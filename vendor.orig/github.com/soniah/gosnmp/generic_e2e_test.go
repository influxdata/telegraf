// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// This set of end-to-end integration tests execute gosnmp against a real
// SNMP MIB-2 host. Potential test systems could include a router, NAS box, printer,
// or a linux box running snmpd, snmpsimd.py, etc.
//
// Ensure "gosnmp-test-host" is defined in your hosts file, and points to your
// generic test system.

// +build all end2end

package gosnmp

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func setupConnection(t *testing.T) {
	envTarget := os.Getenv("GOSNMP_TARGET")
	envPort := os.Getenv("GOSNMP_PORT")

	if len(envTarget) <= 0 {
		t.Error("environment variable not set: GOSNMP_TARGET")
	}
	Default.Target = envTarget

	if len(envPort) <= 0 {
		t.Error("environment variable not set: GOSNMP_PORT")
	}
	port, _ := strconv.ParseUint(envPort, 10, 16)
	Default.Port = uint16(port)

	err := Default.Connect()
	if err != nil {
		if len(envTarget) > 0 {
			t.Fatalf("Connection failed. Is snmpd reachable on %s:%s?\n(err: %v)",
				envTarget, envPort, err)
		}
	}
}

func setupConnectionIPv4(t *testing.T) {
	envTarget := os.Getenv("GOSNMP_TARGET_IPV4")
	envPort := os.Getenv("GOSNMP_PORT_IPV4")

	if len(envTarget) <= 0 {
		t.Error("environment variable not set: GOSNMP_TARGET_IPV4")
	}
	Default.Target = envTarget

	if len(envPort) <= 0 {
		t.Error("environment variable not set: GOSNMP_PORT_IPV4")
	}
	port, _ := strconv.ParseUint(envPort, 10, 16)
	Default.Port = uint16(port)

	err := Default.ConnectIPv4()
	if err != nil {
		if len(envTarget) > 0 {
			t.Fatalf("Connection failed. Is snmpd reachable on %s:%s?\n(err: %v)",
				envTarget, envPort, err)
		}
	}
}

/*
TODO work out ipv6 networking, etc

func setupConnectionIPv6(t *testing.T) {
	envTarget := os.Getenv("GOSNMP_TARGET_IPV6")
	envPort := os.Getenv("GOSNMP_PORT_IPV6")

	if len(envTarget) <= 0 {
		t.Error("environment variable not set: GOSNMP_TARGET_IPV6")
	}
	Default.Target = envTarget

	if len(envPort) <= 0 {
		t.Error("environment variable not set: GOSNMP_PORT_IPV6")
	}
	port, _ := strconv.ParseUint(envPort, 10, 16)
	Default.Port = uint16(port)

	err := Default.ConnectIPv6()
	if err != nil {
		if len(envTarget) > 0 {
			t.Fatalf("Connection failed. Is snmpd reachable on %s:%s?\n(err: %v)",
				envTarget, envPort, err)
		}
	}
}
*/

