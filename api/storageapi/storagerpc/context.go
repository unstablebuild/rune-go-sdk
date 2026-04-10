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

package storagerpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type fieldsKey struct{}

const PartitionMetadataKey = "x-rune-storage-partition"

// WithFields returns a new context that carries the given field names
// for projection. When passed to Client.List, only the specified
// top-level fields will be included in each returned document.
// An empty or nil fields list means no projection (return all fields).
func WithFields(ctx context.Context, fields ...string) context.Context {
	return context.WithValue(ctx, fieldsKey{}, fields)
}

// fieldsFromContext extracts field-projection names from ctx.
// Returns nil if none were set.
func fieldsFromContext(ctx context.Context) []string {
	fields, _ := ctx.Value(fieldsKey{}).([]string)
	return fields
}

// ContextWithPartitions returns a context that sends partition names over gRPC.
func ContextWithPartitions(ctx context.Context, partitions ...string) context.Context {
	if len(partitions) == 0 {
		return ctx
	}
	args := make([]string, 0, len(partitions)*2)
	for _, partition := range partitions {
		args = append(args, PartitionMetadataKey, partition)
	}
	return metadata.AppendToOutgoingContext(ctx, args...)
}

// PartitionsFromIncomingContext extracts partition names sent over gRPC.
func PartitionsFromIncomingContext(ctx context.Context) []string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	partitions := md.Get(PartitionMetadataKey)
	if len(partitions) == 0 {
		return nil
	}
	ret := make([]string, len(partitions))
	copy(ret, partitions)
	return ret
}
