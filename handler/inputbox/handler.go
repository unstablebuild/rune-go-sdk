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

package inputbox

import (
	"io"

	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/handler"
	"github.com/unstablebuild/rune-go-sdk/mouse"
	"github.com/unstablebuild/rune-go-sdk/term"
	"github.com/unstablebuild/tcell/v3"
)

// Handler implements a multi-line text input handler with wrapping,
// text selection, word-wise cursor movement, history navigation,
// reverse search, and tab completion. Note that text selection is
// limited by your terminal's ability to handle shift+arrow keys.
type Handler struct {
	text        []rune
	cursor      int
	voffset     int
	width       int
	height      int
	attrs       term.Attributes
	selAnchor   int // -1 means no selection
	placeholder *component.ResponsiveString

	// Prompt
	prompt      []rune
	promptWidth int

	// History
	history       []string
	historyPos    int
	historyEnd    string
	historyStale  bool
	historyPrefix string

	// Search
	searching      bool
	searchQuery    []rune
	searchIdx      int
	searchFailed   bool
	searchOrigLine string
	searchOrigPos  int

	// Completion
	completer              WordCompleter
	tabStyle               TabStyle
	completions            []string
	completionIdx          int
	completionOn           bool
	completionPrintPending bool
	compHead               string
	compTail               string
	compOriginal           string

	m *mouse.Mouse

	// Highlight range
	hlStart int
	hlEnd   int
	hlAttr  term.Attributes
	hasHL   bool

	// State
	ctrlCAborts bool
	done        bool
	aborted     bool
	eof         bool

	// Rendering
	redact     bool
	redactRune rune
}

// defaultRedactRune is the glyph used to mask buffer contents when redact
// rendering is enabled and no explicit rune has been configured.
const defaultRedactRune = '*'

var _ handler.WithAttributesResponsiveFloating = (*Handler)(nil)

// New creates and initializes a new input box with default attributes.
func New(opts ...Option) *Handler {
	ret := &Handler{
		text:      make([]rune, 0, 64),
		selAnchor: -1,
	}
	for _, o := range opts {
		o(ret)
	}
	if ret.redact && ret.redactRune == 0 {
		ret.redactRune = defaultRedactRune
	}
	ret.historyPos = len(ret.history)
	ret.m = mouse.New(&mouseDelegate{ret})
	return ret
}

// SetAttr satisfies WithAttributes.
func (ib *Handler) SetAttr(attr term.Attributes) (ret term.Attributes) {
	ret = ib.attrs
	ib.attrs = attr
	ib.updatePlaceholderBackground()
	return
}

// SetHighlight applies attr to the text range [start, end).
// The prompt and characters outside the range are unaffected.
func (ib *Handler) SetHighlight(start, end int, attr term.Attributes) {
	ib.hlStart = start
	ib.hlEnd = end
	ib.hlAttr = attr
	ib.hasHL = true
}

// ClearHighlight removes any active highlight range.
func (ib *Handler) ClearHighlight() {
	ib.hasHL = false
}

// Text returns the current text content of the input box.
func (ib *Handler) Text() string {
	return string(ib.text)
}

// Result returns the input text or an error if input
// was aborted or reached EOF.
func (ib *Handler) Result() (string, error) {
	if ib.eof {
		return "", io.EOF
	}
	if ib.aborted {
		return "", ErrAborted
	}
	return string(ib.text), nil
}

// Reset clears the exit state for the next prompt.
func (ib *Handler) Reset() {
	ib.text = ib.text[:0]
	ib.cursor = 0
	ib.selAnchor = -1
	ib.done = false
	ib.aborted = false
	ib.eof = false
	ib.historyPos = len(ib.history)
	ib.historyEnd = ""
	ib.historyStale = true
	ib.historyPrefix = ""
	ib.clearCompletion()
	ib.searching = false
	ib.updateVoffset()
}

