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

import "errors"

// ErrAborted is returned by Result when input was
// cancelled via Ctrl+C.
var ErrAborted = errors.New("aborted")

// TabStyle determines how tab completion candidates
// are presented.
type TabStyle int

const (
	// TabCircular cycles through candidates inline.
	// This is the zero value, making it the default
	// tab completion style.
	TabCircular TabStyle = iota
	// TabPrints shows all candidates in a grid below
	// the input line on the second tab press.
	TabPrints
)

// Completer returns completions for the entire line.
// It receives the line and returns the head (portion
// before completions), the candidates, and the tail
// (portion after completions).
type Completer func(
	line string,
) (head string, completions []string, tail string)

// WordCompleter returns completions for the word at
// the cursor. It receives the line and cursor position
// and returns the head, candidates, and tail.
type WordCompleter func(
	line string, pos int,
) (head string, completions []string, tail string)
