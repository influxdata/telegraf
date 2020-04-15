// +build darwin linux freebsd openbsd netbsd

package serial

import (
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// port implements Port interface.
type port struct {
	fd         int
	oldTermios *syscall.Termios

	timeout time.Duration
}

const (
	rs485Enabled      = 1 << 0
	rs485RTSOnSend    = 1 << 1
	rs485RTSAfterSend = 1 << 2
	rs485RXDuringTX   = 1 << 4
	rs485Tiocs        = 0x542f
)

// rs485_ioctl_opts is used to configure RS485 options in the driver
type rs485_ioctl_opts struct {
	flags                 uint32
	delay_rts_before_send uint32
	delay_rts_after_send  uint32
	padding               [5]uint32
}

// New allocates and returns a new serial port controller.
func New() Port {
	return &port{fd: -1}
}

// Open connects to the given serial port.
func (p *port) Open(c *Config) (err error) {
	termios, err := newTermios(c)
	if err != nil {
		return
	}
	// See man termios(3).
	// O_NOCTTY: no controlling terminal.
	// O_NDELAY: no data carrier detect.
	p.fd, err = syscall.Open(c.Address, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NDELAY|syscall.O_CLOEXEC, 0666)
	if err != nil {
		return
	}
	// Backup current termios to restore on closing.
	p.backupTermios()
	if err = p.setTermios(termios); err != nil {
		// No need to restore termios
		syscall.Close(p.fd)
		p.fd = -1
		p.oldTermios = nil
		return err
	}
	if err = enableRS485(p.fd, &c.RS485); err != nil {
		p.Close()
		return err
	}
	p.timeout = c.Timeout
	return
}

func (p *port) Close() (err error) {
	if p.fd == -1 {
		return
	}
	p.restoreTermios()
	err = syscall.Close(p.fd)
	p.fd = -1
	p.oldTermios = nil
	return
}

// Read reads from serial port. Port must be opened before calling this method.
// It is blocked until all data received or timeout after p.timeout.
func (p *port) Read(b []byte) (n int, err error) {
	var rfds syscall.FdSet

	fd := p.fd
	fdset(fd, &rfds)

	var tv *syscall.Timeval
	if p.timeout > 0 {
		timeout := syscall.NsecToTimeval(p.timeout.Nanoseconds())
		tv = &timeout
	}
	for {
		// If syscall.Select() returns EINTR (Interrupted system call), retry it
		if err = syscallSelect(fd+1, &rfds, nil, nil, tv); err == nil {
			break
		}
		if err != syscall.EINTR {
			err = fmt.Errorf("serial: could not select: %v", err)
			return
		}
	}
	if !fdisset(fd, &rfds) {
		// Timeout
		err = ErrTimeout
		return
	}
	n, err = syscall.Read(fd, b)
	return
}

// Write writes data to the serial port.
func (p *port) Write(b []byte) (n int, err error) {
	n, err = syscall.Write(p.fd, b)
	return
}

func (p *port) setTermios(termios *syscall.Termios) (err error) {
	if err = tcsetattr(p.fd, termios); err != nil {
		err = fmt.Errorf("serial: could not set setting: %v", err)
	}
	return
}

// backupTermios saves current termios setting.
// Make sure that device file has been opened before calling this function.
func (p *port) backupTermios() {
	oldTermios := &syscall.Termios{}
	if err := tcgetattr(p.fd, oldTermios); err != nil {
		// Warning only.
		log.Printf("serial: could not get setting: %v\n", err)
		return
	}
	// Will be reloaded when closing.
	p.oldTermios = oldTermios
}

// restoreTermios restores backed up termios setting.
// Make sure that device file has been opened before calling this function.
func (p *port) restoreTermios() {
	if p.oldTermios == nil {
		return
	}
	if err := tcsetattr(p.fd, p.oldTermios); err != nil {
		// Warning only.
		log.Printf("serial: could not restore setting: %v\n", err)
		return
	}
	p.oldTermios = nil
}

// Helpers for termios

func newTermios(c *Config) (termios *syscall.Termios, err error) {
	termios = &syscall.Termios{}
	flag := termios.Cflag
	// Baud rate
	if c.BaudRate == 0 {
		// 19200 is the required default.
		flag = syscall.B19200
	} else {
		var ok bool
		flag, ok = baudRates[c.BaudRate]
		if !ok {
			err = fmt.Errorf("serial: unsupported baud rate %v", c.BaudRate)
			return
		}
	}
	termios.Cflag |= flag
	// Input baud.
	cfSetIspeed(termios, flag)
	// Output baud.
	cfSetOspeed(termios, flag)
	// Character size.
	if c.DataBits == 0 {
		flag = syscall.CS8
	} else {
		var ok bool
		flag, ok = charSizes[c.DataBits]
		if !ok {
			err = fmt.Errorf("serial: unsupported character size %v", c.DataBits)
			return
		}
	}
	termios.Cflag |= flag
	// Stop bits
	switch c.StopBits {
	case 0, 1:
		// Default is one stop bit.
		// noop
	case 2:
		// CSTOPB: Set two stop bits.
		termios.Cflag |= syscall.CSTOPB
	default:
		err = fmt.Errorf("serial: unsupported stop bits %v", c.StopBits)
		return
	}
	switch c.Parity {
	case "N":
		// noop
	case "O":
		// PARODD: Parity is odd.
		termios.Cflag |= syscall.PARODD
		fallthrough
	case "", "E":
		// As mentioned in the modbus spec, the default parity mode must be Even parity
		// PARENB: Enable parity generation on output.
		termios.Cflag |= syscall.PARENB
		// INPCK: Enable input parity checking.
		termios.Iflag |= syscall.INPCK
	default:
		err = fmt.Errorf("serial: unsupported parity %v", c.Parity)
		return
	}
	// Control modes.
	// CREAD: Enable receiver.
	// CLOCAL: Ignore control lines.
	termios.Cflag |= syscall.CREAD | syscall.CLOCAL
	// Special characters.
	// VMIN: Minimum number of characters for noncanonical read.
	// VTIME: Time in deciseconds for noncanonical read.
	// Both are unused as NDELAY is we utilized when opening device.
	return
}

// enableRS485 enables RS485 functionality of driver via an ioctl if the config says so
func enableRS485(fd int, config *RS485Config) error {
	if !config.Enabled {
		return nil
	}
	rs485 := rs485_ioctl_opts{
		rs485Enabled,
		uint32(config.DelayRtsBeforeSend / time.Millisecond),
		uint32(config.DelayRtsAfterSend / time.Millisecond),
		[5]uint32{0, 0, 0, 0, 0},
	}

	if config.RtsHighDuringSend {
		rs485.flags |= rs485RTSOnSend
	}
	if config.RtsHighAfterSend {
		rs485.flags |= rs485RTSAfterSend
	}
	if config.RxDuringTx {
		rs485.flags |= rs485RXDuringTX
	}

	r, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(rs485Tiocs),
		uintptr(unsafe.Pointer(&rs485)))
	if errno != 0 {
		return os.NewSyscallError("SYS_IOCTL (RS485)", errno)
	}
	if r != 0 {
		return errors.New("serial: unknown error from SYS_IOCTL (RS485)")
	}
	return nil
}
