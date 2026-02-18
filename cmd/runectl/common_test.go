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

package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi/browserrpc"
	"github.com/unstablebuild/rune-go-sdk/api/semanticapi/semanticrpc"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docpb"
	"github.com/unstablebuild/rune-go-sdk/api/syntaxapi/syntaxrpc"
	"github.com/unstablebuild/rune-go-sdk/api/textapi/textrpc"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi/workspacerpc"
	"github.com/unstablebuild/rune-go-sdk/handler/handlerrpc"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/term/termrpc"
	"google.golang.org/grpc"
)

// testEnv contains all mock services and test configuration.
type testEnv struct {
	t        *testing.T
	srv      *grpc.Server
	socket   string
	datadir  string
	notif    *mockNotifications
	wm       *mockWindowManager
	opener   *mockResourceOpener
	scheme   *mockScheme
	editor   *mockEditor
	docStore *mockDocStore
	syntax   *mockSyntax
	executor *mockExecutor
	terminal *mockTerminal
	lsp      *mockLSP
	cleanup  func()
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	datadir := t.TempDir()
	sockPath := filepath.Join(
		datadir, "rune.sock",
	)

	lis, err := net.Listen("unix", sockPath)
	require.NoError(t, err)

	srv := grpc.NewServer()

	notif := &mockNotifications{}
	wm := &mockWindowManager{focusWindowID: 42}
	opener := &mockResourceOpener{}
	scheme := &mockScheme{}
	ed := &mockEditor{cursorX: 10, cursorY: 20}
	ds := newMockDocStore()
	syn := &mockSyntax{}
	exec := &mockExecutor{startPid: 12345}
	term := newMockTerminal()
	lspMock := &mockLSP{}

	browserrpc.RegisterNotificationsServer(
		srv, notif,
	)
	browserrpc.RegisterWindowManagerServer(srv, wm)
	browserrpc.RegisterResourceOpenerServer(
		srv, opener,
	)
	workspacerpc.RegisterSchemeServer(srv, scheme)
	textrpc.RegisterEditorServer(srv, ed)
	docpb.RegisterDocumentStoreServer(srv, ds)
	syntaxrpc.RegisterSyntaxServer(srv, syn)
	workspacerpc.RegisterExecutorServer(srv, exec)
	workspacerpc.RegisterTerminalServer(srv, term)
	semanticrpc.RegisterLSPServer(srv, lspMock)

	go func() { _ = srv.Serve(lis) }()

	t.Setenv("RUNE_SOCKET", sockPath)
	t.Setenv("RUNE_DATADIR", datadir)
	t.Setenv("RUNE_CERT", "")
	t.Setenv("RUNE_TOKEN", "")

	return &testEnv{
		t: t, srv: srv, socket: sockPath,
		datadir: datadir, notif: notif, wm: wm,
		opener: opener, scheme: scheme,
		editor: ed, docStore: ds, syntax: syn,
		executor: exec, terminal: term,
		cleanup: func() {
			srv.Stop()
			term.cleanup()
		},
		lsp: lspMock,
	}
}

func (e *testEnv) run(
	args ...string,
) (string, error) {
	e.t.Helper()

	root := newRootCmd()
	root.SetArgs(args)

	old := os.Stdout
	r, ww, err := os.Pipe()
	require.NoError(e.t, err)
	os.Stdout = ww

	runErr := root.ExecuteContext(context.Background())

	_ = ww.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()

	return buf.String(), runErr
}

// mockNotifications implements browserrpc.NotificationsServer.
type mockNotifications struct {
	browserrpc.UnimplementedNotificationsServer
	lastLevel uint32
	lastMsg   string
}

func (m *mockNotifications) Notify(
	_ context.Context,
	req *browserrpc.NotifyRequest,
) (*browserrpc.NotifyResponse, error) {
	m.lastLevel = req.GetLevel()
	m.lastMsg = req.GetMsg()
	return &browserrpc.NotifyResponse{
		Id: "notif-1",
	}, nil
}

// mockWindowManager implements browserrpc.WindowManagerServer.
type mockWindowManager struct {
	browserrpc.UnimplementedWindowManagerServer
	focusWindowID uint64
	closedID      uint64
}

func (m *mockWindowManager) Focus(
	_ context.Context, _ *browserrpc.FocusRequest,
) (*browserrpc.FocusResponse, error) {
	return &browserrpc.FocusResponse{
		WindowId: m.focusWindowID,
	}, nil
}

func (m *mockWindowManager) CloseWindow(
	_ context.Context,
	req *browserrpc.WindowCloseRequest,
) (*browserrpc.WindowCloseResponse, error) {
	m.closedID = req.GetWindowId()
	return &browserrpc.WindowCloseResponse{}, nil
}

func (m *mockWindowManager) Split(
	stream grpc.BidiStreamingServer[browserrpc.SplitWindowMessage, handlerrpc.ServerMessage],
) error {
	_, err := stream.Recv()
	if err != nil {
		return err
	}
	resp := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Response,
		Response: &handlerrpc.InstallResourceResponse{
			WindowId: 200,
		},
	}
	return stream.Send(resp)
}

