package utils

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"strings"
	"sync/atomic"

	"github.com/jackc/pgx/v4"

	"github.com/influxdata/telegraf"
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

// QuoteIdentifier returns a sanitized string safe to use in SQL as an identifier
func QuoteIdentifier(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

// QuoteLiteral returns a sanitized string safe to use in sql as a string literal
func QuoteLiteral(name string) string {
	return "'" + strings.Replace(name, "'", "''", -1) + "'"
}

// FullTableName returns a sanitized table name with its schema (if supplied)
func FullTableName(schema, name string) pgx.Identifier {
	if schema != "" {
		return pgx.Identifier{schema, name}
	}

	return pgx.Identifier{name}
}

// PGXLogger makes telegraf.Logger compatible with pgx.Logger
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

func GetTagID(metric telegraf.Metric) int64 {
	hash := fnv.New64a()
	for _, tag := range metric.TagList() {
		hash.Write([]byte(tag.Key))   //nolint:revive // all Write() methods for hash in fnv.go returns nil err
		hash.Write([]byte{0})         //nolint:revive // all Write() methods for hash in fnv.go returns nil err
		hash.Write([]byte(tag.Value)) //nolint:revive // all Write() methods for hash in fnv.go returns nil err
		hash.Write([]byte{0})         //nolint:revive // all Write() methods for hash in fnv.go returns nil err
	}
	// Convert to int64 as postgres does not support uint64
	return int64(hash.Sum64())
}

// WaitGroup is similar to sync.WaitGroup, but allows interruptable waiting (e.g. a timeout).
type WaitGroup struct {
	count int32
	done  chan struct{}
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
