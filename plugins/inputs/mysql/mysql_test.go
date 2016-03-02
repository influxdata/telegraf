package mysql

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	assert.True(t, acc.HasMeasurement("mysql"))
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

func TestMysqlDNSAddTimeout(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			"",
			"/?timeout=5s",
		},
		{
			"tcp(192.168.1.1:3306)/",
			"tcp(192.168.1.1:3306)/?timeout=5s",
		},
		{
			"root:passwd@tcp(192.168.1.1:3306)/?tls=false",
			"root:passwd@tcp(192.168.1.1:3306)/?timeout=5s&tls=false",
		},
		{
			"root:passwd@tcp(192.168.1.1:3306)/?tls=false&timeout=10s",
			"root:passwd@tcp(192.168.1.1:3306)/?tls=false&timeout=10s",
		},
	}

	for _, test := range tests {
		output, _ := dsnAddTimeout(test.input)
		if output != test.output {
			t.Errorf("Expected %s, got %s\n", test.output, output)
		}
	}
}

func TestParseValue(t *testing.T) {
	testCases := []struct {
		rawByte   sql.RawBytes
		value     float64
		boolValue bool
	}{
		{sql.RawBytes("Yes"), 1, true},
		{sql.RawBytes("No"), 0, false},
		{sql.RawBytes("ON"), 1, true},
		{sql.RawBytes("OFF"), 0, false},
		{sql.RawBytes("ABC"), 0, false},
	}
	for _, cases := range testCases {
		if value, ok := parseValue(cases.rawByte); value != cases.value && ok != cases.boolValue {
			t.Errorf("want %d with %t, got %d with %t", int(cases.value), cases.boolValue, int(value), ok)
		}
	}
}

func TestNewNamespace(t *testing.T) {
	testCases := []struct {
		words     []string
		namespace string
	}{
		{
			[]string{"thread", "info_scheme", "query update"},
			"thread_info_scheme_query_update",
		},
		{
			[]string{"thread", "info_scheme", "query_update"},
			"thread_info_scheme_query_update",
		},
		{
			[]string{"thread", "info", "scheme", "query", "update"},
			"thread_info_scheme_query_update",
		},
	}
	for _, cases := range testCases {
		if got := newNamespace(cases.words...); got != cases.namespace {
			t.Errorf("want %s, got %s", cases.namespace, got)
		}
	}
}
