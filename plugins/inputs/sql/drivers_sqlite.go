//go:build linux && freebsd && darwin && (!mips || !mips64)

package sql

import (
	// Blank imports to register the sqlite driver
	_ "modernc.org/sqlite"
)
