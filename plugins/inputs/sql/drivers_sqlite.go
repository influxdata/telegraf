//go:build !arm && !mips && !mipsle && !mips64 && !mips64le && !ppc64 && !(freebsd && arm64)
// +build !arm
// +build !mips
// +build !mipsle
// +build !mips64
// +build !mips64le
// +build !ppc64
// +build !freebsd !arm64

package sql

import (
	_ "modernc.org/sqlite"
)
