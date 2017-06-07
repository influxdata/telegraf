package beanstalkd

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Beanstalkd is a Beanstalkd plugin
type Beanstalkd struct {
	Servers []string
}

var sampleConfig = `
  ## An array of address to gather stats about. Specify an ip on hostname
  ## with optional port. ie localhost, 10.0.0.1:11300, etc.
  servers = ["localhost:11300"]
`

var defaultTimeout = 5 * time.Second

// The list of metrics that should be sent
var sendMetrics = []string{
	"current-jobs-urgent",
	"current-jobs-ready",
	"current-jobs-reserved",
	"current-jobs-delayed",
	"current-jobs-buried",
	"cmd-put",
	"cmd-peek",
	"cmd-peek-ready",
	"cmd-peek-delayed",
	"cmd-peek-buried",
	"cmd-reserve",
	"cmd-reserve-with-timeout",
	"cmd-delete",
	"cmd-release",
	"cmd-use",
	"cmd-watch",
	"cmd-ignore",
	"cmd-bury",
	"cmd-kick",
	"cmd-touch",
	"cmd-stats",
	"cmd-stats-job",
	"cmd-stats-tube",
	"cmd-list-tubes",
	"cmd-list-tube-used",
	"cmd-list-tubes-watched",
	"cmd-pause-tube",
	"job-timeouts",
	"total-jobs",
	"current-tubes",
	"current-connections",
	"current-producers",
	"current-workers",
	"current-waiting",
	"total-connections",
	"uptime",
	"binlog-oldest-index",
	"binlog-current-index",
	"binlog-records-migrated",
	"binlog-records-written",
	"binlog-max-size",
}

// SampleConfig returns sample configuration message
func (m *Beanstalkd) SampleConfig() string {
	return sampleConfig
}

// Description returns description of Beanstalkd plugin
func (m *Beanstalkd) Description() string {
	return "Read metrics from one or many Beanstalkd servers"
}

// Gather reads stats from all configured servers accumulates stats
func (m *Beanstalkd) Gather(acc telegraf.Accumulator) error {
	if len(m.Servers) == 0 {
		return m.gatherServer(":11300", acc)
	}

	errChan := errchan.New(len(m.Servers))
	for _, serverAddress := range m.Servers {
		errChan.C <- m.gatherServer(serverAddress, acc)
	}

	return errChan.Error()
}

func (m *Beanstalkd) gatherServer(
	address string,
	acc telegraf.Accumulator,
) error {
	var conn net.Conn
	var err error
	_, _, err = net.SplitHostPort(address)
	if err != nil {
		address = address + ":11300"
	}

	conn, err = net.DialTimeout("tcp", address, defaultTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	if conn == nil {
		return fmt.Errorf("Failed to create net connection")
	}

	// Extend connection
	conn.SetDeadline(time.Now().Add(defaultTimeout))

	// Read and write buffer
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// Send command
	if _, err := fmt.Fprint(rw, "stats\r\n"); err != nil {
		return err
	}
	if err := rw.Flush(); err != nil {
		return err
	}

	values, err := parseResponse(rw.Reader)
	if err != nil {
		return err
	}

	// Add server address as a tag
	tags := map[string]string{"server": address}

	// Process values
	fields := make(map[string]interface{})
	for _, key := range sendMetrics {
		if value, ok := values[key]; ok {
			// Mostly it is the number
			if iValue, errParse := strconv.ParseInt(value, 10, 64); errParse == nil {
				fields[key] = iValue
			} else {
				fields[key] = value
			}
		}
	}
	acc.AddFields("beanstalkd", fields, tags)
	return nil
}

func parseResponse(r *bufio.Reader) (map[string]string, error) {
	values := make(map[string]string)

	for {
		// Read line
		line, _, errRead := r.ReadLine()
		if errRead != nil {
			return values, errRead
		}
		// Done
		if bytes.Equal(line, []byte("")) {
			break
		}

		// Read values
		s := bytes.SplitN(line, []byte(": "), 2)

		if len(s) != 2 {
			s := bytes.SplitN(line, []byte(" "), 2)
			if bytes.Equal(s[0], []byte("---")) || bytes.Equal(s[0], []byte("OK")) {
				continue
			} else {
				return values, fmt.Errorf("unexpected line in stats response: %q", line)
			}
		}

		// Save values
		values[string(s[0])] = string(s[1])
	}
	return values, nil
}

func init() {
	inputs.Add("beanstalkd", func() telegraf.Input {
		return &Beanstalkd{}
	})
}
