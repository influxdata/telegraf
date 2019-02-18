package models

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestFilter_ApplyEmpty(t *testing.T) {
	f := Filter{}
	require.NoError(t, f.Compile())
	require.False(t, f.IsActive())

	m, err := metric.New("m",
		map[string]string{},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	require.NoError(t, err)
	require.True(t, f.Select(m))
}

func TestFilter_ApplyTagsDontPass(t *testing.T) {
	filters := []TagFilter{
		{
			Name:   "cpu",
			Filter: []string{"cpu-*"},
		},
	}
	f := Filter{
		TagDrop: filters,
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	require.True(t, f.IsActive())

	m, err := metric.New("m",
		map[string]string{"cpu": "cpu-total"},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	require.NoError(t, err)
	require.False(t, f.Select(m))
}

func TestFilter_ApplyDeleteFields(t *testing.T) {
	f := Filter{
		FieldDrop: []string{"value"},
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	require.True(t, f.IsActive())

	m, err := metric.New("m",
		map[string]string{},
		map[string]interface{}{
			"value":  int64(1),
			"value2": int64(2),
		},
		time.Now())
	require.NoError(t, err)
	require.True(t, f.Select(m))
	f.Modify(m)
	require.Equal(t, map[string]interface{}{"value2": int64(2)}, m.Fields())
}

func TestFilter_ApplyDeleteAllFields(t *testing.T) {
	f := Filter{
		FieldDrop: []string{"value*"},
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	require.True(t, f.IsActive())

	m, err := metric.New("m",
		map[string]string{},
		map[string]interface{}{
			"value":  int64(1),
			"value2": int64(2),
		},
		time.Now())
	require.NoError(t, err)
	require.True(t, f.Select(m))
	f.Modify(m)
	require.Len(t, m.FieldList(), 0)
}

func TestFilter_Empty(t *testing.T) {
	f := Filter{}

	measurements := []string{
		"foo",
		"bar",
		"barfoo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"supercalifradjulisticexpialidocious",
	}

	for _, measurement := range measurements {
		if !f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}
}

func TestFilter_NamePass(t *testing.T) {
	f := Filter{
		NamePass: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	passes := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	drops := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_NameDrop(t *testing.T) {
	f := Filter{
		NameDrop: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	drops := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	passes := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldNamePass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_FieldPass(t *testing.T) {
	f := Filter{
		FieldPass: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	passes := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	drops := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_FieldDrop(t *testing.T) {
	f := Filter{
		FieldDrop: []string{"foo*", "cpu_usage_idle"},
	}
	require.NoError(t, f.Compile())

	drops := []string{
		"foo",
		"foo_bar",
		"foo.bar",
		"foo-bar",
		"cpu_usage_idle",
	}

	passes := []string{
		"bar",
		"barfoo",
		"bar_foo",
		"cpu_usage_busy",
	}

	for _, measurement := range passes {
		if !f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to pass", measurement)
		}
	}

	for _, measurement := range drops {
		if f.shouldFieldPass(measurement) {
			t.Errorf("Expected measurement %s to drop", measurement)
		}
	}
}

func TestFilter_TagPass(t *testing.T) {
	filters := []TagFilter{
		{
			Name:   "cpu",
			Filter: []string{"cpu-*"},
		},
		{
			Name:   "mem",
			Filter: []string{"mem_free"},
		}}
	f := Filter{
		TagPass: filters,
	}
	require.NoError(t, f.Compile())

	passes := [][]*telegraf.Tag{
		{{Key: "cpu", Value: "cpu-total"}},
		{{Key: "cpu", Value: "cpu-0"}},
		{{Key: "cpu", Value: "cpu-1"}},
		{{Key: "cpu", Value: "cpu-2"}},
		{{Key: "mem", Value: "mem_free"}},
	}

	drops := [][]*telegraf.Tag{
		{{Key: "cpu", Value: "cputotal"}},
		{{Key: "cpu", Value: "cpu0"}},
		{{Key: "cpu", Value: "cpu1"}},
		{{Key: "cpu", Value: "cpu2"}},
		{{Key: "mem", Value: "mem_used"}},
	}

	for _, tags := range passes {
		if !f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to pass", tags)
		}
	}

	for _, tags := range drops {
		if f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to drop", tags)
		}
	}
}

func TestFilter_TagDrop(t *testing.T) {
	filters := []TagFilter{
		{
			Name:   "cpu",
			Filter: []string{"cpu-*"},
		},
		{
			Name:   "mem",
			Filter: []string{"mem_free"},
		}}
	f := Filter{
		TagDrop: filters,
	}
	require.NoError(t, f.Compile())

	drops := [][]*telegraf.Tag{
		{{Key: "cpu", Value: "cpu-total"}},
		{{Key: "cpu", Value: "cpu-0"}},
		{{Key: "cpu", Value: "cpu-1"}},
		{{Key: "cpu", Value: "cpu-2"}},
		{{Key: "mem", Value: "mem_free"}},
	}

	passes := [][]*telegraf.Tag{
		{{Key: "cpu", Value: "cputotal"}},
		{{Key: "cpu", Value: "cpu0"}},
		{{Key: "cpu", Value: "cpu1"}},
		{{Key: "cpu", Value: "cpu2"}},
		{{Key: "mem", Value: "mem_used"}},
	}

	for _, tags := range passes {
		if !f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to pass", tags)
		}
	}

	for _, tags := range drops {
		if f.shouldTagsPass(tags) {
			t.Errorf("Expected tags %v to drop", tags)
		}
	}
}

func TestFilter_FilterTagsNoMatches(t *testing.T) {
	m, err := metric.New("m",
		map[string]string{
			"host":  "localhost",
			"mytag": "foobar",
		},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	require.NoError(t, err)
	f := Filter{
		TagExclude: []string{"nomatch"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(m)
	require.Equal(t, map[string]string{
		"host":  "localhost",
		"mytag": "foobar",
	}, m.Tags())

	f = Filter{
		TagInclude: []string{"nomatch"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(m)
	require.Equal(t, map[string]string{}, m.Tags())
}

func TestFilter_FilterTagsMatches(t *testing.T) {
	m, err := metric.New("m",
		map[string]string{
			"host":  "localhost",
			"mytag": "foobar",
		},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	require.NoError(t, err)
	f := Filter{
		TagExclude: []string{"ho*"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(m)
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, m.Tags())

	m, err = metric.New("m",
		map[string]string{
			"host":  "localhost",
			"mytag": "foobar",
		},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	require.NoError(t, err)
	f = Filter{
		TagInclude: []string{"my*"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(m)
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, m.Tags())
}

// TestFilter_FilterNamePassAndDrop used for check case when
// both parameters were defined
// see: https://github.com/influxdata/telegraf/issues/2860
func TestFilter_FilterNamePassAndDrop(t *testing.T) {

	inputData := []string{"name1", "name2", "name3", "name4"}
	expectedResult := []bool{false, true, false, false}

	f := Filter{
		NamePass: []string{"name1", "name2"},
		NameDrop: []string{"name1", "name3"},
	}

	require.NoError(t, f.Compile())

	for i, name := range inputData {
		require.Equal(t, f.shouldNamePass(name), expectedResult[i])
	}
}

// TestFilter_FilterFieldPassAndDrop used for check case when
// both parameters were defined
// see: https://github.com/influxdata/telegraf/issues/2860
func TestFilter_FilterFieldPassAndDrop(t *testing.T) {

	inputData := []string{"field1", "field2", "field3", "field4"}
	expectedResult := []bool{false, true, false, false}

	f := Filter{
		FieldPass: []string{"field1", "field2"},
		FieldDrop: []string{"field1", "field3"},
	}

	require.NoError(t, f.Compile())

	for i, field := range inputData {
		require.Equal(t, f.shouldFieldPass(field), expectedResult[i])
	}
}

// TestFilter_FilterTagsPassAndDrop used for check case when
// both parameters were defined
// see: https://github.com/influxdata/telegraf/issues/2860
func TestFilter_FilterTagsPassAndDrop(t *testing.T) {
	inputData := [][]*telegraf.Tag{
		{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "3"}},
		{{Key: "tag1", Value: "1"}, {Key: "tag2", Value: "2"}},
		{{Key: "tag1", Value: "2"}, {Key: "tag2", Value: "1"}},
		{{Key: "tag1", Value: "4"}, {Key: "tag2", Value: "1"}},
	}

	expectedResult := []bool{false, true, false, false}

	filterPass := []TagFilter{
		{
			Name:   "tag1",
			Filter: []string{"1", "4"},
		},
	}

	filterDrop := []TagFilter{
		{
			Name:   "tag1",
			Filter: []string{"4"},
		},
		{
			Name:   "tag2",
			Filter: []string{"3"},
		},
	}

	f := Filter{
		TagDrop: filterDrop,
		TagPass: filterPass,
	}

	require.NoError(t, f.Compile())

	for i, tag := range inputData {
		require.Equal(t, f.shouldTagsPass(tag), expectedResult[i])
	}

}

func BenchmarkFilter(b *testing.B) {
	tests := []struct {
		name   string
		filter Filter
		metric telegraf.Metric
	}{
		{
			name:   "empty filter",
			filter: Filter{},
			metric: testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "namepass",
			filter: Filter{
				NamePass: []string{"cpu"},
			},
			metric: testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			require.NoError(b, tt.filter.Compile())
			for n := 0; n < b.N; n++ {
				tt.filter.Select(tt.metric)
			}
		})
	}
}
