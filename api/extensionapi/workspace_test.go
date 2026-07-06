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
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestParseSocket(t *testing.T) {
	tests := []struct {
		name        string
		socket      string
		wantNetwork string
		wantAddress string
	}{
		{
			name:        "bare path dials unix",
			socket:      "/tmp/rune/sockets/abc.sock",
			wantNetwork: "unix",
			wantAddress: "/tmp/rune/sockets/abc.sock",
		},
		{
			name:        "unix scheme dials unix",
			socket:      "unix:///tmp/rune/sockets/abc.sock",
			wantNetwork: "unix",
			wantAddress: "/tmp/rune/sockets/abc.sock",
		},
		{
			name:        "tcp scheme dials tcp",
			socket:      "tcp://192.168.65.1:52341",
			wantNetwork: "tcp",
			wantAddress: "192.168.65.1:52341",
		},
		{
			name:        "empty socket defaults to unix",
			socket:      "",
			wantNetwork: "unix",
			wantAddress: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			network, address := parseSocket(tt.socket)
			assert.Equal(t, tt.wantNetwork, network)
			assert.Equal(t, tt.wantAddress, address)
		})
	}
}

// TestNewWorkspaceDialTargets proves the workspace's context dialer
// reaches a host listening on each supported Config.Socket form. The
// server has no services registered, so a codes.Unimplemented status
// proves the connection (dial + HTTP/2 handshake) was established;
// a transport error would surface as codes.Unavailable.
func TestNewWorkspaceDialTargets(t *testing.T) {
	tests := []struct {
		name   string
		listen func(t *testing.T) (net.Listener, string)
	}{
		{
			name: "tcp socket",
			listen: func(t *testing.T) (net.Listener, string) {
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				return lis, "tcp://" + lis.Addr().String()
			},
		},
		{
			name: "unix socket path",
			listen: func(t *testing.T) (net.Listener, string) {
				tmpDir, err := os.MkdirTemp("", "extapi")
				require.NoError(t, err)
				t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })
				sock := filepath.Join(tmpDir, "s.sock")
				lis, err := net.Listen("unix", sock)
				require.NoError(t, err)
				return lis, sock
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lis, socket := tt.listen(t)
			srv := grpc.NewServer()
			go func() { _ = srv.Serve(lis) }()
			t.Cleanup(srv.Stop)

			w, err := NewWorkspace(
				Config{Socket: socket}, Metadata{ExtensionID: "test"})
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(
				context.Background(), 5*time.Second)
			defer cancel()
			err = w.RawConn().Invoke(ctx, "/test.Probe/Ping", nil, nil)
			require.Error(t, err)
			assert.Equal(t, codes.Unimplemented, status.Code(err),
				"expected Unimplemented from a reachable server, got: %v", err)
		})
	}
}
