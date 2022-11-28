//go:generate ../../../tools/readme_config_includer/generator
package redistimeseries

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v7"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

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
		return errors.New("redis address must be specified")
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
	return "Plugin for sending metrics to RedisTimeSeries"
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
			tags = append(tags, k, v)
		}

		for fieldName, value := range m.Fields() {
			key := name + "_" + fieldName

			addSlice := []interface{}{"TS.ADD", key, now, value}
			addSlice = append(addSlice, tags...)

			if err := r.client.Do(addSlice...).Err(); err != nil {
				return fmt.Errorf("adding sample failed: %v", err)
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
