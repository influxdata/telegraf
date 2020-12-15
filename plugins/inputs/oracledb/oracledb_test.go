package oracledb

import (
  "database/sql/driver"
  "strings"
  "testing"

  "github.com/DATA-DOG/go-sqlmock"
  "github.com/influxdata/telegraf/testutil"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
)

func escapeExpectedQuery(q string) string {
	q = strings.ReplaceAll(q, "(", "\\(")
	return strings.ReplaceAll(q, ")", "\\)")
}

func initMockDB(t *testing.T, expectedQuery query, expectedRows *sqlmock.Rows) *testutil.Accumulator {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unexpected error occured while opening stub db connection: %s", err)
	}

	var (
		oracleMetadataColumns = []string{"SERVER_HOST", "SERVICE_NAME", "INSTANCE_NAME", "DB_UNIQUE_NAME"}
		oracleMetadataValues  = []driver.Value{"TESTHOST", "TESTSERVICE", "TESTINSTANCE", "TESTDB"}
	)

	oracleMetadataQuery := "SELECT SYS_CONTEXT('USERENV', 'SERVER_HOST'), " +
		"SYS_CONTEXT('USERENV', 'SERVICE_NAME'), " +
		"SYS_CONTEXT('USERENV', 'INSTANCE_NAME'), " +
		"SYS_CONTEXT('USERENV', 'DB_UNIQUE_NAME') " +
		"FROM DUAL"

	oracleMetadataRows := sqlmock.NewRows(oracleMetadataColumns).
		AddRow(oracleMetadataValues...)

	mock.ExpectQuery(escapeExpectedQuery(oracleMetadataQuery)).WillReturnRows(oracleMetadataRows)
	mock.ExpectQuery(escapeExpectedQuery(expectedQuery.Sqlquery)).WillReturnRows(expectedRows)

	ora := OracleDB{
		Queries: []query{expectedQuery},
		DB:      db,
		Log:     testutil.Logger{},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(ora.Gather))

	if acc.NMetrics() > 0 {
		for i, tag := range oracleMetadataColumns {
			v := acc.TagValue(pluginName, strings.ToLower(tag))
			ok := v != ""
			assert.True(t, ok)
			assert.Equal(t, oracleMetadataValues[i], v)
		}
	}

	return &acc
}

func TestOracle_GatherBasic(t *testing.T) {
	q := query{
		Sqlquery: "SELECT 1 AS alive FROM dual",
	}
	r := sqlmock.NewRows([]string{"ALIVE"}).
		AddRow(1)

	acc := initMockDB(t, q, r)

	defaultTags := map[string]string{
		"server_host":    "TESTHOST",
		"service_name":   "TESTSERVICE",
		"instance_name":  "TESTINSTANCE",
		"db_unique_name": "TESTDB",
	}

	assert.True(t, acc.HasPoint(pluginName, defaultTags, "ALIVE", int64(1)))
}

func TestOracle_GatherBasicWithExtraTags(t *testing.T) {
	q := query{
		Sqlquery:   "SELECT 1 AS alive, 'tag1' as tag FROM dual",
		TagColumns: []string{"TAG"},
	}
	r := sqlmock.NewRows([]string{"ALIVE", "TAG"}).
		AddRow(1, "tag1")

	acc := initMockDB(t, q, r)

	defaultTags := map[string]string{
		"server_host":    "TESTHOST",
		"service_name":   "TESTSERVICE",
		"instance_name":  "TESTINSTANCE",
		"db_unique_name": "TESTDB",
	}

	defaultTags["TAG"] = "tag1"
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "ALIVE", int64(1)))
}

func TestOracle_GatherMultipleRowsOneFieldEachRow(t *testing.T) {
	q := query{
		Sqlquery:   "SELECT max(value) as field, tag FROM some_table",
		TagColumns: []string{"TAG"},
	}
	r := sqlmock.NewRows([]string{"FIELD", "TAG"}).
		AddRow(10, "tag1").
		AddRow(20, "tag2")

	acc := initMockDB(t, q, r)

	assert.Equal(t, uint64(2), acc.NMetrics())
	assert.Equal(t, 2, acc.NFields())

	defaultTags := map[string]string{
		"server_host":    "TESTHOST",
		"service_name":   "TESTSERVICE",
		"instance_name":  "TESTINSTANCE",
		"db_unique_name": "TESTDB",
	}
	defaultTags["TAG"] = "tag1"
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "FIELD", int64(10)))

	defaultTags["TAG"] = "tag2"
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "FIELD", int64(20)))
}

func TestOracle_GatherMultipleRowsTwoFieldsEachRow(t *testing.T) {
	q := query{
		Sqlquery:   "SELECT min(value) as field1, max(value) as field2, tag FROM some_table",
		TagColumns: []string{"TAG"},
	}
	r := sqlmock.NewRows([]string{"FIELD1", "FIELD2", "TAG"}).
		AddRow(10, 30, "tag1").
		AddRow(20, 40, "tag2")

	acc := initMockDB(t, q, r)

	assert.Equal(t, uint64(2), acc.NMetrics())
	assert.Equal(t, 4, acc.NFields())

	defaultTags := map[string]string{
		"server_host":    "TESTHOST",
		"service_name":   "TESTSERVICE",
		"instance_name":  "TESTINSTANCE",
		"db_unique_name": "TESTDB",
	}

	defaultTags["TAG"] = "tag1"
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "FIELD1", int64(10)))
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "FIELD2", int64(30)))

	defaultTags["TAG"] = "tag2"
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "FIELD1", int64(20)))
	assert.True(t, acc.HasPoint(pluginName, defaultTags, "FIELD2", int64(40)))
}

func TestOracle_GatherEmpty(t *testing.T) {
	q := query{
		Sqlquery: "SELECT some_column FROM some_table",
	}
	r := sqlmock.NewRows([]string{})

	acc := initMockDB(t, q, r)

	assert.Equal(t, uint64(0), acc.NMetrics())
	assert.Equal(t, 0, acc.NFields())
}
