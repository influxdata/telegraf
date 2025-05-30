// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phpfpm

// This file implements FastCGI from the perspective of a child process.

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cgi"
	"strings"
	"sync"
	"time"
)

// request holds the state for an in-progress request. As soon as it's complete,
// it's converted to an http.Request.
type request struct {
	pw        *io.PipeWriter
	reqID     uint16
	params    map[string]string
	buf       [1024]byte
	rawParams []byte
	keepConn  bool
}

func newRequest(reqID uint16, flags uint8) *request {
	r := &request{
		reqID:    reqID,
		params:   make(map[string]string),
		keepConn: flags&flagKeepConn != 0,
	}
	r.rawParams = r.buf[:0]
	return r
}

// parseParams reads an encoded []byte into Params.
func (r *request) parseParams() {
	text := r.rawParams
	r.rawParams = nil
	for len(text) > 0 {
		keyLen, n := readSize(text)
		if n == 0 {
			return
		}
		text = text[n:]
		valLen, n := readSize(text)
		if n == 0 {
			return
		}
		text = text[n:]
		if int(keyLen)+int(valLen) > len(text) {
			return
		}
		key := readString(text, keyLen)
		text = text[keyLen:]
		val := readString(text, valLen)
		text = text[valLen:]
		r.params[key] = val
	}
}

// response implements http.ResponseWriter.
type response struct {
	req         *request
	header      http.Header
	w           *bufWriter
	wroteHeader bool
}

func newResponse(c *child, req *request) *response {
	return &response{
		req:    req,
		header: http.Header{},
		w:      newWriter(c.conn, typeStdout, req.reqID),
	}
}

// Header returns the HTTP headers for the response.
func (r *response) Header() http.Header {
	return r.header
}

func (r *response) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.w.Write(data)
}

// WriteHeader sends an HTTP response header with the provided status code.
func (r *response) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	if code == http.StatusNotModified {
		// Must not have body.
		r.header.Del("Content-Type")
		r.header.Del("Content-Length")
		r.header.Del("Transfer-Encoding")
	} else if r.header.Get("Content-Type") == "" {
		r.header.Set("Content-Type", "text/html; charset=utf-8")
	}

	if r.header.Get("Date") == "" {
		r.header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	fmt.Fprintf(r.w, "Status: %d %s\r\n", code, http.StatusText(code))
	//nolint:errcheck // unable to propagate
	r.header.Write(r.w)
	//nolint:errcheck // unable to propagate
	r.w.WriteString("\r\n")
}

// Flush sends any buffered data to the client.
func (r *response) Flush() {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	_ = r.w.Flush()
}

// Close closes the connection or resource associated with the response.
// It ensures proper cleanup of resources.
func (r *response) Close() error {
	r.Flush()
	return r.w.Close()
}

type child struct {
	conn    *conn
	handler http.Handler

	mu       sync.Mutex          // protects requests:
	requests map[uint16]*request // keyed by request ID
}

func newChild(rwc io.ReadWriteCloser, handler http.Handler) *child {
	return &child{
		conn:     newConn(rwc),
		handler:  handler,
		requests: make(map[uint16]*request),
	}
}

func (c *child) serve() {
	defer c.conn.Close()
	defer c.cleanUp()
	var rec record
	for {
		if err := rec.read(c.conn.rwc); err != nil {
			return
		}
		if err := c.handleRecord(&rec); err != nil {
			return
		}
	}
}

var errCloseConn = errors.New("fcgi: connection should be closed")

var emptyBody = io.NopCloser(strings.NewReader(""))

// errRequestAborted is returned by Read when a handler attempts to read the
// body of a request that has been aborted by the web server.
var errRequestAborted = errors.New("fcgi: request aborted by web server")

// errConnClosed is returned by Read when a handler attempts to read the body of
// a request after the connection to the web server has been closed.
var errConnClosed = errors.New("fcgi: connection to web server closed")

