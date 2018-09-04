// Copyright 2012-2018 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

// +build all marshal

package gosnmp

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"
)

// Tests in alphabetical order of function being tested

// -- Enmarshal ----------------------------------------------------------------

// "Enmarshal" not "Marshal" - easier to select tests via a regex

type testsEnmarshalVarbindPosition struct {
	oid string

	/*
		start and finish position of bytes are calculated with application layer
		starting at byte 0. There are two ways to understand Wireshark dumps,
		switch between them:

		1) the standard decode of the full packet - easier to understand
		what's actually happening

		2) for counting byte positions: select "Simple Network Management
		Protocal" line in Wiresharks middle pane, then right click and choose
		"Export Packet Bytes..." (as .raw). Open the capture in wireshark, it
		will decode as "BER Encoded File". Click on each varbind and the
		"packet bytes" window will highlight the corresponding bytes, then the
		start and end positions can be found.
	*/

	/*
		go-bindata has changed output format. Old style is needed:

		go get -u github.com/jteeuwen/go-bindata/...
		git co 79847ab
		rm ~/go/bin/go-bindata  # belts and braces
		go install
		~/go/bin/go-bindata -uncompressed *.pcap
	*/

	start    int
	finish   int
	pduType  Asn1BER
	pduValue interface{}
}

type testsEnmarshalT struct {
	version     SnmpVersion
	community   string
	requestType PDUType
	requestid   uint32
	msgid       uint32
	// function and function name returning bytes from tcpdump
	goodBytes func() []byte
	funcName  string // could do this via reflection
	// start position of the pdu
	pduStart int
	// start position of the vbl
	vblStart int
	// finish position of pdu, vbl and message - all the same
	finish int
	// a slice of positions containing start and finish of each varbind
	vbPositions []testsEnmarshalVarbindPosition
}

var testsEnmarshal = []testsEnmarshalT{
	{
		Version2c,
		"public",
		GetRequest,
		1871507044,
		0,
		kyoceraRequestBytes,
		"kyocera_request",
		0x0e, // pdu start
		0x1d, // vbl start
		0xa0, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.2.1.1.7.0", 0x20, 0x2d, Null, nil},
			{".1.3.6.1.2.1.2.2.1.10.1", 0x2e, 0x3d, Null, nil},
			{".1.3.6.1.2.1.2.2.1.5.1", 0x3e, 0x4d, Null, nil},
			{".1.3.6.1.2.1.1.4.0", 0x4e, 0x5b, Null, nil},
			{".1.3.6.1.2.1.43.5.1.1.15.1", 0x5c, 0x6c, Null, nil},
			{".1.3.6.1.2.1.4.21.1.1.127.0.0.1", 0x6d, 0x7f, Null, nil},
			{".1.3.6.1.4.1.23.2.5.1.1.1.4.2", 0x80, 0x92, Null, nil},
			{".1.3.6.1.2.1.1.3.0", 0x93, 0xa0, Null, nil},
		},
	},
	{
		Version1,
		"privatelab",
		SetRequest,
		526895288,
		0,
		portOnOutgoing1,
		"portOnOutgoing1",
		0x11, // pdu start
		0x1f, // vbl start
		0x36, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.4.1.318.1.1.4.4.2.1.3.5", 0x21, 0x36, Integer, 1},
		},
	},
	{
		Version1,
		"privatelab",
		SetRequest,
		1826072803,
		0,
		portOffOutgoing1,
		"portOffOutgoing1",
		0x11, // pdu start
		0x1f, // vbl start
		0x36, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.4.1.318.1.1.4.4.2.1.3.5", 0x21, 0x36, Integer, 2},
		},
	},
	// MrSpock Set stuff
	{
		Version2c,
		"private",
		SetRequest,
		756726019,
		0,
		setOctet1,
		"setOctet1",
		0x0e, // pdu start
		0x1c, // vbl start
		0x32, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.4.1.2863.205.1.1.75.1.0",
				0x1e, 0x32, OctetString, []byte{0x80}},
		},
	},
	{
		Version2c,
		"private",
		SetRequest,
		1000552357,
		0,
		setOctet2,
		"setOctet2",
		0x0e, // pdu start
		0x1c, // vbl start
		0x37, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.4.1.2863.205.1.1.75.2.0",
				0x1e, 0x36, OctetString, "telnet"},
		},
	},
	// MrSpock Set stuff
	{
		Version2c,
		"private",
		SetRequest,
		1664317637,
		0,
		setInteger1,
		"setInteger1",
		0x0e, // pdu start
		0x1c, // vbl start
		0x7f, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.4.1.2863.205.10.1.33.2.5.1.2.2", 0x1e, 0x36, Integer, 5001},
			{".1.3.6.1.4.1.2863.205.10.1.33.2.5.1.3.2", 0x37, 0x4f, Integer, 5001},
			{".1.3.6.1.4.1.2863.205.10.1.33.2.5.1.4.2", 0x50, 0x67, Integer, 2},
			{".1.3.6.1.4.1.2863.205.10.1.33.2.5.1.5.2", 0x68, 0x7f, Integer, 1},
		},
	},
	// Issue 35, empty responses.
	{
		Version2c,
		"public",
		GetRequest,
		1883298028,
		0,
		emptyErrRequest,
		"emptyErrRequest",
		0x0d, // pdu start
		0x1b, // vbl start
		0x1c, // finish
		[]testsEnmarshalVarbindPosition{},
	},
	// trap - TimeTicks
	// snmptrap different with timetick 2, integer 5

	// trap1 - capture is from frame - less work, decode easier
	// varbinds - because Wireshark is decoding as BER's, need to subtract 2
	// from start of varbinds
	{
		Version2c,
		"public",
		SNMPv2Trap,
		1918693186,
		0,
		trap1,
		"trap1",
		0x0e, // pdu start
		0x1c, // vbl start
		0x82, // finish
		[]testsEnmarshalVarbindPosition{
			{".1.3.6.1.2.1.1.3.0", 0x1e, 0x2f, TimeTicks, uint32(18542501)},
			{".1.3.6.1.6.3.1.1.4.1.0", 0x30, 0x45, ObjectIdentifier, ".1.3.6.1.2.1.1"},
			{".1.3.6.1.2.1.1.1.0", 0x46, 0x59, OctetString, "red laptop"},
			{".1.3.6.1.2.1.1.7.0", 0x5e, 0x6c, Integer, 5},
			{".1.3.6.1.2.1.1.2", 0x6d, 0x82, ObjectIdentifier, ".1.3.6.1.4.1.2.3.4.5"},
		},
	},
}

// helpers for Enmarshal tests

// vbPosPdus returns a slice of oids in the given test
func vbPosPdus(test testsEnmarshalT) (pdus []SnmpPDU) {
	for _, vbp := range test.vbPositions {
		pdu := SnmpPDU{vbp.oid, vbp.pduType, vbp.pduValue, nil}
		pdus = append(pdus, pdu)
	}
	return
}

// checkByteEquality walks the bytes in testBytes, and compares them to goodBytes
func checkByteEquality(t *testing.T, test testsEnmarshalT, testBytes []byte,
	start int, finish int) {

	testBytesLen := len(testBytes)

	goodBytes := test.goodBytes()
	goodBytes = goodBytes[start : finish+1]
	for cursor := range goodBytes {
		if testBytesLen < cursor {
			t.Errorf("%s: testBytesLen (%d) < cursor (%d)", test.funcName,
				testBytesLen, cursor)
			break
		}
		if testBytes[cursor] != goodBytes[cursor] {
			t.Errorf("%s: cursor %d: testBytes != goodBytes:\n%s\n%s",
				test.funcName,
				cursor,
				dumpBytes2("good", goodBytes, cursor),
				dumpBytes2("test", testBytes, cursor))
			break
		}
	}
}

