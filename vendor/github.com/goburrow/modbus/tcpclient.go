// Copyright 2014 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	tcpProtocolIdentifier uint16 = 0x0000

	// Modbus Application Protocol
	tcpHeaderSize = 7
	tcpMaxLength  = 260
	// Default TCP timeout is not set
	tcpTimeout     = 10 * time.Second
	tcpIdleTimeout = 60 * time.Second
)

// TCPClientHandler implements Packager and Transporter interface.
type TCPClientHandler struct {
	tcpPackager
	tcpTransporter
}

// NewTCPClientHandler allocates a new TCPClientHandler.
func NewTCPClientHandler(address string) *TCPClientHandler {
	h := &TCPClientHandler{}
	h.Address = address
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout
	return h
}

// TCPClient creates TCP client with default handler and given connect string.
func TCPClient(address string) Client {
	handler := NewTCPClientHandler(address)
	return NewClient(handler)
}

// tcpPackager implements Packager interface.
type tcpPackager struct {
	// For synchronization between messages of server & client
	transactionId uint32
	// Broadcast address is 0
	SlaveId byte
}

// Encode adds modbus application protocol header:
//  Transaction identifier: 2 bytes
//  Protocol identifier: 2 bytes
//  Length: 2 bytes
//  Unit identifier: 1 byte
//  Function code: 1 byte
//  Data: n bytes
func (mb *tcpPackager) Encode(pdu *ProtocolDataUnit) (adu []byte, err error) {
	adu = make([]byte, tcpHeaderSize+1+len(pdu.Data))

	// Transaction identifier
	transactionId := atomic.AddUint32(&mb.transactionId, 1)
	binary.BigEndian.PutUint16(adu, uint16(transactionId))
	// Protocol identifier
	binary.BigEndian.PutUint16(adu[2:], tcpProtocolIdentifier)
	// Length = sizeof(SlaveId) + sizeof(FunctionCode) + Data
	length := uint16(1 + 1 + len(pdu.Data))
	binary.BigEndian.PutUint16(adu[4:], length)
	// Unit identifier
	adu[6] = mb.SlaveId

	// PDU
	adu[tcpHeaderSize] = pdu.FunctionCode
	copy(adu[tcpHeaderSize+1:], pdu.Data)
	return
}

// Verify confirms transaction, protocol and unit id.
func (mb *tcpPackager) Verify(aduRequest []byte, aduResponse []byte) (err error) {
	// Transaction id
	responseVal := binary.BigEndian.Uint16(aduResponse)
	requestVal := binary.BigEndian.Uint16(aduRequest)
	if responseVal != requestVal {
		err = fmt.Errorf("modbus: response transaction id '%v' does not match request '%v'", responseVal, requestVal)
		return
	}
	// Protocol id
	responseVal = binary.BigEndian.Uint16(aduResponse[2:])
	requestVal = binary.BigEndian.Uint16(aduRequest[2:])
	if responseVal != requestVal {
		err = fmt.Errorf("modbus: response protocol id '%v' does not match request '%v'", responseVal, requestVal)
		return
	}
	// Unit id (1 byte)
	if aduResponse[6] != aduRequest[6] {
		err = fmt.Errorf("modbus: response unit id '%v' does not match request '%v'", aduResponse[6], aduRequest[6])
		return
	}
	return
}

// Decode extracts PDU from TCP frame:
//  Transaction identifier: 2 bytes
//  Protocol identifier: 2 bytes
//  Length: 2 bytes
//  Unit identifier: 1 byte
func (mb *tcpPackager) Decode(adu []byte) (pdu *ProtocolDataUnit, err error) {
	// Read length value in the header
	length := binary.BigEndian.Uint16(adu[4:])
	pduLength := len(adu) - tcpHeaderSize
	if pduLength <= 0 || pduLength != int(length-1) {
		err = fmt.Errorf("modbus: length in response '%v' does not match pdu data length '%v'", length-1, pduLength)
		return
	}
	pdu = &ProtocolDataUnit{}
	// The first byte after header is function code
	pdu.FunctionCode = adu[tcpHeaderSize]
	pdu.Data = adu[tcpHeaderSize+1:]
	return
}

