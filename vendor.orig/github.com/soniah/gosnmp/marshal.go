// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

package gosnmp

import (
	"bytes"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

//
// Remaining globals and definitions located here.
// See http://www.rane.com/note161.html for a succint description of the SNMP
// protocol.
//

// SnmpVersion 1, 2c and 3 implemented
type SnmpVersion uint8

// SnmpVersion 1, 2c and 3 implemented
const (
	Version1  SnmpVersion = 0x0
	Version2c SnmpVersion = 0x1
	Version3  SnmpVersion = 0x3
)

// SnmpPacket struct represents the entire SNMP Message or Sequence at the
// application layer.
type SnmpPacket struct {
	Version            SnmpVersion
	MsgFlags           SnmpV3MsgFlags
	SecurityModel      SnmpV3SecurityModel
	SecurityParameters SnmpV3SecurityParameters
	ContextEngineID    string
	ContextName        string
	Community          string
	PDUType            PDUType
	MsgID              uint32
	RequestID          uint32
	Error              SNMPError
	ErrorIndex         uint8
	NonRepeaters       uint8
	MaxRepetitions     uint8
	Variables          []SnmpPDU
	Logger             Logger

	// v1 traps have a very different format from v2c and v3 traps.
	//
	// These fields are set via the SnmpTrap parameter to SendTrap().
	SnmpTrap
}

// SnmpTrap is used to define a SNMP trap, and is passed into SendTrap
type SnmpTrap struct {
	Variables []SnmpPDU

	// These fields are required for SNMPV1 Trap Headers
	Enterprise   string
	AgentAddress string
	GenericTrap  int
	SpecificTrap int
	Timestamp    uint
}

// VarBind struct represents an SNMP Varbind.
type VarBind struct {
	Name  asn1.ObjectIdentifier
	Value asn1.RawValue
}

// PDUType describes which SNMP Protocol Data Unit is being sent.
type PDUType byte

// The currently supported PDUType's
const (
	Sequence       PDUType = 0x30
	GetRequest     PDUType = 0xa0
	GetNextRequest PDUType = 0xa1
	GetResponse    PDUType = 0xa2
	SetRequest     PDUType = 0xa3
	Trap           PDUType = 0xa4 // v1
	GetBulkRequest PDUType = 0xa5
	InformRequest  PDUType = 0xa6
	SNMPv2Trap     PDUType = 0xa7 // v2c, v3
	Report         PDUType = 0xa8
)

const rxBufSize = 65535 // max size of IPv4 & IPv6 packet

// Logger is an interface used for debugging. Both Print and
// Printf have the same interfaces as Package Log in the std library. The
// Logger interface is small to give you flexibility in how you do
// your debugging.
//
// For verbose logging to stdout:
//
//     gosnmp_logger = log.New(os.Stdout, "", 0)
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

func (x *GoSNMP) logPrint(v ...interface{}) {
	if x.loggingEnabled {
		x.Logger.Print(v...)
	}
}

func (x *GoSNMP) logPrintf(format string, v ...interface{}) {
	if x.loggingEnabled {
		x.Logger.Printf(format, v...)
	}
}

// send/receive one snmp request
func (x *GoSNMP) sendOneRequest(packetOut *SnmpPacket,
	wait bool) (result *SnmpPacket, err error) {
	finalDeadline := time.Now().Add(x.Timeout)

	allReqIDs := make([]uint32, 0, x.Retries+1)
	allMsgIDs := make([]uint32, 0, x.Retries+1)
	for retries := 0; ; retries++ {
		if retries > 0 {
			x.logPrintf("Retry number %d. Last error was: %v", retries, err)
			if time.Now().After(finalDeadline) {
				err = fmt.Errorf("Request timeout (after %d retries)", retries-1)
				break
			}
			if retries > x.Retries {
				// Report last error
				break
			}
		}
		err = nil

		reqDeadline := time.Now().Add(x.Timeout / time.Duration(x.Retries+1))
		x.Conn.SetDeadline(reqDeadline)

		// Request ID is an atomic counter (started at a random value)
		reqID := atomic.AddUint32(&(x.requestID), 1) // TODO: fix overflows
		allReqIDs = append(allReqIDs, reqID)

		packetOut.RequestID = reqID

		if x.Version == Version3 {
			msgID := atomic.AddUint32(&(x.msgID), 1) // TODO: fix overflows
			allMsgIDs = append(allMsgIDs, msgID)

			packetOut.MsgID = msgID

			err = x.initPacket(packetOut)
			if err != nil {
				break
			}

		}
		if x.loggingEnabled && x.Version == Version3 {
			packetOut.SecurityParameters.Log()
		}

		var outBuf []byte
		outBuf, err = packetOut.marshalMsg()
		if err != nil {
			// Don't retry - not going to get any better!
			err = fmt.Errorf("marshal: %v", err)
			break
		}

		x.logPrintf("SENDING PACKET: %#+v", *packetOut)
		_, err = x.Conn.Write(outBuf)
		if err != nil {
			continue
		}

		// all sends wait for the return packet, except for SNMPv2Trap
		if wait == false {
			return &SnmpPacket{}, nil
		}

	waitingResponse:
		for {
			x.logPrint("WAITING RESPONSE...")
			// Receive response and try receiving again on any decoding error.
			// Let the deadline abort us if we don't receive a valid response.

			var resp []byte
			resp, err = x.receive()
			if err != nil {
				// receive error. retrying won't help. abort
				break
			}
			x.logPrint("GET RESPONSE OK : %+v", resp)
			result = new(SnmpPacket)
			result.Logger = x.Logger

			result.MsgFlags = packetOut.MsgFlags
			if packetOut.SecurityParameters != nil {
				result.SecurityParameters = packetOut.SecurityParameters.Copy()
			}

			var cursor int
			cursor, err = x.unmarshalHeader(resp, result)
			if err != nil {
				x.logPrintf("ERROR on unmarshall header: %s", err)
				err = fmt.Errorf("Unable to decode packet: %s", err.Error())
				continue
			}

			if x.Version == Version3 {
				err = x.testAuthentication(resp, result)
				if err != nil {
					x.logPrintf("ERROR on Test Authentication on v3: %s", err)
					break
				}
				resp, cursor, err = x.decryptPacket(resp, cursor, result)
			}

			err = x.unmarshalPayload(resp, cursor, result)
			if err != nil {
				x.logPrintf("ERROR on UnmarshalPayload on v3: %s", err)
				err = fmt.Errorf("Unable to decode packet: %s", err.Error())
				continue
			}
			if result == nil || len(result.Variables) < 1 {
				x.logPrintf("ERROR on UnmarshalPayload on v3: %s", err)
				err = fmt.Errorf("Unable to decode packet: nil")
				continue
			}

			// Detect usmStats report PDUs and go out of this function with all data
			// (usmStatsNotInTimeWindows [1.3.6.1.6.3.15.1.1.2.0] will be handled by the calling
			// function, and retransmitted.  All others need to be handled by user code)
			if result.Version == Version3 && len(result.Variables) == 1 && result.PDUType == Report {
				switch result.Variables[0].Name {
				case ".1.3.6.1.6.3.15.1.1.1.0", ".1.3.6.1.6.3.15.1.1.2.0",
					".1.3.6.1.6.3.15.1.1.3.0", ".1.3.6.1.6.3.15.1.1.4.0",
					".1.3.6.1.6.3.15.1.1.5.0", ".1.3.6.1.6.3.15.1.1.6.0":
					break waitingResponse
				}
			}

			validID := false
			for _, id := range allReqIDs {
				if id == result.RequestID {
					validID = true
				}
			}
			if result.RequestID == 0 {
				validID = true
			}
			if !validID {
				x.logPrint("ERROR  out of order ")
				err = fmt.Errorf("Out of order response")
				continue
			}

			break
		}
		if err != nil {
			continue
		}

		// Success!
		return result, nil
	}

	// Return last error
	return nil, err
}

// generic "sender" that negotiate any version of snmp request
//
// all sends wait for the return packet, except for SNMPv2Trap
func (x *GoSNMP) send(packetOut *SnmpPacket, wait bool) (result *SnmpPacket, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("recover: %v", e)
		}
	}()

	if x.Conn == nil {
		return nil, fmt.Errorf("&GoSNMP.Conn is missing. Provide a connection or use Connect()")
	}

	if x.Retries < 0 {
		x.Retries = 0
	}
	x.logPrint("SEND INIT")
	if packetOut.Version == Version3 {
		x.logPrint("SEND INIT NEGOTIATE SECURITY PARAMS")
		if err = x.negotiateInitialSecurityParameters(packetOut, wait); err != nil {
			return &SnmpPacket{}, err
		}
		x.logPrint("SEND END NEGOTIATE SECURITY PARAMS")
	}

	// perform request
	result, err = x.sendOneRequest(packetOut, wait)
	if err != nil {
		x.logPrintf("SEND Error on the first Request Error: %s", err)
		return result, err
	}

	if result.Version == Version3 {
		x.logPrintf("SEND STORE SECURITY PARAMS from result: %+v", result)
		err = x.storeSecurityParameters(result)

		// detect out-of-time-window error and retransmit with updated auth engine parameters
		if len(result.Variables) == 1 && result.Variables[0].Name == ".1.3.6.1.6.3.15.1.1.2.0" {
			x.logPrintf("WARNING detected out-of-time-window ERROR")
			err = x.updatePktSecurityParameters(packetOut)
			if err != nil {
				x.logPrintf("ERROR  updatePktSecurityParameters error: %s", err)
				return nil, err
			}
			result, err = x.sendOneRequest(packetOut, wait)
		}
	}

	// detect unknown engine id error and retransmit with updated engine id
	if len(result.Variables) == 1 && result.Variables[0].Name == ".1.3.6.1.6.3.15.1.1.4.0" {
		x.logPrintf("WARNING detected unknown enginer id ERROR")
		err = x.updatePktSecurityParameters(packetOut)
		if err != nil {
			x.logPrintf("ERROR  updatePktSecurityParameters error: %s", err)
			return nil, err
		}
		result, err = x.sendOneRequest(packetOut, wait)
	}
	return result, err
}

