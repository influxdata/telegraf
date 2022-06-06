//go:build linux && freebsd && darwin && windows && (!mips || !mips64)
// +build linux
// +build freebsd
// +build darwin
// +build windows
// +build !mips !mips64

package sql

import (
	_ "modernc.org/sqlite"
)
