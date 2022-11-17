package postgresql

import (
	"time"
)

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

// Types from pguint
const (
	PgUint8 = "uint8"
)

// DerivePgDatatype returns the appropriate PostgreSQL data type
// that could hold the value.
func (p *Postgresql) derivePgDatatype(value interface{}) string {
	if p.Uint64Type == PgUint8 {
		if _, ok := value.(uint64); ok {
			return PgUint8
		}
	}

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
		return PgTimestampWithoutTimeZone
	default:
		return PgText
	}
}
