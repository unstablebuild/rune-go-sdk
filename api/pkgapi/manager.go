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

// Package pkgapi resolves the Rune packages installed on the machine a
// workspace is on.
package pkgapi

import (
	"context"
	"errors"

	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// ErrNotInstalled is reported by Manager.LibDir, or by the iterator it
// returns, when the package is not installed on the workspace host and
// the call did not install it. The user may have declined to install
// it, the package may not be published for the host's platform, or
// installing it may require signing in. Rune has already told the user
// why, so callers should degrade rather than notify again.
var ErrNotInstalled = errors.New("pkg: package not installed")

// Manager resolves the Rune packages installed on the machine a
// workspace is on, which need not be the machine the extension runs on.
type Manager interface {
	// LibDir lists the absolute paths, on the workspace host, of the
	// files in the in-use version of pkgID. The paths are suitable for
	// the workspace's FileSystem and Executor.
	//
	// When no version is installed Rune may ask the user whether to
	// install it, and the user can say no. Iterating then blocks until
	// they answer and any install ends, or until ctx is done. It reports
	// ErrNotInstalled when the package ends up not installed.
	LibDir(ctx context.Context, pkgID string) (iterator.Iterator[string], error)
}
