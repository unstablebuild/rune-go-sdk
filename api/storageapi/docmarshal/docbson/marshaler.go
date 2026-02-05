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

package docbson

import (
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal"
	"gopkg.in/mgo.v2/bson"
)

// Marshaler returns a Marshaler backed by gopkg.in/mgo.v2/bson binary
// marshaler implementation.
func Marshaler() docmarshal.Marshaler {
	return bsonMarshaler{}
}

type bsonMarshaler struct {
}

func (b bsonMarshaler) Marshal(doc interface{}) ([]byte, error) {
	return bson.Marshal(doc)
}
func (b bsonMarshaler) Unmarshal(data []byte, doc interface{}) error {
	return bson.Unmarshal(data, doc)
}

func (j bsonMarshaler) DefaultLowerCase() bool {
	return true
}
