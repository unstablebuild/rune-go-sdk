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


package workspaceapi

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

// URI represents a parsed URI reference.
type URI struct {
	uri    string
	parsed url.URL
	name   string
}

// String returns the full string representation of this URI.
// The result can be fed back into ParseURI to get back a URI.
func (u URI) String() string {
	return u.uri
}

// Name returns a short and non-unique representation of this URI.
func (u URI) Name() string {
	return u.name
}

// Scheme returns the scheme of this URI.
func (u URI) Scheme() string {
	return u.parsed.Scheme
}

// Path relative or absolute path component of this URI.
func (u URI) Path() string {
	return u.parsed.Path
}

// Hostname returns the hostname of this URI or an empty string
// if hostname is not specified.
func (u URI) Hostname() string {
	return u.parsed.Hostname()
}

// Host returns the host of this URI or an empty string
// if host is not specified.
func (u URI) Host() string {
	return u.parsed.Host
}

// Port returns the port of this URI or an empty string
// if port is not specified.
func (u URI) Port() string {
	return u.parsed.Port()
}

// User returns the user of this URI or an empty string
// if a user is not specified.
func (u URI) User() string {
	if u.parsed.User != nil {
		return u.parsed.User.Username()
	}
	return ""
}

// Password returns the password of this URI or an empty string
// if a password is not specified.
func (u URI) Password() (string, bool) {
	if u.parsed.User != nil {
		return u.parsed.User.Password()
	}
	return "", false
}

// Equal returns true if this uri is equal to other.
func (u URI) Equal(other URI) bool {
	return u.Scheme() == other.Scheme() &&
		u.Hostname() == other.Hostname() &&
		u.Port() == other.Port() &&
		u.User() == other.User() &&
		u.Path() == other.Path()
}

// ParseURI parses s as a URI in the form of:
// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
func ParseURI(s string) (URI, error) {
	u, err := url.Parse(s)
	if err != nil {
		return URI{}, fmt.Errorf("url parse: %s", err)
	}
	return uriFromURL(u)
}

// Join joins any number of path elements into this URI's path, separating them
// with slashes. For more details see path.Join.
func Join(uri URI, elem ...string) URI {
	pathElems := make([]string, 0, len(elem)+1)
	pathElems = append(pathElems, uri.Path())
	pathElems = append(pathElems, elem...)
	newPath := path.Join(pathElems...)
	uri.parsed.Path = newPath
	ret, err := ParseURI(uri.parsed.String())
	if err != nil {
		panic("failed to parse internally generated URI")
	}
	return ret
}

// Dir returns all but the last element of this URI's path.
func Dir(uri URI) URI {
	uri.parsed.Path = filepath.Dir(uri.parsed.Path)
	ret, err := ParseURI(uri.parsed.String())
	if err != nil {
		panic("failed to parse internally generated URI")
	}
	return ret
}

// ExpandPathWithURI finds the absolute path of a relative path.
// It uses the given URI to resolve user and cwd so ~ is an alias
// for a path relative to the base of the URI.
// See ExpandPath for more details.
func ExpandPathWithURI(
	path string, uri URI,
) (string, error) {
	return ExpandPath(path, func() (*user.User, error) {
		return &user.User{
			Username: uri.User(),
			HomeDir:  uri.Path(),
		}, nil
	}, func() (string, error) {
		return uri.Path(), nil
	})
}

// ExpandPath finds the absolute path of a relative path and
// expands the home shortcut (~) if any. If path is already
// absolute then this function returns the path unchanged.
func ExpandPath(
	path string, getUser func() (*user.User, error),
	cwdFn func() (string, error),
) (string, error) {
	if path == "~" || path == "/~" {
		usr, err := getUser()
		if err != nil {
			return "", fmt.Errorf("could not get current user: %s", err)
		}
		path = usr.HomeDir
	} else if strings.HasPrefix(path, "~/") {
		usr, err := getUser()
		if err != nil {
			return "", fmt.Errorf("could not get current user: %s", err)
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	} else if strings.HasPrefix(path, "/~/") {
		usr, err := getUser()
		if err != nil {
			return "", fmt.Errorf("could not get current user: %s", err)
		}
		path = filepath.Join(usr.HomeDir, path[3:])
	}
	if filepath.IsAbs(path) {
		return path, nil
	}
	cwd, err := cwdFn()
	if err != nil {
		return "", fmt.Errorf("could not get cwd: %s", err)
	}
	abs := filepath.Join(cwd, path)
	return abs, nil
}

// HasPrefix returns tests whether uri begins with prefix.
func HasPrefix(uri, prefix URI) bool {
	return uri.Scheme() == prefix.Scheme() &&
		uri.Hostname() == prefix.Hostname() &&
		uri.Port() == prefix.Port() &&
		uri.User() == prefix.User() &&
		strings.HasPrefix(uri.Path(), prefix.Path())
}

// RelPath returns target's path as a relative path of base,
// or if that's not possible, it will return target's Path
// without change.
func RelPath(base, target URI) string {
	if HasPrefix(target, base) {
		relpath, err := filepath.Rel(base.Path(), target.Path())
		if err == nil {
			return relpath
		}
	}
	return target.Path()
}

// CurrentUserHostURI builds a URI from a path. If path is relative
// it uses the current working directory as the base of the path and
// if ~ is used to identify the home directory, the current user's home
// directory is used as the base.
// This should only used instead of Manager.URI before a workspace.Manager is
// constructed or for other advanced uses cases.
func CurrentUserHostURI(path string) (URI, error) {
	absPath, err := ExpandPath(path, user.Current, os.Getwd)
	if err != nil {
		return URI{}, err
	}
	return makeLocalURI(absPath)
}

// WithPath maps the given URI to returns a URI with
// the same components but path set to path.
func WithPath(u URI, path string) (URI, error) {
	u.parsed.Path = path
	return uriFromURL(&u.parsed)
}

// RandomURI generates a random URI with the given scheme.
//
// Under the hood it uses math.Rand so clients can manage
// seeding the default rand.Source, if desired.
func RandomURI(scheme string) URI {
	ret, err := ParseURI(fmt.Sprintf("%s:///%d", scheme, rand.Int()))
	if err != nil {
		// this panic indicates a programmer error
		// above when crafting URI
		panic(err)
	}
	return ret
}

func uriFromURL(u *url.URL) (URI, error) {
	if u.Host != "" {
		return makeRemoteURI(u), nil
	}
	return makeFileURI(u)
}

func makeRemoteURI(u *url.URL) URI {
	if u.Path == "" {
		u.Path = "/"
	}
	name := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path.Join("/", u.Path))
	return URI{uri: u.String(), parsed: *u, name: name}
}

func makeFileURI(u *url.URL) (URI, error) {
	path := sanitizeFilePath(u.Path)
	u.Path = path
	if u.Scheme == "" {
		return URI{}, fmt.Errorf("invalid empty scheme %#v", u)
	}
	return URI{uri: u.String(), parsed: *u, name: filepath.Base(path)}, nil
}

func makeLocalURI(path string) (URI, error) {
	uriStr := "file://" + path
	u, err := url.Parse(uriStr)
	if err != nil {
		return URI{}, err
	}

	return makeFileURI(u)
}

func sanitizeFilePath(resource string) string {
	resource = filepath.Clean(resource)
	return sanitizeLine(resource)
}

func sanitizeLine(in string) string {
	var b strings.Builder
	for _, r := range in {
		switch r {
		case '\x00':
		case '\n':
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
