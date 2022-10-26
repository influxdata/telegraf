package kafka

import (
	"fmt"
	"math"
	"time"

	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

// ReadConfig for kafka clients meaning to read from Kafka.
type ReadConfig struct {
	Config
}

// SetConfig on the sarama.Config object from the ReadConfig struct.
func (k *ReadConfig) SetConfig(config *sarama.Config) error {
	config.Consumer.Return.Errors = true

	return k.Config.SetConfig(config)
}

// WriteConfig for kafka clients meaning to write to kafka
type WriteConfig struct {
	Config

	RequiredAcks     int  `toml:"required_acks"`
	MaxRetry         int  `toml:"max_retry"`
	MaxMessageBytes  int  `toml:"max_message_bytes"`
	IdempotentWrites bool `toml:"idempotent_writes"`
}

// SetConfig on the sarama.Config object from the WriteConfig struct.
func (k *WriteConfig) SetConfig(config *sarama.Config) error {
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = k.IdempotentWrites
	config.Producer.Retry.Max = k.MaxRetry
	if k.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = k.MaxMessageBytes
	}
	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	if config.Producer.Idempotent {
		config.Net.MaxOpenRequests = 1
	}
	return k.Config.SetConfig(config)
}

// Config common to all Kafka clients.
type Config struct {
	SASLAuth
	tls.ClientConfig

	Version          string `toml:"version"`
	ClientID         string `toml:"client_id"`
	CompressionCodec int    `toml:"compression_codec"`
	EnableTLS        *bool  `toml:"enable_tls"`

	MetadataRetryMax         int           `toml:"metadata_retry_max"`
	MetadataRetryType        string        `toml:"metadata_retry_type"`
	MetadataRetryBackoff     time.Duration `toml:"metadata_retry_backoff"`
	MetadataRetryMaxDuration time.Duration `toml:"metadata_retry_max_duration"`

	Log telegraf.Logger `toml:"-"`

	// Disable full metadata fetching
	MetadataFull *bool `toml:"metadata_full"`
}

type BackoffFunc func(retries, maxRetries int) time.Duration

func makeBackoffFunc(backoff, maxDuration time.Duration) BackoffFunc {
	return func(retries, maxRetries int) time.Duration {
		d := time.Duration(math.Pow(2, float64(retries))) * backoff
		if maxDuration != 0 && d > maxDuration {
			return maxDuration
		}
		return d
	}
}

// SetConfig on the sarama.Config object from the Config struct.
func (k *Config) SetConfig(config *sarama.Config) error {
	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return err
		}

		config.Version = version
	}

	if k.ClientID != "" {
		config.ClientID = k.ClientID
	} else {
		config.ClientID = "Telegraf"
	}

	config.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)

	if k.EnableTLS != nil && *k.EnableTLS {
		config.Net.TLS.Enable = true
	}

	tlsConfig, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig

		// To maintain backwards compatibility, if the enable_tls option is not
		// set TLS is enabled if a non-default TLS config is used.
		if k.EnableTLS == nil {
			config.Net.TLS.Enable = true
		}
	}

	if k.MetadataFull != nil {
		// Defaults to true in Sarama
		config.Metadata.Full = *k.MetadataFull
	}

	if k.MetadataRetryMax != 0 {
		config.Metadata.Retry.Max = k.MetadataRetryMax
	}

	if k.MetadataRetryBackoff != 0 {
		// If config.Metadata.Retry.BackoffFunc is set, sarama ignores
		// config.Metadata.Retry.Backoff
		config.Metadata.Retry.Backoff = k.MetadataRetryBackoff
	}

	switch t := k.MetadataRetryType; t {
	default:
		return fmt.Errorf("invalid metadata retry type")
	case "exponential":
		if k.MetadataRetryBackoff != 0 {
			config.Metadata.Retry.BackoffFunc = makeBackoffFunc(k.MetadataRetryBackoff, k.MetadataRetryMaxDuration)
		}
	case "none":
		fallthrough
	case "":
	}

	return k.SetSASLConfig(config)
}
