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

package progstatus

import (
	"context"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const ellipsis = "…"

// Reporter sends one program's reports to its terminal.
//
// Reporter attaches the program's app to every report and makes free
// text fit the protocol: control characters and runs of white space
// become a single space, and text over the limit is truncated. It skips
// a report identical to the last one sent for the same record, so
// callers can report progress every time they measure it.
//
// A Reporter is safe for concurrent use. A nil *Reporter discards every
// report, so code can report without checking whether reporting is on.
type Reporter struct {
	mu      sync.Mutex
	w       io.Writer
	app     string
	records map[string]Report
	err     error
}

// NewReporter returns a Reporter that writes the reports of program app
// to w, which should be the program's terminal. app may be empty.
func NewReporter(w io.Writer, app string) *Reporter {
	return &Reporter{w: w, app: app, records: make(map[string]Report)}
}

// Report sends rep. A report that is invalid even once its text fits,
// such as one with an unknown state, is not sent and Err returns the
// first such error.
func (r *Reporter) Report(rep Report) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.report(rep)
}

// Block reports that the program cannot continue until the user does
// something: kind says what and msg says why. A program blocked midway
// through its work keeps reporting its progress. The returned function
// ends the block by restoring the root record that the block replaced,
// or by reporting working when there was none.
func (r *Reporter) Block(kind Kind, msg string) (resume func()) {
	if r == nil {
		return func() {}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	prev, ok := r.records[""]
	if !ok {
		prev = Report{State: Working}
	}
	blocked := Report{State: Blocked, Kind: kind, Msg: msg}
	if prev.State == Working {
		blocked.Progress = prev.Progress
	}
	r.report(blocked)
	return func() { r.Report(prev) }
}

// Err returns the first error that kept a report from being sent.
func (r *Reporter) Err() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

func (r *Reporter) report(rep Report) {
	if rep.App == "" {
		rep.App = r.app
	}
	rep.Title = fitText(rep.Title, maxTitle)
	rep.Msg = fitText(rep.Msg, maxMsg)
	if err := rep.Validate(); err != nil {
		r.fail(err)
		return
	}
	seq := rep.Encode(false)
	if prev, ok := r.records[rep.ID]; ok && prev.Encode(false) == seq {
		return
	}
	if _, err := io.WriteString(r.w, seq); err != nil {
		r.fail(err)
		return
	}
	r.track(rep)
}

// track updates the records the terminal holds after rep was sent.
func (r *Reporter) track(rep Report) {
	if rep.State != Clear {
		r.records[rep.ID] = rep
		return
	}
	for id := range r.records {
		if rep.ID == "" || id == rep.ID || strings.HasPrefix(id, rep.ID+"/") {
			delete(r.records, id)
		}
	}
}

func (r *Reporter) fail(err error) {
	if r.err == nil {
		r.err = err
	}
}

// fitText makes s a free-text value of at most maxLen bytes: one line
// of valid UTF-8 without control characters.
func fitText(s string, maxLen int) string {
	s = strings.ToValidUTF8(s, "\uFFFD")
	s = strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || isControl(r)
	}), " ")
	if len(s) <= maxLen {
		return s
	}
	cut := maxLen - len(ellipsis)
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return strings.TrimRightFunc(s[:cut], unicode.IsSpace) + ellipsis
}

type contextKey struct{}

// NewContext returns a copy of ctx that carries r.
func NewContext(ctx context.Context, r *Reporter) context.Context {
	return context.WithValue(ctx, contextKey{}, r)
}

// FromContext returns the Reporter that ctx carries, or nil, which
// discards reports, when it carries none.
func FromContext(ctx context.Context) *Reporter {
	r, _ := ctx.Value(contextKey{}).(*Reporter)
	return r
}
