package sip

import (
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

func TestInit_Defaults(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://sip.example.com:5060"},
	}

	err := plugin.Init()
	require.NoError(t, err)

	// Check defaults
	require.Equal(t, "OPTIONS", plugin.Method)
	require.Equal(t, "telegraf", plugin.FromUser)
	require.Equal(t, "Telegraf SIP Monitor", plugin.UserAgent)
	require.Equal(t, 200, plugin.ExpectCode)
	require.Equal(t, config.Duration(defaultTimeout), plugin.Timeout)
}

func TestInit_CustomValues(t *testing.T) {
	plugin := &SIP{
		Servers:    []string{"sip://sip.example.com:5061;transport=tcp"},
		Method:     "INVITE",
		FromUser:   "testuser",
		UserAgent:  "Test Agent",
		ExpectCode: 404,
	}

	err := plugin.Init()
	require.NoError(t, err)

	require.Equal(t, "INVITE", plugin.Method)
	require.Equal(t, "testuser", plugin.FromUser)
	require.Equal(t, "Test Agent", plugin.UserAgent)
	require.Equal(t, 404, plugin.ExpectCode)
}

func TestInit_NoServers(t *testing.T) {
	plugin := &SIP{
		Servers: nil,
	}

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one server must be specified")
}

func TestInit_InvalidMethod(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://sip.example.com:5060"},
		Method:  "INVALID",
	}

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid SIP method")
}

func TestInit_InvalidTransport(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://sip.example.com:5060;transport=invalid"},
	}

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid transport")
}

func TestInit_ValidMethods(t *testing.T) {
	validMethods := []string{"OPTIONS", "INVITE", "MESSAGE", "options", "invite", "message"}

	for _, method := range validMethods {
		t.Run(method, func(t *testing.T) {
			plugin := &SIP{
				Servers: []string{"sip://sip.example.com:5060"},
				Method:  method,
			}

			err := plugin.Init()
			require.NoError(t, err)
		})
	}
}

func TestInit_ValidTransports(t *testing.T) {
	validTransports := []struct {
		name string
		url  string
	}{
		{"udp", "sip://sip.example.com:5060"},
		{"tcp", "sip://sip.example.com:5060;transport=tcp"},
		{"tls", "sips://sip.example.com:5061"},
		{"ws", "sip://sip.example.com:5060;transport=ws"},
		{"wss", "sips://sip.example.com:5061;transport=wss"},
	}

	for _, tt := range validTransports {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Servers: []string{tt.url},
			}

			err := plugin.Init()
			require.NoError(t, err)
		})
	}
}

func TestParseServer(t *testing.T) {
	tests := []struct {
		name          string
		server        string
		wantHost      string
		wantPort      int
		wantTransport string
		wantSecure    bool
		expectError   bool
	}{
		{
			name:          "sip:// URL defaults to UDP",
			server:        "sip://sip.example.com:5060",
			wantHost:      "sip.example.com",
			wantPort:      5060,
			wantTransport: "udp",
			wantSecure:    false,
		},
		{
			name:          "sips:// URL defaults to TLS",
			server:        "sips://sip.example.com:5061",
			wantHost:      "sip.example.com",
			wantPort:      5061,
			wantTransport: "tls",
			wantSecure:    true,
		},
		{
			name:          "sip:// with explicit TCP transport",
			server:        "sip://sip.example.com:5060;transport=tcp",
			wantHost:      "sip.example.com",
			wantPort:      5060,
			wantTransport: "tcp",
			wantSecure:    false,
		},
		{
			name:          "sip:// without port defaults to 5060",
			server:        "sip://sip.example.com",
			wantHost:      "sip.example.com",
			wantPort:      5060,
			wantTransport: "udp",
			wantSecure:    false,
		},
		{
			name:          "sips:// without port defaults to 5061",
			server:        "sips://secure.example.com",
			wantHost:      "secure.example.com",
			wantPort:      5061,
			wantTransport: "tls",
			wantSecure:    true,
		},
		{
			name:          "IP address with port",
			server:        "sip://192.168.1.100:5070",
			wantHost:      "192.168.1.100",
			wantPort:      5070,
			wantTransport: "udp",
			wantSecure:    false,
		},
		{
			name:        "missing scheme",
			server:      "sip.example.com:5060",
			expectError: true,
		},
		{
			name:        "invalid transport",
			server:      "sip://sip.example.com:5060;transport=invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseServer(tt.server)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantHost, info.Host)
			require.Equal(t, tt.wantPort, info.Port)
			require.Equal(t, tt.wantTransport, info.Transport)
			require.Equal(t, tt.wantSecure, info.Secure)
		})
	}
}

