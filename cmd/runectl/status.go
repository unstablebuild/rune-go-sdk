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
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Program Status Protocol (OSC 7501) wire format and limits, revision 0.3.
const (
	statusOSC   = "\x1b]7501;"
	statusST    = "\x1b\\"
	statusBEL   = "\x07"
	statusQuery = "?"
	// statusDA requests primary device attributes; every terminal answers
	// it, which bounds the wait for a feature detection reply.
	statusDA = "\x1b[c"

	statusMaxSequence    = 4096
	statusMaxMsgDecoded  = 2048
	statusMaxTitleDecode = 192
	statusMaxApp         = 32
	statusMaxID          = 128
	statusMaxIDSegment   = 32
	statusMaxIDDepth     = 8
)

var (
	statusStates = []string{"idle", "working", "done", "blocked", "error", "clear"}
	statusKinds  = []string{"permission", "question", "auth"}

	errStatusUnsupported = errors.New(
		"terminal does not support program status reports",
	)
)

// statusReport is one Program Status Protocol report.
type statusReport struct {
	State string
	ID    string
	Kind  string
	// Progress is a percentage from 0 to 100, or -1 when absent.
	Progress int
	App      string
	Title    string
	Msg      string
}

func (r statusReport) validate() error {
	if !slices.Contains(statusStates, r.State) {
		return fmt.Errorf(
			"invalid state %q: use idle|working|done|blocked|error|clear",
			r.State,
		)
	}
	if r.ID != "" {
		if err := validateStatusID(r.ID); err != nil {
			return err
		}
	}
	if r.Kind != "" {
		if !slices.Contains(statusKinds, r.Kind) {
			return fmt.Errorf(
				"invalid kind %q: use permission|question|auth", r.Kind,
			)
		}
		if r.State != "blocked" {
			return errors.New("kind only applies to the blocked state")
		}
	}
	if r.Progress != -1 {
		if r.Progress < 0 || r.Progress > 100 {
			return fmt.Errorf(
				"invalid progress %d: must be between 0 and 100", r.Progress,
			)
		}
		if r.State != "working" && r.State != "blocked" {
			return errors.New(
				"progress only applies to the working and blocked states",
			)
		}
	}
	if r.App != "" && !validStatusToken(r.App, statusMaxApp) {
		return fmt.Errorf(
			"invalid app %q: use up to %d characters from [A-Za-z0-9_.+-]",
			r.App, statusMaxApp,
		)
	}
	if err := validateStatusText("title", r.Title, statusMaxTitleDecode); err != nil {
		return err
	}
	if err := validateStatusText("message", r.Msg, statusMaxMsgDecoded); err != nil {
		return err
	}
	if n := len(r.encode(false)); n > statusMaxSequence {
		return fmt.Errorf(
			"report is %d bytes, over the %d byte limit", n, statusMaxSequence,
		)
	}
	return nil
}

// encode returns the full escape sequence for the report, terminated by
// ST or, when bel is set, by BEL.
func (r statusReport) encode(bel bool) string {
	var b strings.Builder
	b.WriteString(statusOSC)
	b.WriteString("state=")
	b.WriteString(r.State)
	if r.ID != "" {
		b.WriteString(":id=")
		b.WriteString(r.ID)
	}
	if r.Kind != "" {
		b.WriteString(":kind=")
		b.WriteString(r.Kind)
	}
	if r.Progress >= 0 {
		b.WriteString(":progress=")
		b.WriteString(strconv.Itoa(r.Progress))
	}
	if r.App != "" {
		b.WriteString(":app=")
		b.WriteString(r.App)
	}
	if r.Title != "" {
		b.WriteString(":title=")
		b.WriteString(base64.StdEncoding.EncodeToString([]byte(r.Title)))
	}
	if r.Msg != "" {
		b.WriteString(":msg=")
		b.WriteString(base64.StdEncoding.EncodeToString([]byte(r.Msg)))
	}
	if bel {
		b.WriteString(statusBEL)
	} else {
		b.WriteString(statusST)
	}
	return b.String()
}

func (r statusReport) String() string {
	var b strings.Builder
	b.WriteString(r.State)
	if r.ID != "" {
		b.WriteString(" id=" + r.ID)
	}
	if r.Kind != "" {
		b.WriteString(" kind=" + r.Kind)
	}
	if r.Progress >= 0 {
		b.WriteString(" progress=" + strconv.Itoa(r.Progress))
	}
	if r.App != "" {
		b.WriteString(" app=" + r.App)
	}
	if r.Title != "" {
		b.WriteString(" title=" + strconv.Quote(r.Title))
	}
	if r.Msg != "" {
		b.WriteString(" msg=" + strconv.Quote(r.Msg))
	}
	return b.String()
}

