//go:build !arm && !mips && !mips64
// +build !arm,!mips,!mips64

package sql

import (
	_ "modernc.org/sqlite"
)
