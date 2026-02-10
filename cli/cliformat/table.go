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
	"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/unstablebuild/rune-go-sdk/api/storageapi"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

// Table returns an IteratorFormatter that formats elements
// into a table of the given fields.
func Table[T any](fields []string) IteratorFormatter[T] {
	set := make(map[string]int)
	for i, f := range fields {
		set[f] = i
	}
	return tableFormatter[T]{fields: fields, set: set}
}

type tableFormatter[T any] struct {
	fields []string
	set    map[string]int
}

func isEncodeable(t any) (reflect.Value, bool) {
	v := reflect.ValueOf(t)
	t, err := storageapi.DerefCreateValue(v)
	if err != nil {
		return reflect.Value{}, false
	}
	v = reflect.ValueOf(t)
	switch v.Kind() {
	case reflect.Map:
		return v, !v.IsNil()
	case reflect.Struct:
		return v, true
	default:
		return reflect.Value{}, false
	}
}

func (f tableFormatter[T]) Format(
	ctx context.Context, w io.Writer, it iterator.Iterator[T],
) error {
	table := tablewriter.NewWriter(w)
	table.Header(f.fields)

	for {
		t, ok := it.Next(ctx)
		if !ok {
			if err := it.Err(); err != nil {
				return err
			}
			break
		}
		v, ok := isEncodeable(t)
		if !ok {
			err := fmt.Errorf("iterator returned value that cannot be formatted: %v", t)
			return err
		}
		row := make([]string, len(f.set))
		for ff, pos := range f.set {
			var field reflect.Value
			if v.Kind() == reflect.Map {
				field = v.MapIndex(reflect.ValueOf(ff))
			} else {
				field = v.FieldByName(ff)
			}
			if field.CanInterface() {
				row[pos] = fmt.Sprintf("%v", field.Interface())
			}
		}
		if err := table.Append(row); err != nil {
			return err
		}
	}
	return table.Render()
}
