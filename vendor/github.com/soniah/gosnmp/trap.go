// Copyright 2012-2020 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

package gosnmp

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//
// Sending Traps ie GoSNMP acting as an Agent
//

// SendTrap sends a SNMP Trap (v2c/v3 only)
//
// pdus[0] can a pdu of Type TimeTicks (with the desired uint32 epoch
// time).  Otherwise a TimeTicks pdu will be prepended, with time set to
// now. This mirrors the behaviour of the Net-SNMP command-line tools.
//
// SendTrap doesn't wait for a return packet from the NMS (Network
// Management Station).
//
// See also Listen() and examples for creating an NMS.
func (x *GoSNMP) SendTrap(trap SnmpTrap) (result *SnmpPacket, err error) {
	var pdutype PDUType

	if len(trap.Variables) == 0 {
		return nil, fmt.Errorf("SendTrap requires at least 1 PDU")
	}

	if trap.Variables[0].Type == TimeTicks {
		// check is uint32
		if _, ok := trap.Variables[0].Value.(uint32); !ok {
			return nil, fmt.Errorf("SendTrap TimeTick must be uint32")
		}
	}

	switch x.Version {
	case Version2c, Version3:
		pdutype = SNMPv2Trap

		if trap.Variables[0].Type != TimeTicks {
			now := uint32(time.Now().Unix())
			timetickPDU := SnmpPDU{"1.3.6.1.2.1.1.3.0", TimeTicks, now, x.Logger}
			// prepend timetickPDU
			trap.Variables = append([]SnmpPDU{timetickPDU}, trap.Variables...)
		}

	case Version1:
		pdutype = Trap
		if len(trap.Enterprise) == 0 {
			return nil, fmt.Errorf("SendTrap for SNMPV1 requires an Enterprise OID")
		}
		if len(trap.AgentAddress) == 0 {
			return nil, fmt.Errorf("SendTrap for SNMPV1 requires an Agent Address")
		}

	default:
		err = fmt.Errorf("SendTrap doesn't support %s", x.Version)
		return nil, err
	}

	packetOut := x.mkSnmpPacket(pdutype, trap.Variables, 0, 0)
	if x.Version == Version1 {
		packetOut.Enterprise = trap.Enterprise
		packetOut.AgentAddress = trap.AgentAddress
		packetOut.GenericTrap = trap.GenericTrap
		packetOut.SpecificTrap = trap.SpecificTrap
		packetOut.Timestamp = trap.Timestamp
	}

	// all sends wait for the return packet, except for SNMPv2Trap
	// -> wait is false
	return x.send(packetOut, false)
}

//
// Receiving Traps ie GoSNMP acting as an NMS (Network Management
// Station).
//
// GoSNMP.unmarshal() currently only handles SNMPv2Trap (ie v2c, v3)
//

// A TrapListener defines parameters for running a SNMP Trap receiver.
// nil values will be replaced by default values.
type TrapListener struct {
	sync.Mutex
	OnNewTrap func(s *SnmpPacket, u *net.UDPAddr)
	Params    *GoSNMP

	// These unexported fields are for letting test cases
	// know we are ready.
	conn  *net.UDPConn
	proto string

	finish    int32 // Atomic flag; set to 1 when closing connection
	done      chan bool
	listening chan bool
}

// NewTrapListener returns an initialized TrapListener.
func NewTrapListener() *TrapListener {
	tl := &TrapListener{}
	tl.finish = 0
	tl.done = make(chan bool)
	// Buffered because one doesn't have to block on it.
	tl.listening = make(chan bool, 1)
	return tl
}

// Listening returns a sentinel channel on which one can block
// until the listener is ready to receive requests.
func (t *TrapListener) Listening() <-chan bool {
	t.Lock()
	defer t.Unlock()
	return t.listening
}

// Close terminates the listening on TrapListener socket
func (t *TrapListener) Close() {
	// Prevent concurrent calls to Close
	if atomic.CompareAndSwapInt32(&t.finish, 0, 1) {
		if t.conn.LocalAddr().Network() == "udp" {
			t.conn.Close()
		}
		<-t.done
	}
}

