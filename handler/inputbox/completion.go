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

func (ib *Handler) handleTab(reverse bool) {
	if ib.completer == nil {
		return
	}

	if ib.completionOn {
		if ib.completionPrintPending {
			// TabPrints second tab: apply common prefix.
			ib.handleTabPrintsPrefix()
			return
		}
		ib.cycleCompletion(reverse)
		return
	}

	line := string(ib.text)
	pos := ib.cursor
	head, candidates, tail := ib.completer(line, pos)
	if len(candidates) == 0 {
		return
	}

	ib.compHead = head
	ib.compTail = tail
	ib.compOriginal = line

	if len(candidates) == 1 {
		ib.applyCompletion(candidates[0])
		return
	}

	if ib.tabStyle == TabPrints {
		ib.handleTabPrints(candidates)
		return
	}

	ib.completions = candidates
	ib.completionIdx = 0
	if reverse {
		ib.completionIdx = len(candidates) - 1
	}
	ib.completionOn = true
	ib.applyCompletion(ib.completions[ib.completionIdx])
}

func (ib *Handler) handleTabPrints(candidates []string) {
	// First tab: show the completion grid without changing text.
	ib.completions = candidates
	ib.completionOn = true
	ib.completionPrintPending = true
}

func (ib *Handler) handleTabPrintsPrefix() {
	// Second tab: apply first candidate, start cycling.
	ib.completionPrintPending = false
	ib.completionIdx = 0
	ib.applyCompletion(ib.completions[0])
}

func (ib *Handler) cycleCompletion(reverse bool) {
	if len(ib.completions) == 0 {
		return
	}
	if reverse {
		ib.completionIdx--
		if ib.completionIdx < 0 {
			ib.completionIdx = len(ib.completions) - 1
		}
	} else {
		ib.completionIdx++
		if ib.completionIdx >= len(ib.completions) {
			ib.completionIdx = 0
		}
	}
	ib.applyCompletion(ib.completions[ib.completionIdx])
}

func (ib *Handler) applyCompletion(candidate string) {
	line := ib.compHead + candidate + ib.compTail
	pos := len([]rune(ib.compHead)) + len([]rune(candidate))
	ib.setLine(line, pos)
}

func (ib *Handler) cancelCompletion() {
	if !ib.completionOn {
		return
	}
	ib.setLine(ib.compOriginal, len([]rune(ib.compOriginal)))
	ib.clearCompletion()
}

func (ib *Handler) clearCompletion() {
	ib.completionOn = false
	ib.completionIdx = 0
	ib.completions = nil
	ib.compHead = ""
	ib.compTail = ""
	ib.compOriginal = ""
	ib.completionPrintPending = false
}

func completionGridHeight(completions []string, width int) int {
	if width == 0 || len(completions) == 0 {
		return 0
	}
	maxLen := 0
	for _, s := range completions {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	colWidth := min(maxLen+2, width)
	cols := max(width/colWidth, 1)
	rows := (len(completions) + cols - 1) / cols
	return rows
}
