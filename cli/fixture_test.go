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
)

type testCLI struct {
	name     string
	summary  string
	synopsis string
	commands []CLI
	options  *FlagSet
}

func (t testCLI) Man() Manual {
	var subMan []Manual
	for _, cmd := range t.commands {
		subMan = append(subMan, cmd.Man())
	}
	return Manual{
		Name:     t.name,
		Summary:  t.summary,
		Synopsis: t.synopsis,
		Commands: subMan,
		Options:  *t.opts(),
	}
}

func (t testCLI) opts() *FlagSet {
	if t.options == nil {
		t.options = NewFlagSet("<default-flagset>")
	}
	return t.options
}

func (t testCLI) Run(ctx context.Context, args []string) error {
	panic("not implemented")
}

var (
	gitFlagSet       = NewFlagSet("git")
	gitLogFlagSet    = NewFlagSet("git-log")
	gitCommitFlagSet = NewFlagSet("git-commit")
)

var (
	emptyManual = testCLI{options: NewFlagSet("empty")}

	gitCommitManual = testCLI{
		name:    "commit",
		summary: "Record changes to the repository",
		options: gitLogFlagSet,
	}

	gitAddManual = testCLI{
		name:     "add",
		synopsis: "[options]",
		summary:  "Add file contents to the index",
		options:  gitCommitFlagSet,
	}

	gitManual = testCLI{
		name:     "git",
		summary:  "Git is the stupid content tracker",
		synopsis: "[options] <cmd>",
		options:  gitFlagSet,
		commands: []CLI{
			gitCommitManual,
			gitAddManual,
		},
	}
)

func init() {
	flagsets := []*FlagSet{gitFlagSet, gitLogFlagSet, gitCommitFlagSet}
	for _, fs := range flagsets {
		var b bool // discard
		fs.BoolVar(&b, "version", false, "display cli version")
	}
}