func validStatusToken(s string, maxLen int) bool {
	if s == "" || len(s) > maxLen {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '_', c == '.', c == '+', c == '-':
		default:
			return false
		}
	}
	return true
}

func validateStatusID(id string) error {
	if len(id) > statusMaxID {
		return fmt.Errorf(
			"invalid id: %d bytes, over the %d byte limit", len(id), statusMaxID,
		)
	}
	segments := strings.Split(id, "/")
	if len(segments) > statusMaxIDDepth {
		return fmt.Errorf(
			"invalid id: %d levels, over the %d level limit",
			len(segments), statusMaxIDDepth,
		)
	}
	for _, seg := range segments {
		if !validStatusToken(seg, statusMaxIDSegment) {
			return fmt.Errorf(
				"invalid id %q: segments use up to %d characters from"+
					" [A-Za-z0-9_.+-] separated by /",
				id, statusMaxIDSegment,
			)
		}
	}
	return nil
}

func validateStatusText(name, s string, maxLen int) error {
	if len(s) > maxLen {
		return fmt.Errorf(
			"%s is %d bytes, over the %d byte limit", name, len(s), maxLen,
		)
	}
	if !utf8.ValidString(s) {
		return fmt.Errorf("%s is not valid UTF-8", name)
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return fmt.Errorf(
				"%s contains control character %U", name, r,
			)
		}
	}
	return nil
}

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

// openStatusTarget resolves where reports are written: the explicit
// path, else stdout when it is a terminal, else the controlling tty.
func openStatusTarget(path string) (io.WriteCloser, error) {
	if path != "" {
		return os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return nopWriteCloser{os.Stdout}, nil
	}
	return os.OpenFile("/dev/tty", os.O_WRONLY, 0)
}

func emitStatus(w io.Writer, r statusReport, bel bool) error {
	if err := r.validate(); err != nil {
		return err
	}
	_, err := io.WriteString(w, r.encode(bel))
	return err
}

func newStatusCmd() *cobra.Command {
	var (
		report statusReport
		bel    bool
		tty    string
		format string
	)

	cmd := &cobra.Command{
		Use:   "status <idle|working|done|blocked|error|clear> [message]",
		Short: "Report program status to the terminal (OSC 7501)",
		Long: "Report what the program is doing to the terminal using the" +
			" Program Status Protocol (OSC 7501). The report goes to the" +
			" terminal directly, so it works in any supporting terminal," +
			" over ssh, and inside containers.",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			report.State = args[0]
			if len(args) == 2 {
				report.Msg = args[1]
			}
			out, err := openStatusTarget(tty)
			if err != nil {
				return err
			}
			defer func() { _ = out.Close() }()
			if err := emitStatus(out, report, bel); err != nil {
				return err
			}
			if format == "" {
				return nil
			}
			return printOK(format)
		},
	}

	cmd.Flags().StringVar(
		&report.ID, "id", "",
		"Record id (path such as build/test); absent means the root record",
	)
	cmd.Flags().StringVar(
		&report.App, "app", "",
		"Stable machine-readable program name, such as cargo or terraform",
	)
	cmd.Flags().StringVar(
		&report.Title, "title", "",
		"Short human-readable label for the record",
	)
	cmd.Flags().StringVar(
		&report.Kind, "kind", "",
		"What a blocked program waits for: permission|question|auth",
	)
	cmd.Flags().IntVar(
		&report.Progress, "progress", -1,
		"Progress percentage 0-100 for working or blocked",
	)
	addStatusTargetFlags(cmd, &bel, &tty)
	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	cmd.AddCommand(
		newStatusQueryCmd(),
		newStatusWrapCmd(),
		newStatusDemoCmd(),
	)

	return cmd
}

func addStatusTargetFlags(cmd *cobra.Command, bel *bool, tty *string) {
	cmd.Flags().BoolVar(
		bel, "bel", false,
		"Terminate reports with BEL instead of ST",
	)
	cmd.Flags().StringVar(
		tty, "tty", "",
		"Terminal to write to (default: stdout if a terminal, else /dev/tty)",
	)
}

// scanStatusQueryReply inspects terminal input received after the
// feature detection query and the primary device attributes request
// were sent, and reports which of the two replies have arrived.
func scanStatusQueryReply(buf []byte) (reply, da bool) {
	reply = bytes.Contains(buf, []byte(statusOSC+statusQuery))
	if j := bytes.Index(buf, []byte("\x1b[?")); j >= 0 {
		da = bytes.IndexByte(buf[j:], 'c') >= 0
	}
	return reply, da
}

