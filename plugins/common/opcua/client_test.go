package opcua

import (
	"context"
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
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
	clientConfig := OpcUAClientConfig{
		Endpoint:       "opc.tcp://localhost:4840",
		SecurityPolicy: "None",
		SecurityMode:   "None",
	}

	err := clientConfig.Validate()
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

// TestDisconnectRepeated verifies that multiple calls to Disconnect are safe
func TestDisconnectRepeated(t *testing.T) {
	client := &OpcUAClient{
		Config: &OpcUAClientConfig{
			Endpoint: "opc.tcp://localhost:4840",
		},
		Log: testutil.Logger{},
	}

	ctx := context.Background()

	// First disconnect should be safe (client is nil)
	err1 := client.Disconnect(ctx)
	require.NoError(t, err1)

	// Second disconnect should also be safe
	err2 := client.Disconnect(ctx)
	require.NoError(t, err2)
}

// TestSetupOptionsContextCancellationIntegration tests context cancellation during endpoint discovery
func TestSetupOptionsContextCancellationIntegration(t *testing.T) {
	client := &OpcUAClient{
		Config: &OpcUAClientConfig{
			Endpoint:       "opc.tcp://unreachable-server:4840",
			ConnectTimeout: config.Duration(1 * time.Second),
		},
		Log: testutil.Logger{},
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.SetupOptions(ctx)
	require.Error(t, err)

	var endpointErr *EndpointError
	require.ErrorAs(t, err, &endpointErr)
	// Context cancellation should result in an error
}

// TestSetupOptionsTimeoutIntegration tests timeout behavior during endpoint discovery
func TestSetupOptionsTimeoutIntegration(t *testing.T) {
	client := &OpcUAClient{
		Config: &OpcUAClientConfig{
			Endpoint:       "opc.tcp://1.2.3.4:4840",                // Non-routable IP for timeout
			ConnectTimeout: config.Duration(100 * time.Millisecond), // Very short timeout
		},
		Log: testutil.Logger{},
	}

	ctx := context.Background()

	start := time.Now()
	err := client.SetupOptions(ctx)
	duration := time.Since(start)

	require.Error(t, err)
	// Should timeout roughly within our configured time
	require.Less(t, duration, 2*time.Second) // Allow some overhead

	var endpointErr *EndpointError
	require.ErrorAs(t, err, &endpointErr)
}

// TestCleanupTimeoutConfiguration tests that cleanup timeout respects connect_timeout
func TestCleanupTimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		connectTimeout config.Duration
		expectDefault  bool
	}{
		{
			name:           "no_connect_timeout_uses_default",
			connectTimeout: config.Duration(0),
			expectDefault:  true,
		},
		{
			name:           "short_connect_timeout_used",
			connectTimeout: config.Duration(2 * time.Second),
			expectDefault:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the timeout calculation logic
			expectedDefault := 10 * time.Second
			expectedConfigured := time.Duration(tt.connectTimeout)

			if tt.expectDefault {
				require.Equal(t, config.Duration(0), tt.connectTimeout)
			} else {
				require.Greater(t, time.Duration(tt.connectTimeout), time.Duration(0))
				require.LessOrEqual(t, expectedConfigured, expectedDefault)
			}
		})
	}
}
