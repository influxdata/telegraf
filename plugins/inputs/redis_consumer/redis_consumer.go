package redis_consumer

import (
	"fmt"
	"log"
	"regexp"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"gopkg.in/redis.v4"
)

// RedisConsumer represents a redis consumer for Telegraf
type RedisConsumer struct {
	Servers  []string
	Channels []string

	clients   []*redis.Client
	acc       telegraf.Accumulator
	accumLock sync.Mutex
	parser    parsers.Parser
	pubsubs   []*redis.PubSub
	finish    chan struct{}
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

func parseChannels(channels []string) (subs, psubs []string, err error) {
	err = nil
	subs = make([]string, 0)
	psubs = make([]string, 0)

	for _, channel := range channels {
		if matched, fail := regexp.MatchString(`[^\\][\[|\(|\*]`, channel); fail != nil {
			err = fmt.Errorf("Could not parse %s : %v", channel, fail)
			return
		} else if matched {
			psubs = append(psubs, channel)
		} else {
			subs = append(subs, channel)
		}
	}
	return
}

func createClient(server string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: server,
	})

	if _, err := client.Ping().Result(); err != nil {
		return client, fmt.Errorf("Unable to ping redis server %s : %v", server, err)
	}

	return client, nil
}

// SetParser allows the consumer to accept multiple data formats
func (r *RedisConsumer) SetParser(parser parsers.Parser) {
	r.parser = parser
}

// SampleConfig provides a sample configuration for the redis consumer
func (r *RedisConsumer) SampleConfig() string {
	return sampleConfig
}

// Description provides a description of the consumer
func (r *RedisConsumer) Description() string {
	return "Reads metrics from Redis channels"
}

// Gather noop for the redis consumer
func (r *RedisConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

// Start starts fetching data from the redis server
func (r *RedisConsumer) Start(acc telegraf.Accumulator) error {
	r.accumLock.Lock()
	defer r.accumLock.Unlock()
	r.acc = acc

	if len(r.Servers) == 0 {
		r.Servers = append(r.Servers, "tcp://localhost:6379")
	}

	// Verify every server can be connected
	for _, server := range r.Servers {
		var client *redis.Client
		var err error

		if client, err = createClient(server); err != nil {
			return fmt.Errorf("Unable to crate redis server %s : %v", server, err)
		}

		r.clients = append(r.clients, client)
	}

	// Verify all subscriptions can be made
	var err error
	if r.pubsubs, err = r.createSubscriptions(); err != nil {
		return err
	}

	r.finish = make(chan struct{})
	// Start listening
	for _, pubsub := range r.pubsubs {
		go r.listen(pubsub)
	}

	return nil
}

func (r *RedisConsumer) createSubscriptions() ([]*redis.PubSub, error) {
	subs, psubs, err := parseChannels(r.Channels)
	if err != nil {
		return nil, err
	}
	pubsubs := []*redis.PubSub{}

	for _, c := range r.clients {
		var s, ps *redis.PubSub

		if len(subs) > 0 {
			s, err = c.Subscribe(subs...)
			if err != nil {
				return nil, fmt.Errorf("Error during subscription creation: %v", err)
			}
			pubsubs = append(pubsubs, s)
		}

		if len(psubs) > 0 {
			ps, err = c.PSubscribe(psubs...)
			if err != nil {
				return nil, fmt.Errorf("Error during psubscription creation: %v", err)
			}
			pubsubs = append(pubsubs, ps)
		}
	}
	return pubsubs, nil
}

func (r *RedisConsumer) listen(pubsub *redis.PubSub) {
	for {
		msg, err := pubsub.ReceiveMessage()

		// Check if the consumer is finishing
		if err != nil {
			select {
			case <-r.finish:
				pubsub.Close()
				return
			default:
				// Nothing todo
			}
		}
		metrics, merr := r.parser.Parse([]byte(msg.Payload))

		if merr != nil {
			log.Printf("Redis Parse Error.\n\tMessage: %s\n\tError: %v", msg.Payload, merr)
			continue
		}

		for _, metric := range metrics {
			r.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
	}
}

// Stop stops fetching data from the redis server
func (r *RedisConsumer) Stop() error {
	r.accumLock.Lock()
	defer r.accumLock.Unlock()

	close(r.finish)
	for _, client := range r.clients {
		if err := client.Close(); err != nil {
			return fmt.Errorf("Error closing redis server: %v", err)
		}
	}

	return nil
}

func init() {
	inputs.Add("redis_consumer", func() telegraf.Input {
		return &RedisConsumer{}
	})
}
