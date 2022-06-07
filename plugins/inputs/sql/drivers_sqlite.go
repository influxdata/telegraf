//go:build !arm && !mips && !mipsle && !mips64 && !mips64le
// +build !arm,!mips,!mipsle,!mips64,!mips64le

package sql

import (
	_ "modernc.org/sqlite"
)
