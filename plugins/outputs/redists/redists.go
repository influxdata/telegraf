package redists

import (
	"github.com/gomodule/redigo/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## The address of the RedisTS server.
  addr = "127.0.0.1:6379"

  ## password to login Redis
  # password = ""
`

type RedisTS struct {
	Addr            string "localhost:6379"
	Password        string ``
	Conn    		redis.Conn
}

func (i *RedisTS) Connect() error {

	conn, err := redis.Dial("tcp", i.Addr, redis.DialPassword(i.Password))
	i.Conn = conn
	return err
}

func (i *RedisTS) Close() error {
	return i.Conn.Close()
}

func (i *RedisTS) Description() string {
	return "Configuration for sending metrics to RedisTS"
}

func (i *RedisTS) SampleConfig() string {
	return sampleConfig
}
func (i *RedisTS) Write(metrics []telegraf.Metric) error {
	
	if len(metrics) == 0 {
		return nil
	}

	for _, m := range metrics {
		now := m.Time().UnixNano() / 1000000000
//		tags := m.Tags()
		name := m.Name()


		for fieldName, value := range m.Fields() {			
			key := name + "_" + fieldName
			_, err := i.Conn.Do("TS.ADD", key, now, value)
			if err  != nil {
				// TODO add tags 
				_, err2 := i.Conn.Do("TS.CREATE", key)
				if err2  != nil {
					return err
				}
				_, err3 := i.Conn.Do("TS.ADD", key, now, value)
				if err3  != nil {
					return err
				}
			}
		}
	}
	return nil
}

func init() {
	outputs.Add("RedisTS", func() telegraf.Output {
		return &RedisTS{
		}
	})
}
