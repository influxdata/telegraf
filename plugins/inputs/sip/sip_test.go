package sip

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestSampleConfig(t *testing.T) {
	plugin := &SIP{}
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitDefaults(t *testing.T) {
	plugin := &SIP{
		Server:  "sip://sip.example.com:5060",
		Timeout: config.Duration(5 * time.Second),
	}

	require.NoError(t, plugin.Init())

	// Check defaults
	require.Equal(t, "OPTIONS", plugin.Method)
	require.Equal(t, "telegraf", plugin.FromUser)
	require.Equal(t, "telegraf", plugin.ToUser)
	require.Equal(t, "udp", plugin.Transport)
}

func TestInitCustomValues(t *testing.T) {
	plugin := &SIP{
		Server:    "sip://sip.example.com:5061",
		Transport: "tcp",
		Method:    "INVITE",
		FromUser:  "testuser",
		ToUser:    "recipient",
		Timeout:   config.Duration(5 * time.Second),
	}

	require.NoError(t, plugin.Init())

	require.Equal(t, "INVITE", plugin.Method)
	require.Equal(t, "testuser", plugin.FromUser)
	require.Equal(t, "recipient", plugin.ToUser)
	require.Equal(t, "tcp", plugin.Transport)
}

func TestInitErrors(t *testing.T) {
	tests := []struct {
		name        string
		plugin      *SIP
		expectedErr string
	}{
		{
			name:        "no_server",
			plugin:      &SIP{},
			expectedErr: "server must be specified",
		},
		{
			name: "invalid_method",
			plugin: &SIP{
				Server:  "sip://sip.example.com:5060",
				Method:  "INVALID",
				Timeout: config.Duration(5 * time.Second),
			},
			expectedErr: "invalid SIP method",
		},
		{
			name: "invalid_transport",
			plugin: &SIP{
				Server:    "sip://sip.example.com:5060",
				Transport: "invalid",
				Timeout:   config.Duration(5 * time.Second),
			},
			expectedErr: "invalid transport",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestInitValidMethods(t *testing.T) {
	validMethods := []string{"OPTIONS", "INVITE", "MESSAGE", "options", "invite", "message"}

	for _, method := range validMethods {
		t.Run(method, func(t *testing.T) {
			plugin := &SIP{
				Server:  "sip://sip.example.com:5060",
				Method:  method,
				Timeout: config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, strings.ToUpper(method), plugin.Method)
		})
	}
}

func TestInitValidTransports(t *testing.T) {
	validTransports := []struct {
		name      string
		server    string
		transport string
	}{
		{"udp", "sip://sip.example.com:5060", "udp"},
		{"tcp", "sip://sip.example.com:5060", "tcp"},
		{"ws", "sip://sip.example.com:5060", "ws"},
		{"wss", "sips://sip.example.com:5061", "wss"},
	}

	for _, tt := range validTransports {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Server:    tt.server,
				Transport: tt.transport,
				Timeout:   config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, tt.transport, plugin.Transport)
		})
	}
}

func TestParseServer(t *testing.T) {
	tests := []struct {
		name      string
		server    string
		transport string
		want      serverInfo
	}{
		{
			name:      "sip:// URL with UDP transport",
			server:    "sip://sip.example.com:5060",
			transport: "udp",
			want:      serverInfo{host: "sip.example.com", port: 5060, transport: "udp", secure: false},
		},
		{
			name:      "sips:// URL with TCP transport",
			server:    "sips://sip.example.com:5061",
			transport: "tcp",
			want:      serverInfo{host: "sip.example.com", port: 5061, transport: "tcp", secure: true},
		},
		{
			name:      "sip:// with TCP transport",
			server:    "sip://sip.example.com:5060",
			transport: "tcp",
			want:      serverInfo{host: "sip.example.com", port: 5060, transport: "tcp", secure: false},
		},
		{
			name:      "sip:// without port defaults to 5060",
			server:    "sip://sip.example.com",
			transport: "udp",
			want:      serverInfo{host: "sip.example.com", port: 5060, transport: "udp", secure: false},
		},
		{
			name:      "sips:// without port defaults to 5061",
			server:    "sips://secure.example.com",
			transport: "tcp",
			want:      serverInfo{host: "secure.example.com", port: 5061, transport: "tcp", secure: true},
		},
		{
			name:      "IP address with port",
			server:    "sip://192.168.1.100:5070",
			transport: "udp",
			want:      serverInfo{host: "192.168.1.100", port: 5070, transport: "udp", secure: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Server:    tt.server,
				Transport: tt.transport,
				Timeout:   config.Duration(5 * time.Second),
			}

			require.NoError(t, plugin.Init())
			require.Equal(t, tt.want, *plugin.serverInfo)
		})
	}
}

func TestStartStop(t *testing.T) {
	plugin := &SIP{
		Server:  "sip://sip.example.com:5060",
		Timeout: config.Duration(5 * time.Second),
	}

	require.NoError(t, plugin.Init())

	acc := &testutil.Accumulator{}

	// Test Start
	require.NoError(t, plugin.Start(acc))

	// Ensure user agent is created
	require.NotNil(t, plugin.ua)

	// Test Stop
	plugin.Stop()

	// Ensure user agent is cleaned up
	require.Nil(t, plugin.ua)
}

func TestTLSConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		server             string
		transport          string
		insecureSkipVerify bool
		shouldInitTLS      bool
	}{
		{
			name:          "TLS via sips:// scheme",
			server:        "sips://sip.example.com:5061",
			transport:     "tcp",
			shouldInitTLS: true,
		},
		{
			name:               "TLS with skip verify",
			server:             "sips://sip.example.com:5061",
			transport:          "tcp",
			insecureSkipVerify: true,
			shouldInitTLS:      true,
		},
		{
			name:          "WSS transport",
			server:        "sips://sip.example.com:5061",
			transport:     "wss",
			shouldInitTLS: true,
		},
		{
			name:          "UDP transport (no TLS)",
			server:        "sip://sip.example.com:5060",
			transport:     "udp",
			shouldInitTLS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Server:    tt.server,
				Transport: tt.transport,
				Method:    "OPTIONS",
				Timeout:   config.Duration(5 * time.Second),
				FromUser:  "telegraf",
				Log:       testutil.Logger{},
			}

			plugin.ClientConfig.InsecureSkipVerify = tt.insecureSkipVerify

			require.NoError(t, plugin.Init())

			if tt.shouldInitTLS {
				if tt.insecureSkipVerify {
					require.NotNil(t, plugin.tlsConfig)
					require.Equal(t, tt.insecureSkipVerify, plugin.tlsConfig.InsecureSkipVerify)
				}
			} else {
				require.Nil(t, plugin.tlsConfig)
			}
		})
	}
}

func TestTLSServerName(t *testing.T) {
	plugin := &SIP{
		Server:    "sips://192.168.1.100:5061",
		Transport: "tcp",
		Method:    "OPTIONS",
		Timeout:   config.Duration(5 * time.Second),
		FromUser:  "telegraf",
		Log:       testutil.Logger{},
	}

	plugin.ClientConfig.ServerName = "sip.example.com"

	require.NoError(t, plugin.Init())
	require.NotNil(t, plugin.tlsConfig)
	require.Equal(t, "sip.example.com", plugin.tlsConfig.ServerName)
}

func TestInitRejectsDeprecatedTLSTransport(t *testing.T) {
	plugin := &SIP{
		Server:    "sip://sip.example.com:5060",
		Transport: "tls",
		Timeout:   config.Duration(5 * time.Second),
	}

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid transport")
}
