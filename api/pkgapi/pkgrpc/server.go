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

	"github.com/unstablebuild/rune-go-sdk/api/pkgapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// libDirBatch bounds how many paths one LibDir message carries, since a
// toolchain package can hold tens of thousands of files.
const libDirBatch = 512

type server struct {
	UnimplementedPackagesServer
	m pkgapi.Manager
}

// NewServer returns a PackagesServer that serves m. An error wrapping
// pkgapi.ErrNotInstalled reaches clients as pkgapi.ErrNotInstalled.
func NewServer(m pkgapi.Manager) PackagesServer {
	return &server{m: m}
}

// LibDir satisfies PackagesServer.
func (s *server) LibDir(
	req *LibDirRequest, stream grpc.ServerStreamingServer[LibDirResponse],
) error {
	ctx := stream.Context()
	it, err := s.m.LibDir(ctx, req.GetPkgId())
	if err != nil {
		return toStatus(err)
	}
	defer func() { _ = it.Close() }()

	batch := make([]string, 0, libDirBatch)
	for {
		p, ok := it.Next(ctx)
		if !ok {
			break
		}
		batch = append(batch, p)
		if len(batch) < libDirBatch {
			continue
		}
		if err := stream.Send(&LibDirResponse{Paths: batch}); err != nil {
			return err
		}
		batch = batch[:0]
	}
	if err := it.Err(); err != nil {
		return toStatus(err)
	}
	if len(batch) == 0 {
		return nil
	}
	return stream.Send(&LibDirResponse{Paths: batch})
}

func toStatus(err error) error {
	switch {
	case errors.Is(err, pkgapi.ErrNotInstalled):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return err
	}
}
