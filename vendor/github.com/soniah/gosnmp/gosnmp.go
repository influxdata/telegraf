// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosnmp

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	// MaxOids is the maximum number of OIDs permitted in a single call,
	// otherwise error. MaxOids too high can cause remote devices to fail
	// strangely. 60 seems to be a common value that works, but you will want
	// to change this in the GoSNMP struct
	MaxOids = 60

	// Base OID for MIB-2 defined SNMP variables
	baseOid = ".1.3.6.1.2.1"

	// Java SNMP uses 50, snmp-net uses 10
	defaultMaxRepetitions = 50
)

// GoSNMP represents GoSNMP library state
type GoSNMP struct {
	// Conn is net connection to use, typically established using GoSNMP.Connect()
	Conn net.Conn

	// Target is an ipv4 address
	Target string

	// Port is a port
	Port uint16

	// Transport is the transport protocol to use ("udp" or "tcp"); if unset "udp" will be used.
	Transport string

	// Community is an SNMP Community string
	Community string

	// Version is an SNMP Version
	Version SnmpVersion

	// Timeout is the timeout for the SNMP Query
	Timeout time.Duration

	// Set the number of retries to attempt within timeout.
	Retries int

	// Double timeout in each retry
	ExponentialTimeout bool

	// Logger is the GoSNMP.Logger to use for debugging. If nil, debugging
	// output will be discarded (/dev/null). For verbose logging to stdout:
	// x.Logger = log.New(os.Stdout, "", 0)
	Logger Logger

	// loggingEnabled is set if the Logger is nil, short circuits any 'Logger' calls
	loggingEnabled bool

	// MaxOids is the maximum number of oids allowed in a Get()
	// (default: MaxOids)
	MaxOids int

	// MaxRepetitions sets the GETBULK max-repetitions used by BulkWalk*
	// Unless MaxRepetitions is specified it will use defaultMaxRepetitions (50)
	// This may cause issues with some devices, if so set MaxRepetitions lower.
	// See comments in https://github.com/soniah/gosnmp/issues/100
	MaxRepetitions uint8

	// NonRepeaters sets the GETBULK max-repeaters used by BulkWalk*
	// (default: 0 as per RFC 1905)
	NonRepeaters int

	// netsnmp has '-C APPOPTS - set various application specific behaviours'
	//
	// - 'c: do not check returned OIDs are increasing' - use AppOpts = map[string]interface{"c":true} with
	//   Walk() or BulkWalk(). The library user needs to implement their own policy for terminating walks.
	// - 'p,i,I,t,E' -> pull requests welcome
	AppOpts map[string]interface{}

	// Internal - used to sync requests to responses
	requestID uint32
	random    *rand.Rand

	rxBuf *[rxBufSize]byte // has to be pointer due to https://github.com/golang/go/issues/11728

	// MsgFlags is an SNMPV3 MsgFlags
	MsgFlags SnmpV3MsgFlags

	// SecurityModel is an SNMPV3 Security Model
	SecurityModel SnmpV3SecurityModel

	// SecurityParameters is an SNMPV3 Security Model parameters struct
	SecurityParameters SnmpV3SecurityParameters

	// ContextEngineID is SNMPV3 ContextEngineID in ScopedPDU
	ContextEngineID string

	// ContextName is SNMPV3 ContextName in ScopedPDU
	ContextName string

	// Internal - used to sync requests to responses - snmpv3
	msgID uint32
}

// Default connection settings
var Default = &GoSNMP{
	Port:               161,
	Transport:          "udp",
	Community:          "public",
	Version:            Version2c,
	Timeout:            time.Duration(2) * time.Second,
	Retries:            3,
	ExponentialTimeout: true,
	MaxOids:            MaxOids,
}

// SnmpPDU will be used when doing SNMP Set's
type SnmpPDU struct {
	// Name is an oid in string format eg ".1.3.6.1.4.9.27"
	Name string

	// The type of the value eg Integer
	Type Asn1BER

	// The value to be set by the SNMP set, or the value when
	// sending a trap
	Value interface{}

	// Logger implements the Logger interface
	Logger Logger
}

