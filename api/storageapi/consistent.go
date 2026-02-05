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

package storageapi

import (
	"context"
	"errors"

	"github.com/unstablebuild/rune-go-sdk/retry"
)

// ConsistentUpdate calls Update on the given service and tries to update
// the underlying document with the provided retry strategy and the
// callback function which gets called before attempting to update
// the document in storage. This can be used to scan the updated
// doc for any data that has been updated in the process of performing
// this operation and to update preconditions with the latest values.
//
// The given doc value will be updated so it should be a valid value as
// defined by Service.Get.
func ConsistentUpdate(
	ctx context.Context, svc Service, ID string, doc interface{},
	retryStrategy retry.Strategy, callback func() ([]Update, []Precondition),
) error {
	return retry.Retry(ctx, retryStrategy, func(ctx context.Context) (bool, error) {
		if err := svc.Get(ctx, ID, doc); err != nil {
			return false, err
		}
		updates, preconditions := callback()
		err := svc.Update(ctx, ID, updates, preconditions...)
		if err == nil {
			return false, svc.Get(ctx, ID, doc)
		}
		return errors.Is(err, ErrPreconditionFailed), err
	})
}
