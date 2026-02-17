// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2017-2026 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package idelsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/schemeapi"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/retry"
)

// ErrNoServer is returned when no language server is
// available for the requested file.
var ErrNoServer = errors.New("no language server")

// Config provides optional configuration for a Manager.
type Config struct {
	Callback           semanticapi.LSPCallback
	MaxRetries         uint
	InitializeTimeout  time.Duration
	CloseTimeout       time.Duration
	EventHandleTimeout time.Duration
	NoInitializeServer bool
}

// Manager is a multi-language LSP server manager.
// It implements semanticapi.LSP and textapi.EventHandler.
type Manager struct {
	cfg           Config
	evs           chan textapi.Event
	mu            sync.Mutex
	rootURI       string
	fileSystem    schemeapi.FileSystem
	executor      schemeapi.Executor
	notifications browserapi.Notifications
	pkgManager    PkgManager
	callback      semanticapi.LSPCallback
	maxRetries    uint
	servers       map[string]*langServer
	files         map[string]*file
	ctx           context.Context
	cancel        context.CancelFunc
	log           *slog.Logger
}

var (
	_ semanticapi.LSP      = (*Manager)(nil)
	_ textapi.EventHandler = (*Manager)(nil)
)

// New creates a new Manager with the given dependencies and configuration.
func New(
	uri workspaceapi.URI, fileSystem schemeapi.FileSystem,
	executor schemeapi.Executor,
	pkgManager PkgManager, notifications browserapi.Notifications,
	opener browserapi.ResourceOpener,
	cfg Config,
) *Manager {
	const eventsBufferSize = 5

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.InitializeTimeout == 0 {
		cfg.InitializeTimeout = 3 * time.Second
	}
	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = 3 * time.Second
	}
	if cfg.EventHandleTimeout == 0 {
		cfg.EventHandleTimeout = 1 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	ret := &Manager{
		cfg:           cfg,
		log:           slog.With("struct", "idelsp.Manager"),
		rootURI:       convertURI(uri),
		fileSystem:    fileSystem,
		executor:      executor,
		pkgManager:    pkgManager,
		notifications: notifications,
		callback:      cfg.Callback,
		servers:       make(map[string]*langServer),
		files:         make(map[string]*file),
		ctx:           ctx,
		cancel:        cancel,
		evs:           make(chan textapi.Event, eventsBufferSize),
	}
	go ret.handleEvs()
	return ret
}