// -- Marshalling Logic --------------------------------------------------------
func (packet *SnmpPacket) MarshalMsg() ([]byte, error) {
	return packet.marshalMsg()
}

// marshal an SNMP message
func (packet *SnmpPacket) marshalMsg() ([]byte, error) {
	var err error
	buf := new(bytes.Buffer)

	// version
	buf.Write([]byte{2, 1, byte(packet.Version)})

	if packet.Version == Version3 {
		buf, err = packet.marshalV3(buf)
		if err != nil {
			return nil, err
		}
	} else {
		// community
		buf.Write([]byte{4, uint8(len(packet.Community))})
		buf.WriteString(packet.Community)
		// pdu
		pdu, err := packet.marshalPDU()
		if err != nil {
			return nil, err
		}
		buf.Write(pdu)
	}

	// build up resulting msg - sequence, length then the tail (buf)
	msg := new(bytes.Buffer)
	msg.WriteByte(byte(Sequence))

	bufLengthBytes, err2 := marshalLength(buf.Len())
	if err2 != nil {
		return nil, err2
	}
	msg.Write(bufLengthBytes)
	buf.WriteTo(msg) // reverse logic - want to do msg.Write(buf)

	authenticatedMessage, err := packet.authenticate(msg.Bytes())
	if err != nil {
		return nil, err
	}

	return authenticatedMessage, nil
}

