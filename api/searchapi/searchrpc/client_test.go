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

package searchrpc_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/api/llmapi"
	"github.com/unstablebuild/rune-go-sdk/api/searchapi"
	"github.com/unstablebuild/rune-go-sdk/api/searchapi/searchrpc"
	"github.com/unstablebuild/rune-go-sdk/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// stubServer streams its results, regexp lines in the messages given and
// semantic results after a status message unless noStatus, then ends the
// stream with err. With block set, each search hands its context over and
// waits for it to end instead of streaming results.
type stubServer struct {
	searchrpc.UnimplementedSearcherServer
	regexp   [][]searchapi.RegexpResult
	semantic []searchapi.SemanticResult
	status   searchapi.SemanticStatus
	noStatus bool
	combined bool
	err      error
	block    chan context.Context
	onWait   func()

	gotRegexp   searchapi.RegexpRequest
	gotSemantic searchapi.SemanticRequest
}

func (s *stubServer) RegexpSearch(
	req *searchrpc.RegexpRequest, stream grpc.ServerStreamingServer[searchrpc.RegexpResponse],
) error {
	s.gotRegexp = searchrpc.FromProtoRegexpRequest(req)
	if s.block != nil {
		return s.wait(stream.Context())
	}
	for _, msg := range s.regexp {
		resp := &searchrpc.RegexpResponse{}
		for _, r := range msg {
			resp.Results = append(resp.Results, searchrpc.ToProtoRegexpResult(r))
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	return s.err
}

func (s *stubServer) SemanticSearch(
	req *searchrpc.SemanticRequest, stream grpc.ServerStreamingServer[searchrpc.SemanticResponse],
) error {
	s.gotSemantic = searchrpc.FromProtoSemanticRequest(req)
	results := s.semantic
	if !s.noStatus {
		resp := &searchrpc.SemanticResponse{Status: searchrpc.ToProtoSemanticStatus(s.status)}
		if s.combined && len(results) > 0 {
			resp.Result = searchrpc.ToProtoSemanticResult(results[0])
			results = results[1:]
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	if s.block != nil {
		return s.wait(stream.Context())
	}
	for _, r := range results {
		if err := stream.Send(&searchrpc.SemanticResponse{Result: searchrpc.ToProtoSemanticResult(r)}); err != nil {
			return err
		}
	}
	return s.err
}

func (s *stubServer) wait(ctx context.Context) error {
	s.block <- ctx
	if s.onWait != nil {
		s.onWait()
	}
	<-ctx.Done()
	return ctx.Err()
}

func dial(t *testing.T, s *stubServer) *searchrpc.Client {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	searchrpc.RegisterSearcherServer(srv, s)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	cc, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close() })
	return searchrpc.NewClient(cc)
}

func TestRegexpSearch(t *testing.T) {
	a := searchapi.RegexpResult{Path: "a.go", Line: 3, Text: "retry(x) retry(y)", Spans: [][2]int{{0, 6}, {9, 15}}}
	b := searchapi.RegexpResult{Path: "dir/b.go", Line: 7, Text: ""}
	c := searchapi.RegexpResult{Path: "dir/c.go", Line: 1, Text: "retry(", Spans: [][2]int{{0, 6}}}
	req := searchapi.RegexpRequest{
		Pattern: `retry\(`, IgnoreCase: true, Roots: []string{"pkg", "/w/cmd"},
		Include: []string{"*.go", "*.md"}, Limit: 50, MaxPerFile: 1,
	}
	tests := []struct {
		name    string
		s       *stubServer
		want    []searchapi.RegexpResult
		wantErr error
	}{
		{name: "no results", s: &stubServer{}, want: []searchapi.RegexpResult{}},
		{
			name: "lines with and without spans",
			s:    &stubServer{regexp: [][]searchapi.RegexpResult{{a, b}}},
			want: []searchapi.RegexpResult{a, b},
		},
		{
			name: "lines across messages",
			s:    &stubServer{regexp: [][]searchapi.RegexpResult{{a}, {}, {b, c}}},
			want: []searchapi.RegexpResult{a, b, c},
		},
		{
			name: "lines that are not UTF-8",
			s: &stubServer{regexp: [][]searchapi.RegexpResult{{
				{Path: "latin1/caf\xe9.txt", Line: 1, Text: "caf\xe9 retry(", Spans: [][2]int{{5, 11}}},
			}}},
			want: []searchapi.RegexpResult{
				{Path: "latin1/caf\xe9.txt", Line: 1, Text: "caf\xe9 retry(", Spans: [][2]int{{5, 11}}},
			},
		},
		{
			name:    "unimplemented is unsupported",
			s:       &stubServer{err: status.Error(codes.Unimplemented, "no index")},
			wantErr: searchapi.ErrUnsupported,
		},
		{
			name: "unimplemented after lines",
			s: &stubServer{
				regexp: [][]searchapi.RegexpResult{{a}},
				err:    status.Error(codes.Unimplemented, "no index"),
			},
			wantErr: searchapi.ErrUnsupported,
		},
		{
			name:    "deadline exceeded",
			s:       &stubServer{err: status.Error(codes.DeadlineExceeded, "walk")},
			wantErr: context.DeadlineExceeded,
		},
		{name: "other failure", s: &stubServer{err: errors.New("bad pattern")}, wantErr: errAny},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := dial(t, tc.s)
			got, err := search(t, func() (iterator.Iterator[searchapi.RegexpResult], error) {
				return c.RegexpSearch(t.Context(), req)
			})
			checkErr(t, tc.wantErr, err)
			if tc.wantErr == nil {
				assert.Equal(t, tc.want, got)
			}
			assert.Equal(t, req, tc.s.gotRegexp)
		})
	}
}

