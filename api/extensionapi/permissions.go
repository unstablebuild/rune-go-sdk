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

package extensionapi

// Permission represents a request to access a resource.
type Permission string

const (
	// PermissionFileSystem requests access to manage
	// the files in a workspace.
	PermissionFileSystem Permission = "permfs"
	// PermissionExecute requests access to execute
	// and stop processes in a workspace.
	PermissionExecute Permission = "permexec"
	// PermissionTerminal requests access to manage a workspace's ptys.
	PermissionTerminal Permission = "permpty"
	// PermissionBrowserWindowManager requests access to a browser's window manager.
	PermissionBrowserWindowManager Permission = "permwm"
	// PermissionBrowserResourceOpener requests access to open new files.
	PermissionBrowserResourceOpener Permission = "permopen"
	// PermissionNotifications requests access to send messages to the UI.
	PermissionNotifications Permission = "permnoti"
	// PermissionInterrupt requests access to interrupt the event loop.
	// This is useful if your extension handler does async updates to its state, as
	// it enables interrupting the main event loop to redraw components.
	PermissionInterrupt Permission = "permint"
	// PermissionEditor requests access to the editor.
	PermissionEditor Permission = "permed"
	// PermissionCommands requests access to registering new commands.
	PermissionCommands Permission = "permcmd"
	// PermissionStorage requests access to persistent storage.
	PermissionStorage Permission = "permstore"
	// PermissionSyntaxTree requests access to AST-level search.
	PermissionSyntaxTree Permission = "permsyntax"
	// PermissionConfig requests access to read the loaded workspace configuration.
	PermissionConfig Permission = "permcfg"
	// PermissionLSP requests access to the Language Server Protocol.
	PermissionLSP Permission = "permlsp"
	// PermissionDebugger requests access to the debugger through the DAP protocol.
	PermissionDebugger Permission = "permdap"
	// PermissionLLM requests access to host LLM completion services.
	PermissionLLM Permission = "permllm"
)

// Permissions is a set of Permission.
type Permissions map[Permission]any

// NewPermissions builds a new set of Permission with the given permissions.
func NewPermissions(perms ...Permission) Permissions {
	ret := Permissions{}
	for _, perm := range perms {
		ret[perm] = nil
	}
	return ret
}

// AllPermissions returns a set of all the permissions.
func AllPermissions() Permissions {
	return Permissions{
		PermissionFileSystem:            nil,
		PermissionExecute:               nil,
		PermissionTerminal:              nil,
		PermissionBrowserWindowManager:  nil,
		PermissionBrowserResourceOpener: nil,
		PermissionNotifications:         nil,
		PermissionInterrupt:             nil,
		PermissionEditor:                nil,
		PermissionCommands:              nil,
		PermissionStorage:               nil,
		PermissionSyntaxTree:            nil,
		PermissionConfig:                nil,
		PermissionLSP:                   nil,
		PermissionDebugger:              nil,
		PermissionLLM:                   nil,
	}
}
