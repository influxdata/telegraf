package redistimeseries

import (
	"github.com/go-redis/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## The address of the RedisTimeSeries server.
  addr = "127.0.0.1:6379"

  ## password to login Redis
  # password = ""
`

type RedisTimeSeries struct {
	Addr            string "localhost:6379"
	Password        string ""
	Client    		*redis.Client
}

func (i *RedisTimeSeries) Connect() error {

	client := redis.NewClient(&redis.Options{
		Addr:     i.Addr,
		Password: i.Password,
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	i.Client = client
	return err
}

func (i *RedisTimeSeries) Close() error {
	return i.Client.Close()
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
			_, err := i.Client.Do("TS.ADD", key, now, value).Result()
			if err  != nil {
				// TODO add tags 
				_, err2 := i.Client.Do("TS.CREATE", key).Result()
				if err2  != nil {
					return err
				}
				_, err3 := i.Client.Do("TS.ADD", key, now, value).Result()
				if err3  != nil {
					return err
				}
			}
		}
		
	}
	return nil
}

func init() {
	outputs.Add("RedisTimeSeries", func() telegraf.Output {
		return &RedisTimeSeries{
		}
	})
}