func TestSemanticSearch(t *testing.T) {
	results := []searchapi.SemanticResult{
		{
			Path: "internal/retry.go", Kind: "code", Score: 3.5, Z: 2.25,
			Cosine: 0.61, KeywordRank: 1, Segment: "lines", Loc: 10,
			LocEnd: 42, Focus: 12, Excerpt: "func «parseRetryBackoff»(s string)",
			Modified: 1700000000, Size: 2048, Text: "\tfunc parseRetryBackoff(s string) {",
		},
		{Path: "README.md", Kind: "doc", Score: 1, Segment: "name"},
	}
	req := searchapi.SemanticRequest{
		Query: "retry backoff", Model: llmapi.ModelEntry{Provider: "llamacpp", Name: "gemma"},
		Roots: []string{"internal"}, Limit: 5, Meaning: true, PrefixLast: true,
	}
	indexing := searchapi.SemanticStatus{Indexing: true, WordsOnly: true, Message: "12 files left to index"}
	tests := []struct {
		name       string
		s          *stubServer
		want       []searchapi.SemanticResult
		wantStatus searchapi.SemanticStatus
		wantErr    error
	}{
		{name: "results round-trip", s: &stubServer{semantic: results}, want: results},
		{name: "no results", s: &stubServer{}, want: []searchapi.SemanticResult{}},
		{
			name:       "status round-trips",
			s:          &stubServer{semantic: results, status: indexing},
			want:       results,
			wantStatus: indexing,
		},
		{
			name:       "status with first result",
			s:          &stubServer{semantic: results, status: indexing, combined: true},
			want:       results,
			wantStatus: indexing,
		},
		{
			name:       "status without results",
			s:          &stubServer{status: indexing},
			want:       []searchapi.SemanticResult{},
			wantStatus: indexing,
		},
		{
			name: "results without a status",
			s:    &stubServer{semantic: results, noStatus: true},
			want: results,
		},
		{
			name: "neither results nor status",
			s:    &stubServer{noStatus: true},
			want: []searchapi.SemanticResult{},
		},
		{
			name:    "unimplemented is unsupported",
			s:       &stubServer{err: status.Error(codes.Unimplemented, "openai")},
			wantErr: searchapi.ErrUnsupported,
		},
		{
			name:    "unimplemented before the status",
			s:       &stubServer{noStatus: true, err: status.Error(codes.Unimplemented, "no index")},
			wantErr: searchapi.ErrUnsupported,
		},
		{
			name:    "not found is an unknown model",
			s:       &stubServer{err: status.Error(codes.NotFound, "gemma")},
			wantErr: llmapi.ErrModelNotFound,
		},
		{
			name:    "canceled after results",
			s:       &stubServer{semantic: results, err: status.Error(codes.Canceled, "embed")},
			wantErr: context.Canceled,
		},
		{name: "other failure", s: &stubServer{err: errors.New("index broke")}, wantErr: errAny},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := dial(t, tc.s)
			var st searchapi.SemanticStatus
			got, err := search(t, func() (iterator.Iterator[searchapi.SemanticResult], error) {
				it, s, err := c.SemanticSearch(t.Context(), req)
				st = s
				return it, err
			})
			checkErr(t, tc.wantErr, err)
			if tc.wantErr == nil {
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.wantStatus, st)
			}
			assert.Equal(t, req, tc.s.gotSemantic)
		})
	}
}

func TestSemanticSearchCancellationWaitingForFirstResponse(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	s := &stubServer{noStatus: true, block: make(chan context.Context, 1), onWait: cancel}
	c := dial(t, s)
	it, st, err := c.SemanticSearch(ctx, searchapi.SemanticRequest{Query: "x"})
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, it)
	assert.Equal(t, searchapi.SemanticStatus{}, st)
	select {
	case srvCtx := <-s.block:
		select {
		case <-srvCtx.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("server stream was not canceled")
		}
	default:
		t.Fatal("search did not reach the server")
	}
}

