package powerdns

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Powerdns struct {
	UnixSockets []string
}

var sampleConfig = `
  ## An array of sockets to gather stats about.
  ## Specify a path to unix socket.
  unix_sockets = ["/var/run/pdns.controlsocket"]
`

var defaultTimeout = 5 * time.Second

func (p *Powerdns) SampleConfig() string {
	return sampleConfig
}

func (p *Powerdns) Description() string {
	return "Read metrics from one or many PowerDNS servers"
}

func (p *Powerdns) Gather(acc telegraf.Accumulator) error {
	if len(p.UnixSockets) == 0 {
		return p.gatherServer("/var/run/pdns.controlsocket", acc)
	}

	for _, serverSocket := range p.UnixSockets {
		if err := p.gatherServer(serverSocket, acc); err != nil {
			acc.AddError(err)
		}
	}

	return nil
}

func (p *Powerdns) gatherServer(address string, acc telegraf.Accumulator) error {
	conn, err := net.DialTimeout("unix", address, defaultTimeout)
	if err != nil {
		return err
	}

	defer conn.Close()

	conn.SetDeadline(time.Now().Add(defaultTimeout))

	// Read and write buffer
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Send command
	if _, err := fmt.Fprint(conn, "show * \n"); err != nil {
		return nil
	}
	if err := rw.Flush(); err != nil {
		return err
	}

	// Read data
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for {
		n, err := rw.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return err
			}

			break
		}
		buf = append(buf, tmp[:n]...)
	}

	metrics := string(buf)

	// Process data
	fields := parseResponse(metrics)

	// Add server socket as a tag
	tags := map[string]string{"server": address}

	acc.AddFields("powerdns", fields, tags)

	return nil
}

func parseResponse(metrics string) map[string]interface{} {
	values := make(map[string]interface{})

	s := strings.Split(metrics, ",")

	for _, metric := range s[:len(s)-1] {
		m := strings.Split(metric, "=")
		if len(m) < 2 {
			continue
		}

		i, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			log.Printf("E! powerdns: Error parsing integer for metric [%s]: %s",
				metric, err)
			continue
		}
		values[m[0]] = i
	}

	return values
}

func init() {
	inputs.Add("powerdns", func() telegraf.Input {
		return &Powerdns{}
	})
}