// queryStatusSupport sends the feature detection query followed by a
// primary device attributes request and waits for the replies. It keeps
// reading until the device attributes reply arrives so that it does not
// leak into the next reader of the terminal, or until timeout elapses.
func queryStatusSupport(rw io.ReadWriter, timeout time.Duration) (bool, error) {
	_, err := io.WriteString(rw, statusOSC+statusQuery+statusST+statusDA)
	if err != nil {
		return false, fmt.Errorf("write query: %w", err)
	}

	type chunk struct {
		data []byte
		err  error
	}
	ch := make(chan chunk)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			buf := make([]byte, 256)
			n, err := rw.Read(buf)
			select {
			case ch <- chunk{data: buf[:n], err: err}:
			case <-done:
				return
			}
			if err != nil {
				return
			}
		}
	}()

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	var buf []byte
	var supported bool
	for {
		select {
		case c := <-ch:
			buf = append(buf, c.data...)
			reply, da := scanStatusQueryReply(buf)
			supported = supported || reply
			if da {
				return supported, nil
			}
			if c.err != nil {
				if errors.Is(c.err, io.EOF) {
					return supported, nil
				}
				return supported, fmt.Errorf("read reply: %w", c.err)
			}
		case <-deadline.C:
			return supported, nil
		}
	}
}

func newStatusQueryCmd() *cobra.Command {
	var (
		timeout time.Duration
		tty     string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Detect whether the terminal supports program status reports",
		Long: "Send the feature detection query and report whether the" +
			" terminal answered. Exits with status 1 when unsupported.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			defer func() { retErr = formatError(format, retErr) }()
			if tty == "" {
				tty = "/dev/tty"
			}
			f, err := os.OpenFile(tty, os.O_RDWR, 0)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			if fd := int(f.Fd()); term.IsTerminal(fd) {
				old, err := term.MakeRaw(fd)
				if err != nil {
					return fmt.Errorf("raw mode: %w", err)
				}
				defer func() { _ = term.Restore(fd, old) }()
			}
			supported, err := queryStatusSupport(f, timeout)
			if err != nil {
				return err
			}
			if !supported {
				return errStatusUnsupported
			}
			return printString(format, "supported", []string{"result"})
		},
	}

	cmd.Flags().DurationVar(
		&timeout, "timeout", time.Second,
		"How long to wait for a reply",
	)
	cmd.Flags().StringVar(
		&tty, "tty", "",
		"Terminal to query (default: /dev/tty)",
	)
	cmd.Flags().StringVarP(
		&format, "format", "F", "",
		"Output format: table, json, or Go template",
	)

	return cmd
}

// exitError asks main to exit with the child's status without printing
// anything, since the child already reported its own failure.
type exitError struct{ code int }

func (e exitError) Error() string {
	return fmt.Sprintf("exit status %d", e.code)
}

func newStatusWrapCmd() *cobra.Command {
	var (
		app string
		msg string
		bel bool
		tty string
	)

	cmd := &cobra.Command{
		Use:   "wrap [flags] -- <command> [args...]",
		Short: "Run a command and report working, then done or error",
		Long: "Run a command with inherited stdio, reporting working" +
			" before it starts and done or error when it exits. SIGINT," +
			" SIGTERM, SIGHUP and SIGQUIT are forwarded to the command." +
			" A command killed by a signal is reported as an error. Exits" +
			" with the command's status.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := openStatusTarget(tty)
			if err != nil {
				return err
			}
			defer func() { _ = out.Close() }()
			if msg == "" {
				msg = strings.Join(args, " ")
			}
			emit := func(state, text string) error {
				return emitStatus(out, statusReport{
					State: state, Progress: -1, App: app, Msg: text,
				}, bel)
			}
			if err := emit("working", msg); err != nil {
				return err
			}
			status, err := runStatusWrapped(args)
			if err != nil {
				_ = emit("error", err.Error())
				return err
			}
			switch {
			case status.Signaled():
				if err := emit("error", fmt.Sprintf("signal: %s", status.Signal())); err != nil {
					return err
				}
				return exitError{code: 128 + int(status.Signal())}
			case status.ExitStatus() == 0:
				return emit("done", msg)
			default:
				code := status.ExitStatus()
				if err := emit("error", fmt.Sprintf("exit status %d", code)); err != nil {
					return err
				}
				return exitError{code: code}
			}
		},
	}
	cmd.Flags().SetInterspersed(false)

	cmd.Flags().StringVar(
		&app, "app", "",
		"Stable machine-readable program name to attach to every report",
	)
	cmd.Flags().StringVar(
		&msg, "msg", "",
		"Message for the working and done reports (default: the command line)",
	)
	addStatusTargetFlags(cmd, &bel, &tty)

	return cmd
}

