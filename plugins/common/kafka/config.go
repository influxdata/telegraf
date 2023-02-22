package kafka

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf"
	tgConf "github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

// ReadConfig for kafka clients meaning to read from Kafka.
type ReadConfig struct {
	Config
}

// SetConfig on the sarama.Config object from the ReadConfig struct.
func (k *ReadConfig) SetConfig(config *sarama.Config, log telegraf.Logger) error {
	config.Consumer.Return.Errors = true

	return k.Config.SetConfig(config, log)
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
func (k *WriteConfig) SetConfig(config *sarama.Config, log telegraf.Logger) error {
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
	return k.Config.SetConfig(config, log)
}

// Config common to all Kafka clients.
type Config struct {
	SASLAuth
	tls.ClientConfig

	Version          string           `toml:"version"`
	ClientID         string           `toml:"client_id"`
	CompressionCodec int              `toml:"compression_codec"`
	EnableTLS        *bool            `toml:"enable_tls"`
	KeepAlivePeriod  *tgConf.Duration `toml:"keep_alive_period"`

	MetadataRetryMax         int             `toml:"metadata_retry_max"`
	MetadataRetryType        string          `toml:"metadata_retry_type"`
	MetadataRetryBackoff     tgConf.Duration `toml:"metadata_retry_backoff"`
	MetadataRetryMaxDuration tgConf.Duration `toml:"metadata_retry_max_duration"`

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
func (k *Config) SetConfig(config *sarama.Config, log telegraf.Logger) error {
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

	if k.KeepAlivePeriod != nil {
		// Defaults to OS setting (15s currently)
		config.Net.KeepAlive = time.Duration(*k.KeepAlivePeriod)
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
		config.Metadata.Retry.Backoff = time.Duration(k.MetadataRetryBackoff)
	}

	switch strings.ToLower(k.MetadataRetryType) {
	default:
		return fmt.Errorf("invalid metadata retry type")
	case "exponential":
		if k.MetadataRetryBackoff == 0 {
			k.MetadataRetryBackoff = tgConf.Duration(250 * time.Millisecond)
			log.Warnf("metadata_retry_backoff is 0, using %s", time.Duration(k.MetadataRetryBackoff))
		}
		config.Metadata.Retry.BackoffFunc = makeBackoffFunc(
			time.Duration(k.MetadataRetryBackoff),
			time.Duration(k.MetadataRetryMaxDuration),
		)
	case "constant", "":
	}

	return k.SetSASLConfig(config)
}
