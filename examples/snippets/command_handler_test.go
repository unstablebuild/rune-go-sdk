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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagestub"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

func TestHandleCommandInsert(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "greet", "hello world")
	s := d.build()

	at := term.Coordinates{X: 2, Y: 3}
	cmd := textapi.Command{
		Name:     "snippets",
		Args:     []string{"insert", "greet"},
		Resource: &fakeResource{},
	}
	cmd.Cursor.Content = at

	if err := s.HandleCommand(context.Background(), cmd); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	spy := d.editor.editSpy
	if !spy.called {
		t.Fatal("expected CellEditor.Edit to be called")
	}
	if spy.str != "hello world" {
		t.Errorf("body = %q, want %q", spy.str, "hello world")
	}
	if spy.start != at || spy.end != at {
		t.Errorf("edit range = %v..%v, want %v", spy.start, spy.end, at)
	}
}

func TestHandleCommandInsertMissing(t *testing.T) {
	d := newDeps()
	s := d.build()

	cmd := textapi.Command{Args: []string{"insert", "nope"}, Resource: &fakeResource{}}
	err := s.HandleCommand(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected an error for missing snippet")
	}
	if d.editor.editSpy.called {
		t.Error("expected no edit for missing snippet")
	}
	// The handler returns the error and lets the host notify, rather than
	// notifying itself.
	if len(d.notif.notifs) != 0 {
		t.Errorf("expected no notifications, got %#v", d.notif.notifs)
	}
}

func TestHandleCommandDelete(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "greet", "hello")
	s := d.build()

	cmd := textapi.Command{Args: []string{"delete", "greet"}}
	if err := s.HandleCommand(context.Background(), cmd); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	var doc snippet
	err := d.storage.Get(context.Background(), "greet", &doc)
	if err == nil {
		t.Fatal("expected snippet to be deleted")
	}
}

func TestHandleCommandUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"bogus"}} {
		d := newDeps()
		s := d.build()
		err := s.HandleCommand(context.Background(), textapi.Command{Args: args})
		if err == nil {
			t.Fatalf("HandleCommand(%v): expected error", args)
		}
		if !strings.Contains(err.Error(), usage) {
			t.Errorf("args %v: error %q does not contain usage %q", args, err, usage)
		}
		if len(d.notif.notifs) != 0 {
			t.Errorf("args %v: expected no notifications, got %#v", args, d.notif.notifs)
		}
	}
}

func TestHandleCommandEditOpensBuffer(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "greet", "existing body")
	s := d.build()

	win := &fakeWindow{id: 7}
	cmd := textapi.Command{Args: []string{"edit", "greet"}, Window: win}
	if err := s.HandleCommand(context.Background(), cmd); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	uri, seed := readOpened(t, d)
	if filepath.Base(uri.Path()) != "greet.snippet" {
		t.Errorf("tab file = %q, want greet.snippet", filepath.Base(uri.Path()))
	}
	if seed != "existing body" {
		t.Errorf("seed = %q, want %q", seed, "existing body")
	}

	// The dispatching window must be switched to the opened buffer.
	if len(d.wm.setContent) != 1 {
		t.Fatalf("expected one SetWindowContent call, got %d", len(d.wm.setContent))
	}
	call := d.wm.setContent[0]
	if call.win != win {
		t.Errorf("focused window = %v, want %v", call.win, win)
	}
	if call.h != d.opener.handle {
		t.Errorf("focused handle = %v, want opened handle %v", call.h, d.opener.handle)
	}
}

func TestHandleCommandEditWithoutWindowDoesNotFocus(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "greet", "existing body")
	s := d.build()

	// No Window on the command (e.g. dispatched while a non-tab is in focus).
	cmd := textapi.Command{Args: []string{"edit", "greet"}}
	if err := s.HandleCommand(context.Background(), cmd); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	if len(d.opener.opened) != 1 {
		t.Fatalf("expected one opened buffer, got %d", len(d.opener.opened))
	}
	if len(d.wm.setContent) != 0 {
		t.Errorf("expected no SetWindowContent call without a window, got %d", len(d.wm.setContent))
	}
}

