package serial

import (
	"fmt"
	"syscall"
	"unsafe"
)

var baudRates = map[int]uint64{
	50:     syscall.B50,
	75:     syscall.B75,
	110:    syscall.B110,
	134:    syscall.B134,
	150:    syscall.B150,
	200:    syscall.B200,
	300:    syscall.B300,
	600:    syscall.B600,
	1200:   syscall.B1200,
	1800:   syscall.B1800,
	2400:   syscall.B2400,
	4800:   syscall.B4800,
	9600:   syscall.B9600,
	19200:  syscall.B19200,
	38400:  syscall.B38400,
	57600:  syscall.B57600,
	115200: syscall.B115200,
	230400: syscall.B230400,
}

var charSizes = map[int]uint64{
	5: syscall.CS5,
	6: syscall.CS6,
	7: syscall.CS7,
	8: syscall.CS8,
}

// syscallSelect is a wapper for syscall.Select that only returns error.
func syscallSelect(n int, r *syscall.FdSet, w *syscall.FdSet, e *syscall.FdSet, tv *syscall.Timeval) error {
	return syscall.Select(n, r, w, e, tv)
}

// tcsetattr sets terminal file descriptor parameters.
// See man tcsetattr(3).
func tcsetattr(fd int, termios *syscall.Termios) (err error) {
	r, _, errno := syscall.Syscall(uintptr(syscall.SYS_IOCTL),
		uintptr(fd), uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(termios)))
	if errno != 0 {
		err = errno
		return
	}
	if r != 0 {
		err = fmt.Errorf("tcsetattr failed %v", r)
	}
	return
}

// tcgetattr gets terminal file descriptor parameters.
// See man tcgetattr(3).
func tcgetattr(fd int, termios *syscall.Termios) (err error) {
	r, _, errno := syscall.Syscall(uintptr(syscall.SYS_IOCTL),
		uintptr(fd), uintptr(syscall.TIOCGETA), uintptr(unsafe.Pointer(termios)))
	if errno != 0 {
		err = errno
		return
	}
	if r != 0 {
		err = fmt.Errorf("tcgetattr failed %v", r)
		return
	}
	return
}

// fdget returns index and offset of fd in fds.
func fdget(fd int, fds *syscall.FdSet) (index, offset int) {
	index = fd / (syscall.FD_SETSIZE / len(fds.Bits)) % len(fds.Bits)
	offset = fd % (syscall.FD_SETSIZE / len(fds.Bits))
	return
}

// fdset implements FD_SET macro.
func fdset(fd int, fds *syscall.FdSet) {
	idx, pos := fdget(fd, fds)
	fds.Bits[idx] = 1 << uint(pos)
}

// fdisset implements FD_ISSET macro.
func fdisset(fd int, fds *syscall.FdSet) bool {
	idx, pos := fdget(fd, fds)
	return fds.Bits[idx]&(1<<uint(pos)) != 0
}
