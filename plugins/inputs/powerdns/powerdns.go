package powerdns

import (
	"fmt"
	"io"
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

var (
	sampleConfig = `
  ## An array of sockets to gather stats about.
  ## Specify a path to unix socket.
  unix_sockets = ["/var/run/pdns.controlsocket"]
`
	defaultTimeout   = 10 * time.Second
	statsCommand     = "show *"
	respsizesCommand = "respsizes"
)

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
	// Fetch general stats.
	generalStatsResp, err := p.executeCommand(statsCommand, address)
	if err != nil {
		return err
	}
	// Fetch response size stats.
	respsizeStatsResp, err := p.executeCommand(respsizesCommand, address)
	if err != nil {
		return err
	}
	// Process general stats data.
	gsFields, err := parseGeneralStats(generalStatsResp)
	if err != nil {
		return err
	}
	// Process resp size stats data.
	rsFields, err := parseRespsizesResponse(respsizeStatsResp)
	if err != nil {
		return err
	}
	// Add server socket as a generic tag for all metrics
	tags := map[string]string{"server": address}
	// Register general stat metrics.
	acc.AddFields("powerdns", gsFields, tags)
	// Register response size stat metrics.
	acc.AddHistogram("powerdns_response_ms", rsFields, tags)
	return nil
}

func (p *Powerdns) executeCommand(command, address string) (string, error) {
	// Initiate connection to PowerDNS Unix Socket.
	conn, err := net.DialTimeout("unix", address, defaultTimeout)
	if err != nil {
		return "", err
	}
	conn.SetDeadline(time.Now().Add(defaultTimeout))
	// Close connection after executing the command.
	defer conn.Close()

	// Send command to connection.
	if _, err := fmt.Fprintln(conn, command); err != nil {
		return "", err
	}
	// Read response from connection.
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for {
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		buf = append(buf, tmp[:n]...)
	}
	return string(buf), nil
}

func parseGeneralStats(rawStats string) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	s := strings.Split(rawStats, ",")
	for _, metric := range s[:len(s)-1] {
		cols := strings.Split(metric, "=")
		if len(cols) < 2 {
			continue
		}
		i, err := strconv.ParseInt(cols[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("E! powerdns: Error parsing integer for metric [%s]: %s", metric, err)
		}
		fields[cols[0]] = i
	}
	return fields, nil
}

func parseRespsizesResponse(rawStats string) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	var upperInclusiveCount, sum int64
	// Make buckets
	for _, line := range strings.Split(rawStats, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) != 2 {
			continue
		}
		latency, err := strconv.ParseInt(cols[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("E! powerdns: Error parsing integer for respsizes metric [%s]: %s", line, err)
		}
		count, err := strconv.ParseInt(cols[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("E! powerdns: Error parsing integer for respsizes metric [%s]: %s", line, err)
		}
		upperInclusiveCount += count
		sum += count * latency
		latencyStr := strconv.FormatInt(latency, 10)
		fields[latencyStr] = upperInclusiveCount
	}
	// Add count and sum fields
	fields["count"] = upperInclusiveCount
	fields["sum"] = sum
	return fields, nil
}

func init() {
	inputs.Add("powerdns", func() telegraf.Input {
		return &Powerdns{}
	})
}
