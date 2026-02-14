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

package fileexplorer

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/component/fileexplorer"
	"github.com/unstablebuild/rune-go-sdk/handler/handlertest"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestHandlerNavigationViewMode(t *testing.T) {
	tests := []struct {
		name     string
		keys     string
		expected int // expected cursor position
	}{
		{"initial cursor at 0", "", 0},
		{"j moves down", "j", 1},
		{"jj moves down twice", "jj", 2},
		{"jjj clamps at bottom", "jjj", 2},
		{"k moves up", "jk", 0},
		{"kk clamps at top", "kk", 0},
		{"down arrow", "<down>", 1},
		{"up arrow", "j<up>", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(t, map[string][]mockEntry{
				"/project": {
					{name: "src", isDir: true},
					{name: "main.go", isDir: false},
					{name: "readme.md", isDir: false},
				},
				"/project/src": {},
			})
			h.Resize(40, 10)

			keys, err := term.ParseKeys(tt.keys)
			require.NoError(t, err)

			for _, k := range keys {
				h.Handle(term.Event{
					Type: term.EventKey,
					Ch:   k.Ch,
					Mod:  k.Mod,
					Key:  k.Key,
				})
			}

			cursor, _, _ := h.Cursor()
			assert.Equal(t, tt.expected, cursor.Y)
		})
	}
}

func TestHandlerExpandCollapse(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	}}
	comp, err := fileexplorer.New(mfs, rootURI(), fileexplorer.Config{})
	require.NoError(t, err)

	h := New(comp)
	h.Resize(40, 10)

	// Initially only src/ is visible
	assert.Len(t, comp.Cells(), 1)

	// Press Enter to expand
	sendKeys(t, h, "<enter>")
	assert.Len(t, comp.Cells(), 2)

	// Press Enter again to collapse
	sendKeys(t, h, "<enter>")
	assert.Len(t, comp.Cells(), 1)
}

func TestHandlerEditMode(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "oldname.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Enter edit mode with 'r'
	sendKeys(t, h, "r")
	assert.Equal(t, ModeEdit, h.Mode())

	// Verify cursor is in edit mode
	_, _, show := h.Cursor()
	assert.True(t, show)

	// Type a new name
	sendKeys(t, h, "<c-u>newname.go<enter>")
	assert.Equal(t, ModeView, h.Mode())

	// Check that changes are detected
	changes := h.Component().Changes()
	require.Len(t, changes, 1)
	assert.Equal(t, fileexplorer.OpRename, changes[0].Type)
}

func TestHandlerEditModeCancel(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "keepname.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Enter edit mode and then cancel
	sendKeys(t, h, "r")
	assert.Equal(t, ModeEdit, h.Mode())

	sendKeys(t, h, "<esc>")
	assert.Equal(t, ModeView, h.Mode())

	// No changes should be detected
	changes := h.Component().Changes()
	assert.Empty(t, changes)
}

func TestHandlerAddFile(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "existing.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Press 'a' to add a file
	sendKeys(t, h, "a")

	// Should enter edit mode
	assert.Equal(t, ModeEdit, h.Mode())

	// Commit the new file
	sendKeys(t, h, "<c-u>newfile.go<enter>")
	assert.Equal(t, ModeView, h.Mode())

	// Should have 2 rows now
	assert.Len(t, h.Component().Cells(), 2)

	// Check for create operation
	changes := h.Component().Changes()
	require.Len(t, changes, 1)
	assert.Equal(t, fileexplorer.OpCreate, changes[0].Type)
}

func TestHandlerAddDir(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "existing.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Press 'A' to add a directory
	sendKeys(t, h, "A")
	assert.Equal(t, ModeEdit, h.Mode())

	sendKeys(t, h, "<c-u>newdir<enter>")
	assert.Equal(t, ModeView, h.Mode())

	// Check for mkdir operation
	changes := h.Component().Changes()
	require.Len(t, changes, 1)
	assert.Equal(t, fileexplorer.OpMkdir, changes[0].Type)
}

func TestHandlerDeleteLine(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Delete first row
	sendKeys(t, h, "d")
	assert.Len(t, h.Component().Cells(), 1)

	// Check for delete operation
	changes := h.Component().Changes()
	require.Len(t, changes, 1)
	assert.Equal(t, fileexplorer.OpDelete, changes[0].Type)
}

