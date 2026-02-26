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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-dap"
	"github.com/unstablebuild/rune-go-sdk/component"
	"github.com/unstablebuild/rune-go-sdk/iterator"
)

type commandFn func(ctx context.Context, args []string) (
	iterator.Iterator[component.Responsive], error,
)

type command struct {
	name    string
	aliases []string
	fn      commandFn
	short   string
	usage   string
}

func errUnknownCommand(name string) error {
	return fmt.Errorf("unknown command %q — type \"help\" for a list of commands", name)
}

func (h *Handler) buildCommands() []command {
	return []command{
		{name: "continue", aliases: []string{"c"}, fn: h.cmdContinue,
			short: "Resume execution", usage: "continue"},
		{name: "next", aliases: []string{"n"}, fn: h.cmdNext,
			short: "Step over to next source line", usage: "next"},
		{name: "step", aliases: []string{"s"}, fn: h.cmdStep,
			short: "Step into function call", usage: "step"},
		{name: "stepout", aliases: []string{"so"}, fn: h.cmdStepOut,
			short: "Step out of current function", usage: "stepout"},
		{name: "pause", fn: h.cmdPause,
			short: "Suspend execution", usage: "pause"},
		{name: "restart", aliases: []string{"r"}, fn: h.cmdRestart,
			short: "Restart debug session", usage: "restart"},

		{name: "break", aliases: []string{"b"}, fn: h.cmdBreak,
			short: "Set a breakpoint", usage: "break <file>:<line> [condition]"},
		{name: "breakpoints", aliases: []string{"bp"}, fn: h.cmdBreakpoints,
			short: "List all breakpoints", usage: "breakpoints"},
		{name: "clear", fn: h.cmdClear,
			short: "Remove a breakpoint", usage: "clear <id>"},
		{name: "clearall", fn: h.cmdClearAll,
			short: "Remove all breakpoints", usage: "clearall"},
		{name: "condition", aliases: []string{"cond"}, fn: h.cmdCondition,
			short: "Set breakpoint condition", usage: "condition <id> <expr>"},

		{name: "print", aliases: []string{"p"}, fn: h.cmdPrint,
			short: "Evaluate an expression", usage: "print <expression>"},
		{name: "set", fn: h.cmdSet,
			short: "Set a variable's value", usage: "set <var> = <value>"},
		{name: "locals", fn: h.cmdLocals,
			short: "Show local variables", usage: "locals"},
		{name: "args", fn: h.cmdArgs,
			short: "Show function arguments", usage: "args"},
		{name: "whatis", fn: h.cmdWhatis,
			short: "Show type of expression", usage: "whatis <expression>"},

		{name: "threads", fn: h.cmdThreads,
			short: "List threads", usage: "threads"},
		{name: "thread", fn: h.cmdThread,
			short: "Switch to a thread", usage: "thread <id>"},

		{name: "stack", aliases: []string{"bt"}, fn: h.cmdStack,
			short: "Print stack trace", usage: "stack [depth]"},
		{name: "frame", fn: h.cmdFrame,
			short: "Select a stack frame", usage: "frame <n>"},
		{name: "up", fn: h.cmdUp,
			short: "Move up the stack", usage: "up [n]"},
		{name: "down", fn: h.cmdDown,
			short: "Move down the stack", usage: "down [n]"},

		{name: "list", aliases: []string{"l"}, fn: h.cmdList,
			short: "Show source around current position", usage: "list"},
		{name: "sources", fn: h.cmdSources,
			short: "List loaded sources", usage: "sources"},
		{name: "modules", fn: h.cmdModules,
			short: "List loaded modules", usage: "modules"},
		{name: "disassemble", aliases: []string{"disass"}, fn: h.cmdDisassemble,
			short: "Disassemble at current position", usage: "disassemble"},

		{name: "help", aliases: []string{"h"}, fn: h.cmdHelp,
			short: "Show help", usage: "help [command]"},
	}
}

