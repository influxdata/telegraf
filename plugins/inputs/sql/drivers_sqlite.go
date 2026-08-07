// According to the support matrix at https://pkg.go.dev/modernc.org/sqlite
//go:build (darwin && (amd64 || arm64)) || (freebsd && (amd64 || arm64)) || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x)) || (openbsd && (amd64 || arm64)) || (windows && (386 || amd64 || arm64))

package sql

import (
	// Blank imports to register the sqlite driver
	_ "modernc.org/sqlite"
)
