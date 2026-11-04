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

package acpapi_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/acpapi"
)

func TestNullableJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
		want acpapi.Nullable[string]
		// again is how want marshals, when not json.
		again string
	}{
		{name: "omitted", json: `{"sessionId":"a"}`},
		{name: "null", json: `{"sessionId":"a","title":null}`, want: acpapi.Null[string]()},
		{name: "value", json: `{"sessionId":"a","title":"Explore"}`, want: acpapi.Some("Explore")},
		{name: "invalid", json: `{"sessionId":"a","title":5}`, again: `{"sessionId":"a"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var u acpapi.SubagentUpdate
			require.NoError(t, json.Unmarshal([]byte(tc.json), &u))
			assert.Equal(t, tc.want, u.Title)
			b, err := json.Marshal(u)
			require.NoError(t, err)
			again := tc.again
			if again == "" {
				again = tc.json
			}
			assert.JSONEq(t, again, string(b))
		})
	}
}
