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

package extensionapi

import (
	"io"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc/grpclog"
)

// LoggingTimestampFormat is the format used
const LoggingTimestampFormat = "2006-01-02T15:04:05.000000Z07:00"

func setupExtensionLogging(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "@timestamp"
				// Also apply custom format if it's a time.Time
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(LoggingTimestampFormat))
				}
			case slog.MessageKey:
				a.Key = "@message"
			}
			return a
		},
	}))
	slog.SetDefault(logger)
	disableGRPCLogging()
}

func disableGRPCLogging() {
	_ = os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "FATAL")
	_ = os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "0")
	discard := grpclog.NewLoggerV2WithVerbosity(io.Discard, io.Discard, io.Discard, 0)
	grpclog.SetLoggerV2(discard)
}
