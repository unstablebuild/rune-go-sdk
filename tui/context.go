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

package tui

import (
	"context"
	"strconv"

	"github.com/unstablebuild/rune-go-sdk/term"
)

// ContextWithIteration returns a new Context that holds the iteration number i as a payload.
func ContextWithIteration(ctx context.Context, i int64) context.Context {
	return term.ContextWithPayload(ctx, []byte(strconv.FormatInt(i, 10)))
}

// IterationFromContext returns the ID value stored in ctx, if any.
func IterationFromContext(ctx context.Context) (int64, bool) {
	payload, ok := term.PayloadFromContext(ctx)
	if !ok {
		return 0, false
	}
	return IterationFromRawBytes(payload)
}

// IterationFromRawBytes parses the given payload into an iteration number,
// as formatted by ContextWithIteration.
func IterationFromRawBytes(payload []byte) (int64, bool) {
	i, err := strconv.ParseInt(string(payload), 10, 64)
	if err != nil {
		return 0, false
	}
	return i, true
}
