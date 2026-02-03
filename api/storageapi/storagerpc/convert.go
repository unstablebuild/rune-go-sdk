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
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docmarshal"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/storagerpc/docpb"
)

// this is just a trick to be able to re-use encode functionality
const protoFieldKey = "X"

func makeProtoUpdates(m docmarshal.Marshaler, updates []storageapi.Update) (
	ret []*docpb.UpdateDocumentRequest_Field,
) {
	slab := make(map[string]any)
	for _, u := range updates {
		// re-use make filter logic
		f := storageapi.Filter{Field: storageapi.Field(u)}
		pf := makeProtoFilter(m, slab, f)
		pu := &docpb.UpdateDocumentRequest_Field{
			FieldPath: pf.FieldPath,
			Data:      pf.Data,
		}
		ret = append(ret, pu)
	}
	return
}

func makeProtoPreconditions(m docmarshal.Marshaler, preconds ...storageapi.Precondition) (
	ret []*docpb.UpdateDocumentRequest_Field,
) {
	slab := make(map[string]any)
	for _, u := range preconds {
		f := storageapi.Filter{Field: storageapi.Field(u)}
		pf := makeProtoFilter(m, slab, f)
		pu := &docpb.UpdateDocumentRequest_Field{
			FieldPath: pf.FieldPath,
			Data:      pf.Data,
		}
		ret = append(ret, pu)
	}
	return
}

func makeProtoFilter(
	m docmarshal.Marshaler,
	slab map[string]any, f storageapi.Filter,
) docpb.ListDocumentRequest_Filter {
	slab[protoFieldKey] = f.Value

	return docpb.ListDocumentRequest_Filter{
		FieldPath: f.FieldPath,
		Data:      storageapi.Encode(m, slab, false),
		Operation: string(f.Op),
	}
}

func makeProtoFilters(m docmarshal.Marshaler, filters []storageapi.Filter) (
	ret []*docpb.ListDocumentRequest_Filter, err error,
) {
	slab := make(map[string]any)
	for _, f := range filters {
		if len(f.FieldPath) == 0 {
			panic("invalid List filter: empty zero-valued FieldPath")
		}
		pf := new(docpb.ListDocumentRequest_Filter)
		*pf = makeProtoFilter(m, slab, f)

		ret = append(ret, pf)
	}
	return
}
