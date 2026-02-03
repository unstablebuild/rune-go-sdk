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


package textrpc

import (
	"context"

	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"github.com/unstablebuild/rune-go-sdk/term"
	termrpc "github.com/unstablebuild/rune-go-sdk/term/termrpc"
)

type clientWriter struct {
	uri    workspaceapi.URI
	client *Client
}

func (w clientWriter) Edit(
	ctx context.Context, start, end term.Coordinates, str string,
) (from, to term.Coordinates, old string, err error) {
	var protoStart, protoEnd termrpc.Coordinates
	protoStart.FromModel(start)
	protoEnd.FromModel(end)
	req := EditCellRequest{
		ResourceName: NewURI(w.uri),
		Start:        &protoStart,
		End:          &protoEnd,
		Str:          str,
	}
	res, err := w.client.ed.EditCell(ctx, &req)
	if err != nil {
		return from, to, "", err
	}

	from = res.GetFrom().ToModel()
	to = res.GetTo().ToModel()
	old = res.GetOld()
	return
}
