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
	"flag"
)

func isBoolFlag(fs *FlagSet, name string) (is bool) {
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == name {
			_, is = f.Value.(interface{ IsBoolFlag() bool })
		}
	})
	return
}

func findLastOptionIdx(fs *FlagSet, args []string) (idx int, length int) {
	isNotFlagArg := true
	for i, arg := range args {
		if arg[0] != '-' {
			if isNotFlagArg {
				idx = i
				return
			}
			isNotFlagArg = true
		} else {
			isNotFlagArg = isBoolFlag(fs, arg[1:])
		}
		length++
	}
	return
}

// Parse parses all flags in args and returns the remaining arguments
// as defined by expectedArgs. Note that options are expected to be
// passed first, then arguments: [options] <arg1> <arg2> ...
//
// Example:
//
//   var myBoolOpt
//   fs := NewFlagSet("ie", []string{"my-mandatory-arg"})
//   fs.BoolVar(&myBoolOpt, "bool-opt", false, "bool opt flag")
//
//	 fs.Parse([]string{"-bool-opt"}) // returns nil, ErrInvalidArgs
//	 fs.Parse([]string{"my-value"}) // returns []string{"my-value"}, nil
//	 fs.Parse([]string{"-bool-opt", "my-value"}) // returns []string{"my-value"}, nil
//	 fs.Parse([]string{"-bool-opt", "my-value", "other-values"}) // returns []string{"my-value"}, nil
//
// Note that ErrHelp is returned if -h or --help flags are in args.
func Parse(fs *FlagSet, expectedArgs int, args []string) (
	actualArgs []string, rest []string, err error,
) {
	if args == nil || fs == nil {
		err = ErrInvalidArgs
		return
	}

	_, optsLen := findLastOptionIdx(fs, args)

	if optsLen != 0 {
		options := args[:optsLen]
		err = fs.Parse(options)
		if err == flag.ErrHelp || fs.help {
			err = ErrHelp
		}
		if err != nil {
			return
		}
	}

	if expectedArgs > len(args)-optsLen {
		err = ErrInvalidArgs
		return
	}

	actualArgs = args[optsLen : optsLen+expectedArgs]
	rest = args[optsLen+expectedArgs:]
	return
}

// ParseUsage calls Parse and handles ErrHelp by returning false. If flags are parsed with no
// errors and usage flag is not passed, then this function returns true.
func ParseUsage(cli CLI, fs *FlagSet, expectedArgs int, args []string) (
	rargs, rest []string, ok bool, err error,
) {

	args, rest, err = Parse(fs, expectedArgs, args)
	if err != nil {
		if handleCommonErrors(cli, err) {
			err = nil
		}
		return
	}
	ok = true
	rargs = args
	return
}