func TestSemanticSearchCloseBeforeResults(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	s := &stubServer{block: make(chan context.Context, 1), status: searchapi.SemanticStatus{Indexing: true}}
	it, st, err := dial(t, s).SemanticSearch(ctx, searchapi.SemanticRequest{Query: "x"})
	require.NoError(t, err)
	assert.Equal(t, s.status, st)
	t.Cleanup(func() { assert.NoError(t, it.Close()) })
	require.NoError(t, it.Close())
	require.NoError(t, it.Close())
	_, ok := it.Next(ctx)
	assert.False(t, ok)
	require.ErrorIs(t, it.Err(), context.Canceled)
	select {
	case srvCtx := <-s.block:
		select {
		case <-srvCtx.Done():
		case <-ctx.Done():
			t.Fatal("server stream was not canceled by Close")
		}
	case <-ctx.Done():
		t.Fatal("server did not enter its wait")
	}
}

func TestSemanticSearchErrors(t *testing.T) {
	for _, tc := range []struct {
		code codes.Code
		want error
	}{
		{codes.Unimplemented, searchapi.ErrUnsupported},
		{codes.NotFound, llmapi.ErrModelNotFound},
		{codes.Canceled, context.Canceled},
		{codes.DeadlineExceeded, context.DeadlineExceeded},
		{codes.Internal, errAny},
	} {
		for _, phase := range []string{"before response", "after status", "after result"} {
			t.Run(tc.code.String()+"/"+phase, func(t *testing.T) {
				st := searchapi.SemanticStatus{Indexing: true, WordsOnly: true, Message: "indexing"}
				s := &stubServer{status: st, noStatus: phase == "before response", err: status.Error(tc.code, "search failed")}
				result := searchapi.SemanticResult{Path: "a.go", Score: 1}
				if phase == "after result" {
					s.semantic = []searchapi.SemanticResult{result}
					s.combined = true
				}
				it, gotStatus, err := dial(t, s).SemanticSearch(t.Context(), searchapi.SemanticRequest{Query: "x"})
				if phase == "before response" {
					checkErr(t, tc.want, err)
					assert.Nil(t, it)
					assert.Equal(t, searchapi.SemanticStatus{}, gotStatus)
					return
				}
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, it.Close()) })
				assert.Equal(t, st, gotStatus)
				require.NoError(t, it.Err())
				if phase == "after result" {
					got, ok := it.Next(t.Context())
					require.True(t, ok)
					assert.Equal(t, result, got)
					require.NoError(t, it.Err())
				}
				for range 2 {
					got, ok := it.Next(t.Context())
					assert.False(t, ok)
					assert.Equal(t, searchapi.SemanticResult{}, got)
					checkErr(t, tc.want, it.Err())
					assert.Contains(t, it.Err().Error(), "search failed")
					if tc.want == errAny {
						assert.Equal(t, tc.code, status.Code(it.Err()))
					}
				}
				require.NoError(t, it.Close())
				checkErr(t, tc.want, it.Err())
			})
		}
	}
}

func TestSemanticSearchModelTravelsByProviderAndName(t *testing.T) {
	s := &stubServer{}
	c := dial(t, s)
	_, err := search(t, func() (iterator.Iterator[searchapi.SemanticResult], error) {
		it, _, err := c.SemanticSearch(t.Context(), searchapi.SemanticRequest{
			Query: "q",
			Model: llmapi.ModelEntry{Provider: "llamacpp", Name: "gemma", ContextWindow: 2048, BaseURL: "http://x"},
		})
		return it, err
	})
	require.NoError(t, err)
	assert.Equal(t, llmapi.ModelEntry{Provider: "llamacpp", Name: "gemma"}, s.gotSemantic.Model)
}

func TestCloseCancelsTheStream(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *searchrpc.Client) (func() error, error)
	}{
		{"regexp", func(ctx context.Context, c *searchrpc.Client) (func() error, error) {
			it, err := c.RegexpSearch(ctx, searchapi.RegexpRequest{Pattern: "x"})
			if err != nil {
				return nil, err
			}
			return it.Close, nil
		}},
		{"semantic", func(ctx context.Context, c *searchrpc.Client) (func() error, error) {
			it, _, err := c.SemanticSearch(ctx, searchapi.SemanticRequest{Query: "x"})
			if err != nil {
				return nil, err
			}
			return it.Close, nil
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &stubServer{block: make(chan context.Context, 1)}
			c := dial(t, s)
			closeIt, err := tc.call(t.Context(), c)
			require.NoError(t, err)
			srvCtx := <-s.block
			require.NoError(t, closeIt())
			<-srvCtx.Done()
		})
	}
}

func search[T any](t *testing.T, call func() (iterator.Iterator[T], error)) ([]T, error) {
	t.Helper()
	it, err := call()
	if err != nil {
		return nil, err
	}
	return iterator.ToSlice(t.Context(), it)
}

func checkErr(t *testing.T, want, got error) {
	t.Helper()
	switch want {
	case nil:
		require.NoError(t, got)
	case errAny:
		require.Error(t, got)
		for _, mapped := range []error{
			searchapi.ErrUnsupported, llmapi.ErrModelNotFound,
			context.Canceled, context.DeadlineExceeded,
		} {
			assert.NotErrorIs(t, got, mapped)
		}
	default:
		require.ErrorIs(t, got, want)
	}
}

var errAny = errors.New("any error")
