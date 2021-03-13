// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phpfpm

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

const requestID uint16 = 1

var sizeTests = []struct {
	size  uint32
	bytes []byte
}{
	{0, []byte{0x00}},
	{127, []byte{0x7F}},
	{128, []byte{0x80, 0x00, 0x00, 0x80}},
	{1000, []byte{0x80, 0x00, 0x03, 0xE8}},
	{33554431, []byte{0x81, 0xFF, 0xFF, 0xFF}},
}

func TestSize(t *testing.T) {
	b := make([]byte, 4)
	for i, test := range sizeTests {
		n := encodeSize(b, test.size)
		if !bytes.Equal(b[:n], test.bytes) {
			t.Errorf("%d expected %x, encoded %x", i, test.bytes, b)
		}
		size, n := readSize(test.bytes)
		if size != test.size {
			t.Errorf("%d expected %d, read %d", i, test.size, size)
		}
		if len(test.bytes) != n {
			t.Errorf("%d did not consume all the bytes", i)
		}
	}
}

var streamTests = []struct {
	desc    string
	recType recType
	reqID   uint16
	content []byte
	raw     []byte
}{
	{"single record", typeStdout, 1, nil,
		[]byte{1, byte(typeStdout), 0, 1, 0, 0, 0, 0},
	},
	// this data will have to be split into two records
	{"two records", typeStdin, 300, make([]byte, 66000),
		bytes.Join([][]byte{
			// header for the first record
			{1, byte(typeStdin), 0x01, 0x2C, 0xFF, 0xFF, 1, 0},
			make([]byte, 65536),
			// header for the second
			{1, byte(typeStdin), 0x01, 0x2C, 0x01, 0xD1, 7, 0},
			make([]byte, 472),
			// header for the empty record
			{1, byte(typeStdin), 0x01, 0x2C, 0, 0, 0, 0},
		},
			nil),
	},
}

type nilCloser struct {
	io.ReadWriter
}

func (c *nilCloser) Close() error { return nil }

func TestStreams(t *testing.T) {
	var rec record
outer:
	for _, test := range streamTests {
		buf := bytes.NewBuffer(test.raw)
		var content []byte
		for buf.Len() > 0 {
			if err := rec.read(buf); err != nil {
				t.Errorf("%s: error reading record: %v", test.desc, err)
				continue outer
			}
			content = append(content, rec.content()...)
		}
		if rec.h.Type != test.recType {
			t.Errorf("%s: got type %d expected %d", test.desc, rec.h.Type, test.recType)
			continue
		}
		if rec.h.ID != test.reqID {
			t.Errorf("%s: got request ID %d expected %d", test.desc, rec.h.ID, test.reqID)
			continue
		}
		if !bytes.Equal(content, test.content) {
			t.Errorf("%s: read wrong content", test.desc)
			continue
		}
		buf.Reset()
		c := newConn(&nilCloser{buf})
		w := newWriter(c, test.recType, test.reqID)
		if _, err := w.Write(test.content); err != nil {
			t.Errorf("%s: error writing record: %v", test.desc, err)
			continue
		}
		if err := w.Close(); err != nil {
			t.Errorf("%s: error closing stream: %v", test.desc, err)
			continue
		}
		if !bytes.Equal(buf.Bytes(), test.raw) {
			t.Errorf("%s: wrote wrong content", test.desc)
		}
	}
}

type writeOnlyConn struct {
	buf []byte
}

func (c *writeOnlyConn) Write(p []byte) (int, error) {
	c.buf = append(c.buf, p...)
	return len(p), nil
}

func (c *writeOnlyConn) Read(_ []byte) (int, error) {
	return 0, errors.New("conn is write-only")
}

func (c *writeOnlyConn) Close() error {
	return nil
}

func TestGetValues(t *testing.T) {
	var rec record
	rec.h.Type = typeGetValues

	wc := new(writeOnlyConn)
	c := newChild(wc, nil)
	err := c.handleRecord(&rec)
	if err != nil {
		t.Fatalf("handleRecord: %v", err)
	}

	const want = "\x01\n\x00\x00\x00\x12\x06\x00" +
		"\x0f\x01FCGI_MPXS_CONNS1" +
		"\x00\x00\x00\x00\x00\x00\x01\n\x00\x00\x00\x00\x00\x00"
	if got := string(wc.buf); got != want {
		t.Errorf(" got: %q\nwant: %q\n", got, want)
	}
}

