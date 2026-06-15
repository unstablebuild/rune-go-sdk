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

package llmrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
)

// schemaMarshaler is a test json.Marshaler used to feed a
// ResponseFormatJSONSchema through the wire.
type schemaMarshaler struct{ payload string }

func (s schemaMarshaler) MarshalJSON() ([]byte, error) { return []byte(s.payload), nil }

func TestMessageReasoningBlocksProtoRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  llmapi.Message
	}{
		{
			name: "thinking block with signature",
			msg: llmapi.Message{
				Role:    llmapi.RoleAssistant,
				Content: "answer",
				ReasoningBlocks: []llmapi.ReasoningBlock{
					{Kind: "thinking", Text: "reasoning", Signature: "sig=="},
				},
			},
		},
		{
			name: "redacted block",
			msg: llmapi.Message{
				Role: llmapi.RoleAssistant,
				ReasoningBlocks: []llmapi.ReasoningBlock{
					{Kind: "redacted", Data: "encrypted"},
				},
			},
		},
		{
			name: "interleaved blocks preserve order with tool calls",
			msg: llmapi.Message{
				Role:    llmapi.RoleAssistant,
				Content: "done",
				ReasoningBlocks: []llmapi.ReasoningBlock{
					{Kind: "thinking", Text: "first", Signature: "s1"},
					{Kind: "redacted", Data: "blob"},
					{Kind: "thinking", Text: "second", Signature: "s2"},
				},
				ToolCalls: []llmapi.ToolCall{
					{ID: "c1", Type: llmapi.ToolTypeFunction, Function: llmapi.FunctionCall{Name: "f", Arguments: "{}"}},
				},
			},
		},
		{
			name: "no reasoning blocks stays nil",
			msg: llmapi.Message{
				Role:    llmapi.RoleAssistant,
				Content: "plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fromProtoMessage(ToProtoMessage(tt.msg))
			assert.Equal(t, tt.msg.ReasoningBlocks, got.ReasoningBlocks)
			assert.Equal(t, tt.msg, got)
		})
	}
}

func TestCompletionHeaderRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		model llmapi.ModelEntry
		req   llmapi.Request
	}{
		{
			name:  "empty request header only",
			model: llmapi.ModelEntry{Name: "m", Provider: "openai", ContextWindow: 128_000},
			req:   llmapi.Request{},
		},
		{
			name:  "config fields without messages",
			model: llmapi.ModelEntry{Name: "gpt-5", Provider: "openai", BaseURL: "https://api.example"},
			req: llmapi.Request{
				ReasoningEffort:  llmapi.ReasoningEffortHigh,
				ReasoningSummary: llmapi.ReasoningSummaryDetailed,
				MaxOutputTokens:  256,
				PromptCacheKey:   "dlg-1",
				TokenCount:       42,
			},
		},
		{
			name:  "tools and response format",
			model: llmapi.ModelEntry{Name: "m"},
			req: llmapi.Request{
				Tools: []llmapi.Tool{{
					Type: llmapi.ToolTypeFunction,
					Function: llmapi.FunctionDefinition{
						Name:        "search",
						Description: "search the web",
						Parameters:  map[string]any{"type": "object"},
					},
				}},
				ResponseFormat: &llmapi.ResponseFormat{
					Type: llmapi.ResponseFormatTypeJSONSchema,
					JSONSchema: &llmapi.ResponseFormatJSONSchema{
						Name:        "Out",
						Description: "output",
						Schema:      schemaMarshaler{payload: `{"type":"object"}`},
						Strict:      true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := ToCompletionHeader(tt.model, tt.req)
			require.NoError(t, err)
			gotModel, gotReq, err := RequestFromHeaderAndMessages(header, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.model, gotModel)

			// FromProtoRequest decodes Parameters/Schema as raw JSON; compare via
			// a reference reconstruction of the same proto round-trip.
			preq, err := ToProtoRequest(tt.req)
			require.NoError(t, err)
			preq.Messages = nil
			wantReq, err := FromProtoRequest(preq)
			require.NoError(t, err)
			assert.Equal(t, wantReq, gotReq)
		})
	}
}

func TestRequestFromHeaderAndMessagesPreservesOrder(t *testing.T) {
	model := llmapi.ModelEntry{Name: "m"}
	msgs := []llmapi.Message{
		{Role: llmapi.RoleUser, Content: "one"},
		{Role: llmapi.RoleAssistant, Content: "two"},
		{Role: llmapi.RoleUser, Content: "three"},
	}
	header, err := ToCompletionHeader(model, llmapi.Request{})
	require.NoError(t, err)
	pmsgs := make([]*Message, len(msgs))
	for i := range msgs {
		pmsgs[i] = ToProtoMessage(msgs[i])
	}
	gotModel, gotReq, err := RequestFromHeaderAndMessages(header, pmsgs)
	require.NoError(t, err)
	assert.Equal(t, model, gotModel)
	require.Len(t, gotReq.Messages, len(msgs))
	for i := range msgs {
		assert.Equal(t, msgs[i].Content, gotReq.Messages[i].Content)
		assert.Equal(t, msgs[i].Role, gotReq.Messages[i].Role)
	}
}

func TestRequestFromHeaderAndMessagesNilHeader(t *testing.T) {
	_, _, err := RequestFromHeaderAndMessages(nil, nil)
	require.Error(t, err)
}
