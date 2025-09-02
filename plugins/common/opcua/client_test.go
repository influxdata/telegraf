package opcua

import (
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

func TestOpcUAClientConfig_ValidateEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		config    OpcUAClientConfig
		expectErr bool
		errType   error
	}{
		{
			name: "valid configuration",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "None",
				SecurityMode:   "None",
			},
			expectErr: false,
		},
		{
			name: "empty endpoint",
			config: OpcUAClientConfig{
				Endpoint: "",
			},
			expectErr: true,
			errType:   ErrInvalidEndpoint,
		},
		{
			name: "invalid endpoint URL",
			config: OpcUAClientConfig{
				Endpoint: "://invalid-url",
			},
			expectErr: true,
			errType:   ErrInvalidEndpoint,
		},
		{
			name: "invalid security policy",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "InvalidPolicy",
				SecurityMode:   "None",
			},
			expectErr: true,
			errType:   ErrInvalidSecurityPolicy,
		},
		{
			name: "invalid security mode",
			config: OpcUAClientConfig{
				Endpoint:       "opc.tcp://localhost:4840",
				SecurityPolicy: "None",
				SecurityMode:   "InvalidMode",
			},
			expectErr: true,
			errType:   ErrInvalidSecurityMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateEndpoint()
			if tt.expectErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.errType)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOpcUAClient_SetupWorkarounds(t *testing.T) {
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
