//go:build !mips || !mips64
// +build !mips !mips64

package sql

import (
	_ "modernc.org/sqlite"
)
