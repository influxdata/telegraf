package pulsar

import (
	"context"
	"fmt"
	"time"

	plsr "github.com/Comcast/pulsar-client-go"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var sampleConfig = `
  ## URL to Pulsar cluster
  ## If you use SSL, then the protocol should be "pulsar+ssl"
  url = "pulsar://localhost:6650"

  ## Timeout while trying to connect
  dial_timeout = "15s"

  ## Timeout while trying to send message
  send_timeout = "5s"

  ## Topic of message
  topic = ""

  ## Name of the producer
  name = ""

  ## Path to certificates and key for TLS
  # tls_ca = ""
  # tls_cert = ""
  # tls_key = ""

  ## Other optionals
  # ping_frequency = "1s"
  # ping_timeout = "1s"
  # initial_reconnect_delay = "3s"
  # max_reconnect_delay = "10s"
  # new_producer_timeout = "10s"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

type Pulsar struct {
	serializer serializers.Serializer
	producer   *plsr.ManagedProducer
	tls.ClientConfig

	URL         string `toml:"url"`
	DialTimeout string `toml:"dial_timeout,omitempty"`
	SendTimeout string `toml:"send_timeout,omitempty"`
	sendTimeout time.Duration

	PingFrequency         string `toml:"ping_frequency,omitempty"`
	PingTimeout           string `toml:"ping_timeout,omitempty"`
	InitialReconnectDelay string `toml:"initial_reconnect_delay,omitempty"`
	MaxReconnectDelay     string `toml:"max_reconnect_delay,omitempty"`
	NewProducerTimeout    string `toml:"new_producer_timeout,omitempty"`

	Topic string `toml:"topic"`
	Name  string `toml:"name"`
}

func (p *Pulsar) SetSerializer(serializer serializers.Serializer) {
	p.serializer = serializer
}

func (p *Pulsar) Connect() error {
	var err error

	p.sendTimeout, err = time.ParseDuration(p.SendTimeout)
	if err != nil {
		return err
	}

	conf := plsr.ManagedProducerConfig{}
	conf.Addr = p.URL

	if p.TLSCA != "" && p.TLSCert != "" && p.TLSKey != "" {
		conf.TLSConfig, err = p.TLSConfig()
	}

	if p.DialTimeout != "" {
		conf.DialTimeout, err = time.ParseDuration(p.DialTimeout)
		conf.ConnectTimeout, err = time.ParseDuration(p.DialTimeout)
	}
	if p.PingFrequency != "" {
		conf.PingFrequency, err = time.ParseDuration(p.PingFrequency)
	}
	if p.PingTimeout != "" {
		conf.PingTimeout, err = time.ParseDuration(p.PingTimeout)
	}
	if p.InitialReconnectDelay != "" {
		conf.InitialReconnectDelay, err = time.ParseDuration(p.InitialReconnectDelay)
	}
	if p.MaxReconnectDelay != "" {
		conf.MaxReconnectDelay, err = time.ParseDuration(p.MaxReconnectDelay)
	}
	if p.NewProducerTimeout != "" {
		conf.NewProducerTimeout, err = time.ParseDuration(p.NewProducerTimeout)
	}
	if p.SendTimeout != "" {
		p.sendTimeout, err = time.ParseDuration(p.SendTimeout)
	} else {
		p.sendTimeout = time.Second * 10
	}

	conf.Topic = p.Topic
	conf.Name = p.Name

	if err != nil {
		return err
	}

	cp := plsr.NewManagedClientPool()
	p.producer = plsr.NewManagedProducer(cp, conf)

	return err
}

func (p *Pulsar) Close() error {
	return nil
}

func (p *Pulsar) SampleConfig() string {
	return sampleConfig
}

func (p *Pulsar) Description() string {
	return "Send telegraf measurements to Apache Pulsar"
}

func (p *Pulsar) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := p.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		ctx, _ := context.WithTimeout(context.Background(), p.sendTimeout)
		_, err = p.producer.Send(ctx, buf)
		if err != nil {
			return fmt.Errorf("FAILED to send Pulsar message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("pulsar", func() telegraf.Output {
		return &Pulsar{}
	})
}