func TestHandleCommandCopySeedsFromSelection(t *testing.T) {
	d := newDeps()
	s := d.build()

	// User selects text first.
	s.Handle(context.Background(), textapi.Event{
		Type:    textapi.EventTypeSelection,
		Content: "selected text",
	})

	cmd := textapi.Command{Args: []string{"copy", "fromsel"}}
	if err := s.HandleCommand(context.Background(), cmd); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	uri, seed := readOpened(t, d)
	if filepath.Base(uri.Path()) != "fromsel.snippet" {
		t.Errorf("tab file = %q, want fromsel.snippet", filepath.Base(uri.Path()))
	}
	if seed != "selected text" {
		t.Errorf("seed = %q, want %q", seed, "selected text")
	}
}

func TestEditNewSnippetPersistedOnClose(t *testing.T) {
	d := newDeps()
	s := d.build()

	cmd := textapi.Command{Args: []string{"edit", "fresh"}}
	if err := s.HandleCommand(context.Background(), cmd); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	uri := d.opener.opened[0]
	// The editor flushes the buffer to its backing file before closing; on
	// close the editor handler is gone, so the content is read from disk.
	if err := os.WriteFile(uri.Path(), []byte("new content"), 0644); err != nil {
		t.Fatalf("write buffer file: %v", err)
	}

	s.Handle(context.Background(), textapi.Event{
		Type:     textapi.EventTypeClose,
		URI:      uri,
		Resource: &fakeResource{uri: uri},
	})

	var doc snippet
	if err := d.storage.Get(context.Background(), "fresh", &doc); err != nil {
		t.Fatalf("expected snippet persisted: %v", err)
	}
	if doc.Body != "new content" {
		t.Errorf("body = %q, want %q", doc.Body, "new content")
	}
	assertScratchGone(t, uri)
}

func TestCloseReadErrorDoesNotOverwrite(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "greet", "original")
	s := d.build()

	if err := s.HandleCommand(
		context.Background(), textapi.Command{Args: []string{"edit", "greet"}},
	); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	uri := d.opener.opened[0]
	if err := os.Remove(uri.Path()); err != nil {
		t.Fatalf("remove buffer file: %v", err)
	}

	s.Handle(context.Background(), textapi.Event{
		Type:     textapi.EventTypeClose,
		URI:      uri,
		Resource: &fakeResource{uri: uri},
	})

	var doc snippet
	if err := d.storage.Get(context.Background(), "greet", &doc); err != nil {
		t.Fatalf("get: %v", err)
	}
	if doc.Body != "original" {
		t.Errorf("body = %q, want unchanged %q", doc.Body, "original")
	}
	if !hasNotif(d.notif, browserapi.LevelError) {
		t.Errorf("expected an error notification, got %#v", d.notif.notifs)
	}
	assertScratchGone(t, uri)
}

func TestEditExistingSnippetOverwrittenOnFlush(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "greet", "old")
	s := d.build()

	if err := s.HandleCommand(
		context.Background(), textapi.Command{Args: []string{"edit", "greet"}},
	); err != nil {
		t.Fatalf("HandleCommand: %v", err)
	}

	uri := d.opener.opened[0]
	s.Handle(context.Background(), textapi.Event{
		Type:    textapi.EventTypeFlush,
		URI:     uri,
		Content: "updated",
	})

	var doc snippet
	if err := d.storage.Get(context.Background(), "greet", &doc); err != nil {
		t.Fatalf("get: %v", err)
	}
	if doc.Body != "updated" {
		t.Errorf("body = %q, want %q", doc.Body, "updated")
	}
}

