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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// usage is the help string included in the error returned when the user issues
// an empty/unknown subcommand or omits the snippet name.
const usage = "Usage: snippets [insert|edit|copy|delete] <name>"

// snippet is the persisted document model. It is stored by ID = Name, and the
// store is auto-partitioned by ExtensionID. Name is duplicated inside the body
// so completion can list names reliably regardless of how List surfaces IDs.
type snippet struct {
	Name      string    `bson:"name"`
	Body      string    `bson:"body"`
	CreatedAt time.Time `bson:"CreatedAt"`
	UpdatedAt time.Time `bson:"UpdatedAt"`
}

// pendingEdit tracks a temporary edit buffer that, once flushed or closed,
// should be persisted back into storage under the given snippet name.
type pendingEdit struct {
	name string
	path string
	dir  string
}

// snippets implements textapi.CommandHandler and textapi.EventHandler. It lets
// users insert stored snippets into the editor and create/edit snippets in a
// scratch buffer that is persisted on flush/close.
type snippets struct {
	storage storageapi.Service
	editor  textapi.Editor
	fs      workspaceapi.FileSystem
	opener  browserapi.ResourceOpener
	wm      browserapi.WindowManager
	notif   browserapi.Notifications

	mu      sync.Mutex
	pending map[string]*pendingEdit // keyed by URI string
	lastSel string                  // last selection content
}

// newSnippets wires the snippets command handler with its dependencies. The
// dependencies are typed as SDK interfaces so tests can stub them.
func newSnippets(
	storage storageapi.Service,
	editor textapi.Editor,
	fs workspaceapi.FileSystem,
	opener browserapi.ResourceOpener,
	wm browserapi.WindowManager,
	notif browserapi.Notifications,
) *snippets {
	return &snippets{
		storage: storage,
		editor:  editor,
		fs:      fs,
		opener:  opener,
		wm:      wm,
		notif:   notif,
		pending: make(map[string]*pendingEdit),
	}
}

// HandleCommand dispatches on the first argument.
func (s *snippets) HandleCommand(ctx context.Context, cmd textapi.Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("missing subcommand: %s", usage)
	}

	sub := cmd.Args[0]
	name := ""
	if len(cmd.Args) > 1 {
		name = cmd.Args[1]
	}

	switch sub {
	case "insert":
		return s.handleInsert(ctx, name, cmd)
	case "edit":
		return s.handleEdit(ctx, name, cmd.Window)
	case "copy":
		return s.handleCopy(name, cmd.Window)
	case "delete":
		return s.handleDelete(ctx, name)
	default:
		return fmt.Errorf("unknown subcommand %q: %s", sub, usage)
	}
}

// Complete expands subcommands and snippet names.
func (s *snippets) Complete(
	ctx context.Context, cmd string, args []string,
) (iterator.Iterator[string], error) {
	if len(args) <= 1 {
		return iterator.FromSlice([]string{"insert", "edit", "copy", "delete"}), nil
	}

	it, err := s.storage.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	docs := iterator.FromDocumentIterator[snippet](it)
	return iterator.Map(docs, func(doc snippet) string { return doc.Name }), nil
}

// Handle tracks selections and persists pending edit buffers on flush/close.
func (s *snippets) Handle(ctx context.Context, ev textapi.Event) bool {
	switch ev.Type {
	case textapi.EventTypeSelection:
		s.mu.Lock()
		s.lastSel = ev.Content
		s.mu.Unlock()
	case textapi.EventTypeFlush:
		if err := s.persist(ctx, ev.URI, ev.Content, false); err != nil {
			// Handle cannot propagate errors to the host, so notify here.
			_, _ = s.notif.Notify(browserapi.LevelError, "snippets: %v", err)
		}
	case textapi.EventTypeClose:
		if err := s.persistOnClose(ctx, ev); err != nil {
			// Handle cannot propagate errors to the host, so notify here.
			_, _ = s.notif.Notify(browserapi.LevelError, "snippets: %v", err)
		}
	}
	return false
}

func (s *snippets) handleInsert(ctx context.Context, name string, cmd textapi.Command) error {
	if name == "" {
		return fmt.Errorf("insert requires a snippet name: %s", usage)
	}

	var doc snippet
	if err := s.storage.Get(ctx, name, &doc); err != nil {
		if errors.Is(err, storageapi.ErrNotFound) {
			return fmt.Errorf("snippet %q not found", name)
		}
		return fmt.Errorf("get snippet %q: %w", name, err)
	}

	if cmd.Resource == nil {
		return errors.New("no editor in focus to insert into")
	}

	at := cmd.Cursor.Content
	if _, _, _, err := s.editor.CellEditor(cmd.Resource).Edit(ctx, at, at, doc.Body); err != nil {
		return fmt.Errorf("insert snippet %q: %w", name, err)
	}
	_, err := s.notif.Notify(browserapi.LevelSuccess, "inserted snippet %q", name)
	return err
}

