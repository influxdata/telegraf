package kafka

import (
	"errors"

	"github.com/Shopify/sarama"
)

var (
	ErrUnknownMechanism = errors.New("unknown mechanism")
)

type SASLConfig struct {
	SASLMechanism string `toml:"sasl_mechanism"`
	SASLUsername  string `toml:"sasl_username"`
	SASLPassword  string `toml:"sasl_password"`
}

type SaramaSASL struct {
	Enable                   bool
	Mechanism                sarama.SASLMechanism
	User                     string
	Password                 string
	SCRAMClientGeneratorFunc func() sarama.SCRAMClient
}

func (c *SASLConfig) SetSaramaSASLConfig(config *sarama.Config) error {
	sasl := config.Net.SASL
	switch c.SASLMechanism {
	case "":
		// SASL is not enabled
		return nil
	case "PLAIN":
		sasl.Mechanism = sarama.SASLTypePlaintext
	case "SCRAM-SHA-256":
		sasl.Mechanism = sarama.SASLTypeSCRAMSHA256
		sasl.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
		}
	case "SCRAM-SHA-512":
		sasl.Mechanism = sarama.SASLTypeSCRAMSHA512
		sasl.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
		}
	default:
		return ErrUnknownMechanism
	}

	if c.SASLUsername != "" || c.SASLPassword != "" {
		sasl.User = c.SASLUsername
		sasl.Password = c.SASLPassword
		sasl.Enable = true
	}

	return nil
}
