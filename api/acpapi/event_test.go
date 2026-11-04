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

package acpapi_test

import (
	"encoding/json"
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
)

func modelOption(current string) acp.SessionConfigOption {
	opt := acp.NewSessionConfigOptionSelect(acp.SessionConfigValueId(current), acp.SessionConfigSelectOptions{
		Ungrouped: &acp.SessionConfigSelectOptionsUngrouped{
			{Name: "Small", Value: "small"},
			{Name: "Large", Value: "large"},
		},
	})
	opt.Select.Id = "model"
	opt.Select.Name = "Model"
	opt.Select.Category = acp.Ptr(acp.SessionConfigOptionCategoryModel)
	return opt
}

func TestApply(t *testing.T) {
	modes := &acp.SessionModeState{
		AvailableModes: []acp.SessionMode{{Id: "ask", Name: "Ask"}, {Id: "code", Name: "Code"}},
		CurrentModeId:  "ask",
	}
	perm := func(id acpapi.PermissionID) acpapi.Event {
		return acpapi.Event{Permission: &acpapi.Permission{ID: id}}
	}
	running := acpapi.SubagentState{State: acpapi.SubagentRunning}
	idle := acpapi.SubagentState{State: acpapi.SubagentIdle, StopReason: acp.StopReasonEndTurn}
	cancel := &acpapi.SubagentCapabilities{Cancel: &acpapi.Capability{}}
	cases := []struct {
		name   string
		events []acpapi.Event
		want   acpapi.SessionState
	}{
		{
			name:   "started",
			events: []acpapi.Event{{Started: &acpapi.Started{ConfigOptions: []acp.SessionConfigOption{modelOption("small")}, Modes: modes}}},
			want:   acpapi.SessionState{ConfigOptions: []acp.SessionConfigOption{modelOption("small")}, Modes: modes},
		},
		{
			name: "config option update replaces options",
			events: []acpapi.Event{
				{Started: &acpapi.Started{ConfigOptions: []acp.SessionConfigOption{modelOption("small")}}},
				{Update: &acp.SessionUpdate{ConfigOptionUpdate: &acp.SessionConfigOptionUpdate{
					ConfigOptions: []acp.SessionConfigOption{modelOption("large")},
				}}},
			},
			want: acpapi.SessionState{ConfigOptions: []acp.SessionConfigOption{modelOption("large")}},
		},
		{
			name: "current mode update",
			events: []acpapi.Event{
				{Started: &acpapi.Started{Modes: modes}},
				{Update: &acp.SessionUpdate{CurrentModeUpdate: &acp.SessionCurrentModeUpdate{CurrentModeId: "code"}}},
			},
			want: acpapi.SessionState{Modes: &acp.SessionModeState{AvailableModes: modes.AvailableModes, CurrentModeId: "code"}},
		},
		{
			name:   "current mode update without modes",
			events: []acpapi.Event{{Update: &acp.SessionUpdate{CurrentModeUpdate: &acp.SessionCurrentModeUpdate{CurrentModeId: "code"}}}},
			want:   acpapi.SessionState{},
		},
		{
			name: "commands plan usage",
			events: []acpapi.Event{
				{Update: &acp.SessionUpdate{AvailableCommandsUpdate: &acp.SessionAvailableCommandsUpdate{
					AvailableCommands: []acp.AvailableCommand{{Name: "review", Description: "Review"}},
				}}},
				{Update: &acp.SessionUpdate{Plan: &acp.SessionUpdatePlan{
					Entries: []acp.PlanEntry{{Content: "step", Status: acp.PlanEntryStatusPending}},
				}}},
				{Update: &acp.SessionUpdate{UsageUpdate: &acp.SessionUsageUpdate{Size: 100, Used: 10}}},
			},
			want: acpapi.SessionState{
				Commands: []acp.AvailableCommand{{Name: "review", Description: "Review"}},
				Plan:     []acp.PlanEntry{{Content: "step", Status: acp.PlanEntryStatusPending}},
				Usage:    &acp.SessionUsageUpdate{Size: 100, Used: 10},
			},
		},
		{
			name:   "prompt is busy",
			events: []acpapi.Event{{Prompt: &acpapi.Prompt{}}},
			want:   acpapi.SessionState{Busy: true},
		},
		{
			name:   "permissions pending and resolved",
			events: []acpapi.Event{{Prompt: &acpapi.Prompt{}}, perm(1), perm(2), {Resolved: &acpapi.Resolution{ID: 1}}},
			want:   acpapi.SessionState{Busy: true, Pending: []acpapi.Permission{{ID: 2}}},
		},
		{
			name:   "last permission resolved",
			events: []acpapi.Event{perm(0), {Resolved: &acpapi.Resolution{ID: 0}}},
			want:   acpapi.SessionState{},
		},
		{
			name:   "done clears busy and pending",
			events: []acpapi.Event{{Prompt: &acpapi.Prompt{}}, perm(1), {Done: &acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn}}},
			want:   acpapi.SessionState{},
		},
		{
			name:   "message chunks do not change state",
			events: []acpapi.Event{{Update: &acp.SessionUpdate{AgentMessageChunk: &acp.SessionUpdateAgentMessageChunk{Content: acp.TextBlock("hi")}}}},
			want:   acpapi.SessionState{},
		},
		{
			name: "subagent announced then patched",
			events: []acpapi.Event{
				{Subagent: &acpapi.SubagentUpdate{
					SessionID: "a", Title: acpapi.Some("Explore"), Description: acpapi.Some("general-purpose"),
					Capabilities: acpapi.Some(*cancel), State: acpapi.Some(running), Cwd: acpapi.Some("/w"),
				}},
				{Subagent: &acpapi.SubagentUpdate{
					SessionID: "a", Title: acpapi.Null[string](), State: acpapi.Some(idle),
					AdditionalDirectories: acpapi.Some([]string{}),
				}},
			},
			want: acpapi.SessionState{Subagents: map[acp.SessionId]acpapi.Subagent{
				"a": {
					Description: "general-purpose", Cwd: "/w", AdditionalDirectories: []string{},
					Capabilities: cancel, State: &idle,
				},
			}},
		},
		{
			name: "subagent state and capabilities cleared",
			events: []acpapi.Event{
				{Subagent: &acpapi.SubagentUpdate{SessionID: "a", Capabilities: acpapi.Some(*cancel), State: acpapi.Some(running)}},
				{Subagent: &acpapi.SubagentUpdate{
					SessionID: "a", Capabilities: acpapi.Null[acpapi.SubagentCapabilities](),
					State: acpapi.Null[acpapi.SubagentState](),
				}},
			},
			want: acpapi.SessionState{Subagents: map[acp.SessionId]acpapi.Subagent{"a": {}}},
		},
		{
			name: "nested subagent keeps its first parent",
			events: []acpapi.Event{
				{Subagent: &acpapi.SubagentUpdate{SessionID: "a"}},
				{Session: "a", Subagent: &acpapi.SubagentUpdate{SessionID: "b"}},
				{Subagent: &acpapi.SubagentUpdate{SessionID: "b", Title: acpapi.Some("Inner")}},
			},
			want: acpapi.SessionState{Subagents: map[acp.SessionId]acpapi.Subagent{
				"a": {},
				"b": {Parent: "a", Title: "Inner"},
			}},
		},
		{
			name: "subagent updates leave the session's state alone",
			events: []acpapi.Event{
				{Subagent: &acpapi.SubagentUpdate{SessionID: "a"}},
				{Session: "a", Update: &acp.SessionUpdate{UsageUpdate: &acp.SessionUsageUpdate{Size: 100, Used: 10}}},
				{Session: "a", Update: &acp.SessionUpdate{ConfigOptionUpdate: &acp.SessionConfigOptionUpdate{
					ConfigOptions: []acp.SessionConfigOption{modelOption("large")},
				}}},
			},
			want: acpapi.SessionState{Subagents: map[acp.SessionId]acpapi.Subagent{"a": {}}},
		},
		{
			name: "done keeps subagent permissions",
			events: []acpapi.Event{
				{Prompt: &acpapi.Prompt{}}, perm(1),
				{Permission: &acpapi.Permission{ID: 2, Session: "a"}},
				{Done: &acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn}},
			},
			want: acpapi.SessionState{Pending: []acpapi.Permission{{ID: 2, Session: "a"}}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s acpapi.SessionState
			for _, ev := range tc.events {
				acpapi.Apply(&s, ev)
			}
			assert.Equal(t, tc.want, s)
		})
	}
}

