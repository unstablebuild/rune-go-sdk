// Copyright 2026 Unstable Build, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package storageapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagestub"
	"github.com/unstablebuild/rune-go-sdk/retry"
)

type testStruct struct {
	Queries   []string
	UpdatedAt time.Time
}

func TestConsistentUpdate(t *testing.T) {
	ctx := context.Background()
	svc := storagestub.NewInMemoryService()
	a := testStruct{Queries: []string{"A"}, UpdatedAt: time.Now()}
	require.NoError(t, svc.Create(ctx, "docID", &a))

	t1 := time.Now().Add(1 * time.Minute)
	b := testStruct{Queries: []string{"B"}, UpdatedAt: t1}

	err := storageapi.ConsistentUpdate(ctx, svc, "docID", &b, retry.LimitStrategy(2),
		func() ([]storageapi.Update, []storageapi.Precondition) {
			return []storageapi.Update{
					{
						FieldPath: []string{"Queries"},
						Value:     append(b.Queries, "C"),
					},
				}, []storageapi.Precondition{
					{
						FieldPath: []string{"UpdatedAt"},
						Value:     b.UpdatedAt,
					},
				}
		},
	)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"A", "C"}, b.Queries)

}

type countingService struct {
	storageapi.Service
	updates int
}

func (c *countingService) Update(
	ctx context.Context, ID string, updates []storageapi.Update,
	preconds ...storageapi.Precondition,
) error {
	c.updates++
	return c.Service.Update(ctx, ID, updates, preconds...)
}

func TestConsistentUpdateNoUpdatesSkipsWrite(t *testing.T) {
	ctx := context.Background()
	svc := &countingService{Service: storagestub.NewInMemoryService()}
	a := testStruct{Queries: []string{"A"}, UpdatedAt: time.Now()}
	require.NoError(t, svc.Create(ctx, "docID", &a))

	var b testStruct
	err := storageapi.ConsistentUpdate(ctx, svc, "docID", &b, retry.LimitStrategy(2),
		func() ([]storageapi.Update, []storageapi.Precondition) {
			return nil, nil
		},
	)
	require.NoError(t, err)
	assert.Zero(t, svc.updates, "Update must not be called when callback returns no updates")

	var got testStruct
	require.NoError(t, svc.Get(ctx, "docID", &got))
	assert.Equal(t, []string{"A"}, got.Queries)
}
