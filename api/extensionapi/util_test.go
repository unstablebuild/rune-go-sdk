// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the
// License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific
// language governing permissions and limitations under the
// License.

package extensionapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionForResource(t *testing.T) {
	tests := []struct {
		name     string
		resource string
		wantPerm Permission
		wantOK   bool
	}{
		{
			name:     "DocumentStore maps to storage",
			resource: "/proto.DocumentStore/Get",
			wantPerm: PermissionStorage,
			wantOK:   true,
		},
		{
			name:     "ResourceOpener maps to browser opener",
			resource: "/browser.ResourceOpener/Open",
			wantPerm: PermissionBrowserResourceOpener,
			wantOK:   true,
		},
		{
			name:     "Notifications maps to notifications",
			resource: "/browser.Notifications/Notify",
			wantPerm: PermissionNotifications,
			wantOK:   true,
		},
		{
			name:     "EventPublisher maps to interrupt",
			resource: "/browser.EventPublisher/Publish",
			wantPerm: PermissionInterrupt,
			wantOK:   true,
		},
		{
			name:     "WindowManager maps to browser wm",
			resource: "/browser.WindowManager/Focus",
			wantPerm: PermissionBrowserWindowManager,
			wantOK:   true,
		},
		{
			name:     "Floating maps to browser wm",
			resource: "/browser.Floating/Dimensions",
			wantPerm: PermissionBrowserWindowManager,
			wantOK:   true,
		},
		{
			name:     "Config maps to config",
			resource: "/config.Config/Get",
			wantPerm: PermissionConfig,
			wantOK:   true,
		},
		{
			name:     "Editor maps to editor",
			resource: "/text.Editor/Edit",
			wantPerm: PermissionEditor,
			wantOK:   true,
		},
		{
			name: "Editor SubscribeCommand " +
				"overrides to commands",
			resource: "/text.Editor/SubscribeCommand",
			wantPerm: PermissionCommands,
			wantOK:   true,
		},
		{
			name: "Editor SubscribeREPLCommand " +
				"overrides to commands",
			resource: "/text.Editor/SubscribeREPLCommand",
			wantPerm: PermissionCommands,
			wantOK:   true,
		},
		{
			name:     "Terminal maps to terminal",
			resource: "/workspace.Terminal/NewPty",
			wantPerm: PermissionTerminal,
			wantOK:   true,
		},
		{
			name:     "Scheme maps to filesystem",
			resource: "/workspace.Scheme/Open",
			wantPerm: PermissionFileSystem,
			wantOK:   true,
		},
		{
			name:     "Executor maps to execute",
			resource: "/workspace.Executor/StartCommand",
			wantPerm: PermissionExecute,
			wantOK:   true,
		},
		{
			name:     "Files maps to filesystem",
			resource: "/workspace.Files/Read",
			wantPerm: PermissionFileSystem,
			wantOK:   true,
		},
		{
			name:     "ProxyScheme maps to filesystem",
			resource: "/workspace.ProxyScheme/InitializeProxy",
			wantPerm: PermissionFileSystem,
			wantOK:   true,
		},
		{
			name:     "Manager maps to filesystem",
			resource: "/workspace.Manager/RegisterScheme",
			wantPerm: PermissionFileSystem,
			wantOK:   true,
		},
		{
			name:     "Syntax maps to syntax tree",
			resource: "/syntax.Syntax/Search",
			wantPerm: PermissionSyntaxTree,
			wantOK:   true,
		},
		{
			name:     "LSP maps to lsp",
			resource: "/semantic.LSP/Hover",
			wantPerm: PermissionLSP,
			wantOK:   true,
		},
		{
			name:     "LLM maps to llm",
			resource: "/llm.LLM/CreateCompletion",
			wantPerm: PermissionLLM,
			wantOK:   true,
		},
		{
			name:     "unknown service returns false",
			resource: "/unknown.Service/Method",
			wantPerm: "",
			wantOK:   false,
		},
		{
			name:     "malformed input without slash",
			resource: "noslash",
			wantPerm: "",
			wantOK:   false,
		},
		{
			name:     "empty string returns false",
			resource: "",
			wantPerm: "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perm, ok := PermissionForResource(tt.resource)
			assert.Equal(t, tt.wantPerm, perm)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}