// Enmarshal tests in order that should be used for troubleshooting
// ie check each varbind is working, then the varbind list, etc

func TestEnmarshalVarbind(t *testing.T) {
	Default.Logger = log.New(ioutil.Discard, "", 0)

	for _, test := range testsEnmarshal {
		for j, test2 := range test.vbPositions {
			snmppdu := &SnmpPDU{test2.oid, test2.pduType, test2.pduValue, nil}
			testBytes, err := marshalVarbind(snmppdu)
			if err != nil {
				t.Errorf("#%s:%d:%s err returned: %v",
					test.funcName, j, test2.oid, err)
			}

			checkByteEquality(t, test, testBytes, test2.start, test2.finish)
		}
	}
}

func TestEnmarshalVBL(t *testing.T) {
	Default.Logger = log.New(ioutil.Discard, "", 0)

	for _, test := range testsEnmarshal {
		x := &SnmpPacket{
			Community: test.community,
			Version:   test.version,
			RequestID: test.requestid,
			Variables: vbPosPdus(test),
		}

		testBytes, err := x.marshalVBL()
		if err != nil {
			t.Errorf("#%s: marshalVBL() err returned: %v", test.funcName, err)
		}

		checkByteEquality(t, test, testBytes, test.vblStart, test.finish)
	}
}

func TestEnmarshalPDU(t *testing.T) {
	Default.Logger = log.New(ioutil.Discard, "", 0)

	for _, test := range testsEnmarshal {
		x := &SnmpPacket{
			Community: test.community,
			Version:   test.version,
			PDUType:   test.requestType,
			RequestID: test.requestid,
			Variables: vbPosPdus(test),
		}

		testBytes, err := x.marshalPDU()
		if err != nil {
			t.Errorf("#%s: marshalPDU() err returned: %v", test.funcName, err)
		}

		checkByteEquality(t, test, testBytes, test.pduStart, test.finish)
	}
}

func TestEnmarshalMsg(t *testing.T) {
	Default.Logger = log.New(ioutil.Discard, "", 0)

	for _, test := range testsEnmarshal {
		x := &SnmpPacket{
			Community: test.community,
			Version:   test.version,
			PDUType:   test.requestType,
			RequestID: test.requestid,
			MsgID:     test.msgid,
			Variables: vbPosPdus(test),
		}

		testBytes, err := x.marshalMsg()
		if err != nil {
			t.Errorf("#%s: marshal() err returned: %v", test.funcName, err)
		}
		checkByteEquality(t, test, testBytes, 0, test.finish)
	}
}

// -- Unmarshal -----------------------------------------------------------------

var testsUnmarshal = []struct {
	in  func() []byte
	out *SnmpPacket
}{
	{kyoceraResponseBytes,
		&SnmpPacket{
			Version:    Version2c,
			Community:  "public",
			PDUType:    GetResponse,
			RequestID:  1066889284,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.2.1.1.7.0",
					Type:  Integer,
					Value: 104,
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.10.1",
					Type:  Counter32,
					Value: 271070065,
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.5.1",
					Type:  Gauge32,
					Value: 100000000,
				},
				{
					Name:  ".1.3.6.1.2.1.1.4.0",
					Type:  OctetString,
					Value: []byte("Administrator"),
				},
				{
					Name:  ".1.3.6.1.2.1.43.5.1.1.15.1",
					Type:  Null,
					Value: nil,
				},
				{
					Name:  ".1.3.6.1.2.1.4.21.1.1.127.0.0.1",
					Type:  IPAddress,
					Value: "127.0.0.1",
				},
				{
					Name:  ".1.3.6.1.4.1.23.2.5.1.1.1.4.2",
					Type:  OctetString,
					Value: []byte{0x00, 0x15, 0x99, 0x37, 0x76, 0x2b},
				},
				{
					Name:  ".1.3.6.1.2.1.1.3.0",
					Type:  TimeTicks,
					Value: 318870100,
				},
			},
		},
	},
	{ciscoResponseBytes,
		&SnmpPacket{
			Version:    Version2c,
			Community:  "public",
			PDUType:    GetResponse,
			RequestID:  4876669,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.2.1.1.7.0",
					Type:  Integer,
					Value: 78,
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.2.6",
					Type:  OctetString,
					Value: []byte("GigabitEthernet0"),
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.5.3",
					Type:  Gauge32,
					Value: uint(4294967295),
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.7.2",
					Type:  NoSuchInstance,
					Value: nil,
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.9.3",
					Type:  TimeTicks,
					Value: 2970,
				},
				{
					Name:  ".1.3.6.1.2.1.3.1.1.2.10.1.10.11.0.17",
					Type:  OctetString,
					Value: []byte{0x00, 0x07, 0x7d, 0x4d, 0x09, 0x00},
				},
				{
					Name:  ".1.3.6.1.2.1.3.1.1.3.10.1.10.11.0.2",
					Type:  IPAddress,
					Value: "10.11.0.2",
				},
				{
					Name:  ".1.3.6.1.2.1.4.20.1.1.110.143.197.1",
					Type:  IPAddress,
					Value: "110.143.197.1",
				},
				{
					Name:  ".1.3.6.1.66.1",
					Type:  NoSuchObject,
					Value: nil,
				},
				{
					Name:  ".1.3.6.1.2.1.1.2.0",
					Type:  ObjectIdentifier,
					Value: ".1.3.6.1.4.1.9.1.1166",
				},
			},
		},
	},
	{portOnIncoming1,
		&SnmpPacket{
			Version:    Version1,
			Community:  "privatelab",
			PDUType:    GetResponse,
			RequestID:  526895288,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.4.1.318.1.1.4.4.2.1.3.5",
					Type:  Integer,
					Value: 1,
				},
			},
		},
	},
	{portOffIncoming1,
		&SnmpPacket{
			Version:    Version1,
			Community:  "privatelab",
			PDUType:    GetResponse,
			RequestID:  1826072803,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.4.1.318.1.1.4.4.2.1.3.5",
					Type:  Integer,
					Value: 2,
				},
			},
		},
	},
	{ciscoGetnextResponseBytes,
		&SnmpPacket{
			Version:    Version2c,
			Community:  "public",
			PDUType:    GetResponse,
			RequestID:  1528674030,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.2.1.3.1.1.3.2.1.192.168.104.2",
					Type:  IPAddress,
					Value: "192.168.104.2",
				},
				{
					Name:  ".1.3.6.1.2.1.92.1.2.1.0",
					Type:  Counter32,
					Value: 0,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.3.3",
					Type:  OctetString,
					Value: []byte("The MIB module for managing IP and ICMP implementations"),
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.2",
					Type:  TimeTicks,
					Value: 21,
				},
				{
					Name:  ".1.3.6.1.2.1.2.1.0",
					Type:  Integer,
					Value: 3,
				},
				{
					Name:  ".1.3.6.1.2.1.1.2.0",
					Type:  ObjectIdentifier,
					Value: ".1.3.6.1.4.1.8072.3.2.10",
				},
			},
		},
	},
	{ciscoGetbulkResponseBytes,
		&SnmpPacket{
			Version:        Version2c,
			Community:      "public",
			PDUType:        GetResponse,
			RequestID:      250000266,
			NonRepeaters:   0,
			MaxRepetitions: 10,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.1",
					Type:  TimeTicks,
					Value: 21,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.2",
					Type:  TimeTicks,
					Value: 21,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.3",
					Type:  TimeTicks,
					Value: 21,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.4",
					Type:  TimeTicks,
					Value: 21,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.5",
					Type:  TimeTicks,
					Value: 21,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.6",
					Type:  TimeTicks,
					Value: 23,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.7",
					Type:  TimeTicks,
					Value: 23,
				},
				{
					Name:  ".1.3.6.1.2.1.1.9.1.4.8",
					Type:  TimeTicks,
					Value: 23,
				},
				{
					Name:  ".1.3.6.1.2.1.2.1.0",
					Type:  Integer,
					Value: 3,
				},
				{
					Name:  ".1.3.6.1.2.1.2.2.1.1.1",
					Type:  Integer,
					Value: 1,
				},
			},
		},
	},
	{emptyErrResponse,
		&SnmpPacket{
			Version:   Version2c,
			Community: "public",
			PDUType:   GetResponse,
			RequestID: 1883298028,
			Error:     0,
			Variables: []SnmpPDU{},
		},
	},
	{counter64Response,
		&SnmpPacket{
			Version:    Version2c,
			Community:  "public",
			PDUType:    GetResponse,
			RequestID:  190378322,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.2.1.31.1.1.1.10.1",
					Type:  Counter64,
					Value: 1527943,
				},
			},
		},
	},
	{opaqueFloatResponse,
		&SnmpPacket{
			Version:    Version2c,
			Community:  "public",
			PDUType:    GetResponse,
			RequestID:  601216773,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.4.1.6574.4.2.12.1.0",
					Type:  OpaqueFloat,
					Value: float32(10.0),
				},
			},
		},
	},
	{opaqueDoubleResponse,
		&SnmpPacket{
			Version:    Version2c,
			Community:  "public",
			PDUType:    GetResponse,
			RequestID:  601216773,
			Error:      0,
			ErrorIndex: 0,
			Variables: []SnmpPDU{
				{
					Name:  ".1.3.6.1.4.1.6574.4.2.12.1.0",
					Type:  OpaqueDouble,
					Value: float64(10.0),
				},
			},
		},
	},
}

