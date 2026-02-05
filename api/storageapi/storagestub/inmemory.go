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
package storagestub

import (
	"context"
	"reflect"
	"sync"

	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal/docbson"
)

type inMemoryService struct {
	marshaler docmarshal.Marshaler
	m         sync.Locker
	storage   map[string][]byte
}

// NewInMemoryService returns an instance of storageapi.Service backed
// by an in-memory map, and the default bson marshaler.
func NewInMemoryService() storageapi.DroppableService {
	return NewInMemoryServiceWithMarshaler(docbson.Marshaler())
}

// NewInMemoryServiceWithMarshaler returns a in-memory service powered
// by the given marshaler.
func NewInMemoryServiceWithMarshaler(m docmarshal.Marshaler) storageapi.DroppableService {
	return &inMemoryService{
		marshaler: m,
		m:         new(sync.Mutex),
		storage:   make(map[string][]byte),
	}
}

func (c *inMemoryService) Set(
	ctx context.Context, ID string, data interface{},
) error {
	return c.set(ctx, ID, data, false)
}

func (c *inMemoryService) Create(
	ctx context.Context, ID string, data interface{},
) error {
	return c.set(ctx, ID, data, true)
}

func (c *inMemoryService) set(
	ctx context.Context, ID string, data interface{},
	errAlreadyExists bool,
) (err error) {
	if data == nil {
		panic("invalid nil data argument to Create/Set")
	}
	data, err = DerefCreateValue(reflect.ValueOf(data))
	if err != nil {
		return
	}

	c.m.Lock()
	defer c.m.Unlock()

	if errAlreadyExists {
		_, ok := c.storage[ID]
		if ok {
			return storageapi.ErrAlreadyExists
		}
	}
	c.storage[ID] = Encode(c.marshaler, data, true)
	return
}

func (c *inMemoryService) getValue(ID string, doc interface{}) (
	err error,
) {
	var ok bool

	var raw []byte
	raw, ok = c.storage[ID]
	if !ok {
		err = storageapi.ErrNotFound
		return
	}

	return SafeDecode(c.marshaler, doc, raw)
}

func (c *inMemoryService) Get(
	ctx context.Context, ID string, to interface{},
) (err error) {
	c.m.Lock()
	defer c.m.Unlock()

	err = c.getValue(ID, to)
	return
}

func (c *inMemoryService) Update(
	ctx context.Context, ID string, updates []storageapi.Update,
	preconds ...storageapi.Precondition,
) error {
	if len(updates) == 0 {
		panic("Update: no paths to update")
	}

	c.m.Lock()
	defer c.m.Unlock()

	var proto map[string]interface{}
	err := c.getValue(ID, &proto)
	if err != nil {
		return err
	}

	err = UpdateProto(c.marshaler, updates, proto, preconds...)
	if err != nil {
		return err
	}

	c.storage[ID] = Encode(c.marshaler, proto, false)

	return nil
}

func (c *inMemoryService) Close() error {
	return nil
}

func (c *inMemoryService) Delete(ctx context.Context, ID string) error {
	c.m.Lock()
	defer c.m.Unlock()

	delete(c.storage, ID)
	return nil
}

func (c *inMemoryService) List(ctx context.Context, filters []storageapi.Filter) (
	it storageapi.Iterator, err error,
) {
	iter := NewListIterator(c.marshaler)

	c.m.Lock()
	defer c.m.Unlock()

	for _, v := range c.storage {
		iter.Extend(filters, v)
	}

	return iter, nil
}

func (c *inMemoryService) Drop(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()

	c.storage = make(map[string][]byte)
	return nil
}
