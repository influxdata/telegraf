package processors

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestRename(t *testing.T) {
	r := Rename{}
	m1 := newMetric("apache", nil, nil)
	m2 := newMetric("phpfpm", nil, nil)
	m3 := newMetric("apache", nil, map[string]interface{}{"scboard_dnslookup": 120})
	m4 := newMetric("etcd", nil, map[string]interface{}{"slow_requests": 10})
	m5 := newMetric("phpfpm", nil, map[string]interface{}{"slow_requests": 10})
	m6 := newMetric("mongodb_col_stats", nil, map[string]interface{}{"ok": 10})
	m7 := newMetric("mongodb", nil, map[string]interface{}{"tcmalloc_heap_size": 10})
	m8 := newMetric("mongodb_db_stats", nil, map[string]interface{}{"avg_obj_size": 10})
	m9 := newMetric("mongodb_shard_stats", nil, map[string]interface{}{"in_use": 10})

	results := r.Process([]telegraf.Metric{m1, m2, m3, m4, m5, m6, m7, m8, m9})
	assert.Equal(t, "apache", results[0].Name())
	assert.Equal(t, "php", results[1].Name())
	assert.Equal(t, "workers.dns", results[2].FieldList()[0].Key)
	assert.Equal(t, "etcd.slow.requests", results[3].FieldList()[0].Key)
	assert.Equal(t, "fpm.requests.slow", results[4].FieldList()[0].Key)
	assert.Equal(t, "mongo", results[5].Name())
	assert.Equal(t, "mongodb_col_stats.ok", results[5].FieldList()[0].Key)
	assert.Equal(t, "tcmalloc_heap_size", results[6].FieldList()[0].Key)
	assert.Equal(t, "database.avg_obj_size", results[7].FieldList()[0].Key)
	assert.Equal(t, "mongodb_shard_stats.in_use", results[8].FieldList()[0].Key)
}

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	return metric.New(name, tags, fields, time.Now())
}
