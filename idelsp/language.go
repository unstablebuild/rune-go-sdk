// Unstable Build LLC ("COMPANY") CONFIDENTIAL
//
// Unpublished Copyright (c) 2017-2026 Unstable Build, All Rights Reserved.
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

package idelsp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/unstablebuild/blue/iterator"
	"github.com/unstablebuild/rune-go-sdk/api/workspaceapi"
	"unstable.build/go-tui/ide/syntax"
)

// PkgManager abstracts the ability to resolve package
// directories for a given package identifier.
type PkgManager interface {
	// LibDir returns an iterator of directory paths where
	// the package's binaries may be found.
	LibDir(ctx context.Context, pkgID string) (
		iterator.Iterator[string], error,
	)
}

type langConfig struct {
	id      string
	command string
	args    []string
}

var languages = map[string]langConfig{
	"go": {
		id:      "go",
		command: "gopls",
		//args: []string{"-mode", "stdio", "-logfile",
		//"/Users/ernestrc/.rune/gopls.log", "-debug", "127.0.0.1:3636", "-rpc.trace", "-v"},
		args: []string{"serve"},
	},
	"python": {
		id:      "python",
		command: "pyright-langserver",
		args:    []string{"--stdio"},
	},
	"typescript": {
		id:      "typescript",
		command: "typescript-language-server",
		args:    []string{"--stdio"},
	},
	"rust": {
		id:      "rust",
		command: "rust-analyzer",
		args:    nil,
	},
	"c": {
		id:      "c",
		command: "clangd",
		args:    nil,
	},
}

func languageForFile(filename workspaceapi.URI) (langConfig, error) {
	return languageForFilename(filename.Path())
}

func languageForFilename(filename string) (langConfig, error) {
	id, err := syntax.LanguageForFile(filepath.Base(filename))
	if err != nil {
		return langConfig{}, err
	}
	lang, ok := languages[id]
	if !ok {
		return langConfig{}, fmt.Errorf("%s language LSP is not supported yet", id)
	}
	return lang, nil
}
