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
// developer and extension metadata. This function blocks until a signal is received
// for the extension to shutdown, in which case the returned error is nil,
// or another error occurs.
func ServeWorkspaceExtension(extension WorkspaceExtension, meta Metadata) error {
	return serveWorkspaceExtension(extension, meta, os.Stdin, os.Stdout)
}

// Config is sent by the host to the extension over stdin.
// It contains the connection configuration needed to establish
// gRPC connections and access workspace resources.
type Config struct {
	// Socket is the unix socket used to establish
	// a secure communication channel with host.
	Socket string `json:"socket"`
	// Token is the oauth2 token used to authenticate
	// and authorize requests against workspace resources.
	Token *oauth2.Token `json:"token"`
	// Certificate is the certificate used to secure the connections.
	Certificate []byte `json:"certificate"`
	// Config is the user's configuration for the running extension.
	Config  map[string]any `json:"config"`
	DataDir string         `json:"datadir"`
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
	in io.ReadCloser, out io.Writer,
) error {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	signal.Ignore(syscall.SIGPIPE)
	defer signal.Stop(sigchan)

	level := getLogLevelEnv()
	setupExtensionLogging(level)

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
		<-sigchan
		return nil
	case <-sigchan:
		return nil
	}
}

type fnWorkspaceExtension struct {
	fn func(context.Context, *Workspace, config.Config) error
}

func (fn fnWorkspaceExtension) ExtendWorkspace(
	ctx context.Context, w *Workspace, cfg config.Config,
) error {
	return fn.fn(ctx, w, cfg)
}
