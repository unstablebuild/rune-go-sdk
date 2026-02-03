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
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docmarshal"
)

// UpdateUpdatedAtField updates the default UpdatedAt field in the given document.
// Deprecated: Use UpdateDefaultUpdatedAtField.
func UpdateUpdatedAtField(marshaler docmarshal.Marshaler, doc any) any {
	now := time.Now()
	return UpdateDefaultUpdatedAtField(doc, now, marshaler.DefaultLowerCase())
}

// UpdateDefaultUpdatedAtField updates the default UpdatedAt field in the given document.
// If lowerCase is false DefaultUpdatedAtField is used, otherwise LowerUpdatedAtField is used.
// If doc is not a map, a struct or a pointer to a struct, this method panics.
func UpdateDefaultUpdatedAtField(doc any, now time.Time, lowerCase bool) any {
	updatedAtField := DefaultUpdatedAtField
	if lowerCase {
		updatedAtField = LowerUpdatedAtField
	}
	m, ok := doc.(map[string]any)
	if ok {
		doc = setMapField(m, updatedAtField, now)
	} else if reflect.ValueOf(doc).Kind() == reflect.Struct {
		dst := cloneReflectValue(doc)
		reflectSetTimeField(dst, updatedAtField, now)
		doc = dst.Interface()
	} else {
		dst := reflect.ValueOf(doc)
		reflectSetTimeField(dst, updatedAtField, now)
		doc = dst.Interface()
	}
	return doc
}

// UpdateCreatedAtField calls UpdateCreatedAtFieldTime with time.Now.
// Deprecated: Use UpdateDefaultCreatedAtField.
func UpdateCreatedAtField(marshaler docmarshal.Marshaler, doc any) any {
	now := time.Now()
	return UpdateDefaultCreatedAtField(doc, now, marshaler.DefaultLowerCase())
}

// DerefUpdateValue dereferences data until it finds a value that can be used
// by UpdateDefaultCreatedAtField and UpdateDefaultUpdatedAtField to update
// their updated at or created at fields.
func DerefUpdateValue(data reflect.Value) (any, error) {
	prev := data
	for {
		switch data.Kind() {
		case reflect.Struct:
			if prev.Kind() == reflect.Ptr {
				return prev.Interface(), nil
			}
			return nil, errors.New("only pointers to a struct or map values are allowed")
		case reflect.Map:
			return data.Interface(), nil
		case reflect.Ptr:
			prev = data
			data = data.Elem()
		case reflect.Interface:
			if data.NumMethod() == 0 {
				prev = data
				data = data.Elem()
				continue
			}
			fallthrough
		default:
			return nil, errors.New("only pointers to a struct or map values are allowed")
		}
	}
}

// UpdateDefaultCreatedAtField updates the default CreatedAt field in the given document
// and the default UpdatedAt field.
// If lowerCase is false DefaultUpdatedAtField is used, otherwise LowerUpdatedAtField is used.
// If doc is not a map, a struct or a pointer to a struct, this method panics.
func UpdateDefaultCreatedAtField(doc any, now time.Time, lowerCase bool) any {
	updatedAtField := DefaultUpdatedAtField
	createdAtField := DefaultCreatedAtField
	m, ok := doc.(map[string]any)
	if lowerCase {
		updatedAtField = LowerUpdatedAtField
		createdAtField = LowerCreatedAtField
	}
	if ok {
		setMapField(m, updatedAtField, now)
		setMapField(m, createdAtField, now)
	} else if reflect.ValueOf(doc).Kind() == reflect.Struct {
		dst := cloneReflectValue(doc)
		now := time.Now()
		reflectSetTimeField(dst, DefaultCreatedAtField, now)
		reflectSetTimeField(dst, DefaultUpdatedAtField, now)
		doc = dst.Elem().Interface()
	} else {
		dst := reflect.ValueOf(doc)
		now := time.Now()
		reflectSetTimeField(dst, DefaultCreatedAtField, now)
		reflectSetTimeField(dst, DefaultUpdatedAtField, now)
		doc = dst.Elem().Interface()
	}
	return doc
}

// Encode encodes doc into a reversible format (via Decode)
// and returns the data in bytes.
func Encode(m docmarshal.Marshaler, doc any, addCreatedAt bool) []byte {
	if addCreatedAt {
		doc = UpdateCreatedAtField(m, doc)
	} else {
		doc = UpdateUpdatedAtField(m, doc)
	}

	b, err := m.Marshal(doc)
	if err != nil {
		panic(err)
	}
	return b
}

