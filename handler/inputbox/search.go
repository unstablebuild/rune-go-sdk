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

func (ib *Handler) enterSearch() {
	if len(ib.history) == 0 {
		return
	}
	ib.searching = true
	ib.searchQuery = ib.searchQuery[:0]
	ib.searchIdx = len(ib.history) - 1
	ib.searchFailed = false
	ib.searchOrigLine = string(ib.text)
	ib.searchOrigPos = ib.cursor
}

func (ib *Handler) searchAddChar(ch rune) {
	ib.searchQuery = append(ib.searchQuery, ch)
	ib.performSearch()
}

func (ib *Handler) searchBackspace() {
	if len(ib.searchQuery) == 0 {
		return
	}
	ib.searchQuery = ib.searchQuery[:len(ib.searchQuery)-1]
	ib.searchIdx = len(ib.history) - 1
	ib.searchFailed = false
	ib.performSearch()
}

func (ib *Handler) searchNext() {
	if ib.searchIdx > 0 {
		ib.searchIdx--
		ib.performSearch()
	}
}

func (ib *Handler) performSearch() {
	query := string(ib.searchQuery)
	for i := ib.searchIdx; i >= 0; i-- {
		if strings.Contains(ib.history[i], query) {
			ib.searchIdx = i
			ib.searchFailed = false
			return
		}
	}
	ib.searchFailed = true
}

func (ib *Handler) acceptSearch() {
	ib.searching = false
	query := string(ib.searchQuery)
	for i := ib.searchIdx; i >= 0; i-- {
		if strings.Contains(ib.history[i], query) {
			ib.setLine(ib.history[i], len([]rune(ib.history[i])))
			return
		}
	}
}

func (ib *Handler) cancelSearch() {
	ib.searching = false
	ib.setLine(ib.searchOrigLine, ib.searchOrigPos)
}