func nameValuePair11(nameData, valueData string) []byte {
	return bytes.Join(
		[][]byte{
			{byte(len(nameData)), byte(len(valueData))},
			[]byte(nameData),
			[]byte(valueData),
		},
		nil,
	)
}

func makeRecord(
	recordType recType,
	contentData []byte,
) []byte {
	requestIDB1 := byte(requestID >> 8)
	requestIDB0 := byte(requestID)

	contentLength := len(contentData)
	contentLengthB1 := byte(contentLength >> 8)
	contentLengthB0 := byte(contentLength)
	return bytes.Join([][]byte{
		{1, byte(recordType), requestIDB1, requestIDB0, contentLengthB1,
			contentLengthB0, 0, 0},
		contentData,
	},
		nil)
}

// a series of FastCGI records that start a request and begin sending the
// request body
var streamBeginTypeStdin = bytes.Join([][]byte{
	// set up request 1
	makeRecord(typeBeginRequest, []byte{0, byte(roleResponder), 0, 0, 0, 0, 0, 0}),
	// add required parameters to request 1
	makeRecord(typeParams, nameValuePair11("REQUEST_METHOD", "GET")),
	makeRecord(typeParams, nameValuePair11("SERVER_PROTOCOL", "HTTP/1.1")),
	makeRecord(typeParams, nil),
	// begin sending body of request 1
	makeRecord(typeStdin, []byte("0123456789abcdef")),
},
	nil)

var cleanUpTests = []struct {
	input []byte
	err   error
}{
	// confirm that child.handleRecord closes req.pw after aborting req
	{
		bytes.Join([][]byte{
			streamBeginTypeStdin,
			makeRecord(typeAbortRequest, nil),
		},
			nil),
		ErrRequestAborted,
	},
	// confirm that child.serve closes all pipes after error reading record
	{
		bytes.Join([][]byte{
			streamBeginTypeStdin,
			nil,
		},
			nil),
		ErrConnClosed,
	},
}

type nopWriteCloser struct {
	io.ReadWriter
}

func (nopWriteCloser) Close() error {
	return nil
}

// Test that child.serve closes the bodies of aborted requests and closes the
// bodies of all requests before returning. Causes deadlock if either condition
// isn't met. See issue 6934.
func TestChildServeCleansUp(t *testing.T) {
	for _, tt := range cleanUpTests {
		input := make([]byte, len(tt.input))
		copy(input, tt.input)
		rc := nopWriteCloser{bytes.NewBuffer(input)}
		done := make(chan bool)
		c := newChild(rc, http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			// block on reading body of request
			_, err := io.Copy(ioutil.Discard, r.Body)
			if err != tt.err {
				t.Errorf("Expected %#v, got %#v", tt.err, err)
			}
			// not reached if body of request isn't closed
			done <- true
		}))
		go c.serve()
		// wait for body of request to be closed or all goroutines to block
		<-done
	}
}

type rwNopCloser struct {
	io.Reader
	io.Writer
}

func (rwNopCloser) Close() error {
	return nil
}

// Verifies it doesn't crash. 	Issue 11824.
func TestMalformedParams(_ *testing.T) {
	input := []byte{
		// beginRequest, requestId=1, contentLength=8, role=1, keepConn=1
		1, 1, 0, 1, 0, 8, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0,
		// params, requestId=1, contentLength=10, k1Len=50, v1Len=50 (malformed, wrong length)
		1, 4, 0, 1, 0, 10, 0, 0, 50, 50, 3, 4, 5, 6, 7, 8, 9, 10,
		// end of params
		1, 4, 0, 1, 0, 0, 0, 0,
	}
	rw := rwNopCloser{bytes.NewReader(input), ioutil.Discard}
	c := newChild(rw, http.DefaultServeMux)
	c.serve()
}
