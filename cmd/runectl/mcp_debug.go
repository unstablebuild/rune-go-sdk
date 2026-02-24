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

package main

import (
	"context"
	"encoding/json"

	"github.com/google/go-dap"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/unstablebuild/rune-go-sdk/api/debugapi"
	"github.com/unstablebuild/rune-go-sdk/api/extensionapi"
)

func registerDebugTools( //nolint:funlen
	s *server.MCPServer,
	w *extensionapi.Workspace,
	bgCtx context.Context,
) {
	d := w.Debugger(bgCtx)
	registerDAPLaunch(s, d, bgCtx)
	registerDAPAttach(s, d, bgCtx)
	registerDAPConfigurationDone(s, d, bgCtx)
	registerDAPDisconnect(s, d, bgCtx)
	registerDAPTerminate(s, d, bgCtx)
	registerDAPRestart(s, d, bgCtx)
	registerDAPSetBreakpoints(s, d, bgCtx)
	registerDAPSetFunctionBreakpoints(s, d, bgCtx)
	registerDAPSetExceptionBreakpoints(s, d, bgCtx)
	registerDAPContinue(s, d, bgCtx)
	registerDAPNext(s, d, bgCtx)
	registerDAPStepIn(s, d, bgCtx)
	registerDAPStepOut(s, d, bgCtx)
	registerDAPStepBack(s, d, bgCtx)
	registerDAPReverseContinue(s, d, bgCtx)
	registerDAPPause(s, d, bgCtx)
	registerDAPThreads(s, d, bgCtx)
	registerDAPStackTrace(s, d, bgCtx)
	registerDAPScopes(s, d, bgCtx)
	registerDAPVariables(s, d, bgCtx)
	registerDAPSetVariable(s, d, bgCtx)
	registerDAPSource(s, d, bgCtx)
	registerDAPEvaluate(s, d, bgCtx)
	registerDAPSetExpression(s, d, bgCtx)
	registerDAPCompletions(s, d, bgCtx)
	registerDAPExceptionInfo(s, d, bgCtx)
	registerDAPModules(s, d, bgCtx)
	registerDAPLoadedSources(s, d, bgCtx)
	registerDAPReadMemory(s, d, bgCtx)
	registerDAPWriteMemory(s, d, bgCtx)
	registerDAPDisassemble(s, d, bgCtx)
	registerDAPGotoTargets(s, d, bgCtx)
	registerDAPGoto(s, d, bgCtx)
}

// mcpOK returns a simple success MCP result.
func mcpOK() *mcp.CallToolResult {
	return mcp.NewToolResultText(`{"ok":true}`)
}

