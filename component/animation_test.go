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

package component

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func newTestInterrupter() (term.Interrupter, chan struct{}) {
	ch := make(chan struct{})
	interrupter := term.FuncInterrupter(func(context.Context) error {
		ch <- struct{}{}
		return nil
	})
	return interrupter, ch
}

func TestAnimation(t *testing.T) {
	t.Parallel()
	suite := []struct {
		desc         string
		newAnimation func(*testing.T, term.Interrupter, []string, []int, int) *Animation
	}{
		{"NewAnimation constructor",
			func(t *testing.T, interrupter term.Interrupter, frames []string, sequence []int, fps int) *Animation {
				return NewAnimation(interrupter, frames, sequence, fps)
			}},
		{"encoded and decoded animation",
			func(t *testing.T, interrupter term.Interrupter, frames []string, sequence []int, fps int) *Animation {
				a := NewAnimation(interrupter, frames, sequence, fps)
				raw := EncodeAnimation(a, 8, 4)
				ret, err := DecodeAnimation(raw, fps, interrupter)
				require.NoError(t, err)
				require.NoError(t, a.Close())
				return ret
			}},
		{"non stringer encoded and decoded animation",
			func(t *testing.T, interrupter term.Interrupter, frames []string, sequence []int, fps int) *Animation {
				components := make([]WithAttributes, len(frames))
				for i, frame := range frames {
					components[i] = noStringerString{NewStringWithConfig(frame, StringConfig{
						Alignment: AlignmentCentered,
					})}
				}
				a := new(Animation)
				a.InitWithComponents(context.Background(), interrupter, components, sequence, fps)

				raw := EncodeAnimation(a, 8, 4)
				ret, err := DecodeAnimation(raw, fps, interrupter)
				require.NoError(t, err)
				require.NoError(t, a.Close())
				return ret
			}},
	}
	for _, test := range suite {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			t.Run("no frames", func(t *testing.T) {
				t.Parallel()
				fps := 30
				interrupter, _ := newTestInterrupter()

				frames := []string{}
				sequence := []int{}
				expected := "        \n        \n        \n        "

				// sut
				c := NewAnimation(interrupter, frames, sequence, fps)
				c.Resize(8, 4)

				for i := 0; i < 30; i++ {
					w := term.NewStringWriter(8, 4)
					c.Draw(w)

					require.NoError(t, w.Flush())
					assert.Equal(t, expected, w.String())
				}

				require.NoError(t, c.Close())
			})

			t.Run("single static frame", func(t *testing.T) {
				t.Parallel()
				fps := 30
				interrupter, ch := newTestInterrupter()

				frames := []string{"1111    \n1111    \n    @@@@\n    @@@@"}
				sequence := []int{0}
				expected := "1111    \n1111    \n    @@@@\n    @@@@"

				// sut
				c := NewAnimation(interrupter, frames, sequence, fps)
				c.Resize(8, 4)

				for i := 0; i < 30; i++ {
					w := term.NewStringWriter(8, 4)
					<-ch
					c.Draw(w)

					require.NoError(t, w.Flush())
					assert.Equal(t, expected, w.String())
				}

				require.NoError(t, c.Close())
			})

			t.Run("multiple frames, repeated or not", func(t *testing.T) {
				t.Parallel()
				fps := 30
				interrupter, ch := newTestInterrupter()

				frames := []string{
					"0000    \n0000    \n    0000\n    0000",
					"1111    \n1111    \n    1111\n    1111",
				}
				sequence := []int{0, 0, 1}

				// sut
				c := test.newAnimation(t, interrupter, frames, sequence, fps)
				c.Resize(8, 4)

				for i := range 30 {
					w := term.NewStringWriter(8, 4)
					<-ch
					c.Draw(w)
					var expected string
					switch i % 3 {
					case 0:
						expected = "0000    \n0000    \n    0000\n    0000"
					case 1:
						expected = "0000    \n0000    \n    0000\n    0000"
					case 2:
						expected = "1111    \n1111    \n    1111\n    1111"
					default:
						panic("hmmm")
					}

					require.NoError(t, w.Flush())
					assert.Equal(t, expected, w.String())
				}

				require.NoError(t, c.Close())
			})
		})
	}
}

type noStringerString struct {
	str String
}

func (n noStringerString) Draw(w term.Writer) {
	n.str.Draw(w)
}
func (n noStringerString) SetAttr(attr term.Attributes) term.Attributes {
	return n.str.SetAttr(attr)
}

func (n noStringerString) Resize(width, height int) {
	n.str.Resize(width, height)
}