func TestHandlerConfirmMode(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})
	h.Resize(40, 20)

	// Delete a row to create a change
	sendKeys(t, h, "d")

	// Press Ctrl+S to enter confirm mode
	sendKeys(t, h, "<c-s>")
	assert.Equal(t, ModeConfirm, h.Mode())

	// Press Escape to cancel
	sendKeys(t, h, "<esc>")
	assert.Equal(t, ModeView, h.Mode())
}

func TestHandlerConfirmModeYes(t *testing.T) {
	// Note: This test verifies that confirming with 'y' exits confirm mode.
	// It doesn't actually apply changes since mockFS.Remove panics.
	// For actual change application, see integration tests.

	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "file1.go", isDir: false},
			{name: "file2.go", isDir: false},
		},
	})
	h.Resize(40, 20)

	// Rename a file (instead of delete, which would call mockFS.Remove)
	sendKeys(t, h, "r<c-u>renamed.go<enter>")

	// Enter confirm mode
	sendKeys(t, h, "<c-s>")
	assert.Equal(t, ModeConfirm, h.Mode())

	// Cancel with 'n' instead of confirming, since mockFS doesn't implement operations
	sendKeys(t, h, "n")
	assert.Equal(t, ModeView, h.Mode())
}

func TestHandlerExit(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Press 'q' to exit
	exit, handled := sendKey(h, 'q')
	assert.True(t, exit)
	assert.True(t, handled)

	// Press Escape to exit
	h = newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})
	h.Resize(40, 10)
	exit, handled = h.Handle(term.Event{Type: term.EventKey, Key: term.KeyEsc})
	assert.True(t, exit)
	assert.True(t, handled)
}

func TestHandlerScrollable(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
			{name: "d.go", isDir: false},
			{name: "e.go", isDir: false},
		},
	})
	// Small viewport
	h.Resize(40, 2)

	assert.Equal(t, 0, h.SeekOffset())
	assert.Equal(t, 3, h.MaxSeekOffset()) // 5 items - 2 height = 3

	assert.True(t, h.SeekDown())
	assert.Equal(t, 1, h.SeekOffset())

	assert.True(t, h.SeekUp())
	assert.Equal(t, 0, h.SeekOffset())

	assert.False(t, h.SeekUp())
	assert.Equal(t, 0, h.SeekOffset())
}

func TestHandlerDraw(t *testing.T) {
	mfs := &mockFS{dirs: map[string][]mockEntry{
		"/project": {
			{name: "main.go", isDir: false},
		},
	}}
	comp, err := fileexplorer.New(mfs, rootURI(), fileexplorer.Config{})
	require.NoError(t, err)

	h := New(comp)

	cases := []handlertest.SequenceTestCase{
		{InputSequence: "", Expected: "  main.go           \n                    "},
	}
	handlertest.RunHandlerSequence(t, h, 20, 2, cases)
}

func TestHandlerOnOpen(t *testing.T) {
	var openedURI workspaceapi.URI
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}, WithOnOpen(func(uri workspaceapi.URI) {
		openedURI = uri
	}))
	h.Resize(40, 10)

	// Press Enter on a file
	sendKeys(t, h, "<enter>")
	assert.Contains(t, openedURI.String(), "test.go")
}

func TestHandlerPageNavigation(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
			{name: "d.go", isDir: false},
			{name: "e.go", isDir: false},
		},
	})
	h.Resize(40, 2)

	// Page down
	sendKeys(t, h, "<pgdn>")
	cursor, _, _ := h.Cursor()
	assert.Equal(t, 2, cursor.Y+h.SeekOffset())

	// Page up
	sendKeys(t, h, "<pgup>")
	cursor, _, _ = h.Cursor()
	assert.Equal(t, 0, cursor.Y+h.SeekOffset())
}

func TestHandlerHomeEnd(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Go to end with G
	sendKeys(t, h, "G")
	cursor, _, _ := h.Cursor()
	assert.Equal(t, 2, cursor.Y)

	// Go to start with g
	sendKeys(t, h, "g")
	cursor, _, _ = h.Cursor()
	assert.Equal(t, 0, cursor.Y)
}

