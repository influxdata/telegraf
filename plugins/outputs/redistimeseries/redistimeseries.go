package redistimeseries

import (
	"github.com/go-redis/redis/v7"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## The address of the RedisTimeSeries server.
  address = "127.0.0.1:6379"

  # Redis ACL username
  # username = ""
  ## password to login Redis
  # password = ""
`

type RedisTimeSeries struct {
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	client   *redis.Client
}

func (i *RedisTimeSeries) Connect() error {

	client := redis.NewClient(&redis.Options{
		Addr:     i.Address,
		Password: i.Password,
		Username: i.Username,
		DB:       0, // use default DB
	})

	err := client.Ping().Err()
	i.client = client
	return err
}

func (i *RedisTimeSeries) Close() error {
	return i.client.Close()
}

func (i *RedisTimeSeries) Description() string {
	return "Configuration for sending metrics to RedisTimeSeries"
}

func (i *RedisTimeSeries) SampleConfig() string {
	return sampleConfig
}
func (i *RedisTimeSeries) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}

	for _, m := range metrics {
		now := m.Time().UnixNano() / 1000000000
		//		tags := m.Tags() TODO add support for tags
		name := m.Name()

		for fieldName, value := range m.Fields() {
			key := name + "_" + fieldName
			err := i.client.Do("TS.ADD", key, now, value).Err()
			if err != nil {
				// TODO add tags
				err2 := i.client.Do("TS.CREATE", key).Err()
				if err2 != nil {
					return err
				}
				err3 := i.client.Do("TS.ADD", key, now, value).Err()
				if err3 != nil {
					return err
				}
			}
		}

	}
	return nil
}

func init() {
	outputs.Add("RedisTimeSeries", func() telegraf.Output {
		return &RedisTimeSeries{}
	})
}
