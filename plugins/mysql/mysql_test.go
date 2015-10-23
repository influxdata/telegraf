package mysql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMysqlGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Mysql{
		Servers: []string{fmt.Sprintf("root@tcp(%s:3306)/", testutil.GetLocalHost())},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	prefixes := []struct {
		prefix string
		count  int
	}{
		{"commands", 147},
		{"handler", 18},
		{"bytes", 2},
		{"innodb", 51},
		{"threads", 4},
		{"aborted", 2},
		{"created", 3},
		{"key", 7},
		{"open", 7},
		{"opened", 3},
		{"qcache", 8},
		{"table", 5},
	}

	intMetrics := []string{
		"queries",
		"slow_queries",
		"connections",
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
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Mysql{
		Servers: []string{fmt.Sprintf("root@tcp(%s:3306)/", testutil.GetLocalHost())},
	}

	var acc testutil.Accumulator

	err := m.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, len(acc.Points) > 0)
}

func TestMysqlParseDSN(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			"",
			"127.0.0.1:3306",
		},
		{
			"localhost",
			"127.0.0.1:3306",
		},
		{
			"127.0.0.1",
			"127.0.0.1:3306",
		},
		{
			"tcp(192.168.1.1:3306)/",
			"192.168.1.1:3306",
		},
		{
			"tcp(localhost)/",
			"localhost",
		},
		{
			"root:passwd@tcp(192.168.1.1:3306)/?tls=false",
			"192.168.1.1:3306",
		},
		{
			"root@tcp(127.0.0.1:3306)/?tls=false",
			"127.0.0.1:3306",
		},
		{
			"root:passwd@tcp(localhost:3036)/dbname?allowOldPasswords=1",
			"localhost:3036",
		},
		{
			"root:foo@bar@tcp(192.1.1.1:3306)/?tls=false",
			"192.1.1.1:3306",
		},
		{
			"root:f00@b4r@tcp(192.1.1.1:3306)/?tls=false",
			"192.1.1.1:3306",
		},
		{
			"root:fl!p11@tcp(192.1.1.1:3306)/?tls=false",
			"192.1.1.1:3306",
		},
	}

	for _, test := range tests {
		output, _ := parseDSN(test.input)
		if output != test.output {
			t.Errorf("Expected %s, got %s\n", test.output, output)
		}
	}
}
