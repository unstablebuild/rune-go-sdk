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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
)

func newNotifyCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "notify <level> <message>",
		Short: "Send a notification (level: error|warn|info|success)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			level, err := parseNotificationLevel(args[0])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			_, err = w.Notifications(cmd.Context()).Notify(level, args[1])
			if err != nil {
				return err
			}
			return printOK(format)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

func parseNotificationLevel(
	s string,
) (browserapi.NotificationLevel, error) {
	switch s {
	case "error":
		return browserapi.LevelError, nil
	case "warn":
		return browserapi.LevelWarn, nil
	case "info":
		return browserapi.LevelInfo, nil
	case "success":
		return browserapi.LevelSuccess, nil
	default:
		return 0, fmt.Errorf(
			"invalid level %q: use error|warn|info|success",
			s,
		)
	}
}
