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

package browserrpc

import handlerrpc "github.com/unstablebuild/rune-go-sdk/handler/handlerrpc"

/* boilerplate so we can re-use handlerrpc stream implementation */

// SetDraw satisfies handlerrpc.StreamMessage.
func (m *FloatingWindowMessage) SetDraw(r *handlerrpc.DrawStreamResponse) {
	m.Draw = r
	m.Type = handlerrpc.MessageType_Draw
}

// SetHandle satisfies handlerrpc.StreamMessage.
func (m *FloatingWindowMessage) SetHandle(r *handlerrpc.HandleStreamResponse) {
	m.Handle = r
	m.Type = handlerrpc.MessageType_Handle
}

// SetClose satisfies handlerrpc.StreamMessage.
func (m *FloatingWindowMessage) SetClose(r *handlerrpc.CloseStreamResponse) {
	m.Close = r
	m.Type = handlerrpc.MessageType_Close
}

// SetCursor satisfies handlerrpc.StreamMessage.
func (m *FloatingWindowMessage) SetCursor(r *handlerrpc.CursorStreamResponse) {
	m.Cursor = r
	m.Type = handlerrpc.MessageType_Cursor
}

// SetSelection satisfies handlerrpc.StreamMessage.
func (m *FloatingWindowMessage) SetSelection(r *handlerrpc.SelectionStreamResponse) {
	m.Selection = r
	m.Type = handlerrpc.MessageType_Selection
}

// SetDimensions satisfies handlerrpc.StreamMessage.
func (m *FloatingWindowMessage) SetDimensions(r *handlerrpc.DimensionsStreamResponse) {
	m.Dimensions = r
	m.Type = handlerrpc.MessageType_Dimensions
}

// SetDraw satisfies handlerrpc.StreamMessage.
func (m *SplitWindowMessage) SetDraw(r *handlerrpc.DrawStreamResponse) {
	m.Draw = r
	m.Type = handlerrpc.MessageType_Draw
}

// SetHandle satisfies handlerrpc.StreamMessage.
func (m *SplitWindowMessage) SetHandle(r *handlerrpc.HandleStreamResponse) {
	m.Handle = r
	m.Type = handlerrpc.MessageType_Handle
}

// SetClose satisfies handlerrpc.StreamMessage.
func (m *SplitWindowMessage) SetClose(r *handlerrpc.CloseStreamResponse) {
	m.Close = r
	m.Type = handlerrpc.MessageType_Close
}

// SetCursor satisfies handlerrpc.StreamMessage.
func (m *SplitWindowMessage) SetCursor(r *handlerrpc.CursorStreamResponse) {
	m.Cursor = r
	m.Type = handlerrpc.MessageType_Cursor
}

// SetSelection satisfies handlerrpc.StreamMessage.
func (m *SplitWindowMessage) SetSelection(r *handlerrpc.SelectionStreamResponse) {
	m.Selection = r
	m.Type = handlerrpc.MessageType_Selection
}

// SetDimensions satisfies handlerrpc.StreamMessage.
func (m *SplitWindowMessage) SetDimensions(r *handlerrpc.DimensionsStreamResponse) {
	m.Dimensions = r
	m.Type = handlerrpc.MessageType_Dimensions
}

// SetDraw satisfies handlerrpc.StreamMessage.
func (m *BarMessage) SetDraw(r *handlerrpc.DrawStreamResponse) {
	m.Draw = r
	m.Type = handlerrpc.MessageType_Draw
}

// SetHandle satisfies handlerrpc.StreamMessage.
func (m *BarMessage) SetHandle(r *handlerrpc.HandleStreamResponse) {
	m.Handle = r
	m.Type = handlerrpc.MessageType_Handle
}

// SetClose satisfies handlerrpc.StreamMessage.
func (m *BarMessage) SetClose(r *handlerrpc.CloseStreamResponse) {
	m.Close = r
	m.Type = handlerrpc.MessageType_Close
}

// SetCursor satisfies handlerrpc.StreamMessage.
func (m *BarMessage) SetCursor(r *handlerrpc.CursorStreamResponse) {
	m.Cursor = r
	m.Type = handlerrpc.MessageType_Cursor
}

// SetSelection satisfies handlerrpc.StreamMessage.
func (m *BarMessage) SetSelection(r *handlerrpc.SelectionStreamResponse) {
	m.Selection = r
	m.Type = handlerrpc.MessageType_Selection
}

// SetDimensions satisfies handlerrpc.StreamMessage.
func (m *BarMessage) SetDimensions(r *handlerrpc.DimensionsStreamResponse) {
	m.Dimensions = r
	m.Type = handlerrpc.MessageType_Dimensions
}

// SetDraw satisfies handlerrpc.StreamMessage.
func (m *WindowSetContentMessage) SetDraw(r *handlerrpc.DrawStreamResponse) {
	m.Draw = r
	m.Type = handlerrpc.MessageType_Draw
}

// SetHandle satisfies handlerrpc.StreamMessage.
func (m *WindowSetContentMessage) SetHandle(r *handlerrpc.HandleStreamResponse) {
	m.Handle = r
	m.Type = handlerrpc.MessageType_Handle
}

// SetClose satisfies handlerrpc.StreamMessage.
func (m *WindowSetContentMessage) SetClose(r *handlerrpc.CloseStreamResponse) {
	m.Close = r
	m.Type = handlerrpc.MessageType_Close
}

// SetCursor satisfies handlerrpc.StreamMessage.
func (m *WindowSetContentMessage) SetCursor(r *handlerrpc.CursorStreamResponse) {
	m.Cursor = r
	m.Type = handlerrpc.MessageType_Cursor
}

// SetSelection satisfies handlerrpc.StreamMessage.
func (m *WindowSetContentMessage) SetSelection(r *handlerrpc.SelectionStreamResponse) {
	m.Selection = r
	m.Type = handlerrpc.MessageType_Selection
}

// SetDimensions satisfies handlerrpc.StreamMessage.
func (m *WindowSetContentMessage) SetDimensions(r *handlerrpc.DimensionsStreamResponse) {
	m.Dimensions = r
	m.Type = handlerrpc.MessageType_Dimensions
}

// SetDraw satisfies handlerrpc.StreamMessage.
func (m *TabMessage) SetDraw(r *handlerrpc.DrawStreamResponse) {
	m.Draw = r
	m.Type = handlerrpc.MessageType_Draw
}

// SetHandle satisfies handlerrpc.StreamMessage.
func (m *TabMessage) SetHandle(r *handlerrpc.HandleStreamResponse) {
	m.Handle = r
	m.Type = handlerrpc.MessageType_Handle
}

// SetClose satisfies handlerrpc.StreamMessage.
func (m *TabMessage) SetClose(r *handlerrpc.CloseStreamResponse) {
	m.Close = r
	m.Type = handlerrpc.MessageType_Close
}

// SetCursor satisfies handlerrpc.StreamMessage.
func (m *TabMessage) SetCursor(r *handlerrpc.CursorStreamResponse) {
	m.Cursor = r
	m.Type = handlerrpc.MessageType_Cursor
}

// SetSelection satisfies handlerrpc.StreamMessage.
func (m *TabMessage) SetSelection(r *handlerrpc.SelectionStreamResponse) {
	m.Selection = r
	m.Type = handlerrpc.MessageType_Selection
}

// SetDimensions satisfies handlerrpc.StreamMessage.
func (m *TabMessage) SetDimensions(r *handlerrpc.DimensionsStreamResponse) {
	m.Dimensions = r
	m.Type = handlerrpc.MessageType_Dimensions
}
