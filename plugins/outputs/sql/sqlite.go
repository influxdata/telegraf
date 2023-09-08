//go:build !mips && !mipsle && !mips64 && !mips64le && !(windows && 386)

package sql

// The modernc.org sqlite driver isn't supported on all
// platforms. Register it with build constraints to prevent build
// failures on unsupported platforms.
import (
	_ "modernc.org/sqlite" // Register sqlite sql driver
)
