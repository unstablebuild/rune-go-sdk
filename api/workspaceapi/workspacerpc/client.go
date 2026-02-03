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
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/schemeapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"google.golang.org/grpc"
)

const defaultTimeout = 5 * time.Second

// for extension-side
var _ workspaceapi.FileSystem = (*Client)(nil)
var _ workspaceapi.Executor = (*Client)(nil)
var _ workspaceapi.Terminal = (*Client)(nil)

// for scheme registry-side
var _ schemeapi.Scheme = (*Client)(nil)

// Client is a scheme client.
type Client struct {
	cc        grpc.ClientConnInterface
	exec      ExecutorClient
	scheme    SchemeClient
	files     FilesClient
	log       *slog.Logger
	term      TerminalClient
	ctx       context.Context
	root      string // different root than cc's root
	cancelCtx func()
}

// NewClient allocates storage for a new Client and
// initializes it with the given connection. Client satisfies schemeapi.Scheme
// by connecting to a Server via the given the given rpc connection.
func NewClient(ctx context.Context, cc grpc.ClientConnInterface) *Client {
	ret := new(Client)
	ret.Init(ctx, cc)
	return ret
}

// Init initializes this client with cc.
func (c *Client) Init(ctx context.Context, cc grpc.ClientConnInterface) {
	c.cc = cc
	c.scheme = NewSchemeClient(cc)
	c.files = NewFilesClient(cc)
	c.term = NewTerminalClient(cc)
	c.exec = NewExecutorClient(cc)
	c.ctx, c.cancelCtx = context.WithCancel(ctx)
	c.log = slog.Default().With("struct", "workspacerpc.Client")
}

// URI satisfies schemeapi.Scheme.
func (c *Client) URI(path string) (workspaceapi.URI, error) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := URIRequest{Root: c.root, Path: path}
	resp, err := c.scheme.URI(ctx, &req)
	if err != nil {
		return workspaceapi.URI{}, err
	}
	uri, err := workspaceapi.ParseURI(resp.GetUri())
	if err != nil {
		return workspaceapi.URI{}, fmt.Errorf("could not parse URI response from server: %w", err)
	}
	return uri, nil
}

// Chroot satisfies schemeapi.Scheme.
func (c *Client) Chroot(path string) (schemeapi.Scheme, error) {
	uri, err := c.URI(path)
	if err != nil {
		return nil, err
	}
	ret := NewClient(context.Background(), c.cc)
	// force root to be absolute
	ret.root = uri.Path()
	return ret, nil
}

// Root satisfies schemeapi.Scheme.
func (c *Client) Root() string {
	if c.root != "" {
		return c.root
	}
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := RootRequest{}
	resp, err := c.scheme.Root(ctx, &req)
	if err != nil {
		// valid, not useful; best effort
		return "."
	}
	return resp.GetPath()
}

// Symlink satisfies schemeapi.Scheme.
func (c *Client) Symlink(target, link string) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := SymlinkRequest{Root: c.root, Target: target, Link: link}
	_, err := c.scheme.Symlink(ctx, &req)
	return err
}

// TempFile satisfies schemeapi.Scheme.
func (c *Client) TempFile(dir, prefix string) (workspaceapi.File, error) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := TempFileRequest{Root: c.root, Dir: dir, Prefix: prefix}
	resp, err := c.scheme.TempFile(ctx, &req)
	if err != nil {
		return nil, err
	}
	ret := newFileClient(c.root, c.ctx, c, c.cc,
		resp.GetFilename(), uintptr(resp.GetFd()))
	return ret, nil
}

// Join satisfies schemeapi.Scheme.
func (c *Client) Join(elem ...string) string {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := JoinRequest{Elem: elem}
	resp, err := c.scheme.Join(ctx, &req)
	if err != nil {
		// best effort
		return filepath.Join(elem...)
	}
	return resp.GetFilename()
}

// Create satisfies schemeapi.Scheme.
func (c *Client) Create(filename string) (workspaceapi.File, error) {
	return c.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

// Open satisfies schemeapi.Scheme.
func (c *Client) Open(filename string) (workspaceapi.File, error) {
	return c.OpenFile(filename, os.O_RDONLY, 0)
}

// OpenFile satisfies schemeapi.Scheme.
func (c *Client) OpenFile(path string, flag int, mode os.FileMode) (
	workspaceapi.File, error,
) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := makeOpenRequest(c.root, path, flag, mode)
	resp, err := c.scheme.Open(ctx, req)
	if err != nil {
		return nil, err
	}
	if werr, ok := isTypedError(resp); ok {
		return nil, werr
	}
	ret := newFileClient(c.root, c.ctx, c, c.cc,
		resp.GetFilename(), uintptr(resp.GetFd()))
	return ret, nil
}

