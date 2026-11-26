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

package browserapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/term"
)

func TestPromptConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     PromptConfig
		wantErr string
	}{
		{
			name: "a message and an option",
			cfg:  PromptConfig{Message: "?", Options: []string{"Ok"}},
		},
		{
			name: "fields with questions",
			cfg: PromptConfig{Message: "?", Options: []string{"Ok"}, Fields: []Field{
				{Question: "User", Answer: "me"},
				{Question: "Password", Secret: true, Mandatory: true},
			}},
		},
		{
			name:    "no message",
			cfg:     PromptConfig{Options: []string{"Ok"}},
			wantErr: "prompt has no message",
		},
		{
			name:    "no options",
			cfg:     PromptConfig{Message: "?", Fields: []Field{{Question: "User"}}},
			wantErr: "prompt has no options",
		},
		{
			name: "fewer bindings than options",
			cfg: PromptConfig{
				Message: "?", Options: []string{"Yes", "No"}, Bindings: []term.KeyComb{{Ch: 'y'}},
			},
			wantErr: "prompt has a different number of bindings and options",
		},
		{
			name:    "a negative max width",
			cfg:     PromptConfig{Message: "?", Options: []string{"Ok"}, MaxWidth: -1},
			wantErr: "prompt has a negative max width",
		},
		{
			name: "a field without a question",
			cfg: PromptConfig{Message: "Rename old to:", Options: []string{"Rename"}, Fields: []Field{
				{Answer: "old", Mandatory: true},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}