func (packet *SnmpPacket) marshalSNMPV1TrapHeader() ([]byte, error) {
	buf := new(bytes.Buffer)

	mOid, err := marshalOID(packet.Enterprise)
	if err != nil {
		return nil, fmt.Errorf("Unable to marshal OID: %s\n", err.Error())
	}

	buf.Write([]byte{ObjectIdentifier, byte(len(mOid))})
	buf.Write(mOid)

	// write IPAddress type, length and ipAddress value
	ip := net.ParseIP(packet.AgentAddress)
	ipAddressBytes := ipv4toBytes(ip)
	buf.Write([]byte{IPAddress, byte(len(ipAddressBytes))})
	buf.Write(ipAddressBytes)

	buf.Write([]byte{Integer, 1})
	buf.WriteByte(byte(packet.GenericTrap))

	buf.Write([]byte{Integer, 1})
	buf.WriteByte(byte(packet.SpecificTrap))

	timeTicks, e := marshalUint32(uint32(packet.Timestamp))
	if e != nil {
		return nil, fmt.Errorf("Unable to Timestamp: %s\n", e.Error())
	}

	buf.Write([]byte{TimeTicks, byte(len(timeTicks))})
	buf.Write(timeTicks)

	return buf.Bytes(), nil
}

// marshal a PDU
func (packet *SnmpPacket) marshalPDU() ([]byte, error) {
	buf := new(bytes.Buffer)

	switch packet.PDUType {

	case GetBulkRequest:
		// requestid
		buf.Write([]byte{2, 4})
		err := binary.Write(buf, binary.BigEndian, packet.RequestID)
		if err != nil {
			return nil, err
		}

		// non repeaters
		buf.Write([]byte{2, 1, packet.NonRepeaters})

		// max repetitions
		buf.Write([]byte{2, 1, packet.MaxRepetitions})

	case Trap:
		// write SNMP V1 Trap Header fields
		snmpV1TrapHeader, err := packet.marshalSNMPV1TrapHeader()
		if err != nil {
			return nil, err
		}

		buf.Write(snmpV1TrapHeader)
	default:
		// requestid
		buf.Write([]byte{2, 4})
		err := binary.Write(buf, binary.BigEndian, packet.RequestID)

		if err != nil {
			return nil, fmt.Errorf("Unable to marshal OID: %s\n", err.Error())
		}

		// error
		buf.Write([]byte{2, 1, 0})

		// error index
		buf.Write([]byte{2, 1, 0})

	}

	// varbind list
	vbl, err := packet.marshalVBL()
	if err != nil {
		return nil, err
	}
	buf.Write(vbl)

	// build up resulting pdu - request type, length, then the tail (buf)
	pdu := new(bytes.Buffer)
	pdu.WriteByte(byte(packet.PDUType))

	bufLengthBytes, err2 := marshalLength(buf.Len())
	if err2 != nil {
		return nil, err2
	}
	pdu.Write(bufLengthBytes)

	buf.WriteTo(pdu) // reverse logic - want to do pdu.Write(buf)
	return pdu.Bytes(), nil
}

