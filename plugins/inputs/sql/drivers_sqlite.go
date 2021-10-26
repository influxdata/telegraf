//go:build linux && freebsd && darwin && (!mips || !mips64)
// +build linux
// +build freebsd
// +build darwin
// +build !mips !mips64

package sql

import (
	_ "modernc.org/sqlite"
)
