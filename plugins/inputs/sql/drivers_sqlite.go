//go:build !mips && !mipsle && !mips64 && !mips64le

package sql

import (
	// Blank imports to register the sqlite driver
	_ "modernc.org/sqlite"
)