func TestCompleteSubcommands(t *testing.T) {
	d := newDeps()
	s := d.build()

	it, err := s.Complete(context.Background(), "snippets", []string{"in"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	got, err := iterator.ToSlice(context.Background(), it)
	if err != nil {
		t.Fatalf("ToSlice: %v", err)
	}
	want := []string{"insert", "edit", "copy", "delete"}
	if len(got) != len(want) {
		t.Fatalf("subcommands = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("subcommands[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestCompleteNames(t *testing.T) {
	d := newDeps()
	seed(t, d.storage, "alpha", "a")
	seed(t, d.storage, "beta", "b")
	seed(t, d.storage, "gamma", "c")
	s := d.build()

	it, err := s.Complete(context.Background(), "snippets", []string{"insert", "al"})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	got, err := iterator.ToSlice(context.Background(), it)
	if err != nil {
		t.Fatalf("ToSlice: %v", err)
	}

	want := map[string]bool{"alpha": false, "beta": false, "gamma": false}
	for _, name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("unexpected name %q", name)
			continue
		}
		want[name] = true
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("missing name %q in %v", name, got)
		}
	}
}

type editSpy struct {
	start, end term.Coordinates
	str        string
	called     bool
}

type fakeCellEditor struct{ spy *editSpy }

func (e *fakeCellEditor) Edit(
	_ context.Context, start, end term.Coordinates, str string,
) (term.Coordinates, term.Coordinates, string, error) {
	e.spy.called = true
	e.spy.start = start
	e.spy.end = end
	e.spy.str = str
	return start, end, "", nil
}

type fakeCellView struct{}

func (v *fakeCellView) RawCells() ([][]term.Cell, error) { return nil, nil }

type fakeEditor struct {
	editSpy *editSpy
	view    *fakeCellView
}

func (e *fakeEditor) SubscribeEvents([]textapi.EventType, textapi.EventHandler) error {
	return nil
}
func (e *fakeEditor) Editor(workspaceapi.URI) (textapi.Handler, error) { return nil, nil }
func (e *fakeEditor) SetLocationList(
	textapi.Handler, textapi.LocationPriority, string, textapi.LocationList,
) error {
	return nil
}
func (e *fakeEditor) MoveToNextLocation(textapi.Handler, string) error { return nil }
func (e *fakeEditor) MoveToPrevLocation(textapi.Handler, string) error { return nil }
func (e *fakeEditor) Cursor(textapi.Handler) (term.Coordinates, error) {
	return term.Coordinates{}, nil
}
func (e *fakeEditor) SetCursor(textapi.Handler, term.Coordinates) error { return nil }
func (e *fakeEditor) CellView(textapi.Handler) textapi.CellView         { return e.view }
func (e *fakeEditor) CellEditor(textapi.Handler) textapi.CellEditor {
	return &fakeCellEditor{spy: e.editSpy}
}
func (e *fakeEditor) SetDefaultAttributes(textapi.Handler, term.Attributes) error { return nil }

// fakeFS resolves URIs and delegates file access to the real OS temp
// filesystem, since the handler writes/reads/removes scratch files there.
type fakeFS struct{}

func newFakeFS() *fakeFS { return &fakeFS{} }

func (fs *fakeFS) URI(path string) (workspaceapi.URI, error) {
	return workspaceapi.ParseURI("file://" + path)
}
func (fs *fakeFS) OpenFile(path string, flag int, mode os.FileMode) (workspaceapi.File, error) {
	return os.OpenFile(path, flag, mode)
}
func (fs *fakeFS) Remove(path string) error              { return os.Remove(path) }
func (fs *fakeFS) Stat(path string) (os.FileInfo, error) { return os.Stat(path) }
func (fs *fakeFS) ReadDir(string) ([]os.DirEntry, error) { return nil, nil }
func (fs *fakeFS) MkdirAll(string, os.FileMode) error    { return nil }

type fakeOpener struct {
	opened []workspaceapi.URI
	handle browserapi.Handler // handle returned by Open
}

func (o *fakeOpener) Open(uri workspaceapi.URI) (browserapi.Handler, error) {
	o.opened = append(o.opened, uri)
	if o.handle == nil {
		o.handle = &fakeResource{uri: uri}
	}
	return o.handle, nil
}

// fakeWindow is a minimal browserapi.Window.
type fakeWindow struct{ id uint64 }

func (w *fakeWindow) WindowID() uint64 { return w.id }

// setContentCall records a SetWindowContent invocation.
type setContentCall struct {
	win browserapi.Window
	h   browserapi.Handler
}

type fakeWM struct{ setContent []setContentCall }

func (m *fakeWM) Focus() (browserapi.Window, error) { return nil, nil }
func (m *fakeWM) Split(browserapi.Orientation, browserapi.Window, browserapi.Handler) (
	browserapi.Window, error,
) {
	return nil, nil
}
func (m *fakeWM) Floating(browserapi.Floating, browserapi.FloatingConfig) (browserapi.Window, error) {
	return nil, nil
}
func (m *fakeWM) Bar(browserapi.BarConfig, tui.Handler) error { return nil }
func (m *fakeWM) Tab(workspaceapi.URI, rune, string, browserapi.Handler) (browserapi.Handler, error) {
	return nil, nil
}
func (m *fakeWM) SetWindowContent(w browserapi.Window, h browserapi.Handler) error {
	m.setContent = append(m.setContent, setContentCall{win: w, h: h})
	return nil
}
func (m *fakeWM) CloseWindow(browserapi.Window) error { return nil }

type notification struct {
	level browserapi.NotificationLevel
	msg   string
}

type fakeNotif struct{ notifs []notification }

func (n *fakeNotif) Notify(
	level browserapi.NotificationLevel, msg string, args ...any,
) (string, error) {
	n.notifs = append(n.notifs, notification{level: level, msg: msg})
	return "", nil
}
func (n *fakeNotif) NotifyOnce(
	level browserapi.NotificationLevel, msg string, args ...any,
) (string, error) {
	return n.Notify(level, msg, args...)
}
func (n *fakeNotif) UpdateNotificationProgress(string, string, int64, int64) error {
	return nil
}

// fakeResource is a minimal textapi.Handler used as cmd.Resource / ev.Resource.
type fakeResource struct{ uri workspaceapi.URI }

func (r *fakeResource) Resource() workspaceapi.URI { return r.uri }
func (r *fakeResource) Close() error               { return nil }
func (r *fakeResource) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	return term.Coordinates{}, 0, false
}
func (r *fakeResource) Draw(term.Writer)               {}
func (r *fakeResource) Handle(term.Event) (bool, bool) { return false, false }
func (r *fakeResource) Resize(int, int)                {}
func (r *fakeResource) Selection() (string, bool)      { return "", false }

type deps struct {
	storage storageapi.Service
	editor  *fakeEditor
	fs      *fakeFS
	opener  *fakeOpener
	wm      *fakeWM
	notif   *fakeNotif
}

func newDeps() *deps {
	return &deps{
		storage: storagestub.NewInMemoryService(),
		editor:  &fakeEditor{editSpy: &editSpy{}, view: &fakeCellView{}},
		fs:      newFakeFS(),
		opener:  &fakeOpener{},
		wm:      &fakeWM{},
		notif:   &fakeNotif{},
	}
}

func (d *deps) build() *snippets {
	return newSnippets(d.storage, d.editor, d.fs, d.opener, d.wm, d.notif)
}

// readOpened reads the scratch file backing the single opened tab from disk.
func readOpened(t *testing.T, d *deps) (workspaceapi.URI, string) {
	t.Helper()
	if len(d.opener.opened) != 1 {
		t.Fatalf("expected one opened buffer, got %d", len(d.opener.opened))
	}
	uri := d.opener.opened[0]
	b, err := os.ReadFile(uri.Path())
	if err != nil {
		t.Fatalf("read scratch file %q: %v", uri.Path(), err)
	}
	return uri, string(b)
}

// assertScratchGone verifies the scratch file (and its temp dir) for the opened
// tab were removed from disk.
func assertScratchGone(t *testing.T, uri workspaceapi.URI) {
	t.Helper()
	if _, err := os.Stat(uri.Path()); !os.IsNotExist(err) {
		t.Errorf("expected scratch file %q removed, stat err = %v", uri.Path(), err)
	}
	if _, err := os.Stat(filepath.Dir(uri.Path())); !os.IsNotExist(err) {
		t.Errorf("expected scratch dir %q removed, stat err = %v", filepath.Dir(uri.Path()), err)
	}
}

func seed(t *testing.T, svc storageapi.Service, name, body string) {
	t.Helper()
	if err := svc.Set(context.Background(), name, snippet{Name: name, Body: body}); err != nil {
		t.Fatalf("seed %q: %v", name, err)
	}
}

func hasNotif(n *fakeNotif, level browserapi.NotificationLevel) bool {
	for _, notif := range n.notifs {
		if notif.level == level {
			return true
		}
	}
	return false
}
