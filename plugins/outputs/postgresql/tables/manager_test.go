package tables

import (
	"errors"
	"testing"

	"github.com/influxdata/telegraf/plugins/outputs/postgresql/db"
	"github.com/influxdata/telegraf/plugins/outputs/postgresql/utils"
	"github.com/jackc/pgx"
	"github.com/stretchr/testify/assert"
)

type mockDb struct {
	exec    pgx.CommandTag
	execErr error
}

func (m *mockDb) Exec(query string, args ...interface{}) (pgx.CommandTag, error) {
	return m.exec, m.execErr
}
func (m *mockDb) DoCopy(fullTableName *pgx.Identifier, colNames []string, batch [][]interface{}) error {
	return nil
}
func (m *mockDb) Query(query string, args ...interface{}) (*pgx.Rows, error) {
	return nil, nil
}
func (m *mockDb) QueryRow(query string, args ...interface{}) *pgx.Row {
	return nil
}
func (m *mockDb) Close() error {
	return nil
}

func (m *mockDb) IsAlive() bool { return true }

func TestNewManager(t *testing.T) {
	db := &mockDb{}
	res := NewManager(db, "schema", "table template").(*defTableManager)
	assert.Equal(t, "table template", res.tableTemplate)
	assert.Equal(t, "schema", res.schema)
	assert.Equal(t, db, res.db)
}

func TestExists(t *testing.T) {
	testCases := []struct {
		desc  string
		in    string
		out   bool
		db    *mockDb
		cache map[string]bool
	}{
		{
			desc:  "table already cached",
			in:    "table",
			db:    &mockDb{execErr: errors.New("should not have called exec")},
			cache: map[string]bool{"table": true},
			out:   true,
		}, {
			desc:  "table not cached, error on check db",
			cache: map[string]bool{},
			in:    "table",
			db:    &mockDb{execErr: errors.New("error on exec")},
		}, {
			desc:  "table not cached, exists in db",
			cache: map[string]bool{},
			in:    "table",
			db:    &mockDb{exec: "0 1"},
			out:   true,
		}, {
			desc:  "table not cached, doesn't exist",
			cache: map[string]bool{},
			in:    "table",
			db:    &mockDb{exec: "0 0"},
			out:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			manager := &defTableManager{
				Tables: tc.cache,
				db:     tc.db,
			}

			got := manager.Exists(tc.in)
			assert.Equal(t, tc.out, got)
		})
	}
}

func TestCreateTable(t *testing.T) {
	testCases := []struct {
		desc     string
		inT      string
		inCD     *utils.TargetColumns
		db       db.Wrapper
		template string
		out      error
	}{
		{
			desc: "error on exec, no table cached",
			inT:  "t",
			inCD: &utils.TargetColumns{
				Names:     []string{"time", "t", "f"},
				Target:    map[string]int{"time": 0, "t": 1, "f": 2},
				DataTypes: []utils.PgDataType{"timestamptz", "text", "float8"},
				Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagColType, utils.FieldColType},
			},
			db:       &mockDb{execErr: errors.New("error on exec")},
			template: "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS}) ",
			out:      errors.New("error on exec"),
		}, {
			desc: "all good, table is cached",
			inT:  "t",
			inCD: &utils.TargetColumns{
				Names:     []string{"time", "t", "f"},
				Target:    map[string]int{"time": 0, "t": 1, "f": 2},
				DataTypes: []utils.PgDataType{"timestamptz", "text", "float8"},
				Roles:     []utils.ColumnRole{utils.TimeColType, utils.TagColType, utils.FieldColType},
			},
			db:       &mockDb{},
			template: "CREATE TABLE IF NOT EXISTS {TABLE}({COLUMNS}) ",
			out:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			manager := &defTableManager{
				Tables:        map[string]bool{},
				db:            tc.db,
				tableTemplate: tc.template,
			}
			got := manager.CreateTable(tc.inT, tc.inCD)
			assert.Equal(t, tc.out, got)
			if tc.out == nil {
				assert.True(t, manager.Tables[tc.inT])
			}
		})
	}
}
