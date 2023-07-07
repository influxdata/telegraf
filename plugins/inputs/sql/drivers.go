package sql

import (
	// Blank imports to register the drivers
	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/apache/arrow/go/v13/arrow/flight/flightsql/driver"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/microsoft/go-mssqldb"
)