func runStatusWrapped(argv []string) (syscall.WaitStatus, error) {
	child := exec.Command(argv[0], argv[1:]...)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		return 0, err
	}

	sigs := make(chan os.Signal, 4)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer signal.Stop(sigs)

	waited := make(chan error, 1)
	go func() { waited <- child.Wait() }()
	for {
		select {
		case sig := <-sigs:
			_ = child.Process.Signal(sig)
		case err := <-waited:
			var exitErr *exec.ExitError
			if err != nil && !errors.As(err, &exitErr) {
				return 0, err
			}
			ws, ok := child.ProcessState.Sys().(syscall.WaitStatus)
			if !ok {
				return 0, errors.New("unexpected process state")
			}
			return ws, nil
		}
	}
}

// statusDemoStep is one step of a demo scenario. A ramped step is a
// working report played as a series of reports whose progress goes
// from 0 to 100%, ignoring the progress the step itself carries.
type statusDemoStep struct {
	statusReport
	ramp bool
}

// statusDemoRampStep is the progress, in percent, between two reports
// of a ramped step.
const statusDemoRampStep = 5

// reports returns the reports that play the step.
func (s statusDemoStep) reports() []statusReport {
	if !s.ramp {
		return []statusReport{s.statusReport}
	}
	var ret []statusReport
	for p := 0; p <= 100; p += statusDemoRampStep {
		r := s.statusReport
		r.Progress = p
		ret = append(ret, r)
	}
	return ret
}

// statusDemoScenarios are the examples from the protocol specification,
// with progress ramped through their main working phase.
var statusDemoScenarios = map[string][]statusDemoStep{
	"brew": {
		{statusReport: statusReport{State: "working", Progress: -1, App: "brew", Msg: "Installing updates"}},
		{statusReport: statusReport{State: "blocked", Progress: -1, Kind: "auth", App: "brew", Msg: "Password required to install updates"}},
		{statusReport: statusReport{State: "working", Progress: -1, App: "brew", Msg: "Installing updates"}, ramp: true},
		{statusReport: statusReport{State: "done", Progress: -1, App: "brew", Msg: "Upgraded 12 packages"}},
	},
	"deploy": {
		{statusReport: statusReport{State: "working", Progress: -1, App: "deploy", Msg: "Deploying v2.4.1"}},
		{statusReport: statusReport{State: "working", Progress: 40, ID: "us-east", Title: "US East", Msg: "Pushing image"}, ramp: true},
		{statusReport: statusReport{State: "blocked", Progress: -1, Kind: "permission", ID: "eu-west", Title: "EU West", Msg: "Approve deploy to eu-west (production)?"}},
		{statusReport: statusReport{State: "done", Progress: -1, ID: "us-east", Title: "US East", Msg: "Healthy"}},
		{statusReport: statusReport{State: "clear", Progress: -1, ID: "us-east"}},
		{statusReport: statusReport{State: "clear", Progress: -1, ID: "eu-west"}},
		{statusReport: statusReport{State: "done", Progress: -1, App: "deploy", Msg: "Deployed to 3 regions"}},
	},
	"rsync": {
		{statusReport: statusReport{State: "working", Progress: -1, Msg: "Syncing photos"}, ramp: true},
		{statusReport: statusReport{State: "done", Progress: -1, Msg: "Photos synced"}},
	},
}

func newStatusDemoCmd() *cobra.Command {
	var (
		scenario string
		speed    time.Duration
		progress time.Duration
		bel      bool
		tty      string
	)

	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Walk through the specification examples, with progress while working",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			steps, ok := statusDemoScenarios[scenario]
			if !ok {
				return fmt.Errorf(
					"invalid scenario %q: use brew|deploy|rsync", scenario,
				)
			}
			out, err := openStatusTarget(tty)
			if err != nil {
				return err
			}
			defer func() { _ = out.Close() }()
			for i, step := range steps {
				if i > 0 {
					if err := sleepStatusDemo(cmd.Context(), speed); err != nil {
						return err
					}
				}
				reports := step.reports()
				for j, r := range reports {
					if j > 0 {
						tick := progress / time.Duration(len(reports)-1)
						if err := sleepStatusDemo(cmd.Context(), tick); err != nil {
							return err
						}
					}
					fmt.Println(r.String())
					if err := emitStatus(out, r, bel); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(
		&scenario, "scenario", "brew",
		"Example to play: brew|deploy|rsync",
	)
	cmd.Flags().DurationVar(
		&speed, "speed", time.Second,
		"Pause between reports",
	)
	cmd.Flags().DurationVar(
		&progress, "progress-duration", 8*time.Second,
		"Time a working step takes to go from 0 to 100% progress",
	)
	addStatusTargetFlags(cmd, &bel, &tty)

	return cmd
}

func sleepStatusDemo(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
