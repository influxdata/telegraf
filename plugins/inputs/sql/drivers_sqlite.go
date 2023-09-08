//go:build !mips && !mipsle && !mips64 && !mips64le && !(windows && 386)

package sql

import (
	// Blank imports to register the sqlite driver
	_ "modernc.org/sqlite"
)