func TestUnmarshal(t *testing.T) {
	Default.Logger = log.New(ioutil.Discard, "", 0)

SANITY:
	for i, test := range testsUnmarshal {
		var err error
		var res = new(SnmpPacket)
		var cursor int

		var buf = test.in()
		cursor, err = Default.unmarshalHeader(buf, res)
		if err != nil {
			t.Errorf("#%d, UnmarshalHeader returned err: %v", i, err)
			continue SANITY
		}
		if res.Version == Version3 {
			buf, cursor, err = Default.decryptPacket(buf, cursor, res)
			if err != nil {
				t.Errorf("#%d, decryptPacket returned err: %v", i, err)
			}
		}
		err = Default.unmarshalPayload(test.in(), cursor, res)
		if err != nil {
			t.Errorf("#%d, UnmarshalPayload returned err: %v", i, err)
		}
		if res == nil {
			t.Errorf("#%d, Unmarshal returned nil", i)
			continue SANITY
		}

		// test "header" fields
		if res.Version != test.out.Version {
			t.Errorf("#%d Version result: %v, test: %v", i, res.Version, test.out.Version)
		}
		if res.Community != test.out.Community {
			t.Errorf("#%d Community result: %v, test: %v", i, res.Community, test.out.Community)
		}
		if res.PDUType != test.out.PDUType {
			t.Errorf("#%d PDUType result: %v, test: %v", i, res.PDUType, test.out.PDUType)
		}
		if res.RequestID != test.out.RequestID {
			t.Errorf("#%d RequestID result: %v, test: %v", i, res.RequestID, test.out.RequestID)
		}
		if res.Error != test.out.Error {
			t.Errorf("#%d Error result: %v, test: %v", i, res.Error, test.out.Error)
		}
		if res.ErrorIndex != test.out.ErrorIndex {
			t.Errorf("#%d ErrorIndex result: %v, test: %v", i, res.ErrorIndex, test.out.ErrorIndex)
		}

		// test varbind values
		for n, vb := range test.out.Variables {
			if len(res.Variables) < n {
				t.Errorf("#%d:%d ran out of varbind results", i, n)
				continue SANITY
			}
			vbr := res.Variables[n]

			if vbr.Name != vb.Name {
				t.Errorf("#%d:%d Name result: %v, test: %v", i, n, vbr.Name, vb.Name)
			}
			if vbr.Type != vb.Type {
				t.Errorf("#%d:%d Type result: %v, test: %v", i, n, vbr.Type, vb.Type)
			}

			switch vb.Type {
			case Integer, Gauge32, Counter32, TimeTicks, Counter64:
				vbval := ToBigInt(vb.Value)
				vbrval := ToBigInt(vbr.Value)
				if vbval.Cmp(vbrval) != 0 {
					t.Errorf("#%d:%d Value result: %v, test: %v", i, n, vbr.Value, vb.Value)
				}
			case OctetString:
				if !bytes.Equal(vb.Value.([]byte), vbr.Value.([]byte)) {
					t.Errorf("#%d:%d Value result: %v, test: %v", i, n, vbr.Value, vb.Value)
				}
			case IPAddress, ObjectIdentifier:
				if vb.Value != vbr.Value {
					t.Errorf("#%d:%d Value result: %v, test: %v", i, n, vbr.Value, vb.Value)
				}
			case Null, NoSuchObject, NoSuchInstance:
				if (vb.Value != nil) || (vbr.Value != nil) {
					t.Errorf("#%d:%d Value result: %v, test: %v", i, n, vbr.Value, vb.Value)
				}
			case OpaqueFloat:
				if vb.Value.(float32) != vbr.Value.(float32) {
					t.Errorf("#%d:%d Value result: %v, test: %v", i, n, vbr.Value, vb.Value)
				}
			case OpaqueDouble:
				if vb.Value.(float64) != vbr.Value.(float64) {
					t.Errorf("#%d:%d Value result: %v, test: %v", i, n, vbr.Value, vb.Value)
				}
			default:
				t.Errorf("#%d:%d Unhandled case result: %v, test: %v", i, n, vbr.Value, vb.Value)
			}

		}
	}
}

// -----------------------------------------------------------------------------

/*

* byte dumps generated using tcpdump and github.com/jteeuwen/go-bindata eg
  `sudo tcpdump -s 0 -i eth0 -w cisco.pcap host 203.50.251.17 and port 161`

* Frame, Ethernet II, IP and UDP layers removed from generated bytes
*/

/*
kyoceraResponseBytes corresponds to the response section of this snmpget

Simple Network Management Protocol
  version: v2c (1)
  community: public
  data: get-response (2)
    get-response
      request-id: 1066889284
      error-status: noError (0)
      error-index: 0
      variable-bindings: 8 items
        1.3.6.1.2.1.1.7.0: 104
        1.3.6.1.2.1.2.2.1.10.1: 271070065
        1.3.6.1.2.1.2.2.1.5.1: 100000000
        1.3.6.1.2.1.1.4.0: 41646d696e6973747261746f72
        1.3.6.1.2.1.43.5.1.1.15.1: Value (Null)
        1.3.6.1.2.1.4.21.1.1.127.0.0.1: 127.0.0.1 (127.0.0.1)
        1.3.6.1.4.1.23.2.5.1.1.1.4.2: 00159937762b
        1.3.6.1.2.1.1.3.0: 318870100
*/

