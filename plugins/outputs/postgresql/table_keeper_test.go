package postgresql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTableKeeper(t *testing.T) {
	mock := &mockWr{}
	tk := newTableKeeper(mock).(*defTableKeeper)
	assert.Equal(t, mock, tk.db)
	assert.Empty(t, tk.Tables)
}

func TestTableKeeperAdd(t *testing.T) {
	tk := newTableKeeper(nil).(*defTableKeeper)
	tk.add("table")
	tk.add("table2")
	assert.Equal(t, 2, len(tk.Tables))
	assert.True(t, tk.Tables["table"])
	assert.True(t, tk.Tables["table2"])
	assert.False(t, tk.Tables["table3"])
	tk.add("table2")
	assert.Equal(t, 2, len(tk.Tables))
}

func TestTableKeeperExists(t *testing.T) {
	mock := &mockWr{}
	tk := newTableKeeper(mock).(*defTableKeeper)
	table := "table name"

	// table cached
	tk.Tables[table] = true
	mock.execErr = fmt.Errorf("should not call execute")
	assert.True(t, tk.exists("", table))

	// error on table exists query
	mock.execErr = fmt.Errorf("error on query execute")
	mock.expected = tableExistsTemplate
	delete(tk.Tables, table)
	assert.False(t, tk.exists("", table))
	assert.Equal(t, 0, len(tk.Tables))

	// fetch from db, doesn't exist
	mock.execErr = nil
	mock.exec = &mockResult{}
	assert.False(t, tk.exists("", table))

	// fetch from db, exists
	mock.exec = &mockResult{rows: 1}
	assert.True(t, tk.exists("", table))
	assert.Equal(t, 1, len(tk.Tables))
	assert.True(t, tk.Tables[table])
}

type mockResult struct {
	rows    int64
	rowErr  error
	last    int64
	lastErr error
}

func (m *mockResult) LastInsertId() (int64, error) {
	return m.last, m.lastErr
}

func (m *mockResult) RowsAffected() (int64, error) {
	return m.rows, m.rowErr
}