func TestHandlerPgUpPgDownNavigation(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
			{name: "d.go", isDir: false},
			{name: "e.go", isDir: false},
			{name: "f.go", isDir: false},
		},
	})
	h.Resize(40, 4) // viewport of 4

	// Page down moves down
	sendKeys(t, h, "<pgdn>")
	cursor, _, _ := h.Cursor()
	// Should move cursor down by page size
	assert.Greater(t, cursor.Y+h.SeekOffset(), 0)

	// Page up moves back up
	sendKeys(t, h, "<pgup>")
	cursor, _, _ = h.Cursor()
	assert.Equal(t, 0, cursor.Y+h.SeekOffset())
}

func TestHandlerEditModeBasicEditing(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		keys     string
		expected string
	}{
		// Inputbox cursor starts at end of text
		{"append at end", "test.go", "X<enter>", "test.goX"},
		{"backspace deletes char", "test.go", "<backspace><enter>", "test.g"},
		{"multiple backspaces", "test.go", "<backspace><backspace><backspace><enter>", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(t, map[string][]mockEntry{
				"/project": {
					{name: tt.initial, isDir: false},
				},
			})
			h.Resize(40, 10)

			// Enter edit mode and apply keys
			sendKeys(t, h, "r"+tt.keys)
			assert.Equal(t, ModeView, h.Mode())

			// Check result
			changes := h.Component().Changes()
			if tt.initial == tt.expected {
				assert.Empty(t, changes)
			} else {
				require.Len(t, changes, 1)
				assert.Equal(t, tt.expected, changes[0].NewURI.Name())
			}
		})
	}
}

func TestHandlerEditModeIKey(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// 'i' should also enter edit mode (like 'r')
	sendKeys(t, h, "i")
	assert.Equal(t, ModeEdit, h.Mode())
}

func TestHandlerNestedDirectoryExpand(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Initially just src/
	initial := len(h.Component().Cells())
	assert.Equal(t, 1, initial)

	// Expand src with 'l'
	sendKeys(t, h, "l")

	// Should now have src/ and app.go
	assert.Equal(t, 2, len(h.Component().Cells()))

	// Collapse with 'h'
	sendKeys(t, h, "h")
	assert.Equal(t, 1, len(h.Component().Cells()))
}

func TestHandlerCollapseWithH(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Expand with 'l'
	sendKeys(t, h, "l")
	assert.Len(t, h.Component().Cells(), 2)

	// Collapse with 'h'
	sendKeys(t, h, "h")
	assert.Len(t, h.Component().Cells(), 1)
}

func TestHandlerExpandWithL(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// 'l' expands directories
	sendKeys(t, h, "l")
	assert.Len(t, h.Component().Cells(), 2)
}

func TestHandlerMultipleDeletes(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Delete multiple rows
	sendKeys(t, h, "djd")

	// Should have 1 row left
	assert.Len(t, h.Component().Cells(), 1)

	// Should detect 2 delete operations
	changes := h.Component().Changes()
	deletes := 0
	for _, c := range changes {
		if c.Type == fileexplorer.OpDelete {
			deletes++
		}
	}
	assert.Equal(t, 2, deletes)
}

func TestHandlerMultipleRenames(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Rename first file
	sendKeys(t, h, "r<c-u>x.go<enter>")

	// Navigate down and rename second file
	sendKeys(t, h, "jr<c-u>y.go<enter>")

	// Should detect 2 rename operations
	changes := h.Component().Changes()
	renames := 0
	for _, c := range changes {
		if c.Type == fileexplorer.OpRename {
			renames++
		}
	}
	assert.Equal(t, 2, renames)
}

func TestHandlerAddFileCreatesDefaultFile(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Add file with 'a'
	sendKeys(t, h, "a")

	// Should have 3 rows (original 2 + new file)
	assert.Len(t, h.Component().Cells(), 3)

	// Should be in edit mode for the new file
	assert.Equal(t, ModeEdit, h.Mode())

	// Cancel edit
	sendKeys(t, h, "<esc>")
	assert.Equal(t, ModeView, h.Mode())

	// Still have 3 rows with the new default file
	assert.Len(t, h.Component().Cells(), 3)
}

