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

package doctoml

import (
	"bytes"

	"github.com/BurntSushi/toml"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi/docmarshal"
)

// Marshaler returns a TOML Marshaler.
func Marshaler() docmarshal.Marshaler {
	return tomlMarshaler{}
}

type tomlMarshaler struct {
}

func (j tomlMarshaler) Marshal(in any) ([]byte, error) {
	var buf bytes.Buffer
	err := toml.NewEncoder(&buf).Encode(in)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (j tomlMarshaler) Unmarshal(data []byte, to any) error {
	_, err := toml.NewDecoder(bytes.NewReader(data)).Decode(to)
	return err
}

func (j tomlMarshaler) DefaultLowerCase() bool {
	return false
}
