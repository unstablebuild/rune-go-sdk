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

	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/cli"
)

type notifyCLI struct {
	app    *app
	fs     *cli.FlagSet
	format string
}

func newNotifyCLI(a *app) *notifyCLI {
	c := &notifyCLI{app: a}
	c.fs = cli.NewFlagSet("runectrl notify")
	c.fs.StringVar(
		&c.format, "F", "",
		"Output format: table, json, or Go template",
	)
	return c
}

func (c *notifyCLI) Run(
	ctx context.Context, args []string,
) (retErr error) {
	defer func() {
		retErr = formatError(c.format, retErr)
	}()
	rargs, _, ok, err := cli.ParseUsage(c, c.fs, 2, args)
	if !ok || err != nil {
		return err
	}
	level, err := parseNotificationLevel(rargs[0])
	if err != nil {
		return err
	}
	w, err := c.app.getWorkspace()
	if err != nil {
		return err
	}
	_, err = w.Notifications(ctx).Notify(level, rargs[1])
	if err != nil {
		return err
	}
	return printOK(c.format)
}

func (c *notifyCLI) Man() cli.Manual {
	return cli.Manual{
		Name: "notify",
		Summary: "Send a notification",
		Synopsis: "[options] <level> <message>" +
			" (level: error|warn|info|success)",
		Options: *c.fs,
	}
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
