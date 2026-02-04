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
	"io"
)

const (
	// DefaultCreatedAtField is the document field that is automatically added
	// by Service implementors. See Service.Create for more info.
	DefaultCreatedAtField = "CreatedAt"

	// DefaultUpdatedAtField is the document field that is automatically updated
	// by Service implementors. See Service.Update for more info.
	DefaultUpdatedAtField = "UpdatedAt"

	// LowerCreatedAtField is the document field that is automatically added
	// by Service implementors when using a marshaler that defaults to lower case,
	// either directly or indirectly through another Service.
	LowerCreatedAtField = "createdat"

	// LowerUpdatedAtField is the document field that is automatically updated
	// by Service implementors when using a marshaler that defaults to lower case,
	// either directly or indirectly through another Service.
	LowerUpdatedAtField = "updatedat"
)

// Service is the interface that encapsulates a document store service.
//
// The following rules must be followed in order to guarantee compatibility
// across implementations:
//
//   - Document structures  must not contain embedded public fields.
//     This causes Update problems, as some implementations store the
//     embedded fields as a nested field, whereas others store them using
//     Go's internal representation.
//   - Document structures must not attempt to rename fields with tags. Some
//     implementations do not take tags, so Update operations, which use field
//     names string literals, would work for some implementations but not others.
type Service interface {
	// Create creates the document with the given data.
	// It returns an ErrAlreadyExists if a document with the same ID already exists.
	// The data argument can be a map with string keys, a struct, or a pointer
	// to a struct. The map keys or exported struct fields become the
	// fields of the document.
	//
	// Pointers and the empty any are permitted as
	// struct attributes or map values, and their elements processed recursively.
	//
	// DefaultCreatedAtField is automatically added and clients can consume it
	// by adding the corresponding property in the document structure.
	// Note that certain implementations might require special field tags.
	Create(ctx context.Context, ID string, doc any) error

	// Set creates a document with the given data or updates it if it already exists.
	//
	// DefaultUpdatedAtField is automatically updated and clients can consume it
	// by adding the corresponding property in the document structure.
	//
	// See Create for more details.
	Set(ctx context.Context, ID string, doc any) error

	// Update updates the document. The values at the given
	// field paths are replaced, but other fields of the stored document
	// are untouched. If one of the preconditions fails, it returns
	// ErrPreconditionFailed. If the document with the given ID doesn't exist
	// it returns ErrNotFound.
	//
	// DefaultUpdatedAtField is automatically updated and clients can consume it
	// by adding the corresponding property in the document structure.
	Update(ctx context.Context, ID string,
		updates []Update, precond ...Precondition) error

	// Get retrieves the document. If the document does not exist,
	// it returns a ErrNotFound error.
	// Parameter doc is used to populate the document's fields.
	// It can be a pointer to a map[string]any or a pointer to a struct.
	// If the document with the given ID doesn't exist
	// it returns ErrNotFound.
	Get(ctx context.Context, ID string, doc any) error

	// Delete deletes the document. If the document doesn't exist,
	// it does nothing and returns no error.
	Delete(ctx context.Context, ID string) error

	// The List operation returns a page of all documents in the collection.
	// To return a subset of the collection, you can provide a set of filters.
	List(ctx context.Context, filters []Filter) (Iterator, error)

	io.Closer
}

// DroppableService wraps a Service and provides a method to delete all records
// efficiently.
type DroppableService interface {
	Service

	// Drop deletes all records in a document.Service. Implementors must guarantee
	// that (1) this is done efficiently and (2) the service remains functional
	// after this operation succeeds.
	Drop(context.Context) error
}

// Iterator is used to collect the results obtained by List.
//
// HasNext is used to check how many results are left in the iterator.
// When HasNext returns false, a call to NextTo will panic.
//
// NextTo marshals the next document into the provided argument.
// It returns an error if marshaling fails. Once marshaled, the
// given document should not be re-used in the next call to NextTo
// otherwise map or slice fields could be overriden, depending on the
// implementation.
type Iterator interface {
	HasNext() bool
	NextTo(doc any) error
	io.Closer
}

// Field represents a document field.
type Field struct {
	FieldPath []string
	Value     any
}

// Update is used to indicate an update operation to a document field.
type Update Field

// Preconditions are optionally passed to Update to fail
// if the document state is not expected by the caller.
type Precondition Field

// Filter is used to construct a filter predicate in a List operation.
type Filter struct {
	Field
	Op
}

// Op represents a type of filter expression in a filter predicate.
type Op string

const (
	// OpEqual evalates true if the value in the Filter and the value
	// stored are equal.
	OpEqual Op = "=="
	// OpGreaterThan evalates true if the value in the Filter
	// is greater than the value stored in the collection.
	OpGreaterThan = ">"
	// OpGreaterThanEqual evalates true if the value in the Filter
	// is greater than or equal to the value stored in the collection.
	OpGreaterThanEqual = ">="
	// OpLessThan evalates true if the value in the Filter
	// is smaller than the value stored in the collection.
	OpLessThan = "<"
	// OpLessThanEqual evalates true if the value in the Filter
	// is smaller than or equal to the value stored in the collection.
	OpLessThanEqual = "<="
)
