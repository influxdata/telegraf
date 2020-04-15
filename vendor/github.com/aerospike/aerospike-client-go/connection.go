// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"crypto/tls"
	"io"
	"net"
	"strconv"
	"time"

	. "github.com/aerospike/aerospike-client-go/logger"
	. "github.com/aerospike/aerospike-client-go/types"
)

// Connection represents a connection with a timeout.
type Connection struct {
	node *Node

	// timeout
	timeout time.Duration

	// duration after which connection is considered idle
	idleTimeout  time.Duration
	idleDeadline time.Time

	// connection object
	conn net.Conn

	// to avoid having a buffer pool and contention
	dataBuffer []byte
}

func errToTimeoutErr(err error) error {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return NewAerospikeError(TIMEOUT, err.Error())
	}
	return err
}

func shouldClose(err error) bool {
	if err == io.EOF {
		return true
	}

	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}

	return false
}

// NewConnection creates a connection on the network and returns the pointer
// A minimum timeout of 2 seconds will always be applied.
// If the connection is not established in the specified timeout,
// an error will be returned
func NewConnection(address string, timeout time.Duration) (*Connection, error) {
	newConn := &Connection{dataBuffer: make([]byte, 1024)}

	// don't wait indefinitely
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		Logger.Error("Connection to address `" + address + "` failed to establish with error: " + err.Error())
		return nil, errToTimeoutErr(err)
	}
	newConn.conn = conn

	// set timeout at the last possible moment
	if err := newConn.SetTimeout(timeout); err != nil {
		return nil, err
	}
	return newConn, nil
}

// NewSecureConnection creates a TLS connection on the network and returns the pointer.
// A minimum timeout of 2 seconds will always be applied.
// If the connection is not established in the specified timeout,
// an error will be returned
func NewSecureConnection(policy *ClientPolicy, host *Host) (*Connection, error) {
	address := net.JoinHostPort(host.Name, strconv.Itoa(host.Port))
	conn, err := NewConnection(address, policy.Timeout)
	if err != nil {
		return nil, err
	}

	if policy.TlsConfig == nil {
		return conn, nil
	}

	// Use version dependent clone function to clone the config
	tlsConfig := cloneTlsConfig(policy.TlsConfig)
	tlsConfig.ServerName = host.TLSName

	sconn := tls.Client(conn.conn, tlsConfig)
	if err := sconn.Handshake(); err != nil {
		sconn.Close()
		return nil, err
	}

	if host.TLSName != "" && !tlsConfig.InsecureSkipVerify {
		if err := sconn.VerifyHostname(host.TLSName); err != nil {
			sconn.Close()
			Logger.Error("Connection to address `" + address + "` failed to establish with error: " + err.Error())
			return nil, errToTimeoutErr(err)
		}
	}

	conn.conn = sconn
	return conn, nil
}

// Write writes the slice to the connection buffer.
func (ctn *Connection) Write(buf []byte) (total int, err error) {
	// make sure all bytes are written
	// Don't worry about the loop, timeout has been set elsewhere
	length := len(buf)
	var r int
	for total < length {
		if r, err = ctn.conn.Write(buf[total:]); err != nil {
			break
		}
		total += r
	}

	if err == nil {
		return total, nil
	}
	return total, errToTimeoutErr(err)
}

// ReadN reads N bytes from connection buffer to the provided Writer.
func (ctn *Connection) ReadN(buf io.Writer, length int64) (total int64, err error) {
	// if all bytes are not read, retry until successful
	// Don't worry about the loop; we've already set the timeout elsewhere
	total, err = io.CopyN(buf, ctn.conn, length)

	if err == nil && total == length {
		return total, nil
	} else if err != nil {
		if shouldClose(err) {
			ctn.Close()
		}
		return total, errToTimeoutErr(err)
	}
	ctn.Close()
	return total, NewAerospikeError(SERVER_ERROR)
}

// Read reads from connection buffer to the provided slice.
func (ctn *Connection) Read(buf []byte, length int) (total int, err error) {
	// if all bytes are not read, retry until successful
	// Don't worry about the loop; we've already set the timeout elsewhere
	var r int
	for total < length {
		r, err = ctn.conn.Read(buf[total:length])
		total += r
		if err != nil {
			break
		}
	}

	if err == nil && total == length {
		return total, nil
	} else if err != nil {
		if shouldClose(err) {
			ctn.Close()
		}
		return total, errToTimeoutErr(err)
	}
	ctn.Close()
	return total, NewAerospikeError(SERVER_ERROR)
}

// IsConnected returns true if the connection is not closed yet.
func (ctn *Connection) IsConnected() bool {
	return ctn.conn != nil
}

// SetTimeout sets connection timeout for both read and write operations.
func (ctn *Connection) SetTimeout(timeout time.Duration) error {
	// Set timeout ONLY if there is or has been a timeout
	if timeout > 0 || ctn.timeout != 0 {
		ctn.timeout = timeout

		// important: remove deadline when not needed; connections are pooled
		if ctn.conn != nil {
			var deadline time.Time
			if timeout > 0 {
				deadline = time.Now().Add(timeout)
			}
			if err := ctn.conn.SetDeadline(deadline); err != nil {
				return err
			}
		}
	}

	return nil
}

// Close closes the connection
func (ctn *Connection) Close() {
	if ctn != nil && ctn.conn != nil {
		// deregister
		if ctn.node != nil {
			ctn.node.connectionCount.DecrementAndGet()
		}

		if err := ctn.conn.Close(); err != nil {
			Logger.Warn(err.Error())
		}
		ctn.conn = nil
	}
}

// Authenticate will send authentication information to the server.
func (ctn *Connection) Authenticate(user string, password []byte) error {
	// need to authenticate
	if user != "" {
		command := newAdminCommand(ctn.dataBuffer)
		if err := command.authenticate(ctn, user, password); err != nil {
			// Socket not authenticated. Do not put back into pool.
			return err
		}
	}
	return nil
}

// setIdleTimeout sets the idle timeout for the connection.
func (ctn *Connection) setIdleTimeout(timeout time.Duration) {
	ctn.idleTimeout = timeout
}

// isIdle returns true if the connection has reached the idle deadline.
func (ctn *Connection) isIdle() bool {
	return ctn.idleTimeout > 0 && !time.Now().Before(ctn.idleDeadline)
}

// refresh extends the idle deadline of the connection.
func (ctn *Connection) refresh() {
	ctn.idleDeadline = time.Now().Add(ctn.idleTimeout)
}
