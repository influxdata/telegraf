package mysql

import (
	"strings"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlGeneratesMetrics(t *testing.T) {
	m := &Mysql{
		Servers: []string{""},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	prefixes := []struct {
		prefix string
		count  int
	}{
		{"commands", 141},
		{"handler", 18},
		{"bytes", 2},
		{"innodb", 51},
		{"threads", 4},
	}

	intMetrics := []string{
		"queries",
		"slow_queries",
	}

	for _, prefix := range prefixes {
		var count int

		for _, p := range acc.Points {
			if strings.HasPrefix(p.Measurement, prefix.prefix) {
				count++
			}
		}

		assert.Equal(t, prefix.count, count)
	}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntValue(metric))
	}
}

func TestMysqlDefaultsToLocal(t *testing.T) {
	m := &Mysql{}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, len(acc.Points) > 0)
}
