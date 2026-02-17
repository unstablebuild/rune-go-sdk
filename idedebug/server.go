package idedebug

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/google/go-dap"

	"github.com/unstablebuild/rune-go-sdk/api/schemeapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
)

type debugServer struct {
	mu         sync.Mutex
	ctx        context.Context
	cancel     func()
	stopCalled bool
	cfg        debugConfig
	binPath    string
	rootURI    string
	pid        workspaceapi.Pid
	watcher    chan error
	executor   schemeapi.Executor
	conn       net.Conn
	reader     *bufio.Reader
	seq        int32
	pending    map[int]chan dap.Message
	pendingMu  sync.Mutex
	events     chan dap.EventMessage
	alive      bool
	log        *slog.Logger
	caps       *dap.Capabilities
}

func newDebugServer(
	ctx context.Context,
	cfg debugConfig,
	binPath string,
	executor schemeapi.Executor,
	rootURI string,
	events chan dap.EventMessage,
) *debugServer {
	ctx, cancel := context.WithCancel(ctx)
	return &debugServer{
		ctx:      ctx,
		cancel:   cancel,
		cfg:      cfg,
		binPath:  binPath,
		executor: executor,
		rootURI:  rootURI,
		events:   events,
		pending:  make(map[int]chan dap.Message),
		log: slog.With(
			"struct", "idedebug.debugServer",
			"adapter", cfg.id, "uri", rootURI,
		),
	}
}

func (s *debugServer) start(ctx context.Context) error {
	fds, err := syscall.Socketpair(
		syscall.AF_UNIX, syscall.SOCK_STREAM, 0,
	)
	if err != nil {
		return fmt.Errorf("create socket pair: %v", err)
	}
	childEnd := os.NewFile(uintptr(fds[1]), "child-dap")
	ourEnd := os.NewFile(uintptr(fds[0]), "ide-dap")

	watchCh := make(chan error, 1)
	watcher := workspaceapi.ChanProcessWatcher(watchCh)

	cmd := workspaceapi.Cmd{
		Path:    s.binPath,
		Args:    s.cfg.args,
		Stdin:   childEnd,
		Stdout:  childEnd,
		Watcher: watcher,
	}

	// do not use context passed to start as it should
	// only be used for initial protocol exchange
	lifecycleCtx := s.ctx

	s.log.Info("starting server",
		"path", cmd.Path, "args", cmd.Args)
	pid, err := s.executor.StartCommand(
		lifecycleCtx, cmd,
	)
	if err != nil {
		_ = childEnd.Close()
		_ = ourEnd.Close()
		return fmt.Errorf(
			"start %s: %w", s.cfg.command, err,
		)
	}

	conn, err := net.FileConn(ourEnd)
	if err != nil {
		return fmt.Errorf("new file conn: %w", err)
	}
	_ = ourEnd.Close()
	_ = childEnd.Close()

	s.mu.Lock()
	s.pid = pid
	s.watcher = watchCh
	s.conn = conn
	s.reader = bufio.NewReader(conn)
	s.alive = true
	s.mu.Unlock()

	go s.readLoop()

	// initialize is called without s.mu held because it
	// calls sendRequest which also acquires s.mu.
	caps, err := s.initialize(ctx)
	if err != nil {
		_ = s.conn.Close()
		return fmt.Errorf(
			"initialize %s: %w", s.cfg.command, err,
		)
	}

	s.mu.Lock()
	s.caps = caps
	s.mu.Unlock()
	return nil
}

func (s *debugServer) initialize(
	ctx context.Context,
) (*dap.Capabilities, error) {
	s.log.Debug("dap initialize")
	req := &dap.InitializeRequest{
		Request: s.newRequest("initialize"),
		Arguments: dap.InitializeRequestArguments{
			ClientID:        "rune",
			ClientName:      "Rune IDE",
			AdapterID:       s.cfg.id,
			LinesStartAt1:   true,
			ColumnsStartAt1: true,
			PathFormat:      "path",
		},
	}
	resp, err := s.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	initResp, ok := resp.(*dap.InitializeResponse)
	if !ok {
		return nil, fmt.Errorf(
			"unexpected response type: %T", resp,
		)
	}
	return &initResp.Body, nil
}

func (s *debugServer) stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopCalled = true
	if !s.alive {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	req := &dap.DisconnectRequest{
		Request: s.newRequest("disconnect"),
		Arguments: &dap.DisconnectArguments{
			TerminateDebuggee: true,
		},
	}
	// best-effort disconnect
	_, _ = s.sendRequest(ctx, req)

	s.mu.Lock()
	s.alive = false
	s.mu.Unlock()

	return s.conn.Close()
}

func (s *debugServer) sendRequest(
	ctx context.Context, req dap.Message,
) (dap.Message, error) {
	s.mu.Lock()
	if !s.alive {
		s.mu.Unlock()
		return nil, ErrNoServer
	}
	s.mu.Unlock()

	seq := req.GetSeq()
	ch := make(chan dap.Message, 1)
	s.pendingMu.Lock()
	s.pending[seq] = ch
	s.pendingMu.Unlock()

	defer func() {
		s.pendingMu.Lock()
		delete(s.pending, seq)
		s.pendingMu.Unlock()
	}()

	if err := dap.WriteProtocolMessage(
		s.conn, req,
	); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-ch:
		if !ok || msg == nil {
			return nil, fmt.Errorf(
				"server closed connection",
			)
		}
		if resp, ok := msg.(dap.ResponseMessage); ok {
			if !resp.GetResponse().Success {
				return nil, fmt.Errorf(
					"dap error: %s",
					resp.GetResponse().Message,
				)
			}
		}
		return msg, nil
	}
}

func (s *debugServer) readLoop() {
	for {
		msg, err := dap.ReadProtocolMessage(s.reader)
		if err != nil {
			s.mu.Lock()
			alive := s.alive
			s.mu.Unlock()
			if alive {
				s.log.Warn("read error", "error", err)
			}
			s.closePending()
			return
		}

		switch m := msg.(type) {
		case dap.ResponseMessage:
			reqSeq := m.GetResponse().RequestSeq
			s.pendingMu.Lock()
			ch, ok := s.pending[reqSeq]
			s.pendingMu.Unlock()
			if ok {
				ch <- msg
			} else {
				s.log.Warn("unmatched response",
					"requestSeq", reqSeq)
			}
		case dap.EventMessage:
			select {
			case s.events <- m:
			default:
				s.log.Warn(
					"events channel full, dropping",
				)
			}
		default:
			s.log.Debug("unknown message type",
				"type", fmt.Sprintf("%T", msg))
		}
	}
}

func (s *debugServer) closePending() {
	s.pendingMu.Lock()
	defer s.pendingMu.Unlock()
	for seq, ch := range s.pending {
		close(ch)
		delete(s.pending, seq)
	}
}

func (s *debugServer) newRequest(
	command string,
) dap.Request {
	seq := int(atomic.AddInt32(&s.seq, 1))
	return dap.Request{
		ProtocolMessage: dap.ProtocolMessage{
			Seq:  seq,
			Type: "request",
		},
		Command: command,
	}
}
