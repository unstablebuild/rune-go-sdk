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
	"github.com/unstablebuild/tcell/v3"
)

func TestAttributesUnion(t *testing.T) {
	tsuite := []struct {
		inA, inB Attributes
		wantOut  Attributes
	}{
		{},

		{Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorDefault}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorDefault}, Attributes{Bg: tcell.ColorRed}},

		{Attributes{Fg: tcell.ColorWhite}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorWhite}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorRed}},

		{Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse},
			Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse},
			Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse}},
		{Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse},
			Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse},
			Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse}},

		{Attributes{Attrs: tcell.AttrBold}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold}},
		{Attributes{Attrs: tcell.AttrBold}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold}},
		{Attributes{Attrs: tcell.AttrUnderline}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrUnderline}},
		{Attributes{Attrs: tcell.AttrUnderline}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrUnderline}},
		{Attributes{Attrs: tcell.AttrReverse}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrReverse}},
		{Attributes{Attrs: tcell.AttrReverse}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrReverse}},

		{Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrBold}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrBold}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold}},
		{Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrUnderline}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrUnderline}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrUnderline}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrUnderline}},
		{Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrReverse}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrReverse}},

		{Attributes{Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline},
			Attributes{Bg: tcell.ColorRed},
			Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline}},
		{Attributes{Bg: tcell.ColorRed},
			Attributes{Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline},
			Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline}},

		{Attributes{Fg: tcell.ColorDefault}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorDefault}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorRed}},
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

		{Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse}, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse},
			Attributes{Fg: tcell.ColorDefault}},
		{Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse}, Attributes{Bg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline | tcell.AttrReverse},
			Attributes{Bg: tcell.ColorDefault}},

		{Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorDefault}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorDefault}, Attributes{Bg: tcell.ColorRed}},

		{Attributes{Fg: tcell.ColorWhite}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorWhite}},
		{Attributes{Bg: tcell.ColorWhite}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorWhite}},

		{Attributes{Fg: tcell.ColorWhite}, Attributes{Fg: tcell.ColorWhite}, Attributes{Fg: tcell.ColorDefault}},
		{Attributes{Bg: tcell.ColorWhite}, Attributes{Bg: tcell.ColorWhite}, Attributes{Bg: tcell.ColorDefault}},

		{Attributes{Attrs: tcell.AttrBold}, Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrBold}},
		{Attributes{Attrs: tcell.AttrBold}, Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrBold}},
		{Attributes{Attrs: tcell.AttrUnderline}, Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrUnderline}},
		{Attributes{Attrs: tcell.AttrUnderline}, Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrUnderline}},
		{Attributes{Attrs: tcell.AttrReverse}, Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse}},
		{Attributes{Attrs: tcell.AttrReverse}, Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse}},

		{Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrBold}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrBold}, Attributes{Bg: tcell.ColorRed}},
		{Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrUnderline}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrUnderline}, Attributes{Bg: tcell.ColorRed}},
		{Attributes{Fg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse}, Attributes{Fg: tcell.ColorRed}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse}, Attributes{Bg: tcell.ColorRed}},

		{Attributes{Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline}, Attributes{Bg: tcell.ColorRed},
			Attributes{Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline}},
		{Attributes{Bg: tcell.ColorRed}, Attributes{Attrs: tcell.AttrReverse | tcell.AttrBold | tcell.AttrUnderline},
			Attributes{Bg: tcell.ColorRed}},

		{Attributes{Fg: tcell.ColorDefault}, Attributes{Fg: tcell.ColorRed}, Attributes{Fg: tcell.ColorDefault}},
		{Attributes{Bg: tcell.ColorDefault}, Attributes{Bg: tcell.ColorRed}, Attributes{Bg: tcell.ColorDefault}},
	}

	for i, tcase := range tsuite {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actualOut := AttributesDifference(tcase.inA, tcase.inB)
			assert.Equal(t, tcase.wantOut, actualOut)
		})
	}
}

func BenchmarkAttributesOperations(b *testing.B) {
	attr := Attributes{Fg: tcell.ColorDefault, Bg: tcell.ColorDefault}
	for i := 0; i < b.N; i++ {
		attr = AttributesUnion(attr, Attributes{Fg: tcell.ColorRed, Attrs: tcell.AttrBold | tcell.AttrUnderline})
		attr = AttributesDifference(attr, Attributes{Attrs: tcell.AttrBold | tcell.AttrReverse})
		attr = AttributesUnion(attr, Attributes{Fg: tcell.ColorGreen, Attrs: tcell.AttrUnderline, Bg: tcell.ColorBlack})
	}
}