// marshal a varbind list
func (packet *SnmpPacket) marshalVBL() ([]byte, error) {

	vblBuf := new(bytes.Buffer)
	for _, pdu := range packet.Variables {
		vb, err := marshalVarbind(&pdu)
		if err != nil {
			return nil, err
		}
		vblBuf.Write(vb)
	}

	vblBytes := vblBuf.Bytes()
	vblLengthBytes, err := marshalLength(len(vblBytes))
	if err != nil {
		return nil, err
	}

	// FIX does bytes.Buffer give better performance than byte slices?
	result := []byte{byte(Sequence)}
	result = append(result, vblLengthBytes...)
	result = append(result, vblBytes...)
	return result, nil
}

// marshal a varbind
func marshalVarbind(pdu *SnmpPDU) ([]byte, error) {
	oid, err := marshalOID(pdu.Name)
	if err != nil {
		return nil, err
	}
	pduBuf := new(bytes.Buffer)
	tmpBuf := new(bytes.Buffer)

	// Marshal the PDU type into the appropriate BER
	switch pdu.Type {

	case Null:
		ltmp, err := marshalLength(len(oid))
		if err != nil {
			return nil, err
		}
		tmpBuf.Write([]byte{byte(ObjectIdentifier)})
		tmpBuf.Write(ltmp)
		tmpBuf.Write(oid)
		tmpBuf.Write([]byte{Null, 0x00})

		ltmp, err = marshalLength(tmpBuf.Len())
		if err != nil {
			return nil, err
		}
		pduBuf.Write([]byte{byte(Sequence)})
		pduBuf.Write(ltmp)
		tmpBuf.WriteTo(pduBuf)

	/*
		snmp Integer32 and INTEGER:
		-2^31 and 2^31-1 inclusive (-2147483648 to 2147483647 decimal)
		(FYI https://groups.google.com/forum/#!topic/comp.protocols.snmp/1xaAMzCe_hE)

		snmp Counter32, Gauge32, TimeTicks, Unsigned32:
		non-negative integer, maximum value of 2^32-1 (4294967295 decimal)
	*/

	case Integer:
		// TODO tests currently only cover positive integers

		// Oid
		tmpBuf.Write([]byte{byte(ObjectIdentifier), byte(len(oid))})
		tmpBuf.Write(oid)

		// Number
		var intBytes []byte
		switch value := pdu.Value.(type) {
		case byte:
			intBytes = []byte{byte(pdu.Value.(int))}
		case int:
			intBytes, err = marshalInt16(value)
			pdu.Check(err)
		default:
			return nil, fmt.Errorf("Unable to marshal PDU Integer; not byte or int.")
		}
		tmpBuf.Write([]byte{byte(Integer), byte(len(intBytes))})
		tmpBuf.Write(intBytes)

		// Sequence, length of oid + integer, then oid/integer data
		pduBuf.WriteByte(byte(Sequence))
		pduBuf.WriteByte(byte(len(oid) + len(intBytes) + 4))
		pduBuf.Write(tmpBuf.Bytes())

	case Counter32, Gauge32, TimeTicks, Uinteger32:
		// Oid
		tmpBuf.Write([]byte{byte(ObjectIdentifier), byte(len(oid))})
		tmpBuf.Write(oid)

		// Number
		var intBytes []byte
		switch value := pdu.Value.(type) {
		case uint32:
			intBytes, err = marshalUint32(value)
			pdu.Check(err)
		default:
			return nil, fmt.Errorf("Unable to marshal pdu.Type %v; unknown pdu.Value %v", pdu.Type, pdu.Value)
		}
		tmpBuf.Write([]byte{byte(pdu.Type), byte(len(intBytes))})
		tmpBuf.Write(intBytes)

		// Sequence, length of oid + integer, then oid/integer data
		pduBuf.WriteByte(byte(Sequence))
		pduBuf.WriteByte(byte(len(oid) + len(intBytes) + 4))
		pduBuf.Write(tmpBuf.Bytes())

	case OctetString:

		//Oid
		tmpBuf.Write([]byte{byte(ObjectIdentifier), byte(len(oid))})
		tmpBuf.Write(oid)

		//OctetString
		var octetStringBytes []byte
		switch value := pdu.Value.(type) {
		case []byte:
			octetStringBytes = value
		case string:
			octetStringBytes = []byte(value)
		default:
			return nil, fmt.Errorf("Unable to marshal PDU OctetString; not []byte or String.")
		}

		var length []byte
		length, err = marshalLength(len(octetStringBytes))
		if err != nil {
			return nil, err
		}
		tmpBuf.WriteByte(byte(OctetString))
		tmpBuf.Write(length)
		tmpBuf.Write(octetStringBytes)

		tmpBytes := tmpBuf.Bytes()

		length, err = marshalLength(len(tmpBytes))
		if err != nil {
			return nil, err
		}
		// Sequence, length of oid + octetstring, then oid/octetstring data
		pduBuf.WriteByte(byte(Sequence))

		pduBuf.Write(length)
		pduBuf.Write(tmpBytes)

	case ObjectIdentifier:

		//Oid
		tmpBuf.Write([]byte{byte(ObjectIdentifier), byte(len(oid))})
		tmpBuf.Write(oid)
		value := pdu.Value.(string)
		oidBytes, err := marshalOID(value)
		pdu.Check(err)

		//Oid data
		var length []byte
		length, err = marshalLength(len(oidBytes))
		if err != nil {
			return nil, err
		}
		tmpBuf.WriteByte(byte(pdu.Type))
		tmpBuf.Write(length)
		tmpBuf.Write(oidBytes)

		tmpBytes := tmpBuf.Bytes()
		length, err = marshalLength(len(tmpBytes))
		if err != nil {
			return nil, err
		}
		// Sequence, length of oid + oid, then oid/oid data
		pduBuf.WriteByte(byte(Sequence))
		pduBuf.Write(length)
		pduBuf.Write(tmpBytes)

	// TODO no tests
	case IPAddress:
		//Oid
		tmpBuf.Write([]byte{byte(ObjectIdentifier), byte(len(oid))})
		tmpBuf.Write(oid)
		//OctetString
		var ipAddressBytes []byte
		switch value := pdu.Value.(type) {
		case []byte:
			ipAddressBytes = value
		case string:
			ip := net.ParseIP(value)
			ipAddressBytes = ipv4toBytes(ip)
		default:
			return nil, fmt.Errorf("Unable to marshal PDU IPAddress; not []byte or String.")
		}
		tmpBuf.Write([]byte{byte(IPAddress), byte(len(ipAddressBytes))})
		tmpBuf.Write(ipAddressBytes)
		// Sequence, length of oid + octetstring, then oid/octetstring data
		pduBuf.WriteByte(byte(Sequence))
		pduBuf.WriteByte(byte(len(oid) + len(ipAddressBytes) + 4))
		pduBuf.Write(tmpBuf.Bytes())

	default:
		return nil, fmt.Errorf("Unable to marshal PDU: unknown BER type %q", pdu.Type)
	}

	return pduBuf.Bytes(), nil
}

