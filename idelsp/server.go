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
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/schemeapi"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/idelsp/jsonrpc2"
)

type langServer struct {
	mu         sync.Mutex
	ctx        context.Context
	cancel     func()
	stopCalled bool
	params     semanticapi.InitializeParams
	cfg        langConfig
	binPath    string
	pid        workspaceapi.Pid
	watcher    chan error
	rootURI    string
	executor   schemeapi.Executor
	handler    jsonrpc2.Handler
	stdin      net.Conn
	stdout     net.Conn
	conn       *jsonrpc2.Connection
	alive      bool
	log        *slog.Logger
	init       semanticapi.InitializeResult
}

// pipeCloser wraps read and write pipes to close them together.
type pipeCloser struct {
	r net.Conn
	w net.Conn
}

func (pc pipeCloser) Close() error {
	slog.Debug("closing pipes")
	_ = pc.w.Close()
	return pc.r.Close()
}

func newLangServer(
	ctx context.Context,
	cfg langConfig,
	binPath string,
	executor schemeapi.Executor,
	rootURI string,
	handler jsonrpc2.Handler,
	params semanticapi.InitializeParams,
) *langServer {
	ctx, cancel := context.WithCancel(ctx)
	return &langServer{
		params:   params,
		ctx:      ctx,
		cancel:   cancel,
		cfg:      cfg,
		binPath:  binPath,
		executor: executor,
		rootURI:  rootURI,
		handler:  handler,
		log: slog.With("struct", "idelsp.langServer",
			"language", cfg.id, "uri", rootURI),
	}
}

func (s *langServer) start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return fmt.Errorf("create socket pair: %v", err)
	}
	stdinR := os.NewFile(uintptr(fds[1]), "stdin-lsp")
	stdinW := os.NewFile(uintptr(fds[0]), "stdin-ide")

	stdoutW := os.NewFile(uintptr(fds[1]), "stdout-lsp")
	stdoutR := os.NewFile(uintptr(fds[0]), "stdout-ide")

	watchCh := make(chan error, 1)
	watcher := workspaceapi.ChanProcessWatcher(watchCh)

	cmd := workspaceapi.Cmd{
		Path:    s.binPath,
		Args:    s.cfg.args,
		Stdin:   stdinR,
		Stdout:  stdoutW,
		Watcher: watcher,
	}

	// do not use context passed to start as it sould only be used
	// for initial protocol exchange
	lifecycleContext := s.ctx

	s.log.Info("starting server", "path", cmd.Path, "args", cmd.Args)
	pid, err := s.executor.StartCommand(lifecycleContext, cmd)
	if err != nil {
		_ = stdinR.Close()
		_ = stdinW.Close()
		_ = stdoutR.Close()
		_ = stdoutW.Close()
		return fmt.Errorf(
			"start %s: %w", s.cfg.command, err,
		)
	}

	stdout, err := net.FileConn(stdoutR)
	if err != nil {
		return fmt.Errorf("new stdout file conn: %w", err)
	}
	stdin, err := net.FileConn(stdinW)
	if err != nil {
		return fmt.Errorf("new stdin file conn: %w", err)
	}
	s.pid = pid
	s.watcher = watchCh
	framer := jsonrpc2.HeaderFramer()
	s.conn = jsonrpc2.NewConnection(lifecycleContext, jsonrpc2.ConnectionConfig{
		Reader: framer.Reader(stdout),
		Writer: framer.Writer(stdin),
		Closer: pipeCloser{r: stdout, w: stdin},
		Bind:   func(*jsonrpc2.Connection) jsonrpc2.Handler { return s.handler },
	})
	s.alive = true
	s.stdin = stdin
	s.stdout = stdin

	resp, err := s.initialize(ctx)
	if err != nil {
		_ = s.conn.Close()
		return fmt.Errorf("initialize %s: %w", s.cfg.command, err)
	}
	s.init = resp
	return nil
}

func (s *langServer) initialize(ctx context.Context) (
	response semanticapi.InitializeResult, err error,
) {

	s.log.Debug("rpc call", "method", "initialize")
	err = s.conn.Call(ctx, "initialize", s.params).Await(ctx, &response)
	if err != nil {
		return
	}

	err = s.conn.Notify(ctx, "initialized", struct{}{})
	return
}

func (s *langServer) stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopCalled = true
	if !s.alive {
		s.mu.Unlock()
		return nil
	}
	s.alive = false
	stdin := s.stdin
	s.mu.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		if err := stdin.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set stdout deadline: %v", err)
		}
		defer stdin.SetDeadline(time.Time{}) // nolint:errcheck
	}

	var raw json.RawMessage
	err := s.conn.Call(ctx, "shutdown", nil).Await(ctx, &raw)
	if err != nil {
		return fmt.Errorf("call shutdown: %w", err)
	}
	err = s.conn.Notify(ctx, "exit", nil)
	if err != nil {
		return fmt.Errorf("notify exit: %w", err)
	}
	return nil
}

func (s *langServer) call(
	ctx context.Context, method string,
	params, result any,
) error {
	s.mu.Lock()
	if !s.alive {
		s.mu.Unlock()
		s.log.Warn("rpc call", "method", method, "error", "no server")
		return ErrNoServer
	}
	conn := s.conn
	stdin, stdout := s.stdin, s.stdout
	s.mu.Unlock()
	s.log.Debug("rpc call", "method", method, "step", "attempt")

	// if context has a deadline, set it on the stdin conn for the request
	if deadline, ok := ctx.Deadline(); ok {
		if err := stdin.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set stdout deadline: %v", err)
		}
		defer stdin.SetDeadline(time.Time{}) // nolint:errcheck
	}
	call := conn.Call(ctx, method, params)

	// if context has a deadline, set it on the stdout for the response
	if deadline, ok := ctx.Deadline(); ok {
		if err := stdout.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set stdout deadline: %v", err)
		}
		defer stdout.SetDeadline(time.Time{}) // nolint:errcheck
	}
	err := call.Await(ctx, result)
	if err != nil {
		s.log.Warn("rpc call", "method", method, "error", err)
	} else {
		s.log.Debug("rpc call", "method", method, "step", "success")
	}
	return err
}

func (s *langServer) notify(
	ctx context.Context, method string, params any,
) error {
	s.mu.Lock()
	if !s.alive {
		s.mu.Unlock()
		s.log.Warn("rpc notify", "method", method, "error", "no server")
		return ErrNoServer
	}
	conn := s.conn
	stdin := s.stdin
	s.mu.Unlock()

	// if context has a deadline, set it on the stdin conn for the request
	if deadline, ok := ctx.Deadline(); ok {
		if err := stdin.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set stdout deadline: %v", err)
		}
		defer stdin.SetDeadline(time.Time{}) // nolint:errcheck
	}

	s.log.Debug("rpc notify", "method", method, "step", "attempt")
	err := conn.Notify(ctx, method, params)
	if err != nil {
		s.log.Warn("rpc notify", "method", method, "step", "error", "error", err)
	} else {
		s.log.Debug("rpc notify", "method", method, "step", "success")
	}
	return err
}
