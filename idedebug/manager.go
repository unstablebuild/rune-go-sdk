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
	MaxRetries         uint
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
	rootURI    string
	executor   schemeapi.Executor
	pkgManager PkgManager
	servers    map[string]*debugServer
	events     chan dap.EventMessage
	ctx        context.Context
	cancel     context.CancelFunc
	log        *slog.Logger
}

var _ textapi.EventHandler = (*Manager)(nil)

// New creates a new Manager with the given dependencies
// and configuration.
func New(
	uri workspaceapi.URI,
	executor schemeapi.Executor,
	pkgManager PkgManager,
	cfg Config,
) *Manager {
	const eventsBufferSize = 5

	if cfg.MaxRetries == 0 {
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
		events: make(
			chan dap.EventMessage, eventsBufferSize,
		),
		ctx:    ctx,
		cancel: cancel,
		evs: make(
			chan textapi.Event, eventsBufferSize,
		),
	}
	go ret.handleEvs()
	return ret
}

// Close shuts down all active debug servers.
func (m *Manager) Close() error {
	defer m.cancel()
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
	defer m.cancel()
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
		context.Background(),
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

	m.mu.Lock()
	if srv, ok := m.servers[cfg.id]; ok {
		m.mu.Unlock()
		return srv, nil
	}
	m.mu.Unlock()

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
	m.mu.Unlock()

	go debug.CapturePanicReport(func() {
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
	defer func() {
		_ = srv.conn.Close()
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
		retry.LimitStrategy(m.cfg.MaxRetries),
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
			m.mu.Unlock()

			go debug.CapturePanicReport(func() {
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
	defer m.mu.Unlock()
	for _, srv := range m.servers {
		return srv, nil
	}
	return nil, ErrNoServer
}

func convertURI(u workspaceapi.URI) string {
	return fmt.Sprintf("file://%s", u.Path())
}
