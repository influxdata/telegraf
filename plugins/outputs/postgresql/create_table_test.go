package postgresql

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCreateTable(t *testing.T) {
	p := newPostgresql()
	p.TagsAsJsonb = true
	p.FieldsAsJsonb = true
	timestamp := time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC)

	var m telegraf.Metric
	m, _ = metric.New("m", nil, map[string]interface{}{"f": float64(3.14)}, timestamp)
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,fields jsonb)`, p.generateCreateTable(m))

	m, _ = metric.New("m", map[string]string{"k": "v"}, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,tags jsonb,fields jsonb)`, p.generateCreateTable(m))

	p.TagsAsJsonb = false
	p.FieldsAsJsonb = false

	m, _ = metric.New("m", nil, map[string]interface{}{"f": float64(3.14)}, timestamp)
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,"f" float8)`, p.generateCreateTable(m))

	m, _ = metric.New("m", nil, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,"i" int8)`, p.generateCreateTable(m))

	m, _ = metric.New("m", map[string]string{"k": "v"}, map[string]interface{}{"i": int(3)}, timestamp)
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,"k" text,"i" int8)`, p.generateCreateTable(m))

	p.TagsAsForeignkeys = true
	assert.Equal(t,
		`CREATE TABLE IF NOT EXISTS "public"."m_tag"(tag_id serial primary key,"k" text,UNIQUE("k"));`+
			`CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,tag_id int,"i" int8)`,
		p.generateCreateTable(m))

	p.TagsAsJsonb = true
	assert.Equal(t,
		`CREATE TABLE IF NOT EXISTS "public"."m_tag"(tag_id serial primary key,tags jsonb,UNIQUE(tags));`+
			`CREATE TABLE IF NOT EXISTS "public"."m"(time timestamptz,tag_id int,"i" int8)`,
		p.generateCreateTable(m))
}
