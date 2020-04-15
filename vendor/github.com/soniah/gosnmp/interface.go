// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gosnmp

import (
	"time"
)

//go:generate mockgen --destination gosnmp_mock.go --package=gosnmp --source interface.go

// Handler is a GoSNMP interface
//
// Handler is provided to assist with testing using mocks
type Handler interface {
	// Connect creates and opens a socket. Because UDP is a connectionless
	// protocol, you won't know if the remote host is responding until you send
	// packets. And if the host is regularly disappearing and reappearing, you won't
	// know if you've only done a Connect().
	//
	// For historical reasons (ie this is part of the public API), the method won't
	// be renamed.
	Connect() error

	// ConnectIPv4 connects using IPv4
	ConnectIPv4() error

	// ConnectIPv6 connects using IPv6
	ConnectIPv6() error

	// Get sends an SNMP GET request
	Get(oids []string) (result *SnmpPacket, err error)

	// GetBulk sends an SNMP GETBULK request
	//
	// For maxRepetitions greater than 255, use BulkWalk() or BulkWalkAll()
	GetBulk(oids []string, nonRepeaters uint8, maxRepetitions uint8) (result *SnmpPacket, err error)

	// GetNext sends an SNMP GETNEXT request
	GetNext(oids []string) (result *SnmpPacket, err error)

	// Walk retrieves a subtree of values using GETNEXT - a request is made for each
	// value, unlike BulkWalk which does this operation in batches. As the tree is
	// walked walkFn is called for each new value. The function immediately returns
	// an error if either there is an underlaying SNMP error (e.g. GetNext fails),
	// or if walkFn returns an error.
	Walk(rootOid string, walkFn WalkFunc) error

	// WalkAll is similar to Walk but returns a filled array of all values rather
	// than using a callback function to stream results.
	WalkAll(rootOid string) (results []SnmpPDU, err error)

	// BulkWalk retrieves a subtree of values using GETBULK. As the tree is
	// walked walkFn is called for each new value. The function immediately returns
	// an error if either there is an underlaying SNMP error (e.g. GetBulk fails),
	// or if walkFn returns an error.
	BulkWalk(rootOid string, walkFn WalkFunc) error

	// BulkWalkAll is similar to BulkWalk but returns a filled array of all values
	// rather than using a callback function to stream results.
	BulkWalkAll(rootOid string) (results []SnmpPDU, err error)

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
	SendTrap(trap SnmpTrap) (result *SnmpPacket, err error)

	// UnmarshalTrap unpacks the SNMP Trap.
	UnmarshalTrap(trap []byte) (result *SnmpPacket)

	// Set sends an SNMP SET request
	Set(pdus []SnmpPDU) (result *SnmpPacket, err error)

	// Check makes checking errors easy, so they actually get a minimal check
	Check(err error)

	// Close closes the connection
	Close() error

	// Target gets the Target
	Target() string

	// SetTarget sets the Target
	SetTarget(target string)

	// Port gets the Port
	Port() uint16

	// SetPort sets the Port
	SetPort(port uint16)

	// Community gets the Community
	Community() string

	// SetCommunity sets the Community
	SetCommunity(community string)

	// Version gets the Version
	Version() SnmpVersion

	// SetVersion sets the Version
	SetVersion(version SnmpVersion)

	// Timeout gets the Timeout
	Timeout() time.Duration

	// SetTimeout sets the Timeout
	SetTimeout(timeout time.Duration)

	// Retries gets the Retries
	Retries() int

	// SetRetries sets the Retries
	SetRetries(retries int)

	// GetExponentialTimeout gets the ExponentialTimeout
	GetExponentialTimeout() bool

	// SetExponentialTimeout sets the ExponentialTimeout
	SetExponentialTimeout(value bool)

	// Logger gets the Logger
	Logger() Logger

	// SetLogger sets the Logger
	SetLogger(logger Logger)

	// MaxOids gets the MaxOids
	MaxOids() int

	// SetMaxOids sets the MaxOids
	SetMaxOids(maxOids int)

	// MaxRepetitions gets the maxRepetitions
	MaxRepetitions() uint8

	// SetMaxRepetitions sets the maxRepetitions
	SetMaxRepetitions(maxRepetitions uint8)

	// NonRepeaters gets the nonRepeaters
	NonRepeaters() int

	// SetNonRepeaters sets the nonRepeaters
	SetNonRepeaters(nonRepeaters int)

	// MsgFlags gets the MsgFlags
	MsgFlags() SnmpV3MsgFlags

	// SetMsgFlags sets the MsgFlags
	SetMsgFlags(msgFlags SnmpV3MsgFlags)

	// SecurityModel gets the SecurityModel
	SecurityModel() SnmpV3SecurityModel

	// SetSecurityModel sets the SecurityModel
	SetSecurityModel(securityModel SnmpV3SecurityModel)

	// SecurityParameters gets the SecurityParameters
	SecurityParameters() SnmpV3SecurityParameters

	// SetSecurityParameters sets the SecurityParameters
	SetSecurityParameters(securityParameters SnmpV3SecurityParameters)

	// ContextEngineID gets the ContextEngineID
	ContextEngineID() string

	// SetContextEngineID sets the ContextEngineID
	SetContextEngineID(contextEngineID string)

	// ContextName gets the ContextName
	ContextName() string

	// SetContextName sets the ContextName
	SetContextName(contextName string)
}

