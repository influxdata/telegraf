package kafka

import (
	"errors"

	"github.com/Shopify/sarama"
)

type SASLAuth struct {
	SASLUsername  string `toml:"sasl_username"`
	SASLPassword  string `toml:"sasl_password"`
	SASLMechanism string `toml:"sasl_mechanism"`
	SASLVersion   *int   `toml:"sasl_version"`

	// GSSAPI config
	SASLGSSAPIServiceName        string `toml:"sasl_gssapi_service_name"`
	SASLGSSAPIAuthType           string `toml:"sasl_gssapi_auth_type"`
	SASLGSSAPIDisablePAFXFAST    bool   `toml:"sasl_gssapi_disable_pafxfast"`
	SASLGSSAPIKerberosConfigPath string `toml:"sasl_gssapi_kerberos_config_path"`
	SASLGSSAPIKeyTabPath         string `toml:"sasl_gssapi_key_tab_path"`
	SASLGSSAPIRealm              string `toml:"sasl_gssapi_realm"`

	// OAUTHBEARER config. experimental. undoubtedly this is not good enough.
	SASLAccessToken string `toml:"sasl_access_token"`
}

// SetSASLConfig configures SASL for kafka (sarama)
func (k *SASLAuth) SetSASLConfig(config *sarama.Config) error {
	config.Net.SASL.User = k.SASLUsername
	config.Net.SASL.Password = k.SASLPassword

	if k.SASLMechanism != "" {
		config.Net.SASL.Mechanism = sarama.SASLMechanism(k.SASLMechanism)
		switch config.Net.SASL.Mechanism {
		case sarama.SASLTypeSCRAMSHA256:
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
			}
		case sarama.SASLTypeSCRAMSHA512:
			config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
				return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
			}
		case sarama.SASLTypeOAuth:
			config.Net.SASL.TokenProvider = k // use self as token provider.
		case sarama.SASLTypeGSSAPI:
			config.Net.SASL.GSSAPI.ServiceName = k.SASLGSSAPIServiceName
			config.Net.SASL.GSSAPI.AuthType = gssapiAuthType(k.SASLGSSAPIAuthType)
			config.Net.SASL.GSSAPI.Username = k.SASLUsername
			config.Net.SASL.GSSAPI.Password = k.SASLPassword
			config.Net.SASL.GSSAPI.DisablePAFXFAST = k.SASLGSSAPIDisablePAFXFAST
			config.Net.SASL.GSSAPI.KerberosConfigPath = k.SASLGSSAPIKerberosConfigPath
			config.Net.SASL.GSSAPI.KeyTabPath = k.SASLGSSAPIKeyTabPath
			config.Net.SASL.GSSAPI.Realm = k.SASLGSSAPIRealm

		case sarama.SASLTypePlaintext:
			// nothing.
		default:
		}
	}

	if k.SASLUsername != "" || k.SASLMechanism != "" {
		config.Net.SASL.Enable = true

		version, err := SASLVersion(config.Version, k.SASLVersion)
		if err != nil {
			return err
		}
		config.Net.SASL.Version = version
	}
	return nil
}

// Token does nothing smart, it just grabs a hard-coded token from config.
func (k *SASLAuth) Token() (*sarama.AccessToken, error) {
	return &sarama.AccessToken{
		Token:      k.SASLAccessToken,
		Extensions: map[string]string{},
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
