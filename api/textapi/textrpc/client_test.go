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

package textrpc_test

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi/textrpc"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler/repl"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ---------------------------------------------------------------
// Test server implementation
// ---------------------------------------------------------------

type testServer struct {
	textrpc.UnimplementedEditorServer

	// onSubscribeCommand is called once the subscription
	// handshake completes. The callback receives the stream
	// so it can send HandleCommandRequest /
	// CompleteCommandRequest messages and read replies.
	onSubscribeCommand func(
		manual *textrpc.CommandManual,
		stream grpc.BidiStreamingServer[
			textrpc.ClientCommandMessage,
			textrpc.ServerCommandMessage,
		],
	) error

	// onSubscribeREPLCommand mirrors the above for REPL
	// commands.
	onSubscribeREPLCommand func(
		manual *textrpc.CommandManual,
		stream grpc.BidiStreamingServer[
			textrpc.ClientREPLCommandMessage,
			textrpc.ServerREPLCommandMessage,
		],
	) error
}

func (s *testServer) SubscribeCommand(
	stream grpc.BidiStreamingServer[
		textrpc.ClientCommandMessage,
		textrpc.ServerCommandMessage,
	],
) error {
	// Receive the subscription request.
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.GetType() != textrpc.ClientCommandMessage_Request {
		return errors.New("expected Request message")
	}
	manual := msg.GetRequest().GetCommand()

	// Acknowledge with a Response.
	err = stream.Send(&textrpc.ServerCommandMessage{
		Type:     textrpc.ServerCommandMessage_Response,
		Response: &textrpc.SubscribeCommandResponse{},
	})
	if err != nil {
		return err
	}

	if s.onSubscribeCommand != nil {
		return s.onSubscribeCommand(manual, stream)
	}
	return nil
}

func (s *testServer) SubscribeREPLCommand(
	stream grpc.BidiStreamingServer[
		textrpc.ClientREPLCommandMessage,
		textrpc.ServerREPLCommandMessage,
	],
) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.GetType() != textrpc.ClientREPLCommandMessage_Request {
		return errors.New("expected Request message")
	}
	manual := msg.GetRequest().GetCommand()

	err = stream.Send(&textrpc.ServerREPLCommandMessage{
		Type: textrpc.ServerREPLCommandMessage_Response,
		Response: &textrpc.SubscribeREPLCommandResponse{},
	})
	if err != nil {
		return err
	}

	if s.onSubscribeREPLCommand != nil {
		return s.onSubscribeREPLCommand(manual, stream)
	}
	return nil
}

// ---------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------

type testEnv struct {
	srv       *grpc.Server
	client    *textrpc.Client
	conn      *grpc.ClientConn
	cancelCtx func()
}

func newTestEnv(
	t *testing.T, impl *testServer,
) *testEnv {
	t.Helper()

	// Use os.MkdirTemp with a short prefix to avoid
	// exceeding the 104-byte unix socket path limit
	// (t.TempDir includes the full test name).
	tmpDir, err := os.MkdirTemp("", "cmdrpc")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	sockPath := filepath.Join(tmpDir, "s.sock")
	lis, err := net.Listen("unix", sockPath)
	require.NoError(t, err)

	srv := grpc.NewServer()
	textrpc.RegisterEditorServer(srv, impl)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient(
		"unix:"+sockPath,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	client := textrpc.NewClient(ctx, conn)

	t.Cleanup(func() {
		_ = client.Close()
		cancel()
		_ = conn.Close()
		srv.Stop()
	})

	return &testEnv{
		srv:       srv,
		client:    client,
		conn:      conn,
		cancelCtx: cancel,
	}
}

// waitChan waits for a value from ch or fails after timeout.
func waitChan[T any](
	t *testing.T, ch <-chan T, timeout time.Duration,
) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatal("timed out waiting for channel")
		var zero T
		return zero
	}
}

const testTimeout = 5 * time.Second

// ---------------------------------------------------------------
// RegisterCommand tests
// ---------------------------------------------------------------

