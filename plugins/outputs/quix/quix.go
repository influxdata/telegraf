// quix.go
package quix

import (
	"crypto/sha256"
	"strings"

	"github.com/IBM/sarama"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

// Quix is the main struct for the Quix plugin
type Quix struct {
	Brokers        []string `toml:"brokers"`
	Topic          string   `toml:"topic"`
	Workspace      string   `toml:"workspace"`
	AuthToken      string   `toml:"auth_token"`
	APIURL         string   `toml:"api_url"`
	TimestampUnits string   `toml:"timestamp_units"`

	producer   sarama.SyncProducer
	Log        telegraf.Logger
	serializer serializers.Serializer
}

// SampleConfig returns a sample configuration for the Quix plugin
func (q *Quix) SampleConfig() string {
	return `
  ## Quix output plugin configuration
  workspace = "your_workspace"
  auth_token = "your_auth_token"
  api_url = "https://portal-api.platform.quix.io"
  topic = "telegraf_metrics"
  data_format = "json" 
  timestamp_units = "1s"
`
}

// Init initializes the Quix plugin and sets up the serializer
func (q *Quix) Init() error {
	duration, err := parseTimestampUnits(q.TimestampUnits)
	if err != nil {
		return err
	}

	q.serializer, err = serializers.NewSerializer(&serializers.Config{
		DataFormat:     "json",
		TimestampUnits: duration,
	})
	if err != nil {
		return err
	}

	q.Log.Infof("Initializing Quix plugin.")
	return nil
}

// Connect establishes the connection to Kafka
func (q *Quix) Connect() error {
	quixConfig, cert, err := q.fetchBrokerConfig()
	if err != nil {
		return err
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = quixConfig.SaslUsername
	config.Net.SASL.Password = quixConfig.SaslPassword
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
		return &XDGSCRAMClient{HashGeneratorFcn: sha256.New}
	}

	tlsConfig, err := q.createTLSConfig(cert)
	if err != nil {
		return err
	}
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig

	producer, err := sarama.NewSyncProducer(strings.Split(quixConfig.BootstrapServers, ","), config)
	if err != nil {
		return err
	}
	q.producer = producer
	q.Log.Infof("Connected to Quix.")
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
