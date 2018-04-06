package proxysql

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type mockDatabase struct {
	queryErr error
	rows     *mockDatabaseRows
}

func (d *mockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (databaseRows, error) {
	return d.rows, d.queryErr
}

type mockDatabaseRows struct {
	scanErr error
	values  [][]interface{}
}

func (r *mockDatabaseRows) Scan(dest ...interface{}) error {
	if r.scanErr != nil {
		return r.scanErr
	}

	// Pop the value off the list and use it
	value := r.values[0]
	r.values = r.values[1:]
	if len(value) != len(dest) {
		panic(fmt.Errorf("invalid number of args given"))
	}

	for i := range dest {
		src := reflect.ValueOf(value[i])
		// Dest is gonna be a pointer
		dest := reflect.ValueOf(dest[i]).Elem()
		dest.Set(src)
	}
	return nil
}

func (r *mockDatabaseRows) Next() bool {
	return len(r.values) > 0
}

func (r *mockDatabaseRows) Close() error {
	return nil
}

func newMockDatabaseReturn(queryErr, scanErr error, values ...[]interface{}) *mockDatabase {
	return &mockDatabase{
		queryErr: queryErr,
		rows: &mockDatabaseRows{
			scanErr: scanErr,
			values:  values,
		},
	}
}

func TestProxySQLGatherGlobalStats(t *testing.T) {
	tests := []struct {
		description string
		queryErr    error
		scanErr     error
		values      [][]interface{}
		defaultTags map[string]string
		expErr      error
		expFields   map[string]interface{}
		expTags     map[string]string
	}{
		{
			description: "query fails",
			queryErr:    fmt.Errorf("failure"),
			expErr:      fmt.Errorf("failure"),
		},
		{
			description: "scan fails",
			values: [][]interface{}{
				{
					"some_counter",
					int64(2018),
				},
				{
					"some_other_counter",
					int64(20180),
				},
			},
			scanErr: fmt.Errorf("scan failure"),
			expErr:  fmt.Errorf("scan failure"),
		},
		{
			description: "some fields added",
			values: [][]interface{}{
				{
					"some_counter",
					int64(2018),
				},
				{
					"some_other_counter",
					int64(20180),
				},
			},
			defaultTags: map[string]string{
				"server": "127.0.0.1:3306",
			},
			expFields: map[string]interface{}{
				"some_counter":       int64(2018),
				"some_other_counter": int64(20180),
			},
			expTags: map[string]string{
				"server": "127.0.0.1:3306",
			},
		},
		{
			description: "21 fields added",
			values: func() [][]interface{} {
				vals := [][]interface{}{}
				for i := 0; i < 22; i++ {
					vals = append(vals, []interface{}{
						fmt.Sprintf("field%d", i),
						int64(i),
					})
				}
				return vals
			}(),

			defaultTags: map[string]string{
				"server": "127.0.0.1:3306",
			},
			expFields: func() map[string]interface{} {
				fields := map[string]interface{}{}
				for i := 0; i < 22; i++ {
					fields[fmt.Sprintf("field%d", i)] = int64(i)
				}
				return fields
			}(),
			expTags: map[string]string{
				"server": "127.0.0.1:3306",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			db := newMockDatabaseReturn(tc.queryErr, tc.scanErr, tc.values...)
			p := &ProxySQL{}
			acc := &testutil.Accumulator{}
			require.Equal(t, tc.expErr, p.gatherGlobalStats(context.Background(), db, acc, tc.defaultTags))
			if tc.expErr == nil {
				acc.AssertContainsTaggedFields(t, "proxysql", tc.expFields, tc.expTags)
			}
		})
	}
}

func TestProxySQLGatherConnectionPoolStats(t *testing.T) {
	tests := []struct {
		description string
		queryErr    error
		scanErr     error
		values      [][]interface{}
		defaultTags map[string]string
		expErr      error
		expFields   []map[string]interface{}
		expTags     []map[string]string
	}{
		{
			description: "query fails",
			queryErr:    fmt.Errorf("failure"),
			expErr:      fmt.Errorf("failure"),
		},
		{
			description: "scan fails",
			values: [][]interface{}{
				{
					"0",
					"127.0.0.1",
					"3306",
					"ONLINE",
					int64(1),
					int64(0),
					int64(12),
					int64(4),
					int64(10),
					int64(1289389),
					int64(1023012),
				},
			},
			scanErr: fmt.Errorf("scan failure"),
			expErr:  fmt.Errorf("scan failure"),
		},
		{
			description: "some fields added",
			values: [][]interface{}{
				{
					"0",
					"127.0.0.2",
					"3306",
					"ONLINE",
					int64(1),
					int64(0),
					int64(12),
					int64(4),
					int64(10),
					int64(1289389),
					int64(1023012),
				},
				{
					"1",
					"127.0.0.2",
					"3306",
					"ONLINE",
					int64(2),
					int64(0),
					int64(12),
					int64(4),
					int64(100),
					int64(128129389),
					int64(102302312),
				},
			},
			defaultTags: map[string]string{
				"server": "127.0.0.1:3306",
			},
			expFields: []map[string]interface{}{
				{
					"connections_used": int64(1),
					"connections_free": int64(0),
					"connections_ok":   int64(12),
					"connections_err":  int64(4),
					"queries":          int64(10),
					"bytes_sent":       int64(1289389),
					"bytes_received":   int64(1023012),
				},
				{
					"connections_used": int64(2),
					"connections_free": int64(0),
					"connections_ok":   int64(12),
					"connections_err":  int64(4),
					"queries":          int64(100),
					"bytes_sent":       int64(128129389),
					"bytes_received":   int64(102302312),
				},
			},
			expTags: []map[string]string{
				{
					"server":         "127.0.0.1:3306",
					"hostgroup":      "0",
					"hostgroup_host": "127.0.0.2:3306",
					"status":         "ONLINE",
				},
				{
					"server":         "127.0.0.1:3306",
					"hostgroup":      "1",
					"hostgroup_host": "127.0.0.2:3306",
					"status":         "ONLINE",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			db := newMockDatabaseReturn(tc.queryErr, tc.scanErr, tc.values...)
			p := &ProxySQL{}
			acc := &testutil.Accumulator{}
			require.Equal(t, tc.expErr, p.gatherConnectionPoolStats(context.Background(), db, acc, tc.defaultTags))
			if tc.expErr == nil {
				for i := range tc.expFields {
					acc.AssertContainsTaggedFields(t, "proxysql_connection_pool", tc.expFields[i], tc.expTags[i])
				}
			}
		})
	}
}

func TestProxySQLGatherCommandCounterStats(t *testing.T) {
	tests := []struct {
		description string
		queryErr    error
		scanErr     error
		values      [][]interface{}
		defaultTags map[string]string
		expErr      error
		expFields   []map[string]interface{}
		expTags     []map[string]string
	}{
		{
			description: "query fails",
			queryErr:    fmt.Errorf("failure"),
			expErr:      fmt.Errorf("failure"),
		},
		{
			description: "scan fails",
			values: [][]interface{}{
				{
					"SELECT",
					int64(120),
					int64(76),
					int64(0),
					int64(1),
					int64(5),
					int64(0),
					int64(60),
					int64(0),
					int64(10),
					int64(0),
					int64(0),
					int64(0),
					int64(0),
					int64(0),
				},
			},
			scanErr: fmt.Errorf("scan failure"),
			expErr:  fmt.Errorf("scan failure"),
		},
		{
			description: "some fields added",
			values: [][]interface{}{
				{
					"SELECT",
					int64(120),
					int64(76),
					int64(0),
					int64(1),
					int64(5),
					int64(0),
					int64(60),
					int64(0),
					int64(10),
					int64(0),
					int64(0),
					int64(0),
					int64(0),
					int64(0),
				},
				{
					"UPDATE",
					int64(12005),
					int64(396),
					int64(0),
					int64(1),
					int64(5),
					int64(0),
					int64(60),
					int64(0),
					int64(100),
					int64(0),
					int64(230),
					int64(0),
					int64(0),
					int64(0),
				},
			},
			defaultTags: map[string]string{
				"server": "127.0.0.1:3306",
			},
			expFields: []map[string]interface{}{
				{
					"total_time":  int64(120),
					"count_total": int64(76),
					"count_100us": int64(0),
					"count_500us": int64(1),
					"count_1ms":   int64(5),
					"count_5ms":   int64(0),
					"count_10ms":  int64(60),
					"count_50ms":  int64(0),
					"count_100ms": int64(10),
					"count_500ms": int64(0),
					"count_1s":    int64(0),
					"count_5s":    int64(0),
					"count_10s":   int64(0),
					"count_inf":   int64(0),
				},
				{
					"total_time":  int64(12005),
					"count_total": int64(396),
					"count_100us": int64(0),
					"count_500us": int64(1),
					"count_1ms":   int64(5),
					"count_5ms":   int64(0),
					"count_10ms":  int64(60),
					"count_50ms":  int64(0),
					"count_100ms": int64(100),
					"count_500ms": int64(0),
					"count_1s":    int64(230),
					"count_5s":    int64(0),
					"count_10s":   int64(0),
					"count_inf":   int64(0),
				},
			},
			expTags: []map[string]string{
				{
					"server":  "127.0.0.1:3306",
					"command": "SELECT",
				},
				{
					"server":  "127.0.0.1:3306",
					"command": "UPDATE",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			db := newMockDatabaseReturn(tc.queryErr, tc.scanErr, tc.values...)
			p := &ProxySQL{}
			acc := &testutil.Accumulator{}
			require.Equal(t, tc.expErr, p.gatherCommandCounterStats(context.Background(), db, acc, tc.defaultTags))
			if tc.expErr == nil {
				for i := range tc.expFields {
					acc.AssertContainsTaggedFields(t, "proxysql_commands", tc.expFields[i], tc.expTags[i])
				}
			}
		})
	}
}
