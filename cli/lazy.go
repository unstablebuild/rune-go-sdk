// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2018-2024 Unstable Build, All Rights Reserved.
//
// NOTICE: All information contained herein is, and remains the property of COMPANY.
// The intellectual and technical concepts contained herein are proprietary to
// COMPANY and may be covered by U.S. and Foreign Patents, patents in process,
// and are protected by trade secret or copyright law. Dissemination of this information
// or reproduction of this material is strictly forbidden unless prior written permission
// is obtained from COMPANY. Access to the source code contained herein is hereby
// forbidden to anyone except current COMPANY employees, managers or contractors who
// have executed Confidentiality and Non-disclosure agreements explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or
// disclosure of this source code, which includes information that is confidential and/or
// proprietary, and is a trade secret, of COMPANY. ANY REPRODUCTION, MODIFICATION,
// DISTRIBUTION, PUBLIC  PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE CODE
// WITHOUT  THE EXPRESS WRITTEN CONSENT OF COMPANY IS STRICTLY PROHIBITED, AND IN
// VIOLATION OF APPLICABLE LAWS AND INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF
// THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY OR IMPLY ANY RIGHTS TO
// REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL
// ANYTHING THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package cli

import (
	"context"
	"log/slog"
)

// Lazy returns a CLI that lazily uses the given constructor to
// initialize a CLI when Run is called for the first time.
func Lazy(constructor func(context.Context) (CLI, error)) CLI {
	return &lazy{
		constructor: constructor,
	}
}

type lazy struct {
	constructor func(context.Context) (CLI, error)
	cli         CLI
}

func (l *lazy) Run(ctx context.Context, args []string) (err error) {
	if l.cli == nil {
		l.cli, err = l.constructor(ctx)
		if err != nil {
			return
		}
	}
	return l.cli.Run(ctx, args)
}

func (l *lazy) Man() Manual {
	var err error
	if l.cli == nil {
		l.cli, err = l.constructor(context.Background())
		if err != nil {
			slog.Error("man: failed to build CLI", "error", err)
			return Manual{}
		}
	}
	return l.cli.Man()
}
