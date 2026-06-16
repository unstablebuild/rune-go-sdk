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
	"errors"
	"fmt"
	"strings"
)

// ParseKeys parses the given sequence of key combinations.
func ParseKeys(sequence string) (ret []KeyComb, err error) {
	runes := []rune(sequence)
	for len(runes) != 0 {
		r := runes[0]
		switch r {
		// escape character
		case '\\':
			if len(runes) == 1 {
				err = errors.New("unterminated escape sequence: " +
					"Either '\\', '>' or '<' must follow a start of escape sequence '\\'")
				return
			}
			switch runes[1] {
			case '>', '<', '\\':
				ret = append(ret, KeyComb{Ch: runes[1]})
				runes = runes[2:]
			default:
				err = fmt.Errorf("invalid escape sequence: " +
					"Either '\\', '>' or '<' must follow a start of escape sequence '\\'")
				return
			}
		// invalid, space should be represented as <space>
		case ' ':
			err = errors.New("invalid escape sequence: " +
				"Space should be represented with '<space>' syntax")
			return

		case '>':
			err = errors.New("invalid escape sequence: unescaped, starting '>' character")
			return
		// start of key
		case '<':
			idxGt := strings.IndexRune(string(runes), '>')
			if idxGt < 0 {
				err = errors.New("unterminated key: '<' found but no matching '>' found")
				return
			}
			var key KeyComb
			key, err = ParseKey(string(runes[0 : idxGt+1]))
			if err != nil {
				return
			}
			ret = append(ret, key)
			runes = runes[idxGt+1:]
		default:
			var key KeyComb
			key, err = ParseKey(string(runes[0]))
			if err != nil {
				return
			}
			ret = append(ret, key)
			runes = runes[1:]
		}
	}
	return
}