// Clear resets the input box to empty state.
func (ib *Handler) Clear() {
	ib.text = ib.text[:0]
	ib.cursor = 0
	ib.selAnchor = -1
	ib.updateVoffset()
}

// Resize satisfies tui.Handler
func (ib *Handler) Resize(width, height int) {
	ib.width = width
	ib.height = height
	if ib.placeholder != nil {
		ib.placeholder.Resize(width, height)
		ib.updatePlaceholderBackground()
	}
	ib.updateVoffset()
}

// firstLineChars returns the number of text characters
// that fit on the first line (after the prompt).
func (ib *Handler) firstLineChars() int {
	return max(ib.width-ib.promptWidth, 1)
}

// textPosToCoords maps a text index to screen x,y
// coordinates, accounting for prompt width on line 0.
func (ib *Handler) textPosToCoords(pos int) (x, y int) {
	if ib.width <= 0 {
		return 0, 0
	}
	flc := ib.firstLineChars()
	if pos < flc {
		return ib.promptWidth + pos, 0
	}
	pos -= flc
	y = 1 + pos/ib.width
	x = pos % ib.width
	return x, y
}

// totalTextLines returns the total number of screen
// lines needed to display the current text and cursor.
func (ib *Handler) totalTextLines() int {
	if ib.width <= 0 {
		return 0
	}
	if len(ib.text) == 0 {
		return 1
	}
	flc := ib.firstLineChars()
	if len(ib.text) <= flc {
		// Check if cursor is at the end and wraps.
		if ib.cursor == len(ib.text) && len(ib.text) == flc {
			return 2
		}
		return 1
	}
	remaining := len(ib.text) - flc
	lines := 1 + (remaining+ib.width-1)/ib.width
	// If cursor is at the end and the last line is full,
	// we need an extra line for the cursor.
	_, cy := ib.textPosToCoords(len(ib.text))
	if ib.cursor == len(ib.text) && cy >= lines {
		lines = cy + 1
	}
	return lines
}

// Height satisfies Responsive. Returns the number of lines needed to
// display all text content and cursor given the specified width.
func (ib *Handler) Height(width int) int {
	if width <= 0 {
		return 0
	}
	if ib.promptWidth == 0 {
		return ib.heightNoPrompt(width)
	}
	// Prompt-aware calculation.
	flc := max(width-ib.promptWidth, 1)
	n := len(ib.text)
	if n == 0 {
		h := 1
		if ib.completionOn && len(ib.completions) > 0 {
			h += completionGridHeight(ib.completions, width)
		}
		return h
	}
	var lines int
	if n <= flc {
		lines = 1
		if n == flc && ib.cursor == n {
			lines = 2
		}
	} else {
		remaining := n - flc
		lines = 1 + (remaining+width-1)/width
		// Extra line if cursor at end on a full last line.
		if ib.cursor == n {
			afterFirst := n - flc
			if afterFirst > 0 && afterFirst%width == 0 {
				lines++
			}
		}
	}
	if ib.completionOn && len(ib.completions) > 0 {
		lines += completionGridHeight(ib.completions, width)
	}
	return lines
}

func (ib *Handler) heightNoPrompt(width int) int {
	if len(ib.text) == 0 {
		h := 1
		if ib.completionOn && len(ib.completions) > 0 {
			h += completionGridHeight(ib.completions, width)
		}
		return h
	}
	textLines := (len(ib.text) + width - 1) / width
	if len(ib.text)%width == 0 && ib.cursor == len(ib.text) {
		textLines++
	}
	if ib.completionOn && len(ib.completions) > 0 {
		textLines += completionGridHeight(ib.completions, width)
	}
	return textLines
}

// Dimensions satisfies component.Floating. It returns
// the width needed to render the entire input box content
// on a single line (height is always 1).
func (ib *Handler) Dimensions() (int, int) {
	return ib.promptWidth + len(ib.text) + 1, 1
}