func (m *mockWindowManager) SetContent(
	stream grpc.BidiStreamingServer[browserrpc.WindowSetContentMessage, handlerrpc.ServerMessage],
) error {
	_, err := stream.Recv()
	if err != nil {
		return err
	}
	resp := &handlerrpc.ServerMessage{
		Type: handlerrpc.MessageType_Response,
		Response: &handlerrpc.InstallResourceResponse{
			WindowId: 300,
		},
	}
	return stream.Send(resp)
}

// mockResourceOpener implements browserrpc.ResourceOpenerServer.
type mockResourceOpener struct {
	browserrpc.UnimplementedResourceOpenerServer
}

func (m *mockResourceOpener) Open(
	_ context.Context,
	req *browserrpc.OpenResourceRequest,
) (*browserrpc.OpenResourceResponse, error) {
	return &browserrpc.OpenResourceResponse{
		Uri: req.GetResource(),
	}, nil
}

// mockScheme implements workspacerpc.SchemeServer.
type mockScheme struct {
	workspacerpc.UnimplementedSchemeServer
}

func (m *mockScheme) URI(
	_ context.Context,
	req *workspacerpc.URIRequest,
) (*workspacerpc.URIResponse, error) {
	uri := fmt.Sprintf(
		"file://%s", req.GetPath(),
	)
	return &workspacerpc.URIResponse{Uri: uri}, nil
}

// mockEditor implements textrpc.EditorServer.
type mockEditor struct {
	textrpc.UnimplementedEditorServer
	cursorX         int32
	cursorY         int32
	setCursorCalled bool
	setAttrsCalled  bool
	setLocCalled    bool
	moveNextCalled  bool
	movePrevCalled  bool
}

func (m *mockEditor) Editor(
	_ context.Context, _ *textrpc.EditorRequest,
) (*textrpc.EditorResponse, error) {
	return &textrpc.EditorResponse{}, nil
}

func (m *mockEditor) Cursor(
	_ context.Context, _ *textrpc.CursorRequest,
) (*textrpc.CursorResponse, error) {
	return &textrpc.CursorResponse{
		Pos: &termrpc.Coordinates{
			X: m.cursorX, Y: m.cursorY,
		},
	}, nil
}

func (m *mockEditor) SetCursor(
	_ context.Context,
	req *textrpc.SetCursorRequest,
) (*textrpc.SetCursorResponse, error) {
	m.setCursorCalled = true
	m.cursorX = req.GetPos().GetX()
	m.cursorY = req.GetPos().GetY()
	return &textrpc.SetCursorResponse{}, nil
}

func (m *mockEditor) EditCell(
	_ context.Context,
	req *textrpc.EditCellRequest,
) (*textrpc.EditCellResponse, error) {
	return &textrpc.EditCellResponse{
		From: &termrpc.Coordinates{
			X: req.GetStart().GetX(),
			Y: req.GetStart().GetY(),
		},
		To: &termrpc.Coordinates{
			X: req.GetEnd().GetX(),
			Y: req.GetEnd().GetY(),
		},
		Old: "old-text",
	}, nil
}

func (m *mockEditor) RawCells(
	_ context.Context,
	_ *textrpc.RawCellsRequest,
) (*textrpc.RawCellsResponse, error) {
	cells := term.StringToCells("hello world")
	return textrpc.NewRawCellsResponse(cells), nil
}

func (m *mockEditor) SetDefaultAttributes(
	_ context.Context,
	_ *textrpc.SetDefaultAttributesRequest,
) (*textrpc.SetDefaultAttributesResponse, error) {
	m.setAttrsCalled = true
	return &textrpc.SetDefaultAttributesResponse{}, nil
}

func (m *mockEditor) SetLocationList(
	_ context.Context,
	_ *textrpc.SetLocationListRequest,
) (*textrpc.SetLocationListResponse, error) {
	m.setLocCalled = true
	return &textrpc.SetLocationListResponse{}, nil
}

func (m *mockEditor) MoveToNextLocation(
	_ context.Context,
	_ *textrpc.MoveToLocationRequest,
) (*textrpc.MoveToLocationResponse, error) {
	m.moveNextCalled = true
	return &textrpc.MoveToLocationResponse{}, nil
}

func (m *mockEditor) MoveToPrevLocation(
	_ context.Context,
	_ *textrpc.MoveToLocationRequest,
) (*textrpc.MoveToLocationResponse, error) {
	m.movePrevCalled = true
	return &textrpc.MoveToLocationResponse{}, nil
}

// mockDocStore implements docpb.DocumentStoreServer.
type mockDocStore struct {
	docpb.UnimplementedDocumentStoreServer
	docs       map[string][]byte
	lastUpdate *docpb.UpdateDocumentRequest
}

func newMockDocStore() *mockDocStore {
	return &mockDocStore{
		docs: make(map[string][]byte),
	}
}

func (m *mockDocStore) Create(
	_ context.Context,
	req *docpb.CreateDocumentRequest,
) (*docpb.CreateDocumentResponse, error) {
	m.docs[req.GetId()] = req.GetData()
	return &docpb.CreateDocumentResponse{}, nil
}

func (m *mockDocStore) Set(
	_ context.Context,
	req *docpb.SetDocumentRequest,
) (*docpb.DocumentResponse, error) {
	m.docs[req.GetId()] = req.GetData()
	return &docpb.DocumentResponse{}, nil
}

