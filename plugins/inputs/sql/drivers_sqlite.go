//go:build !mips && !mipsle && !mips64 && !ppc64 && !riscv64 && !loong64 && !mips64le && !(windows && (386 || arm)) && !(freebsd && (386 || arm))

package sql

import (
	// Blank imports to register the sqlite driver
	_ "modernc.org/sqlite"
)
