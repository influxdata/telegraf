// +build linux,freebsd,darwin
// +build !mips !mips64

package sql

// The modernc.org sqlite driver isn't supported on all
// platforms. Register it with build constraints to prevent build
// failures on unsupported platforms.
import (
	_ "modernc.org/sqlite" // Register sqlite sql driver
)
