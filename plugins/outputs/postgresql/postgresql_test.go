package postgresql

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
)

func TestPostgresqlCreateStatement(t *testing.T) {
	p := newPostgresql()
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)

	var m telegraf.Metric
	m, _ = metric.New("m", nil, map[string]interface{}{"f": float64(3.14)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamp,"f" float8,PRIMARY KEY("time"))`, p.generateCreateTable(m))

	m, _ = metric.New("m", nil, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamp,"i" int8,PRIMARY KEY("time"))`, p.generateCreateTable(m))

	m, _ = metric.New("m", map[string]string{"k": "v"}, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE "m"("time" timestamp,"k" text,"i" int8,PRIMARY KEY("time","k"))`, p.generateCreateTable(m))

}

func TestPostgresqlInsertStatement(t *testing.T) {
	p := newPostgresql()

	sql := p.generateInsert("m", []string{"time", "f"})
	assert.Equal(t, `INSERT INTO "m"("time","f") VALUES($1,$2)`, sql)

	sql = p.generateInsert("m", []string{"time", "i"})
	assert.Equal(t, `INSERT INTO "m"("time","i") VALUES($1,$2)`, sql)

	sql = p.generateInsert("m", []string{"time", "f", "i"})
	assert.Equal(t, `INSERT INTO "m"("time","f","i") VALUES($1,$2,$3)`, sql)

	sql = p.generateInsert("m", []string{"time", "k", "i"})
	assert.Equal(t, `INSERT INTO "m"("time","k","i") VALUES($1,$2,$3)`, sql)

	sql = p.generateInsert("m", []string{"time", "k1", "k2", "i"})
	assert.Equal(t, `INSERT INTO "m"("time","k1","k2","i") VALUES($1,$2,$3,$4)`, sql)
}
