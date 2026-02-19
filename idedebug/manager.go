// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package idedebug

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/go-dap"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/api/schemeapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/retry"
)

// ErrNoServer is returned when no debug server is
// available for the requested operation.
var ErrNoServer = errors.New("no debug server")

// Config provides optional configuration for a Manager.
type Config struct {
	MaxRetries         int
	InitializeTimeout  time.Duration
	CloseTimeout       time.Duration
	EventHandleTimeout time.Duration
}

// Manager is a multi-language DAP server manager.
// It implements debugapi.Debugger and textapi.EventHandler.
type Manager struct {
	cfg        Config
	evs        chan textapi.Event
	mu         sync.Mutex
	wg         sync.WaitGroup
	rootURI    string
	executor   schemeapi.Executor
	pkgManager PkgManager
	servers    map[string]*debugServer
	starting   map[string]chan struct{}
	activeSrv  *debugServer
	events     chan dap.EventMessage
	ctx        context.Context
	cancel     context.CancelFunc
	log        *slog.Logger
}

var (
	_ textapi.EventHandler = (*Manager)(nil)
	_ debugapi.Debugger    = (*Manager)(nil)
)

// New creates a new Manager with the given dependencies
// and configuration.
func New(
	uri workspaceapi.URI,
	executor schemeapi.Executor,
	pkgManager PkgManager,
	cfg Config,
) *Manager {
	// eventsBufferSize allows bursts of DAP events
	// (e.g. multiple breakpoint hits across threads)
	// without blocking the read loop. When full,
	// readLoop drops events with a warning
	// (see debugServer.readLoop).
	const eventsBufferSize = 64

	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.InitializeTimeout == 0 {
		cfg.InitializeTimeout = 10 * time.Second
	}
	if cfg.CloseTimeout == 0 {
		cfg.CloseTimeout = 5 * time.Second
	}
	if cfg.EventHandleTimeout == 0 {
		cfg.EventHandleTimeout = 1 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	ret := &Manager{
		cfg:        cfg,
		log:        slog.With("struct", "idedebug.Manager"),
		rootURI:    convertURI(uri),
		executor:   executor,
		pkgManager: pkgManager,
		servers:    make(map[string]*debugServer),
		starting:   make(map[string]chan struct{}),
		events: make(
			chan dap.EventMessage, eventsBufferSize,
		),
		ctx:    ctx,
		cancel: cancel,
		// Buffer of 1 for evs: allows Handle to
		// return without blocking in the common
		// case while handleEvs processes.
		evs: make(chan textapi.Event, 1),
	}
	ret.wg.Add(1)
	go ret.handleEvs()
	return ret
}

// Close shuts down all active debug servers and waits
// for all background goroutines to exit.
func (m *Manager) Close() error {
	m.cancel()

	m.mu.Lock()
	servers := make(
		[]*debugServer, 0, len(m.servers),
	)
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

	m.wg.Wait()
	return errors.Join(errs...)
}

// Handle implements textapi.EventHandler.
func (m *Manager) Handle(
	_ context.Context, ev textapi.Event,
) bool {
	m.log.Debug("received event",
		"type", ev.Type, "file", ev.URI)
	select {
	case m.evs <- ev:
		return false
	case <-m.ctx.Done():
		return true
	}
}

// Events returns a channel for receiving debugger events.
func (m *Manager) Events() <-chan dap.EventMessage {
	return m.events
}

func (m *Manager) handleEvs() {
	defer m.wg.Done()
	for {
		select {
		case <-m.ctx.Done():
			return
		case ev := <-m.evs:
			err := m.handle(ev)
			if err != nil {
				m.log.Error(
					"handle event", "error", err,
				)
			}
		}
	}
}

func (m *Manager) handle(ev textapi.Event) error {
	ctx, cancel := context.WithTimeout(
		m.ctx,
		m.cfg.EventHandleTimeout,
	)
	defer cancel()

	switch ev.Type {
	case textapi.EventTypeOpen:
		m.log.Debug("process open", "file", ev.URI)
		_, err := m.ensureServer(ctx, ev.URI)
		return err
	default:
		return nil
	}
}

func (m *Manager) ensureServer(
	ctx context.Context, filename workspaceapi.URI,
) (*debugServer, error) {
	cfg, err := debugAdapterForFile(filename)
	if err != nil {
		return nil, err
	}
	return m.getOrCreateServer(ctx, cfg)
}

// getOrCreateServer returns an existing server for the
// given adapter config, or initializes one. Concurrent
// callers for the same adapter ID will block until the
// first caller's initialization completes.
func (m *Manager) getOrCreateServer(
	ctx context.Context, cfg debugConfig,
) (*debugServer, error) {
	m.mu.Lock()
	if srv, ok := m.servers[cfg.id]; ok {
		m.mu.Unlock()
		return srv, nil
	}

	// Another goroutine is already initializing this
	// adapter — wait for it.
	if ch, ok := m.starting[cfg.id]; ok {
		m.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		m.mu.Lock()
		srv := m.servers[cfg.id]
		m.mu.Unlock()
		if srv == nil {
			return nil, fmt.Errorf(
				"%w: %s init failed",
				ErrNoServer, cfg.id,
			)
		}
		return srv, nil
	}

	// We are the first — mark this adapter as starting.
	ch := make(chan struct{})
	m.starting[cfg.id] = ch
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.starting, cfg.id)
		m.mu.Unlock()
		close(ch) // unblock waiting goroutines
	}()

	return m.initializeServer(ctx, cfg)
}

