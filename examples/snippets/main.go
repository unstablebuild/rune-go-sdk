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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/unstablebuild/rune-go-sdk/api/config"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
	"github.com/unstablebuild/rune-go-sdk/api/textapi"
)

func main() {
	meta := extensionapi.Metadata{
		DeveloperID:      "rune-sdk-examples",
		DeveloperKey:     "1234",
		DeveloperEmail:   "it@unstable.build",
		ExtensionID:      "snippets",
		ExtensionName:    "Snippets",
		ExtensionVersion: "0.1.0",
		Permissions: extensionapi.NewPermissions(
			extensionapi.PermissionStorage,
			extensionapi.PermissionEditor,
			extensionapi.PermissionCommands,
			extensionapi.PermissionFileSystem,
			extensionapi.PermissionBrowserResourceOpener,
			extensionapi.PermissionBrowserWindowManager,
			extensionapi.PermissionNotifications,
		),
	}

	err := extensionapi.ServeWorkspaceExtension(
		extensionapi.FuncWorkspaceExtension(run),
		meta,
	)
	if err != nil {
		slog.Error("snippets exited", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, ws *extensionapi.Workspace, cfg config.Config) error {
	s := newSnippets(
		ws.Storage(ctx),
		ws.Editor(ctx),
		ws.FileSystem(ctx),
		ws.ResourceOpener(ctx),
		ws.WindowManager(ctx),
		ws.Notifications(ctx),
	)

	events := []textapi.EventType{
		textapi.EventTypeSelection,
		textapi.EventTypeFlush,
		textapi.EventTypeClose,
	}
	if err := ws.Editor(ctx).SubscribeEvents(events, s); err != nil {
		return fmt.Errorf("subscribe events: %w", err)
	}

	manual := textapi.CommandManual{
		Name:     "snippets",
		Summary:  "Insert, edit, copy, or delete reusable text snippets.",
		Synopsis: "[insert|edit|copy|delete] <name>",
	}
	if err := ws.RegisterCommand(manual, s); err != nil {
		return fmt.Errorf("register command: %w", err)
	}

	slog.Info("snippets extension ready", "tmpdir", os.TempDir())

	return nil
}
