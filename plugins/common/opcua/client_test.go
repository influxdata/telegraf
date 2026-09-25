package opcua

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestSetupWorkarounds(t *testing.T) {
	o := OpcUAClient{
		Config: &OpcUAClientConfig{
			Workarounds: OpcUAWorkarounds{
				AdditionalValidStatusCodes: []string{"0xC0", "0x00AA0000", "0x80000000"},
			},
		},
	}

	err := o.setupWorkarounds()
	require.NoError(t, err)

	require.Len(t, o.codes, 4)
	require.Equal(t, o.codes[0], ua.StatusCode(0))
	require.Equal(t, o.codes[1], ua.StatusCode(192))
	require.Equal(t, o.codes[2], ua.StatusCode(11141120))
	require.Equal(t, o.codes[3], ua.StatusCode(2147483648))
}

func TestCheckStatusCode(t *testing.T) {
	var o OpcUAClient
	o.codes = []ua.StatusCode{ua.StatusCode(0), ua.StatusCode(192), ua.StatusCode(11141120)}
	require.True(t, o.StatusCodeOK(ua.StatusCode(192)))
}

func TestOpcUAClientConfigValidateSuccess(t *testing.T) {
	config := OpcUAClientConfig{
		Endpoint:       "opc.tcp://localhost:4840",
		SecurityPolicy: "None",
		SecurityMode:   "None",
	}

	err := config.Validate()
	require.NoError(t, err)
}

func TestOpcUAClientConfigValidateFail(t *testing.T) {
	tests := []struct {
		name        string
		config      OpcUAClientConfig
		expectedErr error
	}{
		{
			name: "empty endpoint",
			config: OpcUAClientConfig{
				Endpoint: "",
			},
			expectedErr: ErrInvalidEndpoint,
		},
		{
			name: "invalid endpoint URL",
			config: OpcUAClientConfig{
				Endpoint: "://invalid-url",
			},
			expectedErr: ErrInvalidEndpoint,
		},
		{
			name: "invalid security policy",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "InvalidPolicy",
				SecurityMode:   "None",
			},
			expectedErr: ErrInvalidSecurityPolicy,
		},
		{
			name: "invalid security mode",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "None",
				SecurityMode:   "InvalidMode",
			},
			expectedErr: ErrInvalidSecurityMode,
		},
		{
			name: "certificate without private key",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "Basic256",
				SecurityMode:   "SignAndEncrypt",
				Certificate:    "cert.pem",
				PrivateKey:     "",
			},
			expectedErr: ErrInvalidConfiguration,
		},
		{
			name: "private key without certificate",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "Basic256",
				SecurityMode:   "SignAndEncrypt",
				Certificate:    "",
				PrivateKey:     "key.pem",
			},
			expectedErr: ErrInvalidConfiguration,
		},
		{
			name: "invalid auth method",
			config: OpcUAClientConfig{
				Endpoint:   "opc.tcp://localhost:4840",
				AuthMethod: "InvalidAuth",
			},
			expectedErr: ErrInvalidAuthMethod,
		},
		{
			name: "invalid optional field",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				OptionalFields: []string{"InvalidField"},
			},
			expectedErr: ErrInvalidConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.Error(t, err)
			// Check that the error chain contains the expected error type
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestOpcUAClientSetupWorkarounds(t *testing.T) {
	tests := []struct {
		name        string
		statusCodes []string
		expectErr   bool
	}{
		{
			name:        "valid status codes",
			statusCodes: []string{"0x80", "0x00AA0000", "123"},
			expectErr:   false,
		},
		{
			name:        "invalid status code",
			statusCodes: []string{"invalid"},
			expectErr:   true,
		},
		{
			name:        "empty status codes",
			statusCodes: nil,
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &OpcUAClient{
				Config: &OpcUAClientConfig{
					Workarounds: OpcUAWorkarounds{
						AdditionalValidStatusCodes: tt.statusCodes,
					},
				},
			}

			err := client.setupWorkarounds()
			if tt.expectErr {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrStatusCodeParsing)
			} else {
				require.NoError(t, err)
				// Should always have at least StatusOK
				require.GreaterOrEqual(t, len(client.codes), 1)
				require.Equal(t, ua.StatusOK, client.codes[0])
			}
		})
	}
}