// --- Execution commands ---

func (h *Handler) cmdContinue(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	_, err := h.dbg.Continue(ctx, &dap.ContinueArguments{
		ThreadId: h.threadID,
	})
	if err != nil {
		return nil, err
	}
	return stringIter("continuing"), nil
}

func (h *Handler) cmdNext(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	err := h.dbg.Next(ctx, &dap.NextArguments{
		ThreadId: h.threadID,
	})
	if err != nil {
		return nil, err
	}
	return h.showCurrentLocation(ctx)
}

func (h *Handler) cmdStep(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	err := h.dbg.StepIn(ctx, &dap.StepInArguments{
		ThreadId: h.threadID,
	})
	if err != nil {
		return nil, err
	}
	return h.showCurrentLocation(ctx)
}

func (h *Handler) cmdStepOut(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	err := h.dbg.StepOut(ctx, &dap.StepOutArguments{
		ThreadId: h.threadID,
	})
	if err != nil {
		return nil, err
	}
	return h.showCurrentLocation(ctx)
}

func (h *Handler) cmdPause(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	err := h.dbg.Pause(ctx, &dap.PauseArguments{
		ThreadId: h.threadID,
	})
	if err != nil {
		return nil, err
	}
	return stringIter("paused"), nil
}

func (h *Handler) cmdRestart(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	err := h.dbg.Restart(ctx)
	if err != nil {
		return nil, err
	}
	h.frameIndex = 0
	h.frameID = 0
	return stringIter("restarted"), nil
}

// --- Breakpoint commands ---

func (h *Handler) cmdBreak(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: break <file>:<line> [condition]")
	}
	file, line, err := parseFileLocation(args[0])
	if err != nil {
		return nil, err
	}

	var condition string
	if len(args) > 1 {
		condition = strings.Join(args[1:], " ")
	}

	existing := h.breakpoints[file]
	bps := make([]dap.SourceBreakpoint, 0, len(existing)+1)
	for _, bp := range existing {
		bps = append(bps, dap.SourceBreakpoint{
			Line:      bp.Line,
			Condition: bp.condition,
		})
	}
	bps = append(bps, dap.SourceBreakpoint{
		Line:      line,
		Condition: condition,
	})

	result, err := h.dbg.SetBreakpoints(ctx, &dap.SetBreakpointsArguments{
		Source:      dap.Source{Path: file},
		Breakpoints: bps,
	})
	if err != nil {
		return nil, err
	}

	// Rebuild tracked breakpoints pairing DAP results with conditions.
	conditions := make([]string, len(bps))
	for i, bp := range bps {
		conditions[i] = bp.Condition
	}
	h.breakpoints[file] = toTracked(result, conditions)

	if len(result) > 0 {
		bp := result[len(result)-1]
		return formatBreakpointSet(bp), nil
	}
	return stringIter("breakpoint set"), nil
}

func (h *Handler) cmdBreakpoints(
	_ context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	var all []trackedBreakpoint
	for _, bps := range h.breakpoints {
		all = append(all, bps...)
	}
	if len(all) == 0 {
		return stringIter("No breakpoints set."), nil
	}
	return formatBreakpoints(all), nil
}

func (h *Handler) cmdClear(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: clear <id>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid breakpoint id: %s", args[0])
	}

	for file, bps := range h.breakpoints {
		for i, bp := range bps {
			if bp.Id == id {
				remaining := make(
					[]trackedBreakpoint, 0, len(bps)-1,
				)
				remaining = append(remaining, bps[:i]...)
				remaining = append(remaining, bps[i+1:]...)
				return h.setFileBreakpoints(ctx, file, remaining)
			}
		}
	}
	return nil, fmt.Errorf("breakpoint %d not found", id)
}

func (h *Handler) cmdClearAll(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	for file := range h.breakpoints {
		iter, err := h.setFileBreakpoints(ctx, file, nil)
		if err != nil {
			return nil, err
		}
		_ = iter.Close()
	}
	return stringIter("All breakpoints cleared."), nil
}

