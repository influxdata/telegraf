package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/influxdata/telegraf"
)

const (
	insertIntoSQLTemplate = "INSERT INTO %s(%s) VALUES(%s)"
)

func TagListToJSON(tagList []*telegraf.Tag) []byte {
	tags := make(map[string]string, len(tagList))
	for _, tag := range tagList {
		tags[tag.Key] = tag.Value
	}
	bs, _ := json.Marshal(tags)
	return bs
}

func FieldListToJSON(fieldList []*telegraf.Field) ([]byte, error) {
	fields := make(map[string]interface{}, len(fieldList))
	for _, field := range fieldList {
		fields[field.Key] = field.Value
	}
	return json.Marshal(fields)
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
func FullTableName(schema, name string) pgx.Identifier {
	if schema != "" {
		return pgx.Identifier{schema, name}
	}

	return pgx.Identifier{name}
}

// Constants for naming PostgreSQL data types both in
// their short and long versions.
const (
	PgBool                     = "boolean"
	PgSmallInt                 = "smallint"
	PgInteger                  = "integer"
	PgBigInt                   = "bigint"
	PgReal                     = "real"
	PgDoublePrecision          = "double precision"
	PgNumeric                  = "numeric"
	PgText                     = "text"
	PgTimestampWithTimeZone    = "timestamp with time zone"
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
	case uint64:
		return PgNumeric
	case int64, int, uint, uint32:
		return PgBigInt
	case int32:
		return PgInteger
	case int16, int8:
		return PgSmallInt
	case float64:
		return PgDoublePrecision
	case float32:
		return PgReal
	case string:
		return PgText
	case time.Time:
		return PgTimestampWithTimeZone
	default:
		log.Printf("E! Unknown datatype %T(%v)", value, value)
		return PgText
	}
}

// PgTypeCanContain tells you if one PostgreSQL data type can contain
// the values of another without data loss.
func PgTypeCanContain(canThis PgDataType, containThis PgDataType) bool {
	switch canThis {
	case containThis:
		return true
	case PgBigInt:
		return containThis == PgInteger || containThis == PgSmallInt
	case PgInteger:
		return containThis == PgSmallInt
	case PgDoublePrecision, PgReal: // You can store a real in a double, you just lose precision
		return containThis == PgReal || containThis == PgBigInt || containThis == PgInteger || containThis == PgSmallInt
	case PgNumeric:
		return containThis == PgBigInt || containThis == PgSmallInt || containThis == PgInteger || containThis == PgReal || containThis == PgDoublePrecision
	case PgTimestampWithTimeZone:
		return containThis == PgTimestampWithoutTimeZone
	default:
		return false
	}
}

// pgxLogger makes telegraf.Logger compatible with pgx.Logger
type PGXLogger struct {
	telegraf.Logger
}

func (l PGXLogger) Log(_ context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	switch level {
	case pgx.LogLevelError:
		l.Errorf("PG %s - %+v", msg, data)
	case pgx.LogLevelWarn:
		l.Warnf("PG %s - %+v", msg, data)
	case pgx.LogLevelInfo, pgx.LogLevelNone:
		l.Infof("PG %s - %+v", msg, data)
	case pgx.LogLevelDebug, pgx.LogLevelTrace:
		l.Debugf("PG %s - %+v", msg, data)
	default:
		l.Debugf("PG %s - %+v", msg, data)
	}
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

func GetTagID(metric telegraf.Metric) int64 {
	hash := fnv.New64a()
	for _, tag := range metric.TagList() {
		_, _ = hash.Write([]byte(tag.Key))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(tag.Value))
		_, _ = hash.Write([]byte{0})
	}
	// Convert to int64 as postgres does not support uint64
	return int64(hash.Sum64())
}

// WaitGroup is similar to sync.WaitGroup, but allows interruptable waiting (e.g. a timeout).
type WaitGroup struct {
	count int32
	done chan struct{}
}
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		done: make(chan struct{}),
	}
}

func (wg *WaitGroup) Add(i int32) {
	select {
	case <-wg.done:
		panic("use of an already-done WaitGroup")
	default:
	}
	atomic.AddInt32(&wg.count, i)
}

func (wg *WaitGroup) Done() {
	i := atomic.AddInt32(&wg.count, -1)
	if i == 0 {
		close(wg.done)
	}
	if i < 0 {
		panic("too many Done() calls")
	}
}

func (wg *WaitGroup) C() <-chan struct{} {
	return wg.done
}
