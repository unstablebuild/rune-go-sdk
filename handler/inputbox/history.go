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

import "strings"

// AppendHistory adds an item to the history.
func (ib *Handler) AppendHistory(item string) {
	ib.history = append(ib.history, item)
	ib.historyPos = len(ib.history)
}

// ClearHistory removes all history entries.
func (ib *Handler) ClearHistory() {
	ib.history = nil
	ib.historyPos = 0
	ib.historyEnd = ""
}

// SetHistory replaces the history with items.
func (ib *Handler) SetHistory(items []string) {
	ib.history = make([]string, len(items))
	copy(ib.history, items)
	ib.historyPos = len(ib.history)
	ib.historyEnd = ""
}

func (ib *Handler) historyUp() {
	if len(ib.history) == 0 {
		return
	}
	if ib.historyPos == len(ib.history) {
		ib.historyEnd = string(ib.text)
	}
	if ib.historyPrefix != "" {
		ib.historyPrefixUp()
		return
	}
	if ib.historyPos > 0 {
		ib.historyPos--
		ib.setLine(ib.history[ib.historyPos], len([]rune(ib.history[ib.historyPos])))
	}
}

func (ib *Handler) historyDown() {
	if len(ib.history) == 0 {
		return
	}
	if ib.historyPrefix != "" {
		ib.historyPrefixDown()
		return
	}
	if ib.historyPos < len(ib.history)-1 {
		ib.historyPos++
		ib.setLine(ib.history[ib.historyPos], len([]rune(ib.history[ib.historyPos])))
	} else if ib.historyPos == len(ib.history)-1 {
		ib.historyPos = len(ib.history)
		ib.setLine(ib.historyEnd, len([]rune(ib.historyEnd)))
	}
}

func (ib *Handler) historyPrefixUp() {
	for i := ib.historyPos - 1; i >= 0; i-- {
		if strings.HasPrefix(ib.history[i], ib.historyPrefix) {
			ib.historyPos = i
			ib.setLine(ib.history[i], len([]rune(ib.history[i])))
			return
		}
	}
}

func (ib *Handler) historyPrefixDown() {
	for i := ib.historyPos + 1; i < len(ib.history); i++ {
		if strings.HasPrefix(ib.history[i], ib.historyPrefix) {
			ib.historyPos = i
			ib.setLine(ib.history[i], len([]rune(ib.history[i])))
			return
		}
	}
	if ib.historyPos < len(ib.history) {
		ib.historyPos = len(ib.history)
		ib.setLine(ib.historyEnd, len([]rune(ib.historyEnd)))
	}
}
