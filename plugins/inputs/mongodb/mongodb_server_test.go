//go:build integration
// +build integration

package mongodb

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestGetDefaultTags(t *testing.T) {
	var tagTests = []struct {
		in  string
		out string
	}{
		{"hostname", server.hostname},
	}
	defaultTags := server.getDefaultTags()
	for _, tt := range tagTests {
		if defaultTags[tt.in] != tt.out {
			t.Errorf("expected %q, got %q", tt.out, defaultTags[tt.in])
		}
	}
}

func TestAddDefaultStats(t *testing.T) {
	var acc testutil.Accumulator

	err := server.gatherData(&acc, false, true, true, true, []string{"local"})
	require.NoError(t, err)

	// need to call this twice so it can perform the diff
	err = server.gatherData(&acc, false, true, true, true, []string{"local"})
	require.NoError(t, err)

	for key := range defaultStats {
		require.True(t, acc.HasInt64Field("mongodb", key))
	}
}

func TestPoolStatsVersionCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		expectedCommand string
		err             bool
	}{
		{
			name:            "mongodb v3",
			version:         "3.0.0",
			expectedCommand: "shardConnPoolStats",
		},
		{
			name:            "mongodb v4",
			version:         "4.0.0",
			expectedCommand: "shardConnPoolStats",
		},
		{
			name:            "mongodb v5",
			version:         "5.0.0",
			expectedCommand: "connPoolStats",
		},
		{
			name:    "invalid version",
			version: "v4",
			err:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command, err := poolStatsCommand(test.version)
			require.Equal(t, test.expectedCommand, command)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
