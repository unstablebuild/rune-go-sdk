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

package termtest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/handler/inputbox"
)

func TestHarnessStep(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	// Initial render
	out := h.Render()
	assert.Equal(t, "▐         ", out)

	// Type 'h'
	out, err := h.Step("h")
	require.NoError(t, err)
	assert.Equal(t, "h▐        ", out)

	// Type 'ello'
	out, err = h.Step("ello")
	require.NoError(t, err)
	assert.Equal(t, "hello▐    ", out)
}

func TestHarnessRun(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	input := strings.NewReader(`h
e
llo`)
	var output bytes.Buffer

	err := h.Run(input, &output)
	require.NoError(t, err)

	// StringWriter pads output to full width
	expected := "h▐        \n---\nhe▐       \n---\nhello▐    \n"
	assert.Equal(t, expected, output.String())
}

func TestHarnessRunWithComments(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	input := strings.NewReader(`# This is a comment
hello
# Another comment
`)
	var output bytes.Buffer

	err := h.Run(input, &output)
	require.NoError(t, err)

	// Only one output (no separator needed for single output)
	assert.Equal(t, "hello▐    \n", output.String())
}

func TestHarnessRunWithSpecialKeys(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	input := strings.NewReader(`hello
<home>
<delete>`)
	var output bytes.Buffer

	err := h.Run(input, &output)
	require.NoError(t, err)

	lines := strings.Split(output.String(), "---\n")
	require.Len(t, lines, 3)

	// After hello
	assert.Contains(t, lines[0], "hello")
	// After home
	assert.Contains(t, lines[1], "▐ello")
	// After delete
	assert.Contains(t, lines[2], "▐llo")
}

func TestHarnessResize(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	_, err := h.Step("hello")
	require.NoError(t, err)

	// Resize to smaller
	h.Resize(5, 1)
	out := h.Render()
	// The cursor character ▐ is 3 bytes, so 5 runes = 7 bytes
	// We count runes, not bytes
	assert.Equal(t, 5, len([]rune(out)))
}

func TestHarnessEmptyInput(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	input := strings.NewReader(`

`)
	var output bytes.Buffer

	err := h.Run(input, &output)
	require.NoError(t, err)
	assert.Empty(t, output.String())
}

func TestHarnessParseError(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	// Invalid key sequence
	_, err := h.Step("<invalid>")
	assert.Error(t, err)
}

func TestHarnessHandler(t *testing.T) {
	ib := inputbox.New()
	h := NewHarness(ib, 10, 1)

	assert.Equal(t, ib, h.Handler())
}