// SafeDecode checks if the given interface would be decoded by Decode
// and decodes it or otherwise returns an error.
func SafeDecode(m docmarshal.Marshaler, rcv any, raw []byte) error {
	if !IsEncodeable(rcv) {
		return errors.New("receiver is not a pointer and not a map or is nil")
	}
	return decode(m, rcv, raw)
}

// IsEncodeable returns true if doc is a structure that can be safely
// decoded via Decode.
func IsEncodeable(doc any) bool {
	v := reflect.ValueOf(doc)
	return (v.Kind() == reflect.Ptr || v.Kind() == reflect.Map) && !v.IsNil()
}

// ListIterator returns an iterator that iterates over docs.
// It will uson bson to encode and decode the data so
// it shouldn't be used by a document.Service that doesn't use
// the suite of Decode/Encode functions in this package.
func NewListIterator(m docmarshal.Marshaler, docs ...any) *ListIterator {
	iter := &ListIterator{marshaler: m, docs: make([][]byte, 0)}
	for _, data := range docs {
		iter.Extend(nil, Encode(m, data, false))
	}
	return iter
}

// ListIterator satisfies an Iterator with an inmemory
// list of documents.
type ListIterator struct {
	marshaler docmarshal.Marshaler
	docs      [][]byte
}

// HasNext returns false if this Iterator is empty.
func (l *ListIterator) HasNext() bool {
	return len(l.docs) > 0
}

// NextTo decodes the next chunk of data into doc or returns
// an error if there was a decoding issue.
func (l *ListIterator) NextTo(doc any) error {
	if err := SafeDecode(l.marshaler, doc, l.docs[0]); err != nil {
		return err
	}
	l.docs = l.docs[1:]
	return nil
}

// Close does nothing.
func (l *ListIterator) Close() error {
	return nil
}

// Extend extends this iterator if and only if the data chunk's
// structure satisfies all filters.
func (l *ListIterator) Extend(filters []Filter, v []byte) {
	var proto map[string]any
	// if bogus data is inserted into the DB oob,
	// then we ignore it here, rather than failing a list operation
	err := decode(l.marshaler, &proto, v)
	if err != nil {
		slog.Warn("failed to add value to iterator", "reason", "decode", "error", err)
		return
	}
	if !MatchesAllFilters(l.marshaler, proto, filters) {
		return
	}

	copied := make([]byte, len(v))
	copy(copied, v)
	l.docs = append(l.docs, copied)
}

// UpdateProto updates proto with the given slice of updates.
func UpdateProto(m docmarshal.Marshaler, updates []Update,
	proto map[string]any, preconds ...Precondition) error {
	lowerCase := m.DefaultLowerCase()

	updatedAtField := DefaultUpdatedAtField
	if lowerCase {
		updatedAtField = LowerUpdatedAtField
	}

	for _, cond := range preconds {
		if len(cond.FieldPath) == 0 {
			panic("empty field path")
		}

		fieldPath := cond.FieldPath
		if lowerCase {
			fieldPath = make([]string, len(cond.FieldPath))
			for i, comp := range cond.FieldPath {
				fieldPath[i] = strings.ToLower(comp)
			}
		}

		if !precondField(proto, Precondition{FieldPath: fieldPath, Value: cond.Value}) {
			return ErrPreconditionFailed
		}
	}

	for _, update := range updates {
		if len(update.FieldPath) == 0 {
			panic("empty field path")
		}
		if update.FieldPath[0] == updatedAtField {
			continue
		}

		fieldPath := update.FieldPath
		if lowerCase {
			fieldPath = make([]string, len(update.FieldPath))
			for i, comp := range update.FieldPath {
				fieldPath[i] = strings.ToLower(comp)
			}
		}

		updateField(proto, Update{FieldPath: fieldPath, Value: update.Value})
	}

	updateField(proto, Update{
		FieldPath: []string{updatedAtField},
		Value:     time.Now(),
	})

	return nil
}