func TestApplyDoesNotMutatePriorCopies(t *testing.T) {
	var s acpapi.SessionState
	acpapi.Apply(&s, acpapi.Event{Permission: &acpapi.Permission{ID: 1}})
	acpapi.Apply(&s, acpapi.Event{Permission: &acpapi.Permission{ID: 2}})
	before := s
	acpapi.Apply(&s, acpapi.Event{Resolved: &acpapi.Resolution{ID: 1}})
	acpapi.Apply(&s, acpapi.Event{Permission: &acpapi.Permission{ID: 3}})
	assert.Equal(t, []acpapi.Permission{{ID: 1}, {ID: 2}}, before.Pending)
	assert.Equal(t, []acpapi.Permission{{ID: 2}, {ID: 3}}, s.Pending)

	acpapi.Apply(&s, acpapi.Event{Subagent: &acpapi.SubagentUpdate{SessionID: "a", Title: acpapi.Some("A")}})
	before = s
	acpapi.Apply(&s, acpapi.Event{Subagent: &acpapi.SubagentUpdate{SessionID: "a", Title: acpapi.Some("B")}})
	acpapi.Apply(&s, acpapi.Event{Subagent: &acpapi.SubagentUpdate{SessionID: "b"}})
	assert.Equal(t, map[acp.SessionId]acpapi.Subagent{"a": {Title: "A"}}, before.Subagents)
	assert.Equal(t, map[acp.SessionId]acpapi.Subagent{"a": {Title: "B"}, "b": {}}, s.Subagents)
}

