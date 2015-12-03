package memcached

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/influxdb/telegraf/plugins"
)

// Memcached is a memcached plugin
type Memcached struct {
	Servers     []string
	UnixSockets []string
}

var sampleConfig = `
  # An array of address to gather stats about. Specify an ip on hostname
  # with optional port. ie localhost, 10.0.0.1:11211, etc.
  #
  # If no servers are specified, then localhost is used as the host.
  servers = ["localhost:11211"]
  # unix_sockets = ["/var/run/memcached.sock"]
`

var defaultTimeout = 5 * time.Second

// The list of metrics tha should be calculated
var sendAsIs = []string{
	"get_hits",
	"get_misses",
	"evictions",
	"limit_maxbytes",
	"bytes",
	"uptime",
	"curr_items",
	"total_items",
	"curr_connections",
	"total_connections",
	"connection_structures",
	"cmd_get",
	"cmd_set",
	"delete_hits",
	"delete_misses",
	"incr_hits",
	"incr_misses",
	"decr_hits",
	"decr_misses",
	"cas_hits",
	"cas_misses",
	"evictions",
	"bytes_read",
	"bytes_written",
	"threads",
	"conn_yields",
}

// SampleConfig returns sample configuration message
func (m *Memcached) SampleConfig() string {
	return sampleConfig
}

// Description returns description of Memcached plugin
func (m *Memcached) Description() string {
	return "Read metrics from one or many memcached servers"
}

// Gather reads stats from all configured servers accumulates stats
func (m *Memcached) Gather(acc plugins.Accumulator) error {
	if len(m.Servers) == 0 && len(m.UnixSockets) == 0 {
		return m.gatherServer(":11211", false, acc)
	}

	for _, serverAddress := range m.Servers {
		if err := m.gatherServer(serverAddress, false, acc); err != nil {
			return err
		}
	}

	for _, unixAddress := range m.UnixSockets {
		if err := m.gatherServer(unixAddress, true, acc); err != nil {
			return err
		}
	}

	return nil
}

func (m *Memcached) gatherServer(
	address string,
	unix bool,
	acc plugins.Accumulator,
) error {
	var conn net.Conn
	if unix {
		conn, err := net.DialTimeout("unix", address, defaultTimeout)
		if err != nil {
			return err
		}
		defer conn.Close()
	} else {
		_, _, err := net.SplitHostPort(address)
		if err != nil {
			address = address + ":11211"
		}

		conn, err = net.DialTimeout("tcp", address, defaultTimeout)
		if err != nil {
			return err
		}
		defer conn.Close()
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

	// Read response
	values := make(map[string]string)

	for {
		// Read line
		line, _, errRead := rw.Reader.ReadLine()
		if errRead != nil {
			return errRead
		}
		// Done
		if bytes.Equal(line, []byte("END")) {
			break
		}
		// Read values
		s := bytes.SplitN(line, []byte(" "), 3)
		if len(s) != 3 || !bytes.Equal(s[0], []byte("STAT")) {
			return fmt.Errorf("unexpected line in stats response: %q", line)
		}

		// Save values
		values[string(s[1])] = string(s[2])
	}

	// Add server address as a tag
	tags := map[string]string{"server": address}

	// Process values
	for _, key := range sendAsIs {
		if value, ok := values[key]; ok {
			// Mostly it is the number
			if iValue, errParse := strconv.ParseInt(value, 10, 64); errParse != nil {
				acc.Add(key, value, tags)
			} else {
				acc.Add(key, iValue, tags)
			}
		}
	}
	return nil
}

func init() {
	plugins.Add("memcached", func() plugins.Plugin {
		return &Memcached{}
	})
}
