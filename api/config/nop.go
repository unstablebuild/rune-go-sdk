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

import "github.com/unstablebuild/rune-go-sdk/term"

type nopConfig struct{}

// NopConfig returns a Config that always returns ErrNotFound.
func NopConfig() Config {
	return nopConfig{}
}

func (n nopConfig) GetInt(string) (int, error) {
	return 0, ErrNotFound
}

func (n nopConfig) GetFloat(string) (float64, error) {
	return 0, ErrNotFound
}

func (n nopConfig) GetString(string) (string, error) {
	return "", ErrNotFound
}

func (n nopConfig) GetBool(string) (bool, error) {
	return false, ErrNotFound
}

func (n nopConfig) GetConfig(string) (Config, error) {
	return nil, ErrNotFound
}

func (n nopConfig) GetMap(string) (map[string]interface{}, error) {
	return nil, ErrNotFound
}

func (n nopConfig) GetAttribute(string) (term.AttrMask, error) {
	return 0, ErrNotFound
}

func (n nopConfig) GetColor(string) (term.Color, error) {
	return 0, ErrNotFound
}

func (n nopConfig) GetRune(string) (rune, error) {
	return 0, ErrNotFound
}

func (n nopConfig) GetSlice(string) ([]interface{}, error) {
	return nil, ErrNotFound
}

func (n nopConfig) Iterate(fn func(k string, value interface{})) {
}
