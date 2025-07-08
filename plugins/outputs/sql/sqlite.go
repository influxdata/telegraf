//go:build !mips && !mipsle && !mips64 && !ppc64 && !riscv64 && !loong64 && !mips64le && !(windows && (386 || arm)) && !(freebsd && (386 || arm))

package sql

// The modernc.org sqlite driver isn't supported on all
// platforms. Register it with build constraints to prevent build
// failures on unsupported platforms.
import (
	_ "modernc.org/sqlite" // Register sqlite sql driver
)
