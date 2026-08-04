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

package graphemecluster

import "github.com/rivo/uniseg"

// StepString returns the first grapheme cluster (user-perceived character) found in
// the given string. It also returns the monospace width of the cluster.
//
// The returned state is opaque and only valid when passed back into StepString
// for the rest of the same string; pass -1 when starting a new string.
//
// See uniseg.FirstGraphemeClusterInString for more details.
func StepString(str string, state int) (
	cluster, rest string, width uint8, newState int,
) {
	// Deliberately not uniseg.StepString: it additionally runs word-,
	// sentence- and line-break state machines whose results this wrapper
	// has no way to return, roughly tripling the per-rune cost.
	var w int
	cluster, rest, w, newState = uniseg.FirstGraphemeClusterInString(str, state)
	width = graphemeClusterWidth(cluster, w)
	return
}

// StringWidth returns the monospace width for the given string, that is, the
// number of same-size cells to be occupied by the string.
func StringWidth(s string) (width int) {
	state := -1
	var w uint8
	for len(s) > 0 {
		_, s, w, state = StepString(s, state)
		width += int(w)
	}
	return
}

func graphemeClusterWidth(cluster string, unisegWidth int) uint8 {
	if unisegWidth == 0 || unisegWidth > 1 || len(cluster) == 0 /* don't trust uniseg */ {
		return uint8(unisegWidth)
	}

	switch cluster {
	// NOTE: this is just the icons that we're interested in
	// but we should add the full list of nerd font icons
	// width width > 1.
	case "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", " ",
		"󱫆", "", "", "", "", "", "", "", "", "", "", "",
		"", "󰌾", "󰗻", "󱄋", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "󰝤", "", "", "󰅗", "",
		"", "󱡓", "󰄳", "󰗠", "󰐌", "󰏥", "󰄝", "", "󰟓", "", "󰙅", "", "", "󱙺", "",
		"", "":
		return 2
	default:
		return uint8(unisegWidth)
	}
}
