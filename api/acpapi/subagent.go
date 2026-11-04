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

package acpapi

import (
	"bytes"
	"encoding/json"

	acp "github.com/coder/acp-go-sdk"
)

// The payloads below are ACP's unstable Subagents RFD, which the acp
// import does not model yet. Agents only send them to clients that
// advertise the subagents client capability.

// Nullable is a field of an ACP upsert: omitted leaves the stored value
// unchanged, null clears it and a value replaces it. Marshal it with the
// omitzero option so an unset field stays omitted. A value that does not
// decode leaves the field unset, as the RFD asks receivers to tolerate it.
type Nullable[T any] struct {
	// Set reports whether the field was present.
	Set bool
	// Value is the field's value, or nil when it was null.
	Value *T
}

// Some is a Nullable set to v.
func Some[T any](v T) Nullable[T] { return Nullable[T]{Set: true, Value: &v} }

// Null is a Nullable that clears its field.
func Null[T any]() Nullable[T] { return Nullable[T]{Set: true} }

// IsZero reports whether n is unset, for omitzero.
func (n Nullable[T]) IsZero() bool { return !n.Set }

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*n.Value)
}

func (n *Nullable[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		*n = Null[T]()
		return nil
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		*n = Nullable[T]{}
		return nil
	}
	*n = Some(v)
	return nil
}

// patch is old with n applied.
func (n Nullable[T]) patch(old T) T {
	switch {
	case !n.Set:
		return old
	case n.Value == nil:
		var zero T
		return zero
	default:
		return *n.Value
	}
}

// patchPtr is old with n applied, for values stored by pointer.
func (n Nullable[T]) patchPtr(old *T) *T {
	if !n.Set {
		return old
	}
	return n.Value
}

// SubagentUpdate announces a subagent, a session that the session it
// arrives on created and owns, or patches what is known of it. It is
// ACP's subagent_update. The first update for a SessionID announces it;
// later ones never change its parent.
type SubagentUpdate struct {
	SessionID   acp.SessionId    `json:"sessionId"`
	Title       Nullable[string] `json:"title,omitzero"`
	Description Nullable[string] `json:"description,omitzero"`
	// Capabilities replaces the controls the client may use on the
	// subagent as a whole. Null disables them all.
	Capabilities Nullable[SubagentCapabilities] `json:"capabilities,omitzero"`
	// State replaces the subagent's work state snapshot. Null leaves it
	// unconfirmed, which is not idle.
	State Nullable[SubagentState] `json:"state,omitzero"`
	// Cwd and AdditionalDirectories report where the agent placed the
	// subagent. While unset, each is its parent's.
	Cwd                   Nullable[string]         `json:"cwd,omitzero"`
	AdditionalDirectories Nullable[[]string]       `json:"additionalDirectories,omitzero"`
	Meta                  Nullable[map[string]any] `json:"_meta,omitzero"`
}

// SubagentCapabilities are the controls the client may use on a subagent.
// Prompting a subagent is never allowed.
type SubagentCapabilities struct {
	// Cancel is set when the client may stop the subagent's current work by
	// sending session/cancel with its session ID.
	Cancel *Capability     `json:"cancel,omitempty"`
	Meta   map[string]any `json:"_meta,omitempty"`
}

// Capability marks a supported control. Its presence is the support.
type Capability struct {
	Meta map[string]any `json:"_meta,omitempty"`
}

// SubagentWorkState is what a subagent's foreground work is doing. Values
// other than the constants are custom or future states, none of which
// means the work ended.
type SubagentWorkState string

const (
	SubagentRunning        SubagentWorkState = "running"
	SubagentIdle           SubagentWorkState = "idle"
	SubagentRequiresAction SubagentWorkState = "requires_action"
	// SubagentUnknown means the agent lost track of the work, which may
	// still be running.
	SubagentUnknown SubagentWorkState = "unknown"
)

// StopReasonError is the stop reason of an idle subagent whose work
// failed.
const StopReasonError acp.StopReason = "error"

// SubagentState is a snapshot of a subagent's foreground work. Idle does
// not end the subagent: its parent may give it more work.
type SubagentState struct {
	State SubagentWorkState `json:"state"`
	// StopReason, Error and Usage are only reported when idle. Error is
	// the failure when StopReason is StopReasonError.
	StopReason acp.StopReason    `json:"stopReason,omitempty"`
	Error      *acp.RequestError `json:"error,omitempty"`
	Usage      *acp.Usage        `json:"usage,omitempty"`
	Meta       map[string]any    `json:"_meta,omitempty"`
}

// SessionMessage is an upsert of a message between two sessions, as it
// appears in the transcript of the session it arrives on: that session
// sent it when it is the Sender and received it when it is the Recipient.
// It is ACP's session_message. MessageID is only unique within that
// transcript; SenderMessageID and RecipientMessageID link the two views.
type SessionMessage struct {
	MessageID          string                       `json:"messageId"`
	Sender             Nullable[acp.SessionId]      `json:"senderSessionId,omitzero"`
	Recipient          Nullable[acp.SessionId]      `json:"recipientSessionId,omitzero"`
	SenderMessageID    Nullable[string]             `json:"senderMessageId,omitzero"`
	RecipientMessageID Nullable[string]             `json:"recipientMessageId,omitzero"`
	// Content replaces the message's content. Null or empty clears it.
	Content Nullable[[]acp.ContentBlock] `json:"content,omitzero"`
	Meta    Nullable[map[string]any]     `json:"_meta,omitzero"`
}

// SessionMessageChunk appends Content to the SessionMessage MessageID. It
// may arrive before the message's first upsert. It is ACP's
// session_message_chunk.
type SessionMessageChunk struct {
	MessageID string           `json:"messageId"`
	Content   acp.ContentBlock `json:"content"`
	Meta      map[string]any   `json:"_meta,omitempty"`
}

// Subagent is what a session knows of one of its subagents.
type Subagent struct {
	// Parent is the subagent that owns it, or empty when the session
	// itself does.
	Parent      acp.SessionId
	Title       string
	Description string
	// Cwd and AdditionalDirectories are empty while inherited from the
	// parent. A non-nil empty AdditionalDirectories reports none.
	Cwd                   string
	AdditionalDirectories []string
	Capabilities          *SubagentCapabilities
	// State is the last reported work state, or nil while unconfirmed.
	State *SubagentState
}

// apply is s with u applied.
func (s Subagent) apply(u *SubagentUpdate) Subagent {
	s.Title = u.Title.patch(s.Title)
	s.Description = u.Description.patch(s.Description)
	s.Cwd = u.Cwd.patch(s.Cwd)
	s.AdditionalDirectories = u.AdditionalDirectories.patch(s.AdditionalDirectories)
	s.Capabilities = u.Capabilities.patchPtr(s.Capabilities)
	s.State = u.State.patchPtr(s.State)
	return s
}
