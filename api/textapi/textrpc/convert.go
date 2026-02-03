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

package textrpc

import (
	"fmt"

	"github.com/unstablebuild/rune-go-sdk/api/textapi"
)

func protoTypeToModel(protoType EditorEvent_Type) (ev textapi.EventType, err error) {
	switch protoType {
	case EditorEvent_TypeClose:
		ev = textapi.EventTypeClose
	case EditorEvent_TypeFlush:
		ev = textapi.EventTypeFlush
	case EditorEvent_TypeOpen:
		ev = textapi.EventTypeOpen
	case EditorEvent_TypeEdit:
		ev = textapi.EventTypeEdit
	case EditorEvent_TypeScroll:
		ev = textapi.EventTypeScroll
	case EditorEvent_TypeHidden:
		ev = textapi.EventTypeHidden
	case EditorEvent_TypeVisible:
		ev = textapi.EventTypeVisible
	case EditorEvent_TypeCursor:
		ev = textapi.EventTypeCursor
	case EditorEvent_TypeSelection:
		ev = textapi.EventTypeSelection
	case EditorEvent_TypeCreate:
		ev = textapi.EventTypeCreate
	case EditorEvent_TypeChange:
		ev = textapi.EventTypeChange
	case EditorEvent_TypeFocus:
		ev = textapi.EventTypeFocus
	case EditorEvent_TypeUnfocus:
		ev = textapi.EventTypeUnfocus
	case EditorEvent_TypeRemove:
		ev = textapi.EventTypeRemove
	case EditorEvent_TypeRename:
		ev = textapi.EventTypeRename
	default:
		err = fmt.Errorf("failed to convert proto editor event: invalid type: %v",
			protoType)
	}

	return
}

func fromProto(e *textapi.Event, pe *EditorEvent) (err error) {
	e.Type, err = protoTypeToModel(pe.GetType())
	if err != nil {
		return
	}
	if pe.GetResourceName().GetUri() != "" {
		e.URI, err = NewURIFromProto(pe.GetResourceName())
		if err != nil {
			return
		}
	}
	if pe.ResourceName != nil {
		uri, err := NewURIFromProto(pe.GetResourceName())
		if err != nil {
			return err
		}
		e.Resource = Token{
			URI: uri,
		}
	}
	e.Start = pe.GetStart().ToModel()
	e.End = pe.GetEnd().ToModel()
	e.From = pe.GetFrom().ToModel()
	e.To = pe.GetTo().ToModel()
	e.Content = pe.GetContent()
	return nil
}

func protoType(e textapi.Event) EditorEvent_Type {
	switch e.Type {
	case textapi.EventTypeClose:
		return EditorEvent_TypeClose
	case textapi.EventTypeFlush:
		return EditorEvent_TypeFlush
	case textapi.EventTypeOpen:
		return EditorEvent_TypeOpen
	case textapi.EventTypeEdit:
		return EditorEvent_TypeEdit
	case textapi.EventTypeScroll:
		return EditorEvent_TypeScroll
	case textapi.EventTypeHidden:
		return EditorEvent_TypeHidden
	case textapi.EventTypeVisible:
		return EditorEvent_TypeVisible
	case textapi.EventTypeChange:
		return EditorEvent_TypeChange
	case textapi.EventTypeCreate:
		return EditorEvent_TypeCreate
	case textapi.EventTypeCursor:
		return EditorEvent_TypeCursor
	case textapi.EventTypeSelection:
		return EditorEvent_TypeSelection
	case textapi.EventTypeFocus:
		return EditorEvent_TypeFocus
	case textapi.EventTypeUnfocus:
		return EditorEvent_TypeUnfocus
	case textapi.EventTypeRename:
		return EditorEvent_TypeRename
	case textapi.EventTypeRemove:
		return EditorEvent_TypeRemove
	default:
		panic(fmt.Sprintf("failed to convert editor event to proto: invalid type: %v", e.Type))
	}
}
