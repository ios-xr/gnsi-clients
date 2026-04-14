// Copyright 2024 Cisco Systems, Inc.
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

package logger

import (
	"fmt"
	"os"
)

// Logger provides simple logging functionality
type Logger struct {
	verbose bool
}

// New creates a new logger
func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

// Info prints informational messages (always shown)
func (l *Logger) Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Verbose prints verbose messages (only if verbose flag is set)
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// Error prints error messages to stderr
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// Success prints success messages with checkmark
func (l *Logger) Success(format string, args ...interface{}) {
	fmt.Printf("✓ "+format+"\n", args...)
}

// Warning prints warning messages
func (l *Logger) Warning(format string, args ...interface{}) {
	fmt.Printf("⚠ "+format+"\n", args...)
}
