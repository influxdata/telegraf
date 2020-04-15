// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
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

package glog

import (
	"flag"
	"io"
	"os"
)

func init() {
	flag.Var(&logging.verbosity, "v", "log level for V logs")
	flag.Var(&logging.vmodule, "vmodule",
		"comma-separated list of pattern=N settings for file-filtered logging")
	flag.Var(&logging.traceLocation, "log_backtrace_at",
		"when logging hits line file:N, emit a stack trace")

	logging.toWriter = true
	logging.writer = os.Stderr

	logging.setVState(0, nil, false)
}

// SetVGlobal sets the global verbosity level.
func SetVGlobal(level string) {
	// This value doesn't matter, I just need something to call Set on
	l := Level(0)
	l.Set(level)
}

// SetVModule sets the per-module verbosity level.
// Syntax: message=2,routing*=1
func SetVModule(value string) error {
	// This value doesn't matter, I just need something to call Set on
	m := moduleSpec{}
	return m.Set(value)
}

// SetOutput sets the writer for log output. By default this is os.StdErr.
// It returns the writer that was previously set.
func SetOutput(w io.Writer) io.Writer {
	prev := logging.writer
	logging.writer = w
	return prev
}
