package mysql

import (
	"strings"
	"testing"

	"github.com/influxdb/tivan/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlGeneratesMetrics(t *testing.T) {
	m := &Mysql{
		Servers: []*Server{
			{
				Address: "",
			},
		},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	prefixes := []struct {
		prefix string
		count  int
	}{
		{"mysql_commands", 141},
		{"mysql_handler", 18},
		{"mysql_bytes", 2},
		{"mysql_innodb", 51},
		{"mysql_threads", 4},
	}

	intMetrics := []string{
		"mysql_queries",
		"mysql_slow_queries",
	}

	for _, prefix := range prefixes {
		var count int

		for _, p := range acc.Points {
			if strings.HasPrefix(p.Name, prefix.prefix) {
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
