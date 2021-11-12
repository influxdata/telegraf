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

func init() {
	inputs.Add("redis_sentinel", func() telegraf.Input {
		return &RedisSentinel{}
	})
}

func (r *RedisSentinelClient) baseTags() map[string]string {
	return r.tags
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
			return fmt.Errorf("unable to parse to address %q: %v", serv, err)
		}

		if u.Scheme != "tcp" && u.Scheme != "unix" {
			return fmt.Errorf("invalid scheme %q. expected tcp or unix", u.Scheme)
		}

		address := u.Host
		if u.Scheme == "unix" {
			address = u.Path
		}

		sentinel := redis.NewSentinelClient(
			&redis.Options{
				Addr:      address,
				Password:  r.Password,
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

		go func(acc telegraf.Accumulator, client *RedisSentinelClient) {
			defer wg.Done()

			gatherMasterStats(acc, client)
			gatherInfoStats(acc, client)
		}(acc, client)
	}

	wg.Wait()

	return nil
}

func gatherInfoStats(acc telegraf.Accumulator, client *RedisSentinelClient) {
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
	infoTags, infoFields, err := convertSentinelInfoOutput(client.baseTags(), rdr)
	if err != nil {
		acc.AddError(err)
		return
	}
	acc.AddFields(measurementSentinel, infoFields, infoTags)
}

func gatherMasterStats(acc telegraf.Accumulator, client *RedisSentinelClient) {
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

		sentinelMastersTags, sentinelMastersFields, err := convertSentinelMastersOutput(client.baseTags(), m, quorumErr)
		if err == nil {
			acc.AddFields(measurementMasters, sentinelMastersFields, sentinelMastersTags)
		} else {
			acc.AddError(err)
		}

		gatherReplicaStats(acc, client, masterName)
		gatherSentinelStats(acc, client, masterName)
	}
}

func gatherReplicaStats(
	acc telegraf.Accumulator,
	client *RedisSentinelClient,
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

			replicaTags, replicaFields, err := convertSentinelReplicaOutput(client.baseTags(), masterName, rm)
			if err == nil {
				acc.AddFields(measurementReplicas, replicaFields, replicaTags)
			} else {
				acc.AddError(err)
			}
		}
	}
}

func gatherSentinelStats(
	acc telegraf.Accumulator,
	client *RedisSentinelClient,
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

			sentinelTags, sentinelFields, err := convertSentinelSentinelsOutput(client.baseTags(), masterName, sm)
			if err == nil {
				acc.AddFields(measurementSentinels, sentinelFields, sentinelTags)
			} else {
				acc.AddError(err)
			}
		}
	}
}

// converts `sentinel masters <name>` output to tags and fields
func convertSentinelMastersOutput(
	globalTags map[string]string,
	master map[string]string,
	quorumErr error,
) (map[string]string, map[string]interface{}, error) {
	tags := globalTags

	tags["master"] = master["name"]

	fields := make(map[string]interface{})

	fields["has_quorum"] = quorumErr == nil

	for key, val := range master {
		key = strings.Replace(key, "-", "_", -1)

		switch valType := measurementMastersFields[key]; valType {
		case "float":
			val, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", key, err)
			}
			fields[key] = val
		case "integer":
			val, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", key, err)
			}
			fields[key] = val
		case "string":
			fields[key] = val
		}
	}

	return tags, fields, nil
}

// converts `sentinel sentinels <name>` output to tags and fields
func convertSentinelSentinelsOutput(
	globalTags map[string]string,
	masterName string,
	sentinelMaster map[string]string,
) (map[string]string, map[string]interface{}, error) {
	tags := globalTags

	tags["sentinel_ip"] = sentinelMaster["ip"]
	tags["sentinel_port"] = sentinelMaster["port"]
	tags["master"] = masterName

	fields := make(map[string]interface{})

	for key, val := range sentinelMaster {
		key = strings.Replace(key, "-", "_", -1)

		switch valType := measurementSentinelsFields[key]; valType {
		case "float":
			val, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", key, err)
			}
			fields[key] = val
		case "integer":
			val, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", key, err)
			}
			fields[key] = val
		case "string":
			fields[key] = val
		}
	}

	return tags, fields, nil
}

// converts `sentinel replicas <name>` output to tags and fields
func convertSentinelReplicaOutput(
	globalTags map[string]string,
	masterName string,
	replica map[string]string,
) (map[string]string, map[string]interface{}, error) {
	tags := globalTags

	tags["replica_ip"] = replica["ip"]
	tags["replica_port"] = replica["port"]
	tags["master"] = masterName

	fields := make(map[string]interface{})

	for key, val := range replica {
		key = strings.Replace(key, "-", "_", -1)

		switch valType := measurementReplicasFields[key]; valType {
		case "float":
			val, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", key, err)
			}
			fields[key] = val
		case "integer":
			val, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", key, err)
			}
			fields[key] = val
		case "string":
			fields[key] = val
		}
	}

	return tags, fields, nil
}

// convertSentinelInfoOutput parses `INFO` command output
// Largely copied from the Redis input plugin's gatherInfoOutput()
func convertSentinelInfoOutput(
	globalTags map[string]string,
	rdr io.Reader,
) (map[string]string, map[string]interface{}, error) {
	scanner := bufio.NewScanner(rdr)
	fields := make(map[string]interface{})

	tags := globalTags

	var section string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Redis denotes configuration sections with a hashtag
		// This comes in handy when we want to ensure we're processing the correct configuration option
		// For example, when we are renaming fields before sending them to the accumulator
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
				uptimeInSeconds, uptimeParseErr := strconv.ParseInt(parts[1], 10, 64)
				if uptimeParseErr != nil {
					return nil, nil, fmt.Errorf("failed parsing field uptime_in_seconds: %v", uptimeParseErr)
				}
				fields["uptime_ns"] = int64(time.Duration(uptimeInSeconds) * time.Second)

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
			val, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", metric, err)
			}
			fields[metric] = val
		case "integer":
			val, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("failed parsing field %v: %v", metric, err)
			}
			fields[metric] = val
		case "string":
			fields[metric] = val
		}
	}

	return tags, fields, nil
}
