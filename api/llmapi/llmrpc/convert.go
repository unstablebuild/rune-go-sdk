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
	"encoding/json"
	"errors"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToProtoModelEntry converts an llmapi.ModelEntry to its wire form.
func ToProtoModelEntry(m llmapi.ModelEntry) *ModelEntry {
	return &ModelEntry{
		Name:          m.Name,
		Provider:      m.Provider,
		ContextWindow: int32(m.ContextWindow),
		BaseUrl:       m.BaseURL,
	}
}

// FromProtoModelEntry converts a wire ModelEntry to llmapi.ModelEntry.
func FromProtoModelEntry(p *ModelEntry) llmapi.ModelEntry {
	if p == nil {
		return llmapi.ModelEntry{}
	}
	return llmapi.ModelEntry{
		Name:          p.GetName(),
		Provider:      p.GetProvider(),
		ContextWindow: int(p.GetContextWindow()),
		BaseURL:       p.GetBaseUrl(),
	}
}

// ToProtoRequest converts an llmapi.Request to its proto wire form.
func ToProtoRequest(req llmapi.Request) (*Request, error) {
	out := &Request{
		ReasoningEffort:  string(req.ReasoningEffort),
		ReasoningSummary: string(req.ReasoningSummary),
		MaxOutputTokens:  int32(req.MaxOutputTokens),
		PromptCacheKey:   req.PromptCacheKey,
		TokenCount:       int32(req.TokenCount),
	}
	out.Messages = make([]*Message, len(req.Messages))
	for i, m := range req.Messages {
		out.Messages[i] = ToProtoMessage(m)
	}
	if len(req.Tools) > 0 {
		out.Tools = make([]*Tool, len(req.Tools))
		for i, t := range req.Tools {
			pt, err := toProtoTool(t)
			if err != nil {
				return nil, err
			}
			out.Tools[i] = pt
		}
	}
	if req.ResponseFormat != nil {
		rf, err := toProtoResponseFormat(req.ResponseFormat)
		if err != nil {
			return nil, err
		}
		out.ResponseFormat = rf
		out.HasResponseFormat = true
	}
	return out, nil
}

// FromProtoRequest converts a wire Request to llmapi.Request.
func FromProtoRequest(p *Request) (llmapi.Request, error) {
	if p == nil {
		return llmapi.Request{}, nil
	}
	out := llmapi.Request{
		ReasoningEffort:  llmapi.ReasoningEffort(p.GetReasoningEffort()),
		ReasoningSummary: llmapi.ReasoningSummary(p.GetReasoningSummary()),
		MaxOutputTokens:  int(p.GetMaxOutputTokens()),
		PromptCacheKey:   p.GetPromptCacheKey(),
		TokenCount:       int(p.GetTokenCount()),
	}
	if msgs := p.GetMessages(); len(msgs) > 0 {
		out.Messages = make([]llmapi.Message, len(msgs))
		for i, m := range msgs {
			out.Messages[i] = fromProtoMessage(m)
		}
	}
	if tools := p.GetTools(); len(tools) > 0 {
		out.Tools = make([]llmapi.Tool, len(tools))
		for i, t := range tools {
			tt, err := fromProtoTool(t)
			if err != nil {
				return llmapi.Request{}, err
			}
			out.Tools[i] = tt
		}
	}
	if p.GetHasResponseFormat() {
		rf, err := fromProtoResponseFormat(p.GetResponseFormat())
		if err != nil {
			return llmapi.Request{}, err
		}
		out.ResponseFormat = rf
	}
	return out, nil
}

// ToProtoMessage converts an llmapi.Message to its proto wire form.
func ToProtoMessage(m llmapi.Message) *Message {
	out := &Message{
		Role:             toProtoRole(m.Role),
		Content:          m.Content,
		ReasoningContent: m.ReasoningContent,
		ToolCallId:       m.ToolCallID,
		Name:             m.Name,
	}
	if len(m.MultiContent) > 0 {
		out.MultiContent = make([]*ContentPart, len(m.MultiContent))
		for i, cp := range m.MultiContent {
			out.MultiContent[i] = &ContentPart{
				Type:     toProtoContentPartType(cp.Type),
				Text:     cp.Text,
				ImageUrl: cp.ImageURL,
			}
		}
	}
	if len(m.ToolCalls) > 0 {
		out.ToolCalls = make([]*ToolCall, len(m.ToolCalls))
		for i, tc := range m.ToolCalls {
			out.ToolCalls[i] = toProtoToolCall(tc)
		}
	}
	return out
}

