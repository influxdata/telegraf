package microsoft_fabric

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/testutil"
)

func TestEventHouseConnectSuccess(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		database string
	}{
		{
			name:     "valid configuration",
			endpoint: "addr=https://example.com",
			database: "testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &eventhouse{
				connectionString: tt.endpoint,
				Config: adx.Config{
					Database: tt.database,
				},
				log: testutil.Logger{},
			}
			require.NoError(t, plugin.init())

			// Check for successful connection and client creation
			require.NoError(t, plugin.Connect())
			require.NotNil(t, plugin.client)
			// Clean up resources
			require.NoError(t, plugin.Close())
		})
	}
}

func TestIsEventhouseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "data source prefix",
			endpoint: "data source=https://example.com",
		},
		{
			name:     "address prefix",
			endpoint: "address=https://example.com",
		},
		{
			name:     "network address prefix",
			endpoint: "network address=https://example.com",
		},
		{
			name:     "server prefix",
			endpoint: "server=https://example.com",
		},
		{
			name:     "case insensitive prefix",
			endpoint: "DATA SOURCE=https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, isEventhouseEndpoint(tt.endpoint))
		})
	}
}

func TestIsNotEventhouseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "Invalid prefix",
			endpoint: "invalid=https://example.com",
		},
		{
			name:     "Empty string",
			endpoint: "",
		},
		{
			name:     "Just URL",
			endpoint: "https://example.com",
		},
		{
			name:     "eventstream endpoint",
			endpoint: "Endpoint=sb://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.False(t, isEventhouseEndpoint(tt.endpoint))
		})
	}
}
