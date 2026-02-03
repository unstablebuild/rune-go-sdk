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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURI(t *testing.T) {
	tsuite := []struct {
		in      string
		wantOut URI
		wantErr bool
	}{
		{"file:///", URI{
			uri: "file:///", name: "/", parsed: url.URL{Scheme: "file", Path: "/"},
		}, false},
		{"file:///~", URI{
			uri: "file:///~", name: "~", parsed: url.URL{Scheme: "file", Path: "/~"},
		}, false},
		{"file:///~/src", URI{
			uri: "file:///~/src", name: "src", parsed: url.URL{Scheme: "file", Path: "/~/src"},
		}, false},
		{"file:///~/src/../", URI{
			uri: "file:///~", name: "~", parsed: url.URL{Scheme: "file", Path: "/~"},
		}, false},
		{"file:///tmp/a", URI{
			uri: "file:///tmp/a", name: "a", parsed: url.URL{Scheme: "file", Path: "/tmp/a"},
		}, false},
		{"ssh://unstable.build/tmp/a", URI{
			uri:    "ssh://unstable.build/tmp/a",
			name:   "ssh://unstable.build/tmp/a",
			parsed: url.URL{Scheme: "ssh", Host: "unstable.build", Path: "/tmp/a"},
		}, false},
		{"ssh://unstable.build", URI{
			uri:    "ssh://unstable.build/",
			name:   "ssh://unstable.build/",
			parsed: url.URL{Scheme: "ssh", Host: "unstable.build", Path: "/"},
		}, false},
		{"ssh://unstable.build/~", URI{
			uri:    "ssh://unstable.build/~",
			name:   "ssh://unstable.build/~",
			parsed: url.URL{Scheme: "ssh", Host: "unstable.build", Path: "/~"},
		}, false},
		{"ssh://potato:farmer@unstable.build/tmp/a", URI{
			uri:  "ssh://potato:farmer@unstable.build/tmp/a",
			name: "ssh://unstable.build/tmp/a",
			parsed: url.URL{
				Scheme: "ssh",
				Host:   "unstable.build",
				Path:   "/tmp/a",
				User:   url.UserPassword("potato", "farmer"),
			},
		}, false},
		{"/tmp/a", URI{}, true},
		{"tmp/a", URI{}, true},
		{"f$le://", URI{}, true},
	}

	for _, tcase := range tsuite {
		t.Run(fmt.Sprintf("ParseURI(%s)", tcase.in), func(t *testing.T) {
			actualOut, actualErr := ParseURI(tcase.in)
			if tcase.wantErr {
				require.Error(t, actualErr)
			} else {
				require.NoError(t, actualErr)
			}
			assert.Equal(t, tcase.wantOut, actualOut)
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tsuite := []struct {
		in  string
		out string
	}{
		{"", "."},
		{"file.sql", "file.sql"},
		{"/file.sql", "/file.sql"},
		{"w\x00ps", "wps"},
		{"RE\nADME.md\n", "README.md"},
	}

	for _, tcase := range tsuite {
		out := sanitizeFilePath(tcase.in)
		assert.Equal(t, tcase.out, out)
	}
}

func TestJoin(t *testing.T) {
	tsuite := []struct {
		uri      string
		elems    []string
		wantURI  string
		wantName string
	}{
		{"file:///tmp", nil, "file:///tmp", "tmp"},
		{"file:///tmp", []string{".sixrc"}, "file:///tmp/.sixrc", ".sixrc"},
		{"file:///tmp", []string{"test", ".sixrc"}, "file:///tmp/test/.sixrc", ".sixrc"},
		{"file:///", nil, "file:///", "/"},
		{"file:///", []string{".sixrc"}, "file:///.sixrc", ".sixrc"},
		{"file:///", []string{"test", ".sixrc"}, "file:///test/.sixrc", ".sixrc"},
		{"ssh://unstablebuild@six.build/tmp", nil, "ssh://unstablebuild@six.build/tmp", "ssh://six.build/tmp"},
		{"ssh://unstablebuild@six.build/tmp", []string{".sixrc"}, "ssh://unstablebuild@six.build/tmp/.sixrc", "ssh://six.build/tmp/.sixrc"},
		{"ssh://unstablebuild@six.build/tmp", []string{"test", ".sixrc"}, "ssh://unstablebuild@six.build/tmp/test/.sixrc", "ssh://six.build/tmp/test/.sixrc"},
		{"ssh://six.build/", nil, "ssh://six.build/", "ssh://six.build/"},
		{"ssh://six.build/", []string{".sixrc"}, "ssh://six.build/.sixrc", "ssh://six.build/.sixrc"},
		{"ssh://six.build/", []string{"test", ".sixrc"}, "ssh://six.build/test/.sixrc", "ssh://six.build/test/.sixrc"},
	}

	for i, tcase := range tsuite {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			u, err := url.Parse(tcase.uri)
			require.NoError(t, err)
			uri := URI{uri: tcase.uri, parsed: *u}

			o, err := url.Parse(tcase.wantURI)
			require.NoError(t, err)
			want := URI{uri: tcase.wantURI, parsed: *o, name: tcase.wantName}

			// sut
			actual := Join(uri, tcase.elems...)
			assert.Equal(t, want, actual)
		})
	}
}

func TestRelPath(t *testing.T) {
	tsuite := []struct {
		workspaceURI string
		uri          string
		expectedOut  string
	}{
		{"file:///", "file:///tmp", "tmp"},
		{"file:///tmp", "file:///tmp", "."},
		{"file:///var", "file:///tmp/file", "/tmp/file"},
		{"file:///var", "file:///var/file", "file"},
		{"file:///var", "file:///var/dir/dir/dir/file", "dir/dir/dir/file"},
		{"file:///var/", "file:///var/file", "file"},
		{"ssh://user@host:2233/", "ssh://user@host:2234/tmp", "/tmp"},
		{"ssh://user@host:2233/", "ssh://user@otherhost:2233/tmp", "/tmp"},
		{"ssh://user@host:2233/", "ssh://otheruser:2233/tmp", "/tmp"},
		{"ssh://user@host/", "ssh://otheruser/tmp", "/tmp"},
		{"file:///", "ssh:///tmp", "/tmp"},
	}

	for i, tcase := range tsuite {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			inURI, err := ParseURI(tcase.uri)
			require.NoError(t, err)

			inWorkspaceURI, err := ParseURI(tcase.workspaceURI)
			require.NoError(t, err)

			// sut
			actualOut := RelPath(inWorkspaceURI, inURI)
			assert.Equal(t, tcase.expectedOut, actualOut)
		})
	}
}

func TestRandomURI(t *testing.T) {
	uri := RandomURI("myscheme")
	assert.Equal(t, "myscheme", uri.Scheme())
	assert.NotZero(t, uri.Path())

	uri2 := RandomURI("myscheme")
	assert.NotEqual(t, uri2.Path(), uri.Path())

	assert.False(t, uri.Equal(uri2))
	assert.False(t, uri2.Equal(uri))
}