func (t *TrapListener) listenUDP(addr string) error {
	// udp

	udpAddr, err := net.ResolveUDPAddr(t.proto, addr)
	if err != nil {
		return err
	}
	t.conn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	defer t.conn.Close()

	// Mark that we are listening now.
	t.listening <- true

	for {
		switch {
		case atomic.LoadInt32(&t.finish) == 1:
			t.done <- true
			return nil

		default:
			var buf [4096]byte
			rlen, remote, err := t.conn.ReadFromUDP(buf[:])
			if err != nil {
				if atomic.LoadInt32(&t.finish) == 1 {
					// err most likely comes from reading from a closed connection
					continue
				}
				t.Params.logPrintf("TrapListener: error in read %s\n", err)
				continue
			}

			msg := buf[:rlen]
			traps := t.Params.UnmarshalTrap(msg)
			if traps != nil {
				t.OnNewTrap(traps, remote)
			}
		}
	}
}

func (t *TrapListener) handleTCPRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 4096)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		t.Params.logPrintf("TrapListener: error in read %s\n", err)
		return
	}

	//fmt.Printf("TEST: handleTCPRequest:%s, %s", t.proto, conn.RemoteAddr())

	msg := buf[:reqLen]
	traps := t.Params.UnmarshalTrap(msg)

	if traps != nil {
		// TODO: lieing for backward compatibility reason - create UDP Address ... not nice
		r, _ := net.ResolveUDPAddr("", conn.RemoteAddr().String())
		t.OnNewTrap(traps, r)
	}
	// Close the connection when you're done with it.
	conn.Close()
}

func (t *TrapListener) listenTCP(addr string) error {
	// udp

	tcpAddr, err := net.ResolveTCPAddr(t.proto, addr)
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	defer l.Close()

	// Mark that we are listening now.
	t.listening <- true

	for {

		switch {
		case atomic.LoadInt32(&t.finish) == 1:
			t.done <- true
			return nil
		default:

			// Listen for an incoming connection.
			conn, err := l.Accept()
			fmt.Printf("ACCEPT: %s", conn)
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				os.Exit(1)
			}
			// Handle connections in a new goroutine.
			go t.handleTCPRequest(conn)
		}
	}
}

// Listen listens on the UDP address addr and calls the OnNewTrap
// function specified in *TrapListener for every trap received.
func (t *TrapListener) Listen(addr string) error {
	if t.Params == nil {
		t.Params = Default
	}

	t.Params.validateParameters()
	/*
		TODO returning an error causes TestSendTrapBasic() (and others) to hang
		err := t.Params.validateParameters()
		if err != nil {
			return err
		}
	*/

	if t.OnNewTrap == nil {
		t.OnNewTrap = debugTrapHandler
	}

	splitted := strings.SplitN(addr, "://", 2)
	t.proto = "udp"
	if len(splitted) > 1 {
		t.proto = splitted[0]
		addr = splitted[1]
	}

	//fmt.Printf("TEST: Adress:%s, %s", t.proto, addr)

	if t.proto == "tcp" {
		return t.listenTCP(addr)
	} else if t.proto == "udp" {
		return t.listenUDP(addr)
	}

	return fmt.Errorf("Not implemented network protocol: %s [use: tcp/udp]", t.proto)
}

// Default trap handler
func debugTrapHandler(s *SnmpPacket, u *net.UDPAddr) {
	log.Printf("got trapdata from %+v: %+v\n", u, s)
}

// UnmarshalTrap unpacks the SNMP Trap.
func (x *GoSNMP) UnmarshalTrap(trap []byte) (result *SnmpPacket) {
	result = new(SnmpPacket)

	if x.SecurityParameters != nil {
		result.SecurityParameters = x.SecurityParameters.Copy()
	}

	cursor, err := x.unmarshalHeader(trap, result)
	if err != nil {
		x.logPrintf("UnmarshalTrap: %s\n", err)
		return nil
	}

	if result.Version == Version3 {
		if result.SecurityModel == UserSecurityModel {
			err = x.testAuthentication(trap, result)
			if err != nil {
				x.logPrintf("UnmarshalTrap v3 auth: %s\n", err)
				return nil
			}
		}
		trap, cursor, err = x.decryptPacket(trap, cursor, result)
		if err != nil {
			x.logPrintf("UnmarshalTrap v3 decrypt: %s\n", err)
			return nil
		}
	}
	err = x.unmarshalPayload(trap, cursor, result)
	if err != nil {
		x.logPrintf("UnmarshalTrap: %s\n", err)
		return nil
	}
	return result
}
