package redis_sentinel

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type RedisSentinel struct {
	Servers  []string
	Password string
	tls.ClientConfig

	clients     []*RedisSentinelClient
	initialized bool
}

type RedisSentinelClient struct {
	sentinel *redis.SentinelClient
	tags     map[string]string
}

const measurementMasters = "redis_sentinel_masters"
const measurementSentinel = "redis_sentinel"
const measurementSentinels = "redis_sentinels"
const measurementReplicas = "redis_replicas"

// Rename fields (old : new)
var Tracking = map[string]string{
	"connected_clients": "clients",
}

func init() {
	inputs.Add("redis_sentinel", func() telegraf.Input {
		return &RedisSentinel{}
	})
}

func (r *RedisSentinelClient) BaseTags() map[string]string {
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

func (r *RedisSentinel) init(acc telegraf.Accumulator) error {
	if r.initialized {
		return nil
	}

	if len(r.Servers) == 0 {
		r.Servers = []string{"tcp://localhost:26379"}
	}

	r.clients = make([]*RedisSentinelClient, len(r.Servers))

	tlsConfig, err := r.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	for i, serv := range r.Servers {
		if !strings.HasPrefix(serv, "tcp://") && !strings.HasPrefix(serv, "unix://") {
			log.Printf("W! [inputs.redis_sentinel]: server URL found without scheme; please update your configuration file")
			serv = "tcp://" + serv
		}

		u, err := url.Parse(serv)
		if err != nil {
			acc.AddError(fmt.Errorf("Unable to parse to address %q: %v", serv, err))
			continue
		}

		password := ""
		if u.User != nil {
			pw, ok := u.User.Password()
			if ok {
				password = pw
			}
		} else {
			if len(r.Password) > 0 {
				password = r.Password
			}
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

	r.initialized = true
	return nil
}

// Redis list format has string key/values adjacent, so convert to a map for easier use
func toMap(vals []interface{}) map[string]string {
	m := make(map[string]string)

	for idx := 0; idx < len(vals); idx += 2 {
		m[vals[idx].(string)] = vals[idx+1].(string)
	}

	return m
}

// Reads stats from all configured servers accumulates stats.
// Returns one of the errors encountered while gather stats (if any).
func (r *RedisSentinel) Gather(acc telegraf.Accumulator) error {
	if !r.initialized {
		if err := r.init(acc); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	for _, client := range r.clients {
		wg.Add(1)

		go func(client *RedisSentinelClient, acc telegraf.Accumulator) {
			defer wg.Done()

			// First check all masters this sentinel is monitoring
			mastersCmd := redis.NewSliceCmd("sentinel", "masters")
			if smErr := client.sentinel.Process(mastersCmd); smErr != nil {
				acc.AddError(smErr)
				return
			}

			masters, mastersErr := mastersCmd.Result()
			if mastersErr != nil {
				acc.AddError(mastersErr)
				return
			}

			for _, master := range masters {
				m := toMap(master.([]interface{}))

				quorumCmd := redis.NewStringCmd("sentinel", "ckquorum", m["name"])
				if qErr := client.sentinel.Process(quorumCmd); qErr != nil {
					acc.AddError(qErr)
					return
				}

				_, quorumErr := quorumCmd.Result()

				sentinelMastersTags, sentinelMastersFields := convertSentinelMastersOutput(client.BaseTags(), m, quorumErr)
				acc.AddFields(measurementMasters, sentinelMastersFields, sentinelMastersTags)

				// ------------------------------------------------------------

				// Check other Sentinels
				sentinelsCmd := redis.NewSliceCmd("sentinel", "sentinels", m["name"])
				if ssErr := client.sentinel.Process(sentinelsCmd); ssErr != nil {
					acc.AddError(ssErr)
					return
				}

				sentinels, sentinelsErr := sentinelsCmd.Result()
				if sentinelsErr != nil {
					acc.AddError(sentinelsErr)
					return
				}

				for _, sentinel := range sentinels {
					sm := toMap(sentinel.([]interface{}))

					sentinelTags, sentinelFields := convertSentinelSentinelsOutput(client.BaseTags(), m["name"], sm)
					acc.AddFields(measurementSentinels, sentinelFields, sentinelTags)
				}

				// ------------------------------------------------------------

				// Check other Replicas
				replicasCmd := redis.NewSliceCmd("sentinel", "replicas", m["name"])
				if srErr := client.sentinel.Process(replicasCmd); srErr != nil {
					acc.AddError(srErr)
					return
				}

				replicas, replicasErr := replicasCmd.Result()
				if replicasErr != nil {
					acc.AddError(replicasErr)
					return
				}

				for _, replica := range replicas {
					rm := toMap(replica.([]interface{}))

					replicaTags, replicaFields := convertSentinelReplicaOutput(client.BaseTags(), m["name"], rm)
					acc.AddFields(measurementReplicas, replicaFields, replicaTags)
				}

				// ------------------------------------------------------------

				// Get INFO
				infoCmd := redis.NewStringCmd("info", "all")
				if iErr := client.sentinel.Process(infoCmd); iErr != nil {
					acc.AddError(iErr)
					return
				}

				info, infoErr := infoCmd.Result()
				if infoErr != nil {
					acc.AddError(infoErr)
					return
				}

				rdr := strings.NewReader(info)
				infoTags, infoFields := convertSentinelInfoOutput(acc, client.BaseTags(), rdr)

				acc.AddFields(measurementSentinel, infoFields, infoTags)
			}
		}(client, acc)
	}

	wg.Wait()

	return nil
}

// converts `sentinel masters <name>` output to tags and fields
func convertSentinelMastersOutput(
	global_tags map[string]string,
	master map[string]string,
	quorumErr error,
) (map[string]string, map[string]interface{}) {

	tags := make(map[string]string)
	for k, v := range global_tags {
		tags[k] = v
	}

	tags["master_name"] = master["name"]

	fields := make(map[string]interface{})

	fields["has-quorum"] = 0
	if quorumErr == nil {
		fields["has-quorum"] = 1
	}

	for key, val := range master {
		if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
			fields[key] = ival
			continue
		}

		if fval, err := strconv.ParseFloat(val, 64); err == nil {
			fields[key] = fval
			continue
		}

		fields[key] = val
	}

	return tags, fields
}

// converts `sentinel sentinels <name>` output to tags and fields
func convertSentinelSentinelsOutput(
	global_tags map[string]string,
	masterName string,
	sentinelMaster map[string]string,
) (map[string]string, map[string]interface{}) {
	tags := make(map[string]string)
	for k, v := range global_tags {
		tags[k] = v
	}

	tags["sentinel_ip"] = sentinelMaster["ip"]
	tags["sentinel_port"] = sentinelMaster["port"]
	tags["master_name"] = masterName

	fields := make(map[string]interface{})

	for key, val := range sentinelMaster {
		if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
			fields[key] = ival
			continue
		}

		if fval, err := strconv.ParseFloat(val, 64); err == nil {
			fields[key] = fval
			continue
		}

		fields[key] = val
	}

	return tags, fields
}

// converts `sentinel replicas <name>` output to tags and fields
func convertSentinelReplicaOutput(
	global_tags map[string]string,
	masterName string,
	replica map[string]string,
) (map[string]string, map[string]interface{}) {

	tags := make(map[string]string)
	for k, v := range global_tags {
		tags[k] = v
	}

	tags["replica_ip"] = replica["ip"]
	tags["replica_port"] = replica["port"]
	tags["master_name"] = masterName

	fields := make(map[string]interface{})

	for key, val := range replica {
		if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
			fields[key] = ival
			continue
		}

		if fval, err := strconv.ParseFloat(val, 64); err == nil {
			fields[key] = fval
			continue
		}

		fields[key] = val
	}

	return tags, fields
}

// convertSentinelInfoOutput parses `INFO` command output
// Largely copied from the Redis input plugin's gatherInfoOutput()
func convertSentinelInfoOutput(
	acc telegraf.Accumulator,
	global_tags map[string]string,
	rdr io.Reader,
) (map[string]string, map[string]interface{}) {
	scanner := bufio.NewScanner(rdr)
	fields := make(map[string]interface{})

	tags := make(map[string]string)
	for k, v := range global_tags {
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
			if name != "lru_clock" && name != "uptime_in_seconds" && name != "redis_version" {
				continue
			}

			// Rename and convert to nanoseconds
			if name == "uptime_in_seconds" {
				uptimeInSeconds, uptimeParseErr := strconv.ParseInt(parts[1], 10, 64)
				if uptimeParseErr == nil {
					fields["uptime_ns"] = int64(time.Duration(uptimeInSeconds) * time.Second)
					continue
				} else {
					acc.AddError(uptimeParseErr)
				}
			}
		}

		if strings.HasSuffix(name, "_human") {
			continue
		}

		// This data (master0, master1, etc) is already captured by `sentinel masters`,
		// so skip
		if strings.HasPrefix(name, "master") {
			continue
		}

		metric, ok := Tracking[name]
		if !ok {
			metric = name
		}

		val := strings.TrimSpace(parts[1])

		if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
			fields[metric] = ival
			continue
		}

		if fval, err := strconv.ParseFloat(val, 64); err == nil {
			fields[metric] = fval
			continue
		}

		fields[metric] = val
	}

	return tags, fields
}
