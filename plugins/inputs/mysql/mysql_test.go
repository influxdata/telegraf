package mysql

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/testutil"
)

const servicePort = "3306"

func TestMysqlDefaultsToLocalIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := testutil.Container{
		Image: "mysql",
		Env: map[string]string{
			"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
		},
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForLog("/usr/sbin/mysqld: ready for connections"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}

	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	m := &Mysql{
		Servers: []string{fmt.Sprintf("root@tcp(%s:%s)/", container.Address, container.Ports[servicePort])},
	}

	var acc testutil.Accumulator
	err = m.Gather(&acc)
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

	container := testutil.Container{
		Image: "mysql",
		Env: map[string]string{
			"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
		},
		ExposedPorts: []string{servicePort},
		WaitingFor: wait.ForAll(
			wait.ForLog("/usr/sbin/mysqld: ready for connections"),
			wait.ForListeningPort(nat.Port(servicePort)),
		),
	}

	err := container.Start()
	require.NoError(t, err, "failed to start container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating container failed")
	}()

	testServer := fmt.Sprintf("root@tcp(%s:%s)/?tls=false", container.Address, container.Ports[servicePort])
	m := &Mysql{
		Servers:          []string{testServer},
		IntervalSlow:     "30s",
		GatherGlobalVars: true,
		MetricVersion:    2,
	}

	var acc, acc2 testutil.Accumulator
	err = m.Gather(&acc)
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

	type fields []struct {
		key         string
		rawValue    string
		parsedValue interface{}
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
				{"__test__string_variable", "text", "text"},
				{"__test__int_variable", "5", int64(5)},
				{"__test__off_variable", "OFF", int64(0)},
				{"__test__on_variable", "ON", int64(1)},
				{"__test__empty_variable", "", nil},
			},
			tags{"server": "127.0.0.1:3306"},
		},
		{
			"version tag is present",
			fields{
				{"__test__string_variable", "text", "text"},
				{"version", "8.0.27-0ubuntu0.20.04.1", "8.0.27-0ubuntu0.20.04.1"},
			},
			tags{"server": "127.0.0.1:3306", "version": "8.0.27-0ubuntu0.20.04.1"},
		},

		{"", fields{{"delay_key_write", "OFF", "OFF"}}, nil},
		{"", fields{{"delay_key_write", "ON", "ON"}}, nil},
		{"", fields{{"delay_key_write", "ALL", "ALL"}}, nil},
		{"", fields{{"enforce_gtid_consistency", "OFF", "OFF"}}, nil},
		{"", fields{{"enforce_gtid_consistency", "ON", "ON"}}, nil},
		{"", fields{{"enforce_gtid_consistency", "WARN", "WARN"}}, nil},
		{"", fields{{"event_scheduler", "NO", "NO"}}, nil},
		{"", fields{{"event_scheduler", "YES", "YES"}}, nil},
		{"", fields{{"event_scheduler", "DISABLED", "DISABLED"}}, nil},
		{"", fields{{"have_ssl", "DISABLED", int64(0)}}, nil},
		{"", fields{{"have_ssl", "YES", int64(1)}}, nil},
		{"", fields{{"have_symlink", "NO", int64(0)}}, nil},
		{"", fields{{"have_symlink", "DISABLED", int64(0)}}, nil},
		{"", fields{{"have_symlink", "YES", int64(1)}}, nil},
		{"", fields{{"session_track_gtids", "OFF", "OFF"}}, nil},
		{"", fields{{"session_track_gtids", "OWN_GTID", "OWN_GTID"}}, nil},
		{"", fields{{"session_track_gtids", "ALL_GTIDS", "ALL_GTIDS"}}, nil},
		{"", fields{{"session_track_transaction_info", "OFF", "OFF"}}, nil},
		{"", fields{{"session_track_transaction_info", "STATE", "STATE"}}, nil},
		{"", fields{{"session_track_transaction_info", "CHARACTERISTICS", "CHARACTERISTICS"}}, nil},
		{"", fields{{"ssl_fips_mode", "0", "0"}}, nil}, // TODO: map this to OFF or vice versa using integers
		{"", fields{{"ssl_fips_mode", "1", "1"}}, nil}, // TODO: map this to ON or vice versa using integers
		{"", fields{{"ssl_fips_mode", "2", "2"}}, nil}, // TODO: map this to STRICT or vice versa using integers
		{"", fields{{"ssl_fips_mode", "OFF", "OFF"}}, nil},
		{"", fields{{"ssl_fips_mode", "ON", "ON"}}, nil},
		{"", fields{{"ssl_fips_mode", "STRICT", "STRICT"}}, nil},
		{"", fields{{"use_secondary_engine", "OFF", "OFF"}}, nil},
		{"", fields{{"use_secondary_engine", "ON", "ON"}}, nil},
		{"", fields{{"use_secondary_engine", "FORCED", "FORCED"}}, nil},
		{"", fields{{"transaction_write_set_extraction", "OFF", "OFF"}}, nil},
		{"", fields{{"transaction_write_set_extraction", "MURMUR32", "MURMUR32"}}, nil},
		{"", fields{{"transaction_write_set_extraction", "XXHASH64", "XXHASH64"}}, nil},
		{"", fields{{"slave_skip_errors", "OFF", "OFF"}}, nil},
		{"", fields{{"slave_skip_errors", "0", "0"}}, nil},
		{"", fields{{"slave_skip_errors", "1007,1008,1050", "1007,1008,1050"}}, nil},
		{"", fields{{"slave_skip_errors", "all", "all"}}, nil},
		{"", fields{{"slave_skip_errors", "ddl_exist_errors", "ddl_exist_errors"}}, nil},
		{"", fields{{"gtid_mode", "OFF", int64(0)}}, nil},
		{"", fields{{"gtid_mode", "OFF_PERMISSIVE", int64(0)}}, nil},
		{"", fields{{"gtid_mode", "ON", int64(1)}}, nil},
		{"", fields{{"gtid_mode", "ON_PERMISSIVE", int64(1)}}, nil},
	}

	for i, testCase := range testCases {
		if testCase.name == "" {
			testCase.name = fmt.Sprintf("#%d", i)
		}

		t.Run(testCase.name, func(t *testing.T) {
			rows := sqlmock.NewRows(columns)
			for _, field := range testCase.fields {
				rows.AddRow(field.key, field.rawValue)
			}

			mock.ExpectQuery(globalVariablesQuery).WillReturnRows(rows).RowsWillBeClosed()

			acc := &testutil.Accumulator{}

			err = m.gatherGlobalVariables(db, "test", acc)
			require.NoErrorf(t, err, "err on gatherGlobalVariables (test case %q)", testCase.name)

			foundFields := map[string]bool{}

			for _, metric := range acc.Metrics {
				require.Equalf(t, measurement, metric.Measurement, "wrong measurement (test case %q)", testCase.name)

				if testCase.tags != nil {
					require.Equalf(t, testCase.tags, tags(metric.Tags), "wrong tags (test case %q)", testCase.name)
				}

				for key, value := range metric.Fields {
					for _, field := range testCase.fields {
						if field.key == key {
							require.Falsef(t, foundFields[key], "field %s observed multiple times (test case %q)", key, testCase.name)
							require.Equalf(t, field.parsedValue, value, "wrong value for field %s (test case %q)", key, testCase.name)
							foundFields[key] = true
							break
						}
					}

					require.Truef(t, foundFields[key], "unexpected field %s=%v (test case %q)", key, value, testCase.name)
				}
			}

			for _, field := range testCase.fields {
				require.Truef(t, foundFields[field.key], "missing field %s=%v (test case %q)", field.key, field.parsedValue, testCase.name)
			}
		})
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
