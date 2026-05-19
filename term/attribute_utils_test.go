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

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributesUnion(t *testing.T) {
	tsuite := []struct {
		inA, inB Attributes
		wantOut  Attributes
	}{
		{},

		{Attributes{Fg: ColorRed}, Attributes{Fg: ColorDefault}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorRed}, Attributes{Bg: ColorDefault}, Attributes{Bg: ColorRed}},

		{Attributes{Fg: ColorWhite}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorWhite}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorRed}},

		{Attributes{Fg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse},
			Attributes{Fg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse},
			Attributes{Fg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse}},
		{Attributes{Bg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse},
			Attributes{Bg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse},
			Attributes{Bg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse}},

		{Attributes{Attrs: AttrBold}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorRed, Attrs: AttrBold}},
		{Attributes{Attrs: AttrBold}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorRed, Attrs: AttrBold}},
		{Attributes{Attrs: AttrUnderline}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorRed, Attrs: AttrUnderline}},
		{Attributes{Attrs: AttrUnderline}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorRed, Attrs: AttrUnderline}},
		{Attributes{Attrs: AttrReverse}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorRed, Attrs: AttrReverse}},
		{Attributes{Attrs: AttrReverse}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorRed, Attrs: AttrReverse}},

		{Attributes{Fg: ColorRed}, Attributes{Attrs: AttrBold}, Attributes{Fg: ColorRed, Attrs: AttrBold}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrBold}, Attributes{Bg: ColorRed, Attrs: AttrBold}},
		{Attributes{Fg: ColorRed}, Attributes{Attrs: AttrUnderline}, Attributes{Fg: ColorRed, Attrs: AttrUnderline}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrUnderline}, Attributes{Bg: ColorRed, Attrs: AttrUnderline}},
		{Attributes{Fg: ColorRed}, Attributes{Attrs: AttrReverse}, Attributes{Fg: ColorRed, Attrs: AttrReverse}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrReverse}, Attributes{Bg: ColorRed, Attrs: AttrReverse}},

		{Attributes{Attrs: AttrReverse | AttrBold | AttrUnderline},
			Attributes{Bg: ColorRed},
			Attributes{Bg: ColorRed, Attrs: AttrReverse | AttrBold | AttrUnderline}},
		{Attributes{Bg: ColorRed},
			Attributes{Attrs: AttrReverse | AttrBold | AttrUnderline},
			Attributes{Bg: ColorRed, Attrs: AttrReverse | AttrBold | AttrUnderline}},

		{Attributes{Fg: ColorDefault}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorDefault}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorRed}},
	}

	for i, tcase := range tsuite {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actualOut := AttributesUnion(tcase.inA, tcase.inB)
			assert.Equal(t, tcase.wantOut, actualOut)
		})
	}
}

func TestAttributesDifference(t *testing.T) {
	tsuite := []struct {
		inA, inB Attributes
		wantOut  Attributes
	}{
		{},

		{Attributes{Fg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse}, Attributes{Fg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse},
			Attributes{Fg: ColorDefault}},
		{Attributes{Bg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse}, Attributes{Bg: ColorRed, Attrs: AttrBold | AttrUnderline | AttrReverse},
			Attributes{Bg: ColorDefault}},

		{Attributes{Fg: ColorRed}, Attributes{Fg: ColorDefault}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorRed}, Attributes{Bg: ColorDefault}, Attributes{Bg: ColorRed}},

		{Attributes{Fg: ColorWhite}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorWhite}},
		{Attributes{Bg: ColorWhite}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorWhite}},

		{Attributes{Fg: ColorWhite}, Attributes{Fg: ColorWhite}, Attributes{Fg: ColorDefault}},
		{Attributes{Bg: ColorWhite}, Attributes{Bg: ColorWhite}, Attributes{Bg: ColorDefault}},

		{Attributes{Attrs: AttrBold}, Attributes{Fg: ColorRed}, Attributes{Attrs: AttrBold}},
		{Attributes{Attrs: AttrBold}, Attributes{Bg: ColorRed}, Attributes{Attrs: AttrBold}},
		{Attributes{Attrs: AttrUnderline}, Attributes{Fg: ColorRed}, Attributes{Attrs: AttrUnderline}},
		{Attributes{Attrs: AttrUnderline}, Attributes{Bg: ColorRed}, Attributes{Attrs: AttrUnderline}},
		{Attributes{Attrs: AttrReverse}, Attributes{Fg: ColorRed}, Attributes{Attrs: AttrReverse}},
		{Attributes{Attrs: AttrReverse}, Attributes{Bg: ColorRed}, Attributes{Attrs: AttrReverse}},

		{Attributes{Fg: ColorRed}, Attributes{Attrs: AttrBold}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrBold}, Attributes{Bg: ColorRed}},
		{Attributes{Fg: ColorRed}, Attributes{Attrs: AttrUnderline}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrUnderline}, Attributes{Bg: ColorRed}},
		{Attributes{Fg: ColorRed}, Attributes{Attrs: AttrReverse}, Attributes{Fg: ColorRed}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrReverse}, Attributes{Bg: ColorRed}},

		{Attributes{Attrs: AttrReverse | AttrBold | AttrUnderline}, Attributes{Bg: ColorRed},
			Attributes{Attrs: AttrReverse | AttrBold | AttrUnderline}},
		{Attributes{Bg: ColorRed}, Attributes{Attrs: AttrReverse | AttrBold | AttrUnderline},
			Attributes{Bg: ColorRed}},

		{Attributes{Fg: ColorDefault}, Attributes{Fg: ColorRed}, Attributes{Fg: ColorDefault}},
		{Attributes{Bg: ColorDefault}, Attributes{Bg: ColorRed}, Attributes{Bg: ColorDefault}},
	}

	for i, tcase := range tsuite {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actualOut := AttributesDifference(tcase.inA, tcase.inB)
			assert.Equal(t, tcase.wantOut, actualOut)
		})
	}
}

func BenchmarkAttributesOperations(b *testing.B) {
	attr := Attributes{Fg: ColorDefault, Bg: ColorDefault}
	for i := 0; i < b.N; i++ {
		attr = AttributesUnion(attr, Attributes{Fg: ColorRed, Attrs: AttrBold | AttrUnderline})
		attr = AttributesDifference(attr, Attributes{Attrs: AttrBold | AttrReverse})
		attr = AttributesUnion(attr, Attributes{Fg: ColorGreen, Attrs: AttrUnderline, Bg: ColorBlack})
	}
}