func kyoceraResponseBytes() []byte {
	return []byte{
		0x30, 0x81, 0xc2, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c,
		0x69, 0x63, 0xa2, 0x81, 0xb4, 0x02, 0x04, 0x3f, 0x97, 0x70, 0x44, 0x02,
		0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x81, 0xa5, 0x30, 0x0d, 0x06, 0x08,
		0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x07, 0x00, 0x02, 0x01, 0x68, 0x30,
		0x12, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x02, 0x01, 0x0a,
		0x01, 0x41, 0x04, 0x10, 0x28, 0x33, 0x71, 0x30, 0x12, 0x06, 0x0a, 0x2b,
		0x06, 0x01, 0x02, 0x01, 0x02, 0x02, 0x01, 0x05, 0x01, 0x42, 0x04, 0x05,
		0xf5, 0xe1, 0x00, 0x30, 0x19, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01,
		0x01, 0x04, 0x00, 0x04, 0x0d, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x69, 0x73,
		0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x30, 0x0f, 0x06, 0x0b, 0x2b, 0x06,
		0x01, 0x02, 0x01, 0x2b, 0x05, 0x01, 0x01, 0x0f, 0x01, 0x05, 0x00, 0x30,
		0x15, 0x06, 0x0d, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x04, 0x15, 0x01, 0x01,
		0x7f, 0x00, 0x00, 0x01, 0x40, 0x04, 0x7f, 0x00, 0x00, 0x01, 0x30, 0x17,
		0x06, 0x0d, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x17, 0x02, 0x05, 0x01, 0x01,
		0x01, 0x04, 0x02, 0x04, 0x06, 0x00, 0x15, 0x99, 0x37, 0x76, 0x2b, 0x30,
		0x10, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x03, 0x00, 0x43,
		0x04, 0x13, 0x01, 0x92, 0x54,
	}
}

/*
ciscoResponseBytes corresponds to the response section of this snmpget:

% snmpget -On -v2c -c public 203.50.251.17 1.3.6.1.2.1.1.7.0 1.3.6.1.2.1.2.2.1.2.6 1.3.6.1.2.1.2.2.1.5.3 1.3.6.1.2.1.2.2.1.7.2 1.3.6.1.2.1.2.2.1.9.3 1.3.6.1.2.1.3.1.1.2.10.1.10.11.0.17 1.3.6.1.2.1.3.1.1.3.10.1.10.11.0.2 1.3.6.1.2.1.4.20.1.1.110.143.197.1 1.3.6.1.66.1 1.3.6.1.2.1.1.2.0
.1.3.6.1.2.1.1.7.0 = INTEGER: 78
.1.3.6.1.2.1.2.2.1.2.6 = STRING: GigabitEthernet0
.1.3.6.1.2.1.2.2.1.5.3 = Gauge32: 4294967295
.1.3.6.1.2.1.2.2.1.7.2 = No Such Instance currently exists at this OID
.1.3.6.1.2.1.2.2.1.9.3 = Timeticks: (2970) 0:00:29.70
.1.3.6.1.2.1.3.1.1.2.10.1.10.11.0.17 = Hex-STRING: 00 07 7D 4D 09 00
.1.3.6.1.2.1.3.1.1.3.10.1.10.11.0.2 = Network Address: 0A:0B:00:02
.1.3.6.1.2.1.4.20.1.1.110.143.197.1 = IPAddress: 110.143.197.1
.1.3.6.1.66.1 = No Such Object available on this agent at this OID
.1.3.6.1.2.1.1.2.0 = OID: .1.3.6.1.4.1.9.1.1166
*/

func ciscoResponseBytes() []byte {
	return []byte{
		0x30, 0x81,
		0xf1, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63,
		0xa2, 0x81, 0xe3, 0x02, 0x03, 0x4a, 0x69, 0x7d, 0x02, 0x01, 0x00, 0x02,
		0x01, 0x00, 0x30, 0x81, 0xd5, 0x30, 0x0d, 0x06, 0x08, 0x2b, 0x06, 0x01,
		0x02, 0x01, 0x01, 0x07, 0x00, 0x02, 0x01, 0x4e, 0x30, 0x1e, 0x06, 0x0a,
		0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x02, 0x01, 0x02, 0x06, 0x04, 0x10,
		0x47, 0x69, 0x67, 0x61, 0x62, 0x69, 0x74, 0x45, 0x74, 0x68, 0x65, 0x72,
		0x6e, 0x65, 0x74, 0x30, 0x30, 0x13, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02,
		0x01, 0x02, 0x02, 0x01, 0x05, 0x03, 0x42, 0x05, 0x00, 0xff, 0xff, 0xff,
		0xff, 0x30, 0x0e, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x02,
		0x01, 0x07, 0x02, 0x81, 0x00, 0x30, 0x10, 0x06, 0x0a, 0x2b, 0x06, 0x01,
		0x02, 0x01, 0x02, 0x02, 0x01, 0x09, 0x03, 0x43, 0x02, 0x0b, 0x9a, 0x30,
		0x19, 0x06, 0x0f, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x03, 0x01, 0x01, 0x02,
		0x0a, 0x01, 0x0a, 0x0b, 0x00, 0x11, 0x04, 0x06, 0x00, 0x07, 0x7d, 0x4d,
		0x09, 0x00, 0x30, 0x17, 0x06, 0x0f, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x03,
		0x01, 0x01, 0x03, 0x0a, 0x01, 0x0a, 0x0b, 0x00, 0x02, 0x40, 0x04, 0x0a,
		0x0b, 0x00, 0x02, 0x30, 0x17, 0x06, 0x0f, 0x2b, 0x06, 0x01, 0x02, 0x01,
		0x04, 0x14, 0x01, 0x01, 0x6e, 0x81, 0x0f, 0x81, 0x45, 0x01, 0x40, 0x04,
		0x6e, 0x8f, 0xc5, 0x01, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x06, 0x01, 0x42,
		0x01, 0x80, 0x00, 0x30, 0x15, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01,
		0x01, 0x02, 0x00, 0x06, 0x09, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x09, 0x01,
		0x89, 0x0e,
	}
}

/*
kyoceraRequestBytes corresponds to the request section of this snmpget:

snmpget -On -v2c -c public 192.168.1.10 1.3.6.1.2.1.1.7.0 1.3.6.1.2.1.2.2.1.10.1 1.3.6.1.2.1.2.2.1.5.1 1.3.6.1.2.1.1.4.0 1.3.6.1.2.1.43.5.1.1.15.1 1.3.6.1.2.1.4.21.1.1.127.0.0.1 1.3.6.1.4.1.23.2.5.1.1.1.4.2 1.3.6.1.2.1.1.3.0
.1.3.6.1.2.1.1.7.0 = INTEGER: 104
.1.3.6.1.2.1.2.2.1.10.1 = Counter32: 144058856
.1.3.6.1.2.1.2.2.1.5.1 = Gauge32: 100000000
.1.3.6.1.2.1.1.4.0 = STRING: "Administrator"
.1.3.6.1.2.1.43.5.1.1.15.1 = NULL
.1.3.6.1.2.1.4.21.1.1.127.0.0.1 = IPAddress: 127.0.0.1
.1.3.6.1.4.1.23.2.5.1.1.1.4.2 = Hex-STRING: 00 15 99 37 76 2B
.1.3.6.1.2.1.1.3.0 = Timeticks: (120394900) 13 days, 22:25:49.00
*/

func kyoceraRequestBytes() []byte {
	return []byte{
		0x30, 0x81,
		0x9e, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63,
		0xa0, 0x81, 0x90, 0x02, 0x04, 0x6f, 0x8c, 0xee, 0x64, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x81, 0x81, 0x30, 0x0c, 0x06, 0x08, 0x2b, 0x06,
		0x01, 0x02, 0x01, 0x01, 0x07, 0x00, 0x05, 0x00, 0x30, 0x0e, 0x06, 0x0a,
		0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x02, 0x01, 0x0a, 0x01, 0x05, 0x00,
		0x30, 0x0e, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x02, 0x01,
		0x05, 0x01, 0x05, 0x00, 0x30, 0x0c, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02,
		0x01, 0x01, 0x04, 0x00, 0x05, 0x00, 0x30, 0x0f, 0x06, 0x0b, 0x2b, 0x06,
		0x01, 0x02, 0x01, 0x2b, 0x05, 0x01, 0x01, 0x0f, 0x01, 0x05, 0x00, 0x30,
		0x11, 0x06, 0x0d, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x04, 0x15, 0x01, 0x01,
		0x7f, 0x00, 0x00, 0x01, 0x05, 0x00, 0x30, 0x11, 0x06, 0x0d, 0x2b, 0x06,
		0x01, 0x04, 0x01, 0x17, 0x02, 0x05, 0x01, 0x01, 0x01, 0x04, 0x02, 0x05,
		0x00, 0x30, 0x0c, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x03,
		0x00, 0x05, 0x00,
	}
}