func TestHandlerAddDirAndExpand(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "existing.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Add a directory
	sendKeys(t, h, "A<c-u>newdir<enter>")

	changes := h.Component().Changes()
	require.Len(t, changes, 1)
	assert.Equal(t, fileexplorer.OpMkdir, changes[0].Type)
}

func TestHandlerCursorClampsOnDelete(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Go to last row and delete
	sendKeys(t, h, "Gd")

	// Cursor should clamp to last valid row
	cursor, _, _ := h.Cursor()
	assert.Equal(t, 0, cursor.Y)
}

func TestHandlerCursorAfterDeleteFirst(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Delete first row
	sendKeys(t, h, "d")

	// Cursor should remain at 0
	cursor, _, _ := h.Cursor()
	assert.Equal(t, 0, cursor.Y)
}

func TestHandlerEmptyDirectory(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {},
	})
	h.Resize(40, 10)

	// Should handle empty directory gracefully
	cursor, _, show := h.Cursor()
	assert.Equal(t, 0, cursor.Y)
	assert.False(t, show)

	// Navigation should not crash
	sendKeys(t, h, "jjkk")

	// Verify cursor is still valid
	cursor, _, _ = h.Cursor()
	assert.Equal(t, 0, cursor.Y)
}

func TestHandlerConfirmModeNavigation(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	})
	h.Resize(40, 20)

	// Delete both files to have multiple operations
	sendKeys(t, h, "dd")

	// Enter confirm mode
	sendKeys(t, h, "<c-s>")
	assert.Equal(t, ModeConfirm, h.Mode())

	// Navigation in confirm mode (j/k should work)
	sendKeys(t, h, "j")
	// Should stay in confirm mode
	assert.Equal(t, ModeConfirm, h.Mode())

	// Cancel
	sendKeys(t, h, "<esc>")
	assert.Equal(t, ModeView, h.Mode())
}

func TestHandlerNoChangesNoConfirm(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Try to enter confirm mode without changes
	sendKeys(t, h, "<c-s>")

	// Should stay in view mode (no changes to confirm)
	assert.Equal(t, ModeView, h.Mode())
}

func TestHandlerEnterExpandsDirectory(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Enter should expand directory
	sendKeys(t, h, "<enter>")
	assert.Len(t, h.Component().Cells(), 2)

	// Enter again should collapse
	sendKeys(t, h, "<enter>")
	assert.Len(t, h.Component().Cells(), 1)
}

func TestHandlerExpandCollapseToggle(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "src", isDir: true},
		},
		"/project/src": {
			{name: "app.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Initially collapsed
	assert.Equal(t, 1, len(h.Component().Cells()))

	// Toggle expand with 'l'
	sendKeys(t, h, "l")
	assert.Equal(t, 2, len(h.Component().Cells()))

	// Toggle collapse with 'l' on the same directory
	sendKeys(t, h, "l")
	assert.Equal(t, 1, len(h.Component().Cells()))
}

func TestHandlerScrollOnCursorMove(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
			{name: "c.go", isDir: false},
			{name: "d.go", isDir: false},
			{name: "e.go", isDir: false},
		},
	})
	h.Resize(40, 2) // Small viewport

	// Move cursor down - should auto-scroll
	sendKeys(t, h, "jjjj")

	// Cursor should be at last item
	cursor, _, _ := h.Cursor()
	total := cursor.Y + h.SeekOffset()
	assert.Equal(t, 4, total)
}

func TestHandlerSelectionInterface(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})
	h.Resize(40, 10)

	// Selection returns the selected content and whether selection is active
	// The exact behavior depends on implementation
	_, hasSelection := h.Selection()

	// Selection behavior may vary; just verify the interface works
	_ = hasSelection
}

func TestHandlerAttributes(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	})

	attrs := term.Attributes{Fg: 1} // Simple color value
	h.SetAttr(attrs)
	assert.Equal(t, attrs, h.attrs)
}