func TestGenericBasicGet(t *testing.T) {
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestGenericBasicGetIPv4Only(t *testing.T) {
	setupConnectionIPv4(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

/*
func TestGenericBasicGetIPv6Only(t *testing.T) {
	setupConnectionIPv6(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}
*/

func TestGenericMultiGet(t *testing.T) {
	setupConnection(t)
	defer Default.Conn.Close()

	oids := []string{
		".1.3.6.1.2.1.1.1.0", // SNMP MIB-2 sysDescr
		".1.3.6.1.2.1.1.5.0", // SNMP MIB-2 sysName
	}
	result, err := Default.Get(oids)
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 2 {
		t.Fatalf("Expected result of size 2")
	}
	for _, v := range result.Variables {
		if v.Type != OctetString {
			t.Fatalf("Expected OctetString")
		}
	}
}

func TestGenericGetNext(t *testing.T) {
	setupConnection(t)
	defer Default.Conn.Close()

	sysDescrOid := ".1.3.6.1.2.1.1.1.0" // SNMP MIB-2 sysDescr
	result, err := Default.GetNext([]string{sysDescrOid})
	if err != nil {
		t.Fatalf("GetNext() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Name == sysDescrOid {
		t.Fatalf("Expected next OID")
	}
}

func TestGenericWalk(t *testing.T) {
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.WalkAll("")
	if err != nil {
		t.Fatalf("WalkAll() Failed with error => %v", err)
	}
	if len(result) <= 1 {
		t.Fatalf("Expected multiple values, got %d", len(result))
	}
}

func TestGenericBulkWalk(t *testing.T) {
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.BulkWalkAll("")
	if err != nil {
		t.Fatalf("BulkWalkAll() Failed with error => %v", err)
	}
	if len(result) <= 1 {
		t.Fatalf("Expected multiple values, got %d", len(result))
	}
}

// Standard exception/error tests

func TestMaxOids(t *testing.T) {
	setupConnection(t)
	defer Default.Conn.Close()

	Default.MaxOids = 1

	var err error
	oids := []string{".1.3.6.1.2.1.1.7.0",
		".1.3.6.1.2.1.2.2.1.10.1"} // 2 arbitrary Oids
	errString := "oid count (2) is greater than MaxOids (1)"

	_, err = Default.Get(oids)
	if err == nil {
		t.Fatalf("Expected too many oids failure. Got nil")
	} else if err.Error() != errString {
		t.Fatalf("Expected too many oids failure. Got => %v", err)
	}

	_, err = Default.GetNext(oids)
	if err == nil {
		t.Fatalf("Expected too many oids failure. Got nil")
	} else if err.Error() != errString {
		t.Fatalf("Expected too many oids failure. Got => %v", err)
	}

	_, err = Default.GetBulk(oids, 0, 0)
	if err == nil {
		t.Fatalf("Expected too many oids failure. Got nil")
	} else if err.Error() != errString {
		t.Fatalf("Expected too many oids failure. Got => %v", err)
	}
}

func TestGenericFailureUnknownHost(t *testing.T) {
	unknownHost := fmt.Sprintf("gosnmp-test-unknown-host-%d", time.Now().UTC().UnixNano())
	Default.Target = unknownHost
	err := Default.Connect()
	if err == nil {
		t.Fatalf("Expected connection failure due to unknown host")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "no such host") {
		t.Fatalf("Expected connection error of type 'no such host'! Got => %v", err)
	}
	_, err = Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err == nil {
		t.Fatalf("Expected get to fail due to missing connection")
	}
}

func TestGenericFailureConnectionTimeout(t *testing.T) {
	Default.Target = "198.51.100.1" // Black hole
	err := Default.Connect()
	if err != nil {
		t.Fatalf("Did not expect connection error with IP address")
	}
	_, err = Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err == nil {
		t.Fatalf("Expected Get() to fail due to invalid IP")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("Expected timeout error. Got => %v", err)
	}
}

func TestGenericFailureConnectionRefused(t *testing.T) {
	Default.Target = "127.0.0.1"
	Default.Port = 1 // Don't expect SNMP to be running here!
	err := Default.Connect()
	if err != nil {
		t.Fatalf("Did not expect connection error with IP address")
	}
	_, err = Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err == nil {
		t.Fatalf("Expected Get() to fail due to invalid port")
	}
	if !(strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "forcibly closed")) {
		t.Fatalf("Expected connection refused error. Got => %v", err)
	}
}

func TestSnmpV3NoAuthNoPrivBasicGet(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = NoAuthNoPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "noAuthNoPrivUser"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3AuthNoPrivMD5Get(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthNoPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authMD5OnlyUser", AuthenticationProtocol: MD5, AuthenticationPassphrase: "testingpass0123456789"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3AuthNoPrivSHAGet(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthNoPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authSHAOnlyUser", AuthenticationProtocol: SHA, AuthenticationPassphrase: "testingpass9876543210"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3AuthMD5PrivDESGet(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authMD5PrivDESUser",
		AuthenticationProtocol:   MD5,
		AuthenticationPassphrase: "testingpass9876543210",
		PrivacyProtocol:          DES,
		PrivacyPassphrase:        "testingpass9876543210"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3AuthSHAPrivDESGet(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authSHAPrivDESUser",
		AuthenticationProtocol:   SHA,
		AuthenticationPassphrase: "testingpassabc6543210",
		PrivacyProtocol:          DES,
		PrivacyPassphrase:        "testingpassabc6543210"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3AuthMD5PrivAESGet(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authMD5PrivAESUser",
		AuthenticationProtocol:   MD5,
		AuthenticationPassphrase: "AEStestingpass9876543210",
		PrivacyProtocol:          AES,
		PrivacyPassphrase:        "AEStestingpass9876543210"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3AuthSHAPrivAESGet(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authSHAPrivAESUser",
		AuthenticationProtocol:   SHA,
		AuthenticationPassphrase: "AEStestingpassabc6543210",
		PrivacyProtocol:          AES,
		PrivacyPassphrase:        "AEStestingpassabc6543210"}
	setupConnection(t)
	defer Default.Conn.Close()

	result, err := Default.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMP MIB-2 sysDescr
	if err != nil {
		t.Fatalf("Get() failed with error => %v", err)
	}
	if len(result.Variables) != 1 {
		t.Fatalf("Expected result of size 1")
	}
	if result.Variables[0].Type != OctetString {
		t.Fatalf("Expected sysDescr to be OctetString")
	}
	sysDescr := result.Variables[0].Value.([]byte)
	if len(sysDescr) == 0 {
		t.Fatalf("Got a zero length sysDescr")
	}
}

func TestSnmpV3PrivEmptyPrivatePassword(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authSHAPrivAESUser",
		AuthenticationProtocol:   SHA,
		AuthenticationPassphrase: "AEStestingpassabc6543210",
		PrivacyProtocol:          AES,
		PrivacyPassphrase:        ""}

	err := Default.Connect()
	if err == nil {
		t.Fatalf("Expected validation error for empty PrivacyPassphrase")
	}
}

func TestSnmpV3AuthNoPrivEmptyPrivatePassword(t *testing.T) {
	Default.Version = Version3
	Default.MsgFlags = AuthNoPriv
	Default.SecurityModel = UserSecurityModel
	Default.SecurityParameters = &UsmSecurityParameters{UserName: "authSHAOnlyUser",
		AuthenticationProtocol:   SHA,
		AuthenticationPassphrase: "testingpass9876543210",
		PrivacyProtocol:          AES,
		PrivacyPassphrase:        ""}

	err := Default.Connect()
	if err == nil {
		t.Fatalf("Expected validation error for empty PrivacyPassphrase")
	}

}