// === snmpset dumps ===

/*
port_on_*1() correspond to this snmpset and response:

snmpset -v 1 -c privatelab 192.168.100.124 .1.3.6.1.4.1.318.1.1.4.4.2.1.3.5 i 1

Simple Network Management Protocol
  version: version-1 (0)
  community: privatelab
  data: set-request (3)
    set-request
      request-id: 526895288
      error-status: noError (0)
      error-index: 0
      variable-bindings: 1 item
        1.3.6.1.4.1.318.1.1.4.4.2.1.3.5:
          Object Name: 1.3.6.1.4.1.318.1.1.4.4.2.1.3.5 (iso.3.6.1.4.1.318.1.1.4.4.2.1.3.5)
          Value (Integer32): 1

Simple Network Management Protocol
  version: version-1 (0)
  community: privatelab
  data: get-response (2)
    get-response
      request-id: 526895288
      error-status: noError (0)
      error-index: 0
      variable-bindings: 1 item
        1.3.6.1.4.1.318.1.1.4.4.2.1.3.5:
          Object Name: 1.3.6.1.4.1.318.1.1.4.4.2.1.3.5 (iso.3.6.1.4.1.318.1.1.4.4.2.1.3.5)
          Value (Integer32): 1
*/

func portOnOutgoing1() []byte {
	return []byte{
		0x30, 0x35, 0x02, 0x01, 0x00, 0x04, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61,
		0x74, 0x65, 0x6c, 0x61, 0x62, 0xa3, 0x24, 0x02, 0x04, 0x1f, 0x67, 0xc8,
		0xb8, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x16, 0x30, 0x14, 0x06,
		0x0f, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x82, 0x3e, 0x01, 0x01, 0x04, 0x04,
		0x02, 0x01, 0x03, 0x05, 0x02, 0x01, 0x01,
	}
}

func portOnIncoming1() []byte {
	return []byte{
		0x30, 0x82, 0x00, 0x35, 0x02, 0x01, 0x00, 0x04, 0x0a, 0x70, 0x72, 0x69,
		0x76, 0x61, 0x74, 0x65, 0x6c, 0x61, 0x62, 0xa2, 0x24, 0x02, 0x04, 0x1f,
		0x67, 0xc8, 0xb8, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x16, 0x30,
		0x14, 0x06, 0x0f, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x82, 0x3e, 0x01, 0x01,
		0x04, 0x04, 0x02, 0x01, 0x03, 0x05, 0x02, 0x01, 0x01,
	}
}

/*
port_off_*1() correspond to this snmpset and response:

snmpset -v 1 -c privatelab 192.168.100.124 .1.3.6.1.4.1.318.1.1.4.4.2.1.3.5 i 2

Simple Network Management Protocol
  version: version-1 (0)
  community: privatelab
  data: set-request (3)
    set-request
      request-id: 1826072803
      error-status: noError (0)
      error-index: 0
      variable-bindings: 1 item
        1.3.6.1.4.1.318.1.1.4.4.2.1.3.5:
          Object Name: 1.3.6.1.4.1.318.1.1.4.4.2.1.3.5 (iso.3.6.1.4.1.318.1.1.4.4.2.1.3.5)
          Value (Integer32): 2

Simple Network Management Protocol
  version: version-1 (0)
  community: privatelab
  data: get-response (2)
    get-response
      request-id: 1826072803
      error-status: noError (0)
      error-index: 0
      variable-bindings: 1 item
        1.3.6.1.4.1.318.1.1.4.4.2.1.3.5:
          Object Name: 1.3.6.1.4.1.318.1.1.4.4.2.1.3.5 (iso.3.6.1.4.1.318.1.1.4.4.2.1.3.5)
          Value (Integer32): 2
*/

func portOffOutgoing1() []byte {
	return []byte{
		0x30, 0x35, 0x02, 0x01, 0x00, 0x04, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61,
		0x74, 0x65, 0x6c, 0x61, 0x62, 0xa3, 0x24, 0x02, 0x04, 0x6c, 0xd7, 0xa8,
		0xe3, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x16, 0x30, 0x14, 0x06,
		0x0f, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x82, 0x3e, 0x01, 0x01, 0x04, 0x04,
		0x02, 0x01, 0x03, 0x05, 0x02, 0x01, 0x02,
	}
}

func portOffIncoming1() []byte {
	return []byte{
		0x30, 0x82, 0x00, 0x35, 0x02, 0x01, 0x00, 0x04, 0x0a, 0x70, 0x72, 0x69,
		0x76, 0x61, 0x74, 0x65, 0x6c, 0x61, 0x62, 0xa2, 0x24, 0x02, 0x04, 0x6c,
		0xd7, 0xa8, 0xe3, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x16, 0x30,
		0x14, 0x06, 0x0f, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x82, 0x3e, 0x01, 0x01,
		0x04, 0x04, 0x02, 0x01, 0x03, 0x05, 0x02, 0x01, 0x02,
	}
}

// MrSpock START

/*
setOctet1:
Simple Network Management Protocol
  version: v2c (1)
  community: private
  data: set-request (3)
    set-request
      request-id: 756726019
      error-status: noError (0)
      error-index: 0
      variable-bindings: 1 item
        1.3.6.1.4.1.2863.205.1.1.75.1.0: 80
          Object Name: 1.3.6.1.4.1.2863.205.1.1.75.1.0 (iso.3.6.1.4.1.2863.205.1.1.75.1.0)
          Value (OctetString): 80

setOctet2:
Simple Network Management Protocol
    version: v2c (1)
    community: private
    data: set-request (3)
        set-request
            request-id: 1000552357
            error-status: noError (0)
            error-index: 0
            variable-bindings: 1 item
                1.3.6.1.4.1.2863.205.1.1.75.2.0: 74656c6e6574
                    Object Name: 1.3.6.1.4.1.2863.205.1.1.75.2.0 (iso.3.6.1.4.1.2863.205.1.1.75.2.0)
                    Value (OctetString): 74656c6e6574 ("telnet")
*/

func setOctet1() []byte {
	return []byte{
		0x30, 0x31, 0x02, 0x01, 0x01, 0x04, 0x07, 0x70, 0x72, 0x69, 0x76, 0x61,
		0x74, 0x65, 0xa3, 0x23, 0x02, 0x04, 0x2d, 0x1a, 0xb9, 0x03, 0x02, 0x01,
		0x00, 0x02, 0x01, 0x00, 0x30, 0x15, 0x30, 0x13, 0x06, 0x0e, 0x2b, 0x06,
		0x01, 0x04, 0x01, 0x96, 0x2f, 0x81, 0x4d, 0x01, 0x01, 0x4b, 0x01, 0x00,
		0x04, 0x01, 0x80,
	}
}