func TestRemoteCertificateValidation(t *testing.T) {
	tests := []struct {
		name              string
		securityPolicy    string
		securityMode      string
		remoteCertificate string
	}{
		{
			name:              "no remote certificate configured",
			securityPolicy:    "None",
			securityMode:      "None",
			remoteCertificate: "",
		},
		{
			name:              "remote certificate path provided with None security",
			securityPolicy:    "None",
			securityMode:      "None",
			remoteCertificate: "/etc/telegraf/server_cert.pem",
		},
		{
			name:              "remote certificate path provided with SignAndEncrypt",
			securityPolicy:    "Basic256Sha256",
			securityMode:      "SignAndEncrypt",
			remoteCertificate: "/etc/telegraf/server_cert.pem",
		},
		{
			name:              "remote certificate path provided with auto security",
			securityPolicy:    "auto",
			securityMode:      "auto",
			remoteCertificate: "/etc/telegraf/server_cert.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OpcUAClientConfig{
				Endpoint:          "opc.tcp://localhost:4840",
				SecurityPolicy:    tt.securityPolicy,
				SecurityMode:      tt.securityMode,
				RemoteCertificate: tt.remoteCertificate,
			}

			require.NoError(t, config.Validate())
		})
	}
}

func TestGenerateClientOptsExtras(t *testing.T) {
	endpoints := []*ua.EndpointDescription{
		{
			EndpointURL:       "opc.tcp://localhost:4840",
			SecurityPolicyURI: ua.SecurityPolicyURINone,
			SecurityMode:      ua.MessageSecurityModeNone,
			SecurityLevel:     0,
			UserIdentityTokens: []*ua.UserTokenPolicy{
				{TokenType: ua.UserTokenTypeAnonymous},
			},
		},
	}

	newBaseClient := func() *OpcUAClient {
		return &OpcUAClient{
			Config: &OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "None",
				SecurityMode:   "None",
				AuthMethod:     "Anonymous",
			},
			Log: &testutil.Logger{},
		}
	}

	baseOpts, err := newBaseClient().generateClientOpts(endpoints)
	require.NoError(t, err)

	tests := []struct {
		name   string
		modify func(*OpcUAClient)
	}{
		{
			name:   "locales",
			modify: func(c *OpcUAClient) { c.Config.Locales = []string{"en", "de"} },
		},
		{
			name:   "disable auto-reconnect",
			modify: func(c *OpcUAClient) { c.DisableAutoReconnect = true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newBaseClient()
			tt.modify(client)
			opts, err := client.generateClientOpts(endpoints)
			require.NoError(t, err)
			require.Len(t, opts, len(baseOpts)+1)
		})
	}
}

func TestGenerateClientOptsNoChannelCertForNoneChannel(t *testing.T) {
	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")
	_, _, err := generateCert("urn:telegraf:gopcua:client", 2048, certFile, keyFile, time.Hour)
	require.NoError(t, err)

	endpoints := []*ua.EndpointDescription{
		{
			EndpointURL:        "opc.tcp://localhost:4840",
			SecurityPolicyURI:  ua.SecurityPolicyURINone,
			SecurityMode:       ua.MessageSecurityModeNone,
			SecurityLevel:      0,
			UserIdentityTokens: []*ua.UserTokenPolicy{{TokenType: ua.UserTokenTypeAnonymous}},
		},
		{
			EndpointURL:        "opc.tcp://localhost:4840",
			SecurityPolicyURI:  ua.SecurityPolicyURIPrefix + "Basic256Sha256",
			SecurityMode:       ua.MessageSecurityModeSignAndEncrypt,
			SecurityLevel:      100,
			UserIdentityTokens: []*ua.UserTokenPolicy{{TokenType: ua.UserTokenTypeAnonymous}},
		},
	}

	newClient := func(policy, mode string) *OpcUAClient {
		return &OpcUAClient{
			Config: &OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: policy,
				SecurityMode:   mode,
				AuthMethod:     "Anonymous",
				Certificate:    certFile,
				PrivateKey:     keyFile,
			},
			Log: &testutil.Logger{},
		}
	}

	// A genuinely secured channel attaches the client certificate and private key.
	secureOpts, err := newClient("Basic256Sha256", "SignAndEncrypt").generateClientOpts(endpoints)
	require.NoError(t, err)

	// security_mode "None" collapses the channel to None. The client certificate
	// must not be attached, otherwise the OpenSecureChannel request carries a
	// SenderCertificate under security policy None and strict servers reject it.
	noneOpts, err := newClient("Basic256Sha256", "None").generateClientOpts(endpoints)
	require.NoError(t, err)

	require.Len(t, noneOpts, len(secureOpts)-2)
}
