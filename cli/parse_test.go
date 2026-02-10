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
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const errFlagNotDefined = "flag provided but not defined"

var (
	indepeFlag bool
	anyFlag    int
)

func emptyFlagSet() *FlagSet {
	return NewFlagSet("Sad Panda")
}

func nonEmptyFlagSet() *FlagSet {
	fs := NewFlagSet("Els Segadors")
	fs.BoolVar(&indepeFlag, "independencia", true, "Ni Oblit, Ni Perdo.")
	fs.IntVar(&anyFlag, "any", 2019, "DUI Ja.")
	return fs
}

func requireError(t *testing.T, err error, expected interface{}) {
	switch expected.(type) {
	case string:
		require.NotNil(t, err)
		require.Contains(t, err.Error(), expected)
	default:
		require.Equal(t, expected, err)
	}
}

func recoverParsePanic(t *testing.T, testCase int) {
	err := recover()
	if err != nil {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)
		t.Fatalf("test case %d panic:%s\n%s\n", testCase, err, buf)
	}
}

func TestParse(t *testing.T) {
	tsuite := []struct {
		// input
		fs      *FlagSet
		expArgs int
		args    []string
		// output
		expectedArgs []string
		expectedRest []string
		expectedErr  interface{}
	}{
		{emptyFlagSet(), 0, nil, nil, nil, ErrInvalidArgs},
		{emptyFlagSet(), 0, []string{}, []string{}, []string{}, nil},
		{emptyFlagSet(), 1, []string{}, nil, nil, ErrInvalidArgs},
		{emptyFlagSet(), 1, []string{"a"}, []string{"a"}, []string{}, nil},
		{emptyFlagSet(), 2, []string{"a"}, nil, nil, ErrInvalidArgs},
		{emptyFlagSet(), 2, []string{"a", "b", "c"}, []string{"a", "b"}, []string{"c"}, nil},
		{emptyFlagSet(), 0, []string{"-opt", "a"}, nil, nil, errFlagNotDefined},
		{nonEmptyFlagSet(), 0, []string{"-opt", "a"}, nil, nil, errFlagNotDefined},
		{nonEmptyFlagSet(), 0, []string{"-independencia", "a"}, []string{}, []string{"a"}, nil},
		{nonEmptyFlagSet(), 2, []string{"a", "-independencia"}, []string{"a", "-independencia"}, []string{}, nil},
		{nonEmptyFlagSet(), 0, []string{"a", "-independencia"}, []string{}, []string{"a", "-independencia"}, nil},
		{nonEmptyFlagSet(), 0, []string{"-any", "2017", "a"}, []string{}, []string{"a"}, nil},
		{nonEmptyFlagSet(), 1, []string{"-independencia", "-any", "2017", "a", "-opt", "a"}, []string{"a"}, []string{"-opt", "a"}, nil},
		{nonEmptyFlagSet(), 1, []string{"-independencia", "-any", "2017", "-opt", "a"}, nil, nil, errFlagNotDefined},
	}

	for i, _tcase := range tsuite {
		tcase := _tcase
		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			defer recoverParsePanic(t, i)

			args, rest, err := Parse(tcase.fs, tcase.expArgs, tcase.args)

			requireError(t, err, tcase.expectedErr)
			assert.Equal(t, tcase.expectedArgs, args)
			assert.Equal(t, tcase.expectedRest, rest)
		})
	}
}

func TestParseOptions(t *testing.T) {
	defer recoverParsePanic(t, 0)
	indepeFlag = false
	anyFlag = 0

	_, _, err := Parse(nonEmptyFlagSet(), 0, []string{"-independencia", "-any", "1989"})
	requireError(t, err, nil)
	assert.Equal(t, true, indepeFlag)
	assert.Equal(t, 1989, anyFlag)
}
