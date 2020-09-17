// Copyright 2018-2020 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// Package debug provides functions for debug logging.
package debug

import (
	"log"
	"os"
	"strings"
)

// Flags contains the debug flags set by OPC_DEBUG.
//
//  * codec : print detailed debugging information when encoding/decoding
var Flags = os.Getenv("OPC_DEBUG")

// Enable controls whether debug logging is enabled. It is disabled by default.
var Enable bool = FlagSet("debug")

// Logger logs the debug messages when debug logging is enabled.
var Logger = log.New(os.Stderr, "debug: ", 0)

// Printf logs the message with Logger.Printf() when debug logging is enabled.
func Printf(format string, args ...interface{}) {
	if !Enable {
		return
	}
	Logger.Printf(format, args...)
}

// FlagSet returns true if the OPCUA_DEBUG environment variable contains the
// given flag.
func FlagSet(name string) bool {
	return stringSliceContains(name, strings.Fields(Flags))
}

func stringSliceContains(s string, vals []string) bool {
	for _, v := range vals {
		if s == v {
			return true
		}
	}
	return false
}