func setOctet2() []byte {
	return []byte{
		0x30, 0x36, 0x02, 0x01, 0x01, 0x04, 0x07, 0x70, 0x72, 0x69, 0x76, 0x61,
		0x74, 0x65, 0xa3, 0x28, 0x02, 0x04, 0x3b, 0xa3, 0x37, 0xa5, 0x02, 0x01,
		0x00, 0x02, 0x01, 0x00, 0x30, 0x1a, 0x30, 0x18, 0x06, 0x0e, 0x2b, 0x06,
		0x01, 0x04, 0x01, 0x96, 0x2f, 0x81, 0x4d, 0x01, 0x01, 0x4b, 0x02, 0x00,
		0x04, 0x06, 0x74, 0x65, 0x6c, 0x6e, 0x65, 0x74,
	}
}

/* setInteger1:
snmpset -c private -v2c 10.80.0.14 \
	.1.3.6.1.4.1.2863.205.10.1.33.2.5.1.2.2 i 5001 \
	.1.3.6.1.4.1.2863.205.10.1.33.2.5.1.3.2 i 5001 \
	.1.3.6.1.4.1.2863.205.10.1.33.2.5.1.4.2 i 2 \
	.1.3.6.1.4.1.2863.205.10.1.33.2.5.1.5.2 i 1

Simple Network Management Protocol
 version: v2c (1)
 community: private
 data: set-request (3)
  set-request
   request-id: 1664317637
   error-status: noError (0)
   error-index: 0
   variable-bindings: 4 items
    1.3.6.1.4.1.2863.205.10.1.33.2.5.1.2.2:
     Object Name: 1.3.6.1.4.1.2863.205.10.1.33.2.5.1.2.2 (iso.3.6.1.4.1.2863.205.10.1.33.2.5.1.2.2)
     Value (Integer32): 5001
    1.3.6.1.4.1.2863.205.10.1.33.2.5.1.3.2:
     Object Name: 1.3.6.1.4.1.2863.205.10.1.33.2.5.1.3.2 (iso.3.6.1.4.1.2863.205.10.1.33.2.5.1.3.2)
     Value (Integer32): 5001
    1.3.6.1.4.1.2863.205.10.1.33.2.5.1.4.2:
     Object Name: 1.3.6.1.4.1.2863.205.10.1.33.2.5.1.4.2 (iso.3.6.1.4.1.2863.205.10.1.33.2.5.1.4.2)
     Value (Integer32): 2
    1.3.6.1.4.1.2863.205.10.1.33.2.5.1.5.2:
     Object Name: 1.3.6.1.4.1.2863.205.10.1.33.2.5.1.5.2 (iso.3.6.1.4.1.2863.205.10.1.33.2.5.1.5.2)
     Value (Integer32): 1
*/

func setInteger1() []byte {
	return []byte{
		0x30, 0x7e, 0x02, 0x01, 0x01, 0x04, 0x07, 0x70, 0x72, 0x69, 0x76, 0x61,
		0x74, 0x65, 0xa3, 0x70, 0x02, 0x04, 0x63, 0x33, 0x78, 0xc5, 0x02, 0x01,
		0x00, 0x02, 0x01, 0x00, 0x30, 0x62, 0x30, 0x17, 0x06, 0x11, 0x2b, 0x06,
		0x01, 0x04, 0x01, 0x96, 0x2f, 0x81, 0x4d, 0x0a, 0x01, 0x21, 0x02, 0x05,
		0x01, 0x02, 0x02, 0x02, 0x02, 0x13, 0x89, 0x30, 0x17, 0x06, 0x11, 0x2b,
		0x06, 0x01, 0x04, 0x01, 0x96, 0x2f, 0x81, 0x4d, 0x0a, 0x01, 0x21, 0x02,
		0x05, 0x01, 0x03, 0x02, 0x02, 0x02, 0x13, 0x89, 0x30, 0x16, 0x06, 0x11,
		0x2b, 0x06, 0x01, 0x04, 0x01, 0x96, 0x2f, 0x81, 0x4d, 0x0a, 0x01, 0x21,
		0x02, 0x05, 0x01, 0x04, 0x02, 0x02, 0x01, 0x02, 0x30, 0x16, 0x06, 0x11,
		0x2b, 0x06, 0x01, 0x04, 0x01, 0x96, 0x2f, 0x81, 0x4d, 0x0a, 0x01, 0x21,
		0x02, 0x05, 0x01, 0x05, 0x02, 0x02, 0x01, 0x01,
	}
}

// MrSpock FINISH

func ciscoGetnextResponseBytes() []byte {
	return []byte{
		0x30, 0x81,
		0xc8, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63,
		0xa2, 0x81, 0xba, 0x02, 0x04, 0x5b, 0x1d, 0xb6, 0xee, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x81, 0xab, 0x30, 0x19, 0x06, 0x11, 0x2b, 0x06,
		0x01, 0x02, 0x01, 0x03, 0x01, 0x01, 0x03, 0x02, 0x01, 0x81, 0x40, 0x81,
		0x28, 0x68, 0x02, 0x40, 0x04, 0xc0, 0xa8, 0x68, 0x02, 0x30, 0x0f, 0x06,
		0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x5c, 0x01, 0x02, 0x01, 0x00, 0x41,
		0x01, 0x00, 0x30, 0x45, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01,
		0x09, 0x01, 0x03, 0x03, 0x04, 0x37, 0x54, 0x68, 0x65, 0x20, 0x4d, 0x49,
		0x42, 0x20, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x20, 0x66, 0x6f, 0x72,
		0x20, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x20, 0x49, 0x50,
		0x20, 0x61, 0x6e, 0x64, 0x20, 0x49, 0x43, 0x4d, 0x50, 0x20, 0x69, 0x6d,
		0x70, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
		0x73, 0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x09,
		0x01, 0x04, 0x02, 0x43, 0x01, 0x15, 0x30, 0x0d, 0x06, 0x08, 0x2b, 0x06,
		0x01, 0x02, 0x01, 0x02, 0x01, 0x00, 0x02, 0x01, 0x03, 0x30, 0x16, 0x06,
		0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x02, 0x00, 0x06, 0x0a, 0x2b,
		0x06, 0x01, 0x04, 0x01, 0xbf, 0x08, 0x03, 0x02, 0x0a,
	}
}

func ciscoGetnextRequestBytes() []byte {
	return []byte{
		0x30, 0x7e,
		0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0xa1,
		0x71, 0x02, 0x04, 0x5b, 0x1d, 0xb6, 0xee, 0x02, 0x01, 0x00, 0x02, 0x01,
		0x00, 0x30, 0x63, 0x30, 0x15, 0x06, 0x11, 0x2b, 0x06, 0x01, 0x02, 0x01,
		0x03, 0x01, 0x01, 0x03, 0x02, 0x01, 0x81, 0x40, 0x81, 0x28, 0x68, 0x01,
		0x05, 0x00, 0x30, 0x0c, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x5c,
		0x01, 0x02, 0x05, 0x00, 0x30, 0x0e, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02,
		0x01, 0x01, 0x09, 0x01, 0x03, 0x02, 0x05, 0x00, 0x30, 0x0e, 0x06, 0x0a,
		0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x09, 0x01, 0x04, 0x01, 0x05, 0x00,
		0x30, 0x0e, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x09, 0x01,
		0x04, 0x08, 0x05, 0x00, 0x30, 0x0c, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02,
		0x01, 0x01, 0x01, 0x00, 0x05, 0x00,
	}
}

