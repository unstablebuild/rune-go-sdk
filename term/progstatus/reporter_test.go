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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReporter(t *testing.T) {
	tests := []struct {
		name    string
		run     func(r *Reporter)
		want    []Report
		wantErr string
	}{
		{
			name: "app",
			run: func(r *Reporter) {
				r.Report(Report{State: Working, Msg: "Uploading"})
				r.Report(Report{State: Done, App: "other"})
			},
			want: []Report{
				{State: Working, App: "app", Msg: "Uploading"},
				{State: Done, App: "other"},
			},
		},
		{
			name: "dedupe",
			run: func(r *Reporter) {
				for _, p := range []int{0, 0, 1, 1, 1, 2} {
					r.Report(Report{State: Working, Progress: Percent(p)})
				}
			},
			want: []Report{
				{State: Working, App: "app", Progress: Percent(0)},
				{State: Working, App: "app", Progress: Percent(1)},
				{State: Working, App: "app", Progress: Percent(2)},
			},
		},
		{
			name: "dedupe/per-record",
			run: func(r *Reporter) {
				r.Report(Report{State: Working})
				r.Report(Report{State: Working, ID: "a"})
				r.Report(Report{State: Working})
			},
			want: []Report{
				{State: Working, App: "app"},
				{State: Working, ID: "a", App: "app"},
			},
		},
		{
			name: "fit-text",
			run: func(r *Reporter) {
				r.Report(Report{
					State: Error,
					Title: "  a\tb  ",
					Msg:   "2 errors occurred:\n\t* \x1b[31mfirst\n\t* bad\xffbyte\n\n",
				})
			},
			want: []Report{{
				State: Error, App: "app", Title: "a b",
				Msg: "2 errors occurred: * [31mfirst * bad\uFFFDbyte",
			}},
		},
		{
			name: "invalid-not-sent",
			run: func(r *Reporter) {
				r.Report(Report{State: "busy"})
				r.Report(Report{State: Done, Kind: Auth})
				r.Report(Report{State: Idle})
			},
			want:    []Report{{State: Idle, App: "app"}},
			wantErr: `invalid state "busy"`,
		},
		{
			name: "block/resume",
			run: func(r *Reporter) {
				r.Report(Report{State: Working, Progress: Percent(40), Msg: "Uploading"})
				resume := r.Block(Auth, "Passphrase required")
				resume()
			},
			want: []Report{
				{State: Working, App: "app", Progress: Percent(40), Msg: "Uploading"},
				{State: Blocked, Kind: Auth, App: "app", Progress: Percent(40), Msg: "Passphrase required"},
				{State: Working, App: "app", Progress: Percent(40), Msg: "Uploading"},
			},
		},
		{
			name: "block/without-record",
			run: func(r *Reporter) {
				resume := r.Block(Question, "ProjectID")
				resume()
			},
			want: []Report{
				{State: Blocked, Kind: Question, App: "app", Msg: "ProjectID"},
				{State: Working, App: "app"},
			},
		},
		{
			name: "block/ignores-child-records",
			run: func(r *Reporter) {
				r.Report(Report{State: Idle})
				r.Report(Report{State: Working, ID: "a", Progress: Percent(5)})
				resume := r.Block(Permission, "Proceed?")
				resume()
			},
			want: []Report{
				{State: Idle, App: "app"},
				{State: Working, ID: "a", App: "app", Progress: Percent(5)},
				{State: Blocked, Kind: Permission, App: "app", Msg: "Proceed?"},
				{State: Idle, App: "app"},
			},
		},
		{
			name: "clear/forgets-subtree",
			run: func(r *Reporter) {
				r.Report(Report{State: Working, ID: "a/b"})
				r.Report(Report{State: Working, ID: "ab"})
				r.Report(Report{State: Clear, ID: "a"})
				r.Report(Report{State: Working, ID: "a/b"})
				r.Report(Report{State: Working, ID: "ab"})
			},
			want: []Report{
				{State: Working, ID: "a/b", App: "app"},
				{State: Working, ID: "ab", App: "app"},
				{State: Clear, ID: "a", App: "app"},
				{State: Working, ID: "a/b", App: "app"},
			},
		},
		{
			name: "clear/forgets-all",
			run: func(r *Reporter) {
				r.Report(Report{State: Working})
				r.Report(Report{State: Working, ID: "a"})
				r.Report(Report{State: Clear})
				r.Report(Report{State: Working})
				r.Report(Report{State: Working, ID: "a"})
			},
			want: []Report{
				{State: Working, App: "app"},
				{State: Working, ID: "a", App: "app"},
				{State: Clear, App: "app"},
				{State: Working, App: "app"},
				{State: Working, ID: "a", App: "app"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			r := NewReporter(&out, "app")

			tt.run(r)

			var want strings.Builder
			for _, rep := range tt.want {
				want.WriteString(rep.Encode(false))
			}
			require.Equal(t, want.String(), out.String())
			if tt.wantErr == "" {
				require.NoError(t, r.Err())
			} else {
				require.ErrorContains(t, r.Err(), tt.wantErr)
			}
		})
	}
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestReporterWriteError(t *testing.T) {
	broken := errors.New("broken terminal")
	r := NewReporter(failingWriter{err: broken}, "")

	r.Report(Report{State: Working})
	r.Report(Report{State: Done})

	require.ErrorIs(t, r.Err(), broken)
}

func TestReporterNil(t *testing.T) {
	var r *Reporter

	r.Report(Report{State: Working})
	r.Block(Auth, "Passphrase required")()

	require.NoError(t, r.Err())
}

func TestReporterContext(t *testing.T) {
	require.Nil(t, FromContext(context.Background()))

	r := NewReporter(&strings.Builder{}, "app")
	require.Same(t, r, FromContext(NewContext(context.Background(), r)))
}

func TestFitText(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		maxLen int
		want   string
	}{
		{name: "empty", in: "", maxLen: 10, want: ""},
		{name: "fits", in: "hello", maxLen: 5, want: "hello"},
		{name: "white-space", in: " a \r\n b\t", maxLen: 10, want: "a b"},
		{name: "c1-control", in: "a\u0085b\u009bc", maxLen: 10, want: "a b c"},
		{name: "invalid-utf8", in: "a\xffb", maxLen: 10, want: "a\uFFFDb"},
		{name: "truncate", in: "abcdefghij", maxLen: 8, want: "abcde…"},
		{name: "truncate/rune-boundary", in: "ééééé", maxLen: 8, want: "éé…"},
		{name: "truncate/trailing-space", in: "abcd efgh", maxLen: 8, want: "abcd…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fitText(tt.in, tt.maxLen)
			require.Equal(t, tt.want, got)
			require.LessOrEqual(t, len(got), tt.maxLen)
			require.NoError(t, validateText("text", got, tt.maxLen))
		})
	}
}
