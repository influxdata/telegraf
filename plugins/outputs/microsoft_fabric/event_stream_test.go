package microsoft_fabric

import (
	"strconv"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

// Helper function to create a test eventstream instance
func newTestEventStream(t *testing.T) *eventstream {
	e := &eventstream{
		log:     testutil.Logger{},
		timeout: config.Duration(time.Second * 5),
		options: azeventhubs.EventDataBatchOptions{},
	}
	err := e.Init()
	require.NoError(t, err)
	return e
}

func TestEventStream_Init(t *testing.T) {
	tests := []struct {
		name            string
		maxMessageSize  config.Size
		expectedMaxSize uint64
	}{
		{
			name:            "Init with default settings",
			maxMessageSize:  0,
			expectedMaxSize: 0,
		},
		{
			name:            "Init with custom max message size",
			maxMessageSize:  1024,
			expectedMaxSize: 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventstream{
				maxMessageSize: tt.maxMessageSize,
			}
			err := e.Init()
			require.NoError(t, err)
			require.NotNil(t, e.serializer, "serializer should be initialized")
			require.Equal(t, tt.expectedMaxSize, e.options.MaxBytes)
		})
	}
}

func TestEventStream_ParseConnectionString(t *testing.T) {
	tests := []struct {
		name          string
		connString    string
		expected      map[string]string
		expectedError string
	}{
		{
			name:       "Valid connection string with partition key and message size",
			connString: "endpoint=https://example.com;partitionkey=mykey;maxmessagesize=1024",
			expected: map[string]string{
				"partitionkey":   "mykey",
				"maxmessagesize": "1024",
			},
		},
		{
			name:          "Invalid connection string format",
			connString:    "invalid string format",
			expectedError: "invalid connection string format",
		},
		{
			name:       "Case insensitive keys",
			connString: "PARTITIONKEY=mykey;MaxMessageSize=1024",
			expected: map[string]string{
				"partitionkey":   "mykey",
				"maxmessagesize": "1024",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEventStream(t)
			err := e.parseconnectionString(tt.connString)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				if tt.expected["partitionkey"] != "" {
					require.Equal(t, tt.expected["partitionkey"], e.partitionKey)
				}
				if tt.expected["maxmessagesize"] != "" {
					size, err := strconv.ParseInt(tt.expected["maxmessagesize"], 10, 64)
					require.NoError(t, err)
					require.Equal(t, config.Size(size), e.maxMessageSize)
				}
			}
		})
	}
}