func (h *Handler) cmdCondition(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("usage: condition <id> <expr>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid breakpoint id: %s", args[0])
	}
	expr := strings.Join(args[1:], " ")

	for file, bps := range h.breakpoints {
		for i, bp := range bps {
			if bp.Id == id {
				bps[i].condition = expr
				return h.setFileBreakpoints(ctx, file, bps)
			}
		}
	}
	return nil, fmt.Errorf("breakpoint %d not found", id)
}

func (h *Handler) setFileBreakpoints(
	ctx context.Context, file string, bps []trackedBreakpoint,
) (iterator.Iterator[component.Responsive], error) {
	sbps := make([]dap.SourceBreakpoint, len(bps))
	conditions := make([]string, len(bps))
	for i, bp := range bps {
		sbps[i] = dap.SourceBreakpoint{
			Line:      bp.Line,
			Condition: bp.condition,
		}
		conditions[i] = bp.condition
	}

	result, err := h.dbg.SetBreakpoints(ctx, &dap.SetBreakpointsArguments{
		Source:      dap.Source{Path: file},
		Breakpoints: sbps,
	})
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		delete(h.breakpoints, file)
	} else {
		h.breakpoints[file] = toTracked(result, conditions)
	}
	return stringIter("breakpoint updated"), nil
}

func toTracked(
	bps []dap.Breakpoint, conditions []string,
) []trackedBreakpoint {
	tracked := make([]trackedBreakpoint, len(bps))
	for i, bp := range bps {
		var cond string
		if i < len(conditions) {
			cond = conditions[i]
		}
		tracked[i] = trackedBreakpoint{
			Breakpoint: bp,
			condition:  cond,
		}
	}
	return tracked
}

// --- Inspection commands ---

func (h *Handler) cmdPrint(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: print <expression>")
	}
	expr := strings.Join(args, " ")
	resp, err := h.dbg.Evaluate(ctx, &dap.EvaluateArguments{
		Expression: expr,
		FrameId:    h.frameID,
		Context:    "repl",
	})
	if err != nil {
		return nil, err
	}
	return formatEval(resp), nil
}

func (h *Handler) cmdSet(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	// Parse "var = value" or "var=value"
	full := strings.Join(args, " ")
	parts := strings.SplitN(full, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("usage: set <var> = <value>")
	}
	name := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	scopes, err := h.dbg.Scopes(ctx, &dap.ScopesArguments{
		FrameId: h.frameID,
	})
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, scope := range scopes {
		resp, err := h.dbg.SetVariable(ctx, &dap.SetVariableArguments{
			VariablesReference: scope.VariablesReference,
			Name:               name,
			Value:              value,
		})
		if err != nil {
			lastErr = err
			continue
		}
		return stringIter(fmt.Sprintf("%s = %s", name, resp.Value)), nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("variable %q not found in any scope", name)
}

func (h *Handler) cmdLocals(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	return h.scopeVariables(ctx, "Locals", "Local")
}

func (h *Handler) cmdArgs(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	return h.scopeVariables(ctx, "Arguments", "Argument")
}

func (h *Handler) scopeVariables(
	ctx context.Context, names ...string,
) (iterator.Iterator[component.Responsive], error) {
	scopes, err := h.dbg.Scopes(ctx, &dap.ScopesArguments{
		FrameId: h.frameID,
	})
	if err != nil {
		return nil, err
	}

	for _, scope := range scopes {
		for _, name := range names {
			if !strings.EqualFold(scope.Name, name) {
				continue
			}
			vars, err := h.dbg.Variables(ctx, &dap.VariablesArguments{
				VariablesReference: scope.VariablesReference,
			})
			if err != nil {
				return nil, err
			}
			if len(vars) == 0 {
				return stringIter("(no " + strings.ToLower(name) + "s)"), nil
			}
			return formatVariables(vars), nil
		}
	}
	return stringIter("(no matching scope)"), nil
}

func (h *Handler) cmdWhatis(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: whatis <expression>")
	}
	expr := strings.Join(args, " ")
	resp, err := h.dbg.Evaluate(ctx, &dap.EvaluateArguments{
		Expression: expr,
		FrameId:    h.frameID,
		Context:    "hover",
	})
	if err != nil {
		return nil, err
	}
	if resp.Type != "" {
		return stringIter(resp.Type), nil
	}
	return stringIter(resp.Result), nil
}

