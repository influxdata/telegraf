package redis

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type RedisCommand struct {
	Command []interface{}
	Field   string
	Type    string
}

type Redis struct {
	Commands []*RedisCommand
	Servers  []string
	Password string
	tls.ClientConfig

	Log telegraf.Logger

	clients     []Client
	initialized bool
}

type Client interface {
	Do(returnType string, args ...interface{}) (interface{}, error)
	Info() *redis.StringCmd
	BaseTags() map[string]string
}

type RedisClient struct {
	client *redis.Client
	tags   map[string]string
}

type RedisOutputs struct {
	Uptime                     int64   `json:"uptime"`
	LruClock                   int64   `json:"lru_clock"`
	Clients                    int64   `json:"clients"`
	ClientLongestOutputList    int64   `json:"client_longest_output_list"`
	ClientBiggestInputBuf      int64   `json:"client_biggest_input_buf"`
	BlockedClients             int64   `json:"blocked_clients"`
	UsedMemory                 int64   `json:"used_memory"`
	UsedMemoryRss              int64   `json:"used_memory_rss"`
	UsedMemoryPeak             int64   `json:"used_memory_peak"`
	UsedMemoryLua              int64   `json:"used_memory_lua"`
	UsedMemoryPeakPerc         float64 `json:"used_memory_peak_perc"`
	UsedMemoryDatasetPerc      float64 `json:"used_memory_dataset_perc"`
	MemFragmentationRatio      float64 `json:"mem_fragmentation_ratio"`
	Loading                    int64   `json:"loading"`
	RdbChangesSinceLastSave    int64   `json:"rdb_changes_since_last_save"`
	RdbBgsaveInProgress        int64   `json:"rdb_bgsave_in_progress"`
	RdbLastSaveTime            int64   `json:"rdb_last_save_time"`
	RdbLastBgsaveStatus        string  `json:"rdb_last_bgsave_status"`
	RdbLastBgsaveTimeSec       int64   `json:"rdb_last_bgsave_time_sec"`
	RdbCurrentBgsaveTimeSec    int64   `json:"rdb_current_bgsave_time_sec"`
	AofEnabled                 int64   `json:"aof_enabled"`
	AofRewriteInProgress       int64   `json:"aof_rewrite_in_progress"`
	AofRewriteScheduled        int64   `json:"aof_rewrite_scheduled"`
	AofLastRewriteTimeSec      int64   `json:"aof_last_rewrite_time_sec"`
	AofCurrentRewriteTimeSec   int64   `json:"aof_current_rewrite_time_sec"`
	AofLastBgrewriteStatus     string  `json:"aof_last_bgrewrite_status"`
	AofLastWriteStatus         string  `json:"aof_last_write_status"`
	TotalConnectionsReceived   int64   `json:"total_connections_received"`
	TotalCommandsProcessed     int64   `json:"total_commands_processed"`
	InstantaneousOpsPerSec     int64   `json:"instantaneous_ops_per_sec"`
	InstantaneousInputKbps     float64 `json:"instantaneous_input_kbps"`
	InstantaneousOutputKbps    float64 `json:"instantaneous_output_kbps"`
	RejectedConnections        int64   `json:"rejected_connections"`
	SyncFull                   int64   `json:"sync_full"`
	SyncPartialOk              int64   `json:"sync_partial_ok"`
	SyncPartialErr             int64   `json:"sync_partial_err"`
	ExpiredKeys                int64   `json:"expired_keys"`
	EvictedKeys                int64   `json:"evicted_keys"`
	KeyspaceHits               int64   `json:"keyspace_hits"`
	KeyspaceMisses             int64   `json:"keyspace_misses"`
	PubsubChannels             int64   `json:"pubsub_channels"`
	PubsubPatterns             int64   `json:"pubsub_patterns"`
	LatestForkUsec             int64   `json:"latest_fork_usec"`
	ConnectedSlaves            int64   `json:"connected_slaves"`
	MasterReplOffset           int64   `json:"master_repl_offset"`
	ReplBacklogActive          int64   `json:"repl_backlog_active"`
	ReplBacklogSize            int64   `json:"repl_backlog_size"`
	ReplBacklogFirstByteOffset int64   `json:"repl_backlog_first_byte_offset"`
	ReplBacklogHistlen         int64   `json:"repl_backlog_histlen"`
	SecondReplOffset           int64   `json:"second_repl_offset"`
	UsedCpuSys                 float64 `json:"used_cpu_sys"`
	UsedCpuUser                float64 `json:"used_cpu_user"`
	UsedCpuSysChildren         float64 `json:"used_cpu_sys_children"`
	UsedCpuUserChildren        float64 `json:"used_cpu_user_children"`
	KeyspaceHitrate            float64 `json:"keyspace_hitrate"`
	RedisVersion               string  `json:"redis_version"`
}

