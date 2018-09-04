package network // import "collectd.org/network"

import (
	"context"
	"net"

	"collectd.org/api"
)

// ClientOptions holds configuration options for Client.
type ClientOptions struct {
	// SecurityLevel determines whether data is signed, encrypted or sent
	// in plain text.
	SecurityLevel SecurityLevel
	// Username and password for the "Sign" and "Encrypt" security levels.
	Username, Password string
	// Size of the send buffer. When zero, DefaultBufferSize is used.
	BufferSize int
}

// Client is a connection to a collectd server. It implements the
// api.Writer interface.
type Client struct {
	udp    net.Conn
	buffer *Buffer
	opts   ClientOptions
}

// Dial connects to the collectd server at address. "address" must be a network
// address accepted by net.Dial().
func Dial(address string, opts ClientOptions) (*Client, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	b := NewBuffer(opts.BufferSize)
	if opts.SecurityLevel == Sign {
		b.Sign(opts.Username, opts.Password)
	} else if opts.SecurityLevel == Encrypt {
		b.Encrypt(opts.Username, opts.Password)
	}

	return &Client{
		udp:    c,
		buffer: b,
		opts:   opts,
	}, nil
}

// Write adds a ValueList to the internal buffer. Data is only written to
// the network when the buffer is full.
func (c *Client) Write(ctx context.Context, vl *api.ValueList) error {
	if err := c.buffer.Write(ctx, vl); err != ErrNotEnoughSpace {
		return err
	}

	if err := c.Flush(); err != nil {
		return err
	}

	return c.buffer.Write(ctx, vl)
}

// Flush writes the contents of the underlying buffer to the network
// immediately.
func (c *Client) Flush() error {
	_, err := c.buffer.WriteTo(c.udp)
	return err
}

// Close writes remaining data to the network and closes the socket. You must
// not use "c" after this call.
func (c *Client) Close() error {
	if err := c.Flush(); err != nil {
		return err
	}

	if err := c.udp.Close(); err != nil {
		return err
	}

	c.buffer = nil
	return nil
}
