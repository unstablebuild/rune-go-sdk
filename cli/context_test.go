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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newParsedFlagSet(t *testing.T) *FlagSet {
	fs := NewFlagSet("beteve")
	fs.Bool("reapers", true, "Revolution")
	fs.Bool("police", true, "Repression")

	err := fs.Parse([]string{"-reapers", "-police"})
	require.NoError(t, err)

	return fs
}

func TestPopulateContext(t *testing.T) {
	ctx := context.Background()
	fs := newParsedFlagSet(t)

	ctx = ContextWithOptions(ctx, fs)

	reapers, ok1 := OptionFromContext(ctx, "reapers")
	police, ok2 := OptionFromContext(ctx, "police")

	assert.Equal(t, true, ok1)
	assert.Equal(t, true, ok2)
	assert.Equal(t, true, reapers)
	assert.Equal(t, true, police)
}

func TestDoesNotConflict(t *testing.T) {
	ctx := context.Background()
	fs := newParsedFlagSet(t)

	ctx = ContextWithOptions(ctx, fs)

	type myKey string
	var reaper myKey = "reapers"
	var sickles myKey = "sickles"
	ctx = context.WithValue(ctx, reaper, false)
	ctx = context.WithValue(ctx, sickles, true)

	reapers, _ := OptionFromContext(ctx, "reapers")
	assert.Equal(t, true, reapers)

	_, ok := OptionFromContext(ctx, "sickles")
	assert.Equal(t, false, ok)
}
