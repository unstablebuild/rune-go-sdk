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
	"strconv"

	"github.com/spf13/cobra"
	"github.com/unstablebuild/rune-go-sdk/api/browserapi"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/term"
)

// windowID implements browserapi.Window.
type windowID uint64

func (w windowID) WindowID() uint64 { return uint64(w) }

func resolveHandler(
	ctx context.Context,
	a *app,
	uriOrPath string,
) (browserapi.Handler, error) {
	parsed, err := a.resolveURIArg(ctx, uriOrPath)
	if err != nil {
		return nil, err
	}
	w, err := a.getWorkspace()
	if err != nil {
		return nil, err
	}
	return w.ResourceOpener(ctx).Open(parsed)
}

func parseWindowID(s string) (windowID, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid window ID %q: %w", s, err,
		)
	}
	return windowID(v), nil
}

func parseOrientation(
	s string,
) (browserapi.Orientation, error) {
	switch s {
	case "default":
		return browserapi.OrientationDefault, nil
	case "top":
		return browserapi.OrientationTop, nil
	case "bottom":
		return browserapi.OrientationBottom, nil
	case "left":
		return browserapi.OrientationLeft, nil
	case "right":
		return browserapi.OrientationRight, nil
	default:
		return 0, fmt.Errorf(
			"invalid orientation %q:"+
				" use default|top|bottom|left|right", s,
		)
	}
}

func parseAlignment(
	s string,
) (component.Alignment, error) {
	switch s {
	case "", "default":
		return 0, nil
	case "left":
		return component.AlignmentLeft, nil
	case "right":
		return component.AlignmentRight, nil
	case "top":
		return component.AlignmentTop, nil
	case "bottom":
		return component.AlignmentBottom, nil
	case "centered":
		return component.AlignmentCentered, nil
	default:
		return 0, fmt.Errorf(
			"invalid alignment %q", s,
		)
	}
}

func newWMCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wm",
		Short: "Window manager commands",
	}

	cmd.AddCommand(
		newWMFocusCmd(a),
		newWMSplitCmd(a),
		newWMFloatingCmd(a),
		newWMSetContentCmd(a),
		newWMCloseCmd(a),
	)

	return cmd
}

func newWMFocusCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Get the focused window ID",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			win, err := w.WindowManager(cmd.Context()).Focus()
			if err != nil {
				return err
			}
			return printResult(
				cmd.Context(), format, win.WindowID(),
				func(v uint64) { fmt.Println(v) },
				[]string{"WindowID"},
			)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

func newWMSplitCmd(a *app) *cobra.Command {
	var format string
	var orientation string

	cmd := &cobra.Command{
		Use:   "split <window-id> <handler-uri>",
		Short: "Split a window",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			orient, err := parseOrientation(orientation)
			if err != nil {
				return err
			}
			wid, err := parseWindowID(args[0])
			if err != nil {
				return err
			}
			h, err := resolveHandler(cmd.Context(), a, args[1])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			win, err := w.WindowManager(cmd.Context()).Split(
				orient, wid, h,
			)
			if err != nil {
				return err
			}
			return printResult(
				cmd.Context(), format, win.WindowID(),
				func(v uint64) { fmt.Println(v) },
				[]string{"WindowID"},
			)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringVarP(
		&orientation, "orientation", "o", "default",
		"Orientation: default|top|bottom|left|right",
	)

	return cmd
}

func newWMFloatingCmd(a *app) *cobra.Command {
	var format string
	var align string
	var offsetX int
	var offsetY int

	cmd := &cobra.Command{
		Use:   "floating <handler-uri>",
		Short: "Create a floating window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			h, err := resolveHandler(cmd.Context(), a, args[0])
			if err != nil {
				return err
			}
			fh, ok := h.(browserapi.Floating)
			if !ok {
				return fmt.Errorf(
					"handler does not implement Floating",
				)
			}
			alignment, err := parseAlignment(align)
			if err != nil {
				return err
			}
			cfg := browserapi.FloatingConfig{
				Alignment: alignment,
				Offset: term.Coordinates{
					X: offsetX, Y: offsetY,
				},
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			win, err := w.WindowManager(cmd.Context()).Floating(fh, cfg)
			if err != nil {
				return err
			}
			return printResult(
				cmd.Context(), format, win.WindowID(),
				func(v uint64) { fmt.Println(v) },
				[]string{"WindowID"},
			)
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)
	cmd.Flags().StringVarP(
		&align, "alignment", "a", "",
		"Alignment: default|left|right|top|bottom|centered",
	)
	cmd.Flags().IntVarP(&offsetX, "offset-x", "x", 0, "X offset")
	cmd.Flags().IntVarP(&offsetY, "offset-y", "y", 0, "Y offset")

	return cmd
}

func newWMSetContentCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "set-content <window-id> <handler-uri>",
		Short: "Set window content",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			wid, err := parseWindowID(args[0])
			if err != nil {
				return err
			}
			h, err := resolveHandler(cmd.Context(), a, args[1])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			err = w.WindowManager(cmd.Context()).SetWindowContent(wid, h)
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

func newWMCloseCmd(a *app) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "close <window-id>",
		Short: "Close a window",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			wid, err := parseWindowID(args[0])
			if err != nil {
				return err
			}
			w, err := a.getWorkspace()
			if err != nil {
				return err
			}
			err = w.WindowManager(cmd.Context()).CloseWindow(wid)
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
