package utils

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestPostgresqlQuote(t *testing.T) {
	assert.Equal(t, `"foo"`, QuoteIdent("foo"))
	assert.Equal(t, `"fo'o"`, QuoteIdent("fo'o"))
	assert.Equal(t, `"fo""o"`, QuoteIdent("fo\"o"))

	assert.Equal(t, "'foo'", QuoteLiteral("foo"))
	assert.Equal(t, "'fo''o'", QuoteLiteral("fo'o"))
	assert.Equal(t, "'fo\"o'", QuoteLiteral("fo\"o"))
}

func TestBuildJsonb(t *testing.T) {
	testCases := []struct {
		desc string
		in   interface{}
		out  string
	}{
		{
			desc: "simple map",
			in:   map[string]int{"a": 1},
			out:  `{"a":1}`,
		}, {
			desc: "single number",
			in:   1,
			out:  `1`,
		}, {
			desc: "interface map",
			in:   map[int]interface{}{1: "a"},
			out:  `{"1":"a"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			res, err := BuildJsonb(tc.in)
			assert.Nil(t, err)
			assert.Equal(t, tc.out, string(res))

		})
	}
}

func TestFullTableName(t *testing.T) {
	assert.Equal(t, `"tableName"`, FullTableName("", "tableName").Sanitize())
	assert.Equal(t, `"table name"`, FullTableName("", "table name").Sanitize())
	assert.Equal(t, `"table.name"`, FullTableName("", "table.name").Sanitize())
	assert.Equal(t, `"table"."name"`, FullTableName("table", "name").Sanitize())
	assert.Equal(t, `"schema name"."table name"`, FullTableName("schema name", "table name").Sanitize())
}

func TestDerivePgDataType(t *testing.T) {
	assert.Equal(t, PgDataType("boolean"), DerivePgDatatype(true))
	assert.Equal(t, PgDataType("int8"), DerivePgDatatype(uint64(1)))
	assert.Equal(t, PgDataType("int8"), DerivePgDatatype(1))
	assert.Equal(t, PgDataType("int8"), DerivePgDatatype(uint(1)))
	assert.Equal(t, PgDataType("int8"), DerivePgDatatype(int64(1)))
	assert.Equal(t, PgDataType("int4"), DerivePgDatatype(uint32(1)))
	assert.Equal(t, PgDataType("int4"), DerivePgDatatype(int32(1)))
	assert.Equal(t, PgDataType("float8"), DerivePgDatatype(float64(1.0)))
	assert.Equal(t, PgDataType("float8"), DerivePgDatatype(float32(1.0)))
	assert.Equal(t, PgDataType("text"), DerivePgDatatype(""))
	assert.Equal(t, PgDataType("timestamptz"), DerivePgDatatype(time.Now()))
	assert.Equal(t, PgDataType("text"), DerivePgDatatype([]int{}))
}

func TestLongToShortPgType(t *testing.T) {
	assert.Equal(t, PgDataType("boolean"), LongToShortPgType("boolean"))
	assert.Equal(t, PgDataType("int4"), LongToShortPgType("integer"))
	assert.Equal(t, PgDataType("int8"), LongToShortPgType("bigint"))
	assert.Equal(t, PgDataType("float8"), LongToShortPgType("double precision"))
	assert.Equal(t, PgDataType("timestamptz"), LongToShortPgType("timestamp with time zone"))
	assert.Equal(t, PgDataType("timestamp"), LongToShortPgType("timestamp without time zone"))
	assert.Equal(t, PgDataType("jsonb"), LongToShortPgType("jsonb"))
	assert.Equal(t, PgDataType("text"), LongToShortPgType("text"))
	assert.Equal(t, PgDataType("unknown"), LongToShortPgType("unknown"))
}

func TestPgTypeCanContain(t *testing.T) {
	assert.True(t, PgTypeCanContain(PgDataType("bogus same"), PgDataType("bogus same")))
	assert.True(t, PgTypeCanContain(PgDataType("int8"), PgDataType("int4")))
	assert.False(t, PgTypeCanContain(PgDataType("int8"), PgDataType("float8")))
	assert.False(t, PgTypeCanContain(PgDataType("int8"), PgDataType("timestamptz")))

	assert.True(t, PgTypeCanContain(PgDataType("int4"), PgDataType("serial")))
	assert.True(t, PgTypeCanContain(PgDataType("int8"), PgDataType("int4")))
	assert.False(t, PgTypeCanContain(PgDataType("int4"), PgDataType("int8")))

	assert.False(t, PgTypeCanContain(PgDataType("float8"), PgDataType("int8")))
	assert.True(t, PgTypeCanContain(PgDataType("float8"), PgDataType("int4")))

	assert.True(t, PgTypeCanContain(PgDataType("timestamptz"), PgDataType("timestamp")))

	assert.False(t, PgTypeCanContain(PgDataType("text"), PgDataType("timestamp")))
}

func TestGroupMetricsByMeasurement(t *testing.T) {
	m11, _ := metric.New("m", map[string]string{}, map[string]interface{}{}, time.Now())
	m12, _ := metric.New("m", map[string]string{"t1": "tv1"}, map[string]interface{}{"f1": 1}, time.Now())
	m13, _ := metric.New("m", map[string]string{}, map[string]interface{}{"f2": 2}, time.Now())

	m21, _ := metric.New("m2", map[string]string{}, map[string]interface{}{}, time.Now())
	m22, _ := metric.New("m2", map[string]string{"t1": "tv1"}, map[string]interface{}{"f1": 1}, time.Now())
	m23, _ := metric.New("m2", map[string]string{}, map[string]interface{}{"f2": 2}, time.Now())
	in := []telegraf.Metric{m11, m12, m21, m22, m13, m23}
	expected := map[string][]int{
		"m":  {0, 1, 4},
		"m2": {2, 3, 5},
	}
	got := GroupMetricsByMeasurement(in)
	assert.Equal(t, expected, got)
}

func TestGenerateInsert(t *testing.T) {

	sql := GenerateInsert(`"m"`, []string{"time", "f"})
	assert.Equal(t, `INSERT INTO "m"("time","f") VALUES($1,$2)`, sql)

	sql = GenerateInsert(`"m"`, []string{"time", "i"})
	assert.Equal(t, `INSERT INTO "m"("time","i") VALUES($1,$2)`, sql)

	sql = GenerateInsert(`"public"."m"`, []string{"time", "f", "i"})
	assert.Equal(t, `INSERT INTO "public"."m"("time","f","i") VALUES($1,$2,$3)`, sql)

	sql = GenerateInsert(`"public"."m n"`, []string{"time", "k", "i"})
	assert.Equal(t, `INSERT INTO "public"."m n"("time","k","i") VALUES($1,$2,$3)`, sql)

	sql = GenerateInsert("m", []string{"time", "k1", "k2", "i"})
	assert.Equal(t, `INSERT INTO m("time","k1","k2","i") VALUES($1,$2,$3,$4)`, sql)
}