// -- Unmarshalling Logic ------------------------------------------------------

func (x *GoSNMP) unmarshalHeader(packet []byte, response *SnmpPacket) (int, error) {
	if len(packet) < 2 {
		return 0, fmt.Errorf("Cannot unmarshal empty packet")
	}
	if response == nil {
		return 0, fmt.Errorf("Cannot unmarshal response into nil packet reference")
	}

	response.Variables = make([]SnmpPDU, 0, 5)

	// Start parsing the packet
	cursor := 0

	// First bytes should be 0x30
	if PDUType(packet[0]) != Sequence {
		return 0, fmt.Errorf("Invalid packet header\n")
	}

	length, cursor := parseLength(packet)
	if len(packet) != length {
		return 0, fmt.Errorf("Error verifying packet sanity: Got %d Expected: %d\n", len(packet), length)
	}
	x.logPrintf("Packet sanity verified, we got all the bytes (%d)", length)

	// Parse SNMP Version
	rawVersion, count, err := parseRawField(packet[cursor:], "version")
	if err != nil {
		return 0, fmt.Errorf("Error parsing SNMP packet version: %s", err.Error())
	}

	cursor += count
	if version, ok := rawVersion.(int); ok {
		response.Version = SnmpVersion(version)
		x.logPrintf("Parsed version %d", version)
	}

	if response.Version == Version3 {
		cursor, err = x.unmarshalV3Header(packet, cursor, response)
		if err != nil {
			return 0, err
		}
	} else {
		// Parse community
		rawCommunity, count, err := parseRawField(packet[cursor:], "community")
		if err != nil {
			return 0, fmt.Errorf("Error parsing community string: %s", err.Error())
		}
		cursor += count
		if community, ok := rawCommunity.(string); ok {
			response.Community = community
			x.logPrintf("Parsed community %s", community)
		}
	}
	return cursor, nil
}

