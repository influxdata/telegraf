package redis

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"gopkg.in/redis.v4"
)

//use the redis service LIST struct as telegraf output
type RedisOutput struct {
	Server      string            `toml:"server"`
	Password    string            `toml:"password"`
	IdleTimeout internal.Duration `toml:"idle_timeout"`
	Timeout     internal.Duration `toml:"timeout"`
	Queue       string            `toml:"queue_name"`

	server     *redis.Client
	serializer serializers.Serializer
}

var sampleConfig = `
  ## redis service listen addr:port, default 127.0.0.1
  # server = "127.0.0.1:6379"
  ## redis service login password
  # password = ""
  ## redis close connections after remaining idle for this duration.
  ## if the value is zero, then idleconnections are not closed.
  ## shoud set the timeout to a value lessthan the redis server's timeout.
  # idle_timeout = "1s" 
  ## specifies the timeout for reading/writing a single command.
  # timeout = "1s"
  ## redis list name, defalut telegraf/output
  # queue_name = "telegraf/output"
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
`

func (p *RedisOutput) SetSerializer(serializer serializers.Serializer) {
	p.serializer = serializer
}

func (p *RedisOutput) SampleConfig() string {
	return sampleConfig
}

func (p *RedisOutput) Description() string {
	return "Configuration for the redis output"
}

func (p *RedisOutput) Connect() error {
	if p.Server == "" {
		p.Server = "localhost:6379"
	}

	if p.Queue == "" {
		p.Queue = "telegraf/output"
	}

	client := redis.NewClient(
		&redis.Options{
			Network:      "tcp",
			Addr:         p.Server,
			IdleTimeout:  p.IdleTimeout.Duration,
			ReadTimeout:  p.Timeout.Duration,
			WriteTimeout: p.Timeout.Duration,
		})

	if _, err := client.Ping().Result(); err != nil {
		return fmt.Errorf("failed to connect redis: %s", err)
	}

	p.server = client

	return nil
}

func (p *RedisOutput) Close() error {
	if p.server != nil {
		return p.server.Close()
	}
	return nil
}

func (p *RedisOutput) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	pipe := p.server.Pipeline()
	for _, metric := range metrics {
		b, err := p.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}

		pipe.LPush(p.Queue, string(b))
	}
	_, err := pipe.Exec()
	if err != nil {
		return fmt.Errorf("failed to write metric: %s", err)
	}

	return nil
}

func init() {
	outputs.Add("redis", func() telegraf.Output {
		return &RedisOutput{}
	})
}
