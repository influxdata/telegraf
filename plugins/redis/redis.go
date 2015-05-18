package redis

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/influxdb/tivan/plugins"
)

type Redis struct {
	Disabled bool
	Address  string
	Servers  []string

	c   net.Conn
	buf []byte
}

var Tracking = map[string]string{
	"uptime_in_seconds":           "redis_uptime",
	"connected_clients":           "redis_clients",
	"used_memory":                 "redis_used_memory",
	"used_memory_rss":             "redis_used_memory_rss",
	"used_memory_peak":            "redis_used_memory_peak",
	"used_memory_lua":             "redis_used_memory_lua",
	"rdb_changes_since_last_save": "redis_rdb_changes_since_last_save",
	"total_connections_received":  "redis_total_connections_received",
	"total_commands_processed":    "redis_total_commands_processed",
	"instantaneous_ops_per_sec":   "redis_instantaneous_ops_per_sec",
	"sync_full":                   "redis_sync_full",
	"sync_partial_ok":             "redis_sync_partial_ok",
	"sync_partial_err":            "redis_sync_partial_err",
	"expired_keys":                "redis_expired_keys",
	"evicted_keys":                "redis_evicted_keys",
	"keyspace_hits":               "redis_keyspace_hits",
	"keyspace_misses":             "redis_keyspace_misses",
	"pubsub_channels":             "redis_pubsub_channels",
	"pubsub_patterns":             "redis_pubsub_patterns",
	"latest_fork_usec":            "redis_latest_fork_usec",
	"connected_slaves":            "redis_connected_slaves",
	"master_repl_offset":          "redis_master_repl_offset",
	"repl_backlog_active":         "redis_repl_backlog_active",
	"repl_backlog_size":           "redis_repl_backlog_size",
	"repl_backlog_histlen":        "redis_repl_backlog_histlen",
	"mem_fragmentation_ratio":     "redis_mem_fragmentation_ratio",
	"used_cpu_sys":                "redis_used_cpu_sys",
	"used_cpu_user":               "redis_used_cpu_user",
	"used_cpu_sys_children":       "redis_used_cpu_sys_children",
	"used_cpu_user_children":      "redis_used_cpu_user_children",
}

var ErrProtocolError = errors.New("redis protocol error")

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (g *Redis) Gather(acc plugins.Accumulator) error {
	if g.Disabled {
		return nil
	}

	var servers []string

	if g.Address != "" {
		servers = append(servers, g.Address)
	}

	servers = append(servers, g.Servers...)

	if len(servers) == 0 {
		g.gatherServer(":6379", acc)
		return nil
	}

	var wg sync.WaitGroup

	var outerr error

	for _, serv := range servers {
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			outerr = g.gatherServer(serv, acc)
		}(serv)
	}

	wg.Wait()

	return outerr
}

func (g *Redis) gatherServer(addr string, acc plugins.Accumulator) error {
	if g.c == nil {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			return err
		}

		g.c = c
	}

	g.c.Write([]byte("info\n"))

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

		val := strings.TrimSpace(parts[1])

		ival, err := strconv.ParseUint(val, 10, 64)
		if err == nil {
			acc.Add(metric, ival, nil)
			continue
		}

		fval, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}

		acc.Add(metric, fval, nil)
	}

	return nil
}

func init() {
	plugins.Add("redis", func() plugins.Plugin {
		return &Redis{}
	})
}
