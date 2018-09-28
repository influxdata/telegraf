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

func TestMysqlGetDSNTag(t *testing.T) {
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
			"localhost:3306",
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
		output := getDSNTag(test.input)
		if output != test.output {
			t.Errorf("Input: %s Expected %s, got %s\n", test.input, test.output, output)
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
			"tcp(127.0.0.1:3306)/?timeout=5s",
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
			"root:passwd@tcp(192.168.1.1:3306)/?timeout=10s&tls=false",
		},
		{
			"tcp(10.150.1.123:3306)/",
			"tcp(10.150.1.123:3306)/?timeout=5s",
		},
		{
			"root:@!~(*&$#%(&@#(@&#Password@tcp(10.150.1.123:3306)/",
			"root:@!~(*&$#%(&@#(@&#Password@tcp(10.150.1.123:3306)/?timeout=5s",
		},
		{
			"root:Test3a#@!@tcp(10.150.1.123:3306)/",
			"root:Test3a#@!@tcp(10.150.1.123:3306)/?timeout=5s",
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
		output    interface{}
		boolValue bool
	}{
		{sql.RawBytes("123"), int64(123), true},
		{sql.RawBytes("abc"), "abc", true},
		{sql.RawBytes("10.1"), 10.1, true},
		{sql.RawBytes("ON"), 1, true},
		{sql.RawBytes("OFF"), 0, true},
		{sql.RawBytes("NO"), 0, true},
		{sql.RawBytes("YES"), 1, true},
		{sql.RawBytes("No"), 0, true},
		{sql.RawBytes("Yes"), 1, true},
		{sql.RawBytes(""), nil, false},
	}
	for _, cases := range testCases {
		if got, ok := parseValue(cases.rawByte); got != cases.output && ok != cases.boolValue {
			t.Errorf("for %s wanted %t, got %t", string(cases.rawByte), cases.output, got)
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
