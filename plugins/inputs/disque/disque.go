//go:generate ../../../tools/readme_config_includer/generator
package disque

import (
	"bufio"
	_ "embed"
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

//go:embed sample.conf
var sampleConfig string

var (
	defaultTimeout = 5 * time.Second
	tracking       = map[string]string{
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
	errProtocol = errors.New("disque protocol error")
)

const (
	defaultPort = "7711"
)

type Disque struct {
	Servers []string `toml:"servers"`

	c net.Conn
}

func (*Disque) SampleConfig() string {
	return sampleConfig
}

func (d *Disque) Gather(acc telegraf.Accumulator) error {
	if len(d.Servers) == 0 {
		address := &url.URL{
			Host: ":" + defaultPort,
		}
		return d.gatherServer(address, acc)
	}

	var wg sync.WaitGroup
	for _, serv := range d.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to parse to address %q: %w", serv, err))
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

func (d *Disque) gatherServer(addr *url.URL, acc telegraf.Accumulator) error {
	if d.c == nil {
		_, _, err := net.SplitHostPort(addr.Host)
		if err != nil {
			addr.Host = addr.Host + ":" + defaultPort
		}

		c, err := net.DialTimeout("tcp", addr.Host, defaultTimeout)
		if err != nil {
			return fmt.Errorf("unable to connect to disque server %q: %w", addr.Host, err)
		}

		if addr.User != nil {
			pwd, set := addr.User.Password()
			if set && pwd != "" {
				if _, err := fmt.Fprintf(c, "AUTH %s\r\n", pwd); err != nil {
					return err
				}

				r := bufio.NewReader(c)

				line, err := r.ReadString('\n')
				if err != nil {
					return err
				}
				if line[0] != '+' {
					return errors.New(strings.TrimSpace(line)[1:])
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
		return fmt.Errorf("bad line start: %w", errProtocol)
	}

	line = strings.TrimSpace(line)

	szStr := line[1:]

	sz, err := strconv.Atoi(szStr)
	if err != nil {
		return fmt.Errorf("bad size string <<%s>>: %w", szStr, errProtocol)
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

		metric, ok := tracking[name]
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
