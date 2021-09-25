package sqlserver

import (
	"net/url"
	"strings"
)

const (
	emptySQLInstance  = "<empty-sql-instance>"
	emptyDatabaseName = "<empty-database-name>"
)

// getConnectionIdentifiers returns the sqlInstance and databaseName from the given connection string.
// The name of the SQL instance is returned as-is in the connection string
// If the connection string could not be parsed or sqlInstance/databaseName were not present, a placeholder value is returned
func getConnectionIdentifiers(connectionString string) (sqlInstance string, databaseName string) {
	if len(connectionString) == 0 {
		return emptySQLInstance, emptyDatabaseName
	}

	trimmedConnectionString := strings.TrimSpace(connectionString)

	if strings.HasPrefix(trimmedConnectionString, "odbc:") {
		connectionStringWithoutOdbc := strings.TrimPrefix(trimmedConnectionString, "odbc:")
		return parseConnectionStringKeyValue(connectionStringWithoutOdbc)
	}
	if strings.HasPrefix(trimmedConnectionString, "sqlserver://") {
		return parseConnectionStringURL(trimmedConnectionString)
	}
	return parseConnectionStringKeyValue(trimmedConnectionString)
}

// parseConnectionStringKeyValue parses a "key=value;" connection string and returns the SQL instance and database name
func parseConnectionStringKeyValue(connectionString string) (sqlInstance string, databaseName string) {
	sqlInstance = ""
	databaseName = ""

	keyValuePairs := strings.Split(connectionString, ";")
	for _, keyValuePair := range keyValuePairs {
		if len(keyValuePair) == 0 {
			continue
		}

		keyAndValue := strings.SplitN(keyValuePair, "=", 2)
		key := strings.TrimSpace(strings.ToLower(keyAndValue[0]))
		if len(key) == 0 {
			continue
		}

		value := ""
		if len(keyAndValue) > 1 {
			value = strings.TrimSpace(keyAndValue[1])
		}
		if strings.EqualFold("server", key) {
			sqlInstance = value
			continue
		}
		if strings.EqualFold("database", key) {
			databaseName = value
		}
	}

	if sqlInstance == "" {
		sqlInstance = emptySQLInstance
	}
	if databaseName == "" {
		databaseName = emptyDatabaseName
	}

	return sqlInstance, databaseName
}

// parseConnectionStringURL parses a URL-formatted connection string and returns the SQL instance and database name
func parseConnectionStringURL(connectionString string) (sqlInstance string, databaseName string) {
	sqlInstance = emptySQLInstance
	databaseName = emptyDatabaseName

	u, err := url.Parse(connectionString)
	if err != nil {
		return emptySQLInstance, emptyDatabaseName
	}

	sqlInstance = u.Hostname()

	if len(u.Path) > 1 {
		// There was a SQL instance name specified in addition to the host
		// E.g. "the.host.com:1234/InstanceName" or "the.host.com/InstanceName"
		sqlInstance = sqlInstance + "\\" + u.Path[1:]
	}

	query := u.Query()
	for key, value := range query {
		if strings.EqualFold("database", key) {
			databaseName = value[0]
			break
		}
	}

	return sqlInstance, databaseName
}
