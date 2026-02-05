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

package sysclip

import (
	"errors"
	"sync/atomic"

	sysclip "github.com/atotto/clipboard"
	"github.com/unstablebuild/rune-go-sdk/clipboard"
)

type register struct {
	data atomic.Value
}

type registerData struct {
	metadata any
	data     string
}

// NewRegister allocates initializes a new system clipboard.
func NewRegister() (clipboard.Register, error) {
	if sysclip.Unsupported {
		return nil, errors.New("system clipboard unsupported")
	}
	ret := new(register)
	ret.data.Store(registerData{})
	return ret, nil
}

// Paste satisfies clipboard.Register.
func (r *register) Paste(id string) (ret clipboard.Data, err error) {
	text, err := sysclip.ReadAll()
	if err != nil {
		return ret, err
	}
	// only return metadata if it matches last copy
	ret.Text = text
	data := r.data.Load().(registerData)
	if data.data != text {
		return
	}
	ret.Metadata = data.metadata
	return
}

// Copy satisfies clipboard.Register.
func (r *register) Copy(id string, data clipboard.Data) error {
	r.data.Store(registerData{data: data.Text, metadata: data.Metadata})
	return sysclip.WriteAll(data.Text)
}
