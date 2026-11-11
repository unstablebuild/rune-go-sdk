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

package extensionapi

import (
	"context"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/config"
)

func TestServeWorkspaceExtension_ShutsDownWhenHostIsGone(t *testing.T) {
	hostGone := make(chan struct{})
	setup := make(chan struct{})
	ext := FuncWorkspaceExtension(func(context.Context, *Workspace, config.Config) error {
		close(setup)
		return nil
	})
	in := io.NopCloser(strings.NewReader(`{"socket":"unix:///nonexistent","config":{}}` + "\n"))
	done := make(chan error, 1)
	go func() {
		done <- serveWorkspaceExtension(ext, Metadata{ExtensionID: "test"}, in, io.Discard, hostGone)
	}()

	select {
	case <-setup:
	case err := <-done:
		t.Fatalf("serve returned before setup: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("extension was never set up")
	}
	select {
	case err := <-done:
		t.Fatalf("serve returned while host alive: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(hostGone)
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after host was gone")
	}
}

func TestWatchProcess(t *testing.T) {
	var alive atomic.Bool
	alive.Store(true)
	gone := watchProcess(42, time.Millisecond, func(pid int) bool {
		assert.Equal(t, 42, pid)
		return alive.Load()
	})

	select {
	case <-gone:
		t.Fatal("reported gone while alive")
	case <-time.After(20 * time.Millisecond):
	}

	alive.Store(false)
	select {
	case <-gone:
	case <-time.After(5 * time.Second):
		t.Fatal("never reported gone")
	}
}

func TestWatchHost(t *testing.T) {
	tests := []struct {
		name    string
		runePid string
		gone    bool
		watched bool
	}{
		{name: "unset", runePid: ""},
		{name: "garbage", runePid: "not-a-pid"},
		{name: "dead host", runePid: strconv.Itoa(math.MaxInt32), watched: true, gone: true},
		{name: "live host", runePid: strconv.Itoa(os.Getpid()), watched: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RUNE_PID", tt.runePid)
			hostGone := watchHost()
			require.Equal(t, tt.watched, hostGone != nil)
			select {
			case <-hostGone:
				assert.True(t, tt.gone, "host reported gone")
			case <-time.After(50 * time.Millisecond):
				assert.False(t, tt.gone, "host not reported gone")
			}
		})
	}
}

func TestProcessAlive(t *testing.T) {
	require.True(t, processAlive(os.Getpid()))
	require.False(t, processAlive(math.MaxInt32))
}
