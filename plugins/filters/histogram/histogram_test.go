package histogram

import (
	"github.com/gobwas/glob"
	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHistFirstPass(t *testing.T) {
	hist := &Histogram{
		fieldMap:      make(map[metricID]map[string]*Aggregate),
		metricTags:    make(map[metricID]map[string]string),
		rollupMap:     make(map[metricID]*rollup),
		matchGlobs:    make(map[string]glob.Glob),
		FlushInterval: "1s",
		Bucketsize:    20,
		Rollup: []string{
			"(Name new) (Tag interface en*) (Functions mean 0.90)",
		},
	}
	shutdown := make(chan struct{})
	in := make(chan telegraf.Metric)
	out := hist.Pipe(in)
	go func() {
		hist.Start(shutdown)
	}()
	metric, _ := telegraf.NewMetric("metric_1", map[string]string{
		"name": "ali",
	}, map[string]interface{}{
		"ed": 1.1,
	}, time.Now().UTC())
	in <- metric
	item := <-out
	shutdown <- struct{}{}
	mTags := item.Tags()
	fields := metric.Fields()
	assert.Equal(t, item.Name(), "metric_1", "Metric name match")
	ed, ok := fields["ed"].(float64)
	assert.True(t, ok, "Field value is not flaot")
	assert.Equal(t, ed, 1.1, "Field match")
	assert.Equal(t, mTags["name"], mTags["name"], "Field match")
}

func TestHistFirstAggregate(t *testing.T) {
	hist := &Histogram{
		fieldMap:      make(map[metricID]map[string]*Aggregate),
		metricTags:    make(map[metricID]map[string]string),
		rollupMap:     make(map[metricID]*rollup),
		matchGlobs:    make(map[string]glob.Glob),
		FlushInterval: "1s",
		Bucketsize:    20,
		Rollup: []string{
			"(Name rollup_1) (Measurements metric_1) (Functions mean 0.90)",
		},
	}
	shutdown := make(chan struct{})
	in := make(chan telegraf.Metric)
	out := hist.Pipe(in)
	go func() {
		hist.Start(shutdown)
	}()
	metric, _ := telegraf.NewMetric("metric_1", map[string]string{
		"name": "ali",
	}, map[string]interface{}{
		"ed": 1.1,
	}, time.Now().UTC())
	in <- metric
	item := <-out
	shutdown <- struct{}{}
	mTags := item.Tags()
	fields := item.Fields()
	assert.Equal(t, item.Name(), "rollup_1", "Metric name match")
	if assert.NotNil(t, fields["ed.mean"], "Mean is present") {
		ed, ok := fields["ed.mean"].(float64)
		assert.True(t, ok, "Field value is not flaot")
		assert.Equal(t, ed, 1.1, "Field match")
		assert.Equal(t, mTags["name"], mTags["name"], "Field match")
	}
	if assert.NotNil(t, fields["ed.p0.90"], "0.90 perc not present") {
		ed, ok := fields["ed.p0.90"].(float64)
		assert.True(t, ok, "Field value is not flaot")
		assert.Equal(t, ed, 1.1, "Field match")
		assert.Equal(t, mTags["name"], mTags["name"], "Field match")
	}
}

func TestHistPassOldAggregate(t *testing.T) {
	hist := &Histogram{
		fieldMap:      make(map[metricID]map[string]*Aggregate),
		metricTags:    make(map[metricID]map[string]string),
		rollupMap:     make(map[metricID]*rollup),
		matchGlobs:    make(map[string]glob.Glob),
		FlushInterval: "1s",
		Bucketsize:    20,
		Rollup: []string{
			"(Name rollup_1) (Measurements metric_1) (Functions mean 0.90) (Pass)",
		},
	}
	shutdown := make(chan struct{})
	in := make(chan telegraf.Metric)
	out := hist.Pipe(in)
	go func() {
		hist.Start(shutdown)
	}()
	metric, _ := telegraf.NewMetric("metric_1", map[string]string{
		"name": "ali",
	}, map[string]interface{}{
		"ed": 1.1,
	}, time.Now().UTC())
	in <- metric
	originalMetric := <-out
	assert.Equal(t, originalMetric.Name(), "metric_1", "Original matric should present (Pass flag exists)")
	item := <-out
	shutdown <- struct{}{}
	mTags := item.Tags()
	fields := item.Fields()
	assert.Equal(t, item.Name(), "rollup_1", "Metric name match")
	if assert.NotNil(t, fields["ed.mean"], "Mean is present") {
		ed, ok := fields["ed.mean"].(float64)
		assert.True(t, ok, "Field value is not flaot")
		assert.Equal(t, ed, 1.1, "Field match")
		assert.Equal(t, mTags["name"], mTags["name"], "Field match")
	}
	if assert.NotNil(t, fields["ed.p0.90"], "0.90 perc not present") {
		ed, ok := fields["ed.p0.90"].(float64)
		assert.True(t, ok, "Field value is not flaot")
		assert.Equal(t, ed, 1.1, "Field match")
		assert.Equal(t, mTags["name"], mTags["name"], "Field match")
	}
}
