package redis_sentinel

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type RedisSentinel struct {
	Servers  []string `toml:"servers"`
	Password string   `toml:"password"`
	tls.ClientConfig

	clients []*RedisSentinelClient
}

type RedisSentinelClient struct {
	sentinel *redis.SentinelClient
	tags     map[string]string
}

const measurementMasters = "redis_sentinel_masters"
const measurementSentinel = "redis_sentinel"
const measurementSentinels = "redis_sentinel_sentinels"
const measurementReplicas = "redis_sentinel_replicas"

// Supported fields for "redis_sentinel_masters"
var measurementMastersFields = map[string]string{
	"config_epoch":            "integer",
	"down_after_milliseconds": "integer",
	"failover_timeout":        "integer",
	"flags":                   "string",
	"info_refresh":            "integer",
	"ip":                      "string",
	"last_ok_ping_reply":      "integer",
	"last_ping_reply":         "integer",
	"last_ping_sent":          "integer",
	"link_pending_commands":   "integer",
	"link_refcount":           "integer",
	"name":                    "string",
	"num_other_sentinels":     "integer",
	"num_slaves":              "integer",
	"parallel_syncs":          "integer",
	"port":                    "integer",
	"quorum":                  "integer",
	"role_reported":           "string",
	"role_reported_time":      "integer",
}

// Supported fields for "redis_sentinel"
var measurementSentinelFields = map[string]string{
	"active_defrag_hits":              "integer",
	"active_defrag_key_hits":          "integer",
	"active_defrag_key_misses":        "integer",
	"active_defrag_misses":            "integer",
	"blocked_clients":                 "integer",
	"client_recent_max_input_buffer":  "integer",
	"client_recent_max_output_buffer": "integer",
	"clients":                         "integer", // Renamed field
	"evicted_keys":                    "integer",
	"expired_keys":                    "integer",
	"expired_stale_perc":              "float",
	"expired_time_cap_reached_count":  "integer",
	"instantaneous_input_kbps":        "float",
	"instantaneous_ops_per_sec":       "integer",
	"instantaneous_output_kbps":       "float",
	"keyspace_hits":                   "integer",
	"keyspace_misses":                 "integer",
	"latest_fork_usec":                "integer",
	"lru_clock":                       "integer",
	"migrate_cached_sockets":          "integer",
	"pubsub_channels":                 "integer",
	"pubsub_patterns":                 "integer",
	"redis_version":                   "string",
	"rejected_connections":            "integer",
	"sentinel_masters":                "integer",
	"sentinel_running_scripts":        "integer",
	"sentinel_scripts_queue_length":   "integer",
	"sentinel_simulate_failure_flags": "integer",
	"sentinel_tilt":                   "integer",
	"slave_expires_tracked_keys":      "integer",
	"sync_full":                       "integer",
	"sync_partial_err":                "integer",
	"sync_partial_ok":                 "integer",
	"total_commands_processed":        "integer",
	"total_connections_received":      "integer",
	"total_net_input_bytes":           "integer",
	"total_net_output_bytes":          "integer",
	"used_cpu_sys":                    "float",
	"used_cpu_sys_children":           "float",
	"used_cpu_user":                   "float",
	"used_cpu_user_children":          "float",
}

// Supported fields for "redis_sentinel_sentinels"
var measurementSentinelsFields = map[string]string{
	"down_after_milliseconds": "integer",
	"flags":                   "string",
	"last_hello_message":      "integer",
	"last_ok_ping_reply":      "integer",
	"last_ping_reply":         "integer",
	"last_ping_sent":          "integer",
	"link_pending_commands":   "integer",
	"link_refcount":           "integer",
	"name":                    "string",
	"voted_leader":            "string",
	"voted_leader_epoch":      "integer",
}

// Supported fields for "redis_sentinel_replicas"
var measurementReplicasFields = map[string]string{
	"down_after_milliseconds": "integer",
	"flags":                   "string",
	"info_refresh":            "integer",
	"last_ok_ping_reply":      "integer",
	"last_ping_reply":         "integer",
	"last_ping_sent":          "integer",
	"link_pending_commands":   "integer",
	"link_refcount":           "integer",
	"master_host":             "string",
	"master_link_down_time":   "integer",
	"master_link_status":      "string",
	"master_port":             "integer",
	"name":                    "string",
	"role_reported":           "string",
	"role_reported_time":      "integer",
	"slave_priority":          "integer",
	"slave_repl_offset":       "integer",
}

func init() {
	inputs.Add("redis_sentinel", func() telegraf.Input {
		return &RedisSentinel{}
	})
}

func (r *RedisSentinelClient) baseTags() map[string]string {
	tags := make(map[string]string)
	for k, v := range r.tags {
		tags[k] = v
	}
	return tags
}

func (r *RedisSentinel) SampleConfig() string {
	return `
  ## specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:26379
  ##    tcp://:password@192.168.99.100
  ##    unix:///var/run/redis-sentinel.sock
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 26379 is used
  servers = ["tcp://localhost:26379"]

  ## specify server password
  # password = "s#cr@t%"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
`
}

