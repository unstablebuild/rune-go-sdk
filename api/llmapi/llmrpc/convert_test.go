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
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
)

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