// Stat returns a FileInfo describing the named file.
func (c *Client) Stat(name string) (os.FileInfo, error) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := StatRequest{Root: c.root, Filename: name}
	resp, err := c.scheme.Stat(ctx, &req)
	if err != nil {
		return nil, err
	}
	if werr, ok := isTypedError(resp); ok {
		return nil, werr
	}
	return &fileClientInfo{StatResponse: *resp}, nil // nolint:govet
}

// ReadDir reads the named directory, returning all its directory entries.
func (c *Client) ReadDir(name string) ([]os.DirEntry, error) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := ReadDirRequest{Dir: name, Root: c.root}
	resp, err := c.scheme.ReadDir(ctx, &req)
	if err != nil {
		return nil, err
	}
	if werr, ok := isTypedError(resp); ok {
		return nil, werr
	}
	respp := resp.GetPath()
	ret := make([]os.DirEntry, 0, len(respp))
	for _, entry := range respp {
		ret = append(ret, dirEntry{
			c:        c,
			name:     entry.Name,
			isDir:    entry.IsDir,
			modeType: entry.Mode,
		})
	}

	return ret, nil
}

// Remove satisfies schemeapi.Scheme.
func (c *Client) Remove(path string) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := RemoveRequest{Root: c.root, Filename: path}
	resp, err := c.scheme.Remove(ctx, &req)
	if err != nil {
		return err
	}
	if werr, ok := isTypedError(resp); ok {
		return werr
	}
	return nil
}

// Rename satisfies schemeapi.Scheme.
func (c *Client) Rename(oldpath, newpath string) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := RenameRequest{Root: c.root, Filename: oldpath, Newfilename: newpath}
	resp, err := c.scheme.Rename(ctx, &req)
	if err != nil {
		return err
	}
	if werr, ok := isTypedError(resp); ok {
		return werr
	}
	return nil
}

// Lstat satisfies schemeapi.Scheme.
func (c *Client) Lstat(name string) (os.FileInfo, error) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := StatRequest{Root: c.root, Filename: name, Lstat: true}
	resp, err := c.scheme.Stat(ctx, &req)
	if err != nil {
		return nil, err
	}
	if werr, ok := isTypedError(resp); ok {
		return nil, werr
	}
	return &fileClientInfo{StatResponse: *resp}, nil // nolint:govet
}

// Readlink satisfies schemeapi.Scheme.
func (c *Client) Readlink(filename string) (string, error) {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := ReadLinkRequest{Root: c.root, Filename: filename}
	resp, err := c.scheme.ReadLink(ctx, &req)
	if err != nil {
		return "", err
	}
	return resp.GetFilename(), nil
}

// MkdirAll satisfies schemeapi.Scheme.
func (c *Client) MkdirAll(path string, perm os.FileMode) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := MkdirAllRequest{Root: c.root, Path: path, Mode: int32(perm)}
	resp, err := c.scheme.MkdirAll(ctx, &req)
	if err != nil {
		return err
	}
	if werr, ok := isTypedError(resp); ok {
		return werr
	}
	return nil
}

// Start satisfies schemeapi.Scheme
func (c *Client) Start(ctx context.Context, cmd workspaceapi.Cmd) (workspaceapi.Pid, error) {
	return c.StartCommand(ctx, cmd)
}