func (r *RedisClient) Do(returnType string, args ...interface{}) (interface{}, error) {
	rawVal := r.client.Do(args...)

	switch returnType {
	case "integer":
		return rawVal.Int64()
	case "string":
		return rawVal.String()
	case "float":
		return rawVal.Float64()
	default:
		return rawVal.String()
	}
}

func (r *RedisClient) Info() *redis.StringCmd {
	return r.client.Info("ALL")
}

func (r *RedisClient) BaseTags() map[string]string {
	tags := make(map[string]string)
	for k, v := range r.tags {
		tags[k] = v
	}
	return tags
}

var replicationSlaveMetricPrefix = regexp.MustCompile(`^slave\d+`)

var sampleConfig = `
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##    unix:///var/run/redis.sock
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]

  ## Optional. Specify redis commands to retrieve values
  # [[inputs.redis.commands]]
  # command = ["get", "sample-key"]
  # field = "sample-key-value"
  # type = "string"

  ## specify server password
  # password = "s#cr@t%"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
`

func (r *Redis) SampleConfig() string {
	return sampleConfig
}

func (r *Redis) Description() string {
	return "Read metrics from one or many redis servers"
}

var Tracking = map[string]string{
	"uptime_in_seconds": "uptime",
	"connected_clients": "clients",
	"role":              "replication_role",
}

