package postgresql

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
)

func TestPostgresqlCreateStatement(t *testing.T) {
	p := Postgresql{}
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)

	var m telegraf.Metric
	m, _ = metric.New("m", nil, map[string]interface{}{"f": float64(3.14)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamptz,"f" float8,PRIMARY KEY("time"))`, p.generateCreateTable(m))

	m, _ = metric.New("m", nil, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamptz,"i" int8,PRIMARY KEY("time"))`, p.generateCreateTable(m))

	m, _ = metric.New("m", nil, map[string]interface{}{"f": float64(3.14), "i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamptz,"f" float8,"i" int8,PRIMARY KEY("time"))`, p.generateCreateTable(m))

	m, _ = metric.New("m", map[string]string{"k": "v"}, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamptz,"k" text,"i" int8,PRIMARY KEY("time","k"))`, p.generateCreateTable(m))

	m, _ = metric.New("m", map[string]string{"k1": "v1", "k2": "v2"}, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamptz,"k1" text,"k2" text,"i" int8,PRIMARY KEY("time","k1","k2"))`, p.generateCreateTable(m))
}

func TestPostgresqlInsertStatement(t *testing.T) {
	p := Postgresql{}
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)

	sql, _ := p.generateInsert("m", []string{"time", "f"}, []interface{}{timestamp, 3.1})
	assert.Equal(t, `INSERT INTO "m"("time","f") VALUES($1,$2)`, sql)

	sql, _ = p.generateInsert("m", []string{"time", "i"}, []interface{}{timestamp, 3})
	assert.Equal(t, `INSERT INTO "m"("time","i") VALUES($1,$2)`, sql)

	sql, _ = p.generateInsert("m", []string{"time", "f", "i"}, []interface{}{timestamp, 3.1, 3})
	assert.Equal(t, `INSERT INTO "m"("time","f","i") VALUES($1,$2,$3)`, sql)

	sql, _ = p.generateInsert("m", []string{"time", "k", "i"}, []interface{}{timestamp, "v", 3})
	assert.Equal(t, `INSERT INTO "m"("time","k","i") VALUES($1,$2,$3)`, sql)

	sql, _ = p.generateInsert("m", []string{"time", "k1", "k2", "i"}, []interface{}{timestamp, "v1", "v2", 3})
	assert.Equal(t, `INSERT INTO "m"("time","k1","k2","i") VALUES($1,$2,$3,$4)`, sql)
}
