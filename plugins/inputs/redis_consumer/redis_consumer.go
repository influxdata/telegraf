package redis_consumer

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	//"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type RedisConsumer struct {
	Servers  []string
	Channels []string
	parser   parsers.Parser

	con net.Conn
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

	for _, channel := range r.Channels {
		log.Printf("Channel %s", channel)
	}


   	// hasMeta reports whether path contains any of the magic characters
   	// recognized by Match.
   	func hasMeta(path string) bool {
   		// TODO(niemeyer): Should other magic characters be added here?
   		return strings.IndexAny(path, "*?[") >= 0
   	}


	con, err := net.Dial("tcp", "localhost:6379")
	r.con = con

	if err != nil {
		return fmt.Errorf("Could connect to Redis: %s", err)
	}
	log.Printf("Connected to Redis.")

	_, err = r.con.Write([]byte("SUBSCRIBE telegraf\r\n"))
	if err != nil {
		fmt.Errorf("Could not SUBSCRIBE to channels: %s", err)
		return err
	}
	log.Printf("Subscribed to channels.")

	// Redis sends confirmation message, ignore this.
	r.con.Read(make([]byte, 512))

	go r.listen()

	return nil
}

func (r *RedisConsumer) listen() error {

	for {
		buf := make([]byte, 512)
		n, err := r.con.Read(buf)
		if err != nil {
			return fmt.Errorf("Something went wrong while reading channel: %s", err)
		}

		msg := string(buf[:n])
		arr := strings.Split(msg, "\r\n")

		//r.acc.Add("redis_consumer", 1, make(map[string]string), time.Now())
		log.Printf("Length of list %s", len(arr))

		log.Printf("Received %s", msg)
		log.Printf("Received %s", arr)
	}

	return nil
}

func (r *RedisConsumer) Stop() {
	r.con.Close()
}

func init() {
	inputs.Add("redis_consumer", func() telegraf.Input {
		return &RedisConsumer{}
	})
}
