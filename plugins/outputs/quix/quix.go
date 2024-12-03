//go:generate ../../../tools/readme_config_includer/generator
package quix

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/IBM/sarama"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	common_kafka "github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

//go:embed sample.conf
var sampleConfig string

type Quix struct {
	APIURL         string          `toml:"url"`
	Workspace      string          `toml:"workspace"`
	Topic          string          `toml:"topic"`
	Token          config.Secret   `toml:"token"`
	TimestampUnits config.Duration `toml:"timestamp_units"`
	Log            telegraf.Logger `toml:"-"`
	common_http.HTTPClientConfig

	producer   sarama.SyncProducer
	serializer serializers.Serializer
}

func (*Quix) SampleConfig() string {
	return sampleConfig
}

// Init initializes the Quix plugin and sets up the serializer
func (q *Quix) Init() error {
	// Set defaults
	if q.APIURL == "" {
		q.APIURL = "https://portal-api.platform.quix.io"
	}
	q.APIURL = strings.TrimSuffix(q.APIURL, "/")

	// Check input parameters
	if q.Topic == "" {
		return errors.New("option 'topic' must be set")
	}
	if q.Workspace == "" {
		return errors.New("option 'workspace' must be set")
	}
	if q.Token.Empty() {
		return errors.New("option 'token' must be set")
	}

	// Create a JSON serializer for the output
	q.serializer = &json.Serializer{
		TimestampUnits: q.TimestampUnits,
	}

	return nil
}

func (q *Quix) Connect() error {
	// Fetch the Kafka broker configuration from the Quix HTTP endpoint
	quixConfig, err := q.fetchBrokerConfig()
	if err != nil {
		return fmt.Errorf("fetching broker config failed: %w", err)
	}
	brokers := strings.Split(quixConfig.BootstrapServers, ",")
	if len(brokers) == 0 {
		return errors.New("no brokers received")
	}

	// Setup the Kakfa producer config
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Net.SASL.Enable = true
	cfg.Net.SASL.User = quixConfig.SaslUsername
	cfg.Net.SASL.Password = quixConfig.SaslPassword
	cfg.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &common_kafka.XDGSCRAMClient{HashGeneratorFcn: common_kafka.SHA256}
	}

	// Certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(quixConfig.cert) {
		return errors.New("appending CA cert to pool failed")
	}
	cfg.Net.TLS.Enable = true
	cfg.Net.TLS.Config = &tls.Config{RootCAs: certPool}

	// Setup the Kakfa producer itself
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return err
	}
	q.producer = producer

	return nil
}

// Write sends serialized metrics to Quix
func (q *Quix) Write(metrics []telegraf.Metric) error {
	q.Log.Debugf("Sending metrics to Quix.")
	for _, metric := range metrics {
		serialized, err := q.serializer.Serialize(metric)
		if err != nil {
			q.Log.Errorf("Error serializing metric: %v", err)
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic:     q.Workspace + "-" + q.Topic,
			Value:     sarama.ByteEncoder(serialized),
			Timestamp: metric.Time(),
			Key:       sarama.StringEncoder("telegraf"),
		}

		if _, _, err = q.producer.SendMessage(msg); err != nil {
			q.Log.Errorf("Error sending message to Kafka: %v", err)
		}
	}
	q.Log.Debugf("Metrics sent to Quix.")
	return nil
}

// Close shuts down the Kafka producer
func (q *Quix) Close() error {
	if q.producer != nil {
		q.Log.Infof("Closing Quix producer connection.")
		return q.producer.Close()
	}
	return nil
}

// Initialize Quix plugin in Telegraf
func init() {
	outputs.Add("quix", func() telegraf.Output { return &Quix{} })
}
