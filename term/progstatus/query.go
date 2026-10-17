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

package progstatus

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	query = "?"
	// da requests primary device attributes; every terminal answers it,
	// which bounds the wait for a feature detection reply.
	da = "\x1b[c"
)

// scanQueryReply inspects terminal input received after the feature
// detection query and the primary device attributes request were sent,
// and reports which of the two replies have arrived.
func scanQueryReply(buf []byte) (reply, gotDA bool) {
	reply = bytes.Contains(buf, []byte(osc+query))
	if j := bytes.Index(buf, []byte("\x1b[?")); j >= 0 {
		gotDA = bytes.IndexByte(buf[j:], 'c') >= 0
	}
	return reply, gotDA
}

// Query reports whether the terminal on rw supports program status
// reports. It sends the feature detection query followed by a primary
// device attributes request, then reads until the device attributes
// reply arrives, so that it does not leak into the next reader of the
// terminal, or until timeout elapses. The terminal should be in raw
// mode so that the replies are readable as soon as they arrive.
func Query(rw io.ReadWriter, timeout time.Duration) (bool, error) {
	_, err := io.WriteString(rw, osc+query+st+da)
	if err != nil {
		return false, fmt.Errorf("write query: %w", err)
	}

	type chunk struct {
		data []byte
		err  error
	}
	ch := make(chan chunk)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			buf := make([]byte, 256)
			n, err := rw.Read(buf)
			select {
			case ch <- chunk{data: buf[:n], err: err}:
			case <-done:
				return
			}
			if err != nil {
				return
			}
		}
	}()

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	var buf []byte
	var supported bool
	for {
		select {
		case c := <-ch:
			buf = append(buf, c.data...)
			reply, gotDA := scanQueryReply(buf)
			supported = supported || reply
			if gotDA {
				return supported, nil
			}
			if c.err != nil {
				if errors.Is(c.err, io.EOF) {
					return supported, nil
				}
				return supported, fmt.Errorf("read reply: %w", c.err)
			}
		case <-deadline.C:
			return supported, nil
		}
	}
}