func (m *mockDocStore) Get(
	_ context.Context,
	req *docpb.GetDocumentRequest,
) (*docpb.GetDocumentResponse, error) {
	data, ok := m.docs[req.GetId()]
	if !ok {
		return &docpb.GetDocumentResponse{}, nil
	}
	return &docpb.GetDocumentResponse{
		Data: data,
	}, nil
}

func (m *mockDocStore) Delete(
	_ context.Context,
	req *docpb.DeleteDocumentRequest,
) (*docpb.DocumentResponse, error) {
	delete(m.docs, req.GetId())
	return &docpb.DocumentResponse{}, nil
}

func (m *mockDocStore) Update(
	_ context.Context,
	req *docpb.UpdateDocumentRequest,
) (*docpb.UpdateDocumentResponse, error) {
	m.lastUpdate = req
	return &docpb.UpdateDocumentResponse{}, nil
}

func (m *mockDocStore) List(
	_ *docpb.ListDocumentRequest,
	stream grpc.ServerStreamingServer[docpb.ListDocumentResponse],
) error {
	for _, data := range m.docs {
		err := stream.Send(
			&docpb.ListDocumentResponse{Data: data},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// mockTerminal implements workspacerpc.TerminalServer.
type mockTerminal struct {
	workspacerpc.UnimplementedTerminalServer
	masterR *os.File // Read end for master (code reads from here)
	masterW *os.File // Write end for master (test writes here)
	slaveR  *os.File // Read end for slave
	slaveW  *os.File // Write end for slave
}

func newMockTerminal() *mockTerminal {
	// Create pipes for PTY simulation
	masterR, masterW, _ := os.Pipe()
	slaveR, slaveW, _ := os.Pipe()
	return &mockTerminal{
		masterR: masterR,
		masterW: masterW,
		slaveR:  slaveR,
		slaveW:  slaveW,
	}
}

func (m *mockTerminal) NewPty(
	_ context.Context,
	_ *workspacerpc.NewPtyRequest,
) (*workspacerpc.NewPtyResponse, error) {
	return &workspacerpc.NewPtyResponse{
		Master:   m.masterR.Name(),
		MasterFd: uint32(m.masterR.Fd()),
		Slave:    m.slaveW.Name(),
		SlaveFd:  uint32(m.slaveW.Fd()),
	}, nil
}

func (m *mockTerminal) cleanup() {
	_ = m.masterR.Close()
	_ = m.masterW.Close()
	_ = m.slaveR.Close()
	_ = m.slaveW.Close()
}

// mockExecutor implements workspacerpc.ExecutorServer.
type mockExecutor struct {
	workspacerpc.UnimplementedExecutorServer
	lastPid      int64
	lastSignal   int32
	signalCalled bool
	startPid     int64
	lastDir      string
	lastArgs     []string
	stdoutData   string
	stderrData   string
	exitError    string
}

func (m *mockExecutor) StartCommand(
	stream grpc.BidiStreamingServer[workspacerpc.CommandPayload, workspacerpc.CommandPayload],
) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	if msg.GetType() != workspacerpc.CommandPayload_TypeStart {
		return fmt.Errorf("expected TypeStart, got %v", msg.GetType())
	}
	m.lastDir = msg.GetStart().GetDir()
	m.lastArgs = msg.GetStart().GetArgs()

	resp := &workspacerpc.CommandPayload{
		Type:    workspacerpc.CommandPayload_TypeStarted,
		Started: &workspacerpc.CommandPayload_Started{Pid: m.startPid},
	}
	if err := stream.Send(resp); err != nil {
		return err
	}

	// Stream stdout if configured
	if m.stdoutData != "" {
		stdoutResp := &workspacerpc.CommandPayload{
			Type: workspacerpc.CommandPayload_TypeIO,
			Io: &workspacerpc.CommandPayload_IO{
				Type: workspacerpc.CommandPayload_IO_TypeStdout,
				Data: []byte(m.stdoutData),
			},
		}
		if err := stream.Send(stdoutResp); err != nil {
			return err
		}
	}

	// Stream stderr if configured
	if m.stderrData != "" {
		stderrResp := &workspacerpc.CommandPayload{
			Type: workspacerpc.CommandPayload_TypeIO,
			Io: &workspacerpc.CommandPayload_IO{
				Type: workspacerpc.CommandPayload_IO_TypeStderr,
				Data: []byte(m.stderrData),
			},
		}
		if err := stream.Send(stderrResp); err != nil {
			return err
		}
	}

	// Send done message
	doneResp := &workspacerpc.CommandPayload{
		Type: workspacerpc.CommandPayload_TypeDone,
		Done: &workspacerpc.CommandPayload_Done{ExitError: m.exitError},
	}
	return stream.Send(doneResp)
}

func (m *mockExecutor) Signal(
	_ context.Context,
	req *workspacerpc.SignalRequest,
) (*workspacerpc.SignalResponse, error) {
	m.signalCalled = true
	m.lastPid = req.GetPid()
	m.lastSignal = req.GetSig()
	return &workspacerpc.SignalResponse{}, nil
}

// mockSyntax implements syntaxrpc.SyntaxServer.
type mockSyntax struct {
	syntaxrpc.UnimplementedSyntaxServer
	lastCaptures  []string
	lastNodeTypes uint32
	lastQueryURI  string
	lastQuery     string
}

func (m *mockSyntax) Search(
	req *syntaxrpc.SearchRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastCaptures = req.GetCaptureNames()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         "file:///src/main.go",
		Text:        "func main()",
		From:        &termrpc.Coordinates{X: 0, Y: 5},
		To:          &termrpc.Coordinates{X: 12, Y: 5},
		CaptureName: "fn",
	})
}

