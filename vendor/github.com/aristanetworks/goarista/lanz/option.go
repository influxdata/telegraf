// Copyright (c) 2016 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package lanz

import "time"

// Option is a LANZ client factory option.
type Option func(c *client)

// WithAddr specifies the address of the LANZ server.
// If WithConnector is not used, then WithAddr must be used.
// When WithConnector is used, WithAddr can be used to pass the address string for displaying.
func WithAddr(addr string) Option {
	return func(c *client) {
		c.addr = addr
	}
}

// WithConnector specifies a connector used to communicate with LANZ server.
func WithConnector(conn ConnectReadCloser) Option {
	return func(c *client) {
		c.conn = conn
	}
}

// WithTimeout specifies the timeout for connecting to LANZ server.
// It only takes effect for default connector.
func WithTimeout(d time.Duration) Option {
	return func(c *client) {
		c.timeout = d
	}
}

// WithBackoff specifies the backoff time after failed connection to LANZ server.
// It only takes effect for default connector.
func WithBackoff(d time.Duration) Option {
	return func(c *client) {
		c.backoff = d
	}
}
