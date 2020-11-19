package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

type ReadConfig struct {
	Config
}

func (k *ReadConfig) SetConfig(config *sarama.Config) error {
	config.Consumer.Return.Errors = true

	return k.Config.SetConfig(config)
}

type WriteConfig struct {
	Config

	RequiredAcks     int  `toml:"required_acks"`
	MaxRetry         int  `toml:"max_retry"`
	MaxMessageBytes  int  `toml:"max_message_bytes"`
	IdempotentWrites bool `toml:"idempotent_writes"`
}

func (k *WriteConfig) SetConfig(config *sarama.Config) error {
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = k.IdempotentWrites
	config.Producer.Retry.Max = k.MaxRetry
	if k.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = k.MaxMessageBytes
	}
	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	return k.Config.SetConfig(config)
}

type Config struct {
	SASLAuth
	tls.ClientConfig

	Version          string `toml:"version"`
	ClientID         string `toml:"client_id"`
	CompressionCodec int    `toml:"compression_codec"`

	// EnableTLS deprecated
	_EnableTLS *bool `toml:"enable_tls"`
}

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

	tlsConfig, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	if err := k.SetSASLConfig(config); err != nil {
		return err
	}

	return nil
}
