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

package workspacerpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
)

var readBufferSize = 1024 * 64

const watcherWaitTimeout = 2 * time.Minute

type clientCommandStreamer struct {
	req    *StartCommandRequest
	cmd    workspaceapi.Cmd
	log    *slog.Logger
	stream Executor_StartCommandClient
	ctx    context.Context
	quitCh chan struct{}
}

func newClientCommandStreamer(
	ctx context.Context,
	req *StartCommandRequest,
	cmd workspaceapi.Cmd,
	stream Executor_StartCommandClient,
) *clientCommandStreamer {
	ret := new(clientCommandStreamer)
	ret.cmd = cmd
	ret.req = req
	ret.stream = stream
	ret.ctx = ctx
	ret.quitCh = make(chan struct{})
	ret.log = slog.Default().With("struct", "workspacerpc.clientCommandStream")
	return ret
}

func (s *clientCommandStreamer) waitForPid() (workspaceapi.Pid, error) {
	var msg CommandPayload
	err := s.stream.RecvMsg(&msg)
	s.log.Debug("receive first msg", "error", err)
	if err != nil {
		return 0, fmt.Errorf("stream receive msg: %v", err)
	}

	if msg.Type != CommandPayload_TypeStarted || msg.Started == nil {
		return 0, fmt.Errorf("expected stream started msg, found %v", msg.Type)
	}

	return workspaceapi.Pid(msg.Started.Pid), nil
}

func (s *clientCommandStreamer) streamStdin() {
	defer s.stream.CloseSend() // nolint:errcheck

	if s.cmd.Stdin == nil {
		s.log.Debug("no stdin set in cmd, skipping streaming stdin")
		return
	}

	if s.req.StdinFd != 0 {
		s.log.Debug("stdin is a file on the remote server. skipping streaming stdin")
		return
	}

	var err error
	buf := make([]byte, readBufferSize)
	for {
		select {
		case <-s.quitCh:
			return
		case <-s.ctx.Done():
			s.log.Debug("parent context is done")
			return
		default:
		}

		var n int
		n, err = s.cmd.Stdin.Read(buf)
		s.log.Debug("read from stdin", "error", err, "data", n)
		if err != nil && err != io.EOF {
			break
		}
		serr := s.stream.Send(&CommandPayload{
			Type: CommandPayload_TypeIO,
			Io: &CommandPayload_IO{
				Data: buf[:n],
				Type: CommandPayload_IO_TypeStdin,
			},
		})
		if serr != nil {
			err = fmt.Errorf("stream send: %v", serr)
			break
		}
		if err == io.EOF {
			err = nil
			break
		}
	}

	if err != nil {
		s.log.Error("stream stdin stopped with error", "error", err)
	}
}

func (s *clientCommandStreamer) streamCommandData(cancelFn func()) {
	s.log.Debug("streaming command data")
	defer close(s.quitCh)
	defer cancelFn()

	go debug.CapturePanicReport(s.streamStdin)

	var err error
	for {
		var msg CommandPayload
		err = s.stream.RecvMsg(&msg)
		s.log.Debug("receive msg", "error", err)
		if err != nil {
			if err == io.EOF {
				break
			}
			err = fmt.Errorf("stream receive msg: %v", err)
			break
		}

		var n int
		switch msg.Type {
		case CommandPayload_TypeIO:
			switch msg.GetIo().GetType() {
			case CommandPayload_IO_TypeStdout:
				if s.cmd.Stdout == nil {
					s.log.Debug("no stdout set in cmd, dropping data")
					continue
				}
				if s.req.StdoutFd != 0 {
					s.log.Error("stdout is a file on the remote server but received data over the stream")
					return
				}
				if msg.GetIo().Data == nil {
					s.log.Warn("stdout type without stdout data")
					continue
				}
				_, err = s.cmd.Stdout.Write(msg.GetIo().GetData())
				s.log.Debug("wrote to stdout", "error", err, "data", len(msg.GetIo().GetData()))
				if err != nil {
					err = fmt.Errorf("stdout io.Writer write: %v", err)
				}
			case CommandPayload_IO_TypeStderr:
				if s.cmd.Stderr == nil {
					s.log.Debug("no stderr set in cmd, dropping data")
					continue
				}
				if s.req.StderrFd != 0 {
					s.log.Error("stderr is a file on the remote server but received data over the stream")
					return
				}
				if msg.GetIo().Data == nil {
					s.log.Warn("stderr type without stderr data")
					continue
				}
				n, err = s.cmd.Stderr.Write(msg.GetIo().GetData())
				s.log.Debug("wrote to stderr", "error", err, "data", n)
				if err != nil {
					err = fmt.Errorf("stderr io.Writer write: %v", err)
				}
			default:
				err = fmt.Errorf("unexpected io message received %v", msg.GetType())
			}
			if err == nil {
				continue
			}
		case CommandPayload_TypeError:
			err = fmt.Errorf("error reading command stdio: %v", msg.GetError())
		case CommandPayload_TypeDone:
			if msg.GetDone().GetExitError() != "" {
				err = errors.New(msg.GetDone().GetExitError())
			}
		default:
			err = fmt.Errorf("unexpected message received %v", msg.GetType())
		}
		break
	}

	// avoid buggy watchers to cause this goroutine to block forever,
	// so the timeout should be in the order of minutes.
	ctx, cancel := context.WithTimeout(context.Background(),
		watcherWaitTimeout)
	defer cancel()

	if s.cmd.Watcher != nil && s.cmd.Watcher.WatchProcess() != nil {
		select {
		case s.cmd.Watcher.WatchProcess() <- err:
		case <-ctx.Done():
			s.log.Warn("could not deliver error to watcher chan: " +
				"watcher not ready for too long")
		}
	}
	// emulate exec code; pipes should be closed to force EOF
	for _, fd := range [2]io.Writer{s.cmd.Stdout, s.cmd.Stderr} {
		if closer, ok := fd.(io.Closer); ok {
			_ = closer.Close()
		}
	}
	s.log.Debug("done streaming command data", "error", err)
}
