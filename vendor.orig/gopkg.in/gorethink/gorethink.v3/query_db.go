package gorethink

import (
	p "gopkg.in/gorethink/gorethink.v3/ql2"
)

// DBCreate creates a database. A RethinkDB database is a collection of tables,
// similar to relational databases.
//
// Note: that you can only use alphanumeric characters and underscores for the
// database name.
func DBCreate(args ...interface{}) Term {
	return constructRootTerm("DBCreate", p.Term_DB_CREATE, args, map[string]interface{}{})
}

// DBDrop drops a database. The database, all its tables, and corresponding data
// will be deleted.
func DBDrop(args ...interface{}) Term {
	return constructRootTerm("DBDrop", p.Term_DB_DROP, args, map[string]interface{}{})
}

// DBList lists all database names in the system.
func DBList(args ...interface{}) Term {
	return constructRootTerm("DBList", p.Term_DB_LIST, args, map[string]interface{}{})
}
