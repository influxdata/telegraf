package kafka

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/IBM/sarama"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

// ReadConfig for kafka clients meaning to read from Kafka.
type ReadConfig struct {
	Config
}

// SetConfig on the sarama.Config object from the ReadConfig struct.
func (k *ReadConfig) SetConfig(cfg *sarama.Config, log telegraf.Logger) error {
	cfg.Consumer.Return.Errors = true
	return k.Config.SetConfig(cfg, log)
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
func (k *WriteConfig) SetConfig(cfg *sarama.Config, log telegraf.Logger) error {
	cfg.Producer.Return.Successes = true
	cfg.Producer.Idempotent = k.IdempotentWrites
	cfg.Producer.Retry.Max = k.MaxRetry
	if k.MaxMessageBytes > 0 {
		cfg.Producer.MaxMessageBytes = k.MaxMessageBytes
	}
	cfg.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	if cfg.Producer.Idempotent {
		cfg.Net.MaxOpenRequests = 1
	}
	return k.Config.SetConfig(cfg, log)
}

// Config common to all Kafka clients.
type Config struct {
	SASLAuth
	tls.ClientConfig

	Version          string           `toml:"version"`
	ClientID         string           `toml:"client_id"`
	CompressionCodec int              `toml:"compression_codec"`
	EnableTLS        *bool            `toml:"enable_tls"`
	KeepAlivePeriod  *config.Duration `toml:"keep_alive_period"`

	MetadataRetryMax         int             `toml:"metadata_retry_max"`
	MetadataRetryType        string          `toml:"metadata_retry_type"`
	MetadataRetryBackoff     config.Duration `toml:"metadata_retry_backoff"`
	MetadataRetryMaxDuration config.Duration `toml:"metadata_retry_max_duration"`

	// Disable full metadata fetching
	MetadataFull *bool `toml:"metadata_full"`
}

type BackoffFunc func(retries, maxRetries int) time.Duration

func makeBackoffFunc(backoff, maxDuration time.Duration) BackoffFunc {
	return func(retries, _ int) time.Duration {
		d := time.Duration(math.Pow(2, float64(retries))) * backoff
		if maxDuration != 0 && d > maxDuration {
			return maxDuration
		}
		return d
	}
}

// SetConfig on the sarama.Config object from the Config struct.
func (k *Config) SetConfig(cfg *sarama.Config, log telegraf.Logger) error {
	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return fmt.Errorf("parsing kafka version failed: %w", err)
		}

		cfg.Version = version
	}

	if k.ClientID != "" {
		cfg.ClientID = k.ClientID
	} else {
		cfg.ClientID = "Telegraf"
	}

	cfg.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)

	if k.EnableTLS != nil && *k.EnableTLS {
		cfg.Net.TLS.Enable = true
	}

	tlsConfig, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("configuring TLS failed: %w", err)
	}

	if tlsConfig != nil {
		cfg.Net.TLS.Config = tlsConfig

		// To maintain backwards compatibility, if the enable_tls option is not
		// set TLS is enabled if a non-default TLS config is used.
		if k.EnableTLS == nil {
			cfg.Net.TLS.Enable = true
		}
	}

	if k.KeepAlivePeriod != nil {
		// Defaults to OS setting (15s currently)
		cfg.Net.KeepAlive = time.Duration(*k.KeepAlivePeriod)
	}

	if k.MetadataFull != nil {
		// Defaults to true in Sarama
		cfg.Metadata.Full = *k.MetadataFull
	}

	if k.MetadataRetryMax != 0 {
		cfg.Metadata.Retry.Max = k.MetadataRetryMax
	}

	if k.MetadataRetryBackoff != 0 {
		// If cfg.Metadata.Retry.BackoffFunc is set, sarama ignores
		// cfg.Metadata.Retry.Backoff
		cfg.Metadata.Retry.Backoff = time.Duration(k.MetadataRetryBackoff)
	}

	switch strings.ToLower(k.MetadataRetryType) {
	default:
		return errors.New("invalid metadata retry type")
	case "exponential":
		if k.MetadataRetryBackoff == 0 {
			k.MetadataRetryBackoff = config.Duration(250 * time.Millisecond)
			log.Warnf("metadata_retry_backoff is 0, using %s", time.Duration(k.MetadataRetryBackoff))
		}
		cfg.Metadata.Retry.BackoffFunc = makeBackoffFunc(
			time.Duration(k.MetadataRetryBackoff),
			time.Duration(k.MetadataRetryMaxDuration),
		)
	case "constant", "":
	}

	if err := k.SetSASLConfig(cfg); err != nil {
		return fmt.Errorf("configuring SASL failed: %w", err)
	}

	// SASLv0 cannot be used with API requests so disable API requests in this case
	if cfg.Net.SASL.Version == sarama.SASLHandshakeV0 {
		cfg.ApiVersionsRequest = false
	}

	return nil
}
