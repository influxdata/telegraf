package opcua

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"
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

func TestServerCertificateValidationSuccess(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	validCertPath := filepath.Join(tempDir, "server_cert.pem")
	err := os.WriteFile(validCertPath, []byte("fake certificate content"), 0600)
	require.NoError(t, err)

	tests := []struct {
		name              string
		securityPolicy    string
		securityMode      string
		serverCertificate string
	}{
		{
			name:              "no server certificate configured",
			securityPolicy:    "None",
			securityMode:      "None",
			serverCertificate: "",
		},
		{
			name:              "valid server certificate with None security",
			securityPolicy:    "None",
			securityMode:      "None",
			serverCertificate: validCertPath,
		},
		{
			name:              "valid server certificate with SignAndEncrypt",
			securityPolicy:    "Basic256Sha256",
			securityMode:      "SignAndEncrypt",
			serverCertificate: validCertPath,
		},
		{
			name:              "valid server certificate with auto security",
			securityPolicy:    "auto",
			securityMode:      "auto",
			serverCertificate: validCertPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OpcUAClientConfig{
				Endpoint:          "opc.tcp://localhost:4840",
				SecurityPolicy:    tt.securityPolicy,
				SecurityMode:      tt.securityMode,
				ServerCertificate: tt.serverCertificate,
			}

			err := config.Validate()
			require.NoError(t, err)
		})
	}
}

func TestServerCertificateValidationFailure(t *testing.T) {
	tests := []struct {
		name              string
		serverCertificate string
		expectedErr       error
	}{
		{
			name:              "nonexistent server certificate file",
			serverCertificate: "/nonexistent/path/to/cert.pem",
			expectedErr:       ErrInvalidConfiguration,
		},
		{
			name:              "invalid path with special characters",
			serverCertificate: "/path/with\x00null/cert.pem",
			expectedErr:       ErrInvalidConfiguration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := OpcUAClientConfig{
				Endpoint:          "opc.tcp://localhost:4840",
				SecurityPolicy:    "None",
				SecurityMode:      "None",
				ServerCertificate: tt.serverCertificate,
			}

			err := config.Validate()
			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
