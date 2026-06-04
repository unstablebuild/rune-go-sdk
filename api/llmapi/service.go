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

// Package llmapi defines the Service interface for AI model providers
// and the request/response types exchanged with them.
package llmapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// Service encapsulates communications with an AI-capabilities provider.
// All methods take the model name explicitly so a single Service can route
// requests to different providers per call.
type Service interface {
	// CreateCompletion attempts to complete the given request using
	// the given model. Returns a stream of typed events.
	// This method should return ErrContextWindowExceeded if a request
	// exceeds the context window of the selected model.
	CreateCompletion(
		ctx context.Context,
		model ModelEntry,
		request Request,
	) (iterator.Iterator[Event], error)

	// CountTokens returns the approximate token count for the given
	// messages under the given model's tokenizer.
	CountTokens(model ModelEntry, messages []Message) (int, error)

	// Models returns an iterator over all available model entries.
	Models() iterator.Iterator[ModelEntry]

	// GetModel returns the canonical entry for the given model and
	// reports whether the model is known to the service. Callers
	// typically pass a ModelEntry with only Name set; the returned
	// entry carries the full provider metadata.
	GetModel(ctx context.Context, model ModelEntry) (ModelEntry, bool)
}

// ModelEntry describes a model available through a Service.
type ModelEntry struct {
	// Name is the model identifier (e.g. "gpt-4o", "claude-opus-4-6").
	Name string
	// Provider identifies which LLM provider serves this model
	// (e.g. "openai", "anthropic", "gemini", "ollama").
	Provider string
	// ContextWindow is the nominal maximum context window in tokens.
	ContextWindow int
	// BaseURL is the provider-specific API base URL.
	// Empty string means use the provider's default.
	BaseURL string
	// ProjectorPath is the absolute path to an optional multimodal projector
	// GGUF (mmproj) associated with a local llama.cpp model. Empty for
	// providers that do not use a separate projector file.
	ProjectorPath string
}

// ReasoningEffort controls the amount of reasoning effort for reasoning models.
type ReasoningEffort string

const (
	// ReasoningEffortNone disables explicit reasoning effort.
	ReasoningEffortNone ReasoningEffort = "none"
	// ReasoningEffortMinimal requests minimal reasoning effort.
	ReasoningEffortMinimal ReasoningEffort = "minimal"
	// ReasoningEffortLow requests low reasoning effort.
	ReasoningEffortLow ReasoningEffort = "low"
	// ReasoningEffortMedium requests medium reasoning effort.
	ReasoningEffortMedium ReasoningEffort = "medium"
	// ReasoningEffortHigh requests high reasoning effort.
	ReasoningEffortHigh ReasoningEffort = "high"
	// ReasoningEffortXHigh requests extra-high reasoning effort.
	ReasoningEffortXHigh ReasoningEffort = "xhigh"
	// ReasoningEffortMax requests the maximum supported reasoning effort.
	ReasoningEffortMax ReasoningEffort = "max"
)

// ReasoningSummary controls the level of reasoning summary output.
type ReasoningSummary string

const (
	// ReasoningSummaryAuto lets the provider choose the summary level.
	ReasoningSummaryAuto ReasoningSummary = "auto"
	// ReasoningSummaryConcise requests a concise reasoning summary.
	ReasoningSummaryConcise ReasoningSummary = "concise"
	// ReasoningSummaryDetailed requests a detailed reasoning summary.
	ReasoningSummaryDetailed ReasoningSummary = "detailed"
	// ReasoningSummaryDisabled disables reasoning summaries.
	ReasoningSummaryDisabled ReasoningSummary = "disabled"
)

// Request is the request type for chat completions.
type Request struct {
	Messages          []Message
	Tools             []Tool
	ToolChoice        ToolChoice
	ParallelToolCalls *bool
	ReasoningEffort   ReasoningEffort
	ReasoningSummary  ReasoningSummary
	MaxOutputTokens   int
	ResponseFormat    *ResponseFormat

	// PromptCacheKey is a stable identifier (typically the dialogue ID)
	// that the provider uses for server-side prompt caching. Requests
	// sharing the same key get a ~90% discount on repeated input-token
	// prefixes. Works without storing responses server-side, preserving
	// Zero Data Retention compatibility.
	PromptCacheKey string

	// TokenCount, when set, is used for the context-window safety check
	// instead of calling CountTokens. The agent loop pre-computes this
	// from the provider-reported usage (or CountTokens + tool estimate on
	// the first turn), so passing it here avoids a redundant re-count on
	// every CreateCompletion call.
	TokenCount int
}

// ToolChoice controls whether and how the model should use tools when tools
// are present on a request.
type ToolChoice string

const (
	// ToolChoiceAuto lets the provider/model decide whether to call a tool.
	ToolChoiceAuto ToolChoice = "auto"
	// ToolChoiceRequired requires the model to emit at least one tool call.
	ToolChoiceRequired ToolChoice = "required"
	// ToolChoiceNone forbids tool use even if tools are declared.
	ToolChoiceNone ToolChoice = "none"
)

