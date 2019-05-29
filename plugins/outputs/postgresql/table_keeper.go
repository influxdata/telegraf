package postgresql

import (
	"log"
)

const (
	tableExistsTemplate = "SELECT tablename FROM pg_tables WHERE tablename = $1 AND schemaname = $2;"
)

type tableKeeper interface {
	exists(schema, tableName string) bool
	add(tableName string)
}

type defTableKeeper struct {
	Tables map[string]bool
	db     dbWrapper
}

func newTableKeeper(db dbWrapper) tableKeeper {
	return &defTableKeeper{
		Tables: make(map[string]bool),
		db:     db,
	}
}

func (t *defTableKeeper) exists(schema, tableName string) bool {
	if _, ok := t.Tables[tableName]; ok {
		return true
	}

	result, err := t.db.Exec(tableExistsTemplate, tableName, schema)
	if err != nil {
		log.Printf("E! Error checking for existence of metric table %s: %v", tableName, err)
		return false
	}
	if count, _ := result.RowsAffected(); count == 1 {
		t.Tables[tableName] = true
		return true
	}
	return false
}

func (t *defTableKeeper) add(tableName string) {
	t.Tables[tableName] = true
}
