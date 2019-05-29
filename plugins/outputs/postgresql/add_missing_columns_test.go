package postgresql

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareMissingColumnsQuery1(columns []string) string {
	var quotedColumns = make([]string, len(columns))
	for i, column := range columns {
		quotedColumns[i] = quoteLiteral(column)
	}
	return fmt.Sprintf(missingColumnsTemplate, strings.Join(quotedColumns, ","))
}

func TestPrepareMissingColumnsQuery(t *testing.T) {
	columns := []string{}
	assert.Equal(t, `WITH available AS (SELECT column_name as c FROM information_schema.columns WHERE table_schema = $1 and table_name = $2),`+
		`required AS (SELECT c FROM unnest(array []) AS c) `+
		`SELECT required.c, available.c IS NULL FROM required LEFT JOIN available ON required.c = available.c;`,
		prepareMissingColumnsQuery(columns))
	columns = []string{"a", "b", "c"}
	assert.Equal(t, `WITH available AS (SELECT column_name as c FROM information_schema.columns WHERE table_schema = $1 and table_name = $2),`+
		`required AS (SELECT c FROM unnest(array ['a','b','c']) AS c) `+
		`SELECT required.c, available.c IS NULL FROM required LEFT JOIN available ON required.c = available.c;`,
		prepareMissingColumnsQuery(columns))
}

func TestWhichColumnsAreMissing(t *testing.T) {
	mock := &mockWr{}
	p := &Postgresql{db: mock}

	columns := []string{"col1"}
	mock.queryErr = fmt.Errorf("error 1")
	mock.expected = prepareMissingColumnsQuery(columns)
	table := "tableName"
	_, err := p.whichColumnsAreMissing(columns, table)
	assert.Equal(t, err.Error(), "error 1")
}

func TestAddColumnToTable(t *testing.T) {
	mock := &mockWr{}
	p := &Postgresql{db: mock, Schema: "pub"}

	column := "col1"
	dataType := "text"
	tableName := "table"
	mock.execErr = fmt.Errorf("error 1")
	mock.expected = `ALTER TABLE "pub"."table" ADD COLUMN IF NOT EXISTS "col1" text;`
	err := p.addColumnToTable(column, dataType, tableName)
	assert.EqualError(t, err, "error 1")

	mock.execErr = nil
	assert.Nil(t, p.addColumnToTable(column, dataType, tableName))

}

func (p *Postgresql) addColumnToTable1(columnName, dataType, tableName string) error {
	fullTableName := p.fullTableName(tableName)
	addColumnQuery := fmt.Sprintf(addColumnTemplate, fullTableName, quoteIdent(columnName), dataType)
	_, err := p.db.Exec(addColumnQuery)
	return err
}

type mockWr struct {
	expected string
	exec     sql.Result
	execErr  error
	query    *sql.Rows
	queryErr error
}

func (m *mockWr) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.expected != query {
		return nil, fmt.Errorf("unexpected query; exp: '%s'; got: '%s'", m.expected, query)
	}
	return m.exec, m.execErr
}
func (m *mockWr) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if m.expected != query {
		return nil, fmt.Errorf("unexpected query; exp: '%s'; got: '%s'", m.expected, query)
	}
	return m.query, m.queryErr
}
func (m *mockWr) QueryRow(query string, args ...interface{}) *sql.Row {
	return nil
}
func (m *mockWr) Close() error { return nil }
