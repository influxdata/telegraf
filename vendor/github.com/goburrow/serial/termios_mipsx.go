// +build linux
// +build mips mipsle mips64 mips64le

package serial

import (
	"syscall"
)

func cfSetIspeed(termios *syscall.Termios, speed uint32) {
	// MIPS has no Ispeed field in termios.
}

func cfSetOspeed(termios *syscall.Termios, speed uint32) {
	// MIPS has no Ospeed field in termios.
}
