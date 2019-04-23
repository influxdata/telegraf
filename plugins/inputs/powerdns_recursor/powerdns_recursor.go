package powerdns_recursor

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type PowerdnsRecursor struct {
	UnixSockets []string

	SocketDir  string `toml:"socket_dir"`
	SocketMode uint32 `toml:"socket_mode"`
}

var defaultTimeout = 5 * time.Second

var sampleConfig = `
  ## An array of sockets to gather stats about.
  ## Specify a path to unix socket.
  unix_sockets = ["/var/run/pdns_recursor.controlsocket"]

  ## Socket for Receive
  #socket_dir = "/var/run/"
  ## Socket permissions
  #socket_mode = "0666"
`

func (p *PowerdnsRecursor) SampleConfig() string {
	return sampleConfig
}

func (p *PowerdnsRecursor) Description() string {
	return "Read metrics from one or many PowerDNS Recursor servers"
}

func (p *PowerdnsRecursor) Gather(acc telegraf.Accumulator) error {
	if len(p.UnixSockets) == 0 {
		return p.gatherServer("/var/run/pdns_recursor.controlsocket", acc)
	}

	for _, serverSocket := range p.UnixSockets {
		if err := p.gatherServer(serverSocket, acc); err != nil {
			acc.AddError(err)
		}
	}

	return nil
}

func (p *PowerdnsRecursor) gatherServer(address string, acc telegraf.Accumulator) error {
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
	perm := uint32(0666)
	if p.SocketMode > 0 {
		perm = p.SocketMode
	}
	if err := os.Chmod(recvSocket, os.FileMode(perm)); err != nil {
		return err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(defaultTimeout))

	// Read and write buffer
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Send command
	if _, err := fmt.Fprint(rw, "get-all\n"); err != nil {
		return nil
	}
	if err := rw.Flush(); err != nil {
		return err
	}

	// Read data
	buf := make([]byte, 16384)
	n, err := rw.Read(buf)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("no data received")
	}

	metrics := string(buf)

	// Process data
	fields := parseResponse(metrics)

	// Add server socket as a tag
	tags := map[string]string{"server": address}

	acc.AddFields("powerdns_recursor", fields, tags)

	conn.Close()

	return nil
}

func parseResponse(metrics string) map[string]interface{} {
	values := make(map[string]interface{})

	s := strings.Split(metrics, "\n")

	for _, metric := range s[:len(s)-1] {
		m := strings.Split(metric, "\t")
		if len(m) < 2 {
			continue
		}

		i, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			log.Printf("E! [inputs.powerdns_recursor] Error parsing integer for metric [%s] %v",
				metric, err)
			continue
		}
		values[m[0]] = i
	}

	return values
}

func init() {
	inputs.Add("powerdns_recursor", func() telegraf.Input {
		return &PowerdnsRecursor{}
	})
}
