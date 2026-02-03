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
	"context"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type ctxKey int

// pKey is the key for sync.Payload values in Contexts. It is
// unexported; clients use workspace.ContextWithPayload and
// PayloadFromContext instead of using this key directly.
var pKey ctxKey

// ContextWithPayload returns a new Context that holds locker.
func ContextWithPayload(ctx context.Context, payload []byte) context.Context {
	return context.WithValue(ctx, pKey, payload)
}

// PayloadFromContext returns the payload value stored in ctx, if any.
func PayloadFromContext(ctx context.Context) ([]byte, bool) {
	locker, ok := ctx.Value(pKey).([]byte)
	return locker, ok
}
