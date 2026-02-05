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

package clipboard

import "sync/atomic"

type inmemoryRegister struct {
	data atomic.Value
}

// NewInMemory returns a simple in-memory implementation of Register.
func NewInMemory() Register {
	ret := new(inmemoryRegister)
	// initialize atomic so below we can simply to a type-assertion
	// without further branches.
	ret.data.Store(Data{})
	return ret
}

func (d *inmemoryRegister) Paste(id string) (ret Data, err error) {
	return d.data.Load().(Data), nil
}

func (d *inmemoryRegister) Copy(id string, data Data) error {
	d.data.Store(data)
	return nil
}
