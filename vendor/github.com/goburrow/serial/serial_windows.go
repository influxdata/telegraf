package serial

import (
	"fmt"
	"syscall"
)

type port struct {
	handle syscall.Handle

	oldDCB      c_DCB
	oldTimeouts c_COMMTIMEOUTS
}

// New allocates and returns a new serial port controller.
func New() Port {
	return &port{
		handle: syscall.InvalidHandle,
	}
}

// Open connects to the given serial port.
func (p *port) Open(c *Config) (err error) {
	p.handle, err = newHandle(c)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			syscall.CloseHandle(p.handle)
			p.handle = syscall.InvalidHandle
		}
	}()
	err = p.setSerialConfig(c)
	if err != nil {
		return
	}
	err = p.setTimeouts(c)
	return
}

func (p *port) Close() (err error) {
	if p.handle == syscall.InvalidHandle {
		return
	}
	err1 := SetCommTimeouts(p.handle, &p.oldTimeouts)
	err2 := SetCommState(p.handle, &p.oldDCB)
	err = syscall.CloseHandle(p.handle)
	if err == nil {
		if err1 == nil {
			err = err2
		} else {
			err = err1
		}
	}
	p.handle = syscall.InvalidHandle
	return
}

// Read reads from serial port.
// It is blocked until data received or timeout after p.timeout.
func (p *port) Read(b []byte) (n int, err error) {
	var done uint32
	if err = syscall.ReadFile(p.handle, b, &done, nil); err != nil {
		return
	}
	if done == 0 {
		err = ErrTimeout
		return
	}
	n = int(done)
	return
}

// Write writes data to the serial port.
func (p *port) Write(b []byte) (n int, err error) {
	var done uint32
	if err = syscall.WriteFile(p.handle, b, &done, nil); err != nil {
		return
	}
	n = int(done)
	return
}

func (p *port) setTimeouts(c *Config) error {
	var timeouts c_COMMTIMEOUTS
	// Read and write timeout
	if c.Timeout > 0 {
		timeout := toDWORD(int(c.Timeout.Nanoseconds() / 1E6))
		// wait until a byte arrived or time out
		timeouts.ReadIntervalTimeout = c_MAXDWORD
		timeouts.ReadTotalTimeoutMultiplier = c_MAXDWORD
		timeouts.ReadTotalTimeoutConstant = timeout
		timeouts.WriteTotalTimeoutConstant = timeout
	}
	err := GetCommTimeouts(p.handle, &p.oldTimeouts)
	if err != nil {
		return err
	}
	err = SetCommTimeouts(p.handle, &timeouts)
	if err != nil {
		// reset
		SetCommTimeouts(p.handle, &p.oldTimeouts)
	}
	return err
}

func (p *port) setSerialConfig(c *Config) error {
	var dcb c_DCB
	if c.BaudRate == 0 {
		dcb.BaudRate = 19200
	} else {
		dcb.BaudRate = toDWORD(c.BaudRate)
	}
	// Data bits
	if c.DataBits == 0 {
		dcb.ByteSize = 8
	} else {
		dcb.ByteSize = toBYTE(c.DataBits)
	}
	// Stop bits
	switch c.StopBits {
	case 0, 1:
		// Default is one stop bit.
		dcb.StopBits = c_ONESTOPBIT
	case 2:
		dcb.StopBits = c_TWOSTOPBITS
	default:
		return fmt.Errorf("serial: unsupported stop bits %v", c.StopBits)
	}
	// Parity
	switch c.Parity {
	case "", "E":
		// Default parity mode is Even.
		dcb.Parity = c_EVENPARITY
		dcb.Pad_cgo_0[0] |= 0x02 // fParity
	case "O":
		dcb.Parity = c_ODDPARITY
		dcb.Pad_cgo_0[0] |= 0x02 // fParity
	case "N":
		dcb.Parity = c_NOPARITY
	default:
		return fmt.Errorf("serial: unsupported parity %v", c.Parity)
	}
	dcb.Pad_cgo_0[0] |= 0x01 // fBinary

	err := GetCommState(p.handle, &p.oldDCB)
	if err != nil {
		return err
	}
	err = SetCommState(p.handle, &dcb)
	if err != nil {
		SetCommState(p.handle, &p.oldDCB)
	}
	return err
}

func newHandle(c *Config) (handle syscall.Handle, err error) {
	handle, err = syscall.CreateFile(
		syscall.StringToUTF16Ptr(c.Address),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,   // mode
		nil, // security
		syscall.OPEN_EXISTING, // create mode
		0, // attributes
		0) // templates
	return
}
