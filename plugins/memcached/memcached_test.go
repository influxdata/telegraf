package memcached

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemcachedGeneratesMetrics(t *testing.T) {
	m := &Memcached{
		Servers: []string{testutil.GetLocalHost()},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{"get_hits", "get_misses", "evictions", "limit_maxbytes", "bytes"}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric), metric)
	}
}
