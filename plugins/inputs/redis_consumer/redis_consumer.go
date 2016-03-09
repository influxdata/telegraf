package redis_consumer

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/garyburd/redigo/redis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type RedisConsumer struct {
	Servers  []string
	Channels []string

	parser parsers.Parser

	pubsubs []redis.PubSubConn
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

  ##  List of channels to listen to. Selecting channels using Redis'
  ##  pattern-matching is allowed, e.g.:
  ##	channels = ["telegraf:*", "app_[1-3]"]
  ##
  ##  See http://redis.io/topics/pubsub#pattern-matching-subscriptions for
  ##  more info.
  channels = ["telegraf"]

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

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

	// Add default Redis server when no servers are configured.
	if len(r.Servers) == 0 {
		r.Servers = append(r.Servers, "tcp://localhost:6379")
	}

	// Create connections to every configured server.
	for _, server := range r.Servers {

		pubsub, err := r.createPubSub(server)

		if err != nil {
			return fmt.Errorf("Unable to connect to Redis server '%s': %s", server, err)
		}

		r.pubsubs = append(r.pubsubs, pubsub)
	}

	// Subscribe to channels on every server and start listening for messages.
	for _, pubsub := range r.pubsubs {
		for _, channel := range r.Channels {
			// Use PSUBSCRIBE when channels contains glob pattern.

			var err error
			if strings.IndexAny(channel, "*?[") >= 0 {
				err = pubsub.PSubscribe(channel)
			} else {
				err = pubsub.Subscribe(channel)
			}

			if err != nil {
				return fmt.Errorf("Could not (p)subscribe to channel '%s': %s.", channel, err)
			}
		}

		go r.listen(pubsub)

	}

	return nil
}

// Create
func (r *RedisConsumer) createPubSub(server string) (redis.PubSubConn, error) {
	var pubsub redis.PubSubConn
	u, err := url.Parse(server)

	if err != nil {
		return pubsub, fmt.Errorf("Unable to parse to address '%s': %s", server, err)
	}

	if u.Scheme == "" {
		// fallback to simple string based address (i.e. "10.0.0.1:10000")
		u.Scheme = "tcp"
		u.Host = server
		u.Path = ""
	}

	con, err := redis.Dial(u.Scheme, u.Host)

	if err != nil {
		return pubsub, fmt.Errorf("Could connect to Redis: %s", err)
	}

	return redis.PubSubConn{con}, nil

}

func (r *RedisConsumer) listen(pubsub redis.PubSubConn) error {
	for {
		switch v := pubsub.Receive().(type) {
		case redis.Message:
			r.processMessage(v)
		case error:
			return v
		}
	}
}

func (r *RedisConsumer) processMessage(msg redis.Message) error {
	metrics, err := r.parser.Parse(msg.Data)

	if err != nil {
		return err

	}

	for _, metric := range metrics {
		r.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}
	return nil
}

func (r *RedisConsumer) Stop() {
	for _, pubsub := range r.pubsubs {
		pubsub.Close()
	}

}

func init() {
	inputs.Add("redis_consumer", func() telegraf.Input {
		return &RedisConsumer{}
	})
}