func (x *GoSNMP) unmarshalPayload(packet []byte, cursor int, response *SnmpPacket) error {
	var err error
	// Parse SNMP packet type
	requestType := PDUType(packet[cursor])
	switch requestType {
	// known, supported types
	case GetResponse, GetNextRequest, GetBulkRequest, Report, SNMPv2Trap:
		response.PDUType = requestType
		err = x.unmarshalResponse(packet[cursor:], response)
		if err != nil {
			return fmt.Errorf("Error in unmarshalResponse: %s", err.Error())
		}
	case Trap:
		response.PDUType = requestType
		err = x.unmarshalTrapV1(packet[cursor:], response)
		if err != nil {
			return fmt.Errorf("Error in unmarshalTrapV1: %s", err.Error())
		}
	default:
		return fmt.Errorf("Unknown PDUType %#x", requestType)
	}
	return nil
}

func (x *GoSNMP) unmarshalResponse(packet []byte, response *SnmpPacket) error {
	cursor := 0

	getResponseLength, cursor := parseLength(packet)
	if len(packet) != getResponseLength {
		return fmt.Errorf("Error verifying Response sanity: Got %d Expected: %d\n", len(packet), getResponseLength)
	}
	x.logPrintf("getResponseLength: %d", getResponseLength)

	// Parse Request-ID
	rawRequestID, count, err := parseRawField(packet[cursor:], "request id")
	if err != nil {
		return fmt.Errorf("Error parsing SNMP packet request ID: %s", err.Error())
	}
	cursor += count
	if requestid, ok := rawRequestID.(int); ok {
		response.RequestID = uint32(requestid)
		x.logPrintf("requestID: %d", response.RequestID)
	}

	if response.PDUType == GetBulkRequest {
		// Parse Non Repeaters
		rawNonRepeaters, count, err := parseRawField(packet[cursor:], "non repeaters")
		if err != nil {
			return fmt.Errorf("Error parsing SNMP packet non repeaters: %s", err.Error())
		}
		cursor += count
		if nonRepeaters, ok := rawNonRepeaters.(int); ok {
			response.NonRepeaters = uint8(nonRepeaters)
		}

		// Parse Max Repetitions
		rawMaxRepetitions, count, err := parseRawField(packet[cursor:], "max repetitions")
		if err != nil {
			return fmt.Errorf("Error parsing SNMP packet max repetitions: %s", err.Error())
		}
		cursor += count
		if maxRepetitions, ok := rawMaxRepetitions.(int); ok {
			response.MaxRepetitions = uint8(maxRepetitions)
		}
	} else {
		// Parse Error-Status
		rawError, count, err := parseRawField(packet[cursor:], "error-status")
		if err != nil {
			return fmt.Errorf("Error parsing SNMP packet error: %s", err.Error())
		}
		cursor += count
		if errorStatus, ok := rawError.(int); ok {
			response.Error = SNMPError(errorStatus)
			x.logPrintf("errorStatus: %d", uint8(errorStatus))
		}

		// Parse Error-Index
		rawErrorIndex, count, err := parseRawField(packet[cursor:], "error index")
		if err != nil {
			return fmt.Errorf("Error parsing SNMP packet error index: %s", err.Error())
		}
		cursor += count
		if errorindex, ok := rawErrorIndex.(int); ok {
			response.ErrorIndex = uint8(errorindex)
			x.logPrintf("error-index: %d", uint8(errorindex))
		}
	}

	return x.unmarshalVBL(packet[cursor:], response)
}

