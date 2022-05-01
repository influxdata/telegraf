package redistimeseries

import (
	"fmt"

	"github.com/go-redis/redis/v7"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## The address of the RedisTimeSeries server.
  address = "127.0.0.1:6379"

  ## Redis ACL credentials
  # username = ""
  # password = ""
  # database = 0

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
`

type RedisTimeSeries struct {
	Address  string          `toml:"address"`
	Username string          `toml:"username"`
	Password string          `toml:"password"`
	Database int             `toml:"database"`
	Log      telegraf.Logger `toml:"-"`
	tls.ClientConfig
	client *redis.Client
}

func (r *RedisTimeSeries) Connect() error {
	if r.Address == "" {
		return fmt.Errorf("Redis address must be specified")
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		Username: r.Username,
		DB:       r.Database,
	})
	return r.client.Ping().Err()
}

func (r *RedisTimeSeries) Close() error {
	return r.client.Close()
}

func (r *RedisTimeSeries) Description() string {
	return "Configuration for sending metrics to RedisTimeSeries"
}

func (r *RedisTimeSeries) SampleConfig() string {
	return sampleConfig
}
func (r *RedisTimeSeries) Write(metrics []telegraf.Metric) error {
	for _, m := range metrics {
		now := m.Time().UnixNano() / 1000000 // in milliseconds
		name := m.Name()

		var tags []interface{}
		for k, v := range m.Tags() {
			tags = append(tags, k)
			tags = append(tags, v)
		}

		for fieldName, value := range m.Fields() {
			key := name + "_" + fieldName

			var addSlice []interface{}
			addSlice = append(addSlice, "TS.ADD")
			addSlice = append(addSlice, key)
			addSlice = append(addSlice, now)
			addSlice = append(addSlice, value)
			addSlice = append(addSlice, tags...)

			if err := r.client.Do(addSlice...).Err(); err != nil {
				return fmt.Errorf("reattempting adding sample failed: %v", err)
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("redistimeseries", func() telegraf.Output {
		return &RedisTimeSeries{}
	})
}