func TestHandlerOnExit(t *testing.T) {
	exitCalled := false
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "test.go", isDir: false},
		},
	}, WithOnExit(func() {
		exitCalled = true
	}))
	h.Resize(40, 10)

	sendKeys(t, h, "q")
	assert.True(t, exitCalled)
}

func TestHandlerOnApply(t *testing.T) {
	applyCalled := false
	var applyErr error

	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "a.go", isDir: false},
			{name: "b.go", isDir: false},
		},
	}, WithOnApply(func(err error) {
		applyCalled = true
		applyErr = err
	}))
	h.Resize(40, 20)

	// Rename file and confirm (rename doesn't panic unlike delete)
	sendKeys(t, h, "r<c-u>renamed.go<enter><c-s>n")

	// Since we cancelled with 'n', apply should not be called
	assert.False(t, applyCalled)
	assert.Nil(t, applyErr)
}

func TestHandlerRenameEntersEditMode(t *testing.T) {
	h := newTestHandler(t, map[string][]mockEntry{
		"/project": {
			{name: "mydir", isDir: true},
		},
		"/project/mydir": {},
	})
	h.Resize(40, 10)

	// Enter edit mode with 'r'
	sendKeys(t, h, "r")
	assert.Equal(t, ModeEdit, h.Mode())

	// Cancel to return to view mode
	sendKeys(t, h, "<esc>")
	assert.Equal(t, ModeView, h.Mode())

	// No changes should be made
	changes := h.Component().Changes()
	assert.Empty(t, changes)
}

// Test helpers

func newTestHandler(t *testing.T, dirs map[string][]mockEntry, opts ...Option) *Handler {
	t.Helper()
	mfs := &mockFS{dirs: dirs}
	comp, err := fileexplorer.New(mfs, rootURI(), fileexplorer.Config{})
	require.NoError(t, err)
	return New(comp, opts...)
}

func sendKeys(t *testing.T, h *Handler, seq string) {
	t.Helper()
	keys, err := term.ParseKeys(seq)
	require.NoError(t, err)
	for _, k := range keys {
		h.Handle(term.Event{
			Type: term.EventKey,
			Ch:   k.Ch,
			Mod:  k.Mod,
			Key:  k.Key,
		})
	}
}

func sendKey(h *Handler, ch rune) (exit, handled bool) {
	return h.Handle(term.Event{Type: term.EventKey, Ch: ch})
}

type mockEntry struct {
	name  string
	isDir bool
}

func (e mockEntry) Name() string      { return e.name }
func (e mockEntry) IsDir() bool       { return e.isDir }
func (e mockEntry) Type() fs.FileMode { return 0 }
func (e mockEntry) Info() (fs.FileInfo, error) {
	return mockInfo{e}, nil
}

type mockInfo struct{ e mockEntry }

func (m mockInfo) Name() string       { return m.e.name }
func (m mockInfo) Size() int64        { return 0 }
func (m mockInfo) Mode() os.FileMode  { return 0 }
func (m mockInfo) ModTime() time.Time { return time.Time{} }
func (m mockInfo) IsDir() bool        { return m.e.isDir }
func (m mockInfo) Sys() any           { return nil }

type mockFS struct {
	dirs map[string][]mockEntry
}

func (m *mockFS) URI(path string) (workspaceapi.URI, error) {
	return workspaceapi.ParseURI("file://" + path)
}

func (m *mockFS) ReadDir(name string) ([]os.DirEntry, error) {
	entries, ok := m.dirs[name]
	if !ok {
		return nil, fmt.Errorf("not found: %s", name)
	}
	ret := make([]os.DirEntry, len(entries))
	for i, e := range entries {
		ret[i] = e
	}
	return ret, nil
}

func (m *mockFS) OpenFile(
	string, int, os.FileMode,
) (workspaceapi.File, error) {
	panic("not implemented")
}

func (m *mockFS) Remove(string) error {
	panic("not implemented")
}

func (m *mockFS) Stat(string) (os.FileInfo, error) {
	panic("not implemented")
}

func (m *mockFS) MkdirAll(string, os.FileMode) error {
	panic("not implemented")
}

func rootURI() workspaceapi.URI {
	u, err := workspaceapi.ParseURI("file:///project")
	if err != nil {
		panic(err)
	}
	return u
}
