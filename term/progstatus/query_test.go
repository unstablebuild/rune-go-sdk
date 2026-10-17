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
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScanQueryReply(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantReply bool
		wantDA    bool
	}{
		{name: "empty"},
		{name: "reply-only", input: "\x1b]7501;?\x1b\\", wantReply: true},
		{name: "reply-bel", input: "\x1b]7501;?\x07", wantReply: true},
		{name: "reply-with-pairs", input: "\x1b]7501;?:states=idle,working,done,blocked,error\x1b\\", wantReply: true},
		{name: "da-only", input: "\x1b[?6c", wantDA: true},
		{name: "da-multi", input: "\x1b[?62;22c", wantDA: true},
		{name: "da-partial", input: "\x1b[?6"},
		{name: "both", input: "\x1b]7501;?\x1b\\\x1b[?6c", wantReply: true, wantDA: true},
		{name: "reply-partial-then-da", input: "\x1b]750\x1b[?6c", wantDA: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reply, da := scanQueryReply([]byte(tt.input))
			require.Equal(t, tt.wantReply, reply, "reply")
			require.Equal(t, tt.wantDA, da, "da")
		})
	}
}

// fakeTTY serves canned replies one chunk per Read and records what was
// written to it.
type fakeTTY struct {
	writes  bytes.Buffer
	replies []string
	block   chan struct{}
}

func (f *fakeTTY) Write(p []byte) (int, error) {
	return f.writes.Write(p)
}

func (f *fakeTTY) Read(p []byte) (int, error) {
	if len(f.replies) == 0 {
		if f.block != nil {
			<-f.block
		}
		return 0, io.EOF
	}
	n := copy(p, f.replies[0])
	f.replies = f.replies[1:]
	return n, nil
}

func TestQuery(t *testing.T) {
	tests := []struct {
		name    string
		replies []string
		block   bool
		want    bool
	}{
		{
			name:    "supported",
			replies: []string{"\x1b]7501;?\x1b\\", "\x1b[?6c"},
			want:    true,
		},
		{
			name:    "supported-single-chunk",
			replies: []string{"\x1b]7501;?\x1b\\\x1b[?6c"},
			want:    true,
		},
		{
			name:    "supported-split-across-reads",
			replies: []string{"\x1b]75", "01;?\x1b\\\x1b[?", "6c"},
			want:    true,
		},
		{
			name:    "supported-no-da-eof",
			replies: []string{"\x1b]7501;?\x1b\\"},
			want:    true,
		},
		{
			name:    "supported-no-da-timeout",
			replies: []string{"\x1b]7501;?\x1b\\"},
			block:   true,
			want:    true,
		},
		{
			name:    "unsupported-da",
			replies: []string{"\x1b[?62;22c"},
		},
		{
			name: "unsupported-eof",
		},
		{
			name:  "unsupported-timeout",
			block: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tty := &fakeTTY{replies: tt.replies}
			if tt.block {
				tty.block = make(chan struct{})
				defer close(tty.block)
			}
			got, err := Query(tty, 50*time.Millisecond)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			require.Equal(t, "\x1b]7501;?\x1b\\\x1b[c", tty.writes.String())
		})
	}
}