func fromProtoMessage(p *Message) llmapi.Message {
	if p == nil {
		return llmapi.Message{}
	}
	m := llmapi.Message{
		Role:             fromProtoRole(p.GetRole()),
		Content:          p.GetContent(),
		ReasoningContent: p.GetReasoningContent(),
		ToolCallID:       p.GetToolCallId(),
		Name:             p.GetName(),
	}
	if parts := p.GetMultiContent(); len(parts) > 0 {
		m.MultiContent = make([]llmapi.ContentPart, len(parts))
		for i, cp := range parts {
			m.MultiContent[i] = llmapi.ContentPart{
				Type:     fromProtoContentPartType(cp.GetType()),
				Text:     cp.GetText(),
				ImageURL: cp.GetImageUrl(),
			}
		}
	}
	if calls := p.GetToolCalls(); len(calls) > 0 {
		m.ToolCalls = make([]llmapi.ToolCall, len(calls))
		for i, tc := range calls {
			m.ToolCalls[i] = fromProtoToolCall(tc)
		}
	}
	return m
}

func toProtoToolCall(tc llmapi.ToolCall) *ToolCall {
	return &ToolCall{
		Id:   tc.ID,
		Type: toProtoToolType(tc.Type),
		Function: &FunctionCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}

func fromProtoToolCall(p *ToolCall) llmapi.ToolCall {
	if p == nil {
		return llmapi.ToolCall{}
	}
	return llmapi.ToolCall{
		ID:   p.GetId(),
		Type: fromProtoToolType(p.GetType()),
		Function: llmapi.FunctionCall{
			Name:      p.GetFunction().GetName(),
			Arguments: p.GetFunction().GetArguments(),
		},
	}
}

func toProtoTool(t llmapi.Tool) (*Tool, error) {
	def := &FunctionDefinition{
		Name:        t.Function.Name,
		Description: t.Function.Description,
	}
	if t.Function.Parameters != nil {
		switch v := t.Function.Parameters.(type) {
		case []byte:
			def.Parameters = v
		case json.RawMessage:
			def.Parameters = []byte(v)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			def.Parameters = b
		}
	}
	return &Tool{Type: toProtoToolType(t.Type), Function: def}, nil
}

func fromProtoTool(p *Tool) (llmapi.Tool, error) {
	if p == nil {
		return llmapi.Tool{}, nil
	}
	out := llmapi.Tool{
		Type: fromProtoToolType(p.GetType()),
		Function: llmapi.FunctionDefinition{
			Name:        p.GetFunction().GetName(),
			Description: p.GetFunction().GetDescription(),
		},
	}
	if b := p.GetFunction().GetParameters(); len(b) > 0 {
		// Preserve as raw JSON bytes; consumers can decode as needed.
		out.Function.Parameters = json.RawMessage(b)
	}
	return out, nil
}

func toProtoResponseFormat(rf *llmapi.ResponseFormat) (*ResponseFormat, error) {
	out := &ResponseFormat{Type: string(rf.Type)}
	if rf.JSONSchema != nil {
		var schemaBytes []byte
		if rf.JSONSchema.Schema != nil {
			b, err := rf.JSONSchema.Schema.MarshalJSON()
			if err != nil {
				return nil, err
			}
			schemaBytes = b
		}
		out.JsonSchema = &ResponseFormatJSONSchema{
			Name:        rf.JSONSchema.Name,
			Description: rf.JSONSchema.Description,
			Schema:      schemaBytes,
			Strict:      rf.JSONSchema.Strict,
		}
		out.HasJsonSchema = true
	}
	return out, nil
}

func fromProtoResponseFormat(p *ResponseFormat) (*llmapi.ResponseFormat, error) {
	if p == nil {
		return nil, nil
	}
	out := &llmapi.ResponseFormat{Type: llmapi.ResponseFormatType(p.GetType())}
	if p.GetHasJsonSchema() {
		js := p.GetJsonSchema()
		out.JSONSchema = &llmapi.ResponseFormatJSONSchema{
			Name:        js.GetName(),
			Description: js.GetDescription(),
			Strict:      js.GetStrict(),
		}
		if b := js.GetSchema(); len(b) > 0 {
			out.JSONSchema.Schema = rawJSON(b)
		}
	}
	return out, nil
}

// rawJSON wraps a []byte payload so it satisfies json.Marshaler with the
// original bytes preserved.
type rawJSON []byte

// MarshalJSON satisfies json.Marshaler.
func (r rawJSON) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return []byte(r), nil
}