// Draw satisfies tui.Component.
func (ib *Handler) Draw(w term.Writer) {
	if ib.height == 0 || ib.width == 0 {
		return
	}

	for y := range ib.height {
		for x := range ib.width {
			w.UnionAttributes(term.Coordinates{X: x, Y: y}, ib.attrs)
		}
	}

	if ib.searching {
		ib.drawSearch(w)
		return
	}

	if len(ib.text) == 0 && ib.promptWidth == 0 && ib.placeholder != nil {
		ib.placeholder.Draw(w)
		return
	}

	if ib.promptWidth == 0 {
		ib.drawNoPrompt(w)
	} else {
		ib.drawWithPrompt(w)
	}

	if ib.completionOn && len(ib.completions) > 0 {
		ib.drawCompletions(w)
	}
}

func (ib *Handler) drawNoPrompt(w term.Writer) {
	startIdx := ib.voffset * ib.width
	endIdx := min(startIdx+ib.height*ib.width, len(ib.text))

	selStart, selEnd := ib.selectionRange()

	var x, y int
	for i := startIdx; i < endIdx; i++ {
		attrs := ib.attrs
		if ib.hasHL && i >= ib.hlStart && i < ib.hlEnd {
			attrs = ib.hlAttr
		}
		if i >= selStart && i < selEnd {
			attrs.Attrs |= tcell.AttrReverse
		}
		ch := ib.text[i]
		if ib.redact {
			ch = ib.redactRune
		}
		w.SetCell(term.Coordinates{X: x, Y: y}, term.Cell{
			Ch:         ch,
			Attributes: attrs,
			Width:      1,
		})
		x++
		if x < ib.width {
			continue
		}
		x = 0
		y++
		if y >= ib.height {
			break
		}
	}
}

func (ib *Handler) drawWithPrompt(w term.Writer) {
	// Draw prompt on first visible line.
	if ib.voffset == 0 {
		for i, r := range ib.prompt {
			if i >= ib.width {
				break
			}
			w.SetCell(term.Coordinates{X: i, Y: 0}, term.Cell{
				Ch:         r,
				Attributes: ib.attrs,
				Width:      1,
			})
		}
	}

	selStart, selEnd := ib.selectionRange()
	flc := ib.firstLineChars()

	for i := range len(ib.text) {
		sx, sy := ib.textPosToCoords(i)
		sy -= ib.voffset
		if sy < 0 {
			continue
		}
		if sy >= ib.height {
			break
		}
		attrs := ib.attrs
		if ib.hasHL && i >= ib.hlStart && i < ib.hlEnd {
			attrs = ib.hlAttr
		}
		if i >= selStart && i < selEnd {
			attrs.Attrs |= tcell.AttrReverse
		}
		_ = flc
		ch := ib.text[i]
		if ib.redact {
			ch = ib.redactRune
		}
		w.SetCell(term.Coordinates{X: sx, Y: sy}, term.Cell{
			Ch:         ch,
			Attributes: attrs,
			Width:      1,
		})
	}
}

func (ib *Handler) drawSearch(w term.Writer) {
	line := "(reverse-i-search)'" + string(ib.searchQuery) + "': " + ib.searchMatchText()
	runes := []rune(line)
	for i, r := range runes {
		if i >= ib.width {
			break
		}
		w.SetCell(
			term.Coordinates{X: i, Y: 0},
			term.Cell{
				Ch: r, Attributes: ib.attrs, Width: 1,
			},
		)
	}
}

func (ib *Handler) searchMatchText() string {
	if ib.searchFailed || len(ib.history) == 0 {
		return ""
	}
	if ib.searchIdx >= 0 && ib.searchIdx < len(ib.history) {
		return ib.history[ib.searchIdx]
	}
	return ""
}

