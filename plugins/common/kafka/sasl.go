package kafka

import (
	"errors"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/influxdata/telegraf/config"
)

type SASLAuth struct {
	SASLUsername   config.Secret     `toml:"sasl_username"`
	SASLPassword   config.Secret     `toml:"sasl_password"`
	SASLExtensions map[string]string `toml:"sasl_extensions"`
	SASLMechanism  string            `toml:"sasl_mechanism"`
	SASLVersion    *int              `toml:"sasl_version"`

	// GSSAPI config
	SASLGSSAPIServiceName        string `toml:"sasl_gssapi_service_name"`
	SASLGSSAPIAuthType           string `toml:"sasl_gssapi_auth_type"`
	SASLGSSAPIDisablePAFXFAST    bool   `toml:"sasl_gssapi_disable_pafxfast"`
	SASLGSSAPIKerberosConfigPath string `toml:"sasl_gssapi_kerberos_config_path"`
	SASLGSSAPIKeyTabPath         string `toml:"sasl_gssapi_key_tab_path"`
	SASLGSSAPIRealm              string `toml:"sasl_gssapi_realm"`

	// OAUTHBEARER token based config
	SASLAccessToken config.Secret `toml:"sasl_access_token"`

	// OAUTHBEARER AWS MSK IAM based config
	SASLOAuthAWSMSKIAMConfig
}

// SetSASLConfig configures SASL for kafka (sarama)
func (k *SASLAuth) SetSASLConfig(cfg *sarama.Config) error {
	username, err := k.SASLUsername.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	cfg.Net.SASL.User = username.String()
	defer username.Destroy()
	password, err := k.SASLPassword.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	cfg.Net.SASL.Password = password.String()
	defer password.Destroy()

	mechanism := k.SASLMechanism

	switch k.SASLMechanism {
	case sarama.SASLTypeSCRAMSHA256:
		cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA256}
		}
	case sarama.SASLTypeSCRAMSHA512:
		cfg.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: SHA512}
		}
	case sarama.SASLTypeGSSAPI:
		cfg.Net.SASL.GSSAPI.ServiceName = k.SASLGSSAPIServiceName
		cfg.Net.SASL.GSSAPI.AuthType = gssapiAuthType(k.SASLGSSAPIAuthType)
		cfg.Net.SASL.GSSAPI.Username = username.String()
		cfg.Net.SASL.GSSAPI.Password = password.String()
		cfg.Net.SASL.GSSAPI.DisablePAFXFAST = k.SASLGSSAPIDisablePAFXFAST
		cfg.Net.SASL.GSSAPI.KerberosConfigPath = k.SASLGSSAPIKerberosConfigPath
		cfg.Net.SASL.GSSAPI.KeyTabPath = k.SASLGSSAPIKeyTabPath
		cfg.Net.SASL.GSSAPI.Realm = k.SASLGSSAPIRealm
	case sarama.SASLTypeOAuth: // OAUTHBEARER secret based auth
		cfg.Net.SASL.TokenProvider = &oauthToken{
			token:      k.SASLAccessToken,
			extensions: k.SASLExtensions,
		}
	case saslTypeOAuthAWSMSKIAM: // AWS-MSK-IAM based auth
		p, err := k.SASLOAuthAWSMSKIAMConfig.tokenProvider(k.SASLExtensions)
		if err != nil {
			return fmt.Errorf("creating AWS MSK IAM token provider failed: %w", err)
		}
		mechanism = sarama.SASLTypeOAuth
		cfg.Net.SASL.TokenProvider = p
	case sarama.SASLTypePlaintext:
		// nothing.
	case "":
		// no SASL
	}
	cfg.Net.SASL.Mechanism = sarama.SASLMechanism(mechanism)

	if !k.SASLUsername.Empty() || k.SASLMechanism != "" {
		cfg.Net.SASL.Enable = true

		version, err := SASLVersion(cfg.Version, k.SASLVersion)
		if err != nil {
			return err
		}
		cfg.Net.SASL.Version = version
	}
	return nil
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
