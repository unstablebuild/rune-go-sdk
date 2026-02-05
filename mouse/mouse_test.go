// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mouse_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/mouse"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestMouseHandle(t *testing.T) {
	tsuite := []struct {
		desc   string
		expect []expect
		evs    []term.Event
	}{
		{"ignores non-mouse events",
			nil, []term.Event{{Type: term.EventKey}}},
		{"delegates mouse release",
			[]expect{
				expectMouseAction(t, 4, 8, mouse.LeftClick, true, false),
				expectMouseAction(t, 4, 8, mouse.Release, true, false),
			},
			[]term.Event{
				evMouseKey(4, 8, term.MouseLeft),
				evMouseKey(4, 8, term.MouseRelease),
			},
		},
		{"delegates mouse left click",
			[]expect{
				expectMouseAction(t, 4, 8, mouse.LeftClick, true, false),
			},
			[]term.Event{evMouseKey(4, 8, term.MouseLeft)}},
		{"delegates mouse right click",
			[]expect{
				expectMouseAction(t, 4, 8, mouse.RightClick, true, false),
			},
			[]term.Event{evMouseKey(4, 8, term.MouseRight)}},
		{"delegates mouse middle click",
			[]expect{
				expectMouseAction(t, 4, 8, mouse.MiddleClick, true, false),
			},
			[]term.Event{evMouseKey(4, 8, term.MouseMiddle)}},
		{"scrolls up",
			[]expect{
				multiExpect(
					expectIgnoreMouseAction(),
					expectScrollUp(t, 1),
				),
			},
			[]term.Event{evMouseKey(4, 8, term.MouseWheelUp)}},
		{"scrolls down",
			[]expect{
				multiExpect(
					expectIgnoreMouseAction(),
					expectScrollDown(t, 1),
				),
			},
			[]term.Event{evMouseKey(4, 8, term.MouseWheelDown)}},
		{"click drag sets selection",
			[]expect{
				multiExpect(
					expectMouseAction(t, 1, 2, mouse.LeftClick, false, true),
					expectClearSelection(t),
					expectSetSelectionStart(t, 1, 2),
				),
				multiExpect(
					expectScrollDown(t, 1),
					expectSetSelectionEnd(t, 4, 8),
				),
				multiExpect(
					expectScrollDown(t, 1),
					expectSetSelectionEnd(t, 4, 9),
				),
				multiExpect(
					expectSetSelectionEnd(t, 2, 4),
				),
				multiExpect(
					expectIgnoreMouseAction(),
				),
			},
			[]term.Event{
				evMouseKey(1, 2, term.MouseLeft),
				evMouseKey(4, 8, term.MouseLeft),
				evMouseKey(4, 9, term.MouseLeft),
				evMouseKey(2, 4, term.MouseLeft),
				evMouseKey(2, 5, term.MouseRelease),
			},
		},
		{"double click selects word",
			[]expect{
				multiExpect(
					expectIgnoreMouseAction(),
					expectClearSelection(t),
					expectSetSelectionStart(t, 1, 2),
				),
				expectIgnoreMouseAction(),
				multiExpect(
					expectIgnoreMouseAction(),
					expectSelectWord(t, 1, 2),
				),
				expectIgnoreMouseAction(),
			},
			[]term.Event{
				evMouseKey(1, 2, term.MouseLeft),
				evMouseKey(1, 2, term.MouseRelease),
				evMouseKey(1, 2, term.MouseLeft),
				evMouseKey(1, 2, term.MouseRelease),
			},
		},
		{"triple click selects line",
			[]expect{
				multiExpect(
					expectIgnoreMouseAction(),
					expectClearSelection(t),
					expectSetSelectionStart(t, 1, 2),
				),
				expectIgnoreMouseAction(),
				multiExpect(
					expectIgnoreMouseAction(),
					expectSelectWord(t, 1, 2),
				),
				expectIgnoreMouseAction(),
				multiExpect(
					expectIgnoreMouseAction(),
					expectSelectLine(t, 2),
				),
				expectIgnoreMouseAction(),
			},
			[]term.Event{
				evMouseKey(1, 2, term.MouseLeft),
				evMouseKey(1, 2, term.MouseRelease),
				evMouseKey(1, 2, term.MouseLeft),
				evMouseKey(1, 2, term.MouseRelease),
				evMouseKey(1, 2, term.MouseLeft),
				evMouseKey(1, 2, term.MouseRelease),
			},
		},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.desc, func(t *testing.T) {
			mock := new(mock)
			mock.returnWidth = 10
			mock.returnHeight = 10
			m := mouse.New(mock)

			for i, ev := range tcase.evs {
				// sut
				if i < len(tcase.expect) && tcase.expect[i] != nil {
					tcase.expect[i](mock)
				}
				_, _ = m.Handle(ev)
			}
		})
	}
}