// --- Thread commands ---

func (h *Handler) cmdThreads(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	threads, err := h.dbg.Threads(ctx)
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return stringIter("No threads."), nil
	}
	return formatThreads(threads, h.threadID), nil
}

func (h *Handler) cmdThread(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: thread <id>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid thread id: %s", args[0])
	}
	h.threadID = id
	h.frameIndex = 0
	h.frameID = 0
	return h.showCurrentLocation(ctx)
}

// --- Stack commands ---

func (h *Handler) cmdStack(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	depth := 50
	if len(args) > 0 {
		d, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid depth: %s", args[0])
		}
		if d < 1 {
			return nil, fmt.Errorf("depth must be positive")
		}
		depth = d
	}

	resp, err := h.dbg.StackTrace(ctx, &dap.StackTraceArguments{
		ThreadId: h.threadID,
		Levels:   depth,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.StackFrames) == 0 {
		return stringIter("No stack frames."), nil
	}
	return formatStackTrace(resp.StackFrames, h.frameIndex), nil
}

func (h *Handler) cmdFrame(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("usage: frame <n>")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid frame number: %s", args[0])
	}
	return h.selectFrame(ctx, n)
}

func (h *Handler) cmdUp(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	n := 1
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid count: %s", args[0])
		}
		if parsed < 1 {
			return nil, fmt.Errorf("count must be positive")
		}
		n = parsed
	}
	return h.selectFrame(ctx, h.frameIndex+n)
}

func (h *Handler) cmdDown(
	ctx context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	n := 1
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("invalid count: %s", args[0])
		}
		if parsed < 1 {
			return nil, fmt.Errorf("count must be positive")
		}
		n = parsed
	}
	target := h.frameIndex - n
	if target < 0 {
		target = 0
	}
	return h.selectFrame(ctx, target)
}

