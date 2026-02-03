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


package iterator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	myVal string
	myNum int
}

func TestFilter(t *testing.T) {
	tsuite := []struct {
		desc     string
		in       []testStruct
		inFn     func(testStruct) bool
		expected []testStruct
	}{
		{"empty input returns empty out, no filter", nil, func(testStruct) bool { return true }, []testStruct{}},
		{"empty input returns empty out, filter everything", nil, func(testStruct) bool { return false }, []testStruct{}},
		{"input returns empty out, filter everything", []testStruct{{}}, func(testStruct) bool { return false }, []testStruct{}},
		{"returns input as-is, filters nothing", []testStruct{{}}, func(testStruct) bool { return true }, []testStruct{{}}},
		{"filters some values", []testStruct{{"", 0}, {"1", 1}},
			func(t testStruct) bool {
				return t.myVal != ""
			}, []testStruct{{"1", 1}}},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.desc, func(t *testing.T) {
			actualOut := Filter(FromSlice(tcase.in), tcase.inFn)
			actualOutSlice, err := ToSlice(context.Background(), actualOut)
			require.NoError(t, err)
			assert.Equal(t, tcase.expected, actualOutSlice)
		})
	}
}