func (s *snippets) handleEdit(ctx context.Context, name string, win browserapi.Window) error {
	if name == "" {
		return fmt.Errorf("edit requires a snippet name: %s", usage)
	}

	seed := ""
	var doc snippet
	if err := s.storage.Get(ctx, name, &doc); err != nil {
		if !errors.Is(err, storageapi.ErrNotFound) {
			return fmt.Errorf("get snippet %q: %w", name, err)
		}
	} else {
		seed = doc.Body
	}
	return s.openEditBuffer(name, seed, win)
}

func (s *snippets) handleCopy(name string, win browserapi.Window) error {
	if name == "" {
		return fmt.Errorf("copy requires a snippet name: %s", usage)
	}
	s.mu.Lock()
	seed := s.lastSel
	s.mu.Unlock()
	return s.openEditBuffer(name, seed, win)
}

func (s *snippets) handleDelete(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("delete requires a snippet name: %s", usage)
	}
	if err := s.storage.Delete(ctx, name); err != nil {
		return fmt.Errorf("delete snippet %q: %w", name, err)
	}
	_, err := s.notif.Notify(browserapi.LevelSuccess, "deleted snippet %q", name)
	return err
}

// openEditBuffer creates a scratch file in the OS temp directory seeded with
// the given content, records it as pending, and opens it as a tab. A unique
// temp directory (os.MkdirTemp, cross-platform via os.TempDir) keeps the tab
// title readable as "<name>.snippet" while avoiding collisions between
// concurrent edits. If win is non-nil, the dispatching window is switched to
// the newly opened buffer.
func (s *snippets) openEditBuffer(name, seed string, win browserapi.Window) error {
	dir, err := os.MkdirTemp("", "snippets-")
	if err != nil {
		return fmt.Errorf("create edit buffer dir: %w", err)
	}
	path := filepath.Join(dir, name+".snippet")

	if err := os.WriteFile(path, []byte(seed), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("write edit buffer: %w", err)
	}

	uri, err := s.fs.URI(path)
	if err != nil {
		_ = os.RemoveAll(dir)
		return fmt.Errorf("resolve edit buffer uri: %w", err)
	}

	s.mu.Lock()
	s.pending[uri.String()] = &pendingEdit{name: name, path: path, dir: dir}
	s.mu.Unlock()

	h, err := s.opener.Open(uri)
	if err != nil {
		_ = s.dropPending(uri)
		return fmt.Errorf("open edit buffer tab: %w", err)
	}

	// Switch the window that dispatched the command to the newly opened
	// buffer. Open creates the tab but does not change focus on its own.
	if win != nil {
		if err := s.wm.SetWindowContent(win, h); err != nil {
			return fmt.Errorf("focus edit buffer: %w", err)
		}
	}

	_, err = s.notif.Notify(browserapi.LevelInfo, "editing snippet %q (save to persist)", name)
	return err
}

func (s *snippets) persistOnClose(ctx context.Context, ev textapi.Event) error {
	s.mu.Lock()
	pe, ok := s.pending[ev.URI.String()]
	s.mu.Unlock()
	if !ok {
		return nil
	}

	body, err := s.readFile(pe.path)
	if err != nil {
		// Leave the snippet untouched rather than overwriting it with an
		// empty body, but still clean up the scratch directory.
		_ = s.dropPending(ev.URI)
		return fmt.Errorf("read snippet %q on close: %w", pe.name, err)
	}
	return s.persist(ctx, ev.URI, body, true)
}

func (s *snippets) readFile(path string) (string, error) {
	f, err := s.fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *snippets) dropPending(uri workspaceapi.URI) error {
	s.mu.Lock()
	pe, ok := s.pending[uri.String()]
	if ok {
		delete(s.pending, uri.String())
	}
	s.mu.Unlock()
	if ok {
		return s.removeTempDir(pe)
	}
	return nil
}

func (s *snippets) persist(ctx context.Context, uri workspaceapi.URI, body string, closed bool) error {
	key := uri.String()

	s.mu.Lock()
	pe, ok := s.pending[key]
	if ok && closed {
		delete(s.pending, key)
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}

	now := time.Now()
	if err := s.storage.Set(ctx, pe.name, snippet{
		Name:      pe.name,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return fmt.Errorf("save snippet %q: %w", pe.name, err)
	}

	if closed {
		return s.removeTempDir(pe)
	}
	return nil
}

// removeTempDir deletes the scratch directory backing a pending edit.
func (s *snippets) removeTempDir(pe *pendingEdit) error {
	if err := os.RemoveAll(pe.dir); err != nil {
		return fmt.Errorf("remove temp dir for snippet %q: %w", pe.name, err)
	}
	return nil
}