func TestEventsJSONRoundTrip(t *testing.T) {
	l := acpapi.NewLog()
	l.Started(acpapi.Started{ConfigOptions: []acp.SessionConfigOption{modelOption("small")}})
	require.NoError(t, l.Prompt([]acp.ContentBlock{acp.TextBlock("fix it"), acp.ResourceLinkBlock("main.go", "file:///main.go")}))
	l.Update(acp.UpdateAgentThoughtText("thinking"))
	l.Update(acp.StartToolCall("t1", "Read main.go", acp.WithStartKind(acp.ToolKindRead)))
	go func() {
		_, _ = l.Ask(t.Context(), acp.RequestPermissionRequest{
			SessionId: "s",
			ToolCall:  acp.ToolCallUpdate{ToolCallId: "t1"},
			Options:   []acp.PermissionOption{{Kind: acp.PermissionOptionKindAllowOnce, Name: "Allow", OptionId: "allow"}},
		})
	}()
	waitPending(t, l, 1)
	require.NoError(t, l.Respond(l.State().Pending[0].ID, acp.NewRequestPermissionOutcomeSelected("allow")))
	l.Update(acp.UpdateToolCall("t1", acp.WithUpdateStatus(acp.ToolCallStatusCompleted)))
	l.Subagent("", acpapi.SubagentUpdate{
		SessionID: "a", Title: acpapi.Some("Explore"), Description: acpapi.Null[string](),
		Capabilities: acpapi.Some(acpapi.SubagentCapabilities{Cancel: &acpapi.Capability{}}),
		State:        acpapi.Some(acpapi.SubagentState{State: acpapi.SubagentRunning}),
	})
	l.Message("", acpapi.SessionMessage{
		MessageID: "m1", Sender: acpapi.Some[acp.SessionId]("s"), Recipient: acpapi.Some[acp.SessionId]("a"),
		RecipientMessageID: acpapi.Some("m1"), Content: acpapi.Some([]acp.ContentBlock{acp.TextBlock("look")}),
	})
	l.MessageChunk("a", acpapi.SessionMessageChunk{MessageID: "m2", Content: acp.TextBlock("found")})
	l.UpdateSubagent("a", acp.StartToolCall("t2", "Run ls", acp.WithStartKind(acp.ToolKindExecute)))
	l.Subagent("", acpapi.SubagentUpdate{SessionID: "a", State: acpapi.Some(acpapi.SubagentState{
		State: acpapi.SubagentIdle, StopReason: acpapi.StopReasonError,
		Error: &acp.RequestError{Code: -32000, Message: "boom"},
	})})
	l.Update(acp.UpdateAgentMessageText("done"))
	l.Done(acpapi.TurnEnd{StopReason: acp.StopReasonEndTurn})
	l.Close(nil)
	events := drain(t, l.Events())

	b, err := json.Marshal(events)
	require.NoError(t, err)
	var got []acpapi.Event
	require.NoError(t, json.Unmarshal(b, &got))
	again, err := json.Marshal(got)
	require.NoError(t, err)
	assert.JSONEq(t, string(b), string(again))

	var want, replayed acpapi.SessionState
	for i := range events {
		acpapi.Apply(&want, events[i])
		acpapi.Apply(&replayed, got[i])
	}
	assert.Equal(t, want, replayed)
	assert.Equal(t, l.State(), want)
}