func TestRegisterCommand(t *testing.T) {
	tests := []struct {
		name string

		// manual sent by the client during registration.
		manual textapi.CommandManual

		// serverAction is what the test server does after
		// the handshake completes. It sends requests and
		// reads responses through the stream.
		serverAction func(
			t *testing.T,
			manual *textrpc.CommandManual,
			stream grpc.BidiStreamingServer[
				textrpc.ClientCommandMessage,
				textrpc.ServerCommandMessage,
			],
		)

		// handler is the command handler passed to
		// RegisterCommand. The returned channel is closed
		// when the handler has been invoked (or all
		// completions have been received).
		handler func(t *testing.T) (
			textapi.CommandHandler, <-chan struct{},
		)

		// verify runs assertions after the handler has
		// processed.
		verify func(t *testing.T)
	}{
		{
			name: "handle command dispatched to handler",
			manual: textapi.CommandManual{
				Name:    "test-cmd",
				Summary: "A test command",
			},
			handler: func(t *testing.T) (
				textapi.CommandHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				var got textapi.Command
				h := textapi.NopCommandCompleter(
					func(
						_ context.Context,
						cmd textapi.Command,
					) error {
						got = cmd
						close(done)
						return nil
					},
				)
				t.Cleanup(func() {
					waitChan(t, done, testTimeout)
					assert.Equal(t, "test-cmd", got.Name)
					assert.Equal(
						t,
						[]string{"arg1", "arg2"},
						got.Args,
					)
				})
				return h, done
			},
			serverAction: func(
				t *testing.T,
				manual *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				assert.Equal(t, "test-cmd", manual.GetName())
				assert.Equal(
					t, "A test command", manual.GetSummary(),
				)

				err := stream.Send(&textrpc.ServerCommandMessage{
					Type: textrpc.ServerCommandMessage_Handle,
					Handle: &textrpc.HandleCommandRequest{
						Name: "test-cmd",
						Args: []string{"arg1", "arg2"},
					},
				})
				require.NoError(t, err)

				// Read handle response.
				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Equal(
					t,
					textrpc.ClientCommandMessage_Handle,
					resp.GetType(),
				)
				assert.Empty(t, resp.GetHandle().GetError())
			},
		},
		{
			name: "handle command returns error",
			manual: textapi.CommandManual{
				Name: "fail-cmd",
			},
			handler: func(t *testing.T) (
				textapi.CommandHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				h := textapi.NopCommandCompleter(
					func(
						_ context.Context,
						_ textapi.Command,
					) error {
						defer close(done)
						return errors.New("handler failed")
					},
				)
				return h, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				err := stream.Send(&textrpc.ServerCommandMessage{
					Type: textrpc.ServerCommandMessage_Handle,
					Handle: &textrpc.HandleCommandRequest{
						Name: "fail-cmd",
					},
				})
				require.NoError(t, err)

				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Equal(
					t,
					"handler failed",
					resp.GetHandle().GetError(),
				)
			},
		},
		{
			name: "complete returns values then done",
			manual: textapi.CommandManual{
				Name: "comp-cmd",
			},
			handler: func(t *testing.T) (
				textapi.CommandHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				h := textapi.FuncCommandHandler(
					func(
						_ context.Context,
						_ textapi.Command,
					) error {
						return nil
					},
					func(
						_ context.Context,
						cmd string, _ []string,
					) (iterator.Iterator[string], error) {
						defer close(done)
						return iterator.FromSlice(
							[]string{"aa", "ab", "ac"},
						), nil
					},
				)
				return h, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				err := stream.Send(&textrpc.ServerCommandMessage{
					Type: textrpc.ServerCommandMessage_Complete,
					Complete: &textrpc.CompleteCommandRequest{
						Id:   1,
						Name: "comp-cmd",
						Args: []string{"a"},
					},
				})
				require.NoError(t, err)

				var values []string
				for {
					resp, err := stream.Recv()
					require.NoError(t, err)
					if resp.GetType() ==
						textrpc.ClientCommandMessage_CompleteDone {
						assert.Empty(
							t,
							resp.GetCompleteDone().GetError(),
						)
						break
					}
					assert.Equal(
						t,
						textrpc.ClientCommandMessage_CompleteValue,
						resp.GetType(),
					)
					values = append(
						values,
						resp.GetCompleteValue().GetValue(),
					)
				}
				assert.Equal(
					t,
					[]string{"aa", "ab", "ac"},
					values,
				)
			},
		},
		{
			name: "complete error propagated as done error",
			manual: textapi.CommandManual{
				Name: "comp-err",
			},
			handler: func(t *testing.T) (
				textapi.CommandHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				h := textapi.FuncCommandHandler(
					func(
						_ context.Context,
						_ textapi.Command,
					) error {
						return nil
					},
					func(
						_ context.Context,
						_ string, _ []string,
					) (iterator.Iterator[string], error) {
						defer close(done)
						return nil, errors.New(
							"completion failed",
						)
					},
				)
				return h, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				err := stream.Send(&textrpc.ServerCommandMessage{
					Type: textrpc.ServerCommandMessage_Complete,
					Complete: &textrpc.CompleteCommandRequest{
						Id:   2,
						Name: "comp-err",
						Args: nil,
					},
				})
				require.NoError(t, err)

				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Equal(
					t,
					textrpc.ClientCommandMessage_CompleteDone,
					resp.GetType(),
				)
				assert.Contains(
					t,
					resp.GetCompleteDone().GetError(),
					"completion failed",
				)
			},
		},
		{
			name: "manual with nested sub-commands",
			manual: textapi.CommandManual{
				Name:    "parent",
				Summary: "parent summary",
				Commands: []textapi.CommandManual{
					{
						Name:     "child1",
						Summary:  "child1 summary",
						Synopsis: "child1 [opts]",
					},
					{
						Name:    "child2",
						Summary: "child2 summary",
					},
				},
			},
			handler: func(t *testing.T) (
				textapi.CommandHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				close(done)
				return textapi.NopCommandCompleter(
					func(
						_ context.Context,
						_ textapi.Command,
					) error {
						return nil
					},
				), done
			},
			serverAction: func(
				t *testing.T,
				manual *textrpc.CommandManual,
				_ grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				assert.Equal(t, "parent", manual.GetName())
				require.Len(t, manual.GetCommands(), 2)
				assert.Equal(
					t,
					"child1",
					manual.GetCommands()[0].GetName(),
				)
				assert.Equal(
					t,
					"child1 [opts]",
					manual.GetCommands()[0].GetSynopsis(),
				)
				assert.Equal(
					t,
					"child2",
					manual.GetCommands()[1].GetName(),
				)
			},
		},
		{
			name: "handle command with resource and cursor",
			manual: textapi.CommandManual{
				Name: "goto-def",
			},
			handler: func(t *testing.T) (
				textapi.CommandHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				var got textapi.Command
				h := textapi.NopCommandCompleter(
					func(
						_ context.Context,
						cmd textapi.Command,
					) error {
						got = cmd
						close(done)
						return nil
					},
				)
				t.Cleanup(func() {
					waitChan(t, done, testTimeout)
					assert.Equal(
						t, "file:///main.go",
						got.URI.String(),
					)
					assert.Equal(t, 10, got.Cursor.Content.X)
					assert.Equal(t, 20, got.Cursor.Content.Y)
					assert.Equal(t, 5, got.Cursor.Window.X)
					assert.Equal(t, 15, got.Cursor.Window.Y)
				})
				return h, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				err := stream.Send(&textrpc.ServerCommandMessage{
					Type: textrpc.ServerCommandMessage_Handle,
					Handle: &textrpc.HandleCommandRequest{
						Name: "goto-def",
						ResourceName: &textrpc.URI{
							Uri: "file:///main.go",
						},
						CursorContent: &termrpc.Coordinates{
							X: 10, Y: 20,
						},
						CursorWindow: &termrpc.Coordinates{
							X: 5, Y: 15,
						},
						WindowId: 99,
					},
				})
				require.NoError(t, err)

				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Empty(t, resp.GetHandle().GetError())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverDone := make(chan struct{})
			impl := &testServer{
				onSubscribeCommand: func(
					manual *textrpc.CommandManual,
					stream grpc.BidiStreamingServer[
						textrpc.ClientCommandMessage,
						textrpc.ServerCommandMessage,
					],
				) error {
					defer close(serverDone)
					tt.serverAction(t, manual, stream)
					return nil
				},
			}
			env := newTestEnv(t, impl)

			h, handlerDone := tt.handler(t)
			err := env.client.RegisterCommand(tt.manual, h)
			require.NoError(t, err)

			waitChan(t, handlerDone, testTimeout)
			waitChan(t, serverDone, testTimeout)

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

// ---------------------------------------------------------------
// RegisterREPLCommand tests
// ---------------------------------------------------------------

// testREPLHandler implements textapi.REPLHandler.
type testREPLHandler struct {
	handleFn   func(context.Context, repl.Command, repl.ProgressWriter) (iterator.Iterator[component.Responsive], error)
	completeFn func(context.Context, string, []string) (iterator.Iterator[string], error)
	helpFn     func(context.Context, []string) (iterator.Iterator[component.Responsive], error)
}

func (h *testREPLHandler) HandleCommand(
	ctx context.Context, cmd repl.Command, pw repl.ProgressWriter,
) (iterator.Iterator[component.Responsive], error) {
	return h.handleFn(ctx, cmd, pw)
}

func (h *testREPLHandler) Complete(
	ctx context.Context, cmd string, args []string,
) (iterator.Iterator[string], error) {
	return h.completeFn(ctx, cmd, args)
}

func (h *testREPLHandler) Help(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	return h.helpFn(ctx, args)
}

func TestRegisterREPLCommand(t *testing.T) {
	tests := []struct {
		name   string
		manual textapi.CommandManual

		serverAction func(
			t *testing.T,
			manual *textrpc.CommandManual,
			stream grpc.BidiStreamingServer[
				textrpc.ClientREPLCommandMessage,
				textrpc.ServerREPLCommandMessage,
			],
		)

		handler func(t *testing.T) (
			textapi.REPLHandler, <-chan struct{},
		)
	}{
		{
			name: "handle REPL command returns rows then done",
			manual: textapi.CommandManual{
				Name:    "status",
				Summary: "show status",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context,
						cmd repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						defer close(done)
						assert.Equal(t, "status", cmd.Name)
						assert.Equal(
							t,
							[]string{"--all"},
							cmd.Args,
						)
						s := component.NewResponsiveString(
							"ok",
							component.StringResponsiveConfig{},
						)
						return iterator.FromSlice(
							[]component.Responsive{s},
						), nil
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						return iterator.FromSlice[string](
							nil,
						), nil
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				manual *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				assert.Equal(t, "status", manual.GetName())

				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Handle,
						Handle: &textrpc.HandleREPLCommandRequest{
							Name:  "status",
							Args:  []string{"--all"},
							Width: 80,
						},
					},
				)
				require.NoError(t, err)

				// Collect HandleValue messages until
				// HandleDone.
				var gotRows int
				for {
					resp, err := stream.Recv()
					require.NoError(t, err)
					if resp.GetType() ==
						textrpc.ClientREPLCommandMessage_HandleDone {
						assert.Empty(
							t,
							resp.GetHandleDone().GetError(),
						)
						break
					}
					assert.Equal(
						t,
						textrpc.ClientREPLCommandMessage_HandleValue,
						resp.GetType(),
					)
					gotRows += len(
						resp.GetHandleValue().GetRows(),
					)
				}
				assert.Greater(t, gotRows, 0)
			},
		},
		{
			name: "handle REPL command returns error",
			manual: textapi.CommandManual{
				Name: "fail-repl",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, _ repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						defer close(done)
						return nil, errors.New(
							"repl handler error",
						)
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						return iterator.FromSlice[string](
							nil,
						), nil
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Handle,
						Handle: &textrpc.HandleREPLCommandRequest{
							Name:  "fail-repl",
							Width: 80,
						},
					},
				)
				require.NoError(t, err)

				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Equal(
					t,
					textrpc.ClientREPLCommandMessage_HandleDone,
					resp.GetType(),
				)
				assert.Contains(
					t,
					resp.GetHandleDone().GetError(),
						"repl handler error",
					)
			},
		},
		{
			name: "handle REPL command forwards progress to client",
			manual: textapi.CommandManual{
				Name:    "dl",
				Summary: "download",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, _ repl.Command, pw repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						defer close(done)
						// Report progress before yielding any
						// rows so the server-side wrapper must
						// send a HandleProgress message.
						pw.Progress(25, 100, "B")
						pw.Progress(75, 100, "B")
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						return iterator.FromSlice[string](nil), nil
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Handle,
						Handle: &textrpc.HandleREPLCommandRequest{
							Name:  "dl",
							Width: 80,
						},
					},
				)
				require.NoError(t, err)

				var progresses []*textrpc.HandleREPLCommandProgress
				for {
					resp, err := stream.Recv()
					require.NoError(t, err)
					if resp.GetType() ==
						textrpc.ClientREPLCommandMessage_HandleDone {
						assert.Empty(
							t,
							resp.GetHandleDone().GetError(),
						)
						break
					}
					if resp.GetType() ==
						textrpc.ClientREPLCommandMessage_HandleProgress {
						progresses = append(
							progresses,
							resp.GetHandleProgress(),
						)
					}
				}
				require.Len(t, progresses, 2)
				assert.Equal(t, int64(25), progresses[0].GetProgress())
				assert.Equal(t, int64(100), progresses[0].GetTotal())
				assert.Equal(t, "B", progresses[0].GetUnits())
				assert.Equal(t, int64(75), progresses[1].GetProgress())
				assert.Equal(t, int64(100), progresses[1].GetTotal())
			},
		},
		{
			name: "REPL complete returns values then done",
			manual: textapi.CommandManual{
				Name: "repl-comp",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, _ repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
					completeFn: func(
						_ context.Context, cmd string,
						args []string,
					) (iterator.Iterator[string], error) {
						defer close(done)
						return iterator.FromSlice(
							[]string{"x", "y"},
						), nil
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Complete,
						Complete: &textrpc.CompleteCommandRequest{
							Id:   7,
							Name: "repl-comp",
							Args: []string{"x"},
						},
					},
				)
				require.NoError(t, err)

				var values []string
				for {
					resp, err := stream.Recv()
					require.NoError(t, err)
					if resp.GetType() ==
						textrpc.ClientREPLCommandMessage_CompleteDone {
						assert.Empty(
							t,
							resp.GetCompleteDone().GetError(),
						)
						break
					}
					assert.Equal(
						t,
						textrpc.ClientREPLCommandMessage_CompleteValue,
						resp.GetType(),
					)
					values = append(
						values,
						resp.GetCompleteValue().GetValue(),
					)
				}
				assert.Equal(t, []string{"x", "y"}, values)
			},
		},
		{
			name: "REPL help returns rows then done",
			manual: textapi.CommandManual{
				Name: "help-cmd",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, _ repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						return iterator.FromSlice[string](
							nil,
						), nil
					},
					helpFn: func(
						_ context.Context, args []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						defer close(done)
						assert.Equal(
							t,
							[]string{"sub"},
							args,
						)
						s := component.NewResponsiveString(
							"usage: help-cmd sub",
							component.StringResponsiveConfig{},
						)
						return iterator.FromSlice(
							[]component.Responsive{s},
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Help,
						Help: &textrpc.HelpCommandRequest{
							Args:  []string{"sub"},
							Width: 80,
						},
					},
				)
				require.NoError(t, err)

				var gotRows int
				for {
					resp, err := stream.Recv()
					require.NoError(t, err)
					if resp.GetType() ==
						textrpc.ClientREPLCommandMessage_HelpDone {
						assert.Empty(
							t,
							resp.GetHelpDone().GetError(),
						)
						break
					}
					assert.Equal(
						t,
						textrpc.ClientREPLCommandMessage_HelpValue,
						resp.GetType(),
					)
					gotRows += len(
						resp.GetHelpValue().GetRows(),
					)
				}
				assert.Greater(t, gotRows, 0)
			},
		},
		{
			name: "REPL help returns error",
			manual: textapi.CommandManual{
				Name: "help-err",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, _ repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						return iterator.FromSlice[string](
							nil,
						), nil
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						defer close(done)
						return nil, errors.New("no help")
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Help,
						Help: &textrpc.HelpCommandRequest{
							Args:  nil,
							Width: 80,
						},
					},
				)
				require.NoError(t, err)

				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Equal(
					t,
					textrpc.ClientREPLCommandMessage_HelpDone,
					resp.GetType(),
				)
				assert.Contains(
					t,
					resp.GetHelpDone().GetError(),
					"no help",
				)
			},
		},
		{
			name: "REPL complete error propagated",
			manual: textapi.CommandManual{
				Name: "repl-comp-err",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, _ repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						defer close(done)
						return nil, errors.New(
							"repl complete failed",
						)
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Complete,
						Complete: &textrpc.CompleteCommandRequest{
							Id:   3,
							Name: "repl-comp-err",
						},
					},
				)
				require.NoError(t, err)

				resp, err := stream.Recv()
				require.NoError(t, err)
				assert.Equal(
					t,
					textrpc.ClientREPLCommandMessage_CompleteDone,
					resp.GetType(),
				)
				assert.Contains(
					t,
					resp.GetCompleteDone().GetError(),
					"repl complete failed",
				)
			},
		},
		{
			name: "multiple REPL operations sequentially",
			manual: textapi.CommandManual{
				Name: "multi",
			},
			handler: func(t *testing.T) (
				textapi.REPLHandler, <-chan struct{},
			) {
				done := make(chan struct{})
				var mu sync.Mutex
				var callCount int
				return &testREPLHandler{
					handleFn: func(
						_ context.Context, cmd repl.Command, _ repl.ProgressWriter,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						mu.Lock()
						callCount++
						count := callCount
						mu.Unlock()
						if count == 2 {
							close(done)
						}
						s := component.NewResponsiveString(
							cmd.Name,
							component.StringResponsiveConfig{},
						)
						return iterator.FromSlice(
							[]component.Responsive{s},
						), nil
					},
					completeFn: func(
						_ context.Context, _ string,
						_ []string,
					) (iterator.Iterator[string], error) {
						return iterator.FromSlice[string](
							nil,
						), nil
					},
					helpFn: func(
						_ context.Context, _ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return iterator.FromSlice[component.Responsive](
							nil,
						), nil
					},
				}, done
			},
			serverAction: func(
				t *testing.T,
				_ *textrpc.CommandManual,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				// Send two handle requests sequentially.
				for i := range 2 {
					err := stream.Send(
						&textrpc.ServerREPLCommandMessage{
							Type: textrpc.ServerREPLCommandMessage_Handle,
							Handle: &textrpc.HandleREPLCommandRequest{
								Name:  "multi",
								Args:  []string{string(rune('a' + i))},
								Width: 40,
							},
						},
					)
					require.NoError(t, err)

					// Drain HandleValue until HandleDone.
					for {
						resp, err := stream.Recv()
						require.NoError(t, err)
						if resp.GetType() ==
							textrpc.ClientREPLCommandMessage_HandleDone {
							assert.Empty(
								t,
								resp.GetHandleDone().GetError(),
							)
							break
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverDone := make(chan struct{})
			impl := &testServer{
				onSubscribeREPLCommand: func(
					manual *textrpc.CommandManual,
					stream grpc.BidiStreamingServer[
						textrpc.ClientREPLCommandMessage,
						textrpc.ServerREPLCommandMessage,
					],
				) error {
					defer close(serverDone)
					tt.serverAction(t, manual, stream)
					return nil
				},
			}
			env := newTestEnv(t, impl)

			h, handlerDone := tt.handler(t)
			err := env.client.RegisterREPLCommand(
				tt.manual, h,
			)
			require.NoError(t, err)

			waitChan(t, handlerDone, testTimeout)
			waitChan(t, serverDone, testTimeout)
		})
	}
}

// ---------------------------------------------------------------
// blockingIterator blocks on Next until ctx is canceled.
// ---------------------------------------------------------------

type blockingIterator[T any] struct {
	started chan struct{} // closed when Next is entered
	done    chan struct{} // closed when Next returns
}

func newBlockingIterator[T any]() *blockingIterator[T] {
	return &blockingIterator[T]{
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (it *blockingIterator[T]) Next(
	ctx context.Context,
) (T, bool) {
	close(it.started)
	<-ctx.Done()
	close(it.done)
	var zero T
	return zero, false
}

func (it *blockingIterator[T]) Err() error {
	return nil
}

func (it *blockingIterator[T]) Close() error {
	return nil
}

// ---------------------------------------------------------------
// Stream cancellation tests (command)
// ---------------------------------------------------------------

func TestCommandStreamCancellation(t *testing.T) {
	tests := []struct {
		name string

		// handler returns a handler whose callback blocks
		// until the stream ctx is canceled.
		// blocked is closed once the handler has entered its
		// blocking call.
		// canceled is closed once the handler has observed
		// context cancellation.
		handler func() (
			h textapi.CommandHandler,
			blocked <-chan struct{},
			canceled <-chan struct{},
		)

		// serverAction sends the request that triggers the
		// blocking handler. It should NOT read a response —
		// the stream will be torn down by Close().
		serverAction func(
			t *testing.T,
			stream grpc.BidiStreamingServer[
				textrpc.ClientCommandMessage,
				textrpc.ServerCommandMessage,
			],
		)
	}{
		{
			name: "close cancels blocked handler",
			handler: func() (
				textapi.CommandHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				blocked := make(chan struct{})
				canceled := make(chan struct{})
				h := textapi.NopCommandCompleter(
					func(
						ctx context.Context,
						_ textapi.Command,
					) error {
						close(blocked)
						<-ctx.Done()
						close(canceled)
						return ctx.Err()
					},
				)
				return h, blocked, canceled
			},
			serverAction: func(
				t *testing.T,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerCommandMessage{
						Type: textrpc.ServerCommandMessage_Handle,
						Handle: &textrpc.HandleCommandRequest{
							Name: "block",
						},
					},
				)
				require.NoError(t, err)
			},
		},
		{
			name: "close cancels completion streaming",
			handler: func() (
				textapi.CommandHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				it := newBlockingIterator[string]()
				h := textapi.FuncCommandHandler(
					func(
						_ context.Context,
						_ textapi.Command,
					) error {
						return nil
					},
					func(
						_ context.Context,
						_ string, _ []string,
					) (
						iterator.Iterator[string], error,
					) {
						return it, nil
					},
				)
				return h, it.started, it.done
			},
			serverAction: func(
				t *testing.T,
				stream grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerCommandMessage{
						Type: textrpc.ServerCommandMessage_Complete,
						Complete: &textrpc.CompleteCommandRequest{
							Id:   1,
							Name: "block-comp",
						},
					},
				)
				require.NoError(t, err)
			},
		},
		{
			name: "server stream end tears down client",
			handler: func() (
				textapi.CommandHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				// No blocking — the server will just
				// close the stream immediately. The
				// client's receiveMessages loop should
				// exit on EOF.
				blocked := make(chan struct{})
				canceled := make(chan struct{})
				close(blocked)
				close(canceled)
				return textapi.NopCommandCompleter(
					func(
						_ context.Context,
						_ textapi.Command,
					) error {
						return nil
					},
				), blocked, canceled
			},
			serverAction: func(
				_ *testing.T,
				_ grpc.BidiStreamingServer[
					textrpc.ClientCommandMessage,
					textrpc.ServerCommandMessage,
				],
			) {
				// Return immediately — the server
				// handler returning closes the stream.
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverDone := make(chan struct{})
			impl := &testServer{
				onSubscribeCommand: func(
					_ *textrpc.CommandManual,
					stream grpc.BidiStreamingServer[
						textrpc.ClientCommandMessage,
						textrpc.ServerCommandMessage,
					],
				) error {
					defer close(serverDone)
					tt.serverAction(t, stream)
					return nil
				},
			}
			env := newTestEnv(t, impl)

			h, blocked, canceled := tt.handler()
			man := textapi.CommandManual{Name: "c"}
			err := env.client.RegisterCommand(man, h)
			require.NoError(t, err)

			// Wait for handler to enter blocking state.
			waitChan(t, blocked, testTimeout)

			// Cancel by closing the client.
			_ = env.client.Close()

			// Handler should observe cancellation.
			waitChan(t, canceled, testTimeout)
			waitChan(t, serverDone, testTimeout)
		})
	}
}

// ---------------------------------------------------------------
// Stream cancellation tests (REPL)
// ---------------------------------------------------------------

func TestREPLCommandStreamCancellation(t *testing.T) {
	tests := []struct {
		name string

		handler func() (
			h textapi.REPLHandler,
			blocked <-chan struct{},
			canceled <-chan struct{},
		)

		serverAction func(
			t *testing.T,
			stream grpc.BidiStreamingServer[
				textrpc.ClientREPLCommandMessage,
				textrpc.ServerREPLCommandMessage,
			],
		)
	}{
		{
			name: "close cancels REPL handle streaming",
			handler: func() (
				textapi.REPLHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				it := newBlockingIterator[component.Responsive]()
			h := &testREPLHandler{
				handleFn: func(
					_ context.Context,
					_ repl.Command, _ repl.ProgressWriter,
				) (
					iterator.Iterator[component.Responsive],
					error,
					) {
						return it, nil
					},
					completeFn: nopComplete,
					helpFn:     nopHelp,
				}
				return h, it.started, it.done
			},
			serverAction: func(
				t *testing.T,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Handle,
						Handle: &textrpc.HandleREPLCommandRequest{
							Name:  "block",
							Width: 80,
						},
					},
				)
				require.NoError(t, err)
			},
		},
		{
			name: "close cancels REPL help streaming",
			handler: func() (
				textapi.REPLHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				it := newBlockingIterator[component.Responsive]()
				h := &testREPLHandler{
					handleFn:   nopHandle,
					completeFn: nopComplete,
					helpFn: func(
						_ context.Context,
						_ []string,
					) (
						iterator.Iterator[component.Responsive],
						error,
					) {
						return it, nil
					},
				}
				return h, it.started, it.done
			},
			serverAction: func(
				t *testing.T,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Help,
						Help: &textrpc.HelpCommandRequest{
							Args:  nil,
							Width: 80,
						},
					},
				)
				require.NoError(t, err)
			},
		},
		{
			name: "close cancels REPL completion streaming",
			handler: func() (
				textapi.REPLHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				it := newBlockingIterator[string]()
				h := &testREPLHandler{
					handleFn: nopHandle,
					completeFn: func(
						_ context.Context,
						_ string, _ []string,
					) (
						iterator.Iterator[string], error,
					) {
						return it, nil
					},
					helpFn: nopHelp,
				}
				return h, it.started, it.done
			},
			serverAction: func(
				t *testing.T,
				stream grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				err := stream.Send(
					&textrpc.ServerREPLCommandMessage{
						Type: textrpc.ServerREPLCommandMessage_Complete,
						Complete: &textrpc.CompleteCommandRequest{
							Id:   1,
							Name: "block-comp",
						},
					},
				)
				require.NoError(t, err)
			},
		},
		{
			name: "server stream end tears down REPL client",
			handler: func() (
				textapi.REPLHandler,
				<-chan struct{},
				<-chan struct{},
			) {
				blocked := make(chan struct{})
				canceled := make(chan struct{})
				close(blocked)
				close(canceled)
				return &testREPLHandler{
					handleFn:   nopHandle,
					completeFn: nopComplete,
					helpFn:     nopHelp,
				}, blocked, canceled
			},
			serverAction: func(
				_ *testing.T,
				_ grpc.BidiStreamingServer[
					textrpc.ClientREPLCommandMessage,
					textrpc.ServerREPLCommandMessage,
				],
			) {
				// Return immediately — closes stream.
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverDone := make(chan struct{})
			impl := &testServer{
				onSubscribeREPLCommand: func(
					_ *textrpc.CommandManual,
					stream grpc.BidiStreamingServer[
						textrpc.ClientREPLCommandMessage,
						textrpc.ServerREPLCommandMessage,
					],
				) error {
					defer close(serverDone)
					tt.serverAction(t, stream)
					return nil
				},
			}
			env := newTestEnv(t, impl)

			h, blocked, canceled := tt.handler()
			man := textapi.CommandManual{Name: "r"}
			err := env.client.RegisterREPLCommand(man, h)
			require.NoError(t, err)

			waitChan(t, blocked, testTimeout)
			_ = env.client.Close()
			waitChan(t, canceled, testTimeout)
			waitChan(t, serverDone, testTimeout)
		})
	}
}

// ---------------------------------------------------------------
// nop helpers for testREPLHandler fields
// ---------------------------------------------------------------

func nopHandle(
	_ context.Context, _ repl.Command, _ repl.ProgressWriter,
) (iterator.Iterator[component.Responsive], error) {
	return iterator.FromSlice[component.Responsive](nil), nil
}

func nopComplete(
	_ context.Context, _ string, _ []string,
) (iterator.Iterator[string], error) {
	return iterator.FromSlice[string](nil), nil
}

func nopHelp(
	_ context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	return iterator.FromSlice[component.Responsive](nil), nil
}
