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

package term

import "sort"

// Coordinates represent a point in a 2D space.
type Coordinates struct {
	X, Y int
}

// Range is represents a selection in a 2D space.
type Range struct {
	Start Coordinates
	End   Coordinates
}

// CoordinatesBlockSort sorts a pair of coordinates (from/to) such that:
//
//		┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
//		│ f    ││ t    ││    f ││    t ││ t  f ││ f  t │
//		│    t ││    f ││ t    ││ f    ││      ││      │
//		└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
//	      |       |        |       |       |       |
//	      v       v        v       v       v       v
//		┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
//		│ f    ││ f    ││ f    ││ f    ││ f  t ││ f  t │
//		│    t ││    t ││    t ││    t ││      ││      │
//		└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
func CoordinatesBlockSort(from Coordinates, to Coordinates) (
	Coordinates, Coordinates,
) {
	if from.X > to.X {
		temp := from.X
		from.X = to.X
		to.X = temp
	}
	if from.Y > to.Y {
		temp := from.Y
		from.Y = to.Y
		to.Y = temp
	}
	return from, to
}

// CoordinatesSort sorts a pair of coordinates (from/to) such that:
//
//		┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
//		│ f    ││ t    ││    f ││    t ││ t  f ││ f  t │
//		│    t ││    f ││ t    ││ f    ││      ││      │
//		└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
//	      |       |        |       |       |       |
//	      v       v        v       v       v       v
//		┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
//		│ f    ││ f    ││    f ││    f ││ f  t ││ f  t │
//		│    t ││    t ││ t    ││ t    ││      ││      │
//		└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
func CoordinatesSort(from Coordinates, to Coordinates) (
	Coordinates, Coordinates,
) {
	if from.Y > to.Y {
		temp := from.Y
		from.Y = to.Y
		to.Y = temp

		temp = from.X
		from.X = to.X
		to.X = temp
	} else if from.Y == to.Y && from.X > to.X {
		temp := from.X
		from.X = to.X
		to.X = temp
	}
	return from, to
}

// CoordinatesDiff subtracts a from b.
func CoordinatesDiff(a, b Coordinates) Coordinates {
	return Coordinates{
		Y: a.Y - b.Y,
		X: a.X - b.X,
	}
}

// CoordinatesSum adds a to b.
func CoordinatesSum(a, b Coordinates) Coordinates {
	return Coordinates{
		Y: a.Y + b.Y,
		X: a.X + b.X,
	}
}

// CoordinatesIntersection calculates the intersection a ∩ b in a 2D space,
// defined as the set of all those cells which are common to both
// a and b. Both a and b are expected to be right-exclusive ranges.
//
//	┌──────┐     ┌──────┐     ┌──────┐
//	│ AA   │  ∩  │      │  =  │      │
//	│      │     │   BB │     │      │
//	└──────┘     └──────┘     └──────┘
//	┌──────┐     ┌──────┐     ┌──────┐
//	│ AAAAA│  ∩  │      │  =  │      │
//	│AAA   │     │BBBBB │     │CCC   │
//	└──────┘     └──────┘     └──────┘
//	┌──────┐     ┌──────┐     ┌──────┐
//	│  BBBB│  ∩  │      │  =  │      │
//	│BBB   │     │AAAAA │     │CCC   │
//	└──────┘     └──────┘     └──────┘
//	┌──────┐     ┌──────┐     ┌──────┐
//	│     A│  ∩  │      │  =  │      │
//	│AAAAAA│     │BBB   │     │CCC   │
//	└──────┘     └──────┘     └──────┘
func CoordinatesIntersection(
	startA, endA, startB, endB Coordinates,
) (intersectionStart Coordinates, intersectionEnd Coordinates, ok bool) {
	startA, endA = CoordinatesSort(startA, endA)
	startB, endB = CoordinatesSort(startB, endB)

	// conflate non-sorted cases into sorted
	actualStartA, actualStartB := CoordinatesSort(startA, startB)
	if actualStartA != startA {
		temp := endB
		endB = endA
		endA = temp
		startA = actualStartA
		startB = actualStartB
	}

	if startB.Y > endA.Y || (startB.Y == endA.Y && startB.X >= endA.X) ||
		startA == endA || startB == endB {
		return
	}

	intersectionStart = startB
	intersectionEnd, _ = CoordinatesSort(endA, endB)
	ok = true
	return
}

// CoordinatesInBounds returns true if the given position is
// within the given right-exclusive bounds.
func CoordinatesInBounds(pos Coordinates, bounds Coordinates) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < bounds.X && pos.Y < bounds.Y
}

// MergeRanges combines the given range slice, such that
// it returns the smallest set of ranges that is equivalent to
// the given set of ranges, by merging all intersecting ranges.
func MergeRanges(ranges []Range) []Range {
	if len(ranges) == 0 {
		return []Range{}
	}

	// Normalize and sort ranges by their start position in file order
	sorted := make([]Range, len(ranges))
	for i, r := range ranges {
		sorted[i] = normalizeRange(r)
	}

	sort.Slice(sorted, func(i, j int) bool {
		return compareCoords(sorted[i].Start, sorted[j].Start) < 0
	})

	result := []Range{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		current := sorted[i]
		lastIdx := len(result) - 1

		// Check if current range overlaps or is adjacent to the last merged range
		if rangesOverlapOrAdjacent(result[lastIdx], current) {
			result[lastIdx] = mergeRanges(result[lastIdx], current)
		} else {
			result = append(result, current)
		}
	}
	return result
}

// compareCoords compares two coordinates in file order
// Returns: -1 if c1 < c2, 0 if c1 == c2, 1 if c1 > c2
func compareCoords(c1, c2 Coordinates) int {
	if c1.Y != c2.Y {
		if c1.Y < c2.Y {
			return -1
		}
		return 1
	}
	if c1.X != c2.X {
		if c1.X < c2.X {
			return -1
		}
		return 1
	}
	return 0
}

// rangesOverlapOrAdjacent checks if two ranges overlap or are adjacent in file space
func rangesOverlapOrAdjacent(r1, r2 Range) bool {
	// r1 and r2 are already normalized and r1.start <= r2.start (due to sorting)

	// If r2 starts before or at the position where r1 ends, they overlap or are adjacent
	// We use <= because adjacent ranges should be merged
	return compareCoords(r2.Start, r1.End) <= 0
}

// mergeRanges combines two ranges by taking the earliest start and latest end
func mergeRanges(r1, r2 Range) Range {
	start := r1.Start
	if compareCoords(r2.Start, start) < 0 {
		start = r2.Start
	}

	end := r1.End
	if compareCoords(r2.End, end) > 0 {
		end = r2.End
	}

	return Range{Start: start, End: end}
}

// normalizeRange ensures start comes before end in file order
func normalizeRange(r Range) Range {
	if compareCoords(r.Start, r.End) > 0 {
		return Range{Start: r.End, End: r.Start}
	}
	return r
}