// AsnExtensionID mask to identify types > 30 in subsequent byte
const AsnExtensionID = 0x1F

//go:generate stringer -type Asn1BER

// Asn1BER is the type of the SNMP PDU
type Asn1BER byte

// Asn1BER's - http://www.ietf.org/rfc/rfc1442.txt
const (
	EndOfContents     Asn1BER = 0x00
	UnknownType       Asn1BER = 0x00
	Boolean           Asn1BER = 0x01
	Integer           Asn1BER = 0x02
	BitString         Asn1BER = 0x03
	OctetString       Asn1BER = 0x04
	Null              Asn1BER = 0x05
	ObjectIdentifier  Asn1BER = 0x06
	ObjectDescription Asn1BER = 0x07
	IPAddress         Asn1BER = 0x40
	Counter32         Asn1BER = 0x41
	Gauge32           Asn1BER = 0x42
	TimeTicks         Asn1BER = 0x43
	Opaque            Asn1BER = 0x44
	NsapAddress       Asn1BER = 0x45
	Counter64         Asn1BER = 0x46
	Uinteger32        Asn1BER = 0x47
	OpaqueFloat       Asn1BER = 0x78
	OpaqueDouble      Asn1BER = 0x79
	NoSuchObject      Asn1BER = 0x80
	NoSuchInstance    Asn1BER = 0x81
	EndOfMibView      Asn1BER = 0x82
)

//go:generate stringer -type SNMPError

// SNMPError is the type for standard SNMP errors.
type SNMPError uint8

// SNMP Errors
const (
	NoError             SNMPError = iota // No error occurred. This code is also used in all request PDUs, since they have no error status to report.
	TooBig                               // The size of the Response-PDU would be too large to transport.
	NoSuchName                           // The name of a requested object was not found.
	BadValue                             // A value in the request didn't match the structure that the recipient of the request had for the object. For example, an object in the request was specified with an incorrect length or type.
	ReadOnly                             // An attempt was made to set a variable that has an Access value indicating that it is read-only.
	GenErr                               // An error occurred other than one indicated by a more specific error code in this table.
	NoAccess                             // Access was denied to the object for security reasons.
	WrongType                            // The object type in a variable binding is incorrect for the object.
	WrongLength                          // A variable binding specifies a length incorrect for the object.
	WrongEncoding                        // A variable binding specifies an encoding incorrect for the object.
	WrongValue                           // The value given in a variable binding is not possible for the object.
	NoCreation                           // A specified variable does not exist and cannot be created.
	InconsistentValue                    // A variable binding specifies a value that could be held by the variable but cannot be assigned to it at this time.
	ResourceUnavailable                  // An attempt to set a variable required a resource that is not available.
	CommitFailed                         // An attempt to set a particular variable failed.
	UndoFailed                           // An attempt to set a particular variable as part of a group of variables failed, and the attempt to then undo the setting of other variables was not successful.
	AuthorizationError                   // A problem occurred in authorization.
	NotWritable                          // The variable cannot be written or created.
	InconsistentName                     // The name in a variable binding specifies a variable that does not exist.
)

//
// Public Functions (main interface)
//

// Connect creates and opens a socket. Because UDP is a connectionless
// protocol, you won't know if the remote host is responding until you send
// packets. Neither will you know if the host is regularly disappearing and reappearing.
//
// For historical reasons (ie this is part of the public API), the method won't
// be renamed to Dial().
func (x *GoSNMP) Connect() error {
	return x.connect("")
}

// ConnectIPv4 forces an IPv4-only connection
func (x *GoSNMP) ConnectIPv4() error {
	return x.connect("4")
}

// ConnectIPv6 forces an IPv6-only connection
func (x *GoSNMP) ConnectIPv6() error {
	return x.connect("6")
}