func (x *GoSNMP) unmarshalTrapV1(packet []byte, response *SnmpPacket) error {
	cursor := 0

	getResponseLength, cursor := parseLength(packet)
	if len(packet) != getResponseLength {
		return fmt.Errorf("Error verifying Response sanity: Got %d Expected: %d\n", len(packet), getResponseLength)
	}
	x.logPrintf("getResponseLength: %d", getResponseLength)

	// Parse Enterprise
	rawEnterprise, count, err := parseRawField(packet[cursor:], "enterprise")
	if err != nil {
		return fmt.Errorf("Error parsing SNMP packet error: %s", err.Error())
	}
	cursor += count
	if Enterprise, ok := rawEnterprise.([]int); ok {
		response.Enterprise = oidToString(Enterprise)
		x.logPrintf("Enterprise: %+v", Enterprise)
	}

	// Parse AgentAddress
	rawAgentAddress, count, err := parseRawField(packet[cursor:], "agent-address")
	if err != nil {
		return fmt.Errorf("Error parsing SNMP packet error: %s", err.Error())
	}
	cursor += count
	if AgentAddress, ok := rawAgentAddress.(string); ok {
		response.AgentAddress = AgentAddress
		x.logPrintf("AgentAddress: %s", AgentAddress)
	}

	// Parse GenericTrap
	rawGenericTrap, count, err := parseRawField(packet[cursor:], "generic-trap")
	if err != nil {
		return fmt.Errorf("Error parsing SNMP packet error: %s", err.Error())
	}
	cursor += count
	if GenericTrap, ok := rawGenericTrap.(int); ok {
		response.GenericTrap = GenericTrap
		x.logPrintf("GenericTrap: %d", GenericTrap)
	}

	// Parse SpecificTrap
	rawSpecificTrap, count, err := parseRawField(packet[cursor:], "specific-trap")
	if err != nil {
		return fmt.Errorf("Error parsing SNMP packet error: %s", err.Error())
	}
	cursor += count
	if SpecificTrap, ok := rawSpecificTrap.(int); ok {
		response.SpecificTrap = SpecificTrap
		x.logPrintf("SpecificTrap: %d", SpecificTrap)
	}

	// Parse TimeStamp
	rawTimestamp, count, err := parseRawField(packet[cursor:], "time-stamp")
	if err != nil {
		return fmt.Errorf("Error parsing SNMP packet error: %s", err.Error())
	}
	cursor += count
	if Timestamp, ok := rawTimestamp.(uint); ok {
		response.Timestamp = Timestamp
		x.logPrintf("Timestamp: %d", Timestamp)
	}

	return x.unmarshalVBL(packet[cursor:], response)
}

