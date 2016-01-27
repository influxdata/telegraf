// +build integration

package mongodb

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultTags(t *testing.T) {
	var tagTests = []struct {
		in  string
		out string
	}{
		{"hostname", server.Url.Host},
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

	err := server.gatherData(&acc)
	require.NoError(t, err)

	time.Sleep(time.Duration(1) * time.Second)
	// need to call this twice so it can perform the diff
	err = server.gatherData(&acc)
	require.NoError(t, err)

	for key, _ := range DefaultStats {
		assert.True(t, acc.HasIntValue(key))
	}
}
