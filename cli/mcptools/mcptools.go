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

// Package mcptools exposes Rune's language servers and syntax trees as MCP tools.
//
// Tools are grouped into a core set, which is always registered, and optional themes that
// are opted into with [WithThemes] or [WithAllThemes].
package mcptools

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Theme names a group of optional tools.
type Theme string

const (
	// ThemeRename enables lsp_rename and lsp_prepare_rename.
	ThemeRename Theme = "rename"
	// ThemeCodeActions enables lsp_code_actions and lsp_execute_command.
	ThemeCodeActions Theme = "code-actions"
	// ThemeFormatting enables lsp_formatting, lsp_range_formatting and
	// lsp_on_type_formatting.
	ThemeFormatting Theme = "formatting"
	// ThemeCompletion enables lsp_completion and lsp_signature_help.
	ThemeCompletion Theme = "completion"
	// ThemeCallHierarchy enables lsp_prepare_call_hierarchy, lsp_incoming_calls and
	// lsp_outgoing_calls.
	ThemeCallHierarchy Theme = "call-hierarchy"
	// ThemeTypeHierarchy enables lsp_prepare_type_hierarchy, lsp_type_supertypes and
	// lsp_type_subtypes.
	ThemeTypeHierarchy Theme = "type-hierarchy"
	// ThemeTypeInfo enables lsp_type_definition and lsp_declaration.
	ThemeTypeInfo Theme = "type-info"
	// ThemeWorkspaceDiagnostics enables lsp_workspace_diagnostics.
	ThemeWorkspaceDiagnostics Theme = "workspace-diagnostics"
	// ThemeDocumentInfo enables lsp_document_highlight, lsp_code_lens, lsp_folding_range
	// and lsp_selection_range.
	ThemeDocumentInfo Theme = "document-info"
)

// AlwaysLoadMetaKey is the tool _meta key that asks MCP clients which defer tool loading
// to load the tool up front.
const AlwaysLoadMetaKey = "anthropic/alwaysLoad"

// toolThemes maps every optional tool to its theme. Tools that are not listed are core
// tools.
var toolThemes = map[string]Theme{
	"lsp_rename":                 ThemeRename,
	"lsp_prepare_rename":         ThemeRename,
	"lsp_code_actions":           ThemeCodeActions,
	"lsp_execute_command":        ThemeCodeActions,
	"lsp_formatting":             ThemeFormatting,
	"lsp_range_formatting":       ThemeFormatting,
	"lsp_on_type_formatting":     ThemeFormatting,
	"lsp_completion":             ThemeCompletion,
	"lsp_signature_help":         ThemeCompletion,
	"lsp_prepare_call_hierarchy": ThemeCallHierarchy,
	"lsp_incoming_calls":         ThemeCallHierarchy,
	"lsp_outgoing_calls":         ThemeCallHierarchy,
	"lsp_prepare_type_hierarchy": ThemeTypeHierarchy,
	"lsp_type_supertypes":        ThemeTypeHierarchy,
	"lsp_type_subtypes":          ThemeTypeHierarchy,
	"lsp_type_definition":        ThemeTypeInfo,
	"lsp_declaration":            ThemeTypeInfo,
	"lsp_workspace_diagnostics":  ThemeWorkspaceDiagnostics,
	"lsp_document_highlight":     ThemeDocumentInfo,
	"lsp_code_lens":              ThemeDocumentInfo,
	"lsp_folding_range":          ThemeDocumentInfo,
	"lsp_selection_range":        ThemeDocumentInfo,
}

// defaultAlwaysLoad lists the tools marked with [AlwaysLoadMetaKey] unless
// [WithAlwaysLoad] is given.
var defaultAlwaysLoad = []string{
	"lsp_definition",
	"lsp_references",
	"lsp_implementation",
	"lsp_document_symbols",
	"lsp_workspace_symbols",
	"lsp_diagnostics",
}

type options struct {
	all        bool
	themes     map[Theme]bool
	alwaysLoad map[string]bool
}

// Option configures tool registration.
type Option func(*options)