// connect to address addr on the given network
//
// https://golang.org/pkg/net/#Dial gives acceptable network values as:
//   "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp", "udp4" (IPv4-only),"udp6" (IPv6-only), "ip",
//   "ip4" (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram" and "unixpacket"
func (x *GoSNMP) connect(networkSuffix string) error {
	err := x.validateParameters()
	if err != nil {
		return err
	}

	x.Transport = x.Transport + networkSuffix
	err = x.netConnect()
	if err != nil {
		return fmt.Errorf("error establishing connection to host: %s", err.Error())
	}

	if x.random == nil {
		x.random = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	}
	// http://tools.ietf.org/html/rfc3412#section-6 - msgID only
	// uses the first 31 bits
	// msgID INTEGER (0..2147483647)
	x.msgID = uint32(x.random.Int31())
	// RequestID is Integer32 from SNMPV2-SMI and uses all 32 bits
	x.requestID = x.random.Uint32()

	x.rxBuf = new([rxBufSize]byte)

	return nil
}

// Performs the real socket opening network operation. This can be used to do a
// reconnect (needed for TCP)
func (x *GoSNMP) netConnect() error {
	var err error
	addr := net.JoinHostPort(x.Target, strconv.Itoa(int(x.Port)))
	x.Conn, err = net.DialTimeout(x.Transport, addr, x.Timeout)
	return err
}

