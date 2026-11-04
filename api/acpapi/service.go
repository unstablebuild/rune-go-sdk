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

// Package acpapi exposes Agent Client Protocol (ACP) agents and their
// sessions. Sessions are observed through an append-only event log so any
// number of viewers can attach, replay what happened from the start and
// follow along live.
//
// The acp import is github.com/coder/acp-go-sdk; payloads are plain ACP v1
// types, so anything built on this package stays compatible with other ACP
// hosts.
package acpapi

import (
	"context"
	"errors"

	acp "github.com/coder/acp-go-sdk"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

var (
	// ErrNotFound is returned for unknown agents, sessions and permissions.
	ErrNotFound = errors.New("acp: not found")
	// ErrExists is returned by Service.NewSession when the name is taken.
	ErrExists = errors.New("acp: already exists")
	// ErrBusy is returned by Session.Prompt while a turn is running.
	ErrBusy = errors.New("acp: a turn is already running")
	// ErrResolved is returned by Session.Respond when the permission was
	// already answered or cancelled.
	ErrResolved = errors.New("acp: permission already resolved")
	// ErrClosed is returned by operations on a session that has ended.
	ErrClosed = errors.New("acp: session closed")
)

// Service is the workspace's configured ACP agents and their running
// sessions.
type Service interface {
	// Agents lists the configured agents.
	Agents(ctx context.Context) iterator.Iterator[AgentInfo]
	// NewSession starts req.Agent's process if it is not running, then opens a
	// session on it. An empty req.Name gets a unique petname; a taken name is
	// ErrExists. An unknown agent is ErrNotFound.
	NewSession(ctx context.Context, req NewSessionRequest) (Session, error)
	// Sessions lists the running sessions of every agent.
	Sessions(ctx context.Context) iterator.Iterator[SessionInfo]
	// OpenSession returns the running session called name, or ErrNotFound.
	OpenSession(ctx context.Context, name string) (Session, error)
}

// NewSessionRequest selects the agent and names the session.
type NewSessionRequest struct {
	Agent, Name string
	// Session is sent as session/new. Cwd defaults to the workspace root.
	Session acp.NewSessionRequest
}

// AgentInfo describes a configured agent.
type AgentInfo struct {
	Name string
	// Running reports whether the agent's process is up.
	Running bool
}

// SessionInfo identifies a session.
type SessionInfo struct {
	// Name is unique within the Service. Sessions opened directly on an
	// Agent are unnamed.
	Name  string
	Agent string
	ID    acp.SessionId
}

// Agent is one ACP connection seen from the client. Sessions share it.
type Agent interface {
	// Info is the agent's answer to initialize.
	Info() acp.InitializeResponse
	NewSession(ctx context.Context, req acp.NewSessionRequest) (Session, error)
	// LoadSession returns once the agent replayed the session; the replay is
	// the start of the session's events.
	LoadSession(ctx context.Context, req acp.LoadSessionRequest) (Session, error)
	ListSessions(ctx context.Context, req acp.ListSessionsRequest) iterator.Iterator[acp.SessionInfo]
	// Done is closed when the connection ends, which also ends every session.
	Done() <-chan struct{}
	// Err is why the connection ended: nil while it is up or after Close.
	Err() error
	// Close ends every session and the connection.
	Close() error
}

// Session is a conversation with an agent.
type Session interface {
	Info() SessionInfo
	// Events returns a new iterator over every event from Seq 0, then follows
	// live. Iterators are independent; closing one only detaches that viewer.
	// Next's ctx bounds the wait only. It ends when the session ends: Err is
	// nil after Close, otherwise the cause (e.g. the agent exited, with its
	// stderr tail).
	Events() iterator.Iterator[Event]
	// Prompt logs a Prompt event and returns; the turn ends with Done. ErrBusy
	// while a turn runs.
	Prompt(ctx context.Context, content []acp.ContentBlock) error
	// Cancel ends the running turn; pending permissions resolve Cancelled.
	Cancel(ctx context.Context) error
	// Respond answers a pending permission; first answer wins, later
	// ErrResolved.
	Respond(ctx context.Context, id PermissionID, o acp.RequestPermissionOutcome) error
	// SetConfigOption changes a session config option; the options the agent
	// returns are logged as a config_option_update.
	SetConfigOption(ctx context.Context, req acp.SetSessionConfigOptionRequest) error
	// Close ends the session for every viewer; the agent process keeps
	// running.
	Close() error
}