func TestBuildSIPURI(t *testing.T) {
	tests := []struct {
		name               string
		host               string
		port               int
		transport          string
		secure             bool
		fromUser           string
		toUser             string
		wantScheme         string
		wantUser           string
		wantHost           string
		wantPort           int
		wantTransportParam bool
	}{
		{
			name:               "UDP transport (default for sip)",
			host:               "sip.example.com",
			port:               5060,
			transport:          "udp",
			secure:             false,
			fromUser:           "alice",
			toUser:             "bob",
			wantScheme:         "sip",
			wantUser:           "bob",
			wantHost:           "sip.example.com",
			wantPort:           5060,
			wantTransportParam: false, // UDP is default for sip:, no param needed
		},
		{
			name:               "TCP transport",
			host:               "sip.example.com",
			port:               5060,
			transport:          "tcp",
			secure:             false,
			fromUser:           "alice",
			toUser:             "bob",
			wantScheme:         "sip",
			wantUser:           "bob",
			wantHost:           "sip.example.com",
			wantPort:           5060,
			wantTransportParam: true, // TCP is non-default for sip:
		},
		{
			name:               "TLS transport (default for sips)",
			host:               "sip.example.com",
			port:               5061,
			transport:          "tls",
			secure:             true,
			fromUser:           "alice",
			toUser:             "bob",
			wantScheme:         "sips",
			wantUser:           "bob",
			wantHost:           "sip.example.com",
			wantPort:           5061,
			wantTransportParam: false, // TLS is default for sips:, no param needed
		},
		{
			name:       "Empty ToUser uses FromUser",
			host:       "sip.example.com",
			port:       5060,
			transport:  "udp",
			secure:     false,
			fromUser:   "alice",
			toUser:     "",
			wantScheme: "sip",
			wantUser:   "alice",
			wantHost:   "sip.example.com",
			wantPort:   5060,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				FromUser: tt.fromUser,
				ToUser:   tt.toUser,
			}

			uri := plugin.buildSIPURI(tt.host, tt.port, tt.transport, tt.secure)

			require.Equal(t, tt.wantScheme, uri.Scheme)
			require.Equal(t, tt.wantUser, uri.User)
			require.Equal(t, tt.wantHost, uri.Host)
			require.Equal(t, tt.wantPort, uri.Port)

			if tt.wantTransportParam {
				require.NotNil(t, uri.UriParams)
			}
		})
	}
}

func TestSetResult(t *testing.T) {
	tests := []struct {
		name           string
		result         string
		expectedCode   int
		expectedResult string
	}{
		{"success", "success", 0, "success"},
		{"response_code_mismatch", "response_code_mismatch", 1, "response_code_mismatch"},
		{"timeout", "timeout", 2, "timeout"},
		{"connection_refused", "connection_refused", 3, "connection_refused"},
		{"connection_failed", "connection_failed", 4, "connection_failed"},
		{"unknown_result", "unknown_result", 99, "unknown_result"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := make(map[string]any)
			tags := make(map[string]string)

			setResult(tt.result, fields, tags)

			require.Equal(t, tt.expectedResult, tags["result"])
			require.Equal(t, tt.expectedResult, fields["result_type"])
			require.Equal(t, tt.expectedCode, fields["result_code"])
		})
	}
}

func TestStartStop(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://sip.example.com:5060"},
	}

	err := plugin.Init()
	require.NoError(t, err)

	acc := &testutil.Accumulator{}

	// Test Start
	err = plugin.Start(acc)
	require.NoError(t, err)

	// Ensure user agent is created
	require.NotNil(t, plugin.ua)

	// Test Stop
	plugin.Stop()

	// Ensure user agent is cleaned up
	require.Nil(t, plugin.ua)
}

