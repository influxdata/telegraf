package kafka

import (
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf/config"
)

type SASLAuth struct {
	SASLUsername   config.Secret     `toml:"sasl_username"`
	SASLPassword   config.Secret     `toml:"sasl_password"`
	SASLExtentions map[string]string `toml:"sasl_extensions"`
	SASLMechanism  string            `toml:"sasl_mechanism"`
	SASLVersion    *int              `toml:"sasl_version"`

	// GSSAPI config
	SASLGSSAPIServiceName        string `toml:"sasl_gssapi_service_name"`
	SASLGSSAPIAuthType           string `toml:"sasl_gssapi_auth_type"`
	SASLGSSAPIDisablePAFXFAST    bool   `toml:"sasl_gssapi_disable_pafxfast"`
	SASLGSSAPIKerberosConfigPath string `toml:"sasl_gssapi_kerberos_config_path"`
	SASLGSSAPIKeyTabPath         string `toml:"sasl_gssapi_key_tab_path"`
	SASLGSSAPIRealm              string `toml:"sasl_gssapi_realm"`

	// OAUTHBEARER config
	SASLAccessToken config.Secret `toml:"sasl_access_token"`
}

// SetSASLConfig configures SASL for kafka (sarama)
func (k *SASLAuth) SetSASLConfig(cfg *sarama.Config) error {
	username, err := k.SASLUsername.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	cfg.Net.SASL.User = string(username)
	config.ReleaseSecret(username)
	password, err := k.SASLPassword.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	cfg.Net.SASL.Password = string(password)
	config.ReleaseSecret(password)

	if k.SASLMechanism != "" {
		cfg.Net.SASL.Mechanism = sarama.SASLMechanism(k.SASLMechanism)
		switch cfg.Net.SASL.Mechanism {
		case sarama.SASLTypeSCRAMSHA256:
			cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
			}
		case sarama.SASLTypeSCRAMSHA512:
			cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
			}
		case sarama.SASLTypeOAuth:
			cfg.Net.SASL.TokenProvider = k // use self as token provider.
		case sarama.SASLTypeGSSAPI:
			cfg.Net.SASL.GSSAPI.ServiceName = k.SASLGSSAPIServiceName
			cfg.Net.SASL.GSSAPI.AuthType = gssapiAuthType(k.SASLGSSAPIAuthType)
			cfg.Net.SASL.GSSAPI.Username = string(username)
			cfg.Net.SASL.GSSAPI.Password = string(password)
			cfg.Net.SASL.GSSAPI.DisablePAFXFAST = k.SASLGSSAPIDisablePAFXFAST
			cfg.Net.SASL.GSSAPI.KerberosConfigPath = k.SASLGSSAPIKerberosConfigPath
			cfg.Net.SASL.GSSAPI.KeyTabPath = k.SASLGSSAPIKeyTabPath
			cfg.Net.SASL.GSSAPI.Realm = k.SASLGSSAPIRealm

		case sarama.SASLTypePlaintext:
			// nothing.
		default:
		}
	}

	if len(username) > 0 || k.SASLMechanism != "" {
		cfg.Net.SASL.Enable = true

		version, err := SASLVersion(cfg.Version, k.SASLVersion)
		if err != nil {
			return err
		}
		cfg.Net.SASL.Version = version
	}
	return nil
}

// Token does nothing smart, it just grabs a hard-coded token from config.
func (k *SASLAuth) Token() (*sarama.AccessToken, error) {
	token, err := k.SASLAccessToken.Get()
	if err != nil {
		return nil, fmt.Errorf("getting token failed: %w", err)
	}
	defer config.ReleaseSecret(token)
	return &sarama.AccessToken{
		Token:      string(token),
		Extensions: k.SASLExtentions,
	}, nil
}

func SASLVersion(kafkaVersion sarama.KafkaVersion, saslVersion *int) (int16, error) {
	if saslVersion == nil {
		if kafkaVersion.IsAtLeast(sarama.V1_0_0_0) {
			return sarama.SASLHandshakeV1, nil
		}
		return sarama.SASLHandshakeV0, nil
	}

	switch *saslVersion {
	case 0:
		return sarama.SASLHandshakeV0, nil
	case 1:
		return sarama.SASLHandshakeV1, nil
	default:
		return 0, errors.New("invalid SASL version")
	}
}

func gssapiAuthType(authType string) int {
	switch authType {
	case "KRB5_USER_AUTH":
		return sarama.KRB5_USER_AUTH
	case "KRB5_KEYTAB_AUTH":
		return sarama.KRB5_KEYTAB_AUTH
	default:
		return 0
	}
}
