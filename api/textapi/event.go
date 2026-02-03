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


package textapi

import (
	"context"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// EventType is a type of editor event.
type EventType uint8

const (
	// EventTypeOpen is dispatched when an editor is called the Edit method.
	// Content represents the initial content of the underlying resource.
	EventTypeOpen EventType = iota

	// EventTypeClose is dispatched when an editor buffer is closed.
	EventTypeClose

	// EventTypeFlush is dispatched when an editor buffer is Flushed.
	// Content represents the file content that was flushed.
	EventTypeFlush

	// EventTypeCreate is dispatched when a watched resource is created out-of-band
	// or not in the workspace.
	EventTypeCreate

	// EventTypeChange is dispatched when a watched resource is updated out-of-band
	// or not in the workspace.
	EventTypeChange

	// EventTypeRemove is dispatched when a watched resource is removed out-of-band
	// or not in the workspace.
	EventTypeRemove

	// EventTypeRename is dispatched when a watched resource is renamed out-of-band
	// or not in the workspace.
	EventTypeRename

	// EventTypeEdit is dispatched when new content is inserted into an editor buffer.
	// Start, End represent the input to Edit whereas
	// From, To represent output coordinates. See cell.Editor.Edit for
	// more details.
	EventTypeEdit

	// EventTypeScroll is dispatched when content is scrolled to a new offset.
	// Start represents the new scroll offset, whereas From represents the
	// last offset position.
	EventTypeScroll

	// EventTypeHidden is dispatched when some content is hidden.
	// Start represents the start of the hidden block of lines, whereas End
	// represents the end of the hidden block.
	EventTypeHidden

	// EventTypeVisible is dispatched when some content is hidden.
	// Start represents the start of the hidden block of lines, whereas End
	// represents the end of the hidden block.
	EventTypeVisible

	// EventTypeFocus is dispatched when an editor handler is on browser.Focus.
	// Start contains the width (X) and height (Y) of the content in focus.
	// From contains the position of the cursor at that content focus.
	// If content is resized, EventTypeFocus is sent again, with the new dimensions.
	EventTypeFocus

	// EventTypeUnfocus is dispatched when an editor handler is not
	// on browser.Focus anymore.
	EventTypeUnfocus

	// EventTypeCursor is dispatched when the position of the cursor of an
	// editor Handler changes, either in the window coordinate system or the underlying
	// content position. Event.Start will be set to the cursor's window position,
	// and Event.From will be set to the cursor's scroll position.
	EventTypeCursor

	// EventTypeSelection is dispatched when the user selects a chunk of text.
	// Start, End represent the selection coordinates and Content contains
	// the content selected as a result.
	EventTypeSelection
)

// String implements fmt.Stringer interface.
func (e EventType) String() string {
	switch e {
	case EventTypeOpen:
		return "open"
	case EventTypeClose:
		return "close"
	case EventTypeFlush:
		return "flush"
	case EventTypeCreate:
		return "create"
	case EventTypeChange:
		return "change"
	case EventTypeRemove:
		return "remove"
	case EventTypeRename:
		return "rename"
	case EventTypeEdit:
		return "edit"
	case EventTypeScroll:
		return "scroll"
	case EventTypeFocus:
		return "focus"
	case EventTypeUnfocus:
		return "unfocus"
	case EventTypeCursor:
		return "cursor"
	case EventTypeSelection:
		return "selection"
	case EventTypeHidden:
		return "hidden"
	case EventTypeVisible:
		return "visible"
	default:
		panic("uknown event")
	}
}

// AllEvents returns all types of textapi.EventType in a slice.
func AllEvents() []EventType {
	return []EventType{
		EventTypeOpen, EventTypeClose, EventTypeFlush,
		EventTypeCreate, EventTypeChange, EventTypeRemove,
		EventTypeRename, EventTypeEdit, EventTypeScroll,
		EventTypeFocus, EventTypeUnfocus, EventTypeCursor,
		EventTypeSelection, EventTypeHidden, EventTypeVisible,
	}
}

// Event encapsulates eventual information about a particular editor resource.
type Event struct {
	Type     EventType
	URI      workspaceapi.URI
	Resource Handler

	Start, End term.Coordinates
	From, To   term.Coordinates
	Content    string
}

// EventHandler wraps the basic method Handle.
type EventHandler interface {
	// Handle handles Event and returns true if it no longer needs to receive events,
	// in other words it returns true if it's done processing events.
	Handle(context.Context, Event) bool
}

type fnEventHandler struct {
	cb func(context.Context, Event) bool
}

func (f fnEventHandler) Handle(ctx context.Context, ev Event) bool {
	return f.cb(ctx, ev)
}

// FuncEventHandler returns an EventHandler that calls fn
// every time Handle is invoked.
func FuncEventHandler(fn func(context.Context, Event) bool) EventHandler {
	return fnEventHandler{
		cb: fn,
	}
}
