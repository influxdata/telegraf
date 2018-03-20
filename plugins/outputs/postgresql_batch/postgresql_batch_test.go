package postgresql_batch

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
)

func TestBuildInsert(t *testing.T) {
	table := "cpu_usage"
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	tags := map[string]string{"host": "address", "zone": "west"}
	fields := map[string]interface{}{"cpu_perc": float64(0.2)}
	m, _ := metric.New(table, tags, fields, timestamp)

	p := newPostgresqlBatch()
	p.Inserts = make(map[string]string)
	assert.Empty(t, p.Inserts[table])
	p.Columns = make(map[string][]string)
	assert.Empty(t, p.Columns[table])

	p.buildTableInsert(m)
	assert.Equal(t, len(p.Columns[table]), 3)
	assert.Contains(t, p.Columns[table], "host")
	assert.Contains(t, p.Columns[table], "zone")
	assert.Contains(t, p.Columns[table], "cpu_perc")
	assert.Contains(t, p.Inserts[table], "INSERT INTO \"" + table + "\"(")
	assert.Contains(t, p.Inserts[table], "\"host\"")
	assert.Contains(t, p.Inserts[table], "\"zone\"")
	assert.Contains(t, p.Inserts[table], "\"cpu_perc\"")
	assert.Contains(t, p.Inserts[table], ",\"time\") VALUES ")
}

func TestBuildValues(t *testing.T) {
	table := "cpu_usage"
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	tags := map[string]string{"host": "address", "zone": "west"}
	fields := map[string]interface{}{"cpu_perc": float64(0.2)}
	m, _ := metric.New(table, tags, fields, timestamp)

	p := newPostgresqlBatch()
	p.Inserts = make(map[string]string)
	p.Columns = make(map[string][]string)

	p.buildTableInsert(m)
	values := buildValues(m, p.Columns[table])
	assert.Contains(t, values, "'address'")
	assert.Contains(t, values, "'west'")
	assert.Contains(t, values, "'0.2'")
	assert.Contains(t, values, "'2010-11-10")
}