// ParseKey parses str into a KeyComb or returns
// error if it fails to parse it.
func ParseKey(str string) (KeyComb, error) {
	str = strings.TrimSpace(str)
	switch len([]rune(str)) {
	case 0:
		return KeyComb{}, errors.New("invalid empty input")
	case 1:
		ch := []rune(str)[0]
		switch ch {
		case '>', '<', '\\':
			return KeyComb{}, errors.New("characters '>', '<' and '\\' must be escaped with '\\'")
		default:
			// characters that would return len(str) > 1 are processed in prev statement
			return KeyComb{Ch: ch}, nil
		}
	default:
		str = strings.ToLower(str)
		switch str {
		case "\\>":
			return KeyComb{Ch: '>'}, nil
		case "\\<":
			return KeyComb{Ch: '<'}, nil
		case "\\\\":
			return KeyComb{Ch: '\\'}, nil
		case "<f1>":
			return KeyComb{Key: KeyF1}, nil
		case "<f2>":
			return KeyComb{Key: KeyF2}, nil
		case "<f3>":
			return KeyComb{Key: KeyF3}, nil
		case "<f4>":
			return KeyComb{Key: KeyF4}, nil
		case "<f5>":
			return KeyComb{Key: KeyF5}, nil
		case "<f6>":
			return KeyComb{Key: KeyF6}, nil
		case "<f7>":
			return KeyComb{Key: KeyF7}, nil
		case "<f8>":
			return KeyComb{Key: KeyF8}, nil
		case "<f9>":
			return KeyComb{Key: KeyF9}, nil
		case "<f10>":
			return KeyComb{Key: KeyF10}, nil
		case "<f11>":
			return KeyComb{Key: KeyF11}, nil
		case "<f12>":
			return KeyComb{Key: KeyF12}, nil
		case "<insert>":
			return KeyComb{Key: KeyInsert}, nil
		case "<delete>":
			return KeyComb{Key: KeyDelete}, nil
		case "<home>":
			return KeyComb{Key: KeyHome}, nil
		case "<end>":
			return KeyComb{Key: KeyEnd}, nil
		case "<pgup>":
			return KeyComb{Key: KeyPgup}, nil
		case "<pgdn>":
			return KeyComb{Key: KeyPgdn}, nil
		case "<up>":
			return KeyComb{Key: KeyArrowUp}, nil
		case "<down>":
			return KeyComb{Key: KeyArrowDown}, nil
		case "<left>":
			return KeyComb{Key: KeyArrowLeft}, nil
		case "<right>":
			return KeyComb{Key: KeyArrowRight}, nil
		case "<mouse-left>":
			return KeyComb{Key: MouseLeft}, nil
		case "<mouse-middle>":
			return KeyComb{Key: MouseMiddle}, nil
		case "<mouse-right>":
			return KeyComb{Key: MouseRight}, nil
		case "<mouse-release>":
			return KeyComb{Key: MouseRelease}, nil
		case "<mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp}, nil
		case "<mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown}, nil
		case "<tab>":
			return KeyComb{Key: KeyTab}, nil
		case "<enter>":
			return KeyComb{Key: KeyEnter}, nil
		case "<esc>":
			return KeyComb{Key: KeyEsc}, nil
		case "<space>":
			return KeyComb{Key: KeySpace}, nil
		case "<backspace>":
			return KeyComb{Key: KeyBackspace}, nil
		case "<capslock>":
			return KeyComb{Key: KeyCapsLock}, nil
		case "<numlock>":
			return KeyComb{Key: KeyNumLock}, nil
		case "<scrolllock>":
			return KeyComb{Key: KeyScrollLock}, nil
		case "<menu>":
			return KeyComb{Key: KeyMenu}, nil

		// meta
		case "<m-f1>", "<meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModMeta}, nil
		case "<m-f2>", "<meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModMeta}, nil
		case "<m-f3>", "<meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModMeta}, nil
		case "<m-f4>", "<meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModMeta}, nil
		case "<m-f5>", "<meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModMeta}, nil
		case "<m-f6>", "<meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModMeta}, nil
		case "<m-f7>", "<meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModMeta}, nil
		case "<m-f8>", "<meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModMeta}, nil
		case "<m-f9>", "<meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModMeta}, nil
		case "<m-f10>", "<meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModMeta}, nil
		case "<m-f11>", "<meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModMeta}, nil
		case "<m-f12>", "<meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModMeta}, nil
		case "<m-insert>", "<meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModMeta}, nil
		case "<m-delete>", "<meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModMeta}, nil
		case "<m-home>", "<meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModMeta}, nil
		case "<m-end>", "<meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModMeta}, nil
		case "<m-pgup>", "<meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModMeta}, nil
		case "<m-pgdn>", "<meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModMeta}, nil
		case "<m-up>", "<meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModMeta}, nil
		case "<m-down>", "<meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModMeta}, nil
		case "<m-left>", "<meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModMeta}, nil
		case "<m-right>", "<meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModMeta}, nil
		case "<m-mouse-left>", "<meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModMeta}, nil
		case "<m-mouse-middle>", "<meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModMeta}, nil
		case "<m-mouse-right>", "<meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModMeta}, nil
		case "<m-mouse-release>", "<meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModMeta}, nil
		case "<m-mouse-wheel-up>", "<meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModMeta}, nil
		case "<m-mouse-wheel-down>", "<meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModMeta}, nil
		case "<m-enter>", "<meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModMeta}, nil
		case "<m-space>", "<meta-space>":
			return KeyComb{Mod: ModMeta, Key: KeySpace}, nil
		case "<m-backspace>", "<meta-backspace>":
			return KeyComb{Mod: ModMeta, Key: KeyBackspace}, nil
		case "<m-esc>", "<meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModMeta}, nil
		case "<m-1>", "<meta-1>":
			return KeyComb{Ch: '1', Mod: ModMeta}, nil
		case "<m-2>", "<meta-2>":
			return KeyComb{Mod: ModMeta, Ch: '2'}, nil
		case "<m-3>", "<meta-3>":
			return KeyComb{Ch: '3', Mod: ModMeta}, nil
		case "<m-4>", "<meta-4>":
			return KeyComb{Ch: '4', Mod: ModMeta}, nil
		case "<m-5>", "<meta-5>":
			return KeyComb{Ch: '5', Mod: ModMeta}, nil
		case "<m-6>", "<meta-6>":
			return KeyComb{Ch: '6', Mod: ModMeta}, nil
		case "<m-7>", "<meta-7>":
			return KeyComb{Ch: '7', Mod: ModMeta}, nil
		case "<m-8>", "<meta-8>":
			return KeyComb{Ch: '8', Mod: ModMeta}, nil
		case "<m-9>", "<meta-9>":
			return KeyComb{Ch: '9', Mod: ModMeta}, nil
		case "<m-0>", "<meta-0>":
			return KeyComb{Ch: '0', Mod: ModMeta}, nil
		case "<m-`>", "<meta-`>":
			return KeyComb{Mod: ModMeta, Ch: '`'}, nil
		case "<m-a>", "<meta-a>":
			return KeyComb{Mod: ModMeta, Ch: 'a'}, nil
		case "<m-b>", "<meta-b>":
			return KeyComb{Mod: ModMeta, Ch: 'b'}, nil
		case "<m-c>", "<meta-c>":
			return KeyComb{Mod: ModMeta, Ch: 'c'}, nil
		case "<m-d>", "<meta-d>":
			return KeyComb{Mod: ModMeta, Ch: 'd'}, nil
		case "<m-e>", "<meta-e>":
			return KeyComb{Mod: ModMeta, Ch: 'e'}, nil
		case "<m-f>", "<meta-f>":
			return KeyComb{Mod: ModMeta, Ch: 'f'}, nil
		case "<m-g>", "<meta-g>":
			return KeyComb{Mod: ModMeta, Ch: 'g'}, nil
		case "<m-h>", "<meta-h>":
			return KeyComb{Mod: ModMeta, Ch: 'h'}, nil
		case "<m-tab>", "<meta-tab>":
			return KeyComb{Mod: ModMeta, Key: KeyTab}, nil
		case "<m-i>", "<meta-i>":
			return KeyComb{Mod: ModMeta, Ch: 'i'}, nil
		case "<m-j>", "<meta-j>":
			return KeyComb{Mod: ModMeta, Ch: 'j'}, nil
		case "<m-k>", "<meta-k>":
			return KeyComb{Mod: ModMeta, Ch: 'k'}, nil
		case "<m-l>", "<meta-l>":
			return KeyComb{Mod: ModMeta, Ch: 'l'}, nil
		case "<m-m>", "<meta-m>":
			return KeyComb{Mod: ModMeta, Ch: 'm'}, nil
		case "<m-n>", "<meta-n>":
			return KeyComb{Mod: ModMeta, Ch: 'n'}, nil
		case "<m-o>", "<meta-o>":
			return KeyComb{Mod: ModMeta, Ch: 'o'}, nil
		case "<m-p>", "<meta-p>":
			return KeyComb{Mod: ModMeta, Ch: 'p'}, nil
		case "<m-q>", "<meta-q>":
			return KeyComb{Mod: ModMeta, Ch: 'q'}, nil
		case "<m-r>", "<meta-r>":
			return KeyComb{Mod: ModMeta, Ch: 'r'}, nil
		case "<m-s>", "<meta-s>":
			return KeyComb{Mod: ModMeta, Ch: 's'}, nil
		case "<m-t>", "<meta-t>":
			return KeyComb{Mod: ModMeta, Ch: 't'}, nil
		case "<m-u>", "<meta-u>":
			return KeyComb{Mod: ModMeta, Ch: 'u'}, nil
		case "<m-v>", "<meta-v>":
			return KeyComb{Mod: ModMeta, Ch: 'v'}, nil
		case "<m-w>", "<meta-w>":
			return KeyComb{Mod: ModMeta, Ch: 'w'}, nil
		case "<m-x>", "<meta-x>":
			return KeyComb{Mod: ModMeta, Ch: 'x'}, nil
		case "<m-y>", "<meta-y>":
			return KeyComb{Mod: ModMeta, Ch: 'y'}, nil
		case "<m-z>", "<meta-z>":
			return KeyComb{Mod: ModMeta, Ch: 'z'}, nil
		case "<m-[>", "<meta-[>":
			return KeyComb{Mod: ModMeta, Ch: '['}, nil
		case "<m-\\\\>", "<meta-\\\\>":
			return KeyComb{Mod: ModMeta, Ch: '\\'}, nil
		case "<m-]>", "<meta-]>":
			return KeyComb{Mod: ModMeta, Ch: ']'}, nil
		case "<m-/>", "<meta-/>":
			return KeyComb{Mod: ModMeta, Ch: '/'}, nil
		case "<m-_>", "<meta-_>":
			return KeyComb{Mod: ModMeta, Ch: '_'}, nil
		case "<m-.>", "<meta-.>":
			return KeyComb{Mod: ModMeta, Ch: '.'}, nil
		case "<m-,>", "<meta-,>":
			return KeyComb{Mod: ModMeta, Ch: ','}, nil
		case "<m-;>", "<meta-;>":
			return KeyComb{Mod: ModMeta, Ch: ';'}, nil
		case "<m-'>", "<meta-'>":
			return KeyComb{Mod: ModMeta, Ch: '\''}, nil
		case "<m-=>", "<meta-=>":
			return KeyComb{Mod: ModMeta, Ch: '='}, nil
		case "<m-->", "<meta-->":
			return KeyComb{Mod: ModMeta, Ch: '-'}, nil

		// alt
		case "<a-f1>", "<alt-f1>":
			return KeyComb{Key: KeyF1, Mod: ModAlt}, nil
		case "<a-f2>", "<alt-f2>":
			return KeyComb{Key: KeyF2, Mod: ModAlt}, nil
		case "<a-f3>", "<alt-f3>":
			return KeyComb{Key: KeyF3, Mod: ModAlt}, nil
		case "<a-f4>", "<alt-f4>":
			return KeyComb{Key: KeyF4, Mod: ModAlt}, nil
		case "<a-f5>", "<alt-f5>":
			return KeyComb{Key: KeyF5, Mod: ModAlt}, nil
		case "<a-f6>", "<alt-f6>":
			return KeyComb{Key: KeyF6, Mod: ModAlt}, nil
		case "<a-f7>", "<alt-f7>":
			return KeyComb{Key: KeyF7, Mod: ModAlt}, nil
		case "<a-f8>", "<alt-f8>":
			return KeyComb{Key: KeyF8, Mod: ModAlt}, nil
		case "<a-f9>", "<alt-f9>":
			return KeyComb{Key: KeyF9, Mod: ModAlt}, nil
		case "<a-f10>", "<alt-f10>":
			return KeyComb{Key: KeyF10, Mod: ModAlt}, nil
		case "<a-f11>", "<alt-f11>":
			return KeyComb{Key: KeyF11, Mod: ModAlt}, nil
		case "<a-f12>", "<alt-f12>":
			return KeyComb{Key: KeyF12, Mod: ModAlt}, nil
		case "<a-insert>", "<alt-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModAlt}, nil
		case "<a-delete>", "<alt-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModAlt}, nil
		case "<a-home>", "<alt-home>":
			return KeyComb{Key: KeyHome, Mod: ModAlt}, nil
		case "<a-end>", "<alt-end>":
			return KeyComb{Key: KeyEnd, Mod: ModAlt}, nil
		case "<a-pgup>", "<alt-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModAlt}, nil
		case "<a-pgdn>", "<alt-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModAlt}, nil
		case "<a-up>", "<alt-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModAlt}, nil
		case "<a-down>", "<alt-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModAlt}, nil
		case "<a-left>", "<alt-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModAlt}, nil
		case "<a-right>", "<alt-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModAlt}, nil
		case "<a-mouse-left>", "<alt-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModAlt}, nil
		case "<a-mouse-middle>", "<alt-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModAlt}, nil
		case "<a-mouse-right>", "<alt-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModAlt}, nil
		case "<a-mouse-release>", "<alt-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModAlt}, nil
		case "<a-mouse-wheel-up>", "<alt-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModAlt}, nil
		case "<a-mouse-wheel-down>", "<alt-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModAlt}, nil
		case "<a-enter>", "<alt-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModAlt}, nil
		case "<a-space>", "<alt-space>":
			return KeyComb{Mod: ModAlt, Key: KeySpace}, nil
		case "<a-backspace>", "<alt-backspace>":
			return KeyComb{Mod: ModAlt, Key: KeyBackspace}, nil
		case "<a-esc>", "<alt-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModAlt}, nil
		case "<a-1>", "<alt-1>":
			return KeyComb{Ch: '1', Mod: ModAlt}, nil
		case "<a-2>", "<alt-2>":
			return KeyComb{Mod: ModAlt, Ch: '2'}, nil
		case "<a-3>", "<alt-3>":
			return KeyComb{Ch: '3', Mod: ModAlt}, nil
		case "<a-4>", "<alt-4>":
			return KeyComb{Ch: '4', Mod: ModAlt}, nil
		case "<a-5>", "<alt-5>":
			return KeyComb{Ch: '5', Mod: ModAlt}, nil
		case "<a-6>", "<alt-6>":
			return KeyComb{Ch: '6', Mod: ModAlt}, nil
		case "<a-7>", "<alt-7>":
			return KeyComb{Ch: '7', Mod: ModAlt}, nil
		case "<a-8>", "<alt-8>":
			return KeyComb{Ch: '8', Mod: ModAlt}, nil
		case "<a-9>", "<alt-9>":
			return KeyComb{Ch: '9', Mod: ModAlt}, nil
		case "<a-0>", "<alt-0>":
			return KeyComb{Ch: '0', Mod: ModAlt}, nil
		case "<a-`>", "<alt-`>":
			return KeyComb{Mod: ModAlt, Ch: '`'}, nil
		case "<a-a>", "<alt-a>":
			return KeyComb{Mod: ModAlt, Ch: 'a'}, nil
		case "<a-b>", "<alt-b>":
			return KeyComb{Mod: ModAlt, Ch: 'b'}, nil
		case "<a-c>", "<alt-c>":
			return KeyComb{Mod: ModAlt, Ch: 'c'}, nil
		case "<a-d>", "<alt-d>":
			return KeyComb{Mod: ModAlt, Ch: 'd'}, nil
		case "<a-e>", "<alt-e>":
			return KeyComb{Mod: ModAlt, Ch: 'e'}, nil
		case "<a-f>", "<alt-f>":
			return KeyComb{Mod: ModAlt, Ch: 'f'}, nil
		case "<a-g>", "<alt-g>":
			return KeyComb{Mod: ModAlt, Ch: 'g'}, nil
		case "<a-h>", "<alt-h>":
			return KeyComb{Mod: ModAlt, Ch: 'h'}, nil
		case "<a-tab>", "<alt-tab>":
			return KeyComb{Mod: ModAlt, Key: KeyTab}, nil
		case "<a-i>", "<alt-i>":
			return KeyComb{Mod: ModAlt, Ch: 'i'}, nil
		case "<a-j>", "<alt-j>":
			return KeyComb{Mod: ModAlt, Ch: 'j'}, nil
		case "<a-k>", "<alt-k>":
			return KeyComb{Mod: ModAlt, Ch: 'k'}, nil
		case "<a-l>", "<alt-l>":
			return KeyComb{Mod: ModAlt, Ch: 'l'}, nil
		case "<a-m>", "<alt-m>":
			return KeyComb{Mod: ModAlt, Ch: 'm'}, nil
		case "<a-n>", "<alt-n>":
			return KeyComb{Mod: ModAlt, Ch: 'n'}, nil
		case "<a-o>", "<alt-o>":
			return KeyComb{Mod: ModAlt, Ch: 'o'}, nil
		case "<a-p>", "<alt-p>":
			return KeyComb{Mod: ModAlt, Ch: 'p'}, nil
		case "<a-q>", "<alt-q>":
			return KeyComb{Mod: ModAlt, Ch: 'q'}, nil
		case "<a-r>", "<alt-r>":
			return KeyComb{Mod: ModAlt, Ch: 'r'}, nil
		case "<a-s>", "<alt-s>":
			return KeyComb{Mod: ModAlt, Ch: 's'}, nil
		case "<a-t>", "<alt-t>":
			return KeyComb{Mod: ModAlt, Ch: 't'}, nil
		case "<a-u>", "<alt-u>":
			return KeyComb{Mod: ModAlt, Ch: 'u'}, nil
		case "<a-v>", "<alt-v>":
			return KeyComb{Mod: ModAlt, Ch: 'v'}, nil
		case "<a-w>", "<alt-w>":
			return KeyComb{Mod: ModAlt, Ch: 'w'}, nil
		case "<a-x>", "<alt-x>":
			return KeyComb{Mod: ModAlt, Ch: 'x'}, nil
		case "<a-y>", "<alt-y>":
			return KeyComb{Mod: ModAlt, Ch: 'y'}, nil
		case "<a-z>", "<alt-z>":
			return KeyComb{Mod: ModAlt, Ch: 'z'}, nil
		case "<a-[>", "<alt-[>":
			return KeyComb{Mod: ModAlt, Ch: '['}, nil
		case "<a-\\\\>", "<alt-\\\\>":
			return KeyComb{Mod: ModAlt, Ch: '\\'}, nil
		case "<a-]>", "<alt-]>":
			return KeyComb{Mod: ModAlt, Ch: ']'}, nil
		case "<a-/>", "<alt-/>":
			return KeyComb{Mod: ModAlt, Ch: '/'}, nil
		case "<a-_>", "<alt-_>":
			return KeyComb{Mod: ModAlt, Ch: '_'}, nil
		case "<a-.>", "<alt-.>":
			return KeyComb{Mod: ModAlt, Ch: '.'}, nil
		case "<a-,>", "<alt-,>":
			return KeyComb{Mod: ModAlt, Ch: ','}, nil
		case "<a-;>", "<alt-;>":
			return KeyComb{Mod: ModAlt, Ch: ';'}, nil
		case "<a-'>", "<alt-'>":
			return KeyComb{Mod: ModAlt, Ch: '\''}, nil
		case "<a-=>", "<alt-=>":
			return KeyComb{Mod: ModAlt, Ch: '='}, nil
		case "<a-->", "<alt-->":
			return KeyComb{Mod: ModAlt, Ch: '-'}, nil

		// shift
		case "<s-f1>", "<shift-f1>":
			return KeyComb{Key: KeyF1, Mod: ModShift}, nil
		case "<s-f2>", "<shift-f2>":
			return KeyComb{Key: KeyF2, Mod: ModShift}, nil
		case "<s-f3>", "<shift-f3>":
			return KeyComb{Key: KeyF3, Mod: ModShift}, nil
		case "<s-f4>", "<shift-f4>":
			return KeyComb{Key: KeyF4, Mod: ModShift}, nil
		case "<s-f5>", "<shift-f5>":
			return KeyComb{Key: KeyF5, Mod: ModShift}, nil
		case "<s-f6>", "<shift-f6>":
			return KeyComb{Key: KeyF6, Mod: ModShift}, nil
		case "<s-f7>", "<shift-f7>":
			return KeyComb{Key: KeyF7, Mod: ModShift}, nil
		case "<s-f8>", "<shift-f8>":
			return KeyComb{Key: KeyF8, Mod: ModShift}, nil
		case "<s-f9>", "<shift-f9>":
			return KeyComb{Key: KeyF9, Mod: ModShift}, nil
		case "<s-f10>", "<shift-f10>":
			return KeyComb{Key: KeyF10, Mod: ModShift}, nil
		case "<s-f11>", "<shift-f11>":
			return KeyComb{Key: KeyF11, Mod: ModShift}, nil
		case "<s-f12>", "<shift-f12>":
			return KeyComb{Key: KeyF12, Mod: ModShift}, nil
		case "<s-insert>", "<shift-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModShift}, nil
		case "<s-delete>", "<shift-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModShift}, nil
		case "<s-home>", "<shift-home>":
			return KeyComb{Key: KeyHome, Mod: ModShift}, nil
		case "<s-end>", "<shift-end>":
			return KeyComb{Key: KeyEnd, Mod: ModShift}, nil
		case "<s-pgup>", "<shift-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModShift}, nil
		case "<s-pgdn>", "<shift-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModShift}, nil
		case "<s-up>", "<shift-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModShift}, nil
		case "<s-down>", "<shift-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModShift}, nil
		case "<s-left>", "<shift-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModShift}, nil
		case "<s-right>", "<shift-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModShift}, nil
		case "<s-mouse-left>", "<shift-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModShift}, nil
		case "<s-mouse-middle>", "<shift-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModShift}, nil
		case "<s-mouse-right>", "<shift-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModShift}, nil
		case "<s-mouse-release>", "<shift-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModShift}, nil
		case "<s-mouse-wheel-up>", "<shift-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModShift}, nil
		case "<s-mouse-wheel-down>", "<shift-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModShift}, nil
		case "<s-enter>", "<shift-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModShift}, nil
		case "<s-space>", "<shift-space>":
			return KeyComb{Mod: ModShift, Key: KeySpace}, nil
		case "<s-backspace>", "<shift-backspace>":
			return KeyComb{Mod: ModShift, Key: KeyBackspace}, nil
		case "<s-esc>", "<shift-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModShift}, nil
		case "<s-1>", "<shift-1>":
			return KeyComb{Ch: '!'}, nil
		case "<s-2>", "<shift-2>":
			return KeyComb{Ch: '@'}, nil
		case "<s-3>", "<shift-3>":
			return KeyComb{Ch: '#'}, nil
		case "<s-4>", "<shift-4>":
			return KeyComb{Ch: '$'}, nil
		case "<s-5>", "<shift-5>":
			return KeyComb{Ch: '%'}, nil
		case "<s-6>", "<shift-6>":
			return KeyComb{Ch: '^'}, nil
		case "<s-7>", "<shift-7>":
			return KeyComb{Ch: '&'}, nil
		case "<s-8>", "<shift-8>":
			return KeyComb{Ch: '*'}, nil
		case "<s-9>", "<shift-9>":
			return KeyComb{Ch: '('}, nil
		case "<s-0>", "<shift-0>":
			return KeyComb{Ch: ')'}, nil
		case "<s-`>", "<shift-`>":
			return KeyComb{Ch: '~'}, nil
		case "<s-a>", "<shift-a>":
			return KeyComb{Ch: 'A'}, nil
		case "<s-b>", "<shift-b>":
			return KeyComb{Ch: 'B'}, nil
		case "<s-c>", "<shift-c>":
			return KeyComb{Ch: 'C'}, nil
		case "<s-d>", "<shift-d>":
			return KeyComb{Ch: 'D'}, nil
		case "<s-e>", "<shift-e>":
			return KeyComb{Ch: 'E'}, nil
		case "<s-f>", "<shift-f>":
			return KeyComb{Ch: 'F'}, nil
		case "<s-g>", "<shift-g>":
			return KeyComb{Ch: 'G'}, nil
		case "<s-h>", "<shift-h>":
			return KeyComb{Ch: 'H'}, nil
		case "<s-tab>", "<shift-tab>":
			return KeyComb{Mod: ModShift, Key: KeyTab}, nil
		case "<s-i>", "<shift-i>":
			return KeyComb{Ch: 'I'}, nil
		case "<s-j>", "<shift-j>":
			return KeyComb{Ch: 'J'}, nil
		case "<s-k>", "<shift-k>":
			return KeyComb{Ch: 'K'}, nil
		case "<s-l>", "<shift-l>":
			return KeyComb{Ch: 'L'}, nil
		case "<s-m>", "<shift-m>":
			return KeyComb{Ch: 'M'}, nil
		case "<s-n>", "<shift-n>":
			return KeyComb{Ch: 'N'}, nil
		case "<s-o>", "<shift-o>":
			return KeyComb{Ch: 'O'}, nil
		case "<s-p>", "<shift-p>":
			return KeyComb{Ch: 'P'}, nil
		case "<s-q>", "<shift-q>":
			return KeyComb{Ch: 'Q'}, nil
		case "<s-r>", "<shift-r>":
			return KeyComb{Ch: 'R'}, nil
		case "<s-s>", "<shift-s>":
			return KeyComb{Ch: 'S'}, nil
		case "<s-t>", "<shift-t>":
			return KeyComb{Ch: 'T'}, nil
		case "<s-u>", "<shift-u>":
			return KeyComb{Ch: 'U'}, nil
		case "<s-v>", "<shift-v>":
			return KeyComb{Ch: 'V'}, nil
		case "<s-w>", "<shift-w>":
			return KeyComb{Ch: 'W'}, nil
		case "<s-x>", "<shift-x>":
			return KeyComb{Ch: 'X'}, nil
		case "<s-y>", "<shift-y>":
			return KeyComb{Ch: 'Y'}, nil
		case "<s-z>", "<shift-z>":
			return KeyComb{Ch: 'Z'}, nil
		case "<s-[>", "<shift-[>":
			return KeyComb{Ch: '{'}, nil
		case "<s-\\\\>", "<shift-\\\\>":
			return KeyComb{Ch: '|'}, nil
		case "<s-]>", "<shift-]>":
			return KeyComb{Ch: '}'}, nil
		case "<s-/>", "<shift-/>":
			return KeyComb{Ch: '?'}, nil
		case "<s-_>", "<shift-_>":
			return KeyComb{Ch: '_'}, nil
		case "<s-.>", "<shift-.>":
			return KeyComb{Ch: '>'}, nil
		case "<s-,>", "<shift-,>":
			return KeyComb{Ch: '<'}, nil
		case "<s-;>", "<shift-;>":
			return KeyComb{Ch: ':'}, nil
		case "<s-'>", "<shift-'>":
			return KeyComb{Ch: '"'}, nil
		case "<s-=>", "<shift-=>":
			return KeyComb{Ch: '+'}, nil
		case "<s-+>", "<shift-+>":
			return KeyComb{Ch: '+'}, nil
		case "<s-->", "<shift-->":
			return KeyComb{Ch: '_'}, nil

		// ctrl
		case "<c-f1>", "<ctrl-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrl}, nil
		case "<c-f2>", "<ctrl-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrl}, nil
		case "<c-f3>", "<ctrl-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrl}, nil
		case "<c-f4>", "<ctrl-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrl}, nil
		case "<c-f5>", "<ctrl-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrl}, nil
		case "<c-f6>", "<ctrl-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrl}, nil
		case "<c-f7>", "<ctrl-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrl}, nil
		case "<c-f8>", "<ctrl-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrl}, nil
		case "<c-f9>", "<ctrl-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrl}, nil
		case "<c-f10>", "<ctrl-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrl}, nil
		case "<c-f11>", "<ctrl-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrl}, nil
		case "<c-f12>", "<ctrl-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrl}, nil
		case "<c-insert>", "<ctrl-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrl}, nil
		case "<c-delete>", "<ctrl-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrl}, nil
		case "<c-home>", "<ctrl-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrl}, nil
		case "<c-end>", "<ctrl-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrl}, nil
		case "<c-pgup>", "<ctrl-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrl}, nil
		case "<c-pgdn>", "<ctrl-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrl}, nil
		case "<c-up>", "<ctrl-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrl}, nil
		case "<c-down>", "<ctrl-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrl}, nil
		case "<c-left>", "<ctrl-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrl}, nil
		case "<c-right>", "<ctrl-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrl}, nil
		case "<c-mouse-left>", "<ctrl-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrl}, nil
		case "<c-mouse-middle>", "<ctrl-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrl}, nil
		case "<c-mouse-right>", "<ctrl-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrl}, nil
		case "<c-mouse-release>", "<ctrl-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrl}, nil
		case "<c-mouse-wheel-up>", "<ctrl-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrl}, nil
		case "<c-mouse-wheel-down>", "<ctrl-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrl}, nil
		case "<c-enter>", "<ctrl-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrl}, nil
		case "<c-space>", "<ctrl-space>":
			return KeyComb{Mod: ModCtrl, Key: KeySpace}, nil
		case "<c-backspace>", "<ctrl-backspace>":
			return KeyComb{Mod: ModCtrl, Key: KeyBackspace}, nil
		case "<c-esc>", "<ctrl-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrl}, nil
		case "<c-1>", "<ctrl-1>":
			return KeyComb{Ch: '1', Mod: ModCtrl}, nil
		case "<c-2>", "<ctrl-2>":
			return KeyComb{Mod: ModCtrl, Ch: '2'}, nil
		case "<c-3>", "<ctrl-3>":
			return KeyComb{Ch: '3', Mod: ModCtrl}, nil
		case "<c-4>", "<ctrl-4>":
			return KeyComb{Ch: '4', Mod: ModCtrl}, nil
		case "<c-5>", "<ctrl-5>":
			return KeyComb{Ch: '5', Mod: ModCtrl}, nil
		case "<c-6>", "<ctrl-6>":
			return KeyComb{Ch: '6', Mod: ModCtrl}, nil
		case "<c-7>", "<ctrl-7>":
			return KeyComb{Ch: '7', Mod: ModCtrl}, nil
		case "<c-8>", "<ctrl-8>":
			return KeyComb{Ch: '8', Mod: ModCtrl}, nil
		case "<c-9>", "<ctrl-9>":
			return KeyComb{Ch: '9', Mod: ModCtrl}, nil
		case "<c-0>", "<ctrl-0>":
			return KeyComb{Ch: '0', Mod: ModCtrl}, nil
		case "<c-`>", "<ctrl-`>":
			return KeyComb{Mod: ModCtrl, Ch: '`'}, nil
		case "<c-a>", "<ctrl-a>":
			return KeyComb{Mod: ModCtrl, Ch: 'a'}, nil
		case "<c-b>", "<ctrl-b>":
			return KeyComb{Mod: ModCtrl, Ch: 'b'}, nil
		case "<c-c>", "<ctrl-c>":
			return KeyComb{Mod: ModCtrl, Ch: 'c'}, nil
		case "<c-d>", "<ctrl-d>":
			return KeyComb{Mod: ModCtrl, Ch: 'd'}, nil
		case "<c-e>", "<ctrl-e>":
			return KeyComb{Mod: ModCtrl, Ch: 'e'}, nil
		case "<c-f>", "<ctrl-f>":
			return KeyComb{Mod: ModCtrl, Ch: 'f'}, nil
		case "<c-g>", "<ctrl-g>":
			return KeyComb{Mod: ModCtrl, Ch: 'g'}, nil
		case "<c-h>", "<ctrl-h>":
			return KeyComb{Mod: ModCtrl, Ch: 'h'}, nil
		case "<c-tab>", "<ctrl-tab>":
			return KeyComb{Mod: ModCtrl, Key: KeyTab}, nil
		case "<c-i>", "<ctrl-i>":
			return KeyComb{Mod: ModCtrl, Ch: 'i'}, nil
		case "<c-j>", "<ctrl-j>":
			return KeyComb{Mod: ModCtrl, Ch: 'j'}, nil
		case "<c-k>", "<ctrl-k>":
			return KeyComb{Mod: ModCtrl, Ch: 'k'}, nil
		case "<c-l>", "<ctrl-l>":
			return KeyComb{Mod: ModCtrl, Ch: 'l'}, nil
		case "<c-m>", "<ctrl-m>":
			return KeyComb{Mod: ModCtrl, Ch: 'm'}, nil
		case "<c-n>", "<ctrl-n>":
			return KeyComb{Mod: ModCtrl, Ch: 'n'}, nil
		case "<c-o>", "<ctrl-o>":
			return KeyComb{Mod: ModCtrl, Ch: 'o'}, nil
		case "<c-p>", "<ctrl-p>":
			return KeyComb{Mod: ModCtrl, Ch: 'p'}, nil
		case "<c-q>", "<ctrl-q>":
			return KeyComb{Mod: ModCtrl, Ch: 'q'}, nil
		case "<c-r>", "<ctrl-r>":
			return KeyComb{Mod: ModCtrl, Ch: 'r'}, nil
		case "<c-s>", "<ctrl-s>":
			return KeyComb{Mod: ModCtrl, Ch: 's'}, nil
		case "<c-t>", "<ctrl-t>":
			return KeyComb{Mod: ModCtrl, Ch: 't'}, nil
		case "<c-u>", "<ctrl-u>":
			return KeyComb{Mod: ModCtrl, Ch: 'u'}, nil
		case "<c-v>", "<ctrl-v>":
			return KeyComb{Mod: ModCtrl, Ch: 'v'}, nil
		case "<c-w>", "<ctrl-w>":
			return KeyComb{Mod: ModCtrl, Ch: 'w'}, nil
		case "<c-x>", "<ctrl-x>":
			return KeyComb{Mod: ModCtrl, Ch: 'x'}, nil
		case "<c-y>", "<ctrl-y>":
			return KeyComb{Mod: ModCtrl, Ch: 'y'}, nil
		case "<c-z>", "<ctrl-z>":
			return KeyComb{Mod: ModCtrl, Ch: 'z'}, nil
		case "<c-[>", "<ctrl-[>":
			return KeyComb{Mod: ModCtrl, Ch: '['}, nil
		case "<c-\\\\>", "<ctrl-\\\\>":
			return KeyComb{Mod: ModCtrl, Ch: '\\'}, nil
		case "<c-]>", "<ctrl-]>":
			return KeyComb{Mod: ModCtrl, Ch: ']'}, nil
		case "<c-/>", "<ctrl-/>":
			return KeyComb{Mod: ModCtrl, Ch: '/'}, nil
		case "<c-_>", "<ctrl-_>":
			return KeyComb{Mod: ModCtrl, Ch: '_'}, nil
		case "<c-.>", "<ctrl-.>":
			return KeyComb{Mod: ModCtrl, Ch: '.'}, nil
		case "<c-,>", "<ctrl-,>":
			return KeyComb{Mod: ModCtrl, Ch: ','}, nil
		case "<c-;>", "<ctrl-;>":
			return KeyComb{Mod: ModCtrl, Ch: ';'}, nil
		case "<c-'>", "<ctrl-'>":
			return KeyComb{Mod: ModCtrl, Ch: '\''}, nil
		case "<c-=>", "<ctrl-=>":
			return KeyComb{Mod: ModCtrl, Ch: '='}, nil
		case "<c-->", "<ctrl-->":
			return KeyComb{Mod: ModCtrl, Ch: '-'}, nil

		// ctrl+shift
		case "<c-s-f1>", "<ctrl-shift-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrlShift}, nil
		case "<c-s-f2>", "<ctrl-shift-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrlShift}, nil
		case "<c-s-f3>", "<ctrl-shift-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrlShift}, nil
		case "<c-s-f4>", "<ctrl-shift-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrlShift}, nil
		case "<c-s-f5>", "<ctrl-shift-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrlShift}, nil
		case "<c-s-f6>", "<ctrl-shift-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrlShift}, nil
		case "<c-s-f7>", "<ctrl-shift-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrlShift}, nil
		case "<c-s-f8>", "<ctrl-shift-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrlShift}, nil
		case "<c-s-f9>", "<ctrl-shift-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrlShift}, nil
		case "<c-s-f10>", "<ctrl-shift-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrlShift}, nil
		case "<c-s-f11>", "<ctrl-shift-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrlShift}, nil
		case "<c-s-f12>", "<ctrl-shift-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrlShift}, nil
		case "<c-s-insert>", "<ctrl-shift-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrlShift}, nil
		case "<c-s-delete>", "<ctrl-shift-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrlShift}, nil
		case "<c-s-home>", "<ctrl-shift-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrlShift}, nil
		case "<c-s-end>", "<ctrl-shift-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrlShift}, nil
		case "<c-s-pgup>", "<ctrl-shift-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrlShift}, nil
		case "<c-s-pgdn>", "<ctrl-shift-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrlShift}, nil
		case "<c-s-up>", "<ctrl-shift-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrlShift}, nil
		case "<c-s-down>", "<ctrl-shift-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrlShift}, nil
		case "<c-s-left>", "<ctrl-shift-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShift}, nil
		case "<c-s-right>", "<ctrl-shift-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrlShift}, nil
		case "<c-s-mouse-left>", "<ctrl-shift-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrlShift}, nil
		case "<c-s-mouse-middle>", "<ctrl-shift-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrlShift}, nil
		case "<c-s-mouse-right>", "<ctrl-shift-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrlShift}, nil
		case "<c-s-mouse-release>", "<ctrl-shift-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrlShift}, nil
		case "<c-s-mouse-wheel-up>", "<ctrl-shift-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrlShift}, nil
		case "<c-s-mouse-wheel-down>", "<ctrl-shift-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrlShift}, nil
		case "<c-s-enter>", "<ctrl-shift-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrlShift}, nil
		case "<c-s-space>", "<ctrl-shift-space>":
			return KeyComb{Mod: ModCtrlShift, Key: KeySpace}, nil
		case "<c-s-backspace>", "<ctrl-shift-backspace>":
			return KeyComb{Mod: ModCtrlShift, Key: KeyBackspace}, nil
		case "<c-s-esc>", "<ctrl-shift-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrlShift}, nil
		case "<c-s-1>", "<ctrl-shift-1>":
			return KeyComb{Ch: '!', Mod: ModCtrl}, nil
		case "<c-s-2>", "<ctrl-shift-2>":
			return KeyComb{Mod: ModCtrl, Ch: '@'}, nil
		case "<c-s-3>", "<ctrl-shift-3>":
			return KeyComb{Ch: '#', Mod: ModCtrl}, nil
		case "<c-s-4>", "<ctrl-shift-4>":
			return KeyComb{Ch: '$', Mod: ModCtrl}, nil
		case "<c-s-5>", "<ctrl-shift-5>":
			return KeyComb{Ch: '%', Mod: ModCtrl}, nil
		case "<c-s-6>", "<ctrl-shift-6>":
			return KeyComb{Ch: '^', Mod: ModCtrl}, nil
		case "<c-s-7>", "<ctrl-shift-7>":
			return KeyComb{Ch: '&', Mod: ModCtrl}, nil
		case "<c-s-8>", "<ctrl-shift-8>":
			return KeyComb{Ch: '*', Mod: ModCtrl}, nil
		case "<c-s-9>", "<ctrl-shift-9>":
			return KeyComb{Ch: '(', Mod: ModCtrl}, nil
		case "<c-s-0>", "<ctrl-shift-0>":
			return KeyComb{Ch: ')', Mod: ModCtrl}, nil
		case "<c-s-`>", "<ctrl-shift-`>":
			return KeyComb{Mod: ModCtrl, Ch: '~'}, nil
		case "<c-s-a>", "<ctrl-shift-a>":
			return KeyComb{Mod: ModCtrl, Ch: 'A'}, nil
		case "<c-s-b>", "<ctrl-shift-b>":
			return KeyComb{Mod: ModCtrl, Ch: 'B'}, nil
		case "<c-s-c>", "<ctrl-shift-c>":
			return KeyComb{Mod: ModCtrl, Ch: 'C'}, nil
		case "<c-s-d>", "<ctrl-shift-d>":
			return KeyComb{Mod: ModCtrl, Ch: 'D'}, nil
		case "<c-s-e>", "<ctrl-shift-e>":
			return KeyComb{Mod: ModCtrl, Ch: 'E'}, nil
		case "<c-s-f>", "<ctrl-shift-f>":
			return KeyComb{Mod: ModCtrl, Ch: 'F'}, nil
		case "<c-s-g>", "<ctrl-shift-g>":
			return KeyComb{Mod: ModCtrl, Ch: 'G'}, nil
		case "<c-s-h>", "<ctrl-shift-h>":
			return KeyComb{Mod: ModCtrl, Ch: 'H'}, nil
		case "<c-s-tab>", "<ctrl-shift-tab>":
			return KeyComb{Mod: ModCtrlShift, Key: KeyTab}, nil
		case "<c-s-i>", "<ctrl-shift-i>":
			return KeyComb{Mod: ModCtrl, Ch: 'I'}, nil
		case "<c-s-j>", "<ctrl-shift-j>":
			return KeyComb{Mod: ModCtrl, Ch: 'J'}, nil
		case "<c-s-k>", "<ctrl-shift-k>":
			return KeyComb{Mod: ModCtrl, Ch: 'K'}, nil
		case "<c-s-l>", "<ctrl-shift-l>":
			return KeyComb{Mod: ModCtrl, Ch: 'L'}, nil
		case "<c-s-m>", "<ctrl-shift-m>":
			return KeyComb{Mod: ModCtrl, Ch: 'M'}, nil
		case "<c-s-n>", "<ctrl-shift-n>":
			return KeyComb{Mod: ModCtrl, Ch: 'N'}, nil
		case "<c-s-o>", "<ctrl-shift-o>":
			return KeyComb{Mod: ModCtrl, Ch: 'O'}, nil
		case "<c-s-p>", "<ctrl-shift-p>":
			return KeyComb{Mod: ModCtrl, Ch: 'P'}, nil
		case "<c-s-q>", "<ctrl-shift-q>":
			return KeyComb{Mod: ModCtrl, Ch: 'Q'}, nil
		case "<c-s-r>", "<ctrl-shift-r>":
			return KeyComb{Mod: ModCtrl, Ch: 'R'}, nil
		case "<c-s-s>", "<ctrl-shift-s>":
			return KeyComb{Mod: ModCtrl, Ch: 'S'}, nil
		case "<c-s-t>", "<ctrl-shift-t>":
			return KeyComb{Mod: ModCtrl, Ch: 'T'}, nil
		case "<c-s-u>", "<ctrl-shift-u>":
			return KeyComb{Mod: ModCtrl, Ch: 'U'}, nil
		case "<c-s-v>", "<ctrl-shift-v>":
			return KeyComb{Mod: ModCtrl, Ch: 'V'}, nil
		case "<c-s-w>", "<ctrl-shift-w>":
			return KeyComb{Mod: ModCtrl, Ch: 'W'}, nil
		case "<c-s-x>", "<ctrl-shift-x>":
			return KeyComb{Mod: ModCtrl, Ch: 'X'}, nil
		case "<c-s-y>", "<ctrl-shift-y>":
			return KeyComb{Mod: ModCtrl, Ch: 'Y'}, nil
		case "<c-s-z>", "<ctrl-shift-z>":
			return KeyComb{Mod: ModCtrl, Ch: 'Z'}, nil
		case "<c-s-[>", "<ctrl-shift-[>":
			return KeyComb{Mod: ModCtrl, Ch: '{'}, nil
		case "<c-s-\\\\>", "<ctrl-shift-\\\\>":
			return KeyComb{Mod: ModCtrl, Ch: '|'}, nil
		case "<c-s-]>", "<ctrl-shift-]>":
			return KeyComb{Mod: ModCtrl, Ch: '}'}, nil
		case "<c-s-/>", "<ctrl-shift-/>":
			return KeyComb{Mod: ModCtrl, Ch: '?'}, nil
		case "<c-s-_>", "<ctrl-shift-_>":
			return KeyComb{Mod: ModCtrl, Ch: '_'}, nil
		case "<c-s-.>", "<ctrl-shift-.>":
			return KeyComb{Mod: ModCtrl, Ch: '>'}, nil
		case "<c-s-,>", "<ctrl-shift-,>":
			return KeyComb{Mod: ModCtrl, Ch: '<'}, nil
		case "<c-s-;>", "<ctrl-shift-;>":
			return KeyComb{Mod: ModCtrl, Ch: ':'}, nil
		case "<c-s-'>", "<ctrl-shift-'>":
			return KeyComb{Mod: ModCtrl, Ch: '"'}, nil
		case "<c-s-=>", "<ctrl-shift-=>":
			return KeyComb{Mod: ModCtrl, Ch: '+'}, nil
		case "<c-s-+>", "<ctrl-shift-+>":
			return KeyComb{Mod: ModCtrl, Ch: '+'}, nil
		case "<c-+>", "<ctrl-+>":
			return KeyComb{Mod: ModCtrl, Ch: '+'}, nil
		case "<c-s-->", "<ctrl-shift-->":
			return KeyComb{Mod: ModCtrl, Ch: '_'}, nil

		// ctrl+alt
		case "<c-a-f1>", "<ctrl-alt-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrlAlt}, nil
		case "<c-a-f2>", "<ctrl-alt-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrlAlt}, nil
		case "<c-a-f3>", "<ctrl-alt-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrlAlt}, nil
		case "<c-a-f4>", "<ctrl-alt-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrlAlt}, nil
		case "<c-a-f5>", "<ctrl-alt-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrlAlt}, nil
		case "<c-a-f6>", "<ctrl-alt-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrlAlt}, nil
		case "<c-a-f7>", "<ctrl-alt-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrlAlt}, nil
		case "<c-a-f8>", "<ctrl-alt-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrlAlt}, nil
		case "<c-a-f9>", "<ctrl-alt-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrlAlt}, nil
		case "<c-a-f10>", "<ctrl-alt-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrlAlt}, nil
		case "<c-a-f11>", "<ctrl-alt-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrlAlt}, nil
		case "<c-a-f12>", "<ctrl-alt-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrlAlt}, nil
		case "<c-a-insert>", "<ctrl-alt-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrlAlt}, nil
		case "<c-a-delete>", "<ctrl-alt-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrlAlt}, nil
		case "<c-a-home>", "<ctrl-alt-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrlAlt}, nil
		case "<c-a-end>", "<ctrl-alt-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrlAlt}, nil
		case "<c-a-pgup>", "<ctrl-alt-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrlAlt}, nil
		case "<c-a-pgdn>", "<ctrl-alt-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrlAlt}, nil
		case "<c-a-up>", "<ctrl-alt-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrlAlt}, nil
		case "<c-a-down>", "<ctrl-alt-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrlAlt}, nil
		case "<c-a-left>", "<ctrl-alt-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrlAlt}, nil
		case "<c-a-right>", "<ctrl-alt-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrlAlt}, nil
		case "<c-a-mouse-left>", "<ctrl-alt-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrlAlt}, nil
		case "<c-a-mouse-middle>", "<ctrl-alt-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrlAlt}, nil
		case "<c-a-mouse-right>", "<ctrl-alt-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrlAlt}, nil
		case "<c-a-mouse-release>", "<ctrl-alt-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrlAlt}, nil
		case "<c-a-mouse-wheel-up>", "<ctrl-alt-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrlAlt}, nil
		case "<c-a-mouse-wheel-down>", "<ctrl-alt-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrlAlt}, nil
		case "<c-a-enter>", "<ctrl-alt-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrlAlt}, nil
		case "<c-a-space>", "<ctrl-alt-space>":
			return KeyComb{Mod: ModCtrlAlt, Key: KeySpace}, nil
		case "<c-a-backspace>", "<ctrl-alt-backspace>":
			return KeyComb{Mod: ModCtrlAlt, Key: KeyBackspace}, nil
		case "<c-a-esc>", "<ctrl-alt-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrlAlt}, nil
		case "<c-a-1>", "<ctrl-alt-1>":
			return KeyComb{Ch: '1', Mod: ModCtrlAlt}, nil
		case "<c-a-2>", "<ctrl-alt-2>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '2'}, nil
		case "<c-a-3>", "<ctrl-alt-3>":
			return KeyComb{Ch: '3', Mod: ModCtrlAlt}, nil
		case "<c-a-4>", "<ctrl-alt-4>":
			return KeyComb{Ch: '4', Mod: ModCtrlAlt}, nil
		case "<c-a-5>", "<ctrl-alt-5>":
			return KeyComb{Ch: '5', Mod: ModCtrlAlt}, nil
		case "<c-a-6>", "<ctrl-alt-6>":
			return KeyComb{Ch: '6', Mod: ModCtrlAlt}, nil
		case "<c-a-7>", "<ctrl-alt-7>":
			return KeyComb{Ch: '7', Mod: ModCtrlAlt}, nil
		case "<c-a-8>", "<ctrl-alt-8>":
			return KeyComb{Ch: '8', Mod: ModCtrlAlt}, nil
		case "<c-a-9>", "<ctrl-alt-9>":
			return KeyComb{Ch: '9', Mod: ModCtrlAlt}, nil
		case "<c-a-0>", "<ctrl-alt-0>":
			return KeyComb{Ch: '0', Mod: ModCtrlAlt}, nil
		case "<c-a-`>", "<ctrl-alt-`>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '`'}, nil
		case "<c-a-a>", "<ctrl-alt-a>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'a'}, nil
		case "<c-a-b>", "<ctrl-alt-b>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'b'}, nil
		case "<c-a-c>", "<ctrl-alt-c>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'c'}, nil
		case "<c-a-d>", "<ctrl-alt-d>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'd'}, nil
		case "<c-a-e>", "<ctrl-alt-e>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'e'}, nil
		case "<c-a-f>", "<ctrl-alt-f>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'f'}, nil
		case "<c-a-g>", "<ctrl-alt-g>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'g'}, nil
		case "<c-a-h>", "<ctrl-alt-h>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'h'}, nil
		case "<c-a-tab>", "<ctrl-alt-tab>":
			return KeyComb{Mod: ModCtrlAlt, Key: KeyTab}, nil
		case "<c-a-i>", "<ctrl-alt-i>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'i'}, nil
		case "<c-a-j>", "<ctrl-alt-j>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'j'}, nil
		case "<c-a-k>", "<ctrl-alt-k>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'k'}, nil
		case "<c-a-l>", "<ctrl-alt-l>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'l'}, nil
		case "<c-a-m>", "<ctrl-alt-m>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'm'}, nil
		case "<c-a-n>", "<ctrl-alt-n>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'n'}, nil
		case "<c-a-o>", "<ctrl-alt-o>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'o'}, nil
		case "<c-a-p>", "<ctrl-alt-p>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'p'}, nil
		case "<c-a-q>", "<ctrl-alt-q>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'q'}, nil
		case "<c-a-r>", "<ctrl-alt-r>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'r'}, nil
		case "<c-a-s>", "<ctrl-alt-s>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 's'}, nil
		case "<c-a-t>", "<ctrl-alt-t>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 't'}, nil
		case "<c-a-u>", "<ctrl-alt-u>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'u'}, nil
		case "<c-a-v>", "<ctrl-alt-v>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'v'}, nil
		case "<c-a-w>", "<ctrl-alt-w>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'w'}, nil
		case "<c-a-x>", "<ctrl-alt-x>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'x'}, nil
		case "<c-a-y>", "<ctrl-alt-y>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'y'}, nil
		case "<c-a-z>", "<ctrl-alt-z>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'z'}, nil
		case "<c-a-[>", "<ctrl-alt-[>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '['}, nil
		case "<c-a-\\\\>", "<ctrl-alt-\\\\>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '\\'}, nil
		case "<c-a-]>", "<ctrl-alt-]>":
			return KeyComb{Mod: ModCtrlAlt, Ch: ']'}, nil
		case "<c-a-/>", "<ctrl-alt-/>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '/'}, nil
		case "<c-a-_>", "<ctrl-alt-_>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '_'}, nil
		case "<c-a-.>", "<ctrl-alt-.>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '.'}, nil
		case "<c-a-,>", "<ctrl-alt-,>":
			return KeyComb{Mod: ModCtrlAlt, Ch: ','}, nil
		case "<c-a-;>", "<ctrl-alt-;>":
			return KeyComb{Mod: ModCtrlAlt, Ch: ';'}, nil
		case "<c-a-'>", "<ctrl-alt-'>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '\''}, nil
		case "<c-a-=>", "<ctrl-alt-=>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '='}, nil
		case "<c-a-->", "<ctrl-alt-->":
			return KeyComb{Mod: ModCtrlAlt, Ch: '-'}, nil

		// ctrl+meta
		case "<c-m-f1>", "<ctrl-meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrlMeta}, nil
		case "<c-m-f2>", "<ctrl-meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrlMeta}, nil
		case "<c-m-f3>", "<ctrl-meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrlMeta}, nil
		case "<c-m-f4>", "<ctrl-meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrlMeta}, nil
		case "<c-m-f5>", "<ctrl-meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrlMeta}, nil
		case "<c-m-f6>", "<ctrl-meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrlMeta}, nil
		case "<c-m-f7>", "<ctrl-meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrlMeta}, nil
		case "<c-m-f8>", "<ctrl-meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrlMeta}, nil
		case "<c-m-f9>", "<ctrl-meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrlMeta}, nil
		case "<c-m-f10>", "<ctrl-meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrlMeta}, nil
		case "<c-m-f11>", "<ctrl-meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrlMeta}, nil
		case "<c-m-f12>", "<ctrl-meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrlMeta}, nil
		case "<c-m-insert>", "<ctrl-meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrlMeta}, nil
		case "<c-m-delete>", "<ctrl-meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrlMeta}, nil
		case "<c-m-home>", "<ctrl-meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrlMeta}, nil
		case "<c-m-end>", "<ctrl-meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrlMeta}, nil
		case "<c-m-pgup>", "<ctrl-meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrlMeta}, nil
		case "<c-m-pgdn>", "<ctrl-meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrlMeta}, nil
		case "<c-m-up>", "<ctrl-meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrlMeta}, nil
		case "<c-m-down>", "<ctrl-meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrlMeta}, nil
		case "<c-m-left>", "<ctrl-meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrlMeta}, nil
		case "<c-m-right>", "<ctrl-meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrlMeta}, nil
		case "<c-m-mouse-left>", "<ctrl-meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrlMeta}, nil
		case "<c-m-mouse-middle>", "<ctrl-meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrlMeta}, nil
		case "<c-m-mouse-right>", "<ctrl-meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrlMeta}, nil
		case "<c-m-mouse-release>", "<ctrl-meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrlMeta}, nil
		case "<c-m-mouse-wheel-up>", "<ctrl-meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrlMeta}, nil
		case "<c-m-mouse-wheel-down>", "<ctrl-meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrlMeta}, nil
		case "<c-m-enter>", "<ctrl-meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrlMeta}, nil
		case "<c-m-space>", "<ctrl-meta-space>":
			return KeyComb{Mod: ModCtrlMeta, Key: KeySpace}, nil
		case "<c-m-backspace>", "<ctrl-meta-backspace>":
			return KeyComb{Mod: ModCtrlMeta, Key: KeyBackspace}, nil
		case "<c-m-esc>", "<ctrl-meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrlMeta}, nil
		case "<c-m-1>", "<ctrl-meta-1>":
			return KeyComb{Ch: '1', Mod: ModCtrlMeta}, nil
		case "<c-m-2>", "<ctrl-meta-2>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '2'}, nil
		case "<c-m-3>", "<ctrl-meta-3>":
			return KeyComb{Ch: '3', Mod: ModCtrlMeta}, nil
		case "<c-m-4>", "<ctrl-meta-4>":
			return KeyComb{Ch: '4', Mod: ModCtrlMeta}, nil
		case "<c-m-5>", "<ctrl-meta-5>":
			return KeyComb{Ch: '5', Mod: ModCtrlMeta}, nil
		case "<c-m-6>", "<ctrl-meta-6>":
			return KeyComb{Ch: '6', Mod: ModCtrlMeta}, nil
		case "<c-m-7>", "<ctrl-meta-7>":
			return KeyComb{Ch: '7', Mod: ModCtrlMeta}, nil
		case "<c-m-8>", "<ctrl-meta-8>":
			return KeyComb{Ch: '8', Mod: ModCtrlMeta}, nil
		case "<c-m-9>", "<ctrl-meta-9>":
			return KeyComb{Ch: '9', Mod: ModCtrlMeta}, nil
		case "<c-m-0>", "<ctrl-meta-0>":
			return KeyComb{Ch: '0', Mod: ModCtrlMeta}, nil
		case "<c-m-`>", "<ctrl-meta-`>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '`'}, nil
		case "<c-m-a>", "<ctrl-meta-a>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'a'}, nil
		case "<c-m-b>", "<ctrl-meta-b>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'b'}, nil
		case "<c-m-c>", "<ctrl-meta-c>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'c'}, nil
		case "<c-m-d>", "<ctrl-meta-d>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'd'}, nil
		case "<c-m-e>", "<ctrl-meta-e>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'e'}, nil
		case "<c-m-f>", "<ctrl-meta-f>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'f'}, nil
		case "<c-m-g>", "<ctrl-meta-g>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'g'}, nil
		case "<c-m-h>", "<ctrl-meta-h>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'h'}, nil
		case "<c-m-tab>", "<ctrl-meta-tab>":
			return KeyComb{Mod: ModCtrlMeta, Key: KeyTab}, nil
		case "<c-m-i>", "<ctrl-meta-i>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'i'}, nil
		case "<c-m-j>", "<ctrl-meta-j>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'j'}, nil
		case "<c-m-k>", "<ctrl-meta-k>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'k'}, nil
		case "<c-m-l>", "<ctrl-meta-l>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'l'}, nil
		case "<c-m-m>", "<ctrl-meta-m>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'm'}, nil
		case "<c-m-n>", "<ctrl-meta-n>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'n'}, nil
		case "<c-m-o>", "<ctrl-meta-o>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'o'}, nil
		case "<c-m-p>", "<ctrl-meta-p>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'p'}, nil
		case "<c-m-q>", "<ctrl-meta-q>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'q'}, nil
		case "<c-m-r>", "<ctrl-meta-r>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'r'}, nil
		case "<c-m-s>", "<ctrl-meta-s>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 's'}, nil
		case "<c-m-t>", "<ctrl-meta-t>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 't'}, nil
		case "<c-m-u>", "<ctrl-meta-u>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'u'}, nil
		case "<c-m-v>", "<ctrl-meta-v>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'v'}, nil
		case "<c-m-w>", "<ctrl-meta-w>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'w'}, nil
		case "<c-m-x>", "<ctrl-meta-x>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'x'}, nil
		case "<c-m-y>", "<ctrl-meta-y>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'y'}, nil
		case "<c-m-z>", "<ctrl-meta-z>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'z'}, nil
		case "<c-m-[>", "<ctrl-meta-[>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '['}, nil
		case "<c-m-\\\\>", "<ctrl-meta-\\\\>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '\\'}, nil
		case "<c-m-]>", "<ctrl-meta-]>":
			return KeyComb{Mod: ModCtrlMeta, Ch: ']'}, nil
		case "<c-m-/>", "<ctrl-meta-/>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '/'}, nil
		case "<c-m-_>", "<ctrl-meta-_>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '_'}, nil
		case "<c-m-.>", "<ctrl-meta-.>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '.'}, nil
		case "<c-m-,>", "<ctrl-meta-,>":
			return KeyComb{Mod: ModCtrlMeta, Ch: ','}, nil
		case "<c-m-;>", "<ctrl-meta-;>":
			return KeyComb{Mod: ModCtrlMeta, Ch: ';'}, nil
		case "<c-m-'>", "<ctrl-meta-'>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '\''}, nil
		case "<c-m-=>", "<ctrl-meta-=>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '='}, nil
		case "<c-m-->", "<ctrl-meta-->":
			return KeyComb{Mod: ModCtrlMeta, Ch: '-'}, nil

		// ctrl+shift+alt
		case "<c-s-a-f1>", "<ctrl-shift-alt-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f2>", "<ctrl-shift-alt-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f3>", "<ctrl-shift-alt-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f4>", "<ctrl-shift-alt-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f5>", "<ctrl-shift-alt-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f6>", "<ctrl-shift-alt-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f7>", "<ctrl-shift-alt-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f8>", "<ctrl-shift-alt-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f9>", "<ctrl-shift-alt-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f10>", "<ctrl-shift-alt-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f11>", "<ctrl-shift-alt-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-f12>", "<ctrl-shift-alt-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-insert>", "<ctrl-shift-alt-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-delete>", "<ctrl-shift-alt-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-home>", "<ctrl-shift-alt-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-end>", "<ctrl-shift-alt-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-pgup>", "<ctrl-shift-alt-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-pgdn>", "<ctrl-shift-alt-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-up>", "<ctrl-shift-alt-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-down>", "<ctrl-shift-alt-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-left>", "<ctrl-shift-alt-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-right>", "<ctrl-shift-alt-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-mouse-left>", "<ctrl-shift-alt-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-mouse-middle>", "<ctrl-shift-alt-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-mouse-right>", "<ctrl-shift-alt-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-mouse-release>", "<ctrl-shift-alt-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-mouse-wheel-up>", "<ctrl-shift-alt-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-mouse-wheel-down>", "<ctrl-shift-alt-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-enter>", "<ctrl-shift-alt-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-space>", "<ctrl-shift-alt-space>":
			return KeyComb{Mod: ModCtrlShiftAlt, Key: KeySpace}, nil
		case "<c-s-a-backspace>", "<ctrl-shift-alt-backspace>":
			return KeyComb{Mod: ModCtrlShiftAlt, Key: KeyBackspace}, nil
		case "<c-s-a-esc>", "<ctrl-shift-alt-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrlShiftAlt}, nil
		case "<c-s-a-1>", "<ctrl-shift-alt-1>":
			return KeyComb{Ch: '!', Mod: ModCtrlAlt}, nil
		case "<c-s-a-2>", "<ctrl-shift-alt-2>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '@'}, nil
		case "<c-s-a-3>", "<ctrl-shift-alt-3>":
			return KeyComb{Ch: '#', Mod: ModCtrlAlt}, nil
		case "<c-s-a-4>", "<ctrl-shift-alt-4>":
			return KeyComb{Ch: '$', Mod: ModCtrlAlt}, nil
		case "<c-s-a-5>", "<ctrl-shift-alt-5>":
			return KeyComb{Ch: '%', Mod: ModCtrlAlt}, nil
		case "<c-s-a-6>", "<ctrl-shift-alt-6>":
			return KeyComb{Ch: '^', Mod: ModCtrlAlt}, nil
		case "<c-s-a-7>", "<ctrl-shift-alt-7>":
			return KeyComb{Ch: '&', Mod: ModCtrlAlt}, nil
		case "<c-s-a-8>", "<ctrl-shift-alt-8>":
			return KeyComb{Ch: '*', Mod: ModCtrlAlt}, nil
		case "<c-s-a-9>", "<ctrl-shift-alt-9>":
			return KeyComb{Ch: '(', Mod: ModCtrlAlt}, nil
		case "<c-s-a-0>", "<ctrl-shift-alt-0>":
			return KeyComb{Ch: ')', Mod: ModCtrlAlt}, nil
		case "<c-s-a-`>", "<ctrl-shift-alt-`>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '~'}, nil
		case "<c-s-a-a>", "<ctrl-shift-alt-a>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'A'}, nil
		case "<c-s-a-b>", "<ctrl-shift-alt-b>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'B'}, nil
		case "<c-s-a-c>", "<ctrl-shift-alt-c>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'C'}, nil
		case "<c-s-a-d>", "<ctrl-shift-alt-d>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'D'}, nil
		case "<c-s-a-e>", "<ctrl-shift-alt-e>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'E'}, nil
		case "<c-s-a-f>", "<ctrl-shift-alt-f>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'F'}, nil
		case "<c-s-a-g>", "<ctrl-shift-alt-g>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'G'}, nil
		case "<c-s-a-h>", "<ctrl-shift-alt-h>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'H'}, nil
		case "<c-s-a-tab>", "<ctrl-shift-alt-tab>":
			return KeyComb{Mod: ModCtrlShiftAlt, Key: KeyTab}, nil
		case "<c-s-a-i>", "<ctrl-shift-alt-i>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'I'}, nil
		case "<c-s-a-j>", "<ctrl-shift-alt-j>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'J'}, nil
		case "<c-s-a-k>", "<ctrl-shift-alt-k>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'K'}, nil
		case "<c-s-a-l>", "<ctrl-shift-alt-l>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'L'}, nil
		case "<c-s-a-m>", "<ctrl-shift-alt-m>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'M'}, nil
		case "<c-s-a-n>", "<ctrl-shift-alt-n>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'N'}, nil
		case "<c-s-a-o>", "<ctrl-shift-alt-o>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'O'}, nil
		case "<c-s-a-p>", "<ctrl-shift-alt-p>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'P'}, nil
		case "<c-s-a-q>", "<ctrl-shift-alt-q>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'Q'}, nil
		case "<c-s-a-r>", "<ctrl-shift-alt-r>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'R'}, nil
		case "<c-s-a-s>", "<ctrl-shift-alt-s>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'S'}, nil
		case "<c-s-a-t>", "<ctrl-shift-alt-t>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'T'}, nil
		case "<c-s-a-u>", "<ctrl-shift-alt-u>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'U'}, nil
		case "<c-s-a-v>", "<ctrl-shift-alt-v>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'V'}, nil
		case "<c-s-a-w>", "<ctrl-shift-alt-w>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'W'}, nil
		case "<c-s-a-x>", "<ctrl-shift-alt-x>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'X'}, nil
		case "<c-s-a-y>", "<ctrl-shift-alt-y>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'Y'}, nil
		case "<c-s-a-z>", "<ctrl-shift-alt-z>":
			return KeyComb{Mod: ModCtrlAlt, Ch: 'Z'}, nil
		case "<c-s-a-[>", "<ctrl-shift-alt-[>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '{'}, nil
		case "<c-s-a-\\\\>", "<ctrl-shift-alt-\\\\>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '|'}, nil
		case "<c-s-a-]>", "<ctrl-shift-alt-]>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '}'}, nil
		case "<c-s-a-/>", "<ctrl-shift-alt-/>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '?'}, nil
		case "<c-s-a-_>", "<ctrl-shift-alt-_>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '_'}, nil
		case "<c-s-a-.>", "<ctrl-shift-alt-.>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '>'}, nil
		case "<c-s-a-,>", "<ctrl-shift-alt-,>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '<'}, nil
		case "<c-s-a-;>", "<ctrl-shift-alt-;>":
			return KeyComb{Mod: ModCtrlAlt, Ch: ':'}, nil
		case "<c-s-a-'>", "<ctrl-shift-alt-'>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '"'}, nil
		case "<c-s-a-=>", "<ctrl-shift-alt-=>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '+'}, nil
		case "<c-s-a-+>", "<ctrl-shift-alt-+>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '+'}, nil
		case "<c-a-+>", "<ctrl-alt-+>":
			return KeyComb{Mod: ModCtrlAlt, Ch: '+'}, nil
		case "<c-s-a-->", "<ctrl-shift-alt-->":
			return KeyComb{Mod: ModCtrlAlt, Ch: '_'}, nil

		// ctrl+shift+meta
		case "<c-s-m-f1>", "<ctrl-shift-meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f2>", "<ctrl-shift-meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f3>", "<ctrl-shift-meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f4>", "<ctrl-shift-meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f5>", "<ctrl-shift-meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f6>", "<ctrl-shift-meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f7>", "<ctrl-shift-meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f8>", "<ctrl-shift-meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f9>", "<ctrl-shift-meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f10>", "<ctrl-shift-meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f11>", "<ctrl-shift-meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-f12>", "<ctrl-shift-meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-insert>", "<ctrl-shift-meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-delete>", "<ctrl-shift-meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-home>", "<ctrl-shift-meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-end>", "<ctrl-shift-meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-pgup>", "<ctrl-shift-meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-pgdn>", "<ctrl-shift-meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-up>", "<ctrl-shift-meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-down>", "<ctrl-shift-meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-left>", "<ctrl-shift-meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-right>", "<ctrl-shift-meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-mouse-left>", "<ctrl-shift-meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-mouse-middle>", "<ctrl-shift-meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-mouse-right>", "<ctrl-shift-meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-mouse-release>", "<ctrl-shift-meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-mouse-wheel-up>", "<ctrl-shift-meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-mouse-wheel-down>", "<ctrl-shift-meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-enter>", "<ctrl-shift-meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-space>", "<ctrl-shift-meta-space>":
			return KeyComb{Mod: ModCtrlShiftMeta, Key: KeySpace}, nil
		case "<c-s-m-backspace>", "<ctrl-shift-meta-backspace>":
			return KeyComb{Mod: ModCtrlShiftMeta, Key: KeyBackspace}, nil
		case "<c-s-m-esc>", "<ctrl-shift-meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrlShiftMeta}, nil
		case "<c-s-m-1>", "<ctrl-shift-meta-1>":
			return KeyComb{Ch: '!', Mod: ModCtrlMeta}, nil
		case "<c-s-m-2>", "<ctrl-shift-meta-2>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '@'}, nil
		case "<c-s-m-3>", "<ctrl-shift-meta-3>":
			return KeyComb{Ch: '#', Mod: ModCtrlMeta}, nil
		case "<c-s-m-4>", "<ctrl-shift-meta-4>":
			return KeyComb{Ch: '$', Mod: ModCtrlMeta}, nil
		case "<c-s-m-5>", "<ctrl-shift-meta-5>":
			return KeyComb{Ch: '%', Mod: ModCtrlMeta}, nil
		case "<c-s-m-6>", "<ctrl-shift-meta-6>":
			return KeyComb{Ch: '^', Mod: ModCtrlMeta}, nil
		case "<c-s-m-7>", "<ctrl-shift-meta-7>":
			return KeyComb{Ch: '&', Mod: ModCtrlMeta}, nil
		case "<c-s-m-8>", "<ctrl-shift-meta-8>":
			return KeyComb{Ch: '*', Mod: ModCtrlMeta}, nil
		case "<c-s-m-9>", "<ctrl-shift-meta-9>":
			return KeyComb{Ch: '(', Mod: ModCtrlMeta}, nil
		case "<c-s-m-0>", "<ctrl-shift-meta-0>":
			return KeyComb{Ch: ')', Mod: ModCtrlMeta}, nil
		case "<c-s-m-`>", "<ctrl-shift-meta-`>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '~'}, nil
		case "<c-s-m-a>", "<ctrl-shift-meta-a>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'A'}, nil
		case "<c-s-m-b>", "<ctrl-shift-meta-b>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'B'}, nil
		case "<c-s-m-c>", "<ctrl-shift-meta-c>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'C'}, nil
		case "<c-s-m-d>", "<ctrl-shift-meta-d>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'D'}, nil
		case "<c-s-m-e>", "<ctrl-shift-meta-e>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'E'}, nil
		case "<c-s-m-f>", "<ctrl-shift-meta-f>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'F'}, nil
		case "<c-s-m-g>", "<ctrl-shift-meta-g>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'G'}, nil
		case "<c-s-m-h>", "<ctrl-shift-meta-h>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'H'}, nil
		case "<c-s-m-tab>", "<ctrl-shift-meta-tab>":
			return KeyComb{Mod: ModCtrlShiftMeta, Key: KeyTab}, nil
		case "<c-s-m-i>", "<ctrl-shift-meta-i>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'I'}, nil
		case "<c-s-m-j>", "<ctrl-shift-meta-j>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'J'}, nil
		case "<c-s-m-k>", "<ctrl-shift-meta-k>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'K'}, nil
		case "<c-s-m-l>", "<ctrl-shift-meta-l>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'L'}, nil
		case "<c-s-m-m>", "<ctrl-shift-meta-m>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'M'}, nil
		case "<c-s-m-n>", "<ctrl-shift-meta-n>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'N'}, nil
		case "<c-s-m-o>", "<ctrl-shift-meta-o>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'O'}, nil
		case "<c-s-m-p>", "<ctrl-shift-meta-p>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'P'}, nil
		case "<c-s-m-q>", "<ctrl-shift-meta-q>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'Q'}, nil
		case "<c-s-m-r>", "<ctrl-shift-meta-r>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'R'}, nil
		case "<c-s-m-s>", "<ctrl-shift-meta-s>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'S'}, nil
		case "<c-s-m-t>", "<ctrl-shift-meta-t>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'T'}, nil
		case "<c-s-m-u>", "<ctrl-shift-meta-u>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'U'}, nil
		case "<c-s-m-v>", "<ctrl-shift-meta-v>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'V'}, nil
		case "<c-s-m-w>", "<ctrl-shift-meta-w>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'W'}, nil
		case "<c-s-m-x>", "<ctrl-shift-meta-x>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'X'}, nil
		case "<c-s-m-y>", "<ctrl-shift-meta-y>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'Y'}, nil
		case "<c-s-m-z>", "<ctrl-shift-meta-z>":
			return KeyComb{Mod: ModCtrlMeta, Ch: 'Z'}, nil
		case "<c-s-m-[>", "<ctrl-shift-meta-[>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '{'}, nil
		case "<c-s-m-\\\\>", "<ctrl-shift-meta-\\\\>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '|'}, nil
		case "<c-s-m-]>", "<ctrl-shift-meta-]>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '}'}, nil
		case "<c-s-m-/>", "<ctrl-shift-meta-/>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '?'}, nil
		case "<c-s-m-_>", "<ctrl-shift-meta-_>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '_'}, nil
		case "<c-s-m-.>", "<ctrl-shift-meta-.>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '>'}, nil
		case "<c-s-m-,>", "<ctrl-shift-meta-,>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '<'}, nil
		case "<c-s-m-;>", "<ctrl-shift-meta-;>":
			return KeyComb{Mod: ModCtrlMeta, Ch: ':'}, nil
		case "<c-s-m-'>", "<ctrl-shift-meta-'>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '"'}, nil
		case "<c-s-m-=>", "<ctrl-shift-meta-=>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '+'}, nil
		case "<c-s-m-+>", "<ctrl-shift-meta-+>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '+'}, nil
		case "<c-m-+>", "<ctrl-meta-+>":
			return KeyComb{Mod: ModCtrlMeta, Ch: '+'}, nil
		case "<c-s-m-->", "<ctrl-shift-meta-->":
			return KeyComb{Mod: ModCtrlMeta, Ch: '_'}, nil

		// ctrl+alt+meta
		case "<c-a-m-f1>", "<ctrl-alt-meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f2>", "<ctrl-alt-meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f3>", "<ctrl-alt-meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f4>", "<ctrl-alt-meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f5>", "<ctrl-alt-meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f6>", "<ctrl-alt-meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f7>", "<ctrl-alt-meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f8>", "<ctrl-alt-meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f9>", "<ctrl-alt-meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f10>", "<ctrl-alt-meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f11>", "<ctrl-alt-meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-f12>", "<ctrl-alt-meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-insert>", "<ctrl-alt-meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-delete>", "<ctrl-alt-meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-home>", "<ctrl-alt-meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-end>", "<ctrl-alt-meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-pgup>", "<ctrl-alt-meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-pgdn>", "<ctrl-alt-meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-up>", "<ctrl-alt-meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-down>", "<ctrl-alt-meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-left>", "<ctrl-alt-meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-right>", "<ctrl-alt-meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-mouse-left>", "<ctrl-alt-meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-mouse-middle>", "<ctrl-alt-meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-mouse-right>", "<ctrl-alt-meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-mouse-release>", "<ctrl-alt-meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-mouse-wheel-up>", "<ctrl-alt-meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-mouse-wheel-down>", "<ctrl-alt-meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-enter>", "<ctrl-alt-meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-space>", "<ctrl-alt-meta-space>":
			return KeyComb{Mod: ModCtrlAltMeta, Key: KeySpace}, nil
		case "<c-a-m-backspace>", "<ctrl-alt-meta-backspace>":
			return KeyComb{Mod: ModCtrlAltMeta, Key: KeyBackspace}, nil
		case "<c-a-m-esc>", "<ctrl-alt-meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-1>", "<ctrl-alt-meta-1>":
			return KeyComb{Ch: '1', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-2>", "<ctrl-alt-meta-2>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '2'}, nil
		case "<c-a-m-3>", "<ctrl-alt-meta-3>":
			return KeyComb{Ch: '3', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-4>", "<ctrl-alt-meta-4>":
			return KeyComb{Ch: '4', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-5>", "<ctrl-alt-meta-5>":
			return KeyComb{Ch: '5', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-6>", "<ctrl-alt-meta-6>":
			return KeyComb{Ch: '6', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-7>", "<ctrl-alt-meta-7>":
			return KeyComb{Ch: '7', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-8>", "<ctrl-alt-meta-8>":
			return KeyComb{Ch: '8', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-9>", "<ctrl-alt-meta-9>":
			return KeyComb{Ch: '9', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-0>", "<ctrl-alt-meta-0>":
			return KeyComb{Ch: '0', Mod: ModCtrlAltMeta}, nil
		case "<c-a-m-`>", "<ctrl-alt-meta-`>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '`'}, nil
		case "<c-a-m-a>", "<ctrl-alt-meta-a>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'a'}, nil
		case "<c-a-m-b>", "<ctrl-alt-meta-b>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'b'}, nil
		case "<c-a-m-c>", "<ctrl-alt-meta-c>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'c'}, nil
		case "<c-a-m-d>", "<ctrl-alt-meta-d>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'd'}, nil
		case "<c-a-m-e>", "<ctrl-alt-meta-e>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'e'}, nil
		case "<c-a-m-f>", "<ctrl-alt-meta-f>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'f'}, nil
		case "<c-a-m-g>", "<ctrl-alt-meta-g>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'g'}, nil
		case "<c-a-m-h>", "<ctrl-alt-meta-h>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'h'}, nil
		case "<c-a-m-tab>", "<ctrl-alt-meta-tab>":
			return KeyComb{Mod: ModCtrlAltMeta, Key: KeyTab}, nil
		case "<c-a-m-i>", "<ctrl-alt-meta-i>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'i'}, nil
		case "<c-a-m-j>", "<ctrl-alt-meta-j>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'j'}, nil
		case "<c-a-m-k>", "<ctrl-alt-meta-k>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'k'}, nil
		case "<c-a-m-l>", "<ctrl-alt-meta-l>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'l'}, nil
		case "<c-a-m-m>", "<ctrl-alt-meta-m>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'm'}, nil
		case "<c-a-m-n>", "<ctrl-alt-meta-n>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'n'}, nil
		case "<c-a-m-o>", "<ctrl-alt-meta-o>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'o'}, nil
		case "<c-a-m-p>", "<ctrl-alt-meta-p>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'p'}, nil
		case "<c-a-m-q>", "<ctrl-alt-meta-q>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'q'}, nil
		case "<c-a-m-r>", "<ctrl-alt-meta-r>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'r'}, nil
		case "<c-a-m-s>", "<ctrl-alt-meta-s>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 's'}, nil
		case "<c-a-m-t>", "<ctrl-alt-meta-t>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 't'}, nil
		case "<c-a-m-u>", "<ctrl-alt-meta-u>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'u'}, nil
		case "<c-a-m-v>", "<ctrl-alt-meta-v>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'v'}, nil
		case "<c-a-m-w>", "<ctrl-alt-meta-w>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'w'}, nil
		case "<c-a-m-x>", "<ctrl-alt-meta-x>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'x'}, nil
		case "<c-a-m-y>", "<ctrl-alt-meta-y>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'y'}, nil
		case "<c-a-m-z>", "<ctrl-alt-meta-z>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: 'z'}, nil
		case "<c-a-m-[>", "<ctrl-alt-meta-[>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '['}, nil
		case "<c-a-m-\\\\>", "<ctrl-alt-meta-\\\\>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '\\'}, nil
		case "<c-a-m-]>", "<ctrl-alt-meta-]>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: ']'}, nil
		case "<c-a-m-/>", "<ctrl-alt-meta-/>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '/'}, nil
		case "<c-a-m-_>", "<ctrl-alt-meta-_>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '_'}, nil
		case "<c-a-m-.>", "<ctrl-alt-meta-.>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '.'}, nil
		case "<c-a-m-,>", "<ctrl-alt-meta-,>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: ','}, nil
		case "<c-a-m-;>", "<ctrl-alt-meta-;>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: ';'}, nil
		case "<c-a-m-'>", "<ctrl-alt-meta-'>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '\''}, nil
		case "<c-a-m-=>", "<ctrl-alt-meta-=>":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '='}, nil
		case "<c-a-m-->", "<ctrl-alt-meta-->":
			return KeyComb{Mod: ModCtrlAltMeta, Ch: '-'}, nil

		// shift+meta
		case "<s-m-f1>", "<shift-meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModShiftMeta}, nil
		case "<s-m-f2>", "<shift-meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModShiftMeta}, nil
		case "<s-m-f3>", "<shift-meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModShiftMeta}, nil
		case "<s-m-f4>", "<shift-meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModShiftMeta}, nil
		case "<s-m-f5>", "<shift-meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModShiftMeta}, nil
		case "<s-m-f6>", "<shift-meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModShiftMeta}, nil
		case "<s-m-f7>", "<shift-meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModShiftMeta}, nil
		case "<s-m-f8>", "<shift-meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModShiftMeta}, nil
		case "<s-m-f9>", "<shift-meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModShiftMeta}, nil
		case "<s-m-f10>", "<shift-meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModShiftMeta}, nil
		case "<s-m-f11>", "<shift-meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModShiftMeta}, nil
		case "<s-m-f12>", "<shift-meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModShiftMeta}, nil
		case "<s-m-insert>", "<shift-meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModShiftMeta}, nil
		case "<s-m-delete>", "<shift-meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModShiftMeta}, nil
		case "<s-m-home>", "<shift-meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModShiftMeta}, nil
		case "<s-m-end>", "<shift-meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModShiftMeta}, nil
		case "<s-m-pgup>", "<shift-meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModShiftMeta}, nil
		case "<s-m-pgdn>", "<shift-meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModShiftMeta}, nil
		case "<s-m-up>", "<shift-meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModShiftMeta}, nil
		case "<s-m-down>", "<shift-meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModShiftMeta}, nil
		case "<s-m-left>", "<shift-meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModShiftMeta}, nil
		case "<s-m-right>", "<shift-meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModShiftMeta}, nil
		case "<s-m-mouse-left>", "<shift-meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModShiftMeta}, nil
		case "<s-m-mouse-middle>", "<shift-meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModShiftMeta}, nil
		case "<s-m-mouse-right>", "<shift-meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModShiftMeta}, nil
		case "<s-m-mouse-release>", "<shift-meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModShiftMeta}, nil
		case "<s-m-mouse-wheel-up>", "<shift-meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModShiftMeta}, nil
		case "<s-m-mouse-wheel-down>", "<shift-meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModShiftMeta}, nil
		case "<s-m-enter>", "<shift-meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModShiftMeta}, nil
		case "<s-m-space>", "<shift-meta-space>":
			return KeyComb{Mod: ModShiftMeta, Key: KeySpace}, nil
		case "<s-m-backspace>", "<shift-meta-backspace>":
			return KeyComb{Mod: ModShiftMeta, Key: KeyBackspace}, nil
		case "<s-m-esc>", "<shift-meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModShiftMeta}, nil
		case "<s-m-1>", "<shift-meta-1>":
			return KeyComb{Ch: '!', Mod: ModMeta}, nil
		case "<s-m-2>", "<shift-meta-2>":
			return KeyComb{Mod: ModMeta, Ch: '@'}, nil
		case "<s-m-3>", "<shift-meta-3>":
			return KeyComb{Ch: '#', Mod: ModMeta}, nil
		case "<s-m-4>", "<shift-meta-4>":
			return KeyComb{Ch: '$', Mod: ModMeta}, nil
		case "<s-m-5>", "<shift-meta-5>":
			return KeyComb{Ch: '%', Mod: ModMeta}, nil
		case "<s-m-6>", "<shift-meta-6>":
			return KeyComb{Ch: '^', Mod: ModMeta}, nil
		case "<s-m-7>", "<shift-meta-7>":
			return KeyComb{Ch: '&', Mod: ModMeta}, nil
		case "<s-m-8>", "<shift-meta-8>":
			return KeyComb{Ch: '*', Mod: ModMeta}, nil
		case "<s-m-9>", "<shift-meta-9>":
			return KeyComb{Ch: '(', Mod: ModMeta}, nil
		case "<s-m-0>", "<shift-meta-0>":
			return KeyComb{Ch: ')', Mod: ModMeta}, nil
		case "<s-m-`>", "<shift-meta-`>":
			return KeyComb{Mod: ModMeta, Ch: '~'}, nil
		case "<s-m-a>", "<shift-meta-a>":
			return KeyComb{Mod: ModMeta, Ch: 'A'}, nil
		case "<s-m-b>", "<shift-meta-b>":
			return KeyComb{Mod: ModMeta, Ch: 'B'}, nil
		case "<s-m-c>", "<shift-meta-c>":
			return KeyComb{Mod: ModMeta, Ch: 'C'}, nil
		case "<s-m-d>", "<shift-meta-d>":
			return KeyComb{Mod: ModMeta, Ch: 'D'}, nil
		case "<s-m-e>", "<shift-meta-e>":
			return KeyComb{Mod: ModMeta, Ch: 'E'}, nil
		case "<s-m-f>", "<shift-meta-f>":
			return KeyComb{Mod: ModMeta, Ch: 'F'}, nil
		case "<s-m-g>", "<shift-meta-g>":
			return KeyComb{Mod: ModMeta, Ch: 'G'}, nil
		case "<s-m-h>", "<shift-meta-h>":
			return KeyComb{Mod: ModMeta, Ch: 'H'}, nil
		case "<s-m-tab>", "<shift-meta-tab>":
			return KeyComb{Mod: ModShiftMeta, Key: KeyTab}, nil
		case "<s-m-i>", "<shift-meta-i>":
			return KeyComb{Mod: ModMeta, Ch: 'I'}, nil
		case "<s-m-j>", "<shift-meta-j>":
			return KeyComb{Mod: ModMeta, Ch: 'J'}, nil
		case "<s-m-k>", "<shift-meta-k>":
			return KeyComb{Mod: ModMeta, Ch: 'K'}, nil
		case "<s-m-l>", "<shift-meta-l>":
			return KeyComb{Mod: ModMeta, Ch: 'L'}, nil
		case "<s-m-m>", "<shift-meta-m>":
			return KeyComb{Mod: ModMeta, Ch: 'M'}, nil
		case "<s-m-n>", "<shift-meta-n>":
			return KeyComb{Mod: ModMeta, Ch: 'N'}, nil
		case "<s-m-o>", "<shift-meta-o>":
			return KeyComb{Mod: ModMeta, Ch: 'O'}, nil
		case "<s-m-p>", "<shift-meta-p>":
			return KeyComb{Mod: ModMeta, Ch: 'P'}, nil
		case "<s-m-q>", "<shift-meta-q>":
			return KeyComb{Mod: ModMeta, Ch: 'Q'}, nil
		case "<s-m-r>", "<shift-meta-r>":
			return KeyComb{Mod: ModMeta, Ch: 'R'}, nil
		case "<s-m-s>", "<shift-meta-s>":
			return KeyComb{Mod: ModMeta, Ch: 'S'}, nil
		case "<s-m-t>", "<shift-meta-t>":
			return KeyComb{Mod: ModMeta, Ch: 'T'}, nil
		case "<s-m-u>", "<shift-meta-u>":
			return KeyComb{Mod: ModMeta, Ch: 'U'}, nil
		case "<s-m-v>", "<shift-meta-v>":
			return KeyComb{Mod: ModMeta, Ch: 'V'}, nil
		case "<s-m-w>", "<shift-meta-w>":
			return KeyComb{Mod: ModMeta, Ch: 'W'}, nil
		case "<s-m-x>", "<shift-meta-x>":
			return KeyComb{Mod: ModMeta, Ch: 'X'}, nil
		case "<s-m-y>", "<shift-meta-y>":
			return KeyComb{Mod: ModMeta, Ch: 'Y'}, nil
		case "<s-m-z>", "<shift-meta-z>":
			return KeyComb{Mod: ModMeta, Ch: 'Z'}, nil
		case "<s-m-[>", "<shift-meta-[>":
			return KeyComb{Mod: ModMeta, Ch: '{'}, nil
		case "<s-m-\\\\>", "<shift-meta-\\\\>":
			return KeyComb{Mod: ModMeta, Ch: '|'}, nil
		case "<s-m-]>", "<shift-meta-]>":
			return KeyComb{Mod: ModMeta, Ch: '}'}, nil
		case "<s-m-/>", "<shift-meta-/>":
			return KeyComb{Mod: ModMeta, Ch: '?'}, nil
		case "<s-m-_>", "<shift-meta-_>":
			return KeyComb{Mod: ModMeta, Ch: '_'}, nil
		case "<s-m-.>", "<shift-meta-.>":
			return KeyComb{Mod: ModMeta, Ch: '>'}, nil
		case "<s-m-,>", "<shift-meta-,>":
			return KeyComb{Mod: ModMeta, Ch: '<'}, nil
		case "<s-m-;>", "<shift-meta-;>":
			return KeyComb{Mod: ModMeta, Ch: ':'}, nil
		case "<s-m-'>", "<shift-meta-'>":
			return KeyComb{Mod: ModMeta, Ch: '"'}, nil
		case "<s-m-=>", "<shift-meta-=>":
			return KeyComb{Mod: ModMeta, Ch: '+'}, nil
		case "<s-m-+>", "<shift-meta-+>":
			return KeyComb{Mod: ModMeta, Ch: '+'}, nil
		case "<m-+>", "<meta-+>":
			return KeyComb{Mod: ModMeta, Ch: '+'}, nil
		case "<s-m-->", "<shift-meta-->":
			return KeyComb{Mod: ModMeta, Ch: '_'}, nil

		// alt+meta
		case "<a-m-f1>", "<alt-meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModAltMeta}, nil
		case "<a-m-f2>", "<alt-meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModAltMeta}, nil
		case "<a-m-f3>", "<alt-meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModAltMeta}, nil
		case "<a-m-f4>", "<alt-meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModAltMeta}, nil
		case "<a-m-f5>", "<alt-meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModAltMeta}, nil
		case "<a-m-f6>", "<alt-meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModAltMeta}, nil
		case "<a-m-f7>", "<alt-meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModAltMeta}, nil
		case "<a-m-f8>", "<alt-meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModAltMeta}, nil
		case "<a-m-f9>", "<alt-meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModAltMeta}, nil
		case "<a-m-f10>", "<alt-meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModAltMeta}, nil
		case "<a-m-f11>", "<alt-meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModAltMeta}, nil
		case "<a-m-f12>", "<alt-meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModAltMeta}, nil
		case "<a-m-insert>", "<alt-meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModAltMeta}, nil
		case "<a-m-delete>", "<alt-meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModAltMeta}, nil
		case "<a-m-home>", "<alt-meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModAltMeta}, nil
		case "<a-m-end>", "<alt-meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModAltMeta}, nil
		case "<a-m-pgup>", "<alt-meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModAltMeta}, nil
		case "<a-m-pgdn>", "<alt-meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModAltMeta}, nil
		case "<a-m-up>", "<alt-meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModAltMeta}, nil
		case "<a-m-down>", "<alt-meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModAltMeta}, nil
		case "<a-m-left>", "<alt-meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModAltMeta}, nil
		case "<a-m-right>", "<alt-meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModAltMeta}, nil
		case "<a-m-mouse-left>", "<alt-meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModAltMeta}, nil
		case "<a-m-mouse-middle>", "<alt-meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModAltMeta}, nil
		case "<a-m-mouse-right>", "<alt-meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModAltMeta}, nil
		case "<a-m-mouse-release>", "<alt-meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModAltMeta}, nil
		case "<a-m-mouse-wheel-up>", "<alt-meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModAltMeta}, nil
		case "<a-m-mouse-wheel-down>", "<alt-meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModAltMeta}, nil
		case "<a-m-enter>", "<alt-meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModAltMeta}, nil
		case "<a-m-space>", "<alt-meta-space>":
			return KeyComb{Mod: ModAltMeta, Key: KeySpace}, nil
		case "<a-m-backspace>", "<alt-meta-backspace>":
			return KeyComb{Mod: ModAltMeta, Key: KeyBackspace}, nil
		case "<a-m-esc>", "<alt-meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModAltMeta}, nil
		case "<a-m-1>", "<alt-meta-1>":
			return KeyComb{Ch: '1', Mod: ModAltMeta}, nil
		case "<a-m-2>", "<alt-meta-2>":
			return KeyComb{Mod: ModAltMeta, Ch: '2'}, nil
		case "<a-m-3>", "<alt-meta-3>":
			return KeyComb{Ch: '3', Mod: ModAltMeta}, nil
		case "<a-m-4>", "<alt-meta-4>":
			return KeyComb{Ch: '4', Mod: ModAltMeta}, nil
		case "<a-m-5>", "<alt-meta-5>":
			return KeyComb{Ch: '5', Mod: ModAltMeta}, nil
		case "<a-m-6>", "<alt-meta-6>":
			return KeyComb{Ch: '6', Mod: ModAltMeta}, nil
		case "<a-m-7>", "<alt-meta-7>":
			return KeyComb{Ch: '7', Mod: ModAltMeta}, nil
		case "<a-m-8>", "<alt-meta-8>":
			return KeyComb{Ch: '8', Mod: ModAltMeta}, nil
		case "<a-m-9>", "<alt-meta-9>":
			return KeyComb{Ch: '9', Mod: ModAltMeta}, nil
		case "<a-m-0>", "<alt-meta-0>":
			return KeyComb{Ch: '0', Mod: ModAltMeta}, nil
		case "<a-m-`>", "<alt-meta-`>":
			return KeyComb{Mod: ModAltMeta, Ch: '`'}, nil
		case "<a-m-a>", "<alt-meta-a>":
			return KeyComb{Mod: ModAltMeta, Ch: 'a'}, nil
		case "<a-m-b>", "<alt-meta-b>":
			return KeyComb{Mod: ModAltMeta, Ch: 'b'}, nil
		case "<a-m-c>", "<alt-meta-c>":
			return KeyComb{Mod: ModAltMeta, Ch: 'c'}, nil
		case "<a-m-d>", "<alt-meta-d>":
			return KeyComb{Mod: ModAltMeta, Ch: 'd'}, nil
		case "<a-m-e>", "<alt-meta-e>":
			return KeyComb{Mod: ModAltMeta, Ch: 'e'}, nil
		case "<a-m-f>", "<alt-meta-f>":
			return KeyComb{Mod: ModAltMeta, Ch: 'f'}, nil
		case "<a-m-g>", "<alt-meta-g>":
			return KeyComb{Mod: ModAltMeta, Ch: 'g'}, nil
		case "<a-m-h>", "<alt-meta-h>":
			return KeyComb{Mod: ModAltMeta, Ch: 'h'}, nil
		case "<a-m-tab>", "<alt-meta-tab>":
			return KeyComb{Mod: ModAltMeta, Key: KeyTab}, nil
		case "<a-m-i>", "<alt-meta-i>":
			return KeyComb{Mod: ModAltMeta, Ch: 'i'}, nil
		case "<a-m-j>", "<alt-meta-j>":
			return KeyComb{Mod: ModAltMeta, Ch: 'j'}, nil
		case "<a-m-k>", "<alt-meta-k>":
			return KeyComb{Mod: ModAltMeta, Ch: 'k'}, nil
		case "<a-m-l>", "<alt-meta-l>":
			return KeyComb{Mod: ModAltMeta, Ch: 'l'}, nil
		case "<a-m-m>", "<alt-meta-m>":
			return KeyComb{Mod: ModAltMeta, Ch: 'm'}, nil
		case "<a-m-n>", "<alt-meta-n>":
			return KeyComb{Mod: ModAltMeta, Ch: 'n'}, nil
		case "<a-m-o>", "<alt-meta-o>":
			return KeyComb{Mod: ModAltMeta, Ch: 'o'}, nil
		case "<a-m-p>", "<alt-meta-p>":
			return KeyComb{Mod: ModAltMeta, Ch: 'p'}, nil
		case "<a-m-q>", "<alt-meta-q>":
			return KeyComb{Mod: ModAltMeta, Ch: 'q'}, nil
		case "<a-m-r>", "<alt-meta-r>":
			return KeyComb{Mod: ModAltMeta, Ch: 'r'}, nil
		case "<a-m-s>", "<alt-meta-s>":
			return KeyComb{Mod: ModAltMeta, Ch: 's'}, nil
		case "<a-m-t>", "<alt-meta-t>":
			return KeyComb{Mod: ModAltMeta, Ch: 't'}, nil
		case "<a-m-u>", "<alt-meta-u>":
			return KeyComb{Mod: ModAltMeta, Ch: 'u'}, nil
		case "<a-m-v>", "<alt-meta-v>":
			return KeyComb{Mod: ModAltMeta, Ch: 'v'}, nil
		case "<a-m-w>", "<alt-meta-w>":
			return KeyComb{Mod: ModAltMeta, Ch: 'w'}, nil
		case "<a-m-x>", "<alt-meta-x>":
			return KeyComb{Mod: ModAltMeta, Ch: 'x'}, nil
		case "<a-m-y>", "<alt-meta-y>":
			return KeyComb{Mod: ModAltMeta, Ch: 'y'}, nil
		case "<a-m-z>", "<alt-meta-z>":
			return KeyComb{Mod: ModAltMeta, Ch: 'z'}, nil
		case "<a-m-[>", "<alt-meta-[>":
			return KeyComb{Mod: ModAltMeta, Ch: '['}, nil
		case "<a-m-\\\\>", "<alt-meta-\\\\>":
			return KeyComb{Mod: ModAltMeta, Ch: '\\'}, nil
		case "<a-m-]>", "<alt-meta-]>":
			return KeyComb{Mod: ModAltMeta, Ch: ']'}, nil
		case "<a-m-/>", "<alt-meta-/>":
			return KeyComb{Mod: ModAltMeta, Ch: '/'}, nil
		case "<a-m-_>", "<alt-meta-_>":
			return KeyComb{Mod: ModAltMeta, Ch: '_'}, nil
		case "<a-m-.>", "<alt-meta-.>":
			return KeyComb{Mod: ModAltMeta, Ch: '.'}, nil
		case "<a-m-,>", "<alt-meta-,>":
			return KeyComb{Mod: ModAltMeta, Ch: ','}, nil
		case "<a-m-;>", "<alt-meta-;>":
			return KeyComb{Mod: ModAltMeta, Ch: ';'}, nil
		case "<a-m-'>", "<alt-meta-'>":
			return KeyComb{Mod: ModAltMeta, Ch: '\''}, nil
		case "<a-m-=>", "<alt-meta-=>":
			return KeyComb{Mod: ModAltMeta, Ch: '='}, nil
		case "<a-m-->", "<alt-meta-->":
			return KeyComb{Mod: ModAltMeta, Ch: '-'}, nil

		// alt+shift+meta
		case "<a-s-m-f1>", "<alt-shift-meta-f1>":
			return KeyComb{Key: KeyF1, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f2>", "<alt-shift-meta-f2>":
			return KeyComb{Key: KeyF2, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f3>", "<alt-shift-meta-f3>":
			return KeyComb{Key: KeyF3, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f4>", "<alt-shift-meta-f4>":
			return KeyComb{Key: KeyF4, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f5>", "<alt-shift-meta-f5>":
			return KeyComb{Key: KeyF5, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f6>", "<alt-shift-meta-f6>":
			return KeyComb{Key: KeyF6, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f7>", "<alt-shift-meta-f7>":
			return KeyComb{Key: KeyF7, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f8>", "<alt-shift-meta-f8>":
			return KeyComb{Key: KeyF8, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f9>", "<alt-shift-meta-f9>":
			return KeyComb{Key: KeyF9, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f10>", "<alt-shift-meta-f10>":
			return KeyComb{Key: KeyF10, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f11>", "<alt-shift-meta-f11>":
			return KeyComb{Key: KeyF11, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-f12>", "<alt-shift-meta-f12>":
			return KeyComb{Key: KeyF12, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-insert>", "<alt-shift-meta-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-delete>", "<alt-shift-meta-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-home>", "<alt-shift-meta-home>":
			return KeyComb{Key: KeyHome, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-end>", "<alt-shift-meta-end>":
			return KeyComb{Key: KeyEnd, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-pgup>", "<alt-shift-meta-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-pgdn>", "<alt-shift-meta-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-up>", "<alt-shift-meta-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-down>", "<alt-shift-meta-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-left>", "<alt-shift-meta-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-right>", "<alt-shift-meta-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-mouse-left>", "<alt-shift-meta-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-mouse-middle>", "<alt-shift-meta-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-mouse-right>", "<alt-shift-meta-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-mouse-release>", "<alt-shift-meta-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-mouse-wheel-up>", "<alt-shift-meta-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-mouse-wheel-down>", "<alt-shift-meta-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-enter>", "<alt-shift-meta-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-space>", "<alt-shift-meta-space>":
			return KeyComb{Mod: ModAltShiftMeta, Key: KeySpace}, nil
		case "<a-s-m-backspace>", "<alt-shift-meta-backspace>":
			return KeyComb{Mod: ModAltShiftMeta, Key: KeyBackspace}, nil
		case "<a-s-m-esc>", "<alt-shift-meta-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModAltShiftMeta}, nil
		case "<a-s-m-1>", "<alt-shift-meta-1>":
			return KeyComb{Ch: '!', Mod: ModAltMeta}, nil
		case "<a-s-m-2>", "<alt-shift-meta-2>":
			return KeyComb{Mod: ModAltMeta, Ch: '@'}, nil
		case "<a-s-m-3>", "<alt-shift-meta-3>":
			return KeyComb{Ch: '#', Mod: ModAltMeta}, nil
		case "<a-s-m-4>", "<alt-shift-meta-4>":
			return KeyComb{Ch: '$', Mod: ModAltMeta}, nil
		case "<a-s-m-5>", "<alt-shift-meta-5>":
			return KeyComb{Ch: '%', Mod: ModAltMeta}, nil
		case "<a-s-m-6>", "<alt-shift-meta-6>":
			return KeyComb{Ch: '^', Mod: ModAltMeta}, nil
		case "<a-s-m-7>", "<alt-shift-meta-7>":
			return KeyComb{Ch: '&', Mod: ModAltMeta}, nil
		case "<a-s-m-8>", "<alt-shift-meta-8>":
			return KeyComb{Ch: '*', Mod: ModAltMeta}, nil
		case "<a-s-m-9>", "<alt-shift-meta-9>":
			return KeyComb{Ch: '(', Mod: ModAltMeta}, nil
		case "<a-s-m-0>", "<alt-shift-meta-0>":
			return KeyComb{Ch: ')', Mod: ModAltMeta}, nil
		case "<a-s-m-`>", "<alt-shift-meta-`>":
			return KeyComb{Mod: ModAltMeta, Ch: '~'}, nil
		case "<a-s-m-a>", "<alt-shift-meta-a>":
			return KeyComb{Mod: ModAltMeta, Ch: 'A'}, nil
		case "<a-s-m-b>", "<alt-shift-meta-b>":
			return KeyComb{Mod: ModAltMeta, Ch: 'B'}, nil
		case "<a-s-m-c>", "<alt-shift-meta-c>":
			return KeyComb{Mod: ModAltMeta, Ch: 'C'}, nil
		case "<a-s-m-d>", "<alt-shift-meta-d>":
			return KeyComb{Mod: ModAltMeta, Ch: 'D'}, nil
		case "<a-s-m-e>", "<alt-shift-meta-e>":
			return KeyComb{Mod: ModAltMeta, Ch: 'E'}, nil
		case "<a-s-m-f>", "<alt-shift-meta-f>":
			return KeyComb{Mod: ModAltMeta, Ch: 'F'}, nil
		case "<a-s-m-g>", "<alt-shift-meta-g>":
			return KeyComb{Mod: ModAltMeta, Ch: 'G'}, nil
		case "<a-s-m-h>", "<alt-shift-meta-h>":
			return KeyComb{Mod: ModAltMeta, Ch: 'H'}, nil
		case "<a-s-m-tab>", "<alt-shift-meta-tab>":
			return KeyComb{Mod: ModAltShiftMeta, Key: KeyTab}, nil
		case "<a-s-m-i>", "<alt-shift-meta-i>":
			return KeyComb{Mod: ModAltMeta, Ch: 'I'}, nil
		case "<a-s-m-j>", "<alt-shift-meta-j>":
			return KeyComb{Mod: ModAltMeta, Ch: 'J'}, nil
		case "<a-s-m-k>", "<alt-shift-meta-k>":
			return KeyComb{Mod: ModAltMeta, Ch: 'K'}, nil
		case "<a-s-m-l>", "<alt-shift-meta-l>":
			return KeyComb{Mod: ModAltMeta, Ch: 'L'}, nil
		case "<a-s-m-m>", "<alt-shift-meta-m>":
			return KeyComb{Mod: ModAltMeta, Ch: 'M'}, nil
		case "<a-s-m-n>", "<alt-shift-meta-n>":
			return KeyComb{Mod: ModAltMeta, Ch: 'N'}, nil
		case "<a-s-m-o>", "<alt-shift-meta-o>":
			return KeyComb{Mod: ModAltMeta, Ch: 'O'}, nil
		case "<a-s-m-p>", "<alt-shift-meta-p>":
			return KeyComb{Mod: ModAltMeta, Ch: 'P'}, nil
		case "<a-s-m-q>", "<alt-shift-meta-q>":
			return KeyComb{Mod: ModAltMeta, Ch: 'Q'}, nil
		case "<a-s-m-r>", "<alt-shift-meta-r>":
			return KeyComb{Mod: ModAltMeta, Ch: 'R'}, nil
		case "<a-s-m-s>", "<alt-shift-meta-s>":
			return KeyComb{Mod: ModAltMeta, Ch: 'S'}, nil
		case "<a-s-m-t>", "<alt-shift-meta-t>":
			return KeyComb{Mod: ModAltMeta, Ch: 'T'}, nil
		case "<a-s-m-u>", "<alt-shift-meta-u>":
			return KeyComb{Mod: ModAltMeta, Ch: 'U'}, nil
		case "<a-s-m-v>", "<alt-shift-meta-v>":
			return KeyComb{Mod: ModAltMeta, Ch: 'V'}, nil
		case "<a-s-m-w>", "<alt-shift-meta-w>":
			return KeyComb{Mod: ModAltMeta, Ch: 'W'}, nil
		case "<a-s-m-x>", "<alt-shift-meta-x>":
			return KeyComb{Mod: ModAltMeta, Ch: 'X'}, nil
		case "<a-s-m-y>", "<alt-shift-meta-y>":
			return KeyComb{Mod: ModAltMeta, Ch: 'Y'}, nil
		case "<a-s-m-z>", "<alt-shift-meta-z>":
			return KeyComb{Mod: ModAltMeta, Ch: 'Z'}, nil
		case "<a-s-m-[>", "<alt-shift-meta-[>":
			return KeyComb{Mod: ModAltMeta, Ch: '{'}, nil
		case "<a-s-m-\\\\>", "<alt-shift-meta-\\\\>":
			return KeyComb{Mod: ModAltMeta, Ch: '|'}, nil
		case "<a-s-m-]>", "<alt-shift-meta-]>":
			return KeyComb{Mod: ModAltMeta, Ch: '}'}, nil
		case "<a-s-m-/>", "<alt-shift-meta-/>":
			return KeyComb{Mod: ModAltMeta, Ch: '?'}, nil
		case "<a-s-m-_>", "<alt-shift-meta-_>":
			return KeyComb{Mod: ModAltMeta, Ch: '_'}, nil
		case "<a-s-m-.>", "<alt-shift-meta-.>":
			return KeyComb{Mod: ModAltMeta, Ch: '>'}, nil
		case "<a-s-m-,>", "<alt-shift-meta-,>":
			return KeyComb{Mod: ModAltMeta, Ch: '<'}, nil
		case "<a-s-m-;>", "<alt-shift-meta-;>":
			return KeyComb{Mod: ModAltMeta, Ch: ':'}, nil
		case "<a-s-m-'>", "<alt-shift-meta-'>":
			return KeyComb{Mod: ModAltMeta, Ch: '"'}, nil
		case "<a-s-m-=>", "<alt-shift-meta-=>":
			return KeyComb{Mod: ModAltMeta, Ch: '+'}, nil
		case "<a-s-m-+>", "<alt-shift-meta-+>":
			return KeyComb{Mod: ModAltMeta, Ch: '+'}, nil
		case "<a-m-+>", "<alt-meta-+>":
			return KeyComb{Mod: ModAltMeta, Ch: '+'}, nil
		case "<a-s-m-->", "<alt-shift-meta-->":
			return KeyComb{Mod: ModAltMeta, Ch: '_'}, nil

		// alt+shift
		case "<a-s-f1>", "<alt-shift-f1>":
			return KeyComb{Key: KeyF1, Mod: ModAltShift}, nil
		case "<a-s-f2>", "<alt-shift-f2>":
			return KeyComb{Key: KeyF2, Mod: ModAltShift}, nil
		case "<a-s-f3>", "<alt-shift-f3>":
			return KeyComb{Key: KeyF3, Mod: ModAltShift}, nil
		case "<a-s-f4>", "<alt-shift-f4>":
			return KeyComb{Key: KeyF4, Mod: ModAltShift}, nil
		case "<a-s-f5>", "<alt-shift-f5>":
			return KeyComb{Key: KeyF5, Mod: ModAltShift}, nil
		case "<a-s-f6>", "<alt-shift-f6>":
			return KeyComb{Key: KeyF6, Mod: ModAltShift}, nil
		case "<a-s-f7>", "<alt-shift-f7>":
			return KeyComb{Key: KeyF7, Mod: ModAltShift}, nil
		case "<a-s-f8>", "<alt-shift-f8>":
			return KeyComb{Key: KeyF8, Mod: ModAltShift}, nil
		case "<a-s-f9>", "<alt-shift-f9>":
			return KeyComb{Key: KeyF9, Mod: ModAltShift}, nil
		case "<a-s-f10>", "<alt-shift-f10>":
			return KeyComb{Key: KeyF10, Mod: ModAltShift}, nil
		case "<a-s-f11>", "<alt-shift-f11>":
			return KeyComb{Key: KeyF11, Mod: ModAltShift}, nil
		case "<a-s-f12>", "<alt-shift-f12>":
			return KeyComb{Key: KeyF12, Mod: ModAltShift}, nil
		case "<a-s-insert>", "<alt-shift-insert>":
			return KeyComb{Key: KeyInsert, Mod: ModAltShift}, nil
		case "<a-s-delete>", "<alt-shift-delete>":
			return KeyComb{Key: KeyDelete, Mod: ModAltShift}, nil
		case "<a-s-home>", "<alt-shift-home>":
			return KeyComb{Key: KeyHome, Mod: ModAltShift}, nil
		case "<a-s-end>", "<alt-shift-end>":
			return KeyComb{Key: KeyEnd, Mod: ModAltShift}, nil
		case "<a-s-pgup>", "<alt-shift-pgup>":
			return KeyComb{Key: KeyPgup, Mod: ModAltShift}, nil
		case "<a-s-pgdn>", "<alt-shift-pgdn>":
			return KeyComb{Key: KeyPgdn, Mod: ModAltShift}, nil
		case "<a-s-up>", "<alt-shift-up>":
			return KeyComb{Key: KeyArrowUp, Mod: ModAltShift}, nil
		case "<a-s-down>", "<alt-shift-down>":
			return KeyComb{Key: KeyArrowDown, Mod: ModAltShift}, nil
		case "<a-s-left>", "<alt-shift-left>":
			return KeyComb{Key: KeyArrowLeft, Mod: ModAltShift}, nil
		case "<a-s-right>", "<alt-shift-right>":
			return KeyComb{Key: KeyArrowRight, Mod: ModAltShift}, nil
		case "<a-s-mouse-left>", "<alt-shift-mouse-left>":
			return KeyComb{Key: MouseLeft, Mod: ModAltShift}, nil
		case "<a-s-mouse-middle>", "<alt-shift-mouse-middle>":
			return KeyComb{Key: MouseMiddle, Mod: ModAltShift}, nil
		case "<a-s-mouse-right>", "<alt-shift-mouse-right>":
			return KeyComb{Key: MouseRight, Mod: ModAltShift}, nil
		case "<a-s-mouse-release>", "<alt-shift-mouse-release>":
			return KeyComb{Key: MouseRelease, Mod: ModAltShift}, nil
		case "<a-s-mouse-wheel-up>", "<alt-shift-mouse-wheel-up>":
			return KeyComb{Key: MouseWheelUp, Mod: ModAltShift}, nil
		case "<a-s-mouse-wheel-down>", "<alt-shift-mouse-wheel-down>":
			return KeyComb{Key: MouseWheelDown, Mod: ModAltShift}, nil
		case "<a-s-enter>", "<alt-shift-enter>":
			return KeyComb{Key: KeyEnter, Mod: ModAltShift}, nil
		case "<a-s-space>", "<alt-shift-space>":
			return KeyComb{Mod: ModAltShift, Key: KeySpace}, nil
		case "<a-s-backspace>", "<alt-shift-backspace>":
			return KeyComb{Mod: ModAltShift, Key: KeyBackspace}, nil
		case "<a-s-esc>", "<alt-shift-esc>":
			return KeyComb{Key: KeyEsc, Mod: ModAltShift}, nil
		case "<a-s-1>", "<alt-shift-1>":
			return KeyComb{Ch: '!', Mod: ModAlt}, nil
		case "<a-s-2>", "<alt-shift-2>":
			return KeyComb{Mod: ModAlt, Ch: '@'}, nil
		case "<a-s-3>", "<alt-shift-3>":
			return KeyComb{Ch: '#', Mod: ModAlt}, nil
		case "<a-s-4>", "<alt-shift-4>":
			return KeyComb{Ch: '$', Mod: ModAlt}, nil
		case "<a-s-5>", "<alt-shift-5>":
			return KeyComb{Ch: '%', Mod: ModAlt}, nil
		case "<a-s-6>", "<alt-shift-6>":
			return KeyComb{Ch: '^', Mod: ModAlt}, nil
		case "<a-s-7>", "<alt-shift-7>":
			return KeyComb{Ch: '&', Mod: ModAlt}, nil
		case "<a-s-8>", "<alt-shift-8>":
			return KeyComb{Ch: '*', Mod: ModAlt}, nil
		case "<a-s-9>", "<alt-shift-9>":
			return KeyComb{Ch: '(', Mod: ModAlt}, nil
		case "<a-s-0>", "<alt-shift-0>":
			return KeyComb{Ch: ')', Mod: ModAlt}, nil
		case "<a-s-`>", "<alt-shift-`>":
			return KeyComb{Mod: ModAlt, Ch: '~'}, nil
		case "<a-s-a>", "<alt-shift-a>":
			return KeyComb{Mod: ModAlt, Ch: 'A'}, nil
		case "<a-s-b>", "<alt-shift-b>":
			return KeyComb{Mod: ModAlt, Ch: 'B'}, nil
		case "<a-s-c>", "<alt-shift-c>":
			return KeyComb{Mod: ModAlt, Ch: 'C'}, nil
		case "<a-s-d>", "<alt-shift-d>":
			return KeyComb{Mod: ModAlt, Ch: 'D'}, nil
		case "<a-s-e>", "<alt-shift-e>":
			return KeyComb{Mod: ModAlt, Ch: 'E'}, nil
		case "<a-s-f>", "<alt-shift-f>":
			return KeyComb{Mod: ModAlt, Ch: 'F'}, nil
		case "<a-s-g>", "<alt-shift-g>":
			return KeyComb{Mod: ModAlt, Ch: 'G'}, nil
		case "<a-s-h>", "<alt-shift-h>":
			return KeyComb{Mod: ModAlt, Ch: 'H'}, nil
		case "<a-s-tab>", "<alt-shift-tab>":
			return KeyComb{Mod: ModAltShift, Key: KeyTab}, nil
		case "<a-s-i>", "<alt-shift-i>":
			return KeyComb{Mod: ModAlt, Ch: 'I'}, nil
		case "<a-s-j>", "<alt-shift-j>":
			return KeyComb{Mod: ModAlt, Ch: 'J'}, nil
		case "<a-s-k>", "<alt-shift-k>":
			return KeyComb{Mod: ModAlt, Ch: 'K'}, nil
		case "<a-s-l>", "<alt-shift-l>":
			return KeyComb{Mod: ModAlt, Ch: 'L'}, nil
		case "<a-s-m>", "<alt-shift-m>":
			return KeyComb{Mod: ModAlt, Ch: 'M'}, nil
		case "<a-s-n>", "<alt-shift-n>":
			return KeyComb{Mod: ModAlt, Ch: 'N'}, nil
		case "<a-s-o>", "<alt-shift-o>":
			return KeyComb{Mod: ModAlt, Ch: 'O'}, nil
		case "<a-s-p>", "<alt-shift-p>":
			return KeyComb{Mod: ModAlt, Ch: 'P'}, nil
		case "<a-s-q>", "<alt-shift-q>":
			return KeyComb{Mod: ModAlt, Ch: 'Q'}, nil
		case "<a-s-r>", "<alt-shift-r>":
			return KeyComb{Mod: ModAlt, Ch: 'R'}, nil
		case "<a-s-s>", "<alt-shift-s>":
			return KeyComb{Mod: ModAlt, Ch: 'S'}, nil
		case "<a-s-t>", "<alt-shift-t>":
			return KeyComb{Mod: ModAlt, Ch: 'T'}, nil
		case "<a-s-u>", "<alt-shift-u>":
			return KeyComb{Mod: ModAlt, Ch: 'U'}, nil
		case "<a-s-v>", "<alt-shift-v>":
			return KeyComb{Mod: ModAlt, Ch: 'V'}, nil
		case "<a-s-w>", "<alt-shift-w>":
			return KeyComb{Mod: ModAlt, Ch: 'W'}, nil
		case "<a-s-x>", "<alt-shift-x>":
			return KeyComb{Mod: ModAlt, Ch: 'X'}, nil
		case "<a-s-y>", "<alt-shift-y>":
			return KeyComb{Mod: ModAlt, Ch: 'Y'}, nil
		case "<a-s-z>", "<alt-shift-z>":
			return KeyComb{Mod: ModAlt, Ch: 'Z'}, nil
		case "<a-s-[>", "<alt-shift-[>":
			return KeyComb{Mod: ModAlt, Ch: '{'}, nil
		case "<a-s-\\\\>", "<alt-shift-\\\\>":
			return KeyComb{Mod: ModAlt, Ch: '|'}, nil
		case "<a-s-]>", "<alt-shift-]>":
			return KeyComb{Mod: ModAlt, Ch: '}'}, nil
		case "<a-s-/>", "<alt-shift-/>":
			return KeyComb{Mod: ModAlt, Ch: '?'}, nil
		case "<a-s-_>", "<alt-shift-_>":
			return KeyComb{Mod: ModAlt, Ch: '_'}, nil
		case "<a-s-.>", "<alt-shift-.>":
			return KeyComb{Mod: ModAlt, Ch: '>'}, nil
		case "<a-s-,>", "<alt-shift-,>":
			return KeyComb{Mod: ModAlt, Ch: '<'}, nil
		case "<a-s-;>", "<alt-shift-;>":
			return KeyComb{Mod: ModAlt, Ch: ':'}, nil
		case "<a-s-'>", "<alt-shift-'>":
			return KeyComb{Mod: ModAlt, Ch: '"'}, nil
		case "<a-s-=>", "<alt-shift-=>":
			return KeyComb{Mod: ModAlt, Ch: '+'}, nil
		case "<a-s-+>", "<alt-shift-+>":
			return KeyComb{Mod: ModAlt, Ch: '+'}, nil
		case "<a-+>", "<alt-+>":
			return KeyComb{Mod: ModAlt, Ch: '+'}, nil
		case "<a-s-->", "<alt-shift-->":
			return KeyComb{Mod: ModAlt, Ch: '_'}, nil

		case "<alt>":
			return KeyComb{Mod: ModAlt}, nil
		case "<shift>":
			return KeyComb{Mod: ModShift}, nil
		case "<meta>":
			return KeyComb{Mod: ModMeta}, nil
		case "<ctrl>":
			return KeyComb{Mod: ModCtrl}, nil
		case "<ctrl-shift>":
			return KeyComb{Mod: ModCtrlShift}, nil
		case "<ctrl-alt>":
			return KeyComb{Mod: ModCtrlAlt}, nil
		case "<ctrl-meta>":
			return KeyComb{Mod: ModCtrlMeta}, nil
		case "<ctrl-shift-alt>":
			return KeyComb{Mod: ModCtrlShiftAlt}, nil
		case "<ctrl-shift-meta>":
			return KeyComb{Mod: ModCtrlShiftMeta}, nil
		case "<ctrl-alt-meta>":
			return KeyComb{Mod: ModCtrlAltMeta}, nil
		case "<shift-meta>":
			return KeyComb{Mod: ModShiftMeta}, nil
		case "<alt-meta>":
			return KeyComb{Mod: ModAltMeta}, nil
		case "<alt-shift-meta>":
			return KeyComb{Mod: ModAltShiftMeta}, nil
		case "<alt-shift>":
			return KeyComb{Mod: ModAltShift}, nil

		default:
			return KeyComb{}, fmt.Errorf("invalid key: '%s'", str)
		}
	}
}
