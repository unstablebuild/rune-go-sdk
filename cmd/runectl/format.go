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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/unstablebuild/rune-go-sdk/cli/cliformat"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

func printResult[T any](
	ctx context.Context,
	format string,
	val T,
	defaultPrint func(T),
	tableFields []string,
) error {
	if format == "" {
		defaultPrint(val)
		return nil
	}
	if format == "json" {
		return printJSONValue(val, tableFields)
	}
	wrapped := wrapForTable(val, tableFields)
	it := iterator.FromSlice(
		[]map[string]any{wrapped},
	)
	return printIterator(
		ctx, format, it, tableFields,
	)
}

func printJSONValue[T any](
	val T, tableFields []string,
) error {
	v := reflect.ValueOf(val)
	var m map[string]any
	if v.Kind() == reflect.Struct {
		raw, err := json.Marshal(val)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}
	} else {
		m = make(map[string]any)
		if len(tableFields) > 0 {
			m[tableFields[0]] = val
		}
	}
	m["success"] = true
	return json.NewEncoder(os.Stdout).Encode(m)
}

func printIterator[T any](
	ctx context.Context,
	format string,
	it iterator.Iterator[T],
	tableFields []string,
) error {
	var f cliformat.IteratorFormatter[T]
	switch format {
	case "table":
		f = cliformat.Table[T](tableFields)
	case "json":
		f = cliformat.JSON[T]()
	default:
		var err error
		f, err = cliformat.Template[T](format)
		if err != nil {
			return err
		}
	}
	return f.Format(ctx, os.Stdout, it)
}

func printString(
	format, val string, tableFields []string,
) error {
	return printResult(
		context.Background(),
		format,
		val,
		func(v string) { fmt.Println(v) },
		tableFields,
	)
}

func printOK(format string) error {
	if format == "json" {
		_, err := fmt.Println(`{"success":true}`)
		return err
	}
	fmt.Println("OK")
	return nil
}

func formatError(format string, err error) error {
	if err == nil || format != "json" {
		return err
	}
	m := map[string]any{
		"success": false,
		"error":   err.Error(),
	}
	_ = json.NewEncoder(os.Stdout).Encode(m)
	return nil
}

func wrapForTable[T any](
	val T, fields []string,
) map[string]any {
	m := make(map[string]any)
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Struct {
		for _, f := range fields {
			fv := v.FieldByName(f)
			if fv.IsValid() && fv.CanInterface() {
				m[f] = fv.Interface()
			}
		}
	} else if len(fields) > 0 {
		m[fields[0]] = val
	}
	return m
}
