package redistimeseries

import (
	"github.com/go-redis/redis/v7"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## The address of the RedisTimeSeries server.
  address = "127.0.0.1:6379"
  database = 0

  ## Redis ACL username
  # username = ""
  ## password to login Redis
  # password = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

type RedisTimeSeries struct {
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database int    `toml:"database"`
	tls.ClientConfig
	client *redis.Client
}

func (r *RedisTimeSeries) Connect() error {
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
		tags := m.Tags()

		name := m.Name()
		for fieldName, value := range m.Fields() {
			key := name + "_" + fieldName

			var addSlice []interface{}
			addSlice = append(addSlice, "TS.ADD")
			addSlice = append(addSlice, key)
			addSlice = append(addSlice, now)
			addSlice = append(addSlice, value)
			for k, v := range tags {
				addSlice = append(addSlice, k)
				addSlice = append(addSlice, v)
			}

			err := r.client.Do(addSlice...).Err() //
			if err != nil {
				var createSlice []interface{}
				createSlice = append(createSlice, "TS.CREATE")
				createSlice = append(createSlice, key)
				for k, v := range tags {
					createSlice = append(createSlice, k)
					createSlice = append(createSlice, v)
				}
				err2 := r.client.Do(createSlice...).Err() //Create a new timeseries with new labels
				if err2 != nil {
					return err
				}
				err3 := r.client.Do(addSlice...).Err() // Attempt the add again
				if err3 != nil {
					return err
				}
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
