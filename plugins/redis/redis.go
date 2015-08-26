package redis

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdb/telegraf/plugins"
)

type Redis struct {
	Servers []string

	c   net.Conn
	buf []byte
}

var sampleConfig = `
	# An array of URI to gather stats about. Specify an ip or hostname
	# with optional port add password. ie redis://localhost, redis://10.10.3.33:18832,
	# 10.0.0.1:10000, etc.
	#
	# If no servers are specified, then localhost is used as the host.
	servers = ["localhost"]
`

func (r *Redis) SampleConfig() string {
	return sampleConfig
}

func (r *Redis) Description() string {
	return "Read metrics from one or many redis servers"
}

var Tracking = map[string]string{
	"uptime_in_seconds":           "uptime",
	"connected_clients":           "clients",
	"used_memory":                 "used_memory",
	"used_memory_rss":             "used_memory_rss",
	"used_memory_peak":            "used_memory_peak",
	"used_memory_lua":             "used_memory_lua",
	"rdb_changes_since_last_save": "rdb_changes_since_last_save",
	"total_connections_received":  "total_connections_received",
	"total_commands_processed":    "total_commands_processed",
	"instantaneous_ops_per_sec":   "instantaneous_ops_per_sec",
	"sync_full":                   "sync_full",
	"sync_partial_ok":             "sync_partial_ok",
	"sync_partial_err":            "sync_partial_err",
	"expired_keys":                "expired_keys",
	"evicted_keys":                "evicted_keys",
	"keyspace_hits":               "keyspace_hits",
	"keyspace_misses":             "keyspace_misses",
	"pubsub_channels":             "pubsub_channels",
	"pubsub_patterns":             "pubsub_patterns",
	"latest_fork_usec":            "latest_fork_usec",
	"connected_slaves":            "connected_slaves",
	"master_repl_offset":          "master_repl_offset",
	"repl_backlog_active":         "repl_backlog_active",
	"repl_backlog_size":           "repl_backlog_size",
	"repl_backlog_histlen":        "repl_backlog_histlen",
	"mem_fragmentation_ratio":     "mem_fragmentation_ratio",
	"used_cpu_sys":                "used_cpu_sys",
	"used_cpu_user":               "used_cpu_user",
	"used_cpu_sys_children":       "used_cpu_sys_children",
	"used_cpu_user_children":      "used_cpu_user_children",
}

var ErrProtocolError = errors.New("redis protocol error")

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *Redis) Gather(acc plugins.Accumulator) error {
	if len(g.Servers) == 0 {
		url := &url.URL{
			Host: ":6379",
		}
		g.gatherServer(url, acc)
		return nil
	}

	var wg sync.WaitGroup

	var outerr error

	for _, serv := range g.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			return fmt.Errorf("Unable to parse to address '%s': %s", serv, err)
		} else if u.Scheme == "" {
			// fallback to simple string based address (i.e. "10.0.0.1:10000")
			u.Scheme = "tcp"
			u.Host = serv
			u.Path = ""
		}
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = g.gatherServer(u, acc)
		}(serv)
	}

	wg.Wait()

	return outerr
}

const defaultPort = "6379"

func (g *Redis) gatherServer(addr *url.URL, acc plugins.Accumulator) error {
	if g.c == nil {

		_, _, err := net.SplitHostPort(addr.Host)
		if err != nil {
			addr.Host = addr.Host + ":" + defaultPort
		}

		c, err := net.Dial("tcp", addr.Host)
		if err != nil {
			return fmt.Errorf("Unable to connect to redis server '%s': %s", addr.Host, err)
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

		_, rPort, err := net.SplitHostPort(addr.Host)
		if err != nil {
			rPort = defaultPort
		}
		tags := map[string]string{"host": addr.String(), "port": rPort}

		val := strings.TrimSpace(parts[1])

		ival, err := strconv.ParseUint(val, 10, 64)
		if err == nil {
			acc.Add(metric, ival, tags)
			continue
		}

		fval, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}

		acc.Add(metric, fval, tags)
	}

	return nil
}

func init() {
	plugins.Add("redis", func() plugins.Plugin {
		return &Redis{}
	})
}
