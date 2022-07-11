package powerdns_recursor

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

// V2 (4.5.0 - 4.5.9) Protocol:
// Unix datagram socket
// Synchronous request / response, individual datagrams
// Datagram 1 => status: uint32
// Datagram 2 => data: byte[] (max 16_384 bytes)
func (p *PowerdnsRecursor) gatherFromV2Server(address string, acc telegraf.Accumulator) error {
	randomNumber := rand.Int63()
	recvSocket := filepath.Join("/", "var", "run", fmt.Sprintf("pdns_recursor_telegraf%d", randomNumber))
	if p.SocketDir != "" {
		recvSocket = filepath.Join(p.SocketDir, fmt.Sprintf("pdns_recursor_telegraf%d", randomNumber))
	}

	laddr, err := net.ResolveUnixAddr("unixgram", recvSocket)

	if err != nil {
		return err
	}

	defer os.Remove(recvSocket)

	raddr, err := net.ResolveUnixAddr("unixgram", address)

	if err != nil {
		return err
	}
	conn, err := net.DialUnix("unixgram", laddr, raddr)
	if err != nil {
		return err
	}
	if err := os.Chmod(recvSocket, os.FileMode(p.mode)); err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(defaultTimeout)); err != nil {
		return err
	}

	// First send a 0 status code.
	_, err = conn.Write([]byte{0, 0, 0, 0})

	if err != nil {
		return err
	}

	// Then send the get-all command.
	command := "get-all"

	_, err = conn.Write([]byte(command))

	if err != nil {
		return err
	}

	// Read the response status code.
	status := make([]byte, 4)
	n, err := conn.Read(status)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("no status code received")
	}

	// Read the response data.
	buf := make([]byte, 16_384)
	n, err = conn.Read(buf)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("no data received")
	}

	metrics := string(buf)

	// Process data
	fields := parseResponse(metrics)

	// Add server socket as a tag
	tags := map[string]string{"server": address}

	acc.AddFields("powerdns_recursor", fields, tags)

	return conn.Close()
}
