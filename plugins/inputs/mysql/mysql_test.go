package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestMysqlDefaultsToLocalIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Mysql{
		Servers: []string{fmt.Sprintf("root@tcp(%s:3306)/", testutil.GetLocalHost())},
	}

	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("mysql"))
}

func TestMysqlMultipleInstancesIntegration(t *testing.T) {
	// Invoke Gather() from two separate configurations and
	//  confirm they don't interfere with each other
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testServer := "root@tcp(127.0.0.1:3306)/?tls=false"
	m := &Mysql{
		Servers:          []string{testServer},
		IntervalSlow:     "30s",
		GatherGlobalVars: true,
		MetricVersion:    2,
	}

	var acc, acc2 testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)
	require.True(t, acc.HasMeasurement("mysql"))
	// acc should have global variables
	require.True(t, acc.HasMeasurement("mysql_variables"))

	m2 := &Mysql{
		Servers:       []string{testServer},
		MetricVersion: 2,
	}
	err = m2.Gather(&acc2)
	require.NoError(t, err)
	require.True(t, acc2.HasMeasurement("mysql"))
	// acc2 should not have global variables
	require.False(t, acc2.HasMeasurement("mysql_variables"))
}

func TestMysqlMultipleInits(t *testing.T) {
	m := &Mysql{
		IntervalSlow: "30s",
	}
	m2 := &Mysql{}

	m.InitMysql()
	require.True(t, m.initDone)
	require.False(t, m2.initDone)
	require.Equal(t, m.scanIntervalSlow, uint32(30))
	require.Equal(t, m2.scanIntervalSlow, uint32(0))

	m2.InitMysql()
	require.True(t, m.initDone)
	require.True(t, m2.initDone)
	require.Equal(t, m.scanIntervalSlow, uint32(30))
	require.Equal(t, m2.scanIntervalSlow, uint32(0))
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
