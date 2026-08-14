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

import "context"

// BatchOpType selects which write a BatchOp performs.
type BatchOpType int

const (
	// BatchCreate mirrors Service.Create.
	BatchCreate BatchOpType = iota
	// BatchSet mirrors Service.Set.
	BatchSet
	// BatchUpdate mirrors Service.Update.
	BatchUpdate
	// BatchDelete mirrors Service.Delete.
	BatchDelete
)

// BatchOp is a single write of a batch. Doc is only read by BatchCreate
// and BatchSet; Updates and Preconditions only by BatchUpdate.
type BatchOp struct {
	Type          BatchOpType
	ID            string
	Doc           any
	Updates       []Update
	Preconditions []Precondition
}

// BatchOpResult reports the outcome of the BatchOp at the same index.
type BatchOpResult struct {
	// Err is one of ErrAlreadyExists, ErrNotFound or
	// ErrPreconditionFailed when the operation was rejected by the
	// store, and nil when it was applied.
	Err error
}

// BatchWriter is implemented by services that can apply several writes
// in a single atomic round trip.
type BatchWriter interface {
	// ApplyBatch applies every op to the same collection. Operations
	// rejected for document-level reasons (a Create over an existing
	// document, an Update whose preconditions fail) do not abort the
	// batch: their error is reported in the result at the same index
	// and the remaining ops are still applied. Any other failure aborts
	// the batch, leaves the store untouched, and is returned as the
	// call error.
	ApplyBatch(ctx context.Context, ops []BatchOp) ([]BatchOpResult, error)
}
