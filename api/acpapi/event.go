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
	"maps"
	"slices"

	acp "github.com/coder/acp-go-sdk"
)

// PermissionID identifies a permission request within a session. It is the
// Seq of its Permission event.
type PermissionID int

// Event is one entry of a session's log. It is plain data: a []Event marshals
// to JSON and replays into the same transcript. Exactly one field besides Seq
// and Session is set. Viewers share payloads and must treat them as
// read-only.
type Event struct {
	// Seq is the event's index in the log, starting at 0.
	Seq int `json:"seq"`
	// Session is the subagent an Update, Subagent, Message or MessageChunk
	// arrived on, or empty when it arrived on the session itself. A
	// subagent's Update is its own, not the session's.
	Session      acp.SessionId        `json:"session,omitempty"`
	Started      *Started             `json:"started,omitempty"`
	Update       *acp.SessionUpdate   `json:"update,omitempty"`
	Subagent     *SubagentUpdate      `json:"subagent,omitempty"`
	Message      *SessionMessage      `json:"message,omitempty"`
	MessageChunk *SessionMessageChunk `json:"messageChunk,omitempty"`
	Prompt       *Prompt              `json:"prompt,omitempty"`
	Permission   *Permission          `json:"permission,omitempty"`
	Resolved     *Resolution          `json:"resolved,omitempty"`
	Done         *TurnEnd             `json:"done,omitempty"`
}

// Started carries the session state returned by session/new or session/load.
type Started struct {
	ConfigOptions []acp.SessionConfigOption `json:"configOptions,omitempty"`
	Modes         *acp.SessionModeState     `json:"modes,omitempty"`
}

// Prompt is the user message that began a turn.
type Prompt struct {
	Content []acp.ContentBlock `json:"content"`
}

// Permission is a pending session/request_permission.
type Permission struct {
	ID PermissionID `json:"id"`
	// Session is the subagent that asked, or empty when the session itself
	// did.
	Session acp.SessionId                `json:"session,omitempty"`
	Request acp.RequestPermissionRequest `json:"request"`
}

// Resolution is the answer to a Permission. Outcome is Cancelled when the turn
// ended, the session closed or the agent withdrew the request.
type Resolution struct {
	ID      PermissionID                 `json:"id"`
	Outcome acp.RequestPermissionOutcome `json:"outcome"`
}

// TurnEnd ends the turn begun by the last Prompt. Error is set, and
// StopReason may be empty, when the turn failed.
type TurnEnd struct {
	StopReason acp.StopReason `json:"stopReason,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// SessionState is the session state folded from its events by Apply.
type SessionState struct {
	ConfigOptions []acp.SessionConfigOption
	Modes         *acp.SessionModeState
	Commands      []acp.AvailableCommand
	Plan          []acp.PlanEntry
	Usage         *acp.SessionUsageUpdate
	// Pending are the unresolved permissions, the subagents' included,
	// oldest first.
	Pending []Permission
	// Busy reports whether a turn is running. Subagents may work outside
	// turns.
	Busy bool
	// Subagents are the announced subagents, nested ones included, by
	// session ID.
	Subagents map[acp.SessionId]Subagent
}

// Apply folds ev into s. It never mutates slices or values previously stored
// in s, so copies of s taken before Apply stay valid.
func Apply(s *SessionState, ev Event) {
	switch {
	case ev.Started != nil:
		s.ConfigOptions = ev.Started.ConfigOptions
		s.Modes = ev.Started.Modes
	case ev.Update != nil && ev.Session == "":
		applyUpdate(s, ev.Update)
	case ev.Subagent != nil:
		subs := maps.Clone(s.Subagents)
		if subs == nil {
			subs = make(map[acp.SessionId]Subagent)
		}
		sub, known := subs[ev.Subagent.SessionID]
		if !known {
			sub.Parent = ev.Session
		}
		subs[ev.Subagent.SessionID] = sub.apply(ev.Subagent)
		s.Subagents = subs
	case ev.Prompt != nil:
		s.Busy = true
	case ev.Permission != nil:
		s.Pending = append(slices.Clip(s.Pending), *ev.Permission)
	case ev.Resolved != nil:
		id := ev.Resolved.ID
		s.Pending = dropPending(s.Pending, func(p Permission) bool { return p.ID == id })
	case ev.Done != nil:
		s.Busy = false
		// Subagents keep working, and asking, after the turn.
		s.Pending = dropPending(s.Pending, func(p Permission) bool { return p.Session == "" })
	}
}

func dropPending(pending []Permission, drop func(Permission) bool) []Permission {
	pending = slices.DeleteFunc(slices.Clone(pending), drop)
	if len(pending) == 0 {
		return nil
	}
	return pending
}

func applyUpdate(s *SessionState, u *acp.SessionUpdate) {
	switch {
	case u.ConfigOptionUpdate != nil:
		s.ConfigOptions = u.ConfigOptionUpdate.ConfigOptions
	case u.CurrentModeUpdate != nil:
		if s.Modes != nil {
			m := *s.Modes
			m.CurrentModeId = u.CurrentModeUpdate.CurrentModeId
			s.Modes = &m
		}
	case u.AvailableCommandsUpdate != nil:
		s.Commands = u.AvailableCommandsUpdate.AvailableCommands
	case u.Plan != nil:
		s.Plan = u.Plan.Entries
	case u.UsageUpdate != nil:
		s.Usage = u.UsageUpdate
	}
}
