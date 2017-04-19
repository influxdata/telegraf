package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
)

const (
	// UDPPayloadSize is a reasonable default payload size for UDP packets that
	// could be travelling over the internet.
	UDPPayloadSize = 512
)

// UDPConfig is the config data needed to create a UDP Client
type UDPConfig struct {
	// URL should be of the form "udp://host:port"
	// or "udp://[ipv6-host%zone]:port".
	URL string

	// PayloadSize is the maximum size of a UDP client message, optional
	// Tune this based on your network. Defaults to UDPPayloadSize.
	PayloadSize int
}

func NewUDP(config UDPConfig) (Client, error) {
	p, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing UDP url [%s]: %s", config.URL, err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", p.Host)
	if err != nil {
		return nil, fmt.Errorf("Error resolving UDP Address [%s]: %s", p.Host, err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("Error dialing UDP address [%s]: %s",
			udpAddr.String(), err)
	}

	size := config.PayloadSize
	if size == 0 {
		size = UDPPayloadSize
	}
	buf := make([]byte, size)
	return &udpClient{conn: conn, buffer: buf}, nil
}

type udpClient struct {
	conn   *net.UDPConn
	buffer []byte
}

func (c *udpClient) Query(command string) error {
	return nil
}

func (c *udpClient) Write(b []byte) (int, error) {
	return c.WriteStream(bytes.NewReader(b), -1)
}

// write params are ignored by the UDP client
func (c *udpClient) WriteWithParams(b []byte, wp WriteParams) (int, error) {
	return c.WriteStream(bytes.NewReader(b), -1)
}

// contentLength is ignored by the UDP client.
func (c *udpClient) WriteStream(r io.Reader, contentLength int) (int, error) {
	var totaln int
	for {
		nR, err := r.Read(c.buffer)
		if nR == 0 {
			break
		}
		if err != io.EOF && err != nil {
			return totaln, err
		}
		nW, err := c.conn.Write(c.buffer[0:nR])
		totaln += nW
		if err != nil {
			return totaln, err
		}
	}
	return totaln, nil
}

// contentLength is ignored by the UDP client.
// write params are ignored by the UDP client
func (c *udpClient) WriteStreamWithParams(r io.Reader, contentLength int, wp WriteParams) (int, error) {
	return c.WriteStream(r, -1)
}

func (c *udpClient) Close() error {
	return c.conn.Close()
}