func (ib *Handler) drawCompletions(w term.Writer) {
	if ib.width == 0 {
		return
	}
	// Find the first row for completions: after text lines.
	textLines := ib.totalTextLines()
	startY := textLines - ib.voffset

	maxLen := 0
	for _, s := range ib.completions {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	colWidth := min(maxLen+2, ib.width)
	cols := max(ib.width/colWidth, 1)

	y := startY
	for i, s := range ib.completions {
		if y >= ib.height {
			break
		}
		col := i % cols
		x := col * colWidth
		runes := []rune(s)
		for j, r := range runes {
			cx := x + j
			if cx >= ib.width {
				break
			}
			w.SetCell(
				term.Coordinates{X: cx, Y: y},
				term.Cell{
					Ch:         r,
					Attributes: ib.attrs,
					Width:      1,
				},
			)
		}
		if col == cols-1 || i == len(ib.completions)-1 {
			y++
		}
	}
}

// Handle satisfies tui.Handler
func (ib *Handler) Handle(ev term.Event) (exit, handled bool) {
	if ev.Type == term.EventMouse {
		return ib.m.Handle(ev)
	}

	if ev.Type != term.EventKey {
		return false, false
	}

	if ib.searching {
		return ib.handleSearch(ev)
	}

	return ib.handleNormal(ev)
}

func (ib *Handler) handleSearch(ev term.Event) (exit, handled bool) {
	switch {
	case ev.Key == term.KeyEnter && ev.Mod == 0:
		ib.acceptSearch()
		ib.done = true
		return true, true

	case ev.Key == term.KeyEsc && ev.Mod == 0,
		ev.Ch == 'g' && ev.Mod == term.ModCtrl,
		ev.Ch == 'c' && ev.Mod == term.ModCtrl:
		ib.cancelSearch()
		return false, true

	case ev.Key == term.KeyTab && ev.Mod == 0:
		ib.acceptSearch()
		return false, true

	case ev.Key == term.KeyBackspace && ev.Mod == 0,
		ev.Ch == 'h' && ev.Mod == term.ModCtrl:
		ib.searchBackspace()
		return false, true

	case ev.Ch == 'r' && ev.Mod == term.ModCtrl:
		ib.searchNext()
		return false, true

	case ev.Ch != 0 && ev.Mod == 0:
		ib.searchAddChar(ev.Ch)
		return false, true

	case ev.Key == term.KeySpace && ev.Mod == 0:
		ib.searchAddChar(' ')
		return false, true
	}
	return false, false
}

func (ib *Handler) handleNormal(ev term.Event) (exit, handled bool) {
	clearComp := ib.completionOn || ib.completionPrintPending

	// Selection-replacing character input: if there's an
	// active selection and a printable char is typed
	// (without ctrl/alt), delete the selection first.
	if ib.hasSelection() && ev.Ch != 0 &&
		ev.Mod != term.ModCtrl && ev.Mod != term.ModAlt {
		ib.deleteSelection()
	}

	switch {
	case ev.Key == term.KeyEnter && ev.Mod == 0:
		ib.clearCompletion()
		ib.done = true
		return true, true

	case ev.Key == term.KeyTab && ev.Mod == 0:
		ib.handleTab(false)
		return false, true

	case ev.Key == term.KeyTab && ev.Mod == term.ModShift:
		ib.handleTab(true)
		return false, true

	case ev.Key == term.KeyEsc && ev.Mod == 0:
		if clearComp {
			ib.cancelCompletion()
			return false, true
		}
		return true, true

	case ev.Ch == 'c' && ev.Mod == term.ModCtrl:
		ib.clearCompletion()
		if ib.ctrlCAborts {
			ib.aborted = true
			return true, true
		}
		ib.Clear()
		return false, true

	case ev.Ch == 'd' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		if len(ib.text) == 0 {
			ib.eof = true
			return true, true
		}
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor < len(ib.text) {
			ib.text = append(
				ib.text[:ib.cursor],
				ib.text[ib.cursor+1:]...,
			)
		}
		ib.updateVoffset()
		return false, true

	case ev.Ch == 'l' && ev.Mod == term.ModCtrl:
		return false, true

	case ev.Ch == 'r' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.enterSearch()
		return false, true

	case ev.Key == term.KeyArrowUp && ev.Mod == 0,
		ev.Ch == 'p' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.historyUp()
		ib.updateVoffset()
		return false, true

	case ev.Key == term.KeyArrowDown && ev.Mod == 0,
		ev.Ch == 'n' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.historyDown()
		ib.updateVoffset()
		return false, true
	}

	defer ib.updateVoffset()

	switch {
	// shift+ctrl/alt word selection
	case ev.Key == term.KeyArrowLeft &&
		ev.Mod == term.ModCtrlShift,
		ev.Key == term.KeyArrowLeft &&
			ev.Mod == term.ModAltShift:
		if clearComp {
			ib.clearCompletion()
		}
		ib.startSelection()
		ib.moveWordLeft()
		return false, true

	case ev.Key == term.KeyArrowRight &&
		ev.Mod == term.ModCtrlShift,
		ev.Key == term.KeyArrowRight &&
			ev.Mod == term.ModAltShift:
		if clearComp {
			ib.clearCompletion()
		}
		ib.startSelection()
		ib.moveWordRight()
		return false, true

	// ctrl/alt word movement
	case ev.Key == term.KeyArrowLeft &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyArrowLeft &&
			ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.moveWordLeft()
		return false, true

	case ev.Key == term.KeyArrowRight &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyArrowRight &&
			ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.moveWordRight()
		return false, true

	// word deletion
	case ev.Ch == 'w' && ev.Mod == term.ModCtrl,
		ev.Key == term.KeyBackspace &&
			ev.Mod == term.ModCtrl,
		ev.Key == term.KeyBackspace &&
			ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		if ib.hasSelection() {
			ib.deleteSelection()
		} else {
			ib.deleteWordBackward()
		}
		return false, true

	case ev.Key == term.KeyDelete &&
		ev.Mod == term.ModCtrl,
		ev.Key == term.KeyDelete &&
			ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		if ib.hasSelection() {
			ib.deleteSelection()
		} else {
			ib.deleteWordForward()
		}
		return false, true

	// shift+arrow char selection
	case ev.Key == term.KeyArrowLeft &&
		ev.Mod == term.ModShift:
		if clearComp {
			ib.clearCompletion()
		}
		ib.startSelection()
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyArrowRight &&
		ev.Mod == term.ModShift:
		if clearComp {
			ib.clearCompletion()
		}
		ib.startSelection()
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	// shift+home/end selection
	case ev.Key == term.KeyHome &&
		ev.Mod == term.ModShift:
		if clearComp {
			ib.clearCompletion()
		}
		ib.startSelection()
		ib.cursor = 0
		return false, true

	case ev.Key == term.KeyEnd &&
		ev.Mod == term.ModShift:
		if clearComp {
			ib.clearCompletion()
		}
		ib.startSelection()
		ib.cursor = len(ib.text)
		return false, true

	case ev.Key == term.KeyBackspace && ev.Mod == 0,
		ev.Ch == 'h' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor > 0 {
			ib.text = append(
				ib.text[:ib.cursor-1],
				ib.text[ib.cursor:]...,
			)
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyDelete && ev.Mod == 0:
		if clearComp {
			ib.clearCompletion()
		}
		if ib.hasSelection() {
			ib.deleteSelection()
		} else if ib.cursor < len(ib.text) {
			ib.text = append(
				ib.text[:ib.cursor],
				ib.text[ib.cursor+1:]...,
			)
		}
		return false, true

	case ev.Key == term.KeyArrowLeft && ev.Mod == 0:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Key == term.KeyArrowRight && ev.Mod == 0:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Key == term.KeyHome && ev.Mod == 0,
		ev.Ch == 'a' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.cursor = 0
		return false, true

	case ev.Key == term.KeyEnd && ev.Mod == 0,
		ev.Ch == 'e' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.cursor = len(ib.text)
		return false, true

	case ev.Ch == 'b' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		if ib.cursor > 0 {
			ib.cursor--
		}
		return false, true

	case ev.Ch == 'f' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		if ib.cursor < len(ib.text) {
			ib.cursor++
		}
		return false, true

	case ev.Ch == 'k' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		if ib.cursor < len(ib.text) {
			ib.text = ib.text[:ib.cursor]
		}
		return false, true

	case ev.Ch == 'u' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.Clear()
		return false, true

	case ev.Ch == 'b' && ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.moveWordLeft()
		return false, true

	case ev.Ch == 'f' && ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.moveWordRight()
		return false, true

	case ev.Ch == 'd' && ev.Mod == term.ModAlt:
		if clearComp {
			ib.clearCompletion()
		}
		if ib.hasSelection() {
			ib.deleteSelection()
		} else {
			ib.deleteWordForward()
		}
		return false, true

	case ev.Ch == 't' && ev.Mod == term.ModCtrl:
		if clearComp {
			ib.clearCompletion()
		}
		ib.clearSelection()
		ib.transpose()
		return false, true

	case ev.Key == term.KeySpace && ev.Mod == 0:
		if clearComp {
			ib.clearCompletion()
		}
		ib.historyStale = true
		ib.historyPrefix = ""
		ib.clearSelection()
		ib.text = append(
			ib.text[:ib.cursor],
			append([]rune{' '}, ib.text[ib.cursor:]...)...,
		)
		ib.cursor++
		return false, true

	case ev.Ch != 0 && ev.Mod == 0:
		if clearComp {
			ib.clearCompletion()
		}
		ib.historyStale = true
		ib.historyPrefix = ""
		ib.clearSelection()
		ib.text = append(
			ib.text[:ib.cursor],
			append([]rune{ev.Ch}, ib.text[ib.cursor:]...)...,
		)
		ib.cursor++
		return false, true
	}

	return false, false
}

// Cursor satisfies tui.Handler
func (ib *Handler) Cursor() (term.Coordinates, term.CursorStyle, bool) {
	if ib.searching {
		prefix := []rune("(reverse-i-search)'")
		x := len(prefix) + len(ib.searchQuery)
		return term.Coordinates{X: x, Y: 0},
			term.CursorStyleSteadyBar, true
	}
	if ib.promptWidth == 0 {
		cursorLine := ib.cursor / ib.width
		cursorCol := ib.cursor % ib.width
		cursorY := cursorLine - ib.voffset
		return term.Coordinates{X: cursorCol, Y: cursorY},
			term.CursorStyleSteadyBar, true
	}
	x, y := ib.textPosToCoords(ib.cursor)
	y -= ib.voffset
	return term.Coordinates{X: x, Y: y},
		term.CursorStyleSteadyBar, true
}

// Selection satisfies tui.Handler
func (ib *Handler) Selection() (string, bool) {
	if !ib.hasSelection() {
		return "", false
	}
	start, end := ib.selectionRange()
	return string(ib.text[start:end]), true
}

func (ib *Handler) hasSelection() bool {
	return ib.selAnchor >= 0 && ib.selAnchor != ib.cursor
}

func (ib *Handler) selectionRange() (int, int) {
	if ib.selAnchor < 0 {
		return 0, 0
	}
	n := len(ib.text)
	start := min(ib.selAnchor, ib.cursor)
	end := max(ib.selAnchor, ib.cursor)
	start = max(0, min(start, n))
	end = max(start, min(end, n))
	return start, end
}

func (ib *Handler) clearSelection() {
	ib.selAnchor = -1
}

// startSelection sets the anchor to the current cursor position
// if no selection is active yet.
func (ib *Handler) startSelection() {
	if ib.selAnchor < 0 {
		ib.selAnchor = ib.cursor
	}
}

func (ib *Handler) deleteSelection() {
	if !ib.hasSelection() {
		return
	}
	start, end := ib.selectionRange()
	ib.text = append(ib.text[:start], ib.text[end:]...)
	ib.cursor = start
	ib.selAnchor = -1
}

func isWordChar(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func (ib *Handler) moveWordLeft() {
	for ib.cursor > 0 && !isWordChar(ib.text[ib.cursor-1]) {
		ib.cursor--
	}
	for ib.cursor > 0 && isWordChar(ib.text[ib.cursor-1]) {
		ib.cursor--
	}
}

func (ib *Handler) moveWordRight() {
	n := len(ib.text)
	for ib.cursor < n && !isWordChar(ib.text[ib.cursor]) {
		ib.cursor++
	}
	for ib.cursor < n && isWordChar(ib.text[ib.cursor]) {
		ib.cursor++
	}
}

func (ib *Handler) deleteWordBackward() {
	start := ib.cursor
	for start > 0 && !isWordChar(ib.text[start-1]) {
		start--
	}
	for start > 0 && isWordChar(ib.text[start-1]) {
		start--
	}
	ib.text = append(ib.text[:start], ib.text[ib.cursor:]...)
	ib.cursor = start
}

func (ib *Handler) deleteWordForward() {
	end := ib.cursor
	n := len(ib.text)
	for end < n && !isWordChar(ib.text[end]) {
		end++
	}
	for end < n && isWordChar(ib.text[end]) {
		end++
	}
	ib.text = append(ib.text[:ib.cursor], ib.text[end:]...)
}

func (ib *Handler) transpose() {
	if ib.cursor < 2 {
		return
	}
	ib.text[ib.cursor-2], ib.text[ib.cursor-1] =
		ib.text[ib.cursor-1], ib.text[ib.cursor-2]
}

// setLine replaces the text and cursor position.
func (ib *Handler) setLine(text string, pos int) {
	ib.text = []rune(text)
	ib.cursor = max(pos, 0)
	ib.cursor = min(ib.cursor, len(ib.text))
	ib.clearSelection()
	ib.updateVoffset()
}

// updateVoffset adjusts the vertical offset to keep the cursor visible
// and show as much text as possible.
func (ib *Handler) updateVoffset() {
	if ib.width <= 0 || ib.height <= 0 {
		return
	}

	if ib.promptWidth == 0 {
		ib.updateVoffsetNoPrompt()
		return
	}

	_, cursorLine := ib.textPosToCoords(ib.cursor)
	textLines := ib.totalTextLines()

	if textLines <= ib.height {
		ib.voffset = 0
		return
	}

	if cursorLine < ib.voffset {
		ib.voffset = cursorLine
	} else if cursorLine >= ib.voffset+ib.height {
		ib.voffset = cursorLine - ib.height + 1
	}
}

func (ib *Handler) updateVoffsetNoPrompt() {
	cursorLine := ib.cursor / ib.width
	totalLines := cursorLine + 1
	if len(ib.text) > 0 {
		textLines := (len(ib.text) + ib.width - 1) / ib.width
		if textLines > totalLines {
			totalLines = textLines
		}
	}

	if totalLines <= ib.height {
		ib.voffset = 0
		return
	}

	if cursorLine < ib.voffset {
		ib.voffset = cursorLine
	} else if cursorLine >= ib.voffset+ib.height {
		ib.voffset = cursorLine - ib.height + 1
	}

	if ib.voffset > 0 && ib.voffset*ib.width >= len(ib.text) {
		if len(ib.text) > 0 {
			ib.voffset = (len(ib.text) - 1) / ib.width
		} else {
			ib.voffset = 0
		}
		if cursorLine >= ib.voffset+ib.height {
			ib.voffset = cursorLine - ib.height + 1
		}
	}
}

func (ib *Handler) updatePlaceholderBackground() {
	if ib.placeholder != nil {
		_ = ib.placeholder.SetAttr(term.Attributes{Bg: ib.attrs.Bg})
	}
}