// ToolType is the type of tool that a model can invoke.
type ToolType string

const (
	// ToolTypeFunction defines a function tool.
	ToolTypeFunction ToolType = "function"
)

// Tool is a resource that the model may use (like functions, files, etc.).
type Tool struct {
	Type     ToolType
	Function FunctionDefinition
}

// FunctionDefinition defines functions that can be "called" by the model.
type FunctionDefinition struct {
	Name        string
	Description string
	// Parameters is an object describing the function.
	// You can pass a raw byte array describing the schema,
	// or you can pass in a struct which serializes to the proper JSONSchema.
	Parameters any
}

// ToolCall is the result of a model call to a tool.
type ToolCall struct {
	ID       string
	Type     ToolType
	Function FunctionCall
}

// FunctionCall is a function call requested by the model.
type FunctionCall struct {
	Name      string
	Arguments string
}

// ReasoningBlock is a reasoning block preserved for replaying back
// to the model.
type ReasoningBlock struct {
	Kind      string `json:"Kind"`
	Text      string `json:"Text,omitempty"`
	Signature string `json:"Signature,omitempty"`
	Data      string `json:"Data,omitempty"`
}

// Message is a message in a chat with an assistant LLM.
type Message struct {
	Role             Role             `json:"Role"`
	Content          string           `json:"Content"`
	MultiContent     []ContentPart    `json:"MultiContent,omitempty"`
	ToolCalls        []ToolCall       `json:"ToolCalls,omitempty"`
	ToolCallID       string           `json:"ToolCallID,omitempty"`
	Name             string           `json:"Name,omitempty"`
	// ReasoningContent is the human-readable reasoning text for display and is
	// lossy; ReasoningBlocks is the source of truth for replaying reasoning
	// back to the model (it preserves per-block signatures and order). When a
	// provider requires replay, populate ReasoningBlocks, not ReasoningContent.
	ReasoningContent string           `json:"ReasoningContent,omitempty"`
	ReasoningBlocks  []ReasoningBlock `json:"ReasoningBlocks,omitempty"`
	// ProviderItems carries opaque provider-specific items that must be
	// threaded back into the next request to maintain stateful continuity
	// (e.g. ChatGPT Codex backend reasoning items with encrypted_content
	// when store=false). Each entry is the raw JSON of a single item,
	// preserved in the order it was emitted by the provider. Cross-provider
	// code should treat this field as opaque.
	ProviderItems []json.RawMessage `json:"ProviderItems,omitempty"`
}

