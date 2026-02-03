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

package config

import "github.com/unstablebuild/tcell/v3"

// ErrConfig returns a Config that returns err to all methods of Config.
func ErrConfig(err error) Config {
	return errConfig{err: err}
}

type errConfig struct {
	err error
}

func (e errConfig) GetInt(string) (int, error) {
	return 0, e.err
}

func (e errConfig) GetFloat(string) (float64, error) {
	return 0, e.err
}

func (e errConfig) GetString(string) (string, error) {
	return "", e.err
}

func (e errConfig) GetBool(string) (bool, error) {
	return false, e.err
}

func (e errConfig) GetConfig(string) (Config, error) {
	return nil, e.err
}

func (e errConfig) GetMap(string) (map[string]interface{}, error) {
	return nil, e.err
}

func (e errConfig) GetAttribute(string) (tcell.AttrMask, error) {
	return 0, e.err
}

func (e errConfig) GetColor(string) (tcell.Color, error) {
	return 0, e.err
}

func (e errConfig) GetRune(string) (rune, error) {
	return 0, e.err
}

func (e errConfig) GetSlice(string) ([]interface{}, error) {
	return nil, e.err
}

func (e errConfig) Iterate(fn func(k string, value interface{})) {
}