func (r *RedisSentinel) Description() string {
	return "Read metrics from one or many redis-sentinel servers"
}

func (r *RedisSentinel) Init() error {
	if len(r.Servers) == 0 {
		r.Servers = []string{"tcp://localhost:26379"}
	}

	r.clients = make([]*RedisSentinelClient, len(r.Servers))

	tlsConfig, err := r.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	for i, serv := range r.Servers {
		u, err := url.Parse(serv)
		if err != nil {
			return fmt.Errorf("unable to parse to address %q: %s", serv, err.Error())
		}

		if u.Scheme != "tcp" && u.Scheme != "unix" {
			return fmt.Errorf("invalid scheme %q. expected tcp or unix", u.Scheme)
		}

		password := ""

		if len(r.Password) > 0 {
			password = r.Password
		}

		var address string
		if u.Scheme == "unix" {
			address = u.Path
		} else {
			address = u.Host
		}

		sentinel := redis.NewSentinelClient(
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
			tags["source"] = u.Hostname()
			tags["port"] = u.Port()
		}

		r.clients[i] = &RedisSentinelClient{
			sentinel: sentinel,
			tags:     tags,
		}
	}

	return nil
}

// Redis list format has string key/values adjacent, so convert to a map for easier use
func toMap(vals []interface{}) map[string]string {
	m := make(map[string]string)

	for idx := 0; idx < len(vals)-1; idx += 2 {
		key, keyOk := vals[idx].(string)
		value, valueOk := vals[idx+1].(string)

		if keyOk && valueOk {
			m[key] = value
		}
	}

	return m
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (r *RedisSentinel) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	for _, client := range r.clients {
		wg.Add(1)

		go func(client *RedisSentinelClient, acc telegraf.Accumulator) {
			defer wg.Done()

			gatherMasterStats(client, acc)
			gatherInfoStats(client, acc)
		}(client, acc)
	}

	wg.Wait()

	return nil
}

func gatherInfoStats(client *RedisSentinelClient, acc telegraf.Accumulator) {
	infoCmd := redis.NewStringCmd("info", "all")
	// We check the command result below
	//nolint:errcheck
	client.sentinel.Process(infoCmd)

	info, infoErr := infoCmd.Result()
	if infoErr != nil {
		acc.AddError(infoErr)
		return
	}

	rdr := strings.NewReader(info)
	infoTags, infoFields := convertSentinelInfoOutput(acc, client.baseTags(), rdr)

	acc.AddFields(measurementSentinel, infoFields, infoTags)
}

func gatherMasterStats(client *RedisSentinelClient, acc telegraf.Accumulator) {
	mastersCmd := redis.NewSliceCmd("sentinel", "masters")
	// We check the command result below
	//nolint:errcheck
	client.sentinel.Process(mastersCmd)

	masters, mastersErr := mastersCmd.Result()
	if mastersErr != nil {
		acc.AddError(mastersErr)
		return
	}

	for _, master := range masters {
		master, masterOk := master.([]interface{})
		if !masterOk {
			continue
		}

		m := toMap(master)

		masterName, masterNameOk := m["name"]
		if !masterNameOk {
			acc.AddError(fmt.Errorf("unable to resolve master name"))
			return
		}

		quorumCmd := redis.NewStringCmd("sentinel", "ckquorum", masterName)
		// We check the command result below
		//nolint:errcheck
		client.sentinel.Process(quorumCmd)

		_, quorumErr := quorumCmd.Result()

		sentinelMastersTags, sentinelMastersFields := convertSentinelMastersOutput(acc, client.baseTags(), m, quorumErr)
		acc.AddFields(measurementMasters, sentinelMastersFields, sentinelMastersTags)

		gatherReplicaStats(client, acc, masterName)
		gatherSentinelStats(client, acc, masterName)
	}
}

func gatherReplicaStats(
	client *RedisSentinelClient,
	acc telegraf.Accumulator,
	masterName string,
) {
	replicasCmd := redis.NewSliceCmd("sentinel", "replicas", masterName)
	// We check the command result below
	//nolint:errcheck
	client.sentinel.Process(replicasCmd)

	replicas, replicasErr := replicasCmd.Result()
	if replicasErr != nil {
		acc.AddError(replicasErr)
		return
	}

	for _, replica := range replicas {
		if replica, replicaOk := replica.([]interface{}); replicaOk {
			rm := toMap(replica)

			replicaTags, replicaFields := convertSentinelReplicaOutput(acc, client.baseTags(), masterName, rm)
			acc.AddFields(measurementReplicas, replicaFields, replicaTags)
		}
	}
}