// unmarshal a Varbind list
func (x *GoSNMP) unmarshalVBL(packet []byte, response *SnmpPacket) error {

	var cursor, cursorInc int
	var vblLength int
	if packet[cursor] != 0x30 {
		return fmt.Errorf("Expected a sequence when unmarshalling a VBL, got %x", packet[cursor])
	}

	vblLength, cursor = parseLength(packet)
	if len(packet) != vblLength {
		return fmt.Errorf("Error verifying: packet length %d vbl length %d\n", len(packet), vblLength)
	}
	x.logPrintf("vblLength: %d", vblLength)

	// check for an empty response
	if vblLength == 2 && packet[1] == 0x00 {
		return nil
	}

	// Loop & parse Varbinds
	for cursor < vblLength {
		if packet[cursor] != 0x30 {
			return fmt.Errorf("Expected a sequence when unmarshalling a VB, got %x", packet[cursor])
		}

		_, cursorInc = parseLength(packet[cursor:])
		cursor += cursorInc

		// Parse OID
		rawOid, oidLength, err := parseRawField(packet[cursor:], "OID")
		if err != nil {
			return fmt.Errorf("Error parsing OID Value: %s", err.Error())
		}
		cursor += oidLength

		var oid []int
		var ok bool
		if oid, ok = rawOid.([]int); !ok {
			return fmt.Errorf("unable to type assert rawOid |%v| to []int", rawOid)
		}
		oidStr := oidToString(oid)
		x.logPrintf("OID: %s", oidStr)

		// Parse Value
		v, err := x.decodeValue(packet[cursor:], "value")
		if err != nil {
			return fmt.Errorf("Error decoding value: %v", err)
		}
		valueLength, _ := parseLength(packet[cursor:])
		cursor += valueLength
		response.Variables = append(response.Variables, SnmpPDU{oidStr, v.Type, v.Value, x.Logger})
	}
	return nil
}

// receive response from network and read into a byte array
func (x *GoSNMP) receive() ([]byte, error) {
	n, err := x.Conn.Read(x.rxBuf[:])
	if err != nil {
		return nil, fmt.Errorf("Error reading from UDP: %s", err.Error())
	}

	if n == rxBufSize {
		// This should never happen unless we're using something like a unix domain socket.
		return nil, fmt.Errorf("response buffer too small")
	}

	resp := make([]byte, n)
	copy(resp, x.rxBuf[:n])
	return resp, nil
}
