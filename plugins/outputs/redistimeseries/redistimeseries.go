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

  # Redis ACL username
  # username = ""
  ## password to login Redis
  # password = ""
  database = 0
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
	client   *redis.Client
}

func (r *RedisTimeSeries) Connect() error {
	r.client = redis.NewClient(&redis.Options{
		Addr:	  r.Address,
		Password: r.Password,
		Username: r.Username,
		DB:	  r.Database,
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
	if len(metrics) == 0 {
		return nil
	}
	for _, m := range metrics {
		now := m.Time().Unix()
		//		tags := m.Tags() TODO add support for tags
		name := m.Name()
		for fieldName, value := range m.Fields() {
			key := name + "_" + fieldName
			err := r.client.Do("TS.ADD", key, now, value).Err()
			if err != nil {
				// TODO add tags
				err2 := r.client.Do("TS.CREATE", key).Err()
				if err2 != nil {
					return err
				}
				err3 := r.client.Do("TS.ADD", key, now, value).Err()
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