// WithThemes registers the tools of the given themes in addition to the core tools.
func WithThemes(themes ...Theme) Option {
	return func(o *options) {
		for _, t := range themes {
			o.themes[t] = true
		}
	}
}

// WithAllThemes registers every tool.
func WithAllThemes() Option {
	return func(o *options) { o.all = true }
}

// WithAlwaysLoad replaces the set of tools marked with [AlwaysLoadMetaKey]. With no
// arguments no tool is marked. The default set is lsp_definition, lsp_references,
// lsp_implementation, lsp_document_symbols, lsp_workspace_symbols and lsp_diagnostics.
func WithAlwaysLoad(tools ...string) Option {
	return func(o *options) {
		o.alwaysLoad = make(map[string]bool, len(tools))
		for _, t := range tools {
			o.alwaysLoad[t] = true
		}
	}
}

type registrar struct {
	s    *server.MCPServer
	opts options
}

func newRegistrar(s *server.MCPServer, opts []Option) *registrar {
	o := options{
		themes:     make(map[Theme]bool),
		alwaysLoad: make(map[string]bool, len(defaultAlwaysLoad)),
	}
	for _, t := range defaultAlwaysLoad {
		o.alwaysLoad[t] = true
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &registrar{s: s, opts: o}
}

// add registers tool unless it belongs to a theme that is not enabled.
func (r *registrar) add(tool mcp.Tool, h server.ToolHandlerFunc) {
	if t, ok := toolThemes[tool.Name]; ok &&
		!r.opts.all && !r.opts.themes[t] {
		return
	}
	if r.opts.alwaysLoad[tool.Name] {
		tool.Meta = mcp.NewMetaFromMap(map[string]any{AlwaysLoadMetaKey: true})
	}
	r.s.AddTool(tool, h)
}

// mcpJSON marshals v as JSON text and returns it as an MCP tool result.
func mcpJSON(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(string(data)), nil
}

// mcpErr returns an MCP error result from an error.
func mcpErr(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

// requireUint returns the required argument name as a uint32, rejecting numbers that do
// not convert exactly.
func requireUint(req mcp.CallToolRequest, name string) (uint32, error) {
	f, err := req.RequireFloat(name)
	if err != nil {
		return 0, err
	}
	return toUint(name, f)
}

// optionalUint returns the argument name as a uint32, or def when it is absent or null.
func optionalUint(req mcp.CallToolRequest, name string, def uint32) (uint32, error) {
	if v, ok := req.GetArguments()[name]; !ok || v == nil {
		return def, nil
	}
	return requireUint(req, name)
}

// toUint converts f to a uint32. Negative, fractional and out of range values are errors:
// converting them would wrap or truncate to a position the caller did not ask for.
func toUint(name string, f float64) (uint32, error) {
	if f < 0 || f > math.MaxUint32 || f != math.Trunc(f) {
		return 0, fmt.Errorf("argument %q must be a whole number from 0 to %d, got %v",
			name, uint32(math.MaxUint32), f)
	}
	return uint32(f), nil
}

// mcpPos extracts uri, line, and character from an MCP request.
func mcpPos(req mcp.CallToolRequest) (string, uint32, uint32, error) {
	uri, err := req.RequireString("uri")
	if err != nil {
		return "", 0, 0, err
	}
	line, err := requireUint(req, "line")
	if err != nil {
		return "", 0, 0, err
	}
	char, err := requireUint(req, "character")
	if err != nil {
		return "", 0, 0, err
	}
	return uri, line, char, nil
}

// mcpRange extracts range parameters from an MCP request.
func mcpRange(req mcp.CallToolRequest) (string, [4]uint32, error) {
	uri, err := req.RequireString("uri")
	if err != nil {
		return "", [4]uint32{}, err
	}
	var r [4]uint32
	names := [4]string{"start_line", "start_character", "end_line", "end_character"}
	for i, name := range names {
		if r[i], err = requireUint(req, name); err != nil {
			return "", [4]uint32{}, err
		}
	}
	return uri, r, nil
}