func (x *GoSNMP) validateParameters() error {
	if x.Logger == nil {
		x.Logger = log.New(ioutil.Discard, "", 0)
	} else {
		x.loggingEnabled = true
	}

	if x.Transport == "" {
		x.Transport = "udp"
	}

	if x.MaxOids == 0 {
		x.MaxOids = MaxOids
	} else if x.MaxOids < 0 {
		return fmt.Errorf("MaxOids cannot be less than 0")
	}

	if x.Version == Version3 {
		x.MsgFlags |= Reportable // tell the snmp server that a report PDU MUST be sent

		err := x.validateParametersV3()
		if err != nil {
			return err
		}
		err = x.SecurityParameters.init(x.Logger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (x *GoSNMP) mkSnmpPacket(pdutype PDUType, pdus []SnmpPDU, nonRepeaters uint8, maxRepetitions uint8) *SnmpPacket {
	var newSecParams SnmpV3SecurityParameters
	if x.SecurityParameters != nil {
		newSecParams = x.SecurityParameters.Copy()
	}
	return &SnmpPacket{
		Version:            x.Version,
		Community:          x.Community,
		MsgFlags:           x.MsgFlags,
		SecurityModel:      x.SecurityModel,
		SecurityParameters: newSecParams,
		ContextEngineID:    x.ContextEngineID,
		ContextName:        x.ContextName,
		Error:              0,
		ErrorIndex:         0,
		PDUType:            pdutype,
		NonRepeaters:       nonRepeaters,
		MaxRepetitions:     maxRepetitions,
		Variables:          pdus,
	}
}

// Get sends an SNMP GET request
func (x *GoSNMP) Get(oids []string) (result *SnmpPacket, err error) {
	oidCount := len(oids)
	if oidCount > x.MaxOids {
		return nil, fmt.Errorf("oid count (%d) is greater than MaxOids (%d)",
			oidCount, x.MaxOids)
	}
	// convert oids slice to pdu slice
	var pdus []SnmpPDU
	for _, oid := range oids {
		pdus = append(pdus, SnmpPDU{oid, Null, nil, x.Logger})
	}
	// build up SnmpPacket
	packetOut := x.mkSnmpPacket(GetRequest, pdus, 0, 0)
	return x.send(packetOut, true)
}

// Set sends an SNMP SET request
func (x *GoSNMP) Set(pdus []SnmpPDU) (result *SnmpPacket, err error) {
	var packetOut *SnmpPacket
	switch pdus[0].Type {
	// TODO test Gauge32
	case Integer, OctetString, Gauge32, IPAddress:
		packetOut = x.mkSnmpPacket(SetRequest, pdus, 0, 0)
	default:
		return nil, fmt.Errorf("ERR:gosnmp currently only supports SNMP SETs for Integers, IPAddress and OctetStrings")
	}
	return x.send(packetOut, true)
}

// GetNext sends an SNMP GETNEXT request
func (x *GoSNMP) GetNext(oids []string) (result *SnmpPacket, err error) {
	oidCount := len(oids)
	if oidCount > x.MaxOids {
		return nil, fmt.Errorf("oid count (%d) is greater than MaxOids (%d)",
			oidCount, x.MaxOids)
	}

	// convert oids slice to pdu slice
	var pdus []SnmpPDU
	for _, oid := range oids {
		pdus = append(pdus, SnmpPDU{oid, Null, nil, x.Logger})
	}

	// Marshal and send the packet
	packetOut := x.mkSnmpPacket(GetNextRequest, pdus, 0, 0)

	return x.send(packetOut, true)
}

// GetBulk sends an SNMP GETBULK request
//
// For maxRepetitions greater than 255, use BulkWalk() or BulkWalkAll()
func (x *GoSNMP) GetBulk(oids []string, nonRepeaters uint8, maxRepetitions uint8) (result *SnmpPacket, err error) {
	oidCount := len(oids)
	if oidCount > x.MaxOids {
		return nil, fmt.Errorf("oid count (%d) is greater than MaxOids (%d)",
			oidCount, x.MaxOids)
	}

	// convert oids slice to pdu slice
	var pdus []SnmpPDU
	for _, oid := range oids {
		pdus = append(pdus, SnmpPDU{oid, Null, nil, x.Logger})
	}

	// Marshal and send the packet
	packetOut := x.mkSnmpPacket(GetBulkRequest, pdus, nonRepeaters, maxRepetitions)
	return x.send(packetOut, true)
}

// SnmpEncodePacket exposes SNMP packet generation to external callers.
// This is useful for generating traffic for use over separate transport
// stacks and creating traffic samples for test purposes.
func (x *GoSNMP) SnmpEncodePacket(pdutype PDUType, pdus []SnmpPDU, nonRepeaters uint8, maxRepetitions uint8) ([]byte, error) {
	err := x.validateParameters()
	if err != nil {
		return []byte{}, err
	}

	pkt := x.mkSnmpPacket(pdutype, pdus, nonRepeaters, maxRepetitions)

	// Request ID is an atomic counter (started at a random value)
	reqID := atomic.AddUint32(&(x.requestID), 1) // TODO: fix overflows
	pkt.RequestID = reqID

	if x.Version == Version3 {
		msgID := atomic.AddUint32(&(x.msgID), 1) // TODO: fix overflows
		pkt.MsgID = msgID

		err = x.initPacket(pkt)
		if err != nil {
			return []byte{}, err
		}
	}

	var out []byte
	out, err = pkt.marshalMsg()
	if err != nil {
		return []byte{}, err
	}

	return out, nil
}

// SnmpDecodePacket exposes SNMP packet parsing to external callers.
// This is useful for processing traffic from other sources and
// building test harnesses.
func (x *GoSNMP) SnmpDecodePacket(resp []byte) (*SnmpPacket, error) {
	var err error

	result := new(SnmpPacket)

	err = x.validateParameters()
	if err != nil {
		return result, err
	}

	result.Logger = x.Logger
	if x.SecurityParameters != nil {
		result.SecurityParameters = x.SecurityParameters.Copy()
	}

	var cursor int
	cursor, err = x.unmarshalHeader(resp, result)
	if err != nil {
		err = fmt.Errorf("Unable to decode packet header: %s", err.Error())
		return result, err
	}

	if result.Version == Version3 {
		resp, cursor, err = x.decryptPacket(resp, cursor, result)
		if err != nil {
			return result, err
		}
	}

	err = x.unmarshalPayload(resp, cursor, result)
	if err != nil {
		err = fmt.Errorf("Unable to decode packet body: %s", err.Error())
		return result, err
	}

	if result == nil || len(result.Variables) < 1 {
		err = fmt.Errorf("Unable to decode packet: no variables")
		return result, err
	}
	return result, nil
}

// SetRequestID sets the base ID value for future requests
func (x *GoSNMP) SetRequestID(reqID uint32) {
	x.requestID = reqID
}

// SetMsgID sets the base ID value for future messages
func (x *GoSNMP) SetMsgID(msgID uint32) {
	x.msgID = msgID & 0x7fffffff
}

//
// SNMP Walk functions - Analogous to net-snmp's snmpwalk commands
//

// WalkFunc is the type of the function called for each data unit visited
// by the Walk function.  If an error is returned processing stops.
type WalkFunc func(dataUnit SnmpPDU) error

// BulkWalk retrieves a subtree of values using GETBULK. As the tree is
// walked walkFn is called for each new value. The function immediately returns
// an error if either there is an underlaying SNMP error (e.g. GetBulk fails),
// or if walkFn returns an error.
func (x *GoSNMP) BulkWalk(rootOid string, walkFn WalkFunc) error {
	return x.walk(GetBulkRequest, rootOid, walkFn)
}

// BulkWalkAll is similar to BulkWalk but returns a filled array of all values
// rather than using a callback function to stream results. Caution: if you
// have set x.AppOpts to 'c', BulkWalkAll may loop indefinitely and cause an
// Out Of Memory - use BulkWalk instead.
func (x *GoSNMP) BulkWalkAll(rootOid string) (results []SnmpPDU, err error) {
	return x.walkAll(GetBulkRequest, rootOid)
}

// Walk retrieves a subtree of values using GETNEXT - a request is made for each
// value, unlike BulkWalk which does this operation in batches. As the tree is
// walked walkFn is called for each new value. The function immediately returns
// an error if either there is an underlaying SNMP error (e.g. GetNext fails),
// or if walkFn returns an error.
func (x *GoSNMP) Walk(rootOid string, walkFn WalkFunc) error {
	return x.walk(GetNextRequest, rootOid, walkFn)
}

// WalkAll is similar to Walk but returns a filled array of all values rather
// than using a callback function to stream results. Caution: if you have set
// x.AppOpts to 'c', WalkAll may loop indefinitely and cause an Out Of Memory -
// use Walk instead.
func (x *GoSNMP) WalkAll(rootOid string) (results []SnmpPDU, err error) {
	return x.walkAll(GetNextRequest, rootOid)
}

//
// Public Functions (helpers) - in alphabetical order
//

// Partition - returns true when dividing a slice into
// partitionSize lengths, including last partition which may be smaller
// than partitionSize. This is useful when you have a large array of OIDs
// to run Get() on. See the tests for example usage.
//
// For example for a slice of 8 items to be broken into partitions of
// length 3, Partition returns true for the currentPosition having
// the following values:
//
// 0  1  2  3  4  5  6  7
//       T        T     T
//
func Partition(currentPosition, partitionSize, sliceLength int) bool {
	if currentPosition < 0 || currentPosition >= sliceLength {
		return false
	}
	if partitionSize == 1 { // redundant, but an obvious optimisation
		return true
	}
	if currentPosition%partitionSize == partitionSize-1 {
		return true
	}
	if currentPosition == sliceLength-1 {
		return true
	}
	return false
}

// ToBigInt converts SnmpPDU.Value to big.Int, or returns a zero big.Int for
// non int-like types (eg strings).
//
// This is a convenience function to make working with SnmpPDU's easier - it
// reduces the need for type assertions. A big.Int is convenient, as SNMP can
// return int32, uint32, and uint64.
func ToBigInt(value interface{}) *big.Int {
	var val int64
	switch value := value.(type) { // shadow
	case int:
		val = int64(value)
	case int8:
		val = int64(value)
	case int16:
		val = int64(value)
	case int32:
		val = int64(value)
	case int64:
		val = int64(value)
	case uint:
		val = int64(value)
	case uint8:
		val = int64(value)
	case uint16:
		val = int64(value)
	case uint32:
		val = int64(value)
	case uint64:
		return (uint64ToBigInt(value))
	case string:
		// for testing and other apps - numbers may appear as strings
		var err error
		if val, err = strconv.ParseInt(value, 10, 64); err != nil {
			return new(big.Int)
		}
	default:
		return new(big.Int)
	}
	return big.NewInt(val)
}