func gatherSentinelStats(
	client *RedisSentinelClient,
	acc telegraf.Accumulator,
	masterName string,
) {
	sentinelsCmd := redis.NewSliceCmd("sentinel", "sentinels", masterName)
	// We check the command result below
	//nolint:errcheck
	client.sentinel.Process(sentinelsCmd)

	sentinels, sentinelsErr := sentinelsCmd.Result()
	if sentinelsErr != nil {
		acc.AddError(sentinelsErr)
		return
	}

	for _, sentinel := range sentinels {
		if sentinel, sentinelOk := sentinel.([]interface{}); sentinelOk {
			sm := toMap(sentinel)

			sentinelTags, sentinelFields := convertSentinelSentinelsOutput(acc, client.baseTags(), masterName, sm)
			acc.AddFields(measurementSentinels, sentinelFields, sentinelTags)
		}
	}
}

// converts `sentinel masters <name>` output to tags and fields
func convertSentinelMastersOutput(
	acc telegraf.Accumulator,
	globalTags map[string]string,
	master map[string]string,
	quorumErr error,
) (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}

	tags["master"] = master["name"]

	fields := make(map[string]interface{})

	fields["has_quorum"] = false
	if quorumErr == nil {
		fields["has_quorum"] = true
	}

	for key, val := range master {
		key = strings.Replace(key, "-", "_", -1)

		switch valType := measurementMastersFields[key]; valType {
		case "float":
			if val, err := strconv.ParseFloat(val, 64); err == nil {
				fields[key] = val
			} else {
				acc.AddError(err)
			}
		case "integer":
			if val, err := strconv.ParseInt(val, 10, 64); err == nil {
				fields[key] = val
			} else {
				acc.AddError(err)
			}
		case "string":
			fields[key] = val
		default:
			continue
		}
	}

	return tags, fields
}

// converts `sentinel sentinels <name>` output to tags and fields
func convertSentinelSentinelsOutput(
	acc telegraf.Accumulator,
	globalTags map[string]string,
	masterName string,
	sentinelMaster map[string]string,
) (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}

	tags["sentinel_ip"] = sentinelMaster["ip"]
	tags["sentinel_port"] = sentinelMaster["port"]
	tags["master"] = masterName

	fields := make(map[string]interface{})

	for key, val := range sentinelMaster {
		key = strings.Replace(key, "-", "_", -1)

		switch valType := measurementSentinelsFields[key]; valType {
		case "float":
			if val, err := strconv.ParseFloat(val, 64); err == nil {
				fields[key] = val
			} else {
				acc.AddError(err)
			}
		case "integer":
			if val, err := strconv.ParseInt(val, 10, 64); err == nil {
				fields[key] = val
			} else {
				acc.AddError(err)
			}
		case "string":
			fields[key] = val
		default:
			continue
		}
	}

	return tags, fields
}

// converts `sentinel replicas <name>` output to tags and fields
func convertSentinelReplicaOutput(
	acc telegraf.Accumulator,
	globalTags map[string]string,
	masterName string,
	replica map[string]string,
) (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}

	tags["replica_ip"] = replica["ip"]
	tags["replica_port"] = replica["port"]
	tags["master"] = masterName

	fields := make(map[string]interface{})

	for key, val := range replica {
		key = strings.Replace(key, "-", "_", -1)

		switch valType := measurementReplicasFields[key]; valType {
		case "float":
			if val, err := strconv.ParseFloat(val, 64); err == nil {
				fields[key] = val
			} else {
				acc.AddError(err)
			}
		case "integer":
			if val, err := strconv.ParseInt(val, 10, 64); err == nil {
				fields[key] = val
			} else {
				acc.AddError(err)
			}
		case "string":
			fields[key] = val
		default:
			continue
		}
	}

	return tags, fields
}

// convertSentinelInfoOutput parses `INFO` command output
// Largely copied from the Redis input plugin's gatherInfoOutput()
func convertSentinelInfoOutput(
	acc telegraf.Accumulator,
	globalTags map[string]string,
	rdr io.Reader,
) (map[string]string, map[string]interface{}) {
	scanner := bufio.NewScanner(rdr)
	fields := make(map[string]interface{})

	tags := make(map[string]string)
	for k, v := range globalTags {
		tags[k] = v
	}

	var section string
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

		name := parts[0]

		if section == "Server" {
			// Rename and convert to nanoseconds
			if name == "uptime_in_seconds" {
				if uptimeInSeconds, uptimeParseErr := strconv.ParseInt(parts[1], 10, 64); uptimeParseErr == nil {
					fields["uptime_ns"] = int64(time.Duration(uptimeInSeconds) * time.Second)
				} else {
					acc.AddError(uptimeParseErr)
				}

				continue
			}
		} else if section == "Clients" {
			// Rename in order to match the "redis" input plugin
			if name == "connected_clients" {
				name = "clients"
			}
		}

		metric := strings.Replace(name, "-", "_", -1)

		val := strings.TrimSpace(parts[1])

		switch valType := measurementSentinelFields[metric]; valType {
		case "float":
			if val, err := strconv.ParseFloat(val, 64); err == nil {
				fields[metric] = val
			} else {
				acc.AddError(err)
			}
		case "integer":
			if val, err := strconv.ParseInt(val, 10, 64); err == nil {
				fields[metric] = val
			} else {
				acc.AddError(err)
			}
		case "string":
			fields[metric] = val
		default:
			continue
		}
	}

	return tags, fields
}
