package disque

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Disque struct {
	Servers []string

	c   net.Conn
	buf []byte
}

var sampleConfig = `
  ## An array of URI to gather stats about. Specify an ip or hostname
  ## with optional port and password.
  ## ie disque://localhost, disque://10.10.3.33:18832, 10.0.0.1:10000, etc.
  ## If no servers are specified, then localhost is used as the host.
  servers = ["localhost"]
`

var defaultTimeout = 5 * time.Second

func (r *Disque) SampleConfig() string {
	return sampleConfig
}

func (r *Disque) Description() string {
	return "Read metrics from one or many disque servers"
}

var Tracking = map[string]string{
	"uptime_in_seconds":          "uptime",
	"connected_clients":          "clients",
	"blocked_clients":            "blocked_clients",
	"used_memory":                "used_memory",
	"used_memory_rss":            "used_memory_rss",
	"used_memory_peak":           "used_memory_peak",
	"total_connections_received": "total_connections_received",
	"total_commands_processed":   "total_commands_processed",
	"instantaneous_ops_per_sec":  "instantaneous_ops_per_sec",
	"latest_fork_usec":           "latest_fork_usec",
	"mem_fragmentation_ratio":    "mem_fragmentation_ratio",
	"used_cpu_sys":               "used_cpu_sys",
	"used_cpu_user":              "used_cpu_user",
	"used_cpu_sys_children":      "used_cpu_sys_children",
	"used_cpu_user_children":     "used_cpu_user_children",
	"registered_jobs":            "registered_jobs",
	"registered_queues":          "registered_queues",
}

var ErrProtocolError = errors.New("disque protocol error")

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *Disque) Gather(acc telegraf.Accumulator) error {
	if len(g.Servers) == 0 {
		url := &url.URL{
			Host: ":7711",
		}
		g.gatherServer(url, acc)
		return nil
	}

	var wg sync.WaitGroup

	for _, serv := range g.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse to address '%s': %s", serv, err))
			continue
		} else if u.Scheme == "" {
			// fallback to simple string based address (i.e. "10.0.0.1:10000")
			u.Scheme = "tcp"
			u.Host = serv
			u.Path = ""
		}
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			acc.AddError(g.gatherServer(u, acc))
		}(serv)
	}

	wg.Wait()

	return nil
}

const defaultPort = "7711"

func (g *Disque) gatherServer(addr *url.URL, acc telegraf.Accumulator) error {
	if g.c == nil {

		_, _, err := net.SplitHostPort(addr.Host)
		if err != nil {
			addr.Host = addr.Host + ":" + defaultPort
		}

		c, err := net.DialTimeout("tcp", addr.Host, defaultTimeout)
		if err != nil {
			return fmt.Errorf("Unable to connect to disque server '%s': %s", addr.Host, err)
		}

		if addr.User != nil {
			pwd, set := addr.User.Password()
			if set && pwd != "" {
				c.Write([]byte(fmt.Sprintf("AUTH %s\r\n", pwd)))

				r := bufio.NewReader(c)

				line, err := r.ReadString('\n')
				if err != nil {
					return err
				}
				if line[0] != '+' {
					return fmt.Errorf("%s", strings.TrimSpace(line)[1:])
				}
			}
		}

		g.c = c
	}

	// Extend connection
	g.c.SetDeadline(time.Now().Add(defaultTimeout))

	g.c.Write([]byte("info\r\n"))

	r := bufio.NewReader(g.c)

	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	if line[0] != '$' {
		return fmt.Errorf("bad line start: %s", ErrProtocolError)
	}

	line = strings.TrimSpace(line)

	szStr := line[1:]

	sz, err := strconv.Atoi(szStr)
	if err != nil {
		return fmt.Errorf("bad size string <<%s>>: %s", szStr, ErrProtocolError)
	}

	var read int

	fields := make(map[string]interface{})
	tags := map[string]string{"disque_host": addr.String()}
	for read < sz {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}

		read += len(line)

		if len(line) == 1 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, ":", 2)

		name := string(parts[0])

		metric, ok := Tracking[name]
		if !ok {
			continue
		}

		val := strings.TrimSpace(parts[1])

		ival, err := strconv.ParseUint(val, 10, 64)
		if err == nil {
			fields[metric] = ival
			continue
		}

		fval, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}

		fields[metric] = fval
	}
	acc.AddFields("disque", fields, tags)
	return nil
}

func init() {
	inputs.Add("disque", func() telegraf.Input {
		return &Disque{}
	})
}
