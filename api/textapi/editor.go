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

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// Handler just wraps a tui.Handler to indicate that this API's handlers might
// not be compatible with other APIs.
type Handler interface {
	browserapi.Handler

	// This is only used to differentiate editor.Handler from the rest
	// of tui.Handler in a browser.Component.
	Resource() workspaceapi.URI
}

// CellEditor is a cell.Editor that can fail.
type CellEditor interface {
	Edit(ctx context.Context, start, end term.Coordinates, str string) (
		from, to term.Coordinates, old string, err error,
	)
}

// CellView wraps a subset of cell.View behaviour with an API that can fail.
type CellView interface {
	RawCells() ([][]term.Cell, error)
}

// LocationPriority defines the order of which location attributes and
// messages are processed.
type LocationPriority uint

const (
	// LocationPriorityInfo defines an informational location.
	LocationPriorityInfo LocationPriority = iota
	// LocationPriorityWarning defines a warning location.
	LocationPriorityWarning
	// LocationPriorityError defines an error location.
	LocationPriorityError
	// LocationPriorityCritical defines an critical location.
	LocationPriorityCritical
)

// Editor is the interface that wraps an API to manage a text editor.
type Editor interface {
	// SubscribeEvents subscribes EventHandler to events of type EventType.
	SubscribeEvents([]EventType, EventHandler) error

	// Editor returns the editor.Handler that manages the given resource,
	// if there is a resource currently open with the given URI.
	Editor(resource workspaceapi.URI) (Handler, error)

	// SetLocationList sets the Handler's location list for users to
	// navigate the code. See LocationList for more details.
	// In order to remove a location list, SetLocationList must be called
	// with an empty (or nil) LocationList.
	// Locations are removed if underlying buffer is updated. It is the
	// responsibility of the caller to recompute the list of locations
	// and call SetLocationList with the new list of locations after
	// every update. Check cell.Buffer.Subscribe for more details.
	SetLocationList(Handler, LocationPriority, string, LocationList) error

	// Moves cursor to the next location on list with ID.
	MoveToNextLocation(h Handler, ID string) error

	// Moves cursor to the previous location on list with ID.
	MoveToPrevLocation(h Handler, ID string) error

	// Cursor gets the position of Handler's cursor in the underlying
	// content buffer.
	Cursor(Handler) (term.Coordinates, error)

	// SetCursor sets the cursor of Handler to the given Coordinates.
	SetCursor(Handler, term.Coordinates) error

	// CellView returns a CellView which allows to read the editor's internal buffer.
	CellView(Handler) CellView

	// CellEditor returns a CellEditor which allows for direct write access
	// to the editor's internal buffer.
	CellEditor(Handler) CellEditor

	// SetDefaultAttributes sets the default attributes of the given Handler
	// before any LocationList overwrites.
	SetDefaultAttributes(Handler, term.Attributes) error
}
