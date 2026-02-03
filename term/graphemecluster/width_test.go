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

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWidth(t *testing.T) {
	suite := []struct {
		inputStr string
		expected int
	}{
		{"", 0},
		{"a", 1},
		{"aa", 2},
		{"", 2},
		{"1", 3},
		{"1", 3},
		{"", 4},
		{"中", 2},
		{"\t", 0},
		{"\n", 0},
	}

	for i, test := range suite {
		t.Run(fmt.Sprintf("test case %d: %q", i, test.inputStr), func(t *testing.T) {
			actual := StringWidth(test.inputStr)
			assert.Equal(t, test.expected, actual)
		})
	}
}
