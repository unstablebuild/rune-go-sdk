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
	"flag"
	"fmt"
	"os"
)

// Manual represents a CLI's manual and documentation.
type Manual struct {
	Name string

	// Summary is a short 80-100 character description.
	Summary string

	// Synopsis is a single line synopsis of how
	// this CLI is to be used. It should ONLY include
	// the semantic information about how arguments are parsed.
	//
	// Example: [<options>] [<revision-range>] [[--] <path>...]
	Synopsis string

	// Commands is a list of accepted commands or nil
	// if no commands are expected.
	Commands []Manual

	Options FlagSet
}

// CLI represents a command-line interface.
type CLI interface {
	// Parse takes args and interprets them.
	Run(ctx context.Context, args []string) error

	Man() Manual
}

// FlagSet is just a type alias used to foster NewFlagSet usage.
type FlagSet struct {
	flag.FlagSet
	help bool
}

// Exit prints r's usage and exits the program exit code.
func Exit(r CLI, code int) {
	Usage(r)
	os.Exit(code)
}

// Run runs cli with os.Args
func Run(ctx context.Context, cli CLI) error {
	args := os.Args[1:]
	return cli.Run(ctx, args)
}

// RunCommand is a helper that attempts to find and run the next command in cmds.
func RunCommand(ctx context.Context, args []string, cmds map[string]CLI) error {
	if len(args) == 0 {
		return ErrInvalidArgs
	}
	argCmd := args[0]
	if cmd, ok := cmds[argCmd]; ok {
		return cmd.Run(ctx, args[1:])
	}

	return ErrInvalidArgs
}

func handleCommonErrors(c CLI, err error) bool {
	switch err {
	case ErrInvalidArgs:
		fmt.Printf("%s\n\n", err)
		fallthrough
	case ErrHelp:
		Usage(c)
		return true
	default:
		return false
	}
}

// ParseAndRunCommand is a helper to run CLI implementations
// that expect no arguments and simply run a sub-command CLI.
//
// Options parsed are propagated to sub-command CLI via context.Context
// and can be retrieved by using OptionFromContext.
func ParseAndRunCommand(
	ctx context.Context, c CLI, fs *FlagSet, cmds map[string]CLI, args []string,
) error {
	_, rest, err := Parse(fs, 0, args)
	if err != nil {
		if handleCommonErrors(c, err) {
			return nil
		}
		return err
	}

	ctx = ContextWithOptions(ctx, fs)
	err = RunCommand(ctx, rest, cmds)
	if err != nil {
		if handleCommonErrors(c, err) {
			return nil
		}
		return err
	}
	return nil
}

// NewFlagSet is a helper around flag.NewFlagSet that sets sane defaults.
func NewFlagSet(name string) *FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	f := &FlagSet{FlagSet: *fs}
	f.BoolVar(&f.help, "h", false, "Display this message")

	return f
}