/* cisco getbulk bytes corresponds to this snmpbulkget command:

$ snmpbulkget -v2c -cpublic  127.0.0.1:161 1.3.6.1.2.1.1.9.1.3.52
iso.3.6.1.2.1.1.9.1.4.1 = Timeticks: (21) 0:00:00.21
iso.3.6.1.2.1.1.9.1.4.2 = Timeticks: (21) 0:00:00.21
iso.3.6.1.2.1.1.9.1.4.3 = Timeticks: (21) 0:00:00.21
iso.3.6.1.2.1.1.9.1.4.4 = Timeticks: (21) 0:00:00.21
iso.3.6.1.2.1.1.9.1.4.5 = Timeticks: (21) 0:00:00.21
iso.3.6.1.2.1.1.9.1.4.6 = Timeticks: (23) 0:00:00.23
iso.3.6.1.2.1.1.9.1.4.7 = Timeticks: (23) 0:00:00.23
iso.3.6.1.2.1.1.9.1.4.8 = Timeticks: (23) 0:00:00.23
iso.3.6.1.2.1.2.1.0 = INTEGER: 3
iso.3.6.1.2.1.2.2.1.1.1 = INTEGER: 1

*/
func ciscoGetbulkRequestBytes() []byte {
	return []byte{
		0x30, 0x2b,
		0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0xa5,
		0x1e, 0x02, 0x04, 0x7d, 0x89, 0x68, 0xda, 0x02, 0x01, 0x00, 0x02, 0x01,
		0x0a, 0x30, 0x10, 0x30, 0x0e, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01,
		0x01, 0x09, 0x01, 0x03, 0x34, 0x05, 0x00, 0x00,
	}
}

func ciscoGetbulkResponseBytes() []byte {
	return []byte{
		0x30, 0x81,
		0xc5, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63,
		0xa2, 0x81, 0xb7, 0x02, 0x04, 0x0e, 0xe6, 0xb3, 0x8a, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x81, 0xa8, 0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06,
		0x01, 0x02, 0x01, 0x01, 0x09, 0x01, 0x04, 0x01, 0x43, 0x01, 0x15, 0x30,
		0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x09, 0x01, 0x04,
		0x02, 0x43, 0x01, 0x15, 0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02,
		0x01, 0x01, 0x09, 0x01, 0x04, 0x03, 0x43, 0x01, 0x15, 0x30, 0x0f, 0x06,
		0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x09, 0x01, 0x04, 0x04, 0x43,
		0x01, 0x15, 0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01,
		0x09, 0x01, 0x04, 0x05, 0x43, 0x01, 0x15, 0x30, 0x0f, 0x06, 0x0a, 0x2b,
		0x06, 0x01, 0x02, 0x01, 0x01, 0x09, 0x01, 0x04, 0x06, 0x43, 0x01, 0x17,
		0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x09, 0x01,
		0x04, 0x07, 0x43, 0x01, 0x17, 0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01,
		0x02, 0x01, 0x01, 0x09, 0x01, 0x04, 0x08, 0x43, 0x01, 0x17, 0x30, 0x0d,
		0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x01, 0x00, 0x02, 0x01,
		0x03, 0x30, 0x0f, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x02, 0x02,
		0x01, 0x01, 0x01, 0x02, 0x01, 0x01,
	}
}

/*
Issue 35, empty responses.
Simple Network Management Protocol
    version: v2c (1)
    community: public
    data: get-request (0)
        get-request
            request-id: 1883298028
            error-status: noError (0)
            error-index: 0
            variable-bindings: 0 items
*/
func emptyErrRequest() []byte {
	return []byte{
		0x30, 0x1b, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69,
		0x63, 0xa0, 0x0e, 0x02, 0x04, 0x70, 0x40, 0xd8, 0xec, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x00,
	}
}

/*
Issue 35, empty responses.

Simple Network Management Protocol
    version: v2c (1)
    community: public
    data: get-response (2)
        get-response
            request-id: 1883298028
            error-status: noError (0)
            error-index: 0
            variable-bindings: 0 items
*/
func emptyErrResponse() []byte {
	return []byte{
		0x30, 0x1b, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69,
		0x63, 0xa2, 0x0e, 0x02, 0x04, 0x70, 0x40, 0xd8, 0xec, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x00,
	}
}

/*
Issue 15, test Counter64.

Simple Network Management Protocol
    version: v2c (1)
    community: public
    data: get-response (2)
        get-response
            request-id: 190378322
            error-status: noError (0)
            error-index: 0
            variable-bindings: 1 item
                1.3.6.1.2.1.31.1.1.1.10.1: 1527943
                    Object Name: 1.3.6.1.2.1.31.1.1.1.10.1 (iso.3.6.1.2.1.31.1.1.1.10.1)
                    Value (Counter64): 1527943
*/
func counter64Response() []byte {
	return []byte{
		0x30, 0x2f, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69,
		0x63, 0xa2, 0x22, 0x02, 0x04, 0x0b, 0x58, 0xf1, 0x52, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x14, 0x30, 0x12, 0x06, 0x0b, 0x2b, 0x06, 0x01,
		0x02, 0x01, 0x1f, 0x01, 0x01, 0x01, 0x0a, 0x01, 0x46, 0x03, 0x17, 0x50,
		0x87,
	}
}

/*
Opaque Float, observed from Synology NAS UPS MIB
 snmpget -v 2c -c public host 1.3.6.1.4.1.6574.4.2.12.1.0
*/
func opaqueFloatResponse() []byte {
	return []byte{
		0x30, 0x34, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69,
		0x63, 0xa2, 0x27, 0x02, 0x04, 0x23, 0xd5, 0xd7, 0x05, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x19, 0x30, 0x17, 0x06, 0x0c, 0x2b, 0x06, 0x01,
		0x04, 0x01, 0xb3, 0x2e, 0x04, 0x02, 0x0c, 0x01, 0x00, 0x44, 0x07, 0x9f,
		0x78, 0x04, 0x41, 0x20, 0x00, 0x00,
	}
}

