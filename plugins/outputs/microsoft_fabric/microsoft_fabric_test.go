package microsoft_fabric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name             string
		connectionString string
		timeout          config.Duration
		expectPlugin     string // "eventstream" or "eventhouse"
		expectError      bool
		errorContains    string
		initFunc         func(*MicrosoftFabric) // For custom initialization if needed
	}{
		{
			name:             "Empty connection string",
			connectionString: "",
			expectError:      true,
			errorContains:    "endpoint must not be empty",
		},
		{
			name:             "Valid EventStream connection",
			connectionString: "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=keyName;SharedAccessKey=key",
			timeout:          config.Duration(30 * time.Second),
			expectPlugin:     "eventstream",
		},
		{
			name:             "Valid EventHouse connection",
			connectionString: "data source=https://example.kusto.windows.net;Database=db",
			timeout:          config.Duration(30 * time.Second),
			expectPlugin:     "eventhouse",
		},
		{
			name:             "Invalid connection string format",
			connectionString: "invalid=format",
			expectError:      true,
			errorContains:    "invalid connection string",
		},
		{
			name:             "EventStream connection string parsing error",
			connectionString: "Endpoint=sb://namespace.servicebus.windows.net/;invalid_param",
			expectError:      true,
			errorContains:    "parsing connection string failed",
			initFunc: func(mf *MicrosoftFabric) {
				mf.eventstream = &eventstream{}
			},
		},
		{
			name:             "EventHouse connection string parsing error",
			connectionString: "data source=https://example.kusto.windows.net;invalid_param",
			expectError:      true,
			errorContains:    "parsing connection string failed",
			initFunc: func(mf *MicrosoftFabric) {
				mf.eventhouse = &eventhouse{}
			},
		},
		{
			name:             "Malformed connection string",
			connectionString: "endpoint=;key=;",
			expectError:      true,
			errorContains:    "invalid connection string",
		},
		{
			name:             "EventStream with custom timeout",
			connectionString: "Endpoint=sb://namespace.servicebus.windows.net/;SharedAccessKeyName=keyName;SharedAccessKey=key",
			timeout:          config.Duration(60 * time.Second),
			expectPlugin:     "eventstream",
			initFunc: func(mf *MicrosoftFabric) {
				mf.Timeout = config.Duration(60 * time.Second)
			},
		},
		{
			name:             "EventHouse with database configuration",
			connectionString: "data source=https://example.kusto.windows.net;Database=testdb",
			timeout:          config.Duration(30 * time.Second),
			expectPlugin:     "eventhouse",
			initFunc: func(mf *MicrosoftFabric) {
				mf.eventhouse = &eventhouse{
					Config: adx.Config{
						Database: "testdb",
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := &MicrosoftFabric{
				ConnectionString: tt.connectionString,
				Log:              testutil.Logger{},
				Timeout:          tt.timeout,
			}

			// Apply custom initialization if provided
			if tt.initFunc != nil {
				tt.initFunc(mf)
			}

			err := mf.Init()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, mf.activePlugin, "Active plugin should be set")

			// Verify correct plugin type was selected
			switch tt.expectPlugin {
			case "eventstream":
				require.NotNil(t, mf.eventstream, "EventStream should be initialized")
				require.Equal(t, mf.eventstream, mf.activePlugin)
			case "eventhouse":
				require.NotNil(t, mf.eventhouse, "EventHouse should be initialized")
				require.Equal(t, mf.eventhouse, mf.activePlugin)
			}

			// Verify timeout was properly set
			if tt.timeout > 0 {
				switch p := mf.activePlugin.(type) {
				case *eventstream:
					require.Equal(t, tt.timeout, p.timeout)
				case *eventhouse:
					require.Equal(t, tt.timeout, p.Timeout)
				}
			}
		})
	}
}