func (r *Redis) init(acc telegraf.Accumulator) error {
	if r.initialized {
		return nil
	}

	if len(r.Servers) == 0 {
		r.Servers = []string{"tcp://localhost:6379"}
	}

	r.clients = make([]Client, len(r.Servers))

	for i, serv := range r.Servers {
		if !strings.HasPrefix(serv, "tcp://") && !strings.HasPrefix(serv, "unix://") {
			r.Log.Warn("Server URL found without scheme; please update your configuration file")
			serv = "tcp://" + serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			return fmt.Errorf("unable to parse to address %q: %s", serv, err.Error())
		}

		password := ""
		if u.User != nil {
			pw, ok := u.User.Password()
			if ok {
				password = pw
			}
		}
		if len(r.Password) > 0 {
			password = r.Password
		}

		var address string
		if u.Scheme == "unix" {
			address = u.Path
		} else {
			address = u.Host
		}

		tlsConfig, err := r.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		client := redis.NewClient(
			&redis.Options{
				Addr:      address,
				Password:  password,
				Network:   u.Scheme,
				PoolSize:  1,
				TLSConfig: tlsConfig,
			},
		)

		tags := map[string]string{}
		if u.Scheme == "unix" {
			tags["socket"] = u.Path
		} else {
			tags["server"] = u.Hostname()
			tags["port"] = u.Port()
		}

		r.clients[i] = &RedisClient{
			client: client,
			tags:   tags,
		}
	}

	r.initialized = true
	return nil
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (r *Redis) Gather(acc telegraf.Accumulator) error {
	if !r.initialized {
		err := r.init(acc)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	for _, client := range r.clients {
		wg.Add(1)
		go func(client Client) {
			defer wg.Done()
			acc.AddError(r.gatherServer(client, acc))
			acc.AddError(r.gatherCommandValues(client, acc))
		}(client)
	}

	wg.Wait()
	return nil
}

func (r *Redis) gatherCommandValues(client Client, acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	for _, command := range r.Commands {
		val, err := client.Do(command.Type, command.Command...)
		if err != nil {
			return err
		}

		fields[command.Field] = val
	}

	acc.AddFields("redis_commands", fields, client.BaseTags())

	return nil
}

func (r *Redis) gatherServer(client Client, acc telegraf.Accumulator) error {
	info, err := client.Info().Result()
	if err != nil {
		return err
	}

	rdr := strings.NewReader(info)
	return gatherInfoOutput(rdr, acc, client.BaseTags())
}

// gatherInfoOutput gathers
func gatherInfoOutput(
	rdr io.Reader,
	acc telegraf.Accumulator,
	tags map[string]string,
) error {
	var section string
	var keyspace_hits, keyspace_misses int64

	scanner := bufio.NewScanner(rdr)
	fields := make(map[string]interface{})
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			continue
		}

		if line[0] == '#' {
			if len(line) > 2 {
				section = line[2:]
			}
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		name := string(parts[0])

		if section == "Server" {
			if name != "lru_clock" && name != "uptime_in_seconds" && name != "redis_version" {
				continue
			}
		}

		if strings.HasPrefix(name, "master_replid") {
			continue
		}

		if name == "mem_allocator" {
			continue
		}

		if strings.HasSuffix(name, "_human") {
			continue
		}

		metric, ok := Tracking[name]
		if !ok {
			if section == "Keyspace" {
				kline := strings.TrimSpace(string(parts[1]))
				gatherKeyspaceLine(name, kline, acc, tags)
				continue
			}
			if section == "Commandstats" {
				kline := strings.TrimSpace(parts[1])
				gatherCommandstateLine(name, kline, acc, tags)
				continue
			}
			if section == "Replication" && replicationSlaveMetricPrefix.MatchString(name) {
				kline := strings.TrimSpace(parts[1])
				gatherReplicationLine(name, kline, acc, tags)
				continue
			}

			metric = name
		}

		val := strings.TrimSpace(parts[1])

		// Some percentage values have a "%" suffix that we need to get rid of before int/float conversion
		val = strings.TrimSuffix(val, "%")

		// Try parsing as int
		if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
			switch name {
			case "keyspace_hits":
				keyspace_hits = ival
			case "keyspace_misses":
				keyspace_misses = ival
			case "rdb_last_save_time":
				// influxdb can't calculate this, so we have to do it
				fields["rdb_last_save_time_elapsed"] = time.Now().Unix() - ival
			}
			fields[metric] = ival
			continue
		}

		// Try parsing as a float
		if fval, err := strconv.ParseFloat(val, 64); err == nil {
			fields[metric] = fval
			continue
		}

		// Treat it as a string

		if name == "role" {
			tags["replication_role"] = val
			continue
		}

		fields[metric] = val
	}
	var keyspace_hitrate float64 = 0.0
	if keyspace_hits != 0 || keyspace_misses != 0 {
		keyspace_hitrate = float64(keyspace_hits) / float64(keyspace_hits+keyspace_misses)
	}
	fields["keyspace_hitrate"] = keyspace_hitrate

	o := RedisOutputs{}

	m, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	err = json.Unmarshal(m, &o)
	if err != nil {
		return err
	}

	fields["uptime"] = o.Uptime
	fields["lru_clock"] = o.LruClock
	fields["clients"] = o.Clients
	fields["client_longest_output_list"] = o.ClientLongestOutputList
	fields["client_biggest_input_buf"] = o.ClientBiggestInputBuf
	fields["blocked_clients"] = o.BlockedClients
	fields["used_memory"] = o.UsedMemory
	fields["used_memory_rss"] = o.UsedMemoryRss
	fields["used_memory_peak"] = o.UsedMemoryPeak
	fields["used_memory_lua"] = o.UsedMemoryLua
	fields["used_memory_peak_perc"] = o.UsedMemoryPeakPerc
	fields["used_memory_dataset_perc"] = o.UsedMemoryDatasetPerc
	fields["mem_fragmentation_ratio"] = o.MemFragmentationRatio
	fields["loading"] = o.Loading
	fields["rdb_changes_since_last_save"] = o.RdbChangesSinceLastSave
	fields["rdb_bgsave_in_progress"] = o.RdbBgsaveInProgress
	fields["rdb_last_save_time"] = o.RdbLastSaveTime
	fields["rdb_last_bgsave_status"] = o.RdbLastBgsaveStatus
	fields["rdb_last_bgsave_time_sec"] = o.RdbLastBgsaveTimeSec
	fields["rdb_current_bgsave_time_sec"] = o.RdbCurrentBgsaveTimeSec
	fields["aof_enabled"] = o.AofEnabled
	fields["aof_rewrite_in_progress"] = o.AofRewriteInProgress
	fields["aof_rewrite_scheduled"] = o.AofRewriteScheduled
	fields["aof_last_rewrite_time_sec"] = o.AofLastRewriteTimeSec
	fields["aof_current_rewrite_time_sec"] = o.AofCurrentRewriteTimeSec
	fields["aof_last_bgrewrite_status"] = o.AofLastBgrewriteStatus
	fields["aof_last_write_status"] = o.AofLastWriteStatus
	fields["total_connections_received"] = o.TotalConnectionsReceived
	fields["total_commands_processed"] = o.TotalCommandsProcessed
	fields["instantaneous_ops_per_sec"] = o.InstantaneousOpsPerSec
	fields["instantaneous_input_kbps"] = o.InstantaneousInputKbps
	fields["instantaneous_output_kbps"] = o.InstantaneousOutputKbps
	fields["rejected_connections"] = o.RejectedConnections
	fields["sync_full"] = o.SyncFull
	fields["sync_partial_ok"] = o.SyncPartialOk
	fields["sync_partial_err"] = o.SyncPartialErr
	fields["expired_keys"] = o.ExpiredKeys
	fields["evicted_keys"] = o.EvictedKeys
	fields["keyspace_hits"] = o.KeyspaceHits
	fields["keyspace_misses"] = o.KeyspaceMisses
	fields["pubsub_channels"] = o.PubsubChannels
	fields["pubsub_patterns"] = o.PubsubPatterns
	fields["latest_fork_usec"] = o.LatestForkUsec
	fields["connected_slaves"] = o.ConnectedSlaves
	fields["master_repl_offset"] = o.MasterReplOffset
	fields["repl_backlog_active"] = o.ReplBacklogActive
	fields["repl_backlog_size"] = o.ReplBacklogSize
	fields["repl_backlog_first_byte_offset"] = o.ReplBacklogFirstByteOffset
	fields["repl_backlog_histlen"] = o.ReplBacklogHistlen
	fields["second_repl_offset"] = o.SecondReplOffset
	fields["used_cpu_sys"] = o.UsedCpuSys
	fields["used_cpu_user"] = o.UsedCpuUser
	fields["used_cpu_sys_children"] = o.UsedCpuSysChildren
	fields["used_cpu_user_children"] = o.UsedCpuUserChildren
	fields["keyspace_hitrate"] = o.KeyspaceHitrate
	fields["redis_version"] = o.RedisVersion

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
	global_tags map[string]string,
) {
	if strings.Contains(line, "keys=") {
		fields := make(map[string]interface{})
		tags := make(map[string]string)
		for k, v := range global_tags {
			tags[k] = v
		}
		tags["database"] = name
		dbparts := strings.Split(line, ",")
		for _, dbp := range dbparts {
			kv := strings.Split(dbp, "=")
			ival, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				fields[kv[0]] = ival
			}
		}
		acc.AddFields("redis_keyspace", fields, tags)
	}
}

