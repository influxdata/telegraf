// +build !mips,!mipsle,!mips64,!mips64le

package serial

import (
	"syscall"
)

func cfSetIspeed(termios *syscall.Termios, speed uint32) {
	termios.Ispeed = speed
}

func cfSetOspeed(termios *syscall.Termios, speed uint32) {
	termios.Ospeed = speed
}
