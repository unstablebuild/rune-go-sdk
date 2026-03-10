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

import "strings"

// PermissionForResource returns the Permission required to
// access the given gRPC method. The resource should be a full
// gRPC method name such as "/proto.DocumentStore/Get".
// It returns the matching Permission and true if found, or
// an empty Permission and false if the resource is unknown.
func PermissionForResource(resource string) (Permission, bool) {
	trimmed := strings.TrimPrefix(resource, "/")
	service, _, ok := strings.Cut(trimmed, "/")
	if !ok {
		return "", false
	}

	// Method-level overrides take precedence over service.
	if resource == "/text.Editor/SubscribeCommand" ||
		resource == "/text.Editor/SubscribeREPLCommand" {
		return PermissionCommands, true
	}

	switch service {
	case "proto.DocumentStore":
		return PermissionStorage, true
	case "browser.ResourceOpener":
		return PermissionBrowserResourceOpener, true
	case "browser.Notifications":
		return PermissionNotifications, true
	case "browser.EventPublisher":
		return PermissionInterrupt, true
	case "browser.WindowManager", "browser.Floating":
		return PermissionBrowserWindowManager, true
	case "config.Config":
		return PermissionConfig, true
	case "text.Editor":
		return PermissionEditor, true
	case "workspace.Terminal":
		return PermissionTerminal, true
	case "workspace.Scheme",
		"workspace.Files",
		"workspace.ProxyScheme",
		"workspace.Manager":
		return PermissionFileSystem, true
	case "workspace.Executor":
		return PermissionExecute, true
	case "syntax.Syntax":
		return PermissionSyntaxTree, true
	case "semantic.LSP":
		return PermissionLSP, true
	case "debug.Debugger":
		return PermissionDebugger, true
	default:
		return "", false
	}
}