// UnmarshalJSON handles both the new format and the old persisted format
// where tool calls were in a Metadata field and multi-content was OtherContent.
func (m *Message) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion.
	type Alias Message
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*m = Message(alias)

	// Handle old format: extract tool calls from Metadata field.
	var raw struct {
		Metadata     json.RawMessage `json:"Metadata"`
		OtherContent json.RawMessage `json:"OtherContent"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil // ignore — we already parsed the core fields
	}

	if len(m.ToolCalls) == 0 && len(raw.Metadata) > 0 && string(raw.Metadata) != "null" {
		var meta struct {
			ToolCalls []ToolCall `json:"ToolCalls"`
		}
		if json.Unmarshal(raw.Metadata, &meta) == nil && len(meta.ToolCalls) > 0 {
			m.ToolCalls = meta.ToolCalls
		}
	}

	if len(m.MultiContent) == 0 && len(raw.OtherContent) > 0 && string(raw.OtherContent) != "null" {
		var parts []ContentPart
		if json.Unmarshal(raw.OtherContent, &parts) == nil && len(parts) > 0 {
			m.MultiContent = parts
		}
	}

	return nil
}

// EventType identifies the kind of streaming event.
type EventType int

const (
	// EventTextDelta is a text content delta.
	EventTextDelta EventType = iota
	// EventReasoningDelta is a reasoning content delta.
	EventReasoningDelta
	// EventToolCallDone signals a single tool call is complete.
	EventToolCallDone
	// EventStreamDone signals the response is complete.
	EventStreamDone
	// EventStreamError signals a stream error.
	EventStreamError
	// EventStreamReset signals the consumer should discard accumulated state
	// because a mid-stream retry is in progress.
	EventStreamReset
	// EventRateLimitWarning signals a rate limit warning from the provider.
	// The client composes the message; downstream layers display it as-is.
	EventRateLimitWarning
)

// Event is a typed streaming event from an LLM provider.
type Event struct {
	Type      EventType
	Text      string         // EventTextDelta
	Reasoning string         // EventReasoningDelta
	ToolCall  *ToolCall      // EventToolCallDone — complete tool call
	DoneData  *DoneData      // EventStreamDone
	Error     error          // EventStreamError
	RateLimit *RateLimitInfo // EventRateLimitWarning
}

// RateLimitInfo carries a provider-composed rate limit warning.
type RateLimitInfo struct {
	// WaitDuration is non-zero when the client is actively waiting before a retry.
	WaitDuration time.Duration
	// Message is a human-readable warning composed by the client.
	Message string
}

// DoneData holds the final message and usage when a stream completes.
type DoneData struct {
	Message      Message
	FinishReason FinishReason
	Usage        Usage
}

// Usage holds token usage information from a completion.
type Usage struct {
	TokensSent         int
	TokensReceived     int
	TokensReasoned     int
	TokensCached       int
	TokensCacheCreated int
}

// DialogueUsage accumulates usage statistics across an entire dialogue session.
type DialogueUsage struct {
	TokensSent         int           `json:"InputTokens"`
	TokensReceived     int           `json:"OutputTokens"`
	TokensReasoned     int           `json:"ReasoningTokens"`
	TokensCached       int           `json:"CachedTokens"`
	TokensCacheCreated int           `json:"CacheCreatedTokens"`
	Completions        int           `json:"Completions"`
	ToolCalls          int           `json:"ToolCalls"`
	TotalDuration      time.Duration `json:"TotalDuration"`
	InferenceDuration  time.Duration `json:"InferenceDuration"`
	ToolCallDuration   time.Duration `json:"ToolCallDuration"`
}

// Add accumulates a single completion's usage and timing into the dialogue totals.
func (u *DialogueUsage) Add(usage Usage, toolCalls int, inferenceDuration, toolCallDuration time.Duration) {
	u.TokensSent += usage.TokensSent
	u.TokensReceived += usage.TokensReceived
	u.TokensReasoned += usage.TokensReasoned
	u.TokensCached += usage.TokensCached
	u.TokensCacheCreated += usage.TokensCacheCreated
	u.Completions++
	u.ToolCalls += toolCalls
	u.InferenceDuration += inferenceDuration
	u.ToolCallDuration += toolCallDuration
}

// ResponseFormatType identifies the response format type.
type ResponseFormatType string

const (
	// ResponseFormatTypeText requests plain text output.
	ResponseFormatTypeText ResponseFormatType = "text"
	// ResponseFormatTypeJSONObject requests JSON object output.
	ResponseFormatTypeJSONObject ResponseFormatType = "json_object"
	// ResponseFormatTypeJSONSchema requests JSON schema-constrained output.
	ResponseFormatTypeJSONSchema ResponseFormatType = "json_schema"
)

// ResponseFormat specifies per-request structured output.
type ResponseFormat struct {
	Type       ResponseFormatType        `json:"type,omitempty"`
	JSONSchema *ResponseFormatJSONSchema `json:"json_schema,omitempty"`
}

// ResponseFormatJSONSchema is the JSON schema configuration.
type ResponseFormatJSONSchema struct {
	Name        string
	Description string
	Schema      json.Marshaler
	Strict      bool
}

// ContentPartType identifies the type of a ContentPart.
type ContentPartType uint8

const (
	// ContentPartTypeText identifies a text content part.
	ContentPartTypeText ContentPartType = iota
	// ContentPartTypeImageURL identifies an image URL content part.
	ContentPartTypeImageURL
)

// ContentPart represents textual or image content in a Message.
type ContentPart struct {
	Type     ContentPartType `json:"Type"`
	Text     string          `json:"Text,omitempty"`
	ImageURL string          `json:"ImageURL,omitempty"`
}

// NewContentPartFromImageURL returns a ContentPart from an image URL.
func NewContentPartFromImageURL(url string) ContentPart {
	return ContentPart{
		Type:     ContentPartTypeImageURL,
		ImageURL: url,
	}
}

// NewContentPartFromImage converts an image.Image into a ContentPart.
func NewContentPartFromImage(img image.Image) (ContentPart, error) {
	var b bytes.Buffer
	b.WriteString("data:image/png;base64,")
	writer := base64.NewEncoder(base64.StdEncoding, &b)
	err := png.Encode(writer, img)
	if err != nil {
		return ContentPart{}, fmt.Errorf("png encode: %w", err)
	}
	if err := writer.Close(); err != nil {
		return ContentPart{}, fmt.Errorf("base64 close: %w", err)
	}
	return NewContentPartFromImageURL(b.String()), nil
}

// Role is the role of the message author in a message stream.
type Role string

// List of roles assigned to the different messages.
const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// FinishReason is the reason why the message choice was returned.
type FinishReason string

const (
	// FinishReasonStop API returned complete message,
	// or a message terminated by one of the stop sequences provided via the stop parameter
	FinishReasonStop FinishReason = "stop"
	// FinishReasonLength Incomplete model output due to max_tokens parameter or token limit
	FinishReasonLength FinishReason = "length"
	// FinishReasonToolCall The model decided to use one of the tools provided.
	FinishReasonToolCall FinishReason = "tool_calls"
	// FinishReasonContentFilter Omitted content due to a flag from our content filters
	FinishReasonContentFilter FinishReason = "content_filter"
	// FinishReasonNull API response still in progress or incomplete
	FinishReasonNull FinishReason = "null"
)