// ToProtoEvent converts an llmapi.Event to its wire form.
func ToProtoEvent(ev llmapi.Event) *Event {
	out := &Event{
		Type:      toProtoEventType(ev.Type),
		Text:      ev.Text,
		Reasoning: ev.Reasoning,
	}
	if ev.ToolCall != nil {
		out.ToolCall = toProtoToolCall(*ev.ToolCall)
	}
	if ev.DoneData != nil {
		out.DoneData = &DoneData{
			Message:      ToProtoMessage(ev.DoneData.Message),
			FinishReason: string(ev.DoneData.FinishReason),
			Usage:        toProtoUsage(ev.DoneData.Usage),
		}
	}
	if ev.Error != nil {
		out.Error = ev.Error.Error()
	}
	if ev.RateLimit != nil {
		out.RateLimit = &RateLimitInfo{
			WaitDurationNs: int64(ev.RateLimit.WaitDuration),
			Message:        ev.RateLimit.Message,
		}
	}
	return out
}

// FromProtoEvent converts a wire Event to llmapi.Event.
func FromProtoEvent(p *Event) llmapi.Event {
	if p == nil {
		return llmapi.Event{}
	}
	out := llmapi.Event{
		Type:      fromProtoEventType(p.GetType()),
		Text:      p.GetText(),
		Reasoning: p.GetReasoning(),
	}
	if tc := p.GetToolCall(); tc != nil {
		conv := fromProtoToolCall(tc)
		out.ToolCall = &conv
	}
	if dd := p.GetDoneData(); dd != nil {
		out.DoneData = &llmapi.DoneData{
			Message:      fromProtoMessage(dd.GetMessage()),
			FinishReason: llmapi.FinishReason(dd.GetFinishReason()),
			Usage:        fromProtoUsage(dd.GetUsage()),
		}
	}
	if errStr := p.GetError(); errStr != "" {
		out.Error = errors.New(errStr)
	}
	if rl := p.GetRateLimit(); rl != nil {
		out.RateLimit = &llmapi.RateLimitInfo{
			WaitDuration: time.Duration(rl.GetWaitDurationNs()),
			Message:      rl.GetMessage(),
		}
	}
	return out
}

func toProtoUsage(u llmapi.Usage) *Usage {
	return &Usage{
		TokensSent:         int32(u.TokensSent),
		TokensReceived:     int32(u.TokensReceived),
		TokensReasoned:     int32(u.TokensReasoned),
		TokensCached:       int32(u.TokensCached),
		TokensCacheCreated: int32(u.TokensCacheCreated),
	}
}

func fromProtoUsage(p *Usage) llmapi.Usage {
	if p == nil {
		return llmapi.Usage{}
	}
	return llmapi.Usage{
		TokensSent:         int(p.GetTokensSent()),
		TokensReceived:     int(p.GetTokensReceived()),
		TokensReasoned:     int(p.GetTokensReasoned()),
		TokensCached:       int(p.GetTokensCached()),
		TokensCacheCreated: int(p.GetTokensCacheCreated()),
	}
}

func toProtoRole(r llmapi.Role) Role {
	switch r {
	case llmapi.RoleAssistant:
		return Role_ROLE_ASSISTANT
	case llmapi.RoleUser:
		return Role_ROLE_USER
	case llmapi.RoleSystem:
		return Role_ROLE_SYSTEM
	case llmapi.RoleTool:
		return Role_ROLE_TOOL
	}
	return Role_ROLE_UNSPECIFIED
}

