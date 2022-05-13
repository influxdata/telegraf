package mysql

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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
	require.Empty(t, acc.Errors)

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
	require.Empty(t, acc.Errors)
	require.True(t, acc.HasMeasurement("mysql"))
	// acc should have global variables
	require.True(t, acc.HasMeasurement("mysql_variables"))

	m2 := &Mysql{
		Servers:       []string{testServer},
		MetricVersion: 2,
	}
	err = m2.Gather(&acc2)
	require.NoError(t, err)
	require.Empty(t, acc.Errors)
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

func TestGatherGlobalVariables(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	m := Mysql{
		Log:           testutil.Logger{},
		MetricVersion: 2,
	}
	m.InitMysql()

	columns := []string{"Variable_name", "Value"}
	measurement := "mysql_variables"

	type fields []*struct {
		key         string
		rawValue    string
		parsedValue interface{}
		observed    bool
	}
	type tags map[string]string
	testCases := []struct {
		name   string
		fields fields
		tags   tags
	}{
		{
			"basic variables",
			fields{
				{"__test__string_variable", "text", "text", false},
				{"__test__int_variable", "5", int64(5), false},
				{"__test__off_variable", "OFF", int64(0), false},
				{"__test__on_variable", "ON", int64(1), false},
				{"__test__empty_variable", "", nil, false},
			},
			tags{"server": "127.0.0.1:3306"},
		},
		{
			"version tag is present",
			fields{
				{"__test__string_variable", "text", "text", false},
				{"version", "8.0.27-0ubuntu0.20.04.1", "8.0.27-0ubuntu0.20.04.1", false},
			},
			tags{"server": "127.0.0.1:3306", "version": "8.0.27-0ubuntu0.20.04.1"},
		},

		{"", fields{{"delay_key_write", "OFF", "OFF", false}}, nil},
		{"", fields{{"delay_key_write", "ON", "ON", false}}, nil},
		{"", fields{{"delay_key_write", "ALL", "ALL", false}}, nil},
		{"", fields{{"enforce_gtid_consistency", "OFF", "OFF", false}}, nil},
		{"", fields{{"enforce_gtid_consistency", "ON", "ON", false}}, nil},
		{"", fields{{"enforce_gtid_consistency", "WARN", "WARN", false}}, nil},
		{"", fields{{"event_scheduler", "NO", "NO", false}}, nil},
		{"", fields{{"event_scheduler", "YES", "YES", false}}, nil},
		{"", fields{{"event_scheduler", "DISABLED", "DISABLED", false}}, nil},
		{"", fields{{"have_ssl", "DISABLED", int64(0), false}}, nil},
		{"", fields{{"have_ssl", "YES", int64(1), false}}, nil},
		{"", fields{{"have_symlink", "NO", int64(0), false}}, nil},
		{"", fields{{"have_symlink", "DISABLED", int64(0), false}}, nil},
		{"", fields{{"have_symlink", "YES", int64(1), false}}, nil},
		{"", fields{{"session_track_gtids", "OFF", "OFF", false}}, nil},
		{"", fields{{"session_track_gtids", "OWN_GTID", "OWN_GTID", false}}, nil},
		{"", fields{{"session_track_gtids", "ALL_GTIDS", "ALL_GTIDS", false}}, nil},
		{"", fields{{"session_track_transaction_info", "OFF", "OFF", false}}, nil},
		{"", fields{{"session_track_transaction_info", "STATE", "STATE", false}}, nil},
		{"", fields{{"session_track_transaction_info", "CHARACTERISTICS", "CHARACTERISTICS", false}}, nil},
		{"", fields{{"ssl_fips_mode", "0", "0", false}}, nil}, // TODO: map this to OFF or vice versa using integers
		{"", fields{{"ssl_fips_mode", "1", "1", false}}, nil}, // TODO: map this to ON or vice versa using integers
		{"", fields{{"ssl_fips_mode", "2", "2", false}}, nil}, // TODO: map this to STRICT or vice versa using integers
		{"", fields{{"ssl_fips_mode", "OFF", "OFF", false}}, nil},
		{"", fields{{"ssl_fips_mode", "ON", "ON", false}}, nil},
		{"", fields{{"ssl_fips_mode", "STRICT", "STRICT", false}}, nil},
		{"", fields{{"use_secondary_engine", "OFF", "OFF", false}}, nil},
		{"", fields{{"use_secondary_engine", "ON", "ON", false}}, nil},
		{"", fields{{"use_secondary_engine", "FORCED", "FORCED", false}}, nil},
		{"", fields{{"transaction_write_set_extraction", "OFF", "OFF", false}}, nil},
		{"", fields{{"transaction_write_set_extraction", "MURMUR32", "MURMUR32", false}}, nil},
		{"", fields{{"transaction_write_set_extraction", "XXHASH64", "XXHASH64", false}}, nil},
		{"", fields{{"slave_skip_errors", "OFF", "OFF", false}}, nil},
		{"", fields{{"slave_skip_errors", "0", "0", false}}, nil},
		{"", fields{{"slave_skip_errors", "1007,1008,1050", "1007,1008,1050", false}}, nil},
		{"", fields{{"slave_skip_errors", "all", "all", false}}, nil},
		{"", fields{{"slave_skip_errors", "ddl_exist_errors", "ddl_exist_errors", false}}, nil},
		{"", fields{{"gtid_mode", "OFF", int64(0), false}}, nil},
		{"", fields{{"gtid_mode", "OFF_PERMISSIVE", int64(0), false}}, nil},
		{"", fields{{"gtid_mode", "ON", int64(1), false}}, nil},
		{"", fields{{"gtid_mode", "ON_PERMISSIVE", int64(1), false}}, nil},
	}

	for i, testCase := range testCases {
		if testCase.name == "" {
			testCase.name = fmt.Sprintf("#%d", i)
		}

		rows := sqlmock.NewRows(columns)
		for _, field := range testCase.fields {
			rows.AddRow(field.key, field.rawValue)
		}

		mock.ExpectQuery(globalVariablesQuery).WillReturnRows(rows).RowsWillBeClosed()

		acc := &testutil.Accumulator{}

		err = m.gatherGlobalVariables(db, "test", acc)
		if !assert.NoErrorf(t, err, "err on gatherGlobalVariables (test case %q)", testCase.name) {
			continue
		}

		for _, metric := range acc.Metrics {
			assert.Equalf(t, measurement, metric.Measurement, "wrong measurement (test case %q)", testCase.name)

			if testCase.tags != nil {
				assert.Equalf(t, testCase.tags, tags(metric.Tags), "wrong tags (test case %q)", testCase.name)
			}

			for key, value := range metric.Fields {
				foundField := false

				for _, field := range testCase.fields {
					if field.key == key {
						assert.Falsef(t, field.observed, "field %s observed multiple times (test case %q)", key, testCase.name)
						assert.Equalf(t, field.parsedValue, value, "wrong value for field %s (test case %q)", key, testCase.name)
						field.observed = true
						foundField = true
						break
					}
				}

				if !assert.Truef(t, foundField, "unexpected field %s=%v (test case %q)", key, value, testCase.name) {
					continue
				}
			}
		}

		for _, field := range testCase.fields {
			assert.Truef(t, field.observed, "missing field %s=%v (test case %q)", field.key, field.parsedValue, testCase.name)
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
