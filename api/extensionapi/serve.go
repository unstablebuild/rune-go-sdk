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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/unstablebuild/rune-go-sdk/api/config"
	"github.com/unstablebuild/rune-go-sdk/debug"
	"golang.org/x/oauth2"
)

// Metadata represents the developer and extension metadata sent by the extension
// over stdout to the extension host.
type Metadata struct {
	DeveloperID      string      `json:"developer_id"`
	DeveloperEmail   string      `json:"developer_email"`
	DeveloperKey     string      `json:"developer_key"`
	ExtensionID      string      `json:"id"`
	ExtensionName    string      `json:"name"`
	ExtensionVersion string      `json:"version"`
	Permissions      Permissions `json:"permissions"`
}

// WorkspaceExtension abstracts ExtendWorkspace, which needs to
// be satisfied by extensions passed to ServeWorkspaceExtension.
type WorkspaceExtension interface {
	ExtendWorkspace(context.Context, *Workspace, config.Config) error
}

// ServeWorkspaceExtension serves the given workspace extension with the given
// developer and extension metadata. This function blocks until a signal is
// received for the extension to shutdown or the Rune process that started it
// is gone, in which case the returned error is nil, or another error occurs.
func ServeWorkspaceExtension(extension WorkspaceExtension, meta Metadata) error {
	setupExtensionLogging(getLogLevelEnv())
	return serveWorkspaceExtension(extension, meta, os.Stdin, os.Stdout, watchHost())
}

// Config is sent by the host to the extension over stdin.
// It contains the connection configuration needed to establish
// gRPC connections and access workspace resources.
type Config struct {
	// Socket is the address used to establish a secure
	// communication channel with the host: a unix socket path
	// (optionally prefixed with unix://) or a tcp://host:port URI.
	Socket string `json:"socket"`
	// Token is the oauth2 token used to authenticate
	// and authorize requests against workspace resources.
	Token *oauth2.Token `json:"token"`
	// Certificate is the certificate used to secure the connections.
	Certificate []byte `json:"certificate"`
	// Config is the user's configuration for the running extension.
	Config  map[string]any `json:"config"`
	DataDir string         `json:"datadir"`
	// InstallDir is the root, pre-expanded on the workspace host, under
	// which the host provisions resources packaged alongside this
	// extension (e.g. <InstallDir>/bin/<tool>).
	InstallDir string `json:"installdir"`
}

// FuncWorkspaceExtension returns a WorkspaceExtension that calls fn
// on calls to ExtendWorkspace.
func FuncWorkspaceExtension(
	fn func(context.Context, *Workspace, config.Config) error,
) WorkspaceExtension {
	return fnWorkspaceExtension{fn: fn}
}

func serveWorkspaceExtension(
	extension WorkspaceExtension, meta Metadata,
	in io.ReadCloser, out io.Writer, hostGone <-chan struct{},
) error {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)
	defer signal.Stop(sigchan)

	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal protocol response: %w", err)
	}

	if _, err := out.Write(data); err != nil {
		return fmt.Errorf("write protocol response: %w", err)
	}
	_ = os.Stdout.Sync()

	scanner := bufio.NewScanner(in)
	ok := scanner.Scan()
	if !ok {
		return fmt.Errorf("scan protocol exchange: %w", scanner.Err())
	}

	var req Config
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		return fmt.Errorf("unmarshal protocol request: %w", err)
	}

	if req.Socket == "" {
		return errors.New("protocol exchange error: socket is missing" +
			" in request")
	}

	start := time.Now()
	errchan := make(chan error, 1)
	cfg := config.MapConfig(req.Config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go debug.CapturePanicReport(func() {
		workspace, err := NewWorkspace(req, meta)
		if err != nil {
			errchan <- err
			return
		}
		panicValue, captureErr, ok := debug.CapturePanicReportWith(req.DataDir,
			meta.ExtensionID, meta.ExtensionVersion, func() {
				errchan <- extension.ExtendWorkspace(ctx, workspace, cfg)
			})
		if ok {
			return
		}
		if captureErr != nil {
			panic(fmt.Sprintf("capture panic report: "+
				"error capturing: %v: original panic: %v", captureErr, panicValue))
		}
		panicErr := fmt.Errorf("extension panic: %v", panicValue)
		// ensure log is delivered
		_ = os.Stderr.Sync()
		errchan <- panicErr
	})

	select {
	case err := <-errchan:
		if err != nil {
			slog.Error("extension is exiting", "error", err)
			return err
		}
		slog.Debug("extension was setup correctly", "duration", time.Since(start))
	case <-sigchan:
		return nil
	case <-hostGone:
		slog.Info("host process is gone, shutting down")
		return nil
	}
	select {
	case <-sigchan:
	case <-hostGone:
		slog.Info("host process is gone, shutting down")
	}
	return nil
}

// hostPollInterval bounds how long an extension outlives a host that died
// without stopping it.
const hostPollInterval = 30 * time.Second

// watchHost returns a channel that is closed once the Rune process that
// started this extension is gone. Extensions always run next to the IDE,
// so the PID the host names in RUNE_PID is one this process can observe.
// Without RUNE_PID there is nothing to watch and the returned channel
// never fires.
func watchHost() <-chan struct{} {
	pid, err := strconv.Atoi(os.Getenv("RUNE_PID"))
	if err != nil || pid <= 0 {
		return nil
	}
	return watchProcess(pid, hostPollInterval, processAlive)
}

func watchProcess(pid int, interval time.Duration, alive func(int) bool) <-chan struct{} {
	gone := make(chan struct{})
	go debug.CapturePanicReport(func() {
		for alive(pid) {
			time.Sleep(interval)
		}
		close(gone)
	})
	return gone
}

// processAlive reports whether pid names a running process. A process that
// exists but may not be signalled counts as alive.
func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	defer func() { _ = proc.Release() }()
	return !errors.Is(proc.Signal(syscall.Signal(0)), os.ErrProcessDone)
}

type fnWorkspaceExtension struct {
	fn func(context.Context, *Workspace, config.Config) error
}

func (fn fnWorkspaceExtension) ExtendWorkspace(
	ctx context.Context, w *Workspace, cfg config.Config,
) error {
	return fn.fn(ctx, w, cfg)
}
