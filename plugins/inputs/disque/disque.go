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

	c net.Conn
}

var defaultTimeout = 5 * time.Second

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
func (d *Disque) Gather(acc telegraf.Accumulator) error {
	if len(d.Servers) == 0 {
		address := &url.URL{
			Host: ":7711",
		}
		return d.gatherServer(address, acc)
	}

	var wg sync.WaitGroup

	for _, serv := range d.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to parse to address '%s': %s", serv, err))
			continue
		} else if u.Scheme == "" {
			// fallback to simple string based address (i.e. "10.0.0.1:10000")
			u.Scheme = "tcp"
			u.Host = serv
			u.Path = ""
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			acc.AddError(d.gatherServer(u, acc))
		}()
	}

	wg.Wait()

	return nil
}

const defaultPort = "7711"

func (d *Disque) gatherServer(addr *url.URL, acc telegraf.Accumulator) error {
	if d.c == nil {
		_, _, err := net.SplitHostPort(addr.Host)
		if err != nil {
			addr.Host = addr.Host + ":" + defaultPort
		}

		c, err := net.DialTimeout("tcp", addr.Host, defaultTimeout)
		if err != nil {
			return fmt.Errorf("unable to connect to disque server '%s': %s", addr.Host, err)
		}

		if addr.User != nil {
			pwd, set := addr.User.Password()
			if set && pwd != "" {
				if _, err := c.Write([]byte(fmt.Sprintf("AUTH %s\r\n", pwd))); err != nil {
					return err
				}

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

		d.c = c
	}

	// Extend connection
	if err := d.c.SetDeadline(time.Now().Add(defaultTimeout)); err != nil {
		return err
	}

	if _, err := d.c.Write([]byte("info\r\n")); err != nil {
		return err
	}

	r := bufio.NewReader(d.c)

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

		name := parts[0]

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
