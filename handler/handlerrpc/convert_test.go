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

package handlerrpc

import (
	"image"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
	"google.golang.org/protobuf/proto"
)

func nonZero(t *testing.T, v reflect.Value, next *int64) {
	t.Helper()
	*next++
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			nonZero(t, v.Field(i), next)
		}
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(*next)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(*next))
	default:
		t.Fatalf("term.Image has a %s field that ImagePlacement does not mirror", v.Type())
	}
}

func TestImagePlacementRoundTrip(t *testing.T) {
	var full term.Image
	v := reflect.ValueOf(&full).Elem()
	var next int64
	for i := range v.NumField() {
		if v.Type().Field(i).Name == "Src" {
			continue
		}
		nonZero(t, v.Field(i), &next)
	}

	tests := []struct {
		name string
		img  term.Image
	}{
		{name: "every field set", img: full},
		{name: "zero value", img: term.Image{}},
		{name: "negative offsets", img: term.Image{
			ID: 3, Offset: image.Pt(-4, -2), RasterOffset: image.Pt(-1, -7),
			Clip: image.Rect(-1, -1, 5, 5),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wire, err := proto.Marshal(NewImagePlacement(tt.img))
			require.NoError(t, err)
			var decoded ImagePlacement
			require.NoError(t, proto.Unmarshal(wire, &decoded))
			assert.Equal(t, tt.img, decoded.ToModel())
		})
	}
}
