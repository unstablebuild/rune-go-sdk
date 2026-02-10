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

package cliformat

import (
	"context"
	"fmt"
	"io"
	"text/template"

	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// Template returns an IteratorFormatter that formats elements
// according to the given Go text/template template.
// See https://pkg.go.dev/text/template for more details.
func Template[T any](tmpl string) (IteratorFormatter[T], error) {
	t, err := template.New("temp").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("invalid args: not a valid Go template: "+
			"%s. See https://pkg.go.dev/text/template", tmpl)
	}
	return templateFormatter[T]{tmpl: t}, nil
}

type templateFormatter[T any] struct {
	tmpl *template.Template
}

func (f templateFormatter[T]) Format(
	ctx context.Context, w io.Writer, it iterator.Iterator[T],
) error {
	for {
		t, ok := it.Next(ctx)
		if !ok {
			if err := it.Err(); err != nil {
				return err
			}
			break
		}
		err := f.tmpl.Execute(w, t)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(w, "\n")
	}
	return nil
}
