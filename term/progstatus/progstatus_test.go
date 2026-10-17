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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReportEncode(t *testing.T) {
	tests := []struct {
		name   string
		report Report
		bel    bool
		want   string
	}{
		{
			name:   "idle",
			report: Report{State: Idle},
			want:   "\x1b]7501;state=idle\x1b\\",
		},
		{
			name:   "bel",
			report: Report{State: Idle},
			bel:    true,
			want:   "\x1b]7501;state=idle\x07",
		},
		{
			name:   "spec/terraform",
			report: Report{State: Blocked, Kind: Permission, App: "terraform", Msg: "Apply 3 to add, 1 to change, 0 to destroy?"},
			want:   "\x1b]7501;state=blocked:kind=permission:app=terraform:msg=QXBwbHkgMyB0byBhZGQsIDEgdG8gY2hhbmdlLCAwIHRvIGRlc3Ryb3k/\x1b\\",
		},
		{
			name:   "spec/brew-auth",
			report: Report{State: Blocked, Kind: Auth, App: "brew", Msg: "Password required to install updates"},
			want:   "\x1b]7501;state=blocked:kind=auth:app=brew:msg=UGFzc3dvcmQgcmVxdWlyZWQgdG8gaW5zdGFsbCB1cGRhdGVz\x1b\\",
		},
		{
			name:   "spec/brew-done",
			report: Report{State: Done, App: "brew", Msg: "Upgraded 12 packages"},
			want:   "\x1b]7501;state=done:app=brew:msg=VXBncmFkZWQgMTIgcGFja2FnZXM=\x1b\\",
		},
		{
			name:   "spec/deploy-region",
			report: Report{State: Working, ID: "us-east", Progress: Percent(40), Title: "US East", Msg: "Pushing image"},
			want:   "\x1b]7501;state=working:id=us-east:progress=40:title=VVMgRWFzdA==:msg=UHVzaGluZyBpbWFnZQ==\x1b\\",
		},
		{
			name:   "progress-zero",
			report: Report{State: Working, Progress: Percent(0)},
			want:   "\x1b]7501;state=working:progress=0\x1b\\",
		},
		{
			name:   "clear/id",
			report: Report{State: Clear, ID: "build/test"},
			want:   "\x1b]7501;state=clear:id=build/test\x1b\\",
		},
		{
			name:   "unicode",
			report: Report{State: Done, Msg: "Déployé ✓"},
			want:   "\x1b]7501;state=done:msg=RMOpcGxvecOpIOKckw==\x1b\\",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, tt.report.Validate())
			require.Equal(t, tt.want, tt.report.Encode(tt.bel))
		})
	}
}

func TestReportValidate(t *testing.T) {
	tests := []struct {
		name    string
		report  Report
		wantErr string
	}{
		{name: "state", report: Report{State: "busy"}, wantErr: `invalid state "busy"`},
		{name: "state/empty", report: Report{}, wantErr: `invalid state ""`},
		{name: "kind/value", report: Report{State: Blocked, Kind: "waiting"}, wantErr: `invalid kind "waiting"`},
		{name: "kind/state", report: Report{State: Working, Kind: Auth}, wantErr: "kind only applies to the blocked state"},
		{name: "progress/range", report: Report{State: Working, Progress: Percent(101)}, wantErr: "invalid progress 101"},
		{name: "progress/negative", report: Report{State: Working, Progress: Percent(-1)}, wantErr: "invalid progress -1"},
		{name: "progress/state", report: Report{State: Done, Progress: Percent(50)}, wantErr: "progress only applies to the working and blocked states"},
		{name: "id/chars", report: Report{State: Idle, ID: "a b"}, wantErr: `invalid id "a b"`},
		{name: "id/empty-segment", report: Report{State: Idle, ID: "a//b"}, wantErr: `invalid id "a//b"`},
		{name: "id/segment-length", report: Report{State: Idle, ID: strings.Repeat("x", 33)}, wantErr: "invalid id"},
		{name: "id/depth", report: Report{State: Idle, ID: strings.Repeat("a/", 8) + "a"}, wantErr: "9 levels, over the 8 level limit"},
		{name: "id/length", report: Report{State: Idle, ID: strings.Repeat(strings.Repeat("x", 32)+"/", 3) + strings.Repeat("x", 30)}, wantErr: "129 bytes, over the 128 byte limit"},
		{name: "app/chars", report: Report{State: Idle, App: "my app"}, wantErr: `invalid app "my app"`},
		{name: "app/length", report: Report{State: Idle, App: strings.Repeat("a", 33)}, wantErr: "invalid app"},
		{name: "msg/control", report: Report{State: Done, Msg: "line1\nline2"}, wantErr: "message contains control character U+000A"},
		{name: "msg/c1", report: Report{State: Done, Msg: "bad\u0085char"}, wantErr: "message contains control character U+0085"},
		{name: "msg/utf8", report: Report{State: Done, Msg: "bad\xffbyte"}, wantErr: "message is not valid UTF-8"},
		{name: "msg/length", report: Report{State: Done, Msg: strings.Repeat("m", 2049)}, wantErr: "message is 2049 bytes, over the 2048 byte limit"},
		{name: "title/length", report: Report{State: Done, Title: strings.Repeat("t", 193)}, wantErr: "title is 193 bytes, over the 192 byte limit"},
		{name: "title/control", report: Report{State: Done, Title: "a\tb"}, wantErr: "title contains control character U+0009"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestReportMaxFitsSequenceLimit(t *testing.T) {
	r := Report{
		State:    Working,
		ID:       strings.Repeat(strings.Repeat("x", 32)+"/", 3) + strings.Repeat("x", 29),
		Progress: Percent(100),
		App:      strings.Repeat("a", 32),
		Title:    strings.Repeat("t", 192),
		Msg:      strings.Repeat("m", 2048),
	}
	require.NoError(t, r.Validate())
	require.Less(t, len(r.Encode(false)), 3300)
}

func TestReportString(t *testing.T) {
	r := Report{
		State: Blocked, ID: "eu-west", Kind: Permission, Progress: Percent(10),
		App: "deploy", Title: "EU West", Msg: "Approve?",
	}
	require.Equal(t,
		`blocked id=eu-west kind=permission progress=10 app=deploy title="EU West" msg="Approve?"`,
		r.String())
}
