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
	"fmt"
)

// WithPartition wraps a service and creates a partition with the given name.
// All the records created, stored, listed, etc. won't be seen by other partitions
// over the same Service, created by this function.
func WithPartition(other Service, partition string) Service {
	c := new(partitionedService)
	c.other = other
	c.partition = partition
	// field key needs to be unique so partitions can be nested
	c.partitionField = "__partition_" + partition
	return c
}

type partitionedService struct {
	other          Service
	partition      string
	partitionField string
}

func (c *partitionedService) makePartitionID(ID string) string {
	return fmt.Sprintf("%s.%s", c.partition, ID)
}

func (c *partitionedService) setPartitionField(ctx context.Context, ID string) error {
	// enable list to filter document of this partition only
	updates := []Update{{FieldPath: []string{c.partitionField}, Value: c.partition}}
	if err := c.other.Update(ctx, ID, updates); err != nil {
		_ = c.other.Delete(ctx, ID) // best effort
		return err
	}
	return nil
}

func (c *partitionedService) Create(ctx context.Context, ID string, doc interface{}) error {
	ID = c.makePartitionID(ID)
	if err := c.other.Create(ctx, ID, doc); err != nil {
		return err
	}
	return c.setPartitionField(ctx, ID)
}

func (c *partitionedService) Set(ctx context.Context, ID string, doc interface{}) error {
	ID = c.makePartitionID(ID)
	if err := c.other.Set(ctx, ID, doc); err != nil {
		return err
	}
	return c.setPartitionField(ctx, ID)
}

func (c *partitionedService) Update(ctx context.Context, ID string,
	updates []Update, precond ...Precondition) error {
	ID = c.makePartitionID(ID)
	return c.other.Update(ctx, ID, updates, precond...)
}

func (c *partitionedService) Get(ctx context.Context, ID string, doc interface{}) error {
	ID = c.makePartitionID(ID)
	return c.other.Get(ctx, ID, doc)
}

func (c *partitionedService) Delete(ctx context.Context, ID string) error {
	ID = c.makePartitionID(ID)
	return c.other.Delete(ctx, ID)
}

func (c *partitionedService) List(
	ctx context.Context, filters []Filter,
) (Iterator, error) {
	partFilter := Filter{
		Op: OpEqual,
		Field: Field{
			FieldPath: []string{c.partitionField},
			Value:     c.partition,
		},
	}
	filters = append(filters, partFilter)
	return c.other.List(ctx, filters)
}

func (c *partitionedService) Close() error {
	return c.other.Close()
}
