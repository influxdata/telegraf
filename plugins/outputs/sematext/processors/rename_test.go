package processors

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {
	r := Rename{}
	m1 := newMetric("apache", nil, nil)
	m2 := newMetric("phpfpm", nil, nil)
	m3 := newMetric("apache", nil, map[string]interface{}{"scboard_dnslookup": 120})
	m4 := newMetric("etcd", nil, map[string]interface{}{"slow_requests": 10})
	m5 := newMetric("phpfpm", nil, map[string]interface{}{"slow_requests": 10})
	m6 := newMetric("mongodb_col_stats", nil, map[string]interface{}{"ok": 10})

	results, err := r.Process([]telegraf.Metric{m1, m2, m3, m4, m5, m6})
	require.NoError(t, err)
	assert.Equal(t, "apache", results[0].Name())
	assert.Equal(t, "php", results[1].Name())
	assert.Equal(t, "workers.dns", results[2].FieldList()[0].Key)
	assert.Equal(t, "slow_requests", results[3].FieldList()[0].Key)
	assert.Equal(t, "fpm.requests.slow", results[4].FieldList()[0].Key)
	assert.Equal(t, "mongo", results[5].Name())
	assert.Equal(t, "mongodb_col_stats.ok", results[5].FieldList()[0].Key)
}

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m := metric.New(name, tags, fields, time.Now())
	return m
}