func (h *Handler) selectFrame(
	ctx context.Context, index int,
) (iterator.Iterator[component.Responsive], error) {
	resp, err := h.dbg.StackTrace(ctx, &dap.StackTraceArguments{
		ThreadId: h.threadID,
		Levels:   index + 1,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.StackFrames) == 0 {
		return nil, fmt.Errorf("no stack frames available")
	}
	if index >= len(resp.StackFrames) {
		return nil, fmt.Errorf(
			"frame %d out of range (0–%d)",
			index, len(resp.StackFrames)-1,
		)
	}
	h.frameIndex = index
	h.frameID = resp.StackFrames[index].Id
	return formatFrameLocation(resp.StackFrames[index]), nil
}

// --- Source commands ---

func (h *Handler) cmdList(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	resp, err := h.dbg.StackTrace(ctx, &dap.StackTraceArguments{
		ThreadId:   h.threadID,
		StartFrame: h.frameIndex,
		Levels:     1,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.StackFrames) == 0 {
		return stringIter("No source available."), nil
	}
	frame := resp.StackFrames[0]
	if frame.Source == nil {
		return stringIter("No source available."), nil
	}

	src, err := h.dbg.Source(ctx, &dap.SourceArguments{
		Source:          frame.Source,
		SourceReference: frame.Source.SourceReference,
	})
	if err != nil {
		return nil, fmt.Errorf("source: %w", err)
	}
	return formatSource(src.Content, frame.Line), nil
}

func (h *Handler) cmdSources(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	sources, err := h.dbg.LoadedSources(ctx)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return stringIter("No loaded sources."), nil
	}
	return formatSources(sources), nil
}

func (h *Handler) cmdModules(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	resp, err := h.dbg.Modules(ctx, &dap.ModulesArguments{})
	if err != nil {
		return nil, err
	}
	if len(resp.Modules) == 0 {
		return stringIter("No modules loaded."), nil
	}
	return formatModules(resp.Modules), nil
}

func (h *Handler) cmdDisassemble(
	ctx context.Context, _ []string,
) (iterator.Iterator[component.Responsive], error) {
	resp, err := h.dbg.StackTrace(ctx, &dap.StackTraceArguments{
		ThreadId:   h.threadID,
		StartFrame: h.frameIndex,
		Levels:     1,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.StackFrames) == 0 {
		return stringIter("No frame available."), nil
	}
	frame := resp.StackFrames[0]
	ref := frame.InstructionPointerReference
	if ref == "" {
		return stringIter("No instruction pointer available."), nil
	}

	// Show 10 instructions before and after the current address.
	const disasmContextInstrs = 10
	instrs, err := h.dbg.Disassemble(ctx, &dap.DisassembleArguments{
		MemoryReference:   ref,
		InstructionOffset: -disasmContextInstrs,
		InstructionCount:  2*disasmContextInstrs + 1,
	})
	if err != nil {
		return nil, err
	}
	if len(instrs) == 0 {
		return stringIter("No disassembly available."), nil
	}
	return formatDisassembly(instrs, ref), nil
}

// --- Misc commands ---

func (h *Handler) cmdHelp(
	_ context.Context, args []string,
) (iterator.Iterator[component.Responsive], error) {
	if len(args) > 0 {
		c, ok := h.findCommand(args[0])
		if !ok {
			return nil, errUnknownCommand(args[0])
		}
		lines := []string{
			c.usage,
			c.short,
		}
		if len(c.aliases) > 0 {
			lines = append(lines,
				"Aliases: "+strings.Join(c.aliases, ", "),
			)
		}
		return stringsIter(lines), nil
	}

	var lines []string
	for _, c := range h.commands {
		line := fmt.Sprintf("%-14s %s", c.name, c.short)
		lines = append(lines, line)
	}
	return stringsIter(lines), nil
}

// --- Helpers ---

func (h *Handler) showCurrentLocation(
	ctx context.Context,
) (iterator.Iterator[component.Responsive], error) {
	resp, err := h.dbg.StackTrace(ctx, &dap.StackTraceArguments{
		ThreadId:   h.threadID,
		StartFrame: h.frameIndex,
		Levels:     1,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.StackFrames) == 0 {
		return stringIter("(no frames)"), nil
	}
	frame := resp.StackFrames[0]
	h.frameID = frame.Id
	return formatFrameLocation(frame), nil
}

func parseFileLocation(loc string) (string, int, error) {
	idx := strings.LastIndex(loc, ":")
	if idx < 0 {
		return "", 0, fmt.Errorf(
			"invalid location %q — expected <file>:<line>", loc,
		)
	}
	file := loc[:idx]
	line, err := strconv.Atoi(loc[idx+1:])
	if err != nil {
		return "", 0, fmt.Errorf(
			"invalid line number in %q: %w", loc, err,
		)
	}
	return file, line, nil
}

// stringIter returns a single-item iterator with a text line.
func stringIter(s string) iterator.Iterator[component.Responsive] {
	r := component.NewResponsiveString(s, component.StringResponsiveConfig{})
	return iterator.FromSlice([]component.Responsive{r})
}

// stringsIter returns an iterator with one line per string.
func stringsIter(lines []string) iterator.Iterator[component.Responsive] {
	items := make([]component.Responsive, len(lines))
	for i, s := range lines {
		items[i] = component.NewResponsiveString(
			s, component.StringResponsiveConfig{},
		)
	}
	return iterator.FromSlice(items)
}
