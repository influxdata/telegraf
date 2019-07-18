package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/jackc/pgx"
)

const (
	insertIntoSQLTemplate = "INSERT INTO %s(%s) VALUES(%s)"
)

// GroupMetricsByMeasurement groups the list of metrics by the measurement name.
// But the values are the index of the measure from the input list of measures.
// [m, m, m2, m2, m] => {m:[0,1,4], m2:[2,3]}
func GroupMetricsByMeasurement(m []telegraf.Metric) map[string][]int {
	toReturn := make(map[string][]int)
	for i, metric := range m {
		var metricLocations []int
		var ok bool
		name := metric.Name()
		if metricLocations, ok = toReturn[name]; !ok {
			metricLocations = []int{}
			toReturn[name] = metricLocations
		}
		toReturn[name] = append(metricLocations, i)
	}
	return toReturn
}

// BuildJsonb returns a byte array of the json representation
// of the passed object.
func BuildJsonb(data interface{}) ([]byte, error) {
	d, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// QuoteIdent returns a sanitized string safe to use in SQL as an identifier
func QuoteIdent(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

// QuoteLiteral returns a sanitized string safe to use in sql as a string literal
func QuoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

// FullTableName returns a sanitized table name with it's schema (if supplied)
func FullTableName(schema, name string) *pgx.Identifier {
	if schema != "" {
		return &pgx.Identifier{schema, name}
	}

	return &pgx.Identifier{name}
}

// Constants for naming PostgreSQL data types both in
// their short and long versions.
const (
	PgBool                     = "boolean"
	PgInt8                     = "int8"
	PgInt4                     = "int4"
	PgInteger                  = "integer"
	PgBigInt                   = "bigint"
	PgFloat8                   = "float8"
	PgDoublePrecision          = "double precision"
	PgText                     = "text"
	PgTimestamptz              = "timestamptz"
	PgTimestampWithTimeZone    = "timestamp with time zone"
	PgTimestamp                = "timestamp"
	PgTimestampWithoutTimeZone = "timestamp without time zone"
	PgSerial                   = "serial"
	PgJSONb                    = "jsonb"
)

// DerivePgDatatype returns the appropriate PostgreSQL data type
// that could hold the value.
func DerivePgDatatype(value interface{}) PgDataType {
	switch value.(type) {
	case bool:
		return PgBool
	case uint64, int64, int, uint:
		return PgInt8
	case uint32, int32:
		return PgInt4
	case float64, float32:
		return PgFloat8
	case string:
		return PgText
	case time.Time:
		return PgTimestamptz
	default:
		log.Printf("E! Unknown datatype %T(%v)", value, value)
		return PgText
	}
}

// LongToShortPgType returns a PostgreSQL datatype in it's short
// notation form.
func LongToShortPgType(longPgType string) PgDataType {
	switch longPgType {
	case PgInteger:
		return PgInt4
	case PgBigInt:
		return PgInt8
	case PgDoublePrecision:
		return PgFloat8
	case PgTimestampWithTimeZone:
		return PgTimestamptz
	case PgTimestampWithoutTimeZone:
		return PgTimestamp
	default:
		return PgDataType(longPgType)
	}
}

// PgTypeCanContain tells you if one PostgreSQL data type can contain
// the values of another without data loss.
func PgTypeCanContain(canThis PgDataType, containThis PgDataType) bool {
	if canThis == containThis {
		return true
	}
	if canThis == PgInt8 {
		return containThis == PgInt4
	}
	if canThis == PgInt4 {
		return containThis == PgSerial
	}
	if canThis == PgFloat8 {
		return containThis == PgInt4
	}
	if canThis == PgTimestamptz {
		return containThis == PgTimestamp
	}

	return false
}

// GenerateInsert returns a SQL statement to insert values in a table
// with $X placeholders for the values
func GenerateInsert(fullSanitizedTableName string, columns []string) string {
	valuePlaceholders := make([]string, len(columns))
	quotedColumns := make([]string, len(columns))
	for i, column := range columns {
		valuePlaceholders[i] = fmt.Sprintf("$%d", i+1)
		quotedColumns[i] = QuoteIdent(column)
	}

	columnNames := strings.Join(quotedColumns, ",")
	values := strings.Join(valuePlaceholders, ",")
	return fmt.Sprintf(insertIntoSQLTemplate, fullSanitizedTableName, columnNames, values)
}
