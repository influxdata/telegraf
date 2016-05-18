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
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Redis struct {
	Servers []string
}

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]
`

var defaultTimeout = 5 * time.Second

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
	"instantaneous_input_kbps":    "instantaneous_input_kbps",
	"instantaneous_output_kbps":   "instantaneous_output_kbps",
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
	"role": "role",
}

var ErrProtocolError = errors.New("redis protocol error")

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (r *Redis) Gather(acc telegraf.Accumulator) error {
	if len(r.Servers) == 0 {
		url := &url.URL{
			Host: ":6379",
		}
		r.gatherServer(url, acc)
		return nil
	}

	var wg sync.WaitGroup

	var outerr error

	for _, serv := range r.Servers {
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
			outerr = r.gatherServer(u, acc)
		}(serv)
	}

	wg.Wait()

	return outerr
}

const defaultPort = "6379"

func (r *Redis) gatherServer(addr *url.URL, acc telegraf.Accumulator) error {
	_, _, err := net.SplitHostPort(addr.Host)
	if err != nil {
		addr.Host = addr.Host + ":" + defaultPort
	}

	c, err := net.DialTimeout("tcp", addr.Host, defaultTimeout)
	if err != nil {
		return fmt.Errorf("Unable to connect to redis server '%s': %s", addr.Host, err)
	}
	defer c.Close()

	// Extend connection
	c.SetDeadline(time.Now().Add(defaultTimeout))

	if addr.User != nil {
		pwd, set := addr.User.Password()
		if set && pwd != "" {
			c.Write([]byte(fmt.Sprintf("AUTH %s\r\n", pwd)))

			rdr := bufio.NewReader(c)

			line, err := rdr.ReadString('\n')
			if err != nil {
				return err
			}
			if line[0] != '+' {
				return fmt.Errorf("%s", strings.TrimSpace(line)[1:])
			}
		}
	}

	c.Write([]byte("INFO\r\n"))
	c.Write([]byte("EOF\r\n"))
	rdr := bufio.NewReader(c)

	// Setup tags for all redis metrics
	host, port := "unknown", "unknown"
	// If there's an error, ignore and use 'unknown' tags
	host, port, _ = net.SplitHostPort(addr.Host)
	tags := map[string]string{"server": host, "port": port}

	return gatherInfoOutput(rdr, acc, tags)
}

// gatherInfoOutput gathers
func gatherInfoOutput(
	rdr *bufio.Reader,
	acc telegraf.Accumulator,
	tags map[string]string,
) error {
	var keyspace_hits, keyspace_misses uint64 = 0, 0

	scanner := bufio.NewScanner(rdr)
	fields := make(map[string]interface{})
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ERR") {
			break
		}

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		name := string(parts[0])
		metric, ok := Tracking[name]
		if !ok {
			kline := strings.TrimSpace(string(parts[1]))
			gatherKeyspaceLine(name, kline, acc, tags)
			continue
		}

		val := strings.TrimSpace(parts[1])
		ival, err := strconv.ParseUint(val, 10, 64)

		if name == "keyspace_hits" {
			keyspace_hits = ival
		}

		if name == "keyspace_misses" {
			keyspace_misses = ival
		}

		if name == "role" {
			tags["role"] = val
			continue
		}

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
	var keyspace_hitrate float64 = 0.0
	if keyspace_hits != 0 || keyspace_misses != 0 {
		keyspace_hitrate = float64(keyspace_hits) / float64(keyspace_hits+keyspace_misses)
	}
	fields["keyspace_hitrate"] = keyspace_hitrate
	acc.AddFields("redis", fields, tags)
	return nil
}

// Parse the special Keyspace line at end of redis stats
// This is a special line that looks something like:
//     db0:keys=2,expires=0,avg_ttl=0
// And there is one for each db on the redis instance
func gatherKeyspaceLine(
	name string,
	line string,
	acc telegraf.Accumulator,
	tags map[string]string,
) {
	if strings.Contains(line, "keys=") {
		fields := make(map[string]interface{})
		tags["database"] = name
		dbparts := strings.Split(line, ",")
		for _, dbp := range dbparts {
			kv := strings.Split(dbp, "=")
			ival, err := strconv.ParseUint(kv[1], 10, 64)
			if err == nil {
				fields[kv[0]] = ival
			}
		}
		acc.AddFields("redis_keyspace", fields, tags)
	}
}

func init() {
	inputs.Add("redis", func() telegraf.Input {
		return &Redis{}
	})
}
