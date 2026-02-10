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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedGitAddUsage = `Add file contents to the index

Usage: add [options]

Options:
  -h        Display this message [default: false]
  -version  display cli version [default: false]

`

const expectedGitUsage = `Git is the stupid content tracker

Usage: git [options] <cmd>

Options:
  -h        Display this message [default: false]
  -version  display cli version [default: false]

Commands:
  commit    Record changes to the repository
  add       Add file contents to the index

`

const expectedEmptyUsage = "Usage:  \n\nOptions:\n  -h  Display this message [default: false]\n\n"

func TestUsage(t *testing.T) {

	tsuite := []struct {
		in  testCLI
		out string
	}{
		{emptyManual, expectedEmptyUsage},
		{gitManual, expectedGitUsage},
		{gitAddManual, expectedGitAddUsage},
	}

	for _, tcase := range tsuite {
		var b bytes.Buffer
		tcase.in.options.SetOutput(&b)

		Usage(tcase.in)

		assert.Equal(t, tcase.out, b.String())
	}
}
