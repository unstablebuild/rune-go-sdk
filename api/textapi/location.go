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

package textapi

import "github.com/unstablebuild/rune-go-sdk/term"

// LocationList is the interface that groups Prev and Next
// location methods to fetch the previous and next item respectively.
//
// Both Prev/Next return false if reached either end or beginning of list.
//
// Current returns the current result. If there are no results in the list
// then Current returns false.
type LocationList interface {
	Current() (Location, bool)
	Prev() (Location, bool)
	Next() (Location, bool)
}

// Location represents a content location in an Editor's buffer.
// From is inclusive and To is exclusive: the range covers cells
// [From, To). A location where From == To is empty (spans zero
// cells) and will not match any cursor position.
type Location struct {
	From, To term.Coordinates
	Attr     term.Attributes
	Message  string
	Icon     string
}

type sliceLocations struct {
	curr int
	in   []Location
}

func (s *sliceLocations) Current() (Location, bool) {
	if s.curr >= len(s.in) || s.curr < 0 {
		return Location{}, false
	}

	return s.in[s.curr], true
}

func (s *sliceLocations) Prev() (Location, bool) {
	if s.curr-1 < 0 {
		return Location{}, false
	}
	s.curr--
	return s.in[s.curr], true
}

func (s *sliceLocations) Next() (Location, bool) {
	if s.curr+1 >= len(s.in) {
		return Location{}, false
	}
	s.curr++
	return s.in[s.curr], true
}

// LocationSlice returns a LocationList based on in
func LocationSlice(in []Location) LocationList {
	return &sliceLocations{in: in}
}