// DerefCreateValue dereferences data for a service.Create implementation
// until it finds a structure that can be used for Encode/Decode
// or returns an error if no such structure could be found.
func DerefCreateValue(data reflect.Value) (any, error) {
	for {
		switch data.Kind() {
		case reflect.Struct, reflect.Map:
			return data.Interface(), nil
		case reflect.Ptr:
			data = data.Elem()
		case reflect.Interface:
			if data.NumMethod() == 0 {
				data = data.Elem()
				continue
			}
			fallthrough
		default:
			return nil, errors.New("only struct or map values are allowed")
		}
	}
}

func reflectSetTimeField(s reflect.Value, k string, v time.Time) {
	f := s.Elem().FieldByName(k)
	if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.Struct ||
		reflect.TypeOf(f) == reflect.TypeOf((*time.Time)(nil)).Elem() {
		return
	}
	f.Set(reflect.ValueOf(v))
}

func cloneReflectValue(data any) reflect.Value {
	typ := reflect.TypeOf(data)
	src := reflect.ValueOf(data)
	dst := reflect.New(typ)
	for i := 0; i < src.NumField(); i++ {
		dstField := dst.Elem().Field(i)
		if dstField.CanSet() {
			dstField.Set(src.Field(i))
		}
	}
	return dst
}

func setMapField(m map[string]any, key string, value any) map[string]any {
	ret := make(map[string]any)
	for k, v := range m {
		ret[k] = v
	}
	ret[key] = value
	return ret
}

func precondField(proto map[string]any, cond Precondition) bool {
	if len(cond.FieldPath) == 1 {
		return proto[cond.FieldPath[0]] == cond.Value
	}

	field, exist := proto[cond.FieldPath[0]]
	if !exist {
		return false
	}

	m, ok := field.(map[string]any)
	if !ok {
		return false
	}

	cond.FieldPath = cond.FieldPath[1:]
	return precondField(m, cond)
}

func updateField(proto map[string]any, update Update) {
	if len(update.FieldPath) == 1 {
		proto[update.FieldPath[0]] = update.Value
		return
	}

	field, exist := proto[update.FieldPath[0]]
	if !exist {
		// surprisingly, FireStore does not return an error and also
		// it doesn't add the extra fields. Since we are emulating
		// the same behaviour, just return silently instead of
		// creating the attribute path.
		return
	}

	m, ok := field.(map[string]any)
	if !ok {
		return
	}

	update.FieldPath = update.FieldPath[1:]
	updateField(m, update)
}

type filterAsserter struct {
	res *bool
}

func (f *filterAsserter) Errorf(_ string, _ ...any) {
	*f.res = false
}

func doMatchFilter(value any, f Filter) bool {
	matches := true
	t := filterAsserter{res: &matches}

	switch f.Op {
	case OpEqual:
		assert.EqualValues(&t, value, f.Value)
	case OpGreaterThan:
		assert.Greater(&t, value, f.Value)
	case OpLessThan:
		assert.Less(&t, value, f.Value)
	case OpGreaterThanEqual:
		assert.EqualValues(&t, value, f.Value)
		if !matches {
			matches = true
			assert.GreaterOrEqual(&t, value, f.Value)
		}
	case OpLessThanEqual:
		assert.EqualValues(&t, value, f.Value)
		if !matches {
			matches = true
			assert.LessOrEqual(&t, value, f.Value)
		}
	default:
		panic("unexpected op")
	}

	return matches
}

// MatchFilter returns true if proto satisfies the filter condition of f.
func MatchFilter(proto map[string]any, f Filter) bool {
	if len(f.FieldPath) == 1 {
		return doMatchFilter(proto[f.FieldPath[0]], f)
	}

	field, exist := proto[f.FieldPath[0]]
	if !exist || field == nil {
		return false
	}

	m := field.(map[string]any)
	f.FieldPath = f.FieldPath[1:]
	return MatchFilter(m, f)
}

// MatchesAllFilters returns true if proto satisfies all of the give filters.
func MatchesAllFilters(
	m docmarshal.Marshaler, proto map[string]any,
	filters []Filter,
) bool {
	for _, f := range filters {
		if m.DefaultLowerCase() {
			fieldPath := make([]string, len(f.FieldPath))
			for i, comp := range f.FieldPath {
				fieldPath[i] = strings.ToLower(comp)
			}
			f.FieldPath = fieldPath
		}

		if !MatchFilter(proto, f) {
			return false
		}
	}
	return true
}

func decode(m docmarshal.Marshaler, rcv any, raw []byte) error {
	return m.Unmarshal(raw, rcv)
}