/*
Opaque Double, not observed, crafted based on description:
 https://tools.ietf.org/html/draft-perkins-float-00
*/
func opaqueDoubleResponse() []byte {
	return []byte{
		0x30, 0x38, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69,
		0x63, 0xa2, 0x2b, 0x02, 0x04, 0x23, 0xd5, 0xd7, 0x05, 0x02, 0x01, 0x00,
		0x02, 0x01, 0x00, 0x30, 0x1d, 0x30, 0x17, 0x06, 0x0c, 0x2b, 0x06, 0x01,
		0x04, 0x01, 0xb3, 0x2e, 0x04, 0x02, 0x0c, 0x01, 0x00, 0x44, 0x0b, 0x9f,
		0x79, 0x08, 0x40, 0x24, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func TestUnmarshalEmptyPanic(t *testing.T) {
	var in = []byte{}
	var res = new(SnmpPacket)

	_, err := Default.unmarshalHeader(in, res)
	if err == nil {
		t.Errorf("unmarshalHeader did not gracefully detect empty packet")
	}
}

func TestSendOneRequest_dups(t *testing.T) {
	srvr, err := net.ListenUDP("udp4", &net.UDPAddr{})
	defer srvr.Close()

	x := &GoSNMP{
		Version: Version2c,
		Target:  srvr.LocalAddr().(*net.UDPAddr).IP.String(),
		Port:    uint16(srvr.LocalAddr().(*net.UDPAddr).Port),
		Timeout: time.Millisecond * 100,
		Retries: 2,
	}
	if err := x.Connect(); err != nil {
		t.Fatalf("Error connecting: %s", err)
	}

	go func() {
		buf := make([]byte, 256)
		for {
			n, addr, err := srvr.ReadFrom(buf)
			if err != nil {
				return
			}
			buf := buf[:n]

			var reqPkt SnmpPacket
			var cursor int
			cursor, err = x.unmarshalHeader(buf, &reqPkt)
			if err != nil {
				t.Errorf("Error: %s", err)
			}
			// if x.Version == Version3 {
			//	buf, cursor, err = x.decryptPacket(buf, cursor, &reqPkt)
			//	if err != nil {
			//		t.Errorf("Error: %s", err)
			//	}
			//}
			err = x.unmarshalPayload(buf, cursor, &reqPkt)
			if err != nil {
				t.Errorf("Error: %s", err)
			}

			rspPkt := x.mkSnmpPacket(GetResponse, []SnmpPDU{
				{
					Name:  ".1.2",
					Type:  Integer,
					Value: 123,
				},
			}, 0, 0)
			rspPkt.RequestID = reqPkt.RequestID
			outBuf, err := rspPkt.marshalMsg()
			if err != nil {
				t.Errorf("ERR: %s", err)
			}
			srvr.WriteTo(outBuf, addr)
			for i := 0; i <= x.Retries; i++ {
				srvr.WriteTo(outBuf, addr)
			}
		}
	}()

	pdus := []SnmpPDU{SnmpPDU{Name: ".1.2", Type: Null}}
	reqPkt := x.mkSnmpPacket(GetResponse, pdus, 0, 0) //not actually a GetResponse, but we need something our test server can unmarshal

	_, err = x.sendOneRequest(reqPkt, true)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	_, err = x.sendOneRequest(reqPkt, true)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
}

func BenchmarkSendOneRequest(b *testing.B) {
	b.StopTimer()

	srvr, err := net.ListenUDP("udp4", &net.UDPAddr{})
	defer srvr.Close()

	x := &GoSNMP{
		Version: Version2c,
		Target:  srvr.LocalAddr().(*net.UDPAddr).IP.String(),
		Port:    uint16(srvr.LocalAddr().(*net.UDPAddr).Port),
		Timeout: time.Millisecond * 100,
		Retries: 2,
	}
	if err := x.Connect(); err != nil {
		b.Fatalf("Error connecting: %s", err)
	}

	go func() {
		buf := make([]byte, 256)
		outBuf := counter64Response()
		for {
			_, addr, err := srvr.ReadFrom(buf)
			if err != nil {
				return
			}

			copy(outBuf[17:21], buf[11:15]) // evil: copy request ID
			srvr.WriteTo(outBuf, addr)
		}
	}()

	pdus := []SnmpPDU{SnmpPDU{Name: ".1.3.6.1.2.1.31.1.1.1.10.1", Type: Null}}
	reqPkt := x.mkSnmpPacket(GetRequest, pdus, 0, 0)

	// make sure everything works before starting the test
	_, err = x.sendOneRequest(reqPkt, true)
	if err != nil {
		b.Fatalf("Precheck failed: %s", err)
	}

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		_, err = x.sendOneRequest(reqPkt, true)
		if err != nil {
			b.Fatalf("Error: %s", err)
			return
		}
	}
}

/*
$ snmptrap -v 2c -c public 192.168.1.10 '' SNMPv2-MIB::system SNMPv2-MIB::sysDescr.0 s "red laptop" SNMPv2-MIB::sysServices.0 i "5"

Simple Network Management Protocol
    version: v2c (1)
    community: public
    data: snmpV2-trap (7)
        snmpV2-trap
            request-id: 1271509950
            error-status: noError (0)
            error-index: 0
            variable-bindings: 5 items
                1.3.6.1.2.1.1.3.0: 1034156
                    Object Name: 1.3.6.1.2.1.1.3.0 (iso.3.6.1.2.1.1.3.0)
                    Value (Timeticks): 1034156
                1.3.6.1.6.3.1.1.4.1.0: 1.3.6.1.2.1.1 (iso.3.6.1.2.1.1)
                    Object Name: 1.3.6.1.6.3.1.1.4.1.0 (iso.3.6.1.6.3.1.1.4.1.0)
                    Value (OID): 1.3.6.1.2.1.1 (iso.3.6.1.2.1.1)
                1.3.6.1.2.1.1.1.0: 726564206c6170746f70
                    Object Name: 1.3.6.1.2.1.1.1.0 (iso.3.6.1.2.1.1.1.0)
                    Value (OctetString): 726564206c6170746f70
                        Variable-binding-string: red laptop
                1.3.6.1.2.1.1.7.0:
                    Object Name: 1.3.6.1.2.1.1.7.0 (iso.3.6.1.2.1.1.7.0)
                    Value (Integer32): 5
                1.3.6.1.2.1.1.2: 1.3.6.1.4.1.2.3.4.5 (iso.3.6.1.4.1.2.3.4.5)
                    Object Name: 1.3.6.1.2.1.1.2 (iso.3.6.1.2.1.1.2)
                    Value (OID): 1.3.6.1.4.1.2.3.4.5 (iso.3.6.1.4.1.2.3.4.5)
*/

func trap1() []byte {
	return []byte{
		0x30, 0x81, 0x80, 0x02, 0x01, 0x01, 0x04, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0xa7, 0x73,
		0x02, 0x04, 0x72, 0x5c, 0xef, 0x42, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x65, 0x30, 0x10,
		0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x03, 0x00, 0x43, 0x04, 0x01, 0x1a, 0xef, 0xa5,
		0x30, 0x14, 0x06, 0x0a, 0x2b, 0x06, 0x01, 0x06, 0x03, 0x01, 0x01, 0x04, 0x01, 0x00, 0x06, 0x06,
		0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x30, 0x16, 0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01,
		0x01, 0x00, 0x04, 0x0a, 0x72, 0x65, 0x64, 0x20, 0x6c, 0x61, 0x70, 0x74, 0x6f, 0x70, 0x30, 0x0d,
		0x06, 0x08, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x07, 0x00, 0x02, 0x01, 0x05, 0x30, 0x14, 0x06,
		0x07, 0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x02, 0x06, 0x09, 0x2b, 0x06, 0x01, 0x04, 0x01, 0x02,
		0x03, 0x04, 0x05, 0x00, 0x00, 0x00, 0xd0, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x68, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x74, 0x3a, 0x05, 0x00, 0xa1, 0x27, 0x42, 0x0c, 0x46, 0x00,
		0x00, 0x00, 0x46, 0x00, 0x00, 0x00, 0x10, 0x4a, 0x7d, 0x34, 0x3a, 0xa5, 0x74, 0xda, 0x38, 0x4d,
		0x6c, 0x6c, 0x08, 0x00, 0x45, 0x00, 0x00, 0x38, 0xcc, 0xdb, 0x40, 0x00, 0xff, 0x01, 0x2b, 0x74,
		0xc0, 0xa8, 0x01, 0x0a, 0xc0, 0xa8, 0x01, 0x1a, 0x03, 0x03, 0x11, 0x67, 0x00, 0x00, 0x00, 0x00,
		0x45, 0x00, 0x00, 0x9f, 0xe6, 0x8f, 0x40, 0x00, 0x40, 0x11, 0x00, 0x00, 0xc0, 0xa8, 0x01, 0x1a,
		0xc0, 0xa8, 0x01, 0x0a, 0xaf, 0x78, 0x00, 0xa2, 0x00, 0x8b, 0x0b, 0x3a, 0x00, 0x00, 0x68, 0x00,
		0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x6c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x74, 0x3a,
		0x05, 0x00, 0xca, 0x94, 0x67, 0x0c, 0x01, 0x00, 0x1c, 0x00, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x65,
		0x72, 0x73, 0x20, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x64, 0x20, 0x62, 0x79, 0x20, 0x64,
		0x75, 0x6d, 0x70, 0x63, 0x61, 0x70, 0x02, 0x00, 0x08, 0x00, 0x74, 0x3a, 0x05, 0x00, 0xdf, 0xba,
		0x27, 0x0c, 0x03, 0x00, 0x08, 0x00, 0x74, 0x3a, 0x05, 0x00, 0x18, 0x94, 0x67, 0x0c, 0x04, 0x00,
		0x08, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x08, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6c, 0x00, 0x00, 0x00}
}
