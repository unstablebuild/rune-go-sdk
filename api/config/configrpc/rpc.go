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

package configrpc

import (
	"context"
	"fmt"
	sync "sync"

	"github.com/unstablebuild/rune-go-sdk/api/config"
	"google.golang.org/grpc"
)

type configServer struct {
	UnimplementedConfigServer
	cfg    config.JSON
	locker sync.Locker
}

// NewServer returns a configpb.ConfigServer that serves cfg.
func NewServer(cfg config.Config, locker sync.Locker) ConfigServer {
	ret := new(configServer)
	ret.cfg = config.JSONFromConfig(cfg)
	ret.locker = locker
	return ret
}

// Get satisfies configpb.ConfigServer.
func (s *configServer) Get(
	ctx context.Context, req *GetRequest,
) (res *GetResponse, err error) {
	s.locker.Lock()
	defer s.locker.Unlock()

	var data []byte
	data, err = s.cfg.MarshalText()
	if err != nil {
		err = fmt.Errorf("marshal config: %w", err)
		return
	}

	res = new(GetResponse)
	res.Data = string(data)
	return
}

// FetchConfig fetches a config.Config from a configpb.ConfigServer over
// the given connection.
func FetchConfig(cc grpc.ClientConnInterface) (config.Config, error) {
	client := NewConfigClient(cc)
	req := GetRequest{}
	res, err := client.Get(context.Background(), &req)
	if err != nil {
		err = fmt.Errorf("fetch config from server: %w", err)
		return nil, err
	}
	var cfg config.JSON
	err = cfg.UnmarshalText([]byte(res.GetData()))
	if err != nil {
		err = fmt.Errorf("unmarshal config from server: %w", err)
		return nil, err
	}
	return cfg, nil
}
