package sqlserver

import (
	"net"
	"net/url"
	"strings"
)

const emptyServerName = "<empty-server-name>"
const emptyDatabaseName = "<empty-database-name>"
const odbcPrefix = "odbc:"

// getConnectionIdentifiers returns the sqlInstance and databaseName from the given connection string.
// The name of the server is returned as-is in the connection string
// If the connection string could not be parsed or sqlInstance/databaseName were not present, a placeholder value is returned
func getConnectionIdentifiers(connectionString string) (sqlInstance string, databaseName string) {
	if len(connectionString) == 0 {
		return
	}

	trimmedConnectionString := strings.TrimSpace(connectionString)

	if strings.HasPrefix(trimmedConnectionString, odbcPrefix) {
		i := strings.Index(trimmedConnectionString, odbcPrefix)
		connectionStringWithoutOdbc := trimmedConnectionString[i+len(odbcPrefix):]
		sqlInstance, databaseName = parseConnectionStringKeyValue(connectionStringWithoutOdbc)
	} else if strings.HasPrefix(trimmedConnectionString, "sqlserver://") {
		sqlInstance, databaseName = parseConnectionStringURL(trimmedConnectionString)
	} else {
		sqlInstance, databaseName = parseConnectionStringKeyValue(trimmedConnectionString)
	}

	return
}

// parseConnectionStringKeyValue parses a "key=value;" connection string and returns the SQL instance and database name
func parseConnectionStringKeyValue(connectionString string) (sqlInstance string, databaseName string) {
	sqlInstance = emptyServerName
	databaseName = emptyDatabaseName

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

		var value string = ""
		if len(keyAndValue) > 1 {
			value = strings.TrimSpace(keyAndValue[1])
		}

		if isInstanceIdentifier(key) {
			sqlInstance = value
			if databaseName != emptyDatabaseName {
				break
			}
		} else if isDatabaseIdentifier(key) {
			databaseName = value
			if sqlInstance != emptyServerName {
				break
			}
		}
	}

	return
}

// parseConnectionStringURL parses a URL-formatted connection string and returns the SQL instance and database name
func parseConnectionStringURL(connectionString string) (sqlInstance string, databaseName string) {
	sqlInstance = emptyServerName
	databaseName = emptyDatabaseName

	u, err := url.Parse(connectionString)
	if err != nil {
		return
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		// The host did not have a port
		sqlInstance = u.Host
	} else {
		// The host did have a port (e.g. "the.host.com:1234")
		sqlInstance = host
	}

	if len(u.Path) > 1 {
		// There was a SQL instance name specified in addition to the host
		// E.g. "the.host.com:1234/InstanceName" or "the.host.com/InstanceName"
		sqlInstance = sqlInstance + "\\" + u.Path[1:]
	}

	query := u.Query()
	for key, value := range query {
		if isDatabaseIdentifier(key) {
			databaseName = value[0]
			break
		}
	}

	return
}

//isInstanceIdentifier returns true if the input is equal to "server" (case insensitive)
func isInstanceIdentifier(input string) bool {
	// The GoLang SQL driver only supports "server" in the connection string, not any of the following others:
	// "data source", "address", "addr", "network address"
	return strings.EqualFold("server", input)
}

//isDatabaseIdentifier returns true if the input is equal to "database" (case insensitive)
func isDatabaseIdentifier(input string) bool {
	// The GoLang SQL driver only supports "database" in the connection string, not "initial catalog"
	return strings.EqualFold("database", input)
}
