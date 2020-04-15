package serial

import (
	"syscall"
)

func cfSetIspeed(termios *syscall.Termios, speed uint64) {
	termios.Ispeed = speed
}

func cfSetOspeed(termios *syscall.Termios, speed uint64) {
	termios.Ospeed = speed
}