// StartCommand returns the Pid to execute the named program with the given
// arguments. For more details see exec.Command.
func (c *Client) StartCommand(
	commandCtx context.Context, cmd workspaceapi.Cmd,
) (workspaceapi.Pid, error) {
	var cancelFn func()
	commandCtx, cancelFn = mergeContexts(commandCtx, c.ctx)

	stream, err := c.exec.StartCommand(commandCtx)
	if err != nil {
		cancelFn()
		return 0, fmt.Errorf("new stream: %v", err)
	}
	var setsid, setctty bool
	if cmd.SysProcAttr != nil {
		setsid = cmd.SysProcAttr.Setsid
		setctty = cmd.SysProcAttr.Setctty
	}
	req := CommandPayload{
		Type: CommandPayload_TypeStart,
		Start: &StartCommandRequest{
			Name:    cmd.Path,
			Dir:     cmd.Dir,
			Args:    cmd.Args,
			Env:     cmd.Env,
			Stdin:   cmd.Stdin != nil,
			Stdout:  cmd.Stdout != nil,
			Stderr:  cmd.Stderr != nil,
			Setsid:  setsid,
			Setctty: setctty,
		},
	}
	req.Start.StdinFd, req.Start.StdinName = tryUnwrapFile(cmd.Stdin)
	req.Start.StdoutFd, req.Start.StdoutName = tryUnwrapFile(cmd.Stdout)
	req.Start.StderrFd, req.Start.StderrName = tryUnwrapFile(cmd.Stderr)
	streamer := newClientCommandStreamer(commandCtx, req.Start, cmd, stream)

	type result struct {
		pid workspaceapi.Pid
		err error
	}

	// implement rpc timeout
	handshakeCtx, cancel := context.WithTimeout(commandCtx, defaultTimeout)
	defer cancel()

	ch := make(chan result)
	go debug.CapturePanicReport(func() {
		// do not worry about closing stream here something else
		// should take care of closing the connection if deemed appropiate.
		if err := stream.Send(&req); err != nil {
			res := result{err: fmt.Errorf("send start command request: %v", err)}
			select {
			case ch <- res:
			case <-handshakeCtx.Done():
			}
			return
		}
		pid, err := streamer.waitForPid()
		c.log.Debug("wait for pid", "pid", pid, "error", err)
		if err != nil {
			res := result{err: fmt.Errorf("error waiting for pid: %v", err)}
			select {
			case ch <- res:
			case <-handshakeCtx.Done():
			}
			return
		}
		select {
		case ch <- result{pid: pid}:
		case <-handshakeCtx.Done():
		}
	})

	select {
	case res := <-ch:
		if res.err != nil {
			cancelFn()
			return 0, res.err
		}
		go debug.CapturePanicReport(func() {
			streamer.streamCommandData(cancelFn)
		})
		return res.pid, nil
	case <-handshakeCtx.Done():
		cancelFn()
		return 0, handshakeCtx.Err()
	}
}

// Signal sends a signal to the running process.
func (c *Client) Signal(p workspaceapi.Pid, s syscall.Signal) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := SignalRequest{Pid: int64(p), Sig: int32(s)}
	_, err := c.exec.Signal(ctx, &req)
	if err != nil {
		return err
	}
	return nil
}

// StartPty satisfies schemeapi.Scheme
func (c *Client) StartPty() (workspaceapi.Pty, error) {
	return c.NewPty(context.Background())
}

// NewPty creates a new pseudoterminal.
func (c *Client) NewPty(ctx context.Context) (workspaceapi.Pty, error) {
	ctx, cancel := mergeContexts(ctx, c.ctx)
	defer cancel()

	var req NewPtyRequest
	resp, err := c.term.NewPty(ctx, &req)
	if err != nil {
		return workspaceapi.Pty{}, err
	}
	master := newFileClient(c.root, c.ctx, c, c.cc,
		resp.GetMaster(), uintptr(resp.GetMasterFd()))
	slave := newFileClient(c.root, c.ctx, c, c.cc,
		resp.GetSlave(), uintptr(resp.GetSlaveFd()))
	ret := workspaceapi.Pty{
		Master: master,
		Slave:  slave,
	}

	return ret, nil
}

// SetPtySize sets the width and height in columns and rows of
// a pseudoterminal.
func (c *Client) SetPtySize(p workspaceapi.Pty, width, height int) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := SetPtySizeRequest{
		Master:   p.Master.Name(),
		MasterFd: uint32(p.Master.Fd()),
		Slave:    p.Slave.Name(),
		SlaveFd:  uint32(p.Slave.Fd()),
		Width:    int32(width),
		Height:   int32(height),
	}
	_, err := c.term.SetPtySize(ctx, &req)
	return err
}

// NewFile satisfies schemeapi.Scheme.
func (c *Client) NewFile(fd uintptr, filename string) workspaceapi.File {
	return newFileClient(c.root, c.ctx, c, c.cc, filename, fd)
}

// Watch satisfies schemeapi.Scheme.
func (c *Client) Watch(
	path string, ch chan<- schemeapi.EventInfo, events ...schemeapi.Event,
) (int, error) {
	var pbEvents []Event
	for _, ev := range events {
		pbEvents = append(pbEvents, Event(ev))
	}
	req := WatchRequest{Root: c.root, Path: path, Events: pbEvents}
	stream, err := c.scheme.Watch(c.ctx, &req)
	if err != nil {
		return 0, err
	}

	msg, err := stream.Recv()
	if err != nil {
		return 0, fmt.Errorf("stream receive response: %w", err)
	}

	if msg.GetType() != WatchMessage_TypeResponse || msg.GetResponse() == nil {
		return 0, errors.New("received incorrect watch message response")
	}

	go debug.CapturePanicReport(func() {
		defer close(ch)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					c.log.Error("watch stream recv", "error", err)
				}
				break
			}
			if msg.GetType() != WatchMessage_TypeData || msg.GetData() == nil {
				c.log.Error("watch stream recv data: incorrect message type")
				break
			}
			data := msg.GetData()
			uri, err := workspaceapi.ParseURI(data.GetUri())
			if err != nil {
				c.log.Error("could not parse URI response from server", "error", err)
				break
			}
			var ev schemeapi.Event
			switch data.GetEvent() {
			case Event_Create:
				ev = schemeapi.Create
			case Event_Write:
				ev = schemeapi.Write
			case Event_Rename:
				ev = schemeapi.Rename
			case Event_Remove:
				ev = schemeapi.Remove
			}
			fi := watchFileInfo{
				event: ev,
				uri:   uri,
				isDir: data.GetIsDir(),
			}
			select {
			case ch <- fi:
			case <-c.ctx.Done():
				return
			}
		}
	})

	return int(msg.GetResponse().GetId()), nil
}