func TestGather_NotInitialized(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://sip.example.com:5060"},
	}

	err := plugin.Init()
	require.NoError(t, err)

	acc := &testutil.Accumulator{}

	// Try to gather without starting
	err = plugin.Gather(acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not initialized")
}

// TestTLSConfiguration tests TLS configuration for secure transports
func TestTLSConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		serverURL          string
		insecureSkipVerify bool
		expectedScheme     string
		shouldInitTLS      bool
	}{
		{
			name:           "TLS transport (sips://)",
			serverURL:      "sips://sip.example.com:5061",
			expectedScheme: "sips",
			shouldInitTLS:  true,
		},
		{
			name:               "TLS with skip verify",
			serverURL:          "sips://sip.example.com:5061",
			insecureSkipVerify: true,
			expectedScheme:     "sips",
			shouldInitTLS:      true,
		},
		{
			name:           "WSS transport",
			serverURL:      "sips://sip.example.com:5061;transport=wss",
			expectedScheme: "sips",
			shouldInitTLS:  true,
		},
		{
			name:           "UDP transport (no TLS)",
			serverURL:      "sip://sip.example.com:5060",
			expectedScheme: "sip",
			shouldInitTLS:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &SIP{
				Servers:    []string{tt.serverURL},
				Method:     "OPTIONS",
				Timeout:    config.Duration(5 * time.Second),
				FromUser:   "telegraf",
				UserAgent:  "Test Agent",
				ExpectCode: 200,
				Log:        testutil.Logger{},
			}

			plugin.ClientConfig.InsecureSkipVerify = tt.insecureSkipVerify

			err := plugin.Init()
			require.NoError(t, err)

			if tt.shouldInitTLS {
				// TLS config should be initialized for TLS/WSS transports
				// Note: tlsConfig might still be nil if no certs are provided,
				// but that's okay - the transport layer will use defaults
				if tt.insecureSkipVerify {
					require.NotNil(t, plugin.tlsConfig)
					require.Equal(t, tt.insecureSkipVerify, plugin.tlsConfig.InsecureSkipVerify)
				}
			} else {
				require.Nil(t, plugin.tlsConfig)
			}

			// Parse server to get transport info
			info, err := parseServer(tt.serverURL)
			require.NoError(t, err)

			uri := plugin.buildSIPURI(info.Host, info.Port, info.Transport, info.Secure)
			require.Equal(t, tt.expectedScheme, uri.Scheme)
		})
	}
}

// TestTLSServerName tests SNI configuration
func TestTLSServerName(t *testing.T) {
	plugin := &SIP{
		Servers:    []string{"sips://192.168.1.100:5061"},
		Method:     "OPTIONS",
		Timeout:    config.Duration(5 * time.Second),
		FromUser:   "telegraf",
		UserAgent:  "Test Agent",
		ExpectCode: 200,
		Log:        testutil.Logger{},
	}

	plugin.ClientConfig.ServerName = "sip.example.com"

	err := plugin.Init()
	require.NoError(t, err)
	require.NotNil(t, plugin.tlsConfig)
	require.Equal(t, "sip.example.com", plugin.tlsConfig.ServerName)
}

func TestSIPDefaults(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://127.0.0.1:5060"},
		Log:     testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	require.Equal(t, "OPTIONS", plugin.Method)
	require.Equal(t, defaultTimeout, time.Duration(plugin.Timeout))
	require.Equal(t, defaultFromUser, plugin.FromUser)
	require.Equal(t, defaultUserAgent, plugin.UserAgent)
	require.Equal(t, defaultExpectCode, plugin.ExpectCode)
}

func TestSIPInvalidMethod(t *testing.T) {
	plugin := &SIP{
		Servers: []string{"sip://127.0.0.1:5060"},
		Method:  "INVALID",
		Log:     testutil.Logger{},
	}

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid SIP method")
}

func TestSIPNoServers(t *testing.T) {
	plugin := &SIP{
		Servers: nil,
		Log:     testutil.Logger{},
	}

	err := plugin.Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "at least one server must be specified")
}