// snmpHandler is a wrapper around gosnmp
type snmpHandler struct {
	GoSNMP
}

// NewHandler creates a new Handler using gosnmp
func NewHandler() Handler {
	return &snmpHandler{
		GoSNMP{
			Port:      Default.Port,
			Community: Default.Community,
			Version:   Default.Version,
			Timeout:   Default.Timeout,
			Retries:   Default.Retries,
			MaxOids:   Default.MaxOids,
		},
	}
}

func (x *snmpHandler) Target() string {
	// not x.Target because it would reference function Target
	return x.GoSNMP.Target
}

func (x *snmpHandler) SetTarget(target string) {
	x.GoSNMP.Target = target
}

func (x *snmpHandler) Port() uint16 {
	return x.GoSNMP.Port
}

func (x *snmpHandler) SetPort(port uint16) {
	x.GoSNMP.Port = port
}

func (x *snmpHandler) Community() string {
	return x.GoSNMP.Community
}

func (x *snmpHandler) SetCommunity(community string) {
	x.GoSNMP.Community = community
}

func (x *snmpHandler) Version() SnmpVersion {
	return x.GoSNMP.Version
}

func (x *snmpHandler) SetVersion(version SnmpVersion) {
	x.GoSNMP.Version = version
}

func (x *snmpHandler) Timeout() time.Duration {
	return x.GoSNMP.Timeout
}

func (x *snmpHandler) SetTimeout(timeout time.Duration) {
	x.GoSNMP.Timeout = timeout
}

func (x *snmpHandler) Retries() int {
	return x.GoSNMP.Retries
}

func (x *snmpHandler) SetRetries(retries int) {
	x.GoSNMP.Retries = retries
}

func (x *snmpHandler) GetExponentialTimeout() bool {
	return x.GoSNMP.ExponentialTimeout
}

func (x *snmpHandler) SetExponentialTimeout(value bool) {
	x.GoSNMP.ExponentialTimeout = value
}

func (x *snmpHandler) Logger() Logger {
	return x.GoSNMP.Logger
}

func (x *snmpHandler) SetLogger(logger Logger) {
	x.GoSNMP.Logger = logger
}

func (x *snmpHandler) MaxOids() int {
	return x.GoSNMP.MaxOids
}

func (x *snmpHandler) SetMaxOids(maxOids int) {
	x.GoSNMP.MaxOids = maxOids
}

func (x *snmpHandler) MaxRepetitions() uint8 {
	return x.GoSNMP.MaxRepetitions
}

func (x *snmpHandler) SetMaxRepetitions(maxRepetitions uint8) {
	x.GoSNMP.MaxRepetitions = maxRepetitions
}

func (x *snmpHandler) NonRepeaters() int {
	return x.GoSNMP.NonRepeaters
}

func (x *snmpHandler) SetNonRepeaters(nonRepeaters int) {
	x.GoSNMP.NonRepeaters = nonRepeaters
}

func (x *snmpHandler) MsgFlags() SnmpV3MsgFlags {
	return x.GoSNMP.MsgFlags
}

func (x *snmpHandler) SetMsgFlags(msgFlags SnmpV3MsgFlags) {
	x.GoSNMP.MsgFlags = msgFlags
}

func (x *snmpHandler) SecurityModel() SnmpV3SecurityModel {
	return x.GoSNMP.SecurityModel
}

func (x *snmpHandler) SetSecurityModel(securityModel SnmpV3SecurityModel) {
	x.GoSNMP.SecurityModel = securityModel
}

func (x *snmpHandler) SecurityParameters() SnmpV3SecurityParameters {
	return x.GoSNMP.SecurityParameters
}

func (x *snmpHandler) SetSecurityParameters(securityParameters SnmpV3SecurityParameters) {
	x.GoSNMP.SecurityParameters = securityParameters
}

func (x *snmpHandler) ContextEngineID() string {
	return x.GoSNMP.ContextEngineID
}

func (x *snmpHandler) SetContextEngineID(contextEngineID string) {
	x.GoSNMP.ContextEngineID = contextEngineID
}

func (x *snmpHandler) ContextName() string {
	return x.GoSNMP.ContextName
}

func (x *snmpHandler) SetContextName(contextName string) {
	x.GoSNMP.ContextName = contextName
}

func (x *snmpHandler) Close() error {
	// not x.Conn for consistency
	return x.GoSNMP.Conn.Close()
}
