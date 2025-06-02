package microsoft_fabric

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/testutil"
)

// Helper function to create a test eventhouse instance
func newTestEventHouse(t *testing.T) *eventhouse {
	e := &eventhouse{
		log: testutil.Logger{},
	}
	err := e.Init()
	require.NoError(t, err)
	e.Config = adx.Config{}
	return e
}

func TestEventHouse_Init(t *testing.T) {
	e := &eventhouse{}
	err := e.Init()
	require.NoError(t, err)
	require.NotNil(t, e.serializer, "serializer should be initialized")
	require.True(t, e.CreateTables, "CreateTables should be true by default")
}

func TestEventHouse_ParseConnectionString(t *testing.T) {
	tests := []struct {
		name          string
		connString    string
		expected      *eventhouse
		expectedError string
	}{
		{
			name:       "Valid connection string with all parameters",
			connString: "data source=https://example.com;database=mydb;table name=mytable;create tables=true;metrics grouping type=tablepermetric",
			expected: &eventhouse{
				Config: adx.Config{
					Endpoint:        "https://example.com",
					Database:        "mydb",
					TableName:       "mytable",
					CreateTables:    true,
					MetricsGrouping: "tablepermetric",
				},
			},
		},
		{
			name:          "Invalid connection string format",
			connString:    "invalid string format",
			expectedError: "invalid connection string format",
		},
		{
			name:       "Case insensitive parameters",
			connString: "DATA SOURCE=https://example.com;DATABASE=mydb",
			expected: &eventhouse{
				Config: adx.Config{
					Endpoint: "https://example.com",
					Database: "mydb",
				},
			},
		},
		{
			name:       "Server parameter instead of data source",
			connString: "server=https://example.com;database=mydb",
			expected: &eventhouse{
				Config: adx.Config{
					Endpoint: "https://example.com",
					Database: "mydb",
				},
			},
		},
		{
			name:          "Invalid metrics grouping type",
			connString:    "data source=https://example.com;metrics grouping type=Invalid",
			expectedError: "metrics grouping type is not valid:Invalid",
		},
		{
			name:       "Create tables parameter true",
			connString: "data source=https://example.com;database=mydb;create tables=true",
			expected: &eventhouse{
				Config: adx.Config{
					Endpoint:     "https://example.com",
					Database:     "mydb",
					CreateTables: true,
				},
			},
		},
		{
			name:       "Create tables parameter false",
			connString: "data source=https://example.com;database=mydb;create tables=false",
			expected: &eventhouse{
				Config: adx.Config{
					Endpoint:     "https://example.com",
					Database:     "mydb",
					CreateTables: false,
				},
			},
		},
		{
			name:          "Invalid create tables value",
			connString:    "data source=https://example.com;database=mydb;create tables=invalid",
			expectedError: "invalid setting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEventHouse(t)
			err := e.parseconnectionString(tt.connString)
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.expected != nil {
					require.Equal(t, tt.expected.Endpoint, e.Endpoint)
					require.Equal(t, tt.expected.Database, e.Database)
					require.Equal(t, tt.expected.TableName, e.TableName)
					require.Equal(t, tt.expected.CreateTables, e.CreateTables)
					require.Equal(t, tt.expected.MetricsGrouping, e.MetricsGrouping)
				}
			}
		})
	}
}

func TestEventHouse_Connect(t *testing.T) {
	tests := []struct {
		name          string
		endpoint      string
		database      string
		expectError   bool
		errorContains string
	}{
		{
			name:     "Valid configuration",
			endpoint: "https://example.com",
			database: "testdb",
		},
		{
			name:          "Empty endpoint",
			endpoint:      "",
			database:      "testdb",
			expectError:   true,
			errorContains: "endpoint configuration cannot be empty",
		},
		{
			name:          "Empty database",
			endpoint:      "https://example.com",
			database:      "",
			expectError:   true,
			errorContains: "database configuration cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEventHouse(t)
			e.Endpoint = tt.endpoint
			e.Database = tt.database

			err := e.Connect()
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, e.client)
			}
		})
	}
}

func TestIsEventhouseEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{
			name:     "Valid data source prefix",
			endpoint: "data source=https://example.com",
			want:     true,
		},
		{
			name:     "Valid address prefix",
			endpoint: "address=https://example.com",
			want:     true,
		},
		{
			name:     "Valid network address prefix",
			endpoint: "network address=https://example.com",
			want:     true,
		},
		{
			name:     "Valid server prefix",
			endpoint: "server=https://example.com",
			want:     true,
		},
		{
			name:     "Invalid prefix",
			endpoint: "invalid=https://example.com",
			want:     false,
		},
		{
			name:     "Empty string",
			endpoint: "",
			want:     false,
		},
		{
			name:     "Just URL",
			endpoint: "https://example.com",
			want:     false,
		},
		{
			name:     "Case insensitive prefix",
			endpoint: "DATA SOURCE=https://example.com",
			want:     true, // isEventhouseEndpoint is not case sensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEventhouseEndpoint(tt.endpoint)
			require.Equal(t, tt.want, got)
		})
	}
}
