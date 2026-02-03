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

//revive:disable:exported
package schemeapi

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"syscall"

	"github.com/unstablebuild/rune-go-sdk/api/config"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
)

var (
	// ErrSchemeAlreadyRegistered is retured by RegisterScheme when a scheme for the given
	// uri has already been registered.
	ErrSchemeAlreadyRegistered = errors.New("scheme already registered")
)

// Scheme abstracts internal workspace scheme implementations.
type Scheme interface {
	FileSystem
	Executor
	Terminal

	URI(path string) (workspaceapi.URI, error)
	NewFile(fd uintptr, name string) workspaceapi.File
	// Chroot returns a new filesystem from the same type where the new root is
	// the given path.
	Chroot(path string) (Scheme, error)
	// Root returns the root path of the filesystem.
	Root() string

	Watch(path string, c chan<- EventInfo, events ...Event) (int, error)
	StopWatch(int) error
}

// FileSystem abstracts the ability to manipulate the file system.
type FileSystem interface {
	// Create creates the named file with mode 0666 (before umask), truncating
	// it if it already exists. If successful, methods on the returned File can
	// be used for I/O; the associated file descriptor has mode O_RDWR.
	Create(filename string) (workspaceapi.File, error)
	// Open opens the named file for reading. If successful, methods on the
	// returned file can be used for reading; the associated file descriptor has
	// mode O_RDONLY.
	Open(filename string) (workspaceapi.File, error)
	// OpenFile is the generalized open call; most users will use Open or Create
	// instead. It opens the named file with specified flag (O_RDONLY etc.) and
	// perm, (0666 etc.) if applicable. If successful, methods on the returned
	// File can be used for I/O.
	OpenFile(filename string, flag int, perm fs.FileMode) (workspaceapi.File, error)
	// Stat returns a FileInfo describing the named file.
	Stat(filename string) (fs.FileInfo, error)
	// Rename renames (moves) oldpath to newpath. If newpath already exists and
	// is not a directory, Rename replaces it. OS-specific restrictions may
	// apply when oldpath and newpath are in different directories.
	Rename(oldpath, newpath string) error
	// Remove removes the named file or directory.
	Remove(filename string) error
	// Join joins any number of path elements into a single path, adding a
	// Separator if necessary. Join calls filepath.Clean on the result; in
	// particular, all empty strings are ignored. On Windows, the result is a
	// UNC path if and only if the first path element is a UNC path.
	Join(elem ...string) string
	// TempFile creates a new temporary file in the directory dir with a name
	// beginning with prefix, opens the file for reading and writing, and
	// returns the resulting File. If dir is the empty string, TempFile
	// uses the default directory for temporary files (see os.TempDir).
	// Multiple programs calling TempFile simultaneously will not choose the
	// same file. The caller can use f.Name() to find the pathname of the file.
	// It is the caller's responsibility to remove the file when no longer
	// needed.
	TempFile(dir, prefix string) (workspaceapi.File, error)
	// Lstat returns a FileInfo describing the named file. If the file is a
	// symbolic link, the returned FileInfo describes the symbolic link. Lstat
	// makes no attempt to follow the link.
	Lstat(filename string) (fs.FileInfo, error)
	// Symlink creates a symbolic-link from link to target. target may be an
	// absolute or relative path, and need not refer to an existing node.
	// Parent directories of link are created as necessary.
	Symlink(oldname, newname string) error
	// Readlink returns the target path of link.
	Readlink(link string) (string, error)
	// ReadDir reads the directory named by dirname and returns a list of
	// directory entries sorted by filename.
	ReadDir(path string) ([]fs.DirEntry, error)
	// MkdirAll creates a directory named path, along with any necessary
	// parents, and returns nil, or else returns an error. The permission bits
	// perm are used for all directories that MkdirAll creates. If path is/
	// already a directory, MkdirAll does nothing and returns nil.
	MkdirAll(filename string, perm fs.FileMode) error
}

// Event represents the type of filesystem action.
type Event uint32

// Create, Remove, Write and Rename are the only event values guaranteed to be
// present on all platforms.
const (
	Create Event = iota
	Remove
	Write
	Rename
)

// AllEvents returns a slice with all permutations of Event.
func AllEvents() []Event {
	return []Event{Create, Remove, Write, Rename}
}

// String implements fmt.Stringer interface.
func (e Event) String() string {
	switch e {
	case Create:
		return "create"
	case Remove:
		return "remove"
	case Write:
		return "write"
	case Rename:
		return "rename"
	default:
		panic("uknown event")
	}
}

// EventInfo describes an event reported by Scheme.Watch.
type EventInfo interface {
	// Event is one of the reported events.
	Event() Event
	// URI is the uri of the resource.
	URI() workspaceapi.URI
	// IsDir returns true if event is from a directory.
	IsDir() (bool, error)
}

// Terminal abstracts the ability to manage pseudoterminals.
type Terminal interface {
	// NewPty creates a new pseudoterminal.
	// The provided context is used to kill the process (by calling
	// os.Process.Kill) if the context becomes done before the command completes on
	// its own.
	NewPty(context.Context) (workspaceapi.Pty, error)

	// SetPtySize sets the width and height in columns and rows of
	// a pseudoterminal.
	SetPtySize(p workspaceapi.Pty, width, height int) error
}

// Executor is the public facing API of a workspace's command execution.
type Executor interface {
	// StartCommand starts the given cmd and returns the Pid of the underlying
	// process. The provided context is used to kill the process (by calling
	// os.Process.Kill) if the context.Done channel is closed  before the command
	// completes on its own.
	// Implementations must clean all resources associated with a command
	// once the process exits.
	StartCommand(ctx context.Context, cmd workspaceapi.Cmd) (workspaceapi.Pid, error)

	// Signal sends a signal to the running process.
	Signal(workspaceapi.Pid, syscall.Signal) error

	io.Closer
}

// SchemeFunc represents a Scheme constructor.
type SchemeFunc func(context.Context, config.Config, workspaceapi.URI) (Scheme, error)

// SchemeManager abstracts the ability to register new URI schemes.
type SchemeManager interface {
	RegisterScheme(string, SchemeFunc) error
}
