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

package comptest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/rune-go-sdk/tui"
)

// TestCase represents an action and how a component
// is expected to be drawn after this action.
type TestCase struct {
	Action   func()
	Expected string
}

// StringerWriter is a term.Writer that also is able to produce a string screen output.
type StringerWriter interface {
	term.Writer
	fmt.Stringer
	Flush() error
	Clear(attr term.Attributes) (err error)
}

// TestComponent tests a given component against a set of ComponentTestCase.
func TestComponent(
	t *testing.T, m tui.Component,
	w StringerWriter, cases []TestCase,
) {
	t.Helper()
	var err error

	for i, tcase := range cases {
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			if err = w.Clear(term.Attributes{}); err != nil {
				t.Fatal(err)
			}

			if tcase.Action != nil {
				tcase.Action()
			}

			m.Draw(w)

			if err := w.Flush(); err != nil {
				t.Fatal(err)
			}

			// for readability, we expected strings are written starting with \n
			expected := strings.TrimLeft(tcase.Expected, "\n")
			assert.Equal(t, expected, w.String(), "testcase %d failed", i)
		})
	}
}
