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

		var address string
		tags := map[string]string{}
		
		switch u.Scheme {
		case "tcp":
			address = u.Host
			tags["source"] = u.Hostname()
			tags["port"] = u.Port()
		case "unix":
			address = u.Path
			tags["socket"] = u.Path
		default:
			return fmt.Errorf("invalid scheme %q, expected tcp or unix", u.Scheme)
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

func castFieldValue(value string, fieldType configFieldType) (interface{}, error) {
	var castedValue interface{}
	var err error

	switch fieldType {
	case configFieldTypeFloat:
		castedValue, err = strconv.ParseFloat(value, 64)
	case configFieldTypeInteger:
		castedValue, err = strconv.ParseInt(value, 10, 64)
	case configFieldTypeString:
		castedValue = value
	default:
		return nil, fmt.Errorf("unsupported field type %v", fieldType)
	}

	if err != nil {
		return nil, fmt.Errorf("casting value %v failed: %v", value, err)
	}

	return castedValue, nil
}

func prepareFieldValues(
	fields map[string]string,
	configFieldTypeMap map[string]configFieldType,
) (map[string]interface{}, error) {
	preparedFields := make(map[string]interface{})

	for key, val := range fields {
		key = strings.Replace(key, "-", "_", -1)

		valType, valTypeOk := configFieldTypeMap[key]

		if !valTypeOk {
			continue
		}

		castedVal, castedValErr := castFieldValue(val, valType)

		if castedValErr != nil {
			return nil, castedValErr
		}

		preparedFields[key] = castedVal
	}

	return preparedFields, nil
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
	if err := client.sentinel.Process(infoCmd); err != nil {
		acc.AddError(err)
		return
	}

	info, infoErr := infoCmd.Result()
	if infoErr != nil {
		acc.AddError(infoErr)
		return
	}

	rdr := strings.NewReader(info)
	infoTags, infoFields, err := convertSentinelInfoOutput(client.tags, rdr)
	if err != nil {
		acc.AddError(err)
		return
	}

	acc.AddFields(measurementSentinel, infoFields, infoTags)
}

func gatherMasterStats(acc telegraf.Accumulator, client *RedisSentinelClient) {
	mastersCmd := redis.NewSliceCmd("sentinel", "masters")
	if err := client.sentinel.Process(mastersCmd); err != nil {
		acc.AddError(err)
		return
	}

	masters, mastersErr := mastersCmd.Result()
	if mastersErr != nil {
		acc.AddError(mastersErr)
		return
	}

	for _, master := range masters {
		master, masterOk := master.([]interface{})
		if !masterOk {
			acc.AddError(fmt.Errorf("unable to process master response"))
			continue
		}

		m := toMap(master)

		masterName, masterNameOk := m["name"]
		if !masterNameOk {
			acc.AddError(fmt.Errorf("unable to resolve master name"))
			continue
		}

		quorumCmd := redis.NewStringCmd("sentinel", "ckquorum", masterName)

		quorumErr := client.sentinel.Process(quorumCmd)

		sentinelMastersTags, sentinelMastersFields, err := convertSentinelMastersOutput(client.tags, m, quorumErr)
		if err != nil {
			acc.AddError(err)
		} else {
			acc.AddFields(measurementMasters, sentinelMastersFields, sentinelMastersTags)
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
	if err := client.sentinel.Process(replicasCmd); err != nil {
		acc.AddError(err)
		return
	}

	replicas, replicasErr := replicasCmd.Result()
	if replicasErr != nil {
		acc.AddError(replicasErr)
		return
	}

	for _, replica := range replicas {
		replica, replicaOk := replica.([]interface{})
		if !replicaOk {
			acc.AddError(fmt.Errorf("unable to process replica response"))
			continue
		}

		rm := toMap(replica)
		replicaTags, replicaFields, err := convertSentinelReplicaOutput(client.tags, masterName, rm)
		if err != nil {
			acc.AddError(err)
			continue
		}

		acc.AddFields(measurementReplicas, replicaFields, replicaTags)
	}
}

func gatherSentinelStats(
	acc telegraf.Accumulator,
	client *RedisSentinelClient,
	masterName string,
) {
	sentinelsCmd := redis.NewSliceCmd("sentinel", "sentinels", masterName)
	if err := client.sentinel.Process(sentinelsCmd); err != nil {
		acc.AddError(err)
		return
	}

	sentinels, sentinelsErr := sentinelsCmd.Result()
	if sentinelsErr != nil {
		acc.AddError(sentinelsErr)
		return
	}

	for _, sentinel := range sentinels {
		sentinel, sentinelOk := sentinel.([]interface{})
		if !sentinelOk {
			acc.AddError(fmt.Errorf("unable to process sentinel response"))
			continue
		}

		sm := toMap(sentinel)
		sentinelTags, sentinelFields, err := convertSentinelSentinelsOutput(client.tags, masterName, sm)
		if err != nil {
			acc.AddError(err)
			continue
		}

		acc.AddFields(measurementSentinels, sentinelFields, sentinelTags)
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

	fields, fieldsErr := prepareFieldValues(master, measurementMastersFields)

	if fieldsErr != nil {
		return nil, nil, fieldsErr
	}

	fields["has_quorum"] = quorumErr == nil

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

	fields, fieldsErr := prepareFieldValues(sentinelMaster, measurementSentinelsFields)

	if fieldsErr != nil {
		return nil, nil, fieldsErr
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

	fields, fieldsErr := prepareFieldValues(replica, measurementReplicasFields)

	if fieldsErr != nil {
		return nil, nil, fieldsErr
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
	rawFields := make(map[string]string)

	tags := globalTags

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Redis denotes configuration sections with a hashtag
		// Example of the section header: # Clients
		if line[0] == '#' {
			// Nothing interesting here
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			// Not a valid configuration option
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		rawFields[key] = val
	}

	fields, fieldsErr := prepareFieldValues(rawFields, measurementSentinelFields)

	// Rename the field and convert it to nanoseconds
	fields["uptime_ns"] = int64(time.Duration(fields["uptime_in_seconds"].(int64)) * time.Second)
	delete(fields, "uptime_in_seconds")

	// Rename in order to match the "redis" input plugin
	fields["clients"] = fields["connected_clients"]
	delete(fields, "connected_clients")

	if fieldsErr != nil {
		return nil, nil, fieldsErr
	}

	return tags, fields, nil
}
