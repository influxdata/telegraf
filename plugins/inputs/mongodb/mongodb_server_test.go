// +build integration

package mongodb

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.True(t, acc.HasInt64Field("mongodb", key))
	}
}
