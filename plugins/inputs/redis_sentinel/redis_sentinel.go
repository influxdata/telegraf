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

// Rename fields
var Tracking = map[string]string{
	"connected_clients": "clients",
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

type returnLastError struct {
	sync.RWMutex
	err error
}

func (e *returnLastError) Add(err error) {
	e.Lock()
	e.err = err
	e.Unlock()
}

func (e *returnLastError) Get() error {
	e.RLock()
	defer e.RUnlock()
	return e.err
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

	lastError := returnLastError{}

	for _, client := range r.clients {
		wg.Add(1)

		go func(client *RedisSentinelClient, acc telegraf.Accumulator) {
			defer wg.Done()

			mastersCmd := redis.NewSliceCmd("sentinel", "masters")
			if smErr := client.sentinel.Process(mastersCmd); smErr != nil {
				lastError.Add(smErr)
				return
			}

			masters, mastersErr := mastersCmd.Result()
			if mastersErr != nil {
				lastError.Add(mastersErr)
				return
			}

			for _, master := range masters {
				m := toMap(master.([]interface{}))

				masterTags := client.tags
				masterTags["master_name"] = m["name"]

				masterFields := make(map[string]interface{})

				for key, val := range m {
					// Try parsing as int
					if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
						masterFields[key] = ival
						continue
					}

					// Try parsing as a float
					if fval, err := strconv.ParseFloat(val, 64); err == nil {
						masterFields[key] = fval
						continue
					}

					// Treat it as a string
					masterFields[key] = val
				}

				quorumCmd := redis.NewStringCmd("sentinel", "ckquorum", m["name"])
				if qErr := client.sentinel.Process(quorumCmd); qErr != nil {
					lastError.Add(qErr)
					return
				}

				_, quorumErr := quorumCmd.Result()
				masterFields["has-quorum"] = false
				if quorumErr == nil {
					masterFields["has-quorum"] = true
				}

				acc.AddFields("redis_sentinel_masters", masterFields, masterTags)

				// ------------------------------------------------------------

				// Check other Sentinels
				sentinelTags := client.tags

				sentinelsCmd := redis.NewSliceCmd("sentinel", "sentinels", m["name"])
				if ssErr := client.sentinel.Process(sentinelsCmd); ssErr != nil {
					lastError.Add(ssErr)
					return
				}

				sentinels, sentinelsErr := sentinelsCmd.Result()
				if sentinelsErr != nil {
					lastError.Add(sentinelsErr)
					return
				}

				for _, sentinel := range sentinels {
					sm := toMap(sentinel.([]interface{}))
					sentinelTags["sentinel_ip"] = sm["ip"]
					sentinelTags["sentinel_port"] = sm["port"]
					sentinelTags["master_name"] = m["name"]
					sentinelFields := make(map[string]interface{})

					for key, val := range sm {
						if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
							sentinelFields[key] = ival
							continue
						}

						if fval, err := strconv.ParseFloat(val, 64); err == nil {
							sentinelFields[key] = fval
							continue
						}

						sentinelFields[key] = val
					}

					acc.AddFields("redis_sentinels", sentinelFields, sentinelTags)
				}

				// ------------------------------------------------------------

				// Check other Replicas
				replicaTags := client.tags

				replicasCmd := redis.NewSliceCmd("sentinel", "replicas", m["name"])
				if srErr := client.sentinel.Process(replicasCmd); srErr != nil {
					lastError.Add(srErr)
					return
				}

				replicas, replicasErr := replicasCmd.Result()
				if replicasErr != nil {
					lastError.Add(replicasErr)
					return
				}

				for _, replica := range replicas {
					rm := toMap(replica.([]interface{}))
					replicaTags["replica_ip"] = rm["ip"]
					replicaTags["replica_port"] = rm["port"]
					replicaTags["master_name"] = m["name"]
					replicaFields := make(map[string]interface{})

					for key, val := range rm {
						if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
							replicaFields[key] = ival
							continue
						}

						if fval, err := strconv.ParseFloat(val, 64); err == nil {
							replicaFields[key] = fval
							continue
						}

						replicaFields[key] = val
					}

					acc.AddFields("redis_replicas", replicaFields, replicaTags)
				}

				// ------------------------------------------------------------

				infoCmd := redis.NewStringCmd("info", "all")
				if iErr := client.sentinel.Process(infoCmd); iErr != nil {
					lastError.Add(iErr)
					return
				}

				info, infoErr := infoCmd.Result()
				if infoErr != nil {
					lastError.Add(infoErr)
					return
				}

				rdr := strings.NewReader(info)
				gatherSentinelInfoOutput(rdr, acc, client.BaseTags())
			}
		}(client, acc)
	}

	wg.Wait()

	return lastError.Get()
}

func init() {
	inputs.Add("redis_sentinel", func() telegraf.Input {
		return &RedisSentinel{}
	})
}

// gatherSentinelInfoOutput parses `INFO` command output
// Largely copied from the Redis input plugin's gatherInfoOutput()
func gatherSentinelInfoOutput(
	rdr io.Reader,
	acc telegraf.Accumulator,
	global_tags map[string]string,
) {
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

	acc.AddFields("redis_sentinel", fields, tags)
}
