// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package modbus

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/goburrow/serial"
)

const (
	// Default timeout
	serialTimeout     = 5 * time.Second
	serialIdleTimeout = 60 * time.Second
)

// serialPort has configuration and I/O controller.
type serialPort struct {
	// Serial port configuration.
	serial.Config

	Logger      *log.Logger
	IdleTimeout time.Duration

	mu sync.Mutex
	// port is platform-dependent data structure for serial port.
	port         io.ReadWriteCloser
	lastActivity time.Time
	closeTimer   *time.Timer
}

func (mb *serialPort) Connect() (err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return mb.connect()
}

// connect connects to the serial port if it is not connected. Caller must hold the mutex.
func (mb *serialPort) connect() error {
	if mb.port == nil {
		port, err := serial.Open(&mb.Config)
		if err != nil {
			return err
		}
		mb.port = port
	}
	return nil
}

func (mb *serialPort) Close() (err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return mb.close()
}

// close closes the serial port if it is connected. Caller must hold the mutex.
func (mb *serialPort) close() (err error) {
	if mb.port != nil {
		err = mb.port.Close()
		mb.port = nil
	}
	return
}

func (mb *serialPort) logf(format string, v ...interface{}) {
	if mb.Logger != nil {
		mb.Logger.Printf(format, v...)
	}
}

func (mb *serialPort) startCloseTimer() {
	if mb.IdleTimeout <= 0 {
		return
	}
	if mb.closeTimer == nil {
		mb.closeTimer = time.AfterFunc(mb.IdleTimeout, mb.closeIdle)
	} else {
		mb.closeTimer.Reset(mb.IdleTimeout)
	}
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (mb *serialPort) closeIdle() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(mb.lastActivity)
	if idle >= mb.IdleTimeout {
		mb.logf("modbus: closing connection due to idle timeout: %v", idle)
		mb.close()
	}
}
