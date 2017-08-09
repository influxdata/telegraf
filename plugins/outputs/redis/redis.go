package redis

import (
	"fmt"
	"sync"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/tidwall/sjson"
)

//use the redis service LIST struct as telegraf output
type RedisOutput struct {
	Addr     string `toml:"server_addr"`
	Password string `toml:"server_passwd"`
	Queue    string `toml:"queue_name"`

	server *redigo.Pool
	mu     sync.Mutex
}

var sampleConfig = `
  ## redis service listen addr:port, default 127.0.0.1
  # server_addr = "127.0.0.1:6379"
  ## redis service login password
  # server_passwd = ""
  ## redis list name, defalut telegraf/output
  # queue_name = "telegraf/output"
`

func (p *RedisOutput) SampleConfig() string {
	return sampleConfig
}

func (p *RedisOutput) Description() string {
	return "Configuration for the redis output"
}

func (p *RedisOutput) Connect() error {
	if p.Addr == "" {
		p.Addr = "localhost:6379"
	}

	if p.Queue == "" {
		p.Queue = "telegraf/output"
	}

	p.server = p.initRedis()
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

	conn := p.server.Get()
	defer conn.Close()

	for _, point := range metrics {
		var jstr string = `{}`
		var err error

		//Tags is on the first floor of the JSON string
		for k, v := range point.Tags() {
			if jstr, err = sjson.Set(jstr, k, v); err != nil {
				return fmt.Errorf("redis-output plugin sjson.Set tag error, %s", err)
			}
		}

		//Fields is placed in the JSON string body object
		for k, v := range point.Fields() {
			if jstr, err = sjson.Set(jstr, "body."+k, v); err != nil {
				return fmt.Errorf("redis-output plugin sjson.Set field error, %s", err)
			}
		}

		_, err = conn.Do("LPUSH", p.Queue, jstr)
		if err != nil {
			return fmt.Errorf("redis-output plugin LPUSH %s %s %s error, %s", p.Addr, p.Queue, jstr, err)
		}
	}

	return nil
}

func (p *RedisOutput) initRedis() *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", p.Addr)
			if err != nil {
				return nil, err
			}

			if p.Password != "" {
				_, err := c.Do("AUTH", p.Password)
				if err != nil {
					return nil, err
				}
			}

			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := redigo.String(c.Do("PING"))

			return err
		},
	}
}

func init() {
	outputs.Add("redis", func() telegraf.Output {
		return &RedisOutput{}
	})
}