func (m *Manager) initializeServer(
	ctx context.Context, cfg debugConfig,
) (*debugServer, error) {
	if cfg.command == "" {
		return nil, errors.New(
			"debug adapter config with empty command",
		)
	}
	var binPath string
	if filepath.IsAbs(cfg.command) {
		binPath = cfg.command
	} else {
		var err error
		binPath, err = m.findBinary(ctx, &cfg)
		if err != nil {
			m.log.Debug(
				"find debug adapter executable, "+
					"falling back to using PATH",
				"executable", cfg.command,
				"error", err,
			)
			binPath = cfg.command
		}
	}

	srv := newDebugServer(
		m.ctx, cfg, binPath, m.executor,
		m.rootURI, m.events,
	)

	ctx, cancel := context.WithTimeout(
		ctx, m.cfg.InitializeTimeout,
	)
	defer cancel()

	if err := srv.start(ctx); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.servers[cfg.id] = srv
	m.activeSrv = srv
	m.mu.Unlock()

	m.wg.Add(1)
	go debug.CapturePanicReport(func() {
		defer m.wg.Done()
		m.watchServer(&cfg, srv)
	})
	return srv, nil
}

func (m *Manager) findBinary(
	ctx context.Context, cfg *debugConfig,
) (string, error) {
	files, err := m.pkgManager.LibDir(
		ctx, cfg.id,
	)
	if err != nil {
		return "", fmt.Errorf("lib dir: %w", err)
	}
	defer func() { _ = files.Close() }()

	paths, err := iterator.ToSlice(ctx, files)
	if err != nil {
		return "", fmt.Errorf(
			"lib dir iterator: %w", err,
		)
	}

	for _, file := range paths {
		candidate := filepath.Base(file)
		if candidate != cfg.command {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			return file, nil
		}
	}
	return "", fmt.Errorf(
		"%s not found in any package directory",
		cfg.command,
	)
}

func (m *Manager) watchServer(
	cfg *debugConfig, srv *debugServer,
) {
	defer srv.closeConn()

	ch := srv.watcher
	if ch == nil {
		return
	}

	select {
	case <-m.ctx.Done():
		return
	case err := <-ch:
		if err == nil {
			m.log.Debug("server exited gracefully",
				"adapter", cfg.id)
			return
		}
		m.log.Warn("server crashed",
			"adapter", cfg.id, "error", err)
		srv.mu.Lock()
		stopCalled := srv.stopCalled
		srv.mu.Unlock()
		if stopCalled {
			return
		}
	}

	strategy := retry.CombinedStrategy(
		retry.LimitStrategy(uint(m.cfg.MaxRetries)),
		retry.ExponentialStrategy(
			500*time.Millisecond, 5*time.Second,
		),
	)

	retryErr := retry.Retry(m.ctx, strategy,
		func(ctx context.Context) (bool, error) {
			ctx, cancel := context.WithTimeout(
				ctx, m.cfg.InitializeTimeout,
			)
			defer cancel()

			m.log.Debug("restarting debug adapter",
				"adapter", cfg.id)
			newSrv := newDebugServer(
				m.ctx, srv.cfg, srv.binPath,
				m.executor, m.rootURI, m.events,
			)

			if err := newSrv.start(ctx); err != nil {
				return true, err
			}
			m.mu.Lock()
			m.servers[cfg.id] = newSrv
			m.activeSrv = newSrv
			m.mu.Unlock()

			m.wg.Add(1)
			go debug.CapturePanicReport(func() {
				defer m.wg.Done()
				m.watchServer(cfg, newSrv)
			})
			return false, nil
		})

	if retryErr != nil {
		m.log.Error(
			"debug adapter failed after retries",
			"adapter", cfg.command,
			"retries", m.cfg.MaxRetries,
			"error", retryErr,
		)
	}
}

func (m *Manager) serverForFile(
	path string,
) (*debugServer, error) {
	cfg, err := debugAdapterForFilename(path)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	srv, ok := m.servers[cfg.id]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf(
			"%w: adapter %s not running",
			ErrNoServer, cfg.id,
		)
	}
	return srv, nil
}

func (m *Manager) activeServer() (
	*debugServer, error,
) {
	m.mu.Lock()
	srv := m.activeSrv
	m.mu.Unlock()
	if srv == nil {
		return nil, ErrNoServer
	}
	return srv, nil
}

func convertURI(u workspaceapi.URI) string {
	return "file://" + u.Path()
}
