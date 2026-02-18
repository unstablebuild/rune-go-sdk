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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package fileexplorer

import (
	"fmt"

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// LocationSet holds a location list with its metadata.
type LocationSet struct {
	Priority textapi.LocationPriority
	ID       string
	List     textapi.LocationList
}

// EditorHandler wraps a text editor handler. It extends
// browserapi.Handler and component.Scrollable with editor
// specific operations.
type EditorHandler interface {
	browserapi.Handler
	component.Scrollable

	// Resource returns the URI of the underlying
	// resource.
	Resource() workspaceapi.URI

	// SetWrap sets whether lines longer than available
	// width wrap into the next line.
	SetWrap(wrap bool)

	// ShowCommandBar hides or shows the command bar.
	ShowCommandBar(show bool)

	// SetCursorAtScroll sets the cursor at the given
	// scroll coordinates.
	SetCursorAtScroll(term.Coordinates) bool

	// CursorAtScroll returns the cursor position in
	// scroll coordinates.
	CursorAtScroll() term.Coordinates

	// SetLocationList sets a location list by ID.
	SetLocationList(
		textapi.LocationPriority,
		string,
		textapi.LocationList,
	)

	// LocationLists returns all location lists.
	LocationLists() []LocationSet

	// MoveToNextLocation moves cursor to the next
	// location on list with ID.
	MoveToNextLocation(string) bool

	// MoveToPrevLocation moves cursor to the previous
	// location on list with ID.
	MoveToPrevLocation(string) bool

	// CellView returns a CellView for reading the
	// editor's internal buffer.
	CellView() textapi.CellView

	// CellEditor returns a CellEditor for writing to
	// the editor's internal buffer.
	CellEditor() textapi.CellEditor

	// SetDefaultAttributes sets the default attributes
	// before any LocationList overwrites.
	SetDefaultAttributes(term.Attributes)
}

// Editor is an interface that wraps an API to manage a
// text editor.
type Editor interface {
	// SubscribeEvents subscribes to events of given
	// types.
	SubscribeEvents(
		[]textapi.EventType, textapi.EventHandler,
	) error

	// UnsubscribeEvents unsubscribes the given handler.
	UnsubscribeEvents(
		textapi.EventHandler,
	) (bool, error)

	// Editor returns the EditorHandler for the given
	// URI.
	Editor(workspaceapi.URI) (EditorHandler, error)

	// SubscribeCommand registers a command handler.
	SubscribeCommand(
		textapi.CommandManual, textapi.CommandHandler,
	) error

	// UnsubscribeCommand un-registers a command.
	UnsubscribeCommand(string) error
}

var _ textapi.Editor = (*editorAdapter)(nil)

// NewEditorAdapter adapts an Editor to textapi.Editor.
// Handler-level operations (SetLocationList, CellView,
// etc.) are delegated by type-asserting the
// textapi.Handler back to EditorHandler.
func NewEditorAdapter(e Editor) textapi.Editor {
	return &editorAdapter{e: e}
}

type editorAdapter struct {
	e Editor
}

func (a *editorAdapter) SubscribeEvents(
	types []textapi.EventType,
	h textapi.EventHandler,
) error {
	return a.e.SubscribeEvents(types, h)
}

func (a *editorAdapter) Editor(
	uri workspaceapi.URI,
) (textapi.Handler, error) {
	return a.e.Editor(uri)
}

func (a *editorAdapter) SetLocationList(
	h textapi.Handler,
	p textapi.LocationPriority,
	id string,
	l textapi.LocationList,
) error {
	eh, err := asEditorHandler(h)
	if err != nil {
		return err
	}
	eh.SetLocationList(p, id, l)
	return nil
}

func (a *editorAdapter) MoveToNextLocation(
	h textapi.Handler, id string,
) error {
	eh, err := asEditorHandler(h)
	if err != nil {
		return err
	}
	if !eh.MoveToNextLocation(id) {
		return fmt.Errorf(
			"no next location for list %q", id,
		)
	}
	return nil
}

func (a *editorAdapter) MoveToPrevLocation(
	h textapi.Handler, id string,
) error {
	eh, err := asEditorHandler(h)
	if err != nil {
		return err
	}
	if !eh.MoveToPrevLocation(id) {
		return fmt.Errorf(
			"no previous location for list %q",
			id,
		)
	}
	return nil
}

func (a *editorAdapter) Cursor(
	h textapi.Handler,
) (term.Coordinates, error) {
	eh, err := asEditorHandler(h)
	if err != nil {
		return term.Coordinates{}, err
	}
	return eh.CursorAtScroll(), nil
}

func (a *editorAdapter) SetCursor(
	h textapi.Handler, pos term.Coordinates,
) error {
	eh, err := asEditorHandler(h)
	if err != nil {
		return err
	}
	if !eh.SetCursorAtScroll(pos) {
		return fmt.Errorf(
			"could not set cursor at (%d, %d)",
			pos.X, pos.Y,
		)
	}
	return nil
}

func (a *editorAdapter) CellView(
	h textapi.Handler,
) textapi.CellView {
	eh, ok := h.(EditorHandler)
	if !ok {
		return nil
	}
	return eh.CellView()
}

func (a *editorAdapter) CellEditor(
	h textapi.Handler,
) textapi.CellEditor {
	eh, ok := h.(EditorHandler)
	if !ok {
		return nil
	}
	return eh.CellEditor()
}

func (a *editorAdapter) SetDefaultAttributes(
	h textapi.Handler, attr term.Attributes,
) error {
	eh, err := asEditorHandler(h)
	if err != nil {
		return err
	}
	eh.SetDefaultAttributes(attr)
	return nil
}

func asEditorHandler(
	h textapi.Handler,
) (EditorHandler, error) {
	eh, ok := h.(EditorHandler)
	if !ok {
		return nil, fmt.Errorf(
			"handler does not implement EditorHandler",
		)
	}
	return eh, nil
}