func (c *child) handleRecord(rec *record) error {
	c.mu.Lock()
	req, ok := c.requests[rec.h.ID]
	c.mu.Unlock()
	if !ok && rec.h.Type != typeBeginRequest && rec.h.Type != typeGetValues {
		// The spec says to ignore unknown request IDs.
		return nil
	}

	switch rec.h.Type {
	case typeBeginRequest:
		if req != nil {
			// The server is trying to begin a request with the same ID
			// as an in-progress request. This is an error.
			return errors.New("fcgi: received ID that is already in-flight")
		}

		var br beginRequest
		if err := br.read(rec.content()); err != nil {
			return err
		}
		if br.role != roleResponder {
			return c.conn.writeEndRequest(rec.h.ID, 0, statusUnknownRole)
		}
		req = newRequest(rec.h.ID, br.flags)
		c.mu.Lock()
		c.requests[rec.h.ID] = req
		c.mu.Unlock()
		return nil
	case typeParams:
		// NOTE(eds): Technically a key-value pair can straddle the boundary
		// between two packets. We buffer until we've received all parameters.
		if len(rec.content()) > 0 {
			req.rawParams = append(req.rawParams, rec.content()...)
			return nil
		}
		req.parseParams()
		return nil
	case typeStdin:
		content := rec.content()
		if req.pw == nil {
			var body io.ReadCloser
			if len(content) > 0 {
				// body could be an io.LimitReader, but it shouldn't matter
				// as long as both sides are behaving.
				body, req.pw = io.Pipe()
			} else {
				body = emptyBody
			}
			go c.serveRequest(req, body)
		}
		if len(content) > 0 {
			// TODO(eds): This blocks until the handler reads from the pipe.
			// If the handler takes a long time, it might be a problem.
			if _, err := req.pw.Write(content); err != nil {
				return err
			}
		} else if req.pw != nil {
			if err := req.pw.Close(); err != nil {
				return err
			}
		}
		return nil
	case typeGetValues:
		values := map[string]string{"FCGI_MPXS_CONNS": "1"}
		return c.conn.writePairs(typeGetValuesResult, 0, values)
	case typeData:
		// If the filter role is implemented, read the data stream here.
		return nil
	case typeAbortRequest:
		c.mu.Lock()
		delete(c.requests, rec.h.ID)
		c.mu.Unlock()
		if err := c.conn.writeEndRequest(rec.h.ID, 0, statusRequestComplete); err != nil {
			return err
		}
		if req.pw != nil {
			req.pw.CloseWithError(errRequestAborted)
		}
		if !req.keepConn {
			// connection will close upon return
			return errCloseConn
		}
		return nil
	default:
		b := make([]byte, 8)
		b[0] = byte(rec.h.Type)
		return c.conn.writeRecord(typeUnknownType, 0, b)
	}
}

func (c *child) serveRequest(req *request, body io.ReadCloser) {
	r := newResponse(c, req)
	httpReq, err := cgi.RequestFromMap(req.params)
	if err != nil {
		// there was an error reading the request
		r.WriteHeader(http.StatusInternalServerError)
		if err := c.conn.writeRecord(typeStderr, req.reqID, []byte(err.Error())); err != nil {
			return
		}
	} else {
		httpReq.Body = body
		c.handler.ServeHTTP(r, httpReq)
	}
	r.Close()
	c.mu.Lock()
	delete(c.requests, req.reqID)
	c.mu.Unlock()
	if err := c.conn.writeEndRequest(req.reqID, 0, statusRequestComplete); err != nil {
		return
	}

	// Consume the entire body, so the host isn't still writing to
	// us when we close the socket below in the !keepConn case,
	// otherwise we'd send a RST. (golang.org/issue/4183)
	// TODO(bradfitz): also bound this copy in time. Or send
	// some sort of abort request to the host, so the host
	// can properly cut off the client sending all the data.
	// For now just bound it a little and
	io.CopyN(io.Discard, body, 100<<20) //nolint:errcheck // ignore the returned error as we cannot do anything about it anyway
	body.Close()

	if !req.keepConn {
		c.conn.Close()
	}
}

func (c *child) cleanUp() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, req := range c.requests {
		if req.pw != nil {
			// race with call to Close in c.serveRequest doesn't matter because
			// Pipe(Reader|Writer).Close are idempotent
			req.pw.CloseWithError(errConnClosed)
		}
	}
}
