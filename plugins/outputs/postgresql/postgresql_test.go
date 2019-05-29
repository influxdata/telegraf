package postgresql

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	oneMetric, _ := metric.New("m", map[string]string{"t": "tv"}, map[string]interface{}{"f": 1}, timestamp)
	noTags, _ := metric.New("m", nil, map[string]interface{}{"f": 1}, timestamp)
	testCases := []struct {
		desc                string
		input               []telegraf.Metric
		fieldsAsJSON        bool
		execs               []sql.Result
		expectedExecQueries []string
		execErrs            []error
		expectErr           string
	}{
		{
			desc:      "no metrics, no error",
			input:     []telegraf.Metric{},
			expectErr: "",
		}, {
			desc:      "metric table not cached, error on creating it",
			input:     []telegraf.Metric{oneMetric},
			execs:     []sql.Result{nil},
			execErrs:  []error{fmt.Errorf("error on first exec")},
			expectErr: "error on first exec",
		}, {
			desc:         "metric table not cached, gets cached, no tags, fields as json, error on insert",
			input:        []telegraf.Metric{noTags},
			fieldsAsJSON: true,
			execs:        []sql.Result{nil, nil},
			execErrs:     []error{nil, fmt.Errorf("error on batch insert")},
			expectErr:    "error on batch insert",
		}, {
			desc:         "metric table not cached, gets cached, has tags, json fields, all good",
			input:        []telegraf.Metric{oneMetric},
			fieldsAsJSON: true,
			execs:        []sql.Result{nil, nil},
			execErrs:     []error{nil, nil},
			expectedExecQueries: []string{
				`CREATE TABLE IF NOT EXISTS "a"."m"(time timestamptz,"t" text,fields jsonb)`,
				`INSERT INTO "a"."m"("time","t","fields") VALUES($1,$2,$3)`},
		}, {
			desc:     "metric table not cached, gets cached, has tags, all good",
			input:    []telegraf.Metric{oneMetric},
			execs:    []sql.Result{nil, nil},
			execErrs: []error{nil, nil},
			expectedExecQueries: []string{
				`CREATE TABLE IF NOT EXISTS "a"."m"(time timestamptz,"t" text,"f" int8)`,
				`INSERT INTO "a"."m"("time","t","f") VALUES($1,$2,$3)`},
		},
	}

	for _, testCase := range testCases {
		p := &Postgresql{
			tables:        &mockTk{tables: make(map[string]bool)},
			TableTemplate: "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS})",
			Schema:        "a",
			FieldsAsJsonb: testCase.fieldsAsJSON,
			db: &mockDb{
				exec:      testCase.execs,
				execErr:   testCase.execErrs,
				expectedQ: testCase.expectedExecQueries,
			}}
		err := p.Write(testCase.input)
		if testCase.expectErr != "" {
			assert.EqualError(t, err, testCase.expectErr, testCase.desc)
		} else {
			assert.Nil(t, err, testCase.desc)
		}
	}
}
func TestInsertBatches(t *testing.T) {
	sampleData := map[string][]*colsAndValues{
		"tab": {
			{
				cols: []string{"a"},
				vals: []interface{}{1},
			},
		},
	}

	testCases := []struct {
		input           map[string][]*colsAndValues
		desc            string
		resultsFromExec []sql.Result
		errorsFromExec  []error
		errorOnQuery    error
		fieldsAsJSON    bool
		expectErr       string
	}{
		{
			desc:           "no batches, no errors",
			input:          make(map[string][]*colsAndValues),
			errorsFromExec: []error{fmt.Errorf("should not have called exec")},
		}, {
			desc:            "error returned on first insert, fields as json",
			input:           sampleData,
			resultsFromExec: []sql.Result{nil},
			errorsFromExec:  []error{fmt.Errorf("error on first insert")},
			fieldsAsJSON:    true,
			expectErr:       "error on first insert",
		}, {
			desc:            "error returned on first insert, error on add column",
			input:           sampleData,
			resultsFromExec: []sql.Result{nil},
			errorsFromExec:  []error{fmt.Errorf("error on first insert")},
			errorOnQuery:    fmt.Errorf("error on query"),
			expectErr:       "error on query",
		}, {
			desc:            "no error on insert",
			input:           sampleData,
			resultsFromExec: []sql.Result{nil},
			errorsFromExec:  []error{nil},
		},
	}

	for _, testCase := range testCases {
		m := &mockDb{exec: testCase.resultsFromExec,
			execErr:  testCase.errorsFromExec,
			queryErr: testCase.errorOnQuery}
		p := &Postgresql{
			db:            m,
			FieldsAsJsonb: testCase.fieldsAsJSON,
		}

		err := p.insertBatches(testCase.input)
		if testCase.expectErr != "" {
			assert.EqualError(t, err, testCase.expectErr)
		} else {
			assert.Nil(t, err)
		}
	}
}

type mockDb struct {
	currentExec int
	exec        []sql.Result
	expectedQ   []string
	execErr     []error
	query       *sql.Rows
	queryErr    error
}

func (m *mockDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	tmp := m.currentExec
	m.currentExec++
	if m.expectedQ != nil && m.expectedQ[tmp] != query {
		return nil, fmt.Errorf("unexpected query, got: '%s' expected: %s", query, m.expectedQ[tmp])
	}

	return m.exec[tmp], m.execErr[tmp]
}
func (m *mockDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.query, m.queryErr
}
func (m *mockDb) QueryRow(query string, args ...interface{}) *sql.Row { return nil }
func (m *mockDb) Close() error                                        { return nil }

type mockTk struct {
	tables map[string]bool
}

func (m *mockTk) add(tableName string) {
	m.tables[tableName] = true
}

func (m *mockTk) exists(schema, table string) bool {
	_, exists := m.tables[table]
	return exists
}