type mock struct {
	expectOnAction          func(ev term.Event, pos term.Coordinates, action mouse.Action) bool
	expectScrollUp          func(n int) bool
	expectScrollDown        func(n int) bool
	expectSetSelectionEnd   func(pos term.Coordinates)
	expectSetSelectionStart func(pos term.Coordinates)
	expectClearSelection    func()
	expectSelectWordAt      func(pos term.Coordinates)
	expectSelectLine        func(y int)
	returnWidth             int
	returnHeight            int
}

func evMouseKey(x, y int, key term.Key) term.Event {
	return term.Event{Type: term.EventMouse, MouseX: x, MouseY: y, Key: key}
}

func expectScrollUp(t *testing.T, n int) expect {
	return func(m *mock) {
		m.expectScrollUp = func(_n int) bool {
			assert.Equal(t, n, _n)
			return true
		}
	}
}

func expectScrollDown(t *testing.T, n int) expect {
	return func(m *mock) {
		m.expectScrollDown = func(_n int) bool {
			assert.Equal(t, n, _n)
			return true
		}
	}
}

func expectSetSelectionStart(t *testing.T, x, y int) expect {
	return func(m *mock) {
		m.expectSetSelectionStart = func(pos term.Coordinates) {
			assert.Equal(t, term.Coordinates{X: x, Y: y}, pos)
		}
	}
}

func expectClearSelection(t *testing.T) expect {
	return func(m *mock) {
		m.expectClearSelection = func() {}
	}
}

func expectSetSelectionEnd(t *testing.T, x, y int) expect {
	return func(m *mock) {
		m.expectSetSelectionEnd = func(pos term.Coordinates) {
			assert.Equal(t, term.Coordinates{X: x, Y: y}, pos)
		}
	}
}

func expectSelectWord(t *testing.T, x, y int) expect {
	return func(m *mock) {
		m.expectSelectWordAt = func(pos term.Coordinates) {
			assert.Equal(t, term.Coordinates{X: x, Y: y}, pos)
		}
	}
}

func expectSelectLine(t *testing.T, y int) expect {
	return func(m *mock) {
		m.expectSelectLine = func(_y int) {
			assert.Equal(t, y, _y)
		}
	}
}

func expectIgnoreMouseAction() expect {
	return func(m *mock) {
		m.expectOnAction = func(ev term.Event, pos term.Coordinates, action mouse.Action) bool {
			return false
		}
	}
}

func expectMouseAction(t *testing.T, x, y int, action mouse.Action, handled, moved bool) expect {
	return func(m *mock) {
		m.expectOnAction = func(ev term.Event, pos term.Coordinates, _action mouse.Action) bool {
			assert.Equal(t, term.Coordinates{X: x, Y: y}, pos)
			assert.Equal(t, action, _action)
			return handled
		}
	}
}

type expect func(*mock)

func multiExpect(e ...expect) expect {
	return func(m *mock) {
		for _, expect := range e {
			expect(m)
		}
	}
}

func (m *mock) OnAction(ev term.Event, pos term.Coordinates, action mouse.Action) bool {
	return m.expectOnAction(ev, pos, action)
}

func (m *mock) ScrollUp(n int) bool {
	return m.expectScrollUp(n)
}

func (m *mock) ScrollDown(n int) bool {
	return m.expectScrollDown(n)
}

func (m *mock) SetSelectionEnd(pos term.Coordinates) {
	m.expectSetSelectionEnd(pos)
}

func (m *mock) SetSelectionStart(pos term.Coordinates) {
	m.expectSetSelectionStart(pos)
}

func (m *mock) ClearSelection() {
	m.expectClearSelection()
}

func (m *mock) SelectWordAt(pos term.Coordinates) {
	m.expectSelectWordAt(pos)
}

func (m *mock) SelectLine(y int) {
	m.expectSelectLine(y)
}

func (m *mock) Width() int {
	return m.returnWidth
}

func (m *mock) Height() int {
	return m.returnHeight
}
