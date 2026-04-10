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

// WithPartition wraps a service and creates a partition with the given name.
// All the records created, stored, listed, etc. won't be seen by other partitions
// over the same Service, created by this function.
func WithPartition(other Service, partition string) Service {
	svc, err := other.Partition(partition)
	if err != nil {
		return &partitionErrorService{err: err}
	}
	return svc
}

type partitionErrorService struct {
	err error
}

func (s *partitionErrorService) Create(ctx context.Context, ID string, doc any) error {
	return s.err
}

func (s *partitionErrorService) Set(ctx context.Context, ID string, doc any) error {
	return s.err
}

func (s *partitionErrorService) Update(
	ctx context.Context, ID string, updates []Update, precond ...Precondition,
) error {
	return s.err
}

func (s *partitionErrorService) Get(ctx context.Context, ID string, doc any) error {
	return s.err
}

func (s *partitionErrorService) Delete(ctx context.Context, ID string) error {
	return s.err
}

func (s *partitionErrorService) List(ctx context.Context, filters []Filter) (Iterator, error) {
	return nil, s.err
}

func (s *partitionErrorService) Partition(name string) (Service, error) {
	return s, s.err
}

func (s *partitionErrorService) Close() error {
	return nil
}
