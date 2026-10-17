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

// Package progstatus implements the Program Status Protocol (OSC 7501),
// revision 0.4, with which a program tells its terminal whether it is
// idle, working, blocked on the user, done or failed, and why.
//
// See https://www.superlogical.com/rex/docs/build/program-status.
package progstatus

import (
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// State is what a program is doing.
type State string

const (
	// Idle means at rest, waiting for the user's next instruction.
	Idle State = "idle"
	// Working means running, with an optional progress percentage.
	Working State = "working"
	// Done means finished: the result is ready and the user has not
	// seen it yet.
	Done State = "done"
	// Blocked means the program cannot continue until the user does
	// something. The report's Kind says what and its Msg says why.
	Blocked State = "blocked"
	// Error means failed and stopped.
	Error State = "error"
	// Clear is not a state: it removes the addressed record and every
	// record beneath it or, in a report without an id, every record.
	Clear State = "clear"
)

// Kind is what a blocked program waits for.
type Kind string

const (
	// Permission is approval to do something.
	Permission Kind = "permission"
	// Question is an answer the user must type.
	Question Kind = "question"
	// Auth is a login, token or credential.
	Auth Kind = "auth"
)

// Wire format and per-report limits.
const (
	osc = "\x1b]7501;"
	st  = "\x1b\\"
	bel = "\x07"

	maxSequence  = 4096
	maxMsg       = 2048
	maxTitle     = 192
	maxApp       = 32
	maxID        = 128
	maxIDSegment = 32
	maxIDDepth   = 8
)

var (
	states = []State{Idle, Working, Done, Blocked, Error, Clear}
	kinds  = []Kind{Permission, Question, Auth}
)

// Report is one Program Status Protocol report. A report replaces the
// record it addresses completely, so a program that wants App or Title
// on a record includes them in every report.
type Report struct {
	State State
	// ID addresses a record by path, such as build/test. Empty
	// addresses the root record.
	ID string
	// Kind says what a blocked program waits for.
	Kind Kind
	// Progress is a percentage from 0 to 100 for working and blocked
	// reports. Nil means the program is busy without a percentage.
	Progress *int
	// App is a stable machine-readable program name, such as cargo.
	App string
	// Title is a short human-readable label for the record.
	Title string
	// Msg is one human-readable line: what the record is doing,
	// waiting for or finished.
	Msg string
}

// Percent returns p as a Report Progress.
func Percent(p int) *int {
	return &p
}

// Validate returns an error when terminals would discard or ignore r.
func (r Report) Validate() error {
	if !slices.Contains(states, r.State) {
		return fmt.Errorf(
			"invalid state %q: use idle|working|done|blocked|error|clear",
			r.State,
		)
	}
	if r.ID != "" {
		if err := validateID(r.ID); err != nil {
			return err
		}
	}
	if r.Kind != "" {
		if !slices.Contains(kinds, r.Kind) {
			return fmt.Errorf(
				"invalid kind %q: use permission|question|auth", r.Kind,
			)
		}
		if r.State != Blocked {
			return errors.New("kind only applies to the blocked state")
		}
	}
	if r.Progress != nil {
		if p := *r.Progress; p < 0 || p > 100 {
			return fmt.Errorf(
				"invalid progress %d: must be between 0 and 100", p,
			)
		}
		if r.State != Working && r.State != Blocked {
			return errors.New(
				"progress only applies to the working and blocked states",
			)
		}
	}
	if r.App != "" && !validToken(r.App, maxApp) {
		return fmt.Errorf(
			"invalid app %q: use up to %d characters from [A-Za-z0-9_.+-]",
			r.App, maxApp,
		)
	}
	if err := validateText("title", r.Title, maxTitle); err != nil {
		return err
	}
	if err := validateText("message", r.Msg, maxMsg); err != nil {
		return err
	}
	if n := len(r.Encode(false)); n > maxSequence {
		return fmt.Errorf(
			"report is %d bytes, over the %d byte limit", n, maxSequence,
		)
	}
	return nil
}

// Encode returns the escape sequence for r, terminated by ST or, when
// useBEL is set, by BEL. It does not validate r.
func (r Report) Encode(useBEL bool) string {
	var b strings.Builder
	b.WriteString(osc)
	b.WriteString("state=")
	b.WriteString(string(r.State))
	if r.ID != "" {
		b.WriteString(":id=")
		b.WriteString(r.ID)
	}
	if r.Kind != "" {
		b.WriteString(":kind=")
		b.WriteString(string(r.Kind))
	}
	if r.Progress != nil {
		b.WriteString(":progress=")
		b.WriteString(strconv.Itoa(*r.Progress))
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
	if useBEL {
		b.WriteString(bel)
	} else {
		b.WriteString(st)
	}
	return b.String()
}

// String returns r in human-readable form, with text values quoted.
func (r Report) String() string {
	var b strings.Builder
	b.WriteString(string(r.State))
	if r.ID != "" {
		fmt.Fprintf(&b, " id=%s", r.ID)
	}
	if r.Kind != "" {
		fmt.Fprintf(&b, " kind=%s", r.Kind)
	}
	if r.Progress != nil {
		fmt.Fprintf(&b, " progress=%d", *r.Progress)
	}
	if r.App != "" {
		fmt.Fprintf(&b, " app=%s", r.App)
	}
	if r.Title != "" {
		fmt.Fprintf(&b, " title=%q", r.Title)
	}
	if r.Msg != "" {
		fmt.Fprintf(&b, " msg=%q", r.Msg)
	}
	return b.String()
}

func validToken(s string, maxLen int) bool {
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

func validateID(id string) error {
	if len(id) > maxID {
		return fmt.Errorf(
			"invalid id: %d bytes, over the %d byte limit", len(id), maxID,
		)
	}
	segments := strings.Split(id, "/")
	if len(segments) > maxIDDepth {
		return fmt.Errorf(
			"invalid id: %d levels, over the %d level limit",
			len(segments), maxIDDepth,
		)
	}
	for _, seg := range segments {
		if !validToken(seg, maxIDSegment) {
			return fmt.Errorf(
				"invalid id %q: segments use up to %d characters from"+
					" [A-Za-z0-9_.+-] separated by /",
				id, maxIDSegment,
			)
		}
	}
	return nil
}

// isControl reports whether r is a control character, which free-text
// values must not contain once decoded.
func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}

func validateText(name, s string, maxLen int) error {
	if len(s) > maxLen {
		return fmt.Errorf(
			"%s is %d bytes, over the %d byte limit", name, len(s), maxLen,
		)
	}
	if !utf8.ValidString(s) {
		return fmt.Errorf("%s is not valid UTF-8", name)
	}
	for _, r := range s {
		if isControl(r) {
			return fmt.Errorf(
				"%s contains control character %U", name, r,
			)
		}
	}
	return nil
}