func fromProtoRole(r Role) llmapi.Role {
	switch r {
	case Role_ROLE_ASSISTANT:
		return llmapi.RoleAssistant
	case Role_ROLE_USER:
		return llmapi.RoleUser
	case Role_ROLE_SYSTEM:
		return llmapi.RoleSystem
	case Role_ROLE_TOOL:
		return llmapi.RoleTool
	}
	return ""
}

func toProtoToolType(t llmapi.ToolType) ToolType {
	if t == llmapi.ToolTypeFunction {
		return ToolType_TOOL_TYPE_FUNCTION
	}
	return ToolType_TOOL_TYPE_UNSPECIFIED
}

func fromProtoToolType(t ToolType) llmapi.ToolType {
	if t == ToolType_TOOL_TYPE_FUNCTION {
		return llmapi.ToolTypeFunction
	}
	return ""
}

func toProtoContentPartType(t llmapi.ContentPartType) ContentPartType {
	if t == llmapi.ContentPartTypeImageURL {
		return ContentPartType_CONTENT_PART_TYPE_IMAGE_URL
	}
	return ContentPartType_CONTENT_PART_TYPE_TEXT
}

func fromProtoContentPartType(t ContentPartType) llmapi.ContentPartType {
	if t == ContentPartType_CONTENT_PART_TYPE_IMAGE_URL {
		return llmapi.ContentPartTypeImageURL
	}
	return llmapi.ContentPartTypeText
}

func toProtoEventType(t llmapi.EventType) EventType {
	switch t {
	case llmapi.EventTextDelta:
		return EventType_EVENT_TYPE_TEXT_DELTA
	case llmapi.EventReasoningDelta:
		return EventType_EVENT_TYPE_REASONING_DELTA
	case llmapi.EventToolCallDone:
		return EventType_EVENT_TYPE_TOOL_CALL_DONE
	case llmapi.EventStreamDone:
		return EventType_EVENT_TYPE_STREAM_DONE
	case llmapi.EventStreamError:
		return EventType_EVENT_TYPE_STREAM_ERROR
	case llmapi.EventStreamReset:
		return EventType_EVENT_TYPE_STREAM_RESET
	case llmapi.EventRateLimitWarning:
		return EventType_EVENT_TYPE_RATE_LIMIT_WARNING
	}
	return EventType_EVENT_TYPE_TEXT_DELTA
}

func fromProtoEventType(t EventType) llmapi.EventType {
	switch t {
	case EventType_EVENT_TYPE_TEXT_DELTA:
		return llmapi.EventTextDelta
	case EventType_EVENT_TYPE_REASONING_DELTA:
		return llmapi.EventReasoningDelta
	case EventType_EVENT_TYPE_TOOL_CALL_DONE:
		return llmapi.EventToolCallDone
	case EventType_EVENT_TYPE_STREAM_DONE:
		return llmapi.EventStreamDone
	case EventType_EVENT_TYPE_STREAM_ERROR:
		return llmapi.EventStreamError
	case EventType_EVENT_TYPE_STREAM_RESET:
		return llmapi.EventStreamReset
	case EventType_EVENT_TYPE_RATE_LIMIT_WARNING:
		return llmapi.EventRateLimitWarning
	}
	return llmapi.EventTextDelta
}

// ContextWindowExceededStatus encodes ErrContextWindowExceeded as a gRPC
// status with a typed detail so the client can recover the original error.
func ContextWindowExceededStatus(err *llmapi.ErrContextWindowExceeded) error {
	st := status.New(codes.ResourceExhausted, err.Error())
	withDetail, dErr := st.WithDetails(&ContextWindowExceededDetail{
		Count: int32(err.Count),
		Max:   int32(err.Max),
	})
	if dErr != nil {
		return st.Err()
	}
	return withDetail.Err()
}

// ContextWindowExceededFromStatus inspects err and, if it carries a
// ContextWindowExceededDetail, returns the typed llmapi error.
func ContextWindowExceededFromStatus(err error) (*llmapi.ErrContextWindowExceeded, bool) {
	st, ok := status.FromError(err)
	if !ok {
		return nil, false
	}
	for _, d := range st.Details() {
		if det, ok := d.(*ContextWindowExceededDetail); ok {
			return &llmapi.ErrContextWindowExceeded{
				Count: int(det.GetCount()),
				Max:   int(det.GetMax()),
			}, true
		}
	}
	return nil, false
}