func (m *mockSyntax) SearchNode(
	req *syntaxrpc.SearchNodeRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastNodeTypes = req.GetNodeTypes()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         "file:///src/main.go",
		Text:        "MyFunc",
		From:        &termrpc.Coordinates{X: 5, Y: 10},
		To:          &termrpc.Coordinates{X: 11, Y: 10},
		CaptureName: "definition.func",
	})
}

func (m *mockSyntax) Query(
	req *syntaxrpc.QueryRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastQueryURI = req.GetUri()
	m.lastQuery = req.GetQuery()
	m.lastCaptures = req.GetCaptureNames()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         req.GetUri(),
		Text:        "package main",
		From:        &termrpc.Coordinates{X: 0, Y: 0},
		To:          &termrpc.Coordinates{X: 12, Y: 0},
		CaptureName: "pkg",
	})
}

func (m *mockSyntax) QueryNode(
	req *syntaxrpc.QueryNodeRequest,
	stream grpc.ServerStreamingServer[syntaxrpc.SearchResponse],
) error {
	m.lastQueryURI = req.GetUri()
	m.lastNodeTypes = req.GetNodeTypes()
	return stream.Send(&syntaxrpc.SearchResponse{
		Uri:         req.GetUri(),
		Text:        "main",
		From:        &termrpc.Coordinates{X: 8, Y: 0},
		To:          &termrpc.Coordinates{X: 12, Y: 0},
		CaptureName: "definition.namespace",
	})
}

// mockLSP implements semanticrpc.LSPServer.
type mockLSP struct {
	semanticrpc.UnimplementedLSPServer
}

func (m *mockLSP) Hover(
	_ context.Context,
	req *semanticrpc.HoverRequest,
) (*semanticrpc.HoverResponse, error) {
	return &semanticrpc.HoverResponse{
		HasResult: true,
		Result: &semanticrpc.Hover{
			Contents: &semanticrpc.MarkupContent{
				Kind:  1,
				Value: "func main()",
			},
		},
	}, nil
}

func (m *mockLSP) Definition(
	_ context.Context,
	req *semanticrpc.DefinitionRequest,
) (*semanticrpc.DefinitionResponse, error) {
	return &semanticrpc.DefinitionResponse{
		Locations: []*semanticrpc.Location{{
			Uri: "file:///src/main.go",
			Range: &semanticrpc.Range{
				Start: &semanticrpc.Position{
					Line: 10, Character: 5,
				},
				End: &semanticrpc.Position{
					Line: 10, Character: 15,
				},
			},
		}},
	}, nil
}

func (m *mockLSP) References(
	_ context.Context,
	req *semanticrpc.ReferencesRequest,
) (*semanticrpc.ReferencesResponse, error) {
	return &semanticrpc.ReferencesResponse{
		Locations: []*semanticrpc.Location{
			{
				Uri: "file:///src/main.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 10, Character: 5,
					},
					End: &semanticrpc.Position{
						Line: 10, Character: 15,
					},
				},
			},
			{
				Uri: "file:///src/util.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 20, Character: 3,
					},
					End: &semanticrpc.Position{
						Line: 20, Character: 13,
					},
				},
			},
		},
	}, nil
}

func (m *mockLSP) DocumentSymbol(
	_ context.Context,
	req *semanticrpc.DocumentSymbolRequest,
) (*semanticrpc.DocumentSymbolResponse, error) {
	return &semanticrpc.DocumentSymbolResponse{
		Symbols: []*semanticrpc.DocumentSymbol{{
			Name: "main",
			Kind: 12,
			Range: &semanticrpc.Range{
				Start: &semanticrpc.Position{
					Line: 5, Character: 0,
				},
				End: &semanticrpc.Position{
					Line: 10, Character: 1,
				},
			},
			SelectionRange: &semanticrpc.Range{
				Start: &semanticrpc.Position{
					Line: 5, Character: 5,
				},
				End: &semanticrpc.Position{
					Line: 5, Character: 9,
				},
			},
			Children: []*semanticrpc.DocumentSymbol{{
				Name: "x",
				Kind: 13,
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 6, Character: 1,
					},
					End: &semanticrpc.Position{
						Line: 6, Character: 10,
					},
				},
				SelectionRange: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 6, Character: 1,
					},
					End: &semanticrpc.Position{
						Line: 6, Character: 2,
					},
				},
			}},
		}},
	}, nil
}

func (m *mockLSP) WorkspaceSymbol(
	_ context.Context,
	req *semanticrpc.WorkspaceSymbolRequest,
) (*semanticrpc.WorkspaceSymbolResponse, error) {
	return &semanticrpc.WorkspaceSymbolResponse{
		Symbols: []*semanticrpc.SymbolInformation{{
			Name: "MyFunc",
			Kind: 12,
			Location: &semanticrpc.Location{
				Uri: "file:///src/main.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 5, Character: 0,
					},
					End: &semanticrpc.Position{
						Line: 5, Character: 10,
					},
				},
			},
		}},
	}, nil
}

