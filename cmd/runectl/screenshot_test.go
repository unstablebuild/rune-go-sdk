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
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestScreenshotValidation(t *testing.T) {
	for _, tc := range []struct {
		name, pid, datadir, want string
		args                    []string
		unsetPID, canceled      bool
	}{
		{name: "missing pid", unsetPID: true, want: "RUNE_PID is not set"},
		{name: "remote workspace", pid: "", want: "commands in remote workspaces run on the remote host"},
		{name: "malformed pid", pid: "abc", want: "invalid RUNE_PID"},
		{name: "zero pid", pid: "0", want: "invalid RUNE_PID"},
		{name: "negative pid", pid: "-1", want: "invalid RUNE_PID"},
		{name: "overflow pid", pid: "4294967297", want: "invalid RUNE_PID"},
		{name: "missing datadir", pid: "self", want: "RUNE_DATADIR"},
		{name: "canceled", pid: "self", datadir: "valid", canceled: true, want: "context canceled"},
		{name: "bad directory", pid: "self", datadir: "file", want: "read screenshot directory"},
		{name: "signal failure", pid: "2147483647", datadir: "valid", want: "signal Rune process"},
		{name: "arguments", args: []string{"extra"}, want: "unknown command"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pid := tc.pid
			if pid == "self" {
				pid = strconv.Itoa(os.Getpid())
			}
			dir := ""
			if tc.datadir != "" {
				dir = t.TempDir()
			}
			if tc.datadir == "file" {
				if err := os.WriteFile(filepath.Join(dir, "screenshots"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("RUNE_PID", pid)
			if tc.unsetPID {
				if err := os.Unsetenv("RUNE_PID"); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("RUNE_DATADIR", dir)
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, syscall.SIGUSR2)
			defer signal.Stop(signals)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.canceled {
				cancel()
			}
			cmd := newRootCmd()
			cmd.SetArgs(append([]string{"screenshot"}, tc.args...))
			var out bytes.Buffer
			cmd.SetOut(&out)
			err := cmd.ExecuteContext(ctx)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v; want %q", err, tc.want)
			}
			if out.Len() != 0 {
				t.Fatalf("unexpected stdout: %q", out.String())
			}
			select {
			case <-signals:
				t.Fatal("signaled before validation")
			case <-time.After(10 * time.Millisecond):
			}
		})
	}
}

func TestScreenshotCommand(t *testing.T) {
	for _, tc := range []struct {
		name                                 string
		missing, timeout, canceled, temporary bool
	}{
		{name: "existing directory"},
		{name: "missing directory", missing: true},
		{name: "timeout", timeout: true},
		{name: "canceled after signal", canceled: true},
		{name: "temporary then renamed", temporary: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte("png")
			dir := t.TempDir()
			screenshots := filepath.Join(dir, "screenshots")
			if !tc.missing {
				if err := os.Mkdir(screenshots, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(screenshots, "screenshot-old.png"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("RUNE_PID", strconv.Itoa(os.Getpid()))
			t.Setenv("RUNE_DATADIR", dir)
			signals := make(chan os.Signal, 4)
			signal.Notify(signals, syscall.SIGUSR2)
			defer signal.Stop(signals)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			cmd := newRootCmd()
			cmd.SetArgs([]string{"screenshot"})
			var out bytes.Buffer
			cmd.SetOut(&out)
			done := make(chan error, 1)
			go func() { done <- cmd.ExecuteContext(ctx) }()
			select {
			case <-signals:
			case err := <-done:
				t.Fatalf("returned before signal: %v", err)
			case <-ctx.Done():
				t.Fatal("no signal")
			}
			path := filepath.Join(screenshots, "screenshot-new-random.png")
			if tc.canceled {
				cancel()
			}
			if !tc.timeout && !tc.canceled {
				if err := os.MkdirAll(screenshots, 0700); err != nil {
					t.Fatal(err)
				}
				if tc.temporary {
					temp := filepath.Join(screenshots, ".screenshot-new-random.png")
					if err := os.WriteFile(temp, data, 0600); err != nil {
						t.Fatal(err)
					}
					select {
					case err := <-done:
						t.Fatalf("returned for temporary file: %v", err)
					case <-time.After(150 * time.Millisecond):
					}
					if err := os.Rename(temp, path); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			err := <-done
			if tc.timeout || tc.canceled {
				want := context.DeadlineExceeded
				if tc.canceled {
					want = context.Canceled
				}
				if !errors.Is(err, want) || out.Len() != 0 {
					t.Fatalf("error=%v stdout=%q", err, out.String())
				}
				if hint := "screenshots require Rune's GUI"; tc.timeout != strings.Contains(err.Error(), hint) {
					t.Fatalf("error=%v; want hint %q only on timeout", err, hint)
				}
			} else if err != nil || out.String() != path+"\n" {
				t.Fatalf("error=%v stdout=%q; want %q", err, out.String(), path+"\n")
			}
			select {
			case <-signals:
				t.Fatal("signaled more than once")
			default:
			}
		})
	}
}

func TestWaitScreenshot(t *testing.T) {
	for _, tc := range []struct {
		name     string
		want     bool
		canceled bool
	}{
		{name: "complete", want: true},
		{name: "old"}, {name: "unrelated"}, {name: "hidden"},
		{name: "directory"}, {name: "symlink"},
		{name: "missing"}, {name: "canceled", canceled: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			name := "screenshot-test.png"
			if tc.name == "unrelated" {
				name = "other.png"
			}
			if tc.name == "hidden" {
				name = ".screenshot-temp.png"
			}
			path := filepath.Join(dir, name)
			baseline := map[string]bool{}
			if tc.name == "old" {
				baseline[name] = true
			}
			var err error
			switch tc.name {
			case "directory":
				err = os.Mkdir(path, 0700)
			case "symlink":
				err = os.Symlink("missing", path)
			case "missing":
				dir = filepath.Join(dir, "absent")
			default:
				err = os.WriteFile(path, []byte("png"), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
			defer cancel()
			if tc.canceled {
				cancel()
			}
			got, err := waitScreenshot(ctx, dir, baseline)
			if tc.want {
				if err != nil || got != path {
					t.Fatalf("path=%q error=%v", got, err)
				}
			} else {
				want := context.DeadlineExceeded
				if tc.canceled {
					want = context.Canceled
				}
				if !errors.Is(err, want) || got != "" {
					t.Fatalf("path=%q error=%v", got, err)
				}
			}
		})
	}
}
