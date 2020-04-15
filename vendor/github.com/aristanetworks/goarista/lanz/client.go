// Copyright (c) 2016 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

// Package lanz implements a LANZ client that will listen to notofications from LANZ streaming
// server and will decode them and send them as a protobuf over a channel to a receiver.
package lanz

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"time"

	pb "github.com/aristanetworks/goarista/lanz/proto"

	"github.com/aristanetworks/glog"
	"github.com/golang/protobuf/proto"
)

const (
	defaultConnectTimeout = 10 * time.Second
	defaultConnectBackoff = 30 * time.Second
)

// Client is the LANZ client interface.
type Client interface {
	// Run is the main loop of the client.
	// It connects to the LANZ server and reads the notifications, decodes them
	// and sends them to the channel.
	// In case of disconnect, it will reconnect automatically.
	Run(ch chan<- *pb.LanzRecord)
	// Stops the client.
	Stop()
}

// ConnectReadCloser extends the io.ReadCloser interface with a Connect method.
type ConnectReadCloser interface {
	io.ReadCloser
	// Connect connects to the address, returning an error if it fails.
	Connect() error
}

type client struct {
	addr     string
	stopping bool
	timeout  time.Duration
	backoff  time.Duration
	conn     ConnectReadCloser
}

// New creates a new client with default TCP connection to the LANZ server.
func New(opts ...Option) Client {
	c := &client{
		stopping: false,
		timeout:  defaultConnectTimeout,
		backoff:  defaultConnectBackoff,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.conn == nil {
		if c.addr == "" {
			panic("Neither address, nor connector specified")
		}
		c.conn = &netConnector{
			addr:    c.addr,
			timeout: c.timeout,
			backoff: c.backoff,
		}
	}

	return c
}

func (c *client) Run(ch chan<- *pb.LanzRecord) {
	for !c.stopping {
		if err := c.conn.Connect(); err != nil && !c.stopping {
			glog.V(1).Infof("Can't connect to LANZ server: %v", err)
			time.Sleep(c.backoff)
			continue
		}
		glog.V(1).Infof("Connected successfully to LANZ server: %v", c.addr)
		if err := c.read(bufio.NewReader(c.conn), ch); err != nil && !c.stopping {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				glog.Errorf("Error receiving LANZ events: %v", err)
			}
			c.conn.Close()
			time.Sleep(c.backoff)
		}
	}

	close(ch)
}

func (c *client) read(r *bufio.Reader, ch chan<- *pb.LanzRecord) error {
	for !c.stopping {
		len, err := binary.ReadUvarint(r)
		if err != nil {
			return err
		}

		buf := make([]byte, len)
		if _, err = io.ReadFull(r, buf); err != nil {
			return err
		}

		rec := &pb.LanzRecord{}
		if err = proto.Unmarshal(buf, rec); err != nil {
			return err
		}

		ch <- rec
	}

	return nil
}

func (c *client) Stop() {
	if c.stopping {
		return
	}

	c.stopping = true
	c.conn.Close()
}

type netConnector struct {
	net.Conn
	addr    string
	timeout time.Duration
	backoff time.Duration
}

func (c *netConnector) Connect() (err error) {
	c.Conn, err = net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
	}
	return
}