func (m *mockLSP) Diagnostic(
	_ context.Context,
	req *semanticrpc.DiagnosticRequest,
) (*semanticrpc.DiagnosticResponse, error) {
	return &semanticrpc.DiagnosticResponse{
		Report: &semanticrpc.DocumentDiagnosticReport{
			Kind: "full",
			Items: []*semanticrpc.Diagnostic{{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 3, Character: 10,
					},
					End: &semanticrpc.Position{
						Line: 3, Character: 15,
					},
				},
				Severity: 1,
				Code:     "E001",
				Source:   "test",
				Message:  "undefined variable",
			}},
		},
	}, nil
}

func (m *mockLSP) Rename(
	_ context.Context,
	req *semanticrpc.RenameRequest,
) (*semanticrpc.RenameResponse, error) {
	return &semanticrpc.RenameResponse{
		HasResult: true,
		Result: &semanticrpc.WorkspaceEdit{
			Changes: map[string]*semanticrpc.TextEditList{
				"file:///src/main.go": {
					Edits: []*semanticrpc.TextEdit{
						{
							Range: &semanticrpc.Range{
								Start: &semanticrpc.Position{
									Line: 5, Character: 5,
								},
								End: &semanticrpc.Position{
									Line: 5, Character: 9,
								},
							},
							NewText: req.GetNewName(),
						},
						{
							Range: &semanticrpc.Range{
								Start: &semanticrpc.Position{
									Line: 20, Character: 3,
								},
								End: &semanticrpc.Position{
									Line: 20, Character: 7,
								},
							},
							NewText: req.GetNewName(),
						},
					},
				},
			},
		},
	}, nil
}

func (m *mockLSP) CodeAction(
	_ context.Context,
	req *semanticrpc.CodeActionRequest,
) (*semanticrpc.CodeActionResponse, error) {
	return &semanticrpc.CodeActionResponse{
		Items: []*semanticrpc.CodeActionResultItem{
			{
				CodeAction: &semanticrpc.CodeAction{
					Title: "Extract variable",
					Kind:  "refactor.extract",
				},
			},
			{
				CodeAction: &semanticrpc.CodeAction{
					Title: "Organize imports",
					Kind:  "source.organizeImports",
					Edit: &semanticrpc.WorkspaceEdit{
						Changes: map[string]*semanticrpc.TextEditList{
							"file:///src/main.go": {
								Edits: []*semanticrpc.TextEdit{
									{
										Range: &semanticrpc.Range{
											Start: &semanticrpc.Position{
												Line: 2, Character: 0,
											},
											End: &semanticrpc.Position{
												Line: 4, Character: 0,
											},
										},
										NewText: "import (\n\t\"fmt\"\n)\n",
									},
								},
							},
						},
					},
				},
			},
			{
				Command: &semanticrpc.Command{
					Title:     "Run test",
					Command:   "test.run",
					Arguments: []string{`{"uri":"file:///test.go"}`},
				},
			},
		},
	}, nil
}

func (m *mockLSP) Completion(
	_ context.Context,
	req *semanticrpc.CompletionRequest,
) (*semanticrpc.CompletionResponse, error) {
	return &semanticrpc.CompletionResponse{
		Result: &semanticrpc.CompletionResult{
			Items: []*semanticrpc.CompletionItem{
				{
					Label:  "fmt",
					Kind:   9, // Module
					Detail: "package fmt",
				},
				{
					Label:  "Println",
					Kind:   3, // Function
					Detail: "func Println(a ...any)",
				},
			},
		},
	}, nil
}

func (m *mockLSP) SignatureHelp(
	_ context.Context,
	req *semanticrpc.SignatureHelpRequest,
) (*semanticrpc.SignatureHelpResponse, error) {
	return &semanticrpc.SignatureHelpResponse{
		HasResult: true,
		Result: &semanticrpc.SignatureHelp{
			Signatures: []*semanticrpc.SignatureInformation{
				{
					Label: "func Println(a ...any) (n int, err error)",
					Parameters: []*semanticrpc.ParameterInformation{
						{Label: "a ...any"},
					},
				},
			},
			ActiveSignature: 0,
			ActiveParameter: 0,
		},
	}, nil
}

func (m *mockLSP) Declaration(
	_ context.Context,
	req *semanticrpc.DeclarationRequest,
) (*semanticrpc.DeclarationResponse, error) {
	return &semanticrpc.DeclarationResponse{
		Locations: []*semanticrpc.Location{{
			Uri: "file:///src/types.go",
			Range: &semanticrpc.Range{
				Start: &semanticrpc.Position{
					Line: 15, Character: 5,
				},
				End: &semanticrpc.Position{
					Line: 15, Character: 15,
				},
			},
		}},
	}, nil
}

func (m *mockLSP) TypeDefinition(
	_ context.Context,
	req *semanticrpc.TypeDefinitionRequest,
) (*semanticrpc.TypeDefinitionResponse, error) {
	return &semanticrpc.TypeDefinitionResponse{
		Locations: []*semanticrpc.Location{{
			Uri: "file:///src/types.go",
			Range: &semanticrpc.Range{
				Start: &semanticrpc.Position{
					Line: 20, Character: 5,
				},
				End: &semanticrpc.Position{
					Line: 20, Character: 12,
				},
			},
		}},
	}, nil
}