// Close shuts down all active language servers.
func (m *Manager) Close() error {
	defer m.cancel()
	m.mu.Lock()
	servers := make([]*langServer, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.Unlock()

	var errs []error
	ctx, cancel := context.WithTimeout(
		context.Background(), m.cfg.CloseTimeout,
	)
	defer cancel()
	for _, s := range servers {
		if err := s.stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Handle implements textapi.EventHandler.
func (m *Manager) Handle(
	_ context.Context, ev textapi.Event,
) bool {
	m.log.Debug("received event", "type", ev.Type, "file", ev.URI)
	select {
	case m.evs <- ev:
		return false
	case <-m.ctx.Done():
		return true
	}
}

func (m *Manager) handleEvs() {
	// ensure Handle doesn't block if this goroutine panics
	// and no goroutine is listening on m.evs.
	defer m.cancel()
	for {
		select {
		case <-m.ctx.Done():
			return
		case ev := <-m.evs:
			err := m.handle(ev)
			if err != nil {
				m.log.Error("handle event", "error", err)
			}
		}
	}
}

func (m *Manager) handle(ev textapi.Event) error {
	uri := convertURI(ev.URI)

	ctx, cancel := context.WithTimeout(context.Background(),
		m.cfg.EventHandleTimeout)
	defer cancel()

	switch ev.Type {
	case textapi.EventTypeOpen:
		m.log.Debug("process open", "file", uri, "size", len(ev.Content))
		srv, err := m.ensureServer(ctx, ev.URI)
		if err != nil {
			return err
		}
		f, err := m.ensureFile(ev.URI, ev.Content, srv.cfg.id)
		if err != nil {
			return err
		}
		return srv.notify(ctx, "textDocument/didOpen",
			semanticapi.DidOpenTextDocumentParams{
				TextDocument: semanticapi.TextDocumentItem{
					URI:        uri,
					LanguageID: f.languageID,
					Version:    f.version,
					Text:       ev.Content,
				},
			})

	case textapi.EventTypeClose:
		srv, err := m.serverForURI(uri)
		if err != nil {
			return err
		}
		return srv.notify(ctx, "textDocument/didClose",
			semanticapi.DidCloseTextDocumentParams{
				TextDocument: semanticapi.TextDocumentIdentifier{
					URI: uri,
				},
			})

	case textapi.EventTypeEdit:
		srv, err := m.serverForURI(uri)
		if err != nil {
			return err
		}
		f, err := m.ensureFile(ev.URI, "", srv.cfg.id)
		if err != nil {
			return err
		}
		m.mu.Lock()
		m.files[f.docID.URI].version++
		m.mu.Unlock()
		return srv.notify(ctx, "textDocument/didChange",
			semanticapi.DidChangeTextDocumentParams{
				TextDocument: semanticapi.VersionedTextDocumentIdentifier{
					Version: f.version,
					URI:     uri,
				},
				ContentChanges: []semanticapi.TextDocumentContentChangeEvent{
					{
						Range: &semanticapi.Range{
							Start: semanticapi.Position{
								Line:      uint32(ev.Start.Y),
								Character: uint32(ev.Start.X),
							},
							End: semanticapi.Position{
								Line:      uint32(ev.End.Y),
								Character: uint32(ev.End.X),
							},
						},
						Text: ev.Content,
					},
				},
			})

	case textapi.EventTypeFlush:
		srv, err := m.serverForURI(uri)
		if err != nil {
			return err
		}
		f, err := m.ensureFile(ev.URI, ev.Content, srv.cfg.id)
		if err != nil {
			return err
		}
		return srv.notify(ctx, "textDocument/didSave",
			semanticapi.DidSaveTextDocumentParams{
				TextDocument: semanticapi.TextDocumentIdentifier{
					URI: uri,
				},
				Text: f.content,
			})

	case textapi.EventTypeCreate:
		return m.broadcastNotify(ctx, ev.URI, "workspace/didCreateFiles",
			semanticapi.CreateFilesParams{
				Files: []semanticapi.FileCreate{
					{URI: uri},
				},
			})

	case textapi.EventTypeChange:
		return m.broadcastNotify(ctx,
			ev.URI, "workspace/didChangeWatchedFiles",
			semanticapi.DidChangeWatchedFilesParams{
				Changes: []semanticapi.FileEvent{
					{
						URI:  uri,
						Type: semanticapi.FileChangeTypeChanged,
					},
				},
			})

	case textapi.EventTypeRemove:
		return m.broadcastNotify(ctx, ev.URI, "workspace/didDeleteFiles",
			semanticapi.DeleteFilesParams{
				Files: []semanticapi.FileDelete{
					{URI: uri},
				},
			})

	case textapi.EventTypeRename:
		return m.broadcastNotify(ctx, ev.URI, "workspace/didRenameFiles",
			semanticapi.RenameFilesParams{
				Files: []semanticapi.FileRename{
					{OldURI: uri, NewURI: ev.Content},
				},
			})
	default:
		return fmt.Errorf("extraneous event %v", ev.Type)
	}
}

func (m *Manager) ensureFile(
	uri workspaceapi.URI,
	content string, languageID string,
) (*file, error) {
	uriStr := convertURI(uri)
	if languageID == "" || uri == (workspaceapi.URI{}) {
		panic("empty params for ensuring available file")
	}
	m.mu.Lock()
	f, ok := m.files[uriStr]
	if ok {
		if content != "" {
			m.files[uriStr].content = content
			m.files[uriStr].version++
		}
		m.mu.Unlock()
		return f, nil
	}
	m.mu.Unlock()
	if content == "" {
		f, err := m.fileSystem.Open(uri.Path())
		if err != nil {
			return nil, fmt.Errorf("open file for reading: %w", err)
		}
		defer f.Close() // nolint:errcheck
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("read workspace file: %w", err)
		}
		content = string(data)
	}
	f = newFile(uri, content, languageID)
	m.mu.Lock()
	m.files[uriStr] = f
	m.mu.Unlock()

	return f, nil
}

func (m *Manager) ensureServer(
	ctx context.Context, filename workspaceapi.URI,
) (*langServer, error) {
	if m.cfg.NoInitializeServer {
		return nil, fmt.Errorf("server not initialized and auto-initialize config is false")
	}
	lang, err := languageForFile(filename)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	if srv, ok := m.servers[lang.id]; ok {
		m.mu.Unlock()
		return srv, nil
	}
	m.mu.Unlock()

	return m.initializeServer(ctx, lang, autoInitParams(m.rootURI))
}

func (m *Manager) initializeServer(
	ctx context.Context, lang langConfig, params semanticapi.InitializeParams,
) (*langServer, error) {
	if lang.command == "" {
		return nil, errors.New("language configuration with empty command")
	}
	var binPath string
	if filepath.IsAbs(lang.command) {
		binPath = lang.command
	} else {
		var err error
		binPath, err = m.findBinary(ctx, &lang)
		if err != nil {
			m.log.Debug(
				"find lsp executable, falling back to using PATH",
				"executable", lang.command, "error", err,
			)
			binPath = lang.command
		}
	}

	srv := newLangServer(
		m.ctx, lang, binPath, m.executor, m.rootURI,
		newCallbackAdapter(m.callback),
		params,
	)

	ctx, cancel := context.WithTimeout(ctx, m.cfg.InitializeTimeout)
	defer cancel()

	if err := srv.start(ctx); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.servers[lang.id] = srv
	m.mu.Unlock()

	go debug.CapturePanicReport(func() {
		m.watchServer(&lang, srv)
	})
	return srv, nil
}

func (m *Manager) findBinary(
	ctx context.Context, lang *langConfig,
) (string, error) {
	files, err := m.pkgManager.LibDir(
		ctx, lang.id,
	)
	if err != nil {
		return "", fmt.Errorf("lib dir: %w", err)
	}
	defer func() { _ = files.Close() }()

	paths, err := iterator.ToSlice(ctx, files)
	if err != nil {
		return "", fmt.Errorf("lib dir iterator: %w", err)
	}

	for _, file := range paths {
		candidate := filepath.Base(file)
		if candidate != lang.command {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return file, nil
		}
	}
	return "", fmt.Errorf(
		"%s not found in any package directory",
		lang.command,
	)
}

func (m *Manager) watchServer(
	lang *langConfig, srv *langServer,
) {
	defer func() {
		err := srv.conn.Close()
		if err != nil {
			m.log.Warn("jsonrpc2 connection close", "error", err)
		}
	}()

	ch := srv.watcher
	if ch == nil {
		return
	}

	select {
	case <-m.ctx.Done():
		return
	case err := <-ch:
		if err == nil {
			m.log.Debug("server exited gracefully", "language", lang.id)
			return
		}
		m.log.Warn("server crashed",
			"language", lang.id, "error", err)
		srv.mu.Lock()
		stopCalled := srv.stopCalled
		srv.mu.Unlock()
		if stopCalled {
			return
		}
	}

	strategy := retry.CombinedStrategy(
		retry.LimitStrategy(m.maxRetries),
		retry.ExponentialStrategy(500*time.Millisecond, 5*time.Second),
	)

	retryErr := retry.Retry(m.ctx, strategy,
		func(ctx context.Context) (bool, error) {
			ctx, cancel := context.WithTimeout(ctx, m.cfg.InitializeTimeout)
			defer cancel()

			m.log.Debug("restarting lsp server", "language", lang.id)
			srv := newLangServer(
				m.ctx, srv.cfg, srv.binPath, m.executor, m.rootURI,
				newCallbackAdapter(m.callback),
				srv.params,
			)

			if err := srv.start(ctx); err != nil {
				return true, err
			}
			m.mu.Lock()
			m.servers[lang.id] = srv
			m.mu.Unlock()

			m.reopenFiles(ctx, lang.id, srv)
			go debug.CapturePanicReport(func() {
				m.watchServer(lang, srv)
			})
			return false, nil
		})

	if retryErr != nil && m.notifications != nil {
		_, _ = m.notifications.Notify(
			browserapi.LevelError,
			"LSP server %s failed after %d retries: %s",
			lang.command, m.maxRetries, retryErr,
		)
	}
}

func (m *Manager) reopenFiles(
	ctx context.Context, langID string,
	srv *langServer,
) {
	m.mu.Lock()
	var files []*file
	for _, file := range m.files {
		if file.languageID == langID {
			files = append(files, file)
		}
	}
	m.mu.Unlock()

	for _, f := range files {
		_ = srv.notify(ctx, "textDocument/didOpen",
			semanticapi.DidOpenTextDocumentParams{
				TextDocument: semanticapi.TextDocumentItem{
					URI:        f.docID.URI,
					LanguageID: langID,
					Version:    f.version,
					Text:       f.content,
				},
			})
	}
}

func (m *Manager) serverForURI(uri string) (*langServer, error) {
	m.mu.Lock()
	cfg, err := languageForFilename(uri)
	if err != nil {
		m.mu.Unlock()
		return nil, err
	}
	srv, ok := m.servers[cfg.id]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("%w: server %s not running", ErrNoServer, cfg.id)
	}
	return srv, nil
}

func (m *Manager) allServers() []*langServer {
	m.mu.Lock()
	defer m.mu.Unlock()
	servers := make(
		[]*langServer, 0, len(m.servers),
	)
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	return servers
}

func (m *Manager) broadcastNotify(
	ctx context.Context, uri workspaceapi.URI, method string, params any,
) (ret error) {
	if uri == (workspaceapi.URI{}) {
		for _, srv := range m.allServers() {
			if err := srv.notify(ctx, method, params); err != nil {
				ret = errors.Join(ret, err)
			}
		}
		return
	}
	cfg, err := languageForFile(uri)
	if err != nil {
		return err
	}
	for _, srv := range m.allServers() {
		if srv.cfg.id != cfg.id {
			continue
		}
		if err := srv.notify(ctx, method, params); err != nil {
			ret = errors.Join(ret, err)
		}
	}
	return ret
}

func autoInitParams(rootURI string) semanticapi.InitializeParams {
	capabilities := map[string]any{
		"textDocument": map[string]any{
			"completion":     map[string]any{},
			"hover":          map[string]any{},
			"signatureHelp":  map[string]any{},
			"definition":     map[string]any{},
			"references":     map[string]any{},
			"documentSymbol": map[string]any{},
			"formatting":     map[string]any{},
			"rename": map[string]any{
				"prepareSupport": true,
			},
			"codeAction": map[string]any{},
			"foldingRange": map[string]any{
				"lineFoldingOnly": false,
			},
			"selectionRange":    map[string]any{},
			"documentHighlight": map[string]any{},
			"callHierarchy":     map[string]any{},
			"codeLens":          map[string]any{},
			"inlayHint":         map[string]any{},
			"semanticTokens": map[string]any{
				"requests": map[string]any{
					"full":  true,
					"range": true,
				},
				"tokenTypes": []string{
					"namespace", "type", "class",
					"enum", "interface", "struct",
					"typeParameter", "parameter",
					"variable", "property",
					"enumMember", "event",
					"function", "method", "macro",
					"keyword", "modifier",
					"comment", "string", "number",
					"regexp", "operator",
					"decorator", "label",
				},
				"tokenModifiers": []string{
					"declaration", "definition",
					"readonly", "static",
					"deprecated", "abstract",
					"async", "modification",
					"documentation",
					"defaultLibrary",
				},
				"formats": []string{"relative"},
			},
		},
		"workspace": map[string]any{
			"symbol": map[string]any{},
		},
	}
	capabilitiesData, err := json.Marshal(capabilities)
	if err != nil {
		panic("marshal capabilities")
	}
	return semanticapi.InitializeParams{
		RootURI:      rootURI,
		Capabilities: json.RawMessage(capabilitiesData),
	}
}
