package postgresql_copy

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
)

func TestBuildColumns(t *testing.T) {
	table := "cpu_usage"
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	tags := map[string]string{"host": "address", "zone": "west"}
	fields := map[string]interface{}{"cpu_perc": float64(0.2)}
	m, _ := metric.New(table, tags, fields, timestamp)

	p := newPostgresqlCopy()
	p.Columns = make(map[string][]string)
	assert.Empty(t, p.Columns[table])

	p.buildColumns(table, m)
	assert.Equal(t, len(p.Columns[table]), 3)
	assert.Equal(t, p.Columns[table][0], "cpu_perc")
	assert.Contains(t, p.Columns[table], "host")
	assert.Contains(t, p.Columns[table], "zone")
}

func TestBuildValues(t *testing.T) {
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)
	table := "cpu_usage"
	tags := map[string]string{"host": "address", "zone": "west"}
	fields := map[string]interface{}{"cpu_perc": float64(0.2)}
	m, _ := metric.New(table, tags, fields, timestamp)

	p := newPostgresqlCopy()
	p.Columns = make(map[string][]string)
	p.buildColumns(table, m)

	values := buildValues(m, p.Columns[table])
	assert.Equal(t, len(values), 4)
	assert.Equal(t, values[0], 0.2)
	assert.Contains(t, values, "address")
	assert.Contains(t, values, "west")
	assert.Equal(t, values[3], m.Time())
}
