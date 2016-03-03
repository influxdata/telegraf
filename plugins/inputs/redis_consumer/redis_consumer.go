package redis_consumer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	//"time"

	"github.com/garyburd/redigo/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type RedisConsumer struct {
	Servers  []string
	Channels []string
	parser   parsers.Parser

	pubsub redis.PubSubConn
	sync.Mutex
	acc telegraf.Accumulator
}

var sampleConfig = `
  ## Specify servers via a url matching:
  ##  [protocol://][:password]@address[:port]
  ##  e.g.
  ##    tcp://localhost:6379
  ##    tcp://:password@192.168.99.100
  ##
  ## If no servers are specified, then localhost is used as the host.
  ## If no port is specified, 6379 is used
  servers = ["tcp://localhost:6379"]

  ##  List of channels to listen to. Selecting channels using pattern-matching
  ## is allowed.
  channels = []

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

const defaultPort = "6379"

func (r *RedisConsumer) SampleConfig() string {
	return sampleConfig
}

func (r *RedisConsumer) Description() string {
	return "Read metrics from Redis channels."
}

func (r *RedisConsumer) SetParser(parser parsers.Parser) {
	r.parser = parser
}

func (r *RedisConsumer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (r *RedisConsumer) Start(acc telegraf.Accumulator) error {
	r.acc = acc

	con, err := redis.Dial("tcp", "localhost:6379")

	if err != nil {
		return fmt.Errorf("Could connect to Redis: %s", err)
	}

	r.pubsub = redis.PubSubConn{con}

	for _, channel := range r.Channels {
		// Use PSUBSCRIBE when channels contains glob pattern.
		if strings.IndexAny(channel, "*?[") >= 0 {
			err = r.pubsub.PSubscribe(channel)
		} else {
			err = r.pubsub.Subscribe(channel)
		}

		if err != nil {
			return fmt.Errorf("Could not (p)subscribe to channel '%s': %s.", channel, err)
		}
	}

	log.Printf("Connected to Redis.")

	go r.listen()

	return nil
}

func (r *RedisConsumer) listen() error {
	for {
		switch v := r.pubsub.Receive().(type) {
		case redis.Message:
			r.processMessage(v)
			fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
		case error:
			return v
		}
	}
}

func (r *RedisConsumer) processMessage(msg redis.Message) error {
	return nil

}

func (r *RedisConsumer) Stop() {
	r.pubsub.Close()
}

func init() {
	inputs.Add("redis_consumer", func() telegraf.Input {
		return &RedisConsumer{}
	})
}
