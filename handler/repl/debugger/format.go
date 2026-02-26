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

package debugger

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-dap"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

const sourceContextLines = 5

func formatStackTrace(
	frames []dap.StackFrame, currentIndex int,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(frames))
	for i, f := range frames {
		marker := "  "
		if i == currentIndex {
			marker = "> "
		}
		loc := formatLocation(f.Source, f.Line)
		lines[i] = fmt.Sprintf("%s%d  %s  %s", marker, i, f.Name, loc)
	}
	return stringsIter(lines)
}

func formatFrameLocation(
	frame dap.StackFrame,
) iterator.Iterator[component.Responsive] {
	loc := formatLocation(frame.Source, frame.Line)
	line := fmt.Sprintf("> %s() %s", frame.Name, loc)
	return stringIter(line)
}

func formatVariables(
	vars []dap.Variable,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(vars))
	for i, v := range vars {
		if v.Type != "" {
			lines[i] = fmt.Sprintf("%s = %s (%s)", v.Name, v.Value, v.Type)
		} else {
			lines[i] = fmt.Sprintf("%s = %s", v.Name, v.Value)
		}
	}
	return stringsIter(lines)
}

func formatBreakpointSet(
	bp dap.Breakpoint,
) iterator.Iterator[component.Responsive] {
	loc := formatLocation(bp.Source, bp.Line)
	var status string
	if bp.Verified {
		status = fmt.Sprintf("Breakpoint %d set at %s", bp.Id, loc)
	} else {
		status = fmt.Sprintf(
			"Breakpoint %d pending at %s", bp.Id, loc,
		)
		if bp.Message != "" {
			status += " (" + bp.Message + ")"
		}
	}
	return stringIter(status)
}

func formatBreakpoints(
	bps []trackedBreakpoint,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(bps))
	for i, bp := range bps {
		loc := formatLocation(bp.Source, bp.Line)
		verified := "verified"
		if !bp.Verified {
			verified = "pending"
		}
		line := fmt.Sprintf(
			"Breakpoint %d at %s [%s]", bp.Id, loc, verified,
		)
		if bp.condition != "" {
			line += fmt.Sprintf(" cond %s", bp.condition)
		}
		lines[i] = line
	}
	return stringsIter(lines)
}

func formatThreads(
	threads []dap.Thread, currentThread int,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(threads))
	for i, t := range threads {
		marker := "  "
		if t.Id == currentThread {
			marker = "* "
		}
		lines[i] = fmt.Sprintf(
			"%sThread %d: %s", marker, t.Id, t.Name,
		)
	}
	return stringsIter(lines)
}

func formatEval(
	resp *dap.EvaluateResponseBody,
) iterator.Iterator[component.Responsive] {
	if resp.Type != "" {
		return stringIter(
			fmt.Sprintf("%s (%s)", resp.Result, resp.Type),
		)
	}
	return stringIter(resp.Result)
}

func formatSource(
	content string, currentLine int,
) iterator.Iterator[component.Responsive] {
	srcLines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	start := currentLine - sourceContextLines - 1
	if start < 0 {
		start = 0
	}
	end := currentLine + sourceContextLines
	if end > len(srcLines) {
		end = len(srcLines)
	}

	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		lineNum := i + 1
		marker := "   "
		if lineNum == currentLine {
			marker = "=> "
		}
		lines = append(lines, fmt.Sprintf(
			"%s%4d: %s", marker, lineNum, srcLines[i],
		))
	}
	return stringsIter(lines)
}

func formatSources(
	sources []dap.Source,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(sources))
	for i, s := range sources {
		if s.Path != "" {
			lines[i] = s.Path
		} else {
			lines[i] = s.Name
		}
	}
	return stringsIter(lines)
}

func formatModules(
	modules []dap.Module,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(modules))
	for i, m := range modules {
		if m.Version != "" {
			lines[i] = fmt.Sprintf("%v: %s (%s)", m.Id, m.Name, m.Version)
		} else {
			lines[i] = fmt.Sprintf("%v: %s", m.Id, m.Name)
		}
	}
	return stringsIter(lines)
}

func formatDisassembly(
	instrs []dap.DisassembledInstruction, currentAddr string,
) iterator.Iterator[component.Responsive] {
	lines := make([]string, len(instrs))
	for i, inst := range instrs {
		marker := "   "
		if inst.Address == currentAddr {
			marker = "=> "
		}
		sym := ""
		if inst.Symbol != "" {
			sym = inst.Symbol + ": "
		}
		lines[i] = fmt.Sprintf(
			"%s%s %s%s", marker, inst.Address, sym, inst.Instruction,
		)
	}
	return stringsIter(lines)
}

func formatLocation(src *dap.Source, line int) string {
	if src == nil {
		return fmt.Sprintf("<unknown>:%d", line)
	}
	path := src.Path
	if path == "" {
		path = src.Name
	}
	if path == "" {
		return fmt.Sprintf("<source ref %d>:%d", src.SourceReference, line)
	}
	return fmt.Sprintf("%s:%d", filepath.Base(path), line)
}
