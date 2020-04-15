// Package apcupsd provides a client for the apcupsd Network Information
// Server (NIS).
package apcupsd

import (
	"context"
	"io"
	"net"
)

// Client is a client for the apcupsd Network Information Server (NIS).
type Client struct {
	rwc io.ReadWriteCloser
}

// Dial dials a connection to an NIS using the address on the named
// network, and creates a Client with the connection.
//
// Typically, network will be one of: "tcp", "tcp4", or "tcp6".
func Dial(network string, addr string) (*Client, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	return New(c), nil
}

// DialContext takes a context and dials a connection to an NIS using the address on the named
// network, and creates a Client with the connection.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
//
// Typically, network will be one of: "tcp", "tcp4", or "tcp6".
func DialContext(ctx context.Context, network, address string) (*Client, error) {
	var d net.Dialer
	c, err := d.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return New(c), nil
}

// New wraps an existing io.ReadWriteCloser to create a Client for communication
// with an NIS.  Client's Close method will close the io.ReadWriteCloser
// when called.
func New(rwc io.ReadWriteCloser) *Client {
	return &Client{
		rwc: newNISReadWriteCloser(rwc),
	}
}

// Close closes the connection to an NIS.
func (c *Client) Close() error {
	return c.rwc.Close()
}

const (
	// maxString is the maximum string length for a NIS key/value pair.
	// Value copied from apcupsd source code, v3.14.14.
	maxString = 256
)

// Status retrieves the current UPS status from the NIS.
func (c *Client) Status() (*Status, error) {
	// Issue a status command to NIS
	_, err := c.rwc.Write([]byte("status"))
	if err != nil {
		return nil, err
	}

	b := make([]byte, maxString)
	s := new(Status)

	// NIS server sends text lines containing key/value pairs, so must keep
	// iterating until EOF to parse them all
	for {
		n, err := c.rwc.Read(b)
		if err == io.EOF {
			// Received key/value pair with length 0
			break
		}
		if err != nil {
			return nil, err
		}

		// Parse key/value pair into appropriate struct field
		if err := s.parseKV(string(b[:n])); err != nil {
			return nil, err
		}
	}

	return s, nil
}
