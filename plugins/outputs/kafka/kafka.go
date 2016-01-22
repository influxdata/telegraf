package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/telegraf/plugins/outputs"
	"io/ioutil"
)

type Kafka struct {
	// Kafka brokers to send metrics to
	Brokers []string
	// Kafka topic
	Topic string
	// Routing Key Tag
	RoutingTag string `toml:"routing_tag"`
	// TLS client certificate
	Certificate string
	// TLS client key
	Key string
	// TLS certificate authority
	CA string
	// Verfiy SSL certificate chain
	VerifySsl bool

	tlsConfig tls.Config
	producer  sarama.SyncProducer
}

var sampleConfig = `
  # URLs of kafka brokers
  brokers = ["localhost:9092"]
  # Kafka topic for producer messages
  topic = "telegraf"
  # Telegraf tag to use as a routing key
  #  ie, if this tag exists, it's value will be used as the routing key
  routing_tag = "host"

  # Optional TLS configuration:
  # Client certificate
  certificate = ""
  # Client key
  key = ""
  # Certificate authority file
  ca = ""
  # Verify SSL certificate chain
  verify_ssl = false
`

func createTlsConfiguration(k *Kafka) (t *tls.Config, err error) {
	if k.Certificate != "" && k.Key != "" && k.CA != "" {
		cert, err := tls.LoadX509KeyPair(k.Certificate, k.Key)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cout not load Kafka TLS client key/certificate: %s",
				err))
		}

		caCert, err := ioutil.ReadFile(k.CA)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Cout not load Kafka TLS CA: %s",
				err))
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: k.VerifySsl,
		}
	}
	// will be nil by default if nothing is provided
	return t, nil
}

func (k *Kafka) Connect() error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	tlsConfig, err := createTlsConfiguration(k)
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	producer, err := sarama.NewSyncProducer(k.Brokers, config)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}

func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Description() string {
	return "Configuration for the Kafka server to send metrics to"
}

func (k *Kafka) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}

	for _, p := range points {
		// Combine tags from Point and BatchPoints and grab the resulting
		// line-protocol output string to write to Kafka
		value := p.String()

		m := &sarama.ProducerMessage{
			Topic: k.Topic,
			Value: sarama.StringEncoder(value),
		}
		if h, ok := p.Tags()[k.RoutingTag]; ok {
			m.Key = sarama.StringEncoder(h)
		}

		_, _, err := k.producer.SendMessage(m)
		if err != nil {
			return errors.New(fmt.Sprintf("FAILED to send kafka message: %s\n",
				err))
		}
	}
	return nil
}

func init() {
	outputs.Add("kafka", func() outputs.Output {
		return &Kafka{}
	})
}
