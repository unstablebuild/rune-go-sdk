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

package pkgrpc_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/pkgapi"
	"github.com/unstablebuild/rune-go-sdk/api/pkgapi/pkgrpc"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fakeManager struct {
	paths   []string
	callErr error
	iterErr error
	gotCtx  chan context.Context
}

func (f *fakeManager) LibDir(ctx context.Context, _ string) (iterator.Iterator[string], error) {
	if f.gotCtx != nil {
		f.gotCtx <- ctx
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if f.callErr != nil {
		return nil, f.callErr
	}
	if f.iterErr != nil {
		return iterator.Error[string](f.iterErr), nil
	}
	return iterator.FromSlice(f.paths), nil
}

func serve(t *testing.T, m pkgapi.Manager) *pkgrpc.Client {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	pkgrpc.RegisterPackagesServer(srv, pkgrpc.NewServer(m))
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	cc, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close() })
	return pkgrpc.NewClient(cc)
}

func TestLibDir(t *testing.T) {
	many := make([]string, 1300)
	for i := range many {
		many[i] = fmt.Sprintf("/home/me/.rune/pkg/python/1/bin/tool%d", i)
	}
	other := errors.New("disk on fire")
	tests := []struct {
		name    string
		m       *fakeManager
		want    []string
		wantErr error
	}{
		{name: "no files", m: &fakeManager{}, want: []string{}},
		{name: "some files", m: &fakeManager{paths: many[:3]}, want: many[:3]},
		{name: "more files than one message holds", m: &fakeManager{paths: many}, want: many},
		{
			name:    "not installed when called",
			m:       &fakeManager{callErr: fmt.Errorf("declined: %w", pkgapi.ErrNotInstalled)},
			wantErr: pkgapi.ErrNotInstalled,
		},
		{
			name:    "not installed after asking",
			m:       &fakeManager{iterErr: fmt.Errorf("declined: %w", pkgapi.ErrNotInstalled)},
			wantErr: pkgapi.ErrNotInstalled,
		},
		{name: "other failure", m: &fakeManager{iterErr: other}, wantErr: errAny},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := serve(t, tc.m)
			it, err := c.LibDir(t.Context(), "python")
			require.NoError(t, err)
			got, err := iterator.ToSlice(t.Context(), it)
			switch tc.wantErr {
			case nil:
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			case errAny:
				require.Error(t, err)
				assert.NotErrorIs(t, err, pkgapi.ErrNotInstalled)
			default:
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestLibDirCancelReachesServer(t *testing.T) {
	m := &fakeManager{gotCtx: make(chan context.Context, 1)}
	c := serve(t, m)
	ctx, cancel := context.WithCancel(t.Context())
	it, err := c.LibDir(ctx, "python")
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		_, err := iterator.ToSlice(t.Context(), it)
		done <- err
	}()
	srvCtx := <-m.gotCtx
	cancel()
	<-srvCtx.Done()
	require.ErrorIs(t, <-done, context.Canceled)
}

var errAny = errors.New("any error")
