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
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

func TestTemplate(t *testing.T) {
	var ifc interface{} = testStruct1{Public: "hello"}
	tsuite := []struct {
		desc        string
		inTemplate  string
		inEls       []interface{}
		expectedOut string
		expectedErr bool
	}{
		{"empty iterator should print nothing",
			"{{ .Public }}", []interface{}{}, "", false},
		{"non-object types should error (string)",
			"{{ .Public }}", []interface{}{"one"}, "", true},
		{"non-object types should error (int)",
			"{{ .Public }}", []interface{}{1}, "", true},
		{"non-object types should error (bool)",
			"{{ .Public }}", []interface{}{true}, "", true},
		{"pass map by value",
			"{{ .Public }}", []interface{}{map[string]interface{}{"Public": "hello"}},
			"hello\n", false},
		{"pass struct by value",
			"{{ .Public }}", []interface{}{testStruct1{Public: "hello"}},
			"hello\n", false},
		{"pass struct by ref",
			"{{ .Public }}", []interface{}{&testStruct1{Public: "hello"}},
			"hello\n", false},
		{"pass struct by interface{}",
			"{{ .Public }}", []interface{}{ifc},
			"hello\n", false},
		{"preserve order of headers",
			"{{ .Public2 }} {{ .Public }}", []interface{}{
				&testStruct1{Public: "x", Public2: 0},
				testStruct1{Public: "deux", Public2: 2},
				map[string]interface{}{"Public": "trois", "Public2": "3"},
			},
			"0 x\n2 deux\n3 trois\n", false},
		{"private fields should error",
			"{{ .Public }} {{ .private }}", []interface{}{&testStruct1{private: "bla", Public: ""}},
			` `, true},
		{"struct field",
			"{{ .Composite }}", []interface{}{testStruct1{Composite: testStruct2{
				Map:      map[string]interface{}{"Bootx": "Torn ACL"},
				SliceIfc: []interface{}{"1", 1, true, false},
				SliceInt: []int{0},
			}}},
			"{map[Bootx:Torn ACL] [1 1 true false] [0]}\n", false},
	}

	for _, tcase := range tsuite {
		t.Run(tcase.desc, func(t *testing.T) {
			table, err := Template[interface{}](tcase.inTemplate)
			require.NoError(t, err)
			var buf bytes.Buffer
			it := iterator.FromSlice[any](tcase.inEls)
			err = table.Format(context.Background(), &buf, it)
			if tcase.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tcase.expectedOut, buf.String())
			}
		})
	}
}