// Parse the special cmdstat lines.
// Example:
//     cmdstat_publish:calls=33791,usec=208789,usec_per_call=6.18
// Tag: cmdstat=publish; Fields: calls=33791i,usec=208789i,usec_per_call=6.18
func gatherCommandstateLine(
	name string,
	line string,
	acc telegraf.Accumulator,
	global_tags map[string]string,
) {
	if !strings.HasPrefix(name, "cmdstat") {
		return
	}

	fields := make(map[string]interface{})
	tags := make(map[string]string)
	for k, v := range global_tags {
		tags[k] = v
	}
	tags["command"] = strings.TrimPrefix(name, "cmdstat_")
	parts := strings.Split(line, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "calls":
			fallthrough
		case "usec":
			ival, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				fields[kv[0]] = ival
			}
		case "usec_per_call":
			fval, err := strconv.ParseFloat(kv[1], 64)
			if err == nil {
				fields[kv[0]] = fval
			}
		}
	}
	acc.AddFields("redis_cmdstat", fields, tags)
}

// Parse the special Replication line
// Example:
//     slave0:ip=127.0.0.1,port=7379,state=online,offset=4556468,lag=0
// This line will only be visible when a node has a replica attached.
func gatherReplicationLine(
	name string,
	line string,
	acc telegraf.Accumulator,
	global_tags map[string]string,
) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	for k, v := range global_tags {
		tags[k] = v
	}

	tags["replica_id"] = strings.TrimLeft(name, "slave")
	tags["replication_role"] = "slave"

	parts := strings.Split(line, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "ip":
			tags["replica_ip"] = kv[1]
		case "port":
			tags["replica_port"] = kv[1]
		case "state":
			tags[kv[0]] = kv[1]
		default:
			ival, err := strconv.ParseInt(kv[1], 10, 64)
			if err == nil {
				fields[kv[0]] = ival
			}
		}
	}

	acc.AddFields("redis_replication", fields, tags)
}

func init() {
	inputs.Add("redis", func() telegraf.Input {
		return &Redis{}
	})
}
