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

package debug

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"gopkg.in/yaml.v3"
)

// CapturePanicReportWith captures a panic in fn and writes to disk
// a yaml report. The boolean value returned indicates if fn run with no panics.
// If the value is false, the returned string indicates the fs location of the report.
// An error is returned if there was a panic but the report couldn't be stored.
func CapturePanicReportWith(dir, pkg, version string, run func()) (
	panicValue any, err error, ok bool,
) {
	defer func() {
		panicValue = recover()
		if panicValue == nil {
			return
		}
		report := BuildCrashReport(pkg, version, panicValue)

		var data []byte
		data, err = yaml.Marshal(report)
		if err != nil {
			slog.Error("yaml marshal", "error", err)
			return
		}
		var f *os.File
		f, err = os.CreateTemp(ReportsDir, fmt.Sprintf("%s_crash_report_", Package))
		if err != nil {
			slog.Error("create temp file", "error", err)
			return
		}
		if _, err = f.Write(data); err != nil {
			slog.Error("write to report %q: %v", f.Name(), err)
			return
		}
		slog.Warn("saved crash report", "file", "file://"+f.Name())
	}()

	run()
	ok = true
	return
}

// CapturePanicReport captures a panic with CapturePanicReportWith,
// and exits or simply returns if there was no panic in fn. It uses
// the compile-time variables Tag, Package and ReportsDir, so make
// sure they're injected at compile-time when using this helper.
func CapturePanicReport(fn func()) {
	defer func() {
		panicValue := recover()
		if panicValue == nil {
			return
		}
		defer panic(panicValue)
		report := BuildCrashReport(Package, Tag, panicValue)

		var data []byte
		data, err := yaml.Marshal(report)
		if err != nil {
			slog.Error("yaml marshal", "error", err)
			return
		}
		var f *os.File
		f, err = os.CreateTemp(ReportsDir, fmt.Sprintf("%s_crash_report_", Package))
		if err != nil {
			slog.Error("create temp file", "error", err)
			return
		}
		if _, err = f.Write(data); err != nil {
			slog.Error("write to", "report", f.Name(), "error", err)
			return
		}
		slog.Warn("saved crash report", "file", "file://"+f.Name())
		fmt.Fprintf(os.Stderr, "saved crash report file://%v", f.Name())
	}()

	fn()
}

// BuildCrashReport builds a crash report with the given panic value, for
// the given package and version.
func BuildCrashReport(pkg, version string, panicValue any) (
	report Report,
) {
	report = Report{
		Author:    "debug.CapturePanic",
		Package:   pkg,
		Version:   version,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	bi, ok := debug.ReadBuildInfo()
	if ok {
		report.Build = *bi
	}

	var errStr string
	switch x := panicValue.(type) {
	case string:
		errStr = x
	case error:
		errStr = x.Error()
	default:
		errStr = fmt.Sprintf("unknown: %v", panicValue)
	}
	report.Metadata[reportMetadataStackTraceField] = string(debug.Stack())
	report.Metadata[reportMetadataBugField] = ""
	report.Metadata[reportMetadataPanicField] = ""
	report.Subject = fmt.Sprintf("%20s", errStr)
	report.Metadata[reportMetadataErrorField] = errStr

	return
}

// Report contains informatin about a suspected or confirmed bug.
type Report struct {
	Package   string
	Version   string
	Author    string
	Subject   string
	Notes     string
	Build     debug.BuildInfo `yaml:"build,omitempty"`
	CreatedAt time.Time       `yaml:"created_at,omitempty"`
	UpdatedAt time.Time       `yaml:"updated_at,omitempty"`
	UpdatedBy string          `yaml:"-"`
	Closed    bool            `yaml:"closed,omitempty"`
	ClosedAt  time.Time       `yaml:"closed_at,omitempty"`
	Metadata  map[string]string
}

const (
	reportMetadataErrorField      = "error"
	reportMetadataStackTraceField = "stack"
	reportMetadataPanicField      = "panic"
	reportMetadataBugField        = "bug"
)