// tcpTransporter implements Transporter interface.
type tcpTransporter struct {
	// Connect string
	Address string
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
	// Transmission logger
	Logger *log.Logger

	// TCP connection
	mu           sync.Mutex
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time
}

// Send sends data to server and ensures response length is greater than header length.
func (mb *tcpTransporter) Send(aduRequest []byte) (aduResponse []byte, err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Establish a new connection if not connected
	if err = mb.connect(); err != nil {
		return
	}
	// Set timer to close when idle
	mb.lastActivity = time.Now()
	mb.startCloseTimer()
	// Set write and read timeout
	var timeout time.Time
	if mb.Timeout > 0 {
		timeout = mb.lastActivity.Add(mb.Timeout)
	}
	if err = mb.conn.SetDeadline(timeout); err != nil {
		return
	}
	// Send data
	mb.logf("modbus: sending % x", aduRequest)
	if _, err = mb.conn.Write(aduRequest); err != nil {
		return
	}
	// Read header first
	var data [tcpMaxLength]byte
	if _, err = io.ReadFull(mb.conn, data[:tcpHeaderSize]); err != nil {
		return
	}
	// Read length, ignore transaction & protocol id (4 bytes)
	length := int(binary.BigEndian.Uint16(data[4:]))
	if length <= 0 {
		mb.flush(data[:])
		err = fmt.Errorf("modbus: length in response header '%v' must not be zero", length)
		return
	}
	if length > (tcpMaxLength - (tcpHeaderSize - 1)) {
		mb.flush(data[:])
		err = fmt.Errorf("modbus: length in response header '%v' must not greater than '%v'", length, tcpMaxLength-tcpHeaderSize+1)
		return
	}
	// Skip unit id
	length += tcpHeaderSize - 1
	if _, err = io.ReadFull(mb.conn, data[tcpHeaderSize:length]); err != nil {
		return
	}
	aduResponse = data[:length]
	mb.logf("modbus: received % x\n", aduResponse)
	return
}

// Connect establishes a new connection to the address in Address.
// Connect and Close are exported so that multiple requests can be done with one session
func (mb *tcpTransporter) Connect() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return mb.connect()
}

func (mb *tcpTransporter) connect() error {
	if mb.conn == nil {
		dialer := net.Dialer{Timeout: mb.Timeout}
		conn, err := dialer.Dial("tcp", mb.Address)
		if err != nil {
			return err
		}
		mb.conn = conn
	}
	return nil
}

func (mb *tcpTransporter) startCloseTimer() {
	if mb.IdleTimeout <= 0 {
		return
	}
	if mb.closeTimer == nil {
		mb.closeTimer = time.AfterFunc(mb.IdleTimeout, mb.closeIdle)
	} else {
		mb.closeTimer.Reset(mb.IdleTimeout)
	}
}

// Close closes current connection.
func (mb *tcpTransporter) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return mb.close()
}

// flush flushes pending data in the connection,
// returns io.EOF if connection is closed.
func (mb *tcpTransporter) flush(b []byte) (err error) {
	if err = mb.conn.SetReadDeadline(time.Now()); err != nil {
		return
	}
	// Timeout setting will be reset when reading
	if _, err = mb.conn.Read(b); err != nil {
		// Ignore timeout error
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			err = nil
		}
	}
	return
}

func (mb *tcpTransporter) logf(format string, v ...interface{}) {
	if mb.Logger != nil {
		mb.Logger.Printf(format, v...)
	}
}

// closeLocked closes current connection. Caller must hold the mutex before calling this method.
func (mb *tcpTransporter) close() (err error) {
	if mb.conn != nil {
		err = mb.conn.Close()
		mb.conn = nil
	}
	return
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (mb *tcpTransporter) closeIdle() {
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
