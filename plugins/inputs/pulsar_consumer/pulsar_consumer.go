package pulsar_consumer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Comcast/pulsar-client-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var (
	defaultMaxUndeliveredMessages = 1000
)

type empty struct{}
type semaphore chan empty

type pulsarConsumer struct {
	consumer *pulsar.ManagedConsumer
	connPool *pulsar.ManagedClientPool

	parser parsers.Parser
	in     chan pulsar.Message
	errs   chan error
	acc    telegraf.TrackingAccumulator
	wg     sync.WaitGroup
	cancel context.CancelFunc

	tls.ClientConfig

	URL         string `toml:"url"`
	DialTimeout string `toml:"dial_timeout,omitempty"`
	RecvTimeout string `toml:"recv_timeout,omitempty"`
	recvTimeout time.Duration

	PingFrequency         string `toml:"ping_frequency,omitempty"`
	PingTimeout           string `toml:"ping_timeout,omitempty"`
	InitialReconnectDelay string `toml:"initial_reconnect_delay,omitempty"`
	MaxReconnectDelay     string `toml:"max_reconnect_delay,omitempty"`
	NewConsumerTimeout    string `toml:"new_consumer_timeout,omitempty"`

	Topic     string `toml:"topic"`
	Name      string `toml:"name"`
	Exclusive bool   `toml:"exclusive,omitempty"`
	QueueSize int    `toml:"queue_size,omitempty"`

	MaxUndeliveredMessages int `toml:"max_undelivered_messages,omitempty"`
}

var sampleConfig = `
  # URL to Pulsar cluster
  # If you use SSL, then the protocol should be "pulsar+ssl"
  url = "pulsar://localhost:6650"

  # Timeout while trying to connect
  dial_timeout = "15s"

  # Timeout while trying to receive message
  recv_timeout = "5s"

  # Topic of message
  topic = ""

  # Name of the consumer
  name = ""

  ## Is consumer exclusive
  # exclusive = false

  ## Queue size for this consumer
  # queue_size = 1000

  ## Path to certificates and key for TLS
  # tls_ca = ""
  # tls_cert = ""
  # tls_key = ""

  ## Other optionals
  ping_frequency = "1s"
  ping_timeout = "1s"
  initial_reconnect_delay = "3s"
  max_reconnect_delay = "10s"
  new_consumer_timeout = "10s"

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (p *pulsarConsumer) SampleConfig() string {
	return sampleConfig
}

func (p *pulsarConsumer) Description() string {
	return "Read metrics from Pulsar topic"
}

func (p *pulsarConsumer) SetParser(parser parsers.Parser) {
	p.parser = parser
}

func (p *pulsarConsumer) Start(acc telegraf.Accumulator) error {
	p.acc = acc.WithTracking(p.MaxUndeliveredMessages)
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	if p.connPool == nil && p.consumer == nil {
		p.in = make(chan pulsar.Message, p.MaxUndeliveredMessages)
		p.errs = make(chan error)

		var err error
		p.recvTimeout, err = time.ParseDuration(p.RecvTimeout)

		conf := pulsar.ManagedConsumerConfig{}
		conf.Addr = p.URL
		conf.Errs = p.errs
		conf.DialTimeout, err = time.ParseDuration(p.DialTimeout)
		conf.ConnectTimeout = conf.DialTimeout
		conf.PingFrequency, err = time.ParseDuration(p.PingFrequency)
		conf.PingTimeout, err = time.ParseDuration(p.PingTimeout)
		conf.InitialReconnectDelay, err = time.ParseDuration(p.InitialReconnectDelay)
		conf.MaxReconnectDelay, err = time.ParseDuration(p.MaxReconnectDelay)
		conf.NewConsumerTimeout, err = time.ParseDuration(p.NewConsumerTimeout)
		if p.TLSCA != "" && p.TLSCert != "" && p.TLSKey != "" {
			conf.TLSConfig, err = p.TLSConfig()
		}
		if err != nil {
			return err
		}

		conf.Topic = p.Topic
		conf.Name = p.Name
		conf.Exclusive = p.Exclusive
		conf.QueueSize = p.QueueSize

		p.connPool = pulsar.NewManagedClientPool()
		p.consumer = pulsar.NewManagedConsumer(p.connPool, conf)

		go func() {
			err = p.consumer.ReceiveAsync(ctx, p.in)
			if err != nil {
				p.errs <- err
			}
		}()
	}

	// Start the message reader
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		go p.receiver(ctx)
	}()

	log.Printf("I! Started the Pulsar consumer service, pulsar: %v, topic: %v, name: %v\n",
		p.URL, p.Topic, p.Name)

	return nil
}

func (p *pulsarConsumer) receiver(ctx context.Context) {
	sem := make(semaphore, p.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.acc.Delivered():
			<-sem
		case err := <-p.errs:
			p.acc.AddError(err)
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case err := <-p.errs:
				<-sem
				log.Printf("E! Unexpected error, pulsar: %v error: %v, topic: %v, name: %v\n",
					p.URL, err, p.Topic, p.Name)
				p.acc.AddError(err)
			case <-p.acc.Delivered():
				<-sem
				<-sem
			case msg := <-p.in:
				metrics, err := p.parser.Parse(msg.Payload)
				if err != nil {
					p.acc.AddError(fmt.Errorf("topic: %s, error: %s", msg.Topic, err.Error()))
					<-sem
					continue
				}
				p.acc.AddTrackingMetricGroup(metrics)

				timeoutCtx, cancel := context.WithTimeout(context.Background(), p.recvTimeout)
				defer cancel()
				err = p.consumer.Ack(timeoutCtx, msg)
				if err != nil {
					p.acc.AddError(fmt.Errorf("topic: %s, error: %s", msg.Topic, err.Error()))
					<-sem
					continue
				}
			}
		}
	}
}

func (p *pulsarConsumer) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *pulsarConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("pulsar_consumer", func() telegraf.Input {
		return &pulsarConsumer{
			URL:                    "pulsar://localhost:6650",
			DialTimeout:            "15s",
			RecvTimeout:            "15s",
			PingFrequency:          "30s",
			PingTimeout:            "15s",
			InitialReconnectDelay:  "15s",
			MaxReconnectDelay:      "15s",
			NewConsumerTimeout:     "15s",
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
			QueueSize:              defaultMaxUndeliveredMessages,
			Name:                   "telegraf-pulsar-consumer",
		}
	})
}