func (m *mockLSP) Implementation(
	_ context.Context,
	req *semanticrpc.ImplementationRequest,
) (*semanticrpc.ImplementationResponse, error) {
	return &semanticrpc.ImplementationResponse{
		Locations: []*semanticrpc.Location{
			{
				Uri: "file:///src/impl1.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 10, Character: 0,
					},
					End: &semanticrpc.Position{
						Line: 10, Character: 10,
					},
				},
			},
			{
				Uri: "file:///src/impl2.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 25, Character: 0,
					},
					End: &semanticrpc.Position{
						Line: 25, Character: 10,
					},
				},
			},
		},
	}, nil
}

func (m *mockLSP) Formatting(
	_ context.Context,
	req *semanticrpc.FormattingRequest,
) (*semanticrpc.FormattingResponse, error) {
	return &semanticrpc.FormattingResponse{
		Edits: []*semanticrpc.TextEdit{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 0, Character: 0,
					},
					End: &semanticrpc.Position{
						Line: 0, Character: 5,
					},
				},
				NewText: "package",
			},
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{
						Line: 5, Character: 0,
					},
					End: &semanticrpc.Position{
						Line: 5, Character: 2,
					},
				},
				NewText: "\t",
			},
		},
	}, nil
}

func (m *mockLSP) PrepareRename(
	_ context.Context,
	req *semanticrpc.PrepareRenameRequest,
) (*semanticrpc.PrepareRenameResponse, error) {
	return &semanticrpc.PrepareRenameResponse{
		HasResult: true,
		Result: &semanticrpc.PrepareRenameResult{
			Range: &semanticrpc.Range{
				Start: &semanticrpc.Position{
					Line: 5, Character: 5,
				},
				End: &semanticrpc.Position{
					Line: 5, Character: 9,
				},
			},
			Placeholder: "main",
		},
	}, nil
}

func (m *mockLSP) DocumentHighlight(
	_ context.Context,
	req *semanticrpc.DocumentHighlightRequest,
) (*semanticrpc.DocumentHighlightResponse, error) {
	return &semanticrpc.DocumentHighlightResponse{
		Highlights: []*semanticrpc.DocumentHighlight{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 5},
					End:   &semanticrpc.Position{Line: 5, Character: 9},
				},
				Kind: 2, // Read
			},
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 10, Character: 3},
					End:   &semanticrpc.Position{Line: 10, Character: 7},
				},
				Kind: 3, // Write
			},
		},
	}, nil
}

func (m *mockLSP) CodeLens(
	_ context.Context,
	req *semanticrpc.CodeLensRequest,
) (*semanticrpc.CodeLensResponse, error) {
	return &semanticrpc.CodeLensResponse{
		Lenses: []*semanticrpc.CodeLens{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 0},
					End:   &semanticrpc.Position{Line: 5, Character: 10},
				},
				Command: &semanticrpc.Command{
					Title:   "Run test",
					Command: "test.run",
				},
			},
		},
	}, nil
}

func (m *mockLSP) RangeFormatting(
	_ context.Context,
	req *semanticrpc.RangeFormattingRequest,
) (*semanticrpc.RangeFormattingResponse, error) {
	return &semanticrpc.RangeFormattingResponse{
		Edits: []*semanticrpc.TextEdit{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 0, Character: 0},
					End:   &semanticrpc.Position{Line: 0, Character: 5},
				},
				NewText: "fixed",
			},
		},
	}, nil
}

func (m *mockLSP) FoldingRange(
	_ context.Context,
	req *semanticrpc.FoldingRangeRequest,
) (*semanticrpc.FoldingRangeResponse, error) {
	return &semanticrpc.FoldingRangeResponse{
		Ranges: []*semanticrpc.FoldingRange{
			{
				StartLine:      5,
				StartCharacter: 0,
				EndLine:        20,
				EndCharacter:   1,
				Kind:           "region",
			},
			{
				StartLine:      1,
				StartCharacter: 0,
				EndLine:        3,
				EndCharacter:   0,
				Kind:           "imports",
			},
		},
	}, nil
}

func (m *mockLSP) SelectionRange(
	_ context.Context,
	req *semanticrpc.SelectionRangeRequest,
) (*semanticrpc.SelectionRangeResponse, error) {
	return &semanticrpc.SelectionRangeResponse{
		Ranges: []*semanticrpc.SelectionRange{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 5},
					End:   &semanticrpc.Position{Line: 5, Character: 9},
				},
				Parent: &semanticrpc.SelectionRange{
					Range: &semanticrpc.Range{
						Start: &semanticrpc.Position{Line: 5, Character: 0},
						End:   &semanticrpc.Position{Line: 10, Character: 1},
					},
				},
			},
		},
	}, nil
}

func (m *mockLSP) ExecuteCommand(
	_ context.Context,
	req *semanticrpc.ExecuteCommandRequest,
) (*semanticrpc.ExecuteCommandResponse, error) {
	return &semanticrpc.ExecuteCommandResponse{
		Result: "command executed: " + req.GetCommand(),
	}, nil
}

func (m *mockLSP) InlayHint(
	_ context.Context,
	req *semanticrpc.InlayHintRequest,
) (*semanticrpc.InlayHintResponse, error) {
	return &semanticrpc.InlayHintResponse{
		Hints: []*semanticrpc.InlayHint{
			{
				Position: &semanticrpc.Position{Line: 10, Character: 15},
				Label:    ": string",
				Kind:     1, // Type
			},
			{
				Position: &semanticrpc.Position{Line: 12, Character: 8},
				Label:    "name:",
				Kind:     2, // Parameter
			},
		},
	}, nil
}

