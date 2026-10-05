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

package pkgrpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/unstablebuild/rune-go-sdk/api/pkgapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pkgapi.Manager = (*Client)(nil)

// Client satisfies pkgapi.Manager via gRPC.
type Client struct {
	pb PackagesClient
}

// NewClient returns a Client that calls the Packages service on cc.
func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{pb: NewPackagesClient(cc)}
}

// LibDir satisfies pkgapi.Manager.
func (c *Client) LibDir(
	ctx context.Context, pkgID string,
) (iterator.Iterator[string], error) {
	callCtx, cancel := context.WithCancel(ctx)
	stream, err := c.pb.LibDir(callCtx, &LibDirRequest{PkgId: pkgID})
	if err != nil {
		cancel()
		return nil, fromStatus(pkgID, err)
	}
	return &pathIterator{pkgID: pkgID, stream: stream, cancel: cancel}, nil
}

type pathIterator struct {
	pkgID  string
	stream grpc.ServerStreamingClient[LibDirResponse]
	cancel context.CancelFunc
	batch  []string
	err    error
	done   bool
}

func (it *pathIterator) Next(context.Context) (string, bool) {
	for len(it.batch) == 0 {
		if it.done {
			return "", false
		}
		resp, err := it.stream.Recv()
		if err != nil {
			it.done = true
			if !errors.Is(err, io.EOF) {
				it.err = fromStatus(it.pkgID, err)
			}
			return "", false
		}
		it.batch = resp.GetPaths()
	}
	ret := it.batch[0]
	it.batch = it.batch[1:]
	return ret, true
}

func (it *pathIterator) Err() error { return it.err }

func (it *pathIterator) Close() error {
	it.cancel()
	return nil
}

func fromStatus(pkgID string, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("%s: %w", pkgID, pkgapi.ErrNotInstalled)
	case codes.Canceled:
		return fmt.Errorf("%s: %w", st.Message(), context.Canceled)
	case codes.DeadlineExceeded:
		return fmt.Errorf("%s: %w", st.Message(), context.DeadlineExceeded)
	default:
		return err
	}
}
