//go:generate ../../../tools/readme_config_includer/generator
package redistimeseries

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type RedisTimeSeries struct {
	Address             string           `toml:"address"`
	Username            config.Secret    `toml:"username"`
	Password            config.Secret    `toml:"password"`
	Database            int              `toml:"database"`
	ConvertStringFields bool             `toml:"convert_string_fields"`
	Timeout             config.Duration  `toml:"timeout"`
	Expire              *config.Duration `toml:"expire"`
	Log                 telegraf.Logger  `toml:"-"`
	tls.ClientConfig
	client *redis.Client
}

func (r *RedisTimeSeries) Init() error {
	if r.Expire != nil && time.Duration(*r.Expire) <= 0 {
		return errors.New("expire must be a positive duration")
	}
	return nil
}

func (r *RedisTimeSeries) Connect() error {
	if r.Address == "" {
		return errors.New("redis address must be specified")
	}

	username, err := r.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	defer username.Destroy()

	password, err := r.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	defer password.Destroy()

	r.client = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Username: username.String(),
		Password: password.String(),
		DB:       r.Database,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout))
	defer cancel()
	return r.client.Ping(ctx).Err()
}

func (r *RedisTimeSeries) Close() error {
	return r.client.Close()
}

func (*RedisTimeSeries) Description() string {
	return "Plugin for sending metrics to RedisTimeSeries"
}

func (*RedisTimeSeries) SampleConfig() string {
	return sampleConfig
}
func (r *RedisTimeSeries) Write(metrics []telegraf.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout))
	defer cancel()

	for _, m := range metrics {
		for name, fv := range m.Fields() {
			key := m.Name() + "_" + name

			var value float64
			switch v := fv.(type) {
			case float64:
				value = v
			case string:
				if !r.ConvertStringFields {
					r.Log.Debugf("Dropping string field %q of metric %q", name, m.Name())
					continue
				}
				var err error
				value, err = strconv.ParseFloat(v, 64)
				if err != nil {
					r.Log.Debugf("Converting string field %q of metric %q failed: %v", name, m.Name(), err)
					continue
				}
			default:
				var err error
				value, err = internal.ToFloat64(v)
				if err != nil {
					r.Log.Errorf("Converting field %q (%T) of metric %q failed: %v", name, v, m.Name(), err)
					continue
				}
			}

			if r.Expire != nil {
				pipe := r.client.Pipeline()
				pipe.TSAddWithArgs(ctx, key, m.Time().UnixMilli(), value, &redis.TSOptions{Labels: m.Tags()})
				pipe.Expire(ctx, key, time.Duration(*r.Expire))
				if _, err := pipe.Exec(ctx); err != nil {
					return fmt.Errorf("writing sample %q with expiry failed: %w", key, err)
				}
			} else {
				resp := r.client.TSAddWithArgs(ctx, key, m.Time().UnixMilli(), value, &redis.TSOptions{Labels: m.Tags()})
				if err := resp.Err(); err != nil {
					return fmt.Errorf("adding sample %q failed: %w", key, err)
				}
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("redistimeseries", func() telegraf.Output {
		return &RedisTimeSeries{
			ConvertStringFields: true,
			Timeout:             config.Duration(10 * time.Second),
		}
	})
}
