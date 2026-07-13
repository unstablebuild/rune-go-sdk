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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi/workspacerpc"
)

// statScheme is a minimal SchemeServer whose Stat answers from an
// in-memory map of path -> isDir. Any path not present reports a
// not-exist error, mirroring a real scheme. Only Stat is served; every
// other method returns Unimplemented via the embedded base.
type statScheme struct {
	workspacerpc.UnimplementedSchemeServer
	files map[string]bool // path -> isDir
}

func (s statScheme) Stat(
	_ context.Context, req *workspacerpc.StatRequest,
) (*workspacerpc.StatResponse, error) {
	isDir, ok := s.files[req.GetFilename()]
	if !ok {
		return &workspacerpc.StatResponse{IsNotExistErr: true}, nil
	}
	return &workspacerpc.StatResponse{Name: req.GetFilename(), IsDir: isDir}, nil
}

// newStatWorkspace serves statScheme over a unix socket and returns a
// Workspace dialing it, so FindInstalled* resolve through the real
// workspacerpc FileSystem client rather than a hand-rolled fake.
func newStatWorkspace(t *testing.T, installDir string, files map[string]bool) *Workspace {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "s.sock")
	lis, err := net.Listen("unix", sock)
	require.NoError(t, err)

	srv := grpc.NewServer()
	workspacerpc.RegisterSchemeServer(srv, statScheme{files: files})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	w, err := NewWorkspace(
		Config{Socket: sock, InstallDir: installDir},
		Metadata{ExtensionID: "test"})
	require.NoError(t, err)
	return w
}

func TestFindInstalledExecutable(t *testing.T) {
	root := "/opt/install"
	w := newStatWorkspace(t, root, map[string]bool{
		root + "/bin/gopls": false,
		root + "/bin/adir":  true,
	})
	ctx := context.Background()

	got, err := w.FindInstalledExecutable(ctx, "gopls")
	require.NoError(t, err)
	assert.Equal(t, root+"/bin/gopls", got)

	_, err = w.FindInstalledExecutable(ctx, "missing")
	assert.ErrorIs(t, err, os.ErrNotExist)

	// A directory under bin/ must be rejected as an executable.
	_, err = w.FindInstalledExecutable(ctx, "adir")
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestFindInstalledResource(t *testing.T) {
	root := "/opt/install"
	w := newStatWorkspace(t, root, map[string]bool{
		root + "/lib/data.txt": false,
		root + "/lib":          true,
	})
	ctx := context.Background()

	got, err := w.FindInstalledResource(ctx, "lib/data.txt")
	require.NoError(t, err)
	assert.Equal(t, root+"/lib/data.txt", got)

	// A directory is a valid resource, unlike an executable.
	got, err = w.FindInstalledResource(ctx, "lib")
	require.NoError(t, err)
	assert.Equal(t, root+"/lib", got)

	_, err = w.FindInstalledResource(ctx, "nope/missing")
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestNewWorkspaceInstallDirFallback(t *testing.T) {
	w, err := NewWorkspace(
		Config{Socket: "x", DataDir: "/local/dd"}, Metadata{ExtensionID: "t"})
	require.NoError(t, err)
	assert.Equal(t, "/local/dd", w.installDir)

	w, err = NewWorkspace(
		Config{Socket: "x", DataDir: "/local/dd", InstallDir: "/remote/id"},
		Metadata{ExtensionID: "t"})
	require.NoError(t, err)
	assert.Equal(t, "/remote/id", w.installDir)
}