func registerDAPLaunch(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_launch",
			mcp.WithDescription(
				"Starts a debug session by launching a program. "+
					"This is typically the first step in a debug session. "+
					"After launching, set breakpoints with dap_set_breakpoints, "+
					"then call dap_configuration_done to signal that setup is "+
					"complete before calling dap_continue.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithString("program", mcp.Required(), mcp.Description("Path to the program to debug")),
			mcp.WithArray("args", mcp.Description("Command line arguments")),
			mcp.WithString("cwd", mcp.Description("Working directory for the program")),
			mcp.WithString("env", mcp.Description(
				"Environment variables as JSON object string, "+
					`e.g. '{"KEY":"value"}'`,
			),
			),
			mcp.WithBoolean("stop_on_entry", mcp.Description("Stop at program entry point")),
			mcp.WithBoolean("no_debug", mcp.Description("Launch without debugging")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			program, err := req.RequireString("program")
			if err != nil {
				return mcpErr(err), nil
			}
			args := debugapi.LaunchRequestArguments{
				Program:     program,
				Args:        req.GetStringSlice("args", nil),
				Cwd:         req.GetString("cwd", ""),
				StopOnEntry: req.GetBool("stop_on_entry", false),
				NoDebug:     req.GetBool("no_debug", false),
			}
			if envStr := req.GetString("env", ""); envStr != "" {
				env := make(map[string]string)
				if jErr := json.Unmarshal([]byte(envStr), &env); jErr != nil {
					return mcpErr(jErr), nil
				}
				args.Env = env
			}
			if err := d.Launch(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPAttach(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_attach",
			mcp.WithDescription(
				"Attaches to an already-running process for debugging. "+
					"Alternative to dap_launch when the program is already running. "+
					"After attaching, set breakpoints then call "+
					"dap_configuration_done.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("pid", mcp.Description("Process ID to attach to")),
			mcp.WithString("program", mcp.Description("Path to the program binary")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := debugapi.AttachRequestArguments{
				PID:     int(req.GetFloat("pid", 0)),
				Program: req.GetString("program", ""),
			}
			if err := d.Attach(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPConfigurationDone(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_configuration_done",
			mcp.WithDescription(
				"Tells the debug adapter that all configuration "+
					"(breakpoints, exception settings) is done. Must be called "+
					"after dap_launch or dap_attach and after initial breakpoint "+
					"setup. The debuggee may start or resume execution after "+
					"this call.\n\n"+
					"Returns confirmation or error.",
			),
		),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if err := d.ConfigurationDone(bgCtx); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPDisconnect(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_disconnect",
			mcp.WithDescription(
				"Ends the debug session. Use this when you're done debugging. "+
					"Can optionally terminate or restart the debuggee.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithBoolean("restart", mcp.Description("Restart the debuggee after disconnecting")),
			mcp.WithBoolean("terminate_debuggee", mcp.Description("Terminate the debuggee process")),
			mcp.WithBoolean("suspend_debuggee", mcp.Description("Suspend the debuggee after disconnecting")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := &dap.DisconnectArguments{
				Restart:           req.GetBool("restart", false),
				TerminateDebuggee: req.GetBool("terminate_debuggee", false),
				SuspendDebuggee:   req.GetBool("suspend_debuggee", false),
			}
			if err := d.Disconnect(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPTerminate(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_terminate",
			mcp.WithDescription(
				"Requests graceful termination of the debuggee process. "+
					"Unlike dap_disconnect, this specifically targets the "+
					"debuggee. Use when you want to stop the program but may "+
					"continue the debug session.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithBoolean("restart", mcp.Description("Restart after termination")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := &dap.TerminateArguments{
				Restart: req.GetBool("restart", false),
			}
			if err := d.Terminate(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPRestart(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_restart",
			mcp.WithDescription(
				"Restarts the debug session from the beginning. "+
					"Equivalent to terminate + launch. Breakpoints and "+
					"other configuration are preserved.\n\n"+
					"Returns confirmation or error.",
			),
		),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if err := d.Restart(bgCtx); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPSetBreakpoints(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_set_breakpoints",
			mcp.WithDescription(
				"Sets all breakpoints for a source file, replacing any "+
					"previously set breakpoints in that file. Each breakpoint "+
					"can have optional conditions, hit counts, and log "+
					"messages.\n\n"+
					"To clear all breakpoints in a file, call with an empty "+
					"breakpoints array.\n\n"+
					"Returns the actual breakpoints set (may differ from "+
					"requested if lines are adjusted).",
			),
			mcp.WithString("source_path", mcp.Required(), mcp.Description("Absolute path to the source file")),
			mcp.WithString("breakpoints", mcp.Required(), mcp.Description(
					`JSON array of breakpoints. Example: `+
						`[{"line":10},{"line":20,"condition":"x>5"}]`,
				),
			),
			mcp.WithBoolean("source_modified", mcp.Description("Whether the source has been modified since last build")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			srcPath, err := req.RequireString("source_path")
			if err != nil {
				return mcpErr(err), nil
			}
			bpJSON, err := req.RequireString("breakpoints")
			if err != nil {
				return mcpErr(err), nil
			}
			var bps []dap.SourceBreakpoint
			if jErr := json.Unmarshal([]byte(bpJSON), &bps); jErr != nil {
				return mcpErr(jErr), nil
			}
			args := &dap.SetBreakpointsArguments{
				Source:         dap.Source{Path: srcPath},
				Breakpoints:    bps,
				SourceModified: req.GetBool("source_modified", false),
			}
			result, err := d.SetBreakpoints(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPSetFunctionBreakpoints(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_set_function_breakpoints",
			mcp.WithDescription(
				"Sets breakpoints on function entry points by name. "+
					"Replaces all previously set function breakpoints. "+
					"Useful when you don't know the exact source location.\n\n"+
					"Returns the actual breakpoints set.",
			),
			mcp.WithString("breakpoints", mcp.Required(), mcp.Description(
					`JSON array of function breakpoints. Example: `+
						`[{"name":"main.handleRequest"},`+
						`{"name":"processData","condition":"len(items)>0"}]`,
				),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			bpJSON, err := req.RequireString("breakpoints")
			if err != nil {
				return mcpErr(err), nil
			}
			var bps []dap.FunctionBreakpoint
			if jErr := json.Unmarshal([]byte(bpJSON), &bps); jErr != nil {
				return mcpErr(jErr), nil
			}
			args := &dap.SetFunctionBreakpointsArguments{
				Breakpoints: bps,
			}
			result, err := d.SetFunctionBreakpoints(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPSetExceptionBreakpoints(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_set_exception_breakpoints",
			mcp.WithDescription(
				"Configures which exceptions should cause the debugger to "+
					"break. The available filter IDs depend on the debug adapter "+
					"(e.g., 'raised', 'uncaught' for Python; 'all', 'unhandled' "+
					"for Go).\n\n"+
					"Returns the actual breakpoints set.",
			),
			mcp.WithArray("filters", mcp.Required(), mcp.Description("Exception filter IDs to enable")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			filters := req.GetStringSlice("filters", nil)
			args := &dap.SetExceptionBreakpointsArguments{
				Filters: filters,
			}
			result, err := d.SetExceptionBreakpoints(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPContinue(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_continue",
			mcp.WithDescription(
				"Resumes execution of all threads (or a single thread). "+
					"The debuggee will run until it hits a breakpoint, completes, "+
					"or encounters an error. After calling, use dap_threads to "+
					"check thread states when execution stops.\n\n"+
					"Returns {all_threads_continued}.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to continue")),
			mcp.WithBoolean("single_thread", mcp.Description("Only continue this thread")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.ContinueArguments{
				ThreadId:     int(tid),
				SingleThread: req.GetBool("single_thread", false),
			}
			result, err := d.Continue(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPNext(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_next",
			mcp.WithDescription(
				"Executes one step, stepping over function calls. "+
					"The thread will stop at the next statement in the "+
					"current function.\n\n"+
					"After stepping, inspect state with dap_stack_trace "+
					"→ dap_scopes → dap_variables.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to step")),
			mcp.WithBoolean("single_thread", mcp.Description("Only step this thread")),
			mcp.WithString("granularity", mcp.Description("Stepping granularity: statement, line, or instruction")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.NextArguments{
				ThreadId:     int(tid),
				SingleThread: req.GetBool("single_thread", false),
				Granularity:  dap.SteppingGranularity(req.GetString("granularity", "")),
			}
			if err := d.Next(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPStepIn(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_step_in",
			mcp.WithDescription(
				"Steps into a function call on the current line. "+
					"If the current line has no function call, behaves "+
					"like dap_next.\n\n"+
					"After stepping, inspect state with dap_stack_trace "+
					"→ dap_scopes → dap_variables.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to step")),
			mcp.WithBoolean("single_thread", mcp.Description("Only step this thread")),
			mcp.WithNumber("target_id", mcp.Description("Step into target ID from step-in targets")),
			mcp.WithString("granularity", mcp.Description("Stepping granularity: statement, line, or instruction")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.StepInArguments{
				ThreadId:     int(tid),
				SingleThread: req.GetBool("single_thread", false),
				TargetId:     int(req.GetFloat("target_id", 0)),
				Granularity:  dap.SteppingGranularity(req.GetString("granularity", "")),
			}
			if err := d.StepIn(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPStepOut(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_step_out",
			mcp.WithDescription(
				"Runs until the current function returns, then stops. "+
					"Use this to finish executing the current function and "+
					"return to the caller.\n\n"+
					"After stepping, inspect state with dap_stack_trace "+
					"→ dap_scopes → dap_variables.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to step out of")),
			mcp.WithBoolean("single_thread", mcp.Description("Only step this thread")),
			mcp.WithString("granularity", mcp.Description("Stepping granularity: statement, line, or instruction")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.StepOutArguments{
				ThreadId:     int(tid),
				SingleThread: req.GetBool("single_thread", false),
				Granularity:  dap.SteppingGranularity(req.GetString("granularity", "")),
			}
			if err := d.StepOut(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPStepBack(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_step_back",
			mcp.WithDescription(
				"Steps backward by one statement. Only available if "+
					"the debug adapter supports reverse debugging (rare).\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to step back")),
			mcp.WithBoolean("single_thread", mcp.Description("Only step this thread")),
			mcp.WithString("granularity", mcp.Description("Stepping granularity: statement, line, or instruction")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.StepBackArguments{
				ThreadId:     int(tid),
				SingleThread: req.GetBool("single_thread", false),
				Granularity:  dap.SteppingGranularity(req.GetString("granularity", "")),
			}
			if err := d.StepBack(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPReverseContinue(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_reverse_continue",
			mcp.WithDescription(
				"Resumes execution backward until a breakpoint is hit. "+
					"Only available if the debug adapter supports reverse "+
					"debugging.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to reverse-continue")),
			mcp.WithBoolean("single_thread", mcp.Description("Only continue this thread")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.ReverseContinueArguments{
				ThreadId:     int(tid),
				SingleThread: req.GetBool("single_thread", false),
			}
			if err := d.ReverseContinue(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPPause(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_pause",
			mcp.WithDescription(
				"Suspends execution of a running thread. Use this when "+
					"the program is running and you want to inspect its state. "+
					"After pausing, use dap_threads and dap_stack_trace to "+
					"examine where each thread stopped.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to pause")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.PauseArguments{ThreadId: int(tid)}
			if err := d.Pause(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}

func registerDAPThreads(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_threads",
			mcp.WithDescription(
				"Returns all threads of the debuggee. Call this after "+
					"execution stops to see which threads are available. "+
					"Thread IDs are needed for most other debugging "+
					"operations.\n\n"+
					"Returns [{id, name}].",
			),
		),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			threads, err := d.Threads(bgCtx)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(threads)
		},
	)
}

func registerDAPStackTrace(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_stack_trace",
			mcp.WithDescription(
				"Returns the call stack for a specific thread. Call after "+
					"execution stops (breakpoint, step, pause). Each frame has "+
					"an id needed by dap_scopes.\n\n"+
					"Typical flow: dap_threads → dap_stack_trace → dap_scopes "+
					"→ dap_variables.\n\n"+
					"Returns {stack_frames: [{id, name, source, line, column}], "+
					"total_frames}.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to get stack trace for")),
			mcp.WithNumber("start_frame", mcp.Description("Index of the first frame to return (default 0)")),
			mcp.WithNumber("levels", mcp.Description("Maximum number of frames to return (0 = all)")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.StackTraceArguments{
				ThreadId:   int(tid),
				StartFrame: int(req.GetFloat("start_frame", 0)),
				Levels:     int(req.GetFloat("levels", 0)),
			}
			result, err := d.StackTrace(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPScopes(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_scopes",
			mcp.WithDescription(
				"Returns the variable scopes (local, global, etc.) for a "+
					"specific stack frame. Use the frame_id from dap_stack_trace. "+
					"Each scope has a variables_reference for dap_variables.\n\n"+
					"Typical flow: dap_stack_trace → dap_scopes → "+
					"dap_variables.\n\n"+
					"Returns [{name, variables_reference, named_variables, "+
					"indexed_variables, expensive}].",
			),
			mcp.WithNumber("frame_id", mcp.Required(), mcp.Description("Stack frame ID from dap_stack_trace")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			fid, err := req.RequireFloat("frame_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.ScopesArguments{FrameId: int(fid)}
			scopes, err := d.Scopes(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(scopes)
		},
	)
}

func registerDAPVariables(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_variables",
			mcp.WithDescription(
				"Returns child variables for a given variables_reference. "+
					"Use the variables_reference from dap_scopes to get "+
					"top-level variables, or from a previous dap_variables "+
					"call to expand nested objects.\n\n"+
					"Variables with variables_reference > 0 can be expanded "+
					"further.\n\n"+
					"Returns [{name, value, type, variables_reference}].",
			),
			mcp.WithNumber("variables_reference", mcp.Required(), mcp.Description("Reference from dap_scopes or dap_variables")),
			mcp.WithString("filter", mcp.Description("Filter: 'indexed' or 'named'")),
			mcp.WithNumber("start", mcp.Description("Start index for paged variables")),
			mcp.WithNumber("count", mcp.Description("Number of variables to return for paging")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			vref, err := req.RequireFloat("variables_reference")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.VariablesArguments{
				VariablesReference: int(vref),
				Filter:             req.GetString("filter", ""),
				Start:              int(req.GetFloat("start", 0)),
				Count:              int(req.GetFloat("count", 0)),
			}
			vars, err := d.Variables(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(vars)
		},
	)
}

func registerDAPSetVariable(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_set_variable",
			mcp.WithDescription(
				"Changes the value of a variable during debugging. "+
					"Use the variables_reference from the variable's parent "+
					"scope and the variable's name.\n\n"+
					"Returns {value, type, variables_reference}.",
			),
			mcp.WithNumber("variables_reference", mcp.Required(), mcp.Description("Parent scope reference")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Variable name to modify")),
			mcp.WithString("value", mcp.Required(), mcp.Description("New value for the variable")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			vref, err := req.RequireFloat("variables_reference")
			if err != nil {
				return mcpErr(err), nil
			}
			name, err := req.RequireString("name")
			if err != nil {
				return mcpErr(err), nil
			}
			value, err := req.RequireString("value")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.SetVariableArguments{
				VariablesReference: int(vref),
				Name:               name,
				Value:              value,
			}
			result, err := d.SetVariable(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPSource(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_source",
			mcp.WithDescription(
				"Retrieves the source code for a given source reference. "+
					"Useful for sources without a file path (e.g., decompiled "+
					"or generated code). For regular files, read the file "+
					"directly instead.\n\n"+
					"Returns {content, mime_type}.",
			),
			mcp.WithNumber("source_reference", mcp.Required(), mcp.Description("Source reference from a stack frame")),
			mcp.WithString("source_path", mcp.Description("Optional source path hint")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sref, err := req.RequireFloat("source_reference")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.SourceArguments{
				SourceReference: int(sref),
			}
			if p := req.GetString("source_path", ""); p != "" {
				args.Source = &dap.Source{Path: p}
			}
			result, err := d.Source(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPEvaluate(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_evaluate",
			mcp.WithDescription(
				"Evaluates an expression in the context of a stack frame. "+
					"Supports watch expressions, REPL evaluation, and hover "+
					"inspection depending on the context parameter.\n\n"+
					"Returns {result, type, variables_reference}.",
			),
			mcp.WithString("expression", mcp.Required(), mcp.Description("Expression to evaluate")),
			mcp.WithNumber("frame_id", mcp.Description("Stack frame context for evaluation")),
			mcp.WithString("context", mcp.Description("Evaluation context: watch, repl, hover, or clipboard")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			expr, err := req.RequireString("expression")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.EvaluateArguments{
				Expression: expr,
				FrameId:    int(req.GetFloat("frame_id", 0)),
				Context:    req.GetString("context", ""),
			}
			result, err := d.Evaluate(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPSetExpression(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_set_expression",
			mcp.WithDescription(
				"Assigns a value to an expression (typically a variable "+
					"or property path). Similar to dap_set_variable but works "+
					"with arbitrary expressions like 'obj.field' or "+
					"'arr[0]'.\n\n"+
					"Returns {value, type, variables_reference}.",
			),
			mcp.WithString("expression", mcp.Required(), mcp.Description("Expression to assign to")),
			mcp.WithString("value", mcp.Required(), mcp.Description("New value to assign")),
			mcp.WithNumber("frame_id", mcp.Description("Stack frame context")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			expr, err := req.RequireString("expression")
			if err != nil {
				return mcpErr(err), nil
			}
			value, err := req.RequireString("value")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.SetExpressionArguments{
				Expression: expr,
				Value:      value,
				FrameId:    int(req.GetFloat("frame_id", 0)),
			}
			result, err := d.SetExpression(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPCompletions(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_completions",
			mcp.WithDescription(
				"Returns completion suggestions for a partial expression "+
					"in the debug console. Useful for discovering available "+
					"variables and methods at the current execution point.\n\n"+
					"Returns [{label, text, type}].",
			),
			mcp.WithString("text", mcp.Required(), mcp.Description("Partial expression to complete")),
			mcp.WithNumber("column", mcp.Required(), mcp.Description("Cursor column in the text (1-based)")),
			mcp.WithNumber("frame_id", mcp.Description("Stack frame context")),
			mcp.WithNumber("line", mcp.Description("Line number in the text (0-based)")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			text, err := req.RequireString("text")
			if err != nil {
				return mcpErr(err), nil
			}
			col, err := req.RequireFloat("column")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.CompletionsArguments{
				Text:    text,
				Column:  int(col),
				FrameId: int(req.GetFloat("frame_id", 0)),
				Line:    int(req.GetFloat("line", 0)),
			}
			items, err := d.Completions(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(items)
		},
	)
}

func registerDAPExceptionInfo(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_exception_info",
			mcp.WithDescription(
				"Retrieves detailed information about the exception that "+
					"caused execution to stop on a specific thread. Call this "+
					"when execution stops due to an exception breakpoint.\n\n"+
					"Returns {exception_id, description, break_mode}.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread that hit the exception")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.ExceptionInfoArguments{ThreadId: int(tid)}
			result, err := d.ExceptionInfo(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPModules(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_modules",
			mcp.WithDescription(
				"Returns information about loaded modules/libraries. "+
					"Useful for understanding which code is loaded and whether "+
					"debug symbols are available.\n\n"+
					"Returns {modules: [{id, name, path}], total_modules}.",
			),
			mcp.WithNumber("start_module", mcp.Description("Index of first module to return")),
			mcp.WithNumber("module_count", mcp.Description("Number of modules to return")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := &dap.ModulesArguments{
				StartModule: int(req.GetFloat("start_module", 0)),
				ModuleCount: int(req.GetFloat("module_count", 0)),
			}
			result, err := d.Modules(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPLoadedSources(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_loaded_sources",
			mcp.WithDescription(
				"Returns all source files that are currently loaded by "+
					"the debugger. Useful for discovering what source files "+
					"are available for setting breakpoints.\n\n"+
					"Returns [{name, path, source_reference}].",
			),
		),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			sources, err := d.LoadedSources(bgCtx)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(sources)
		},
	)
}

func registerDAPReadMemory(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_read_memory",
			mcp.WithDescription(
				"Reads raw bytes from a memory address. The "+
					"memory_reference comes from a variable's "+
					"memory_reference field.\n\n"+
					"Returns {address, data, unreadable_bytes}.",
			),
			mcp.WithString("memory_reference", mcp.Required(), mcp.Description("Memory reference from a variable")),
			mcp.WithNumber("offset", mcp.Description("Byte offset from the reference")),
			mcp.WithNumber("count", mcp.Required(), mcp.Description("Number of bytes to read")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			memRef, err := req.RequireString("memory_reference")
			if err != nil {
				return mcpErr(err), nil
			}
			count, err := req.RequireFloat("count")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.ReadMemoryArguments{
				MemoryReference: memRef,
				Offset:          int(req.GetFloat("offset", 0)),
				Count:           int(count),
			}
			result, err := d.ReadMemory(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPWriteMemory(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_write_memory",
			mcp.WithDescription(
				"Writes raw bytes to a memory address. Use with caution "+
					"as this directly modifies process memory.\n\n"+
					"Returns {offset, bytes_written}.",
			),
			mcp.WithString("memory_reference", mcp.Required(), mcp.Description("Memory reference to write to")),
			mcp.WithNumber("offset", mcp.Description("Byte offset from the reference")),
			mcp.WithString("data", mcp.Required(), mcp.Description("Base64-encoded bytes to write")),
			mcp.WithBoolean("allow_partial", mcp.Description("Allow partial writes")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			memRef, err := req.RequireString("memory_reference")
			if err != nil {
				return mcpErr(err), nil
			}
			data, err := req.RequireString("data")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.WriteMemoryArguments{
				MemoryReference: memRef,
				Offset:          int(req.GetFloat("offset", 0)),
				Data:            data,
				AllowPartial:    req.GetBool("allow_partial", false),
			}
			result, err := d.WriteMemory(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPDisassemble(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_disassemble",
			mcp.WithDescription(
				"Returns disassembled CPU instructions at a memory "+
					"address. Useful for low-level debugging when source "+
					"code is not available.\n\n"+
					"Returns [{address, instruction, instruction_bytes, "+
					"symbol}].",
			),
			mcp.WithString("memory_reference", mcp.Required(), mcp.Description("Memory address to disassemble")),
			mcp.WithNumber("offset", mcp.Description("Byte offset from reference")),
			mcp.WithNumber("instruction_offset", mcp.Description("Instruction offset from reference")),
			mcp.WithNumber("instruction_count", mcp.Required(), mcp.Description("Number of instructions to disassemble")),
			mcp.WithBoolean("resolve_symbols", mcp.Description("Resolve symbol names for addresses")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			memRef, err := req.RequireString("memory_reference")
			if err != nil {
				return mcpErr(err), nil
			}
			instCount, err := req.RequireFloat("instruction_count")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.DisassembleArguments{
				MemoryReference:   memRef,
				Offset:            int(req.GetFloat("offset", 0)),
				InstructionOffset: int(req.GetFloat("instruction_offset", 0)),
				InstructionCount:  int(instCount),
				ResolveSymbols:    req.GetBool("resolve_symbols", false),
			}
			result, err := d.Disassemble(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(result)
		},
	)
}

func registerDAPGotoTargets(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_goto_targets",
			mcp.WithDescription(
				"Returns the possible goto targets (valid execution "+
					"points) for a given source location. Use before "+
					"dap_goto to discover valid jump destinations.\n\n"+
					"Returns [{id, label, line, column}].",
			),
			mcp.WithString("source_path", mcp.Required(), mcp.Description("Absolute path to the source file")),
			mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number in the source")),
			mcp.WithNumber("column", mcp.Description("Column in the source")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			srcPath, err := req.RequireString("source_path")
			if err != nil {
				return mcpErr(err), nil
			}
			line, err := req.RequireFloat("line")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.GotoTargetsArguments{
				Source: dap.Source{Path: srcPath},
				Line:   int(line),
				Column: int(req.GetFloat("column", 0)),
			}
			targets, err := d.GotoTargets(bgCtx, args)
			if err != nil {
				return mcpErr(err), nil
			}
			return mcpJSON(targets)
		},
	)
}

func registerDAPGoto(s *server.MCPServer, d debugapi.Debugger, bgCtx context.Context) {
	s.AddTool(
		mcp.NewTool("dap_goto",
			mcp.WithDescription(
				"Sets the execution point to a specific goto target. "+
					"The thread will continue from the target location. "+
					"Use dap_goto_targets first to get valid target IDs.\n\n"+
					"Returns confirmation or error.",
			),
			mcp.WithNumber("thread_id", mcp.Required(), mcp.Description("Thread to set execution point")),
			mcp.WithNumber("target_id", mcp.Required(), mcp.Description("Goto target ID from dap_goto_targets")),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tid, err := req.RequireFloat("thread_id")
			if err != nil {
				return mcpErr(err), nil
			}
			targetID, err := req.RequireFloat("target_id")
			if err != nil {
				return mcpErr(err), nil
			}
			args := &dap.GotoArguments{
				ThreadId: int(tid),
				TargetId: int(targetID),
			}
			if err := d.Goto(bgCtx, args); err != nil {
				return mcpErr(err), nil
			}
			return mcpOK(), nil
		},
	)
}