func (m *mockLSP) PrepareCallHierarchy(
	_ context.Context,
	req *semanticrpc.PrepareCallHierarchyRequest,
) (*semanticrpc.PrepareCallHierarchyResponse, error) {
	return &semanticrpc.PrepareCallHierarchyResponse{
		Items: []*semanticrpc.CallHierarchyItem{
			{
				Name: "main",
				Kind: 12, // Function
				Uri:  "file:///src/main.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 0},
					End:   &semanticrpc.Position{Line: 10, Character: 1},
				},
				SelectionRange: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 5},
					End:   &semanticrpc.Position{Line: 5, Character: 9},
				},
			},
		},
	}, nil
}

func (m *mockLSP) CallHierarchyIncomingCalls(
	_ context.Context,
	req *semanticrpc.CallHierarchyIncomingCallsRequest,
) (*semanticrpc.CallHierarchyIncomingCallsResponse, error) {
	return &semanticrpc.CallHierarchyIncomingCallsResponse{
		Calls: []*semanticrpc.CallHierarchyIncomingCall{
			{
				From: &semanticrpc.CallHierarchyItem{
					Name: "runApp",
					Kind: 12,
					Uri:  "file:///src/app.go",
					Range: &semanticrpc.Range{
						Start: &semanticrpc.Position{Line: 20, Character: 0},
						End:   &semanticrpc.Position{Line: 30, Character: 1},
					},
					SelectionRange: &semanticrpc.Range{
						Start: &semanticrpc.Position{Line: 20, Character: 5},
						End:   &semanticrpc.Position{Line: 20, Character: 11},
					},
				},
				FromRanges: []*semanticrpc.Range{
					{
						Start: &semanticrpc.Position{Line: 25, Character: 2},
						End:   &semanticrpc.Position{Line: 25, Character: 8},
					},
				},
			},
		},
	}, nil
}

func (m *mockLSP) CallHierarchyOutgoingCalls(
	_ context.Context,
	req *semanticrpc.CallHierarchyOutgoingCallsRequest,
) (*semanticrpc.CallHierarchyOutgoingCallsResponse, error) {
	return &semanticrpc.CallHierarchyOutgoingCallsResponse{
		Calls: []*semanticrpc.CallHierarchyOutgoingCall{
			{
				To: &semanticrpc.CallHierarchyItem{
					Name: "fmt.Println",
					Kind: 12,
					Uri:  "file:///go/src/fmt/print.go",
					Range: &semanticrpc.Range{
						Start: &semanticrpc.Position{Line: 100, Character: 0},
						End:   &semanticrpc.Position{Line: 110, Character: 1},
					},
					SelectionRange: &semanticrpc.Range{
						Start: &semanticrpc.Position{Line: 100, Character: 5},
						End:   &semanticrpc.Position{Line: 100, Character: 12},
					},
				},
				FromRanges: []*semanticrpc.Range{
					{
						Start: &semanticrpc.Position{Line: 7, Character: 2},
						End:   &semanticrpc.Position{Line: 7, Character: 13},
					},
				},
			},
		},
	}, nil
}

func (m *mockLSP) PrepareTypeHierarchy(
	_ context.Context,
	req *semanticrpc.PrepareTypeHierarchyRequest,
) (*semanticrpc.PrepareTypeHierarchyResponse, error) {
	return &semanticrpc.PrepareTypeHierarchyResponse{
		Items: []*semanticrpc.TypeHierarchyItem{
			{
				Name:   "Reader",
				Kind:   11, // Interface
				Uri:    "file:///src/io.go",
				Detail: "interface",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 15, Character: 0},
					End:   &semanticrpc.Position{Line: 18, Character: 1},
				},
				SelectionRange: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 15, Character: 5},
					End:   &semanticrpc.Position{Line: 15, Character: 11},
				},
			},
		},
	}, nil
}

func (m *mockLSP) TypeHierarchySupertypes(
	_ context.Context,
	req *semanticrpc.TypeHierarchySupertypesRequest,
) (*semanticrpc.TypeHierarchySupertypesResponse, error) {
	return &semanticrpc.TypeHierarchySupertypesResponse{
		Items: []*semanticrpc.TypeHierarchyItem{
			{
				Name: "io.Reader",
				Kind: 11, // Interface
				Uri:  "file:///go/src/io/io.go",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 50, Character: 0},
					End:   &semanticrpc.Position{Line: 53, Character: 1},
				},
				SelectionRange: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 50, Character: 5},
					End:   &semanticrpc.Position{Line: 50, Character: 11},
				},
			},
		},
	}, nil
}

