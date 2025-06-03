package microsoft_fabric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitFail(t *testing.T) {
	tests := []struct {
		name       string
		connection string
		expected   string
	}{
		{
			name:     "empty connection string",
			expected: "endpoint must not be empty",
		},
		{
			name:       "invalid connection string format",
			connection: "invalid=format",
			expected:   "invalid connection string",
		},
		{
			name:       "Malformed connection string",
			connection: "endpoint=;key=;",
			expected:   "invalid connection string",
		},
		{
			name:       "invalid eventhouse connection string",
			connection: "data source=https://example.kusto.windows.net;invalid_param",
			expected:   "parsing connection string failed",
		},
		{
			name:       "invalid eventhouse connection string format",
			connection: "invalid string format",
			expected:   "invalid connection string format",
		},
		{
			name:       "invalid eventhouse metrics grouping type",
			connection: "data source=https://example.com;metrics grouping type=Invalid",
			expected:   "metrics grouping type is not valid:Invalid",
		},
		{
			name:       "invalid eventhouse create tables value",
			connection: "data source=https://example.com;database=mydb;create tables=invalid",
			expected:   "invalid setting",
		},
		{
			name:       "invalid eventstream connection string",
			connection: "Endpoint=sb://namespace.servicebus.windows.net/;invalid_param",
			expected:   "parsing connection string failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin
			plugin := &MicrosoftFabric{
				ConnectionString: tt.connection,
				Log:              testutil.Logger{},
			}

			// Check the returned error
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestInitEventHouse(t *testing.T) {
	tests := []struct {
		name       string
		connection string
		timeout    config.Duration
		expected   adx.Config
	}{
		{
			name:       "valid configuration",
			connection: "data source=https://example.kusto.windows.net;Database=testdb",
			expected: adx.Config{
				Endpoint:     "https://example.kusto.windows.net",
				Database:     "testdb",
				CreateTables: true,
				Timeout:      config.Duration(30 * time.Second),
			},
		},
		{
			name:       "valid configuration with timeout",
			connection: "data source=https://example.kusto.windows.net;Database=testdb",
			timeout:    config.Duration(60 * time.Second),
			expected: adx.Config{
				Endpoint:     "https://example.kusto.windows.net",
				Database:     "testdb",
				CreateTables: true,
				Timeout:      config.Duration(60 * time.Second),
			},
		},
		{
			name:       "valid connection string with all parameters",
			connection: "data source=https://example.com;database=mydb;table name=mytable;create tables=true;metrics grouping type=tablepermetric",
			expected: adx.Config{
				Endpoint:        "https://example.com",
				Database:        "mydb",
				TableName:       "mytable",
				CreateTables:    true,
				MetricsGrouping: "tablepermetric",
			},
		},
		{
			name:       "case insensitive parameters",
			connection: "DATA SOURCE=https://example.com;DATABASE=mydb",
			expected: adx.Config{
				Endpoint: "https://example.com",
				Database: "mydb",
			},
		},
		{
			name:       "server parameter instead of data source",
			connection: "server=https://example.com;database=mydb",
			expected: adx.Config{
				Endpoint: "https://example.com",
				Database: "mydb",
			},
		},
		{
			name:       "create tables parameter true",
			connection: "data source=https://example.com;database=mydb;create tables=true",
			expected: adx.Config{
				Endpoint:     "https://example.com",
				Database:     "mydb",
				CreateTables: true,
			},
		},
		{
			name:       "create tables parameter false",
			connection: "data source=https://example.com;database=mydb;create tables=false",
			expected: adx.Config{
				Endpoint:     "https://example.com",
				Database:     "mydb",
				CreateTables: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin
			plugin := &MicrosoftFabric{
				ConnectionString: tt.connection,
				Timeout:          tt.timeout,
				Log:              testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Check the created plugin
			require.NotNil(t, plugin.activePlugin, "active plugin should have been set")
			ap, ok := plugin.activePlugin.(*eventhouse)
			require.Truef(t, ok, "expected evenhouse plugin but got %T", plugin.activePlugin)
			require.Equal(t, tt.expected, ap.Config)
		})
	}
}

func TestInitEventStream(t *testing.T) {
	tests := []struct {
		name       string
		connection string
		timeout    config.Duration
		expected   eventstream
	}{
		{
			name:       "valid connection",
			connection: "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=keyName;SharedAccessKey=key",
			expected: eventstream{
				connectionString: "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=keyName;SharedAccessKey=key",
				timeout:          config.Duration(30 * time.Second),
			},
		},
		{
			name:       "valid connection with timeout",
			connection: "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=keyName;SharedAccessKey=key",
			timeout:    config.Duration(60 * time.Second),
			expected: eventstream{
				connectionString: "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=keyName;SharedAccessKey=key",
				timeout:          config.Duration(30 * time.Second),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &MicrosoftFabric{
				ConnectionString: tt.connection,
				Timeout:          tt.timeout,
				Log:              testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Check the created plugin
			require.NotNil(t, plugin.activePlugin, "active plugin should have been set")
			ap, ok := plugin.activePlugin.(*eventstream)
			require.Truef(t, ok, "expected evenstream plugin but got %T", plugin.activePlugin)
			require.Equal(t, tt.expected.connectionString, ap.connectionString)
			require.Equal(t, tt.expected.timeout, ap.timeout)
			require.Equal(t, tt.expected.partitionKey, ap.partitionKey)
			require.Equal(t, tt.expected.maxMessageSize, ap.maxMessageSize)
		})
	}
}