// StopWatch satisfies schemeapi.Scheme.
func (c *Client) StopWatch(id int) error {
	ctx, cleanup := ctxWithTimeout(c.ctx)
	defer cleanup()

	req := StopWatchRequest{
		Root: c.root,
		Id:   int64(id),
	}
	_, err := c.scheme.StopWatch(ctx, &req)
	return err
}

// Close closes all resources associated with this client.
func (c *Client) Close() (ret error) {
	c.cancelCtx()
	return
}

type fileClientInfo struct {
	StatResponse
}

func (f *fileClientInfo) Name() string {
	return f.GetName()
}

func (f *fileClientInfo) Size() int64 {
	return f.GetSize()
}

func (f *fileClientInfo) Mode() os.FileMode {
	return os.FileMode(f.GetMode())
}

func (f *fileClientInfo) ModTime() time.Time {
	return f.StatResponse.GetModTime().AsTime()
}

func (f *fileClientInfo) IsDir() bool {
	return f.GetIsDir()
}

func (f *fileClientInfo) Sys() interface{} {
	return nil
}

type dirEntry struct {
	c        *Client
	name     string
	isDir    bool
	modeType int32
}

func (e dirEntry) Name() string {
	return e.name
}

func (e dirEntry) IsDir() bool {
	return e.isDir
}

func (e dirEntry) Type() os.FileMode {
	return os.FileMode(e.modeType)
}

func (e dirEntry) Info() (os.FileInfo, error) {
	return e.c.Stat(e.Name())
}

func makeOpenRequest(root, name string, flag int, perm os.FileMode) *OpenRequest {
	return &OpenRequest{
		Root:     root,
		Filename: name,
		Mode:     int32(perm),
		O_RDONLY: flag&^(os.O_APPEND|os.O_CREATE|os.O_EXCL|os.O_SYNC|os.O_TRUNC) == os.O_RDONLY,
		O_WRONLY: flag&^(os.O_APPEND|os.O_CREATE|os.O_EXCL|os.O_SYNC|os.O_TRUNC) == os.O_WRONLY,
		O_APPEND: flag&os.O_APPEND != 0,
		O_CREATE: flag&os.O_CREATE != 0,
		O_EXCL:   flag&os.O_EXCL != 0,
		O_SYNC:   flag&os.O_SYNC != 0,
		O_TRUNC:  flag&os.O_TRUNC != 0,
	}
}

type errResponse interface {
	GetIsExistErr() bool
	GetIsNotExistErr() bool
	GetIsPermissionErr() bool
}

func isTypedError(resp errResponse) (error, bool) {
	switch {
	case resp.GetIsExistErr():
		return os.ErrExist, true
	case resp.GetIsNotExistErr():
		return os.ErrNotExist, true
	case resp.GetIsPermissionErr():
		return os.ErrPermission, true
	default:
		return nil, false
	}
}

func ctxWithTimeout(resourceCtx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithTimeout(resourceCtx, defaultTimeout)
	return ctx, cancel
}

func tryUnwrapFile(ifc interface{}) (uint32, string) {
	fc, ok := ifc.(*FileClient)
	if !ok {
		return 0, ""
	}
	return uint32(fc.Fd()), fc.Name()
}

type watchFileInfo struct {
	event schemeapi.Event
	uri   workspaceapi.URI
	isDir bool
}

func (w watchFileInfo) Event() schemeapi.Event {
	return w.event
}

func (w watchFileInfo) URI() workspaceapi.URI {
	return w.uri
}

func (w watchFileInfo) Sys() interface{} {
	return nil
}

func (w watchFileInfo) IsDir() (bool, error) {
	return w.isDir, nil
}

func mergeContexts(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx1)
	context.AfterFunc(ctx2, cancel)
	return ctx, cancel
}