func (m *mockLSP) TypeHierarchySubtypes(
	_ context.Context,
	req *semanticrpc.TypeHierarchySubtypesRequest,
) (*semanticrpc.TypeHierarchySubtypesResponse, error) {
	return &semanticrpc.TypeHierarchySubtypesResponse{
		Items: []*semanticrpc.TypeHierarchyItem{
			{
				Name:   "MyReader",
				Kind:   23, // Struct
				Uri:    "file:///src/reader.go",
				Detail: "struct",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 10, Character: 0},
					End:   &semanticrpc.Position{Line: 15, Character: 1},
				},
				SelectionRange: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 10, Character: 5},
					End:   &semanticrpc.Position{Line: 10, Character: 13},
				},
			},
			{
				Name:   "BufferedReader",
				Kind:   23, // Struct
				Uri:    "file:///src/buffered.go",
				Detail: "struct",
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 8, Character: 0},
					End:   &semanticrpc.Position{Line: 12, Character: 1},
				},
				SelectionRange: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 8, Character: 5},
					End:   &semanticrpc.Position{Line: 8, Character: 19},
				},
			},
		},
	}, nil
}

func (m *mockLSP) SemanticTokensFull(
	_ context.Context,
	req *semanticrpc.SemanticTokensFullRequest,
) (*semanticrpc.SemanticTokensFullResponse, error) {
	return &semanticrpc.SemanticTokensFullResponse{
		HasResult: true,
		Result: &semanticrpc.SemanticTokens{
			ResultId: "test-result-id",
			Data:     []uint32{0, 0, 5, 0, 0, 1, 0, 10, 1, 0},
		},
	}, nil
}

func (m *mockLSP) SemanticTokensRange(
	_ context.Context,
	req *semanticrpc.SemanticTokensRangeRequest,
) (*semanticrpc.SemanticTokensRangeResponse, error) {
	return &semanticrpc.SemanticTokensRangeResponse{
		HasResult: true,
		Result: &semanticrpc.SemanticTokens{
			ResultId: "range-result-id",
			Data:     []uint32{0, 5, 8, 2, 0},
		},
	}, nil
}

func (m *mockLSP) DocumentColor(
	_ context.Context,
	req *semanticrpc.DocumentColorRequest,
) (*semanticrpc.DocumentColorResponse, error) {
	return &semanticrpc.DocumentColorResponse{
		Colors: []*semanticrpc.ColorInformation{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 10},
					End:   &semanticrpc.Position{Line: 5, Character: 17},
				},
				Color: &semanticrpc.Color{
					Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0,
				},
			},
		},
	}, nil
}

func (m *mockLSP) ColorPresentation(
	_ context.Context,
	req *semanticrpc.ColorPresentationRequest,
) (*semanticrpc.ColorPresentationResponse, error) {
	return &semanticrpc.ColorPresentationResponse{
		Presentations: []*semanticrpc.ColorPresentation{
			{Label: "#FF0000"},
			{Label: "rgb(255, 0, 0)"},
		},
	}, nil
}

func (m *mockLSP) DocumentLink(
	_ context.Context,
	req *semanticrpc.DocumentLinkRequest,
) (*semanticrpc.DocumentLinkResponse, error) {
	return &semanticrpc.DocumentLinkResponse{
		Links: []*semanticrpc.DocumentLink{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 3, Character: 5},
					End:   &semanticrpc.Position{Line: 3, Character: 30},
				},
				Target:  "https://example.com/docs",
				Tooltip: "Documentation link",
			},
		},
	}, nil
}

func (m *mockLSP) OnTypeFormatting(
	_ context.Context,
	req *semanticrpc.OnTypeFormattingRequest,
) (*semanticrpc.OnTypeFormattingResponse, error) {
	return &semanticrpc.OnTypeFormattingResponse{
		Edits: []*semanticrpc.TextEdit{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 5, Character: 0},
					End:   &semanticrpc.Position{Line: 5, Character: 2},
				},
				NewText: "\t",
			},
		},
	}, nil
}

func (m *mockLSP) LinkedEditingRange(
	_ context.Context,
	req *semanticrpc.LinkedEditingRangeRequest,
) (*semanticrpc.LinkedEditingRangeResponse, error) {
	return &semanticrpc.LinkedEditingRangeResponse{
		HasResult: true,
		Result: &semanticrpc.LinkedEditingRanges{
			Ranges: []*semanticrpc.Range{
				{
					Start: &semanticrpc.Position{Line: 1, Character: 5},
					End:   &semanticrpc.Position{Line: 1, Character: 10},
				},
				{
					Start: &semanticrpc.Position{Line: 5, Character: 8},
					End:   &semanticrpc.Position{Line: 5, Character: 13},
				},
			},
			WordPattern: "\\w+",
		},
	}, nil
}

func (m *mockLSP) Moniker(
	_ context.Context,
	req *semanticrpc.MonikerRequest,
) (*semanticrpc.MonikerResponse, error) {
	return &semanticrpc.MonikerResponse{
		Monikers: []*semanticrpc.Moniker{
			{
				Scheme:     "go",
				Identifier: "github.com/example/pkg.Func",
				Unique:     "global",
				Kind:       "export",
			},
		},
	}, nil
}

func (m *mockLSP) InlineValue(
	_ context.Context,
	req *semanticrpc.InlineValueRequest,
) (*semanticrpc.InlineValueResponse, error) {
	return &semanticrpc.InlineValueResponse{
		Values: []*semanticrpc.InlineValue{
			{
				Range: &semanticrpc.Range{
					Start: &semanticrpc.Position{Line: 10, Character: 5},
					End:   &semanticrpc.Position{Line: 10, Character: 15},
				},
				Text:         "value: 42",
				VariableName: "x",
			},
		},
	}, nil
}
