package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestFilter_ApplyEmpty(t *testing.T) {
	f := Filter{}
	require.NoError(t, f.Compile())
	require.False(t, f.IsActive())

	m := metric.New("m",
		map[string]string{},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	selected, err := f.Select(m)
	require.NoError(t, err)
	require.True(t, selected)
}

func TestFilter_ApplyTagsDontPass(t *testing.T) {
	filters := []TagFilter{
		{
			Name:   "cpu",
			Values: []string{"cpu-*"},
		},
	}
	f := Filter{
		TagDropFilters: filters,
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	require.True(t, f.IsActive())

	m := metric.New("m",
		map[string]string{"cpu": "cpu-total"},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	selected, err := f.Select(m)
	require.NoError(t, err)
	require.False(t, selected)
}

func TestFilter_ApplyDeleteFields(t *testing.T) {
	f := Filter{
		FieldExclude: []string{"value"},
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	require.True(t, f.IsActive())

	m := metric.New("m",
		map[string]string{},
		map[string]interface{}{
			"value":  int64(1),
			"value2": int64(2),
		},
		time.Now())
	selected, err := f.Select(m)
	require.NoError(t, err)
	require.True(t, selected)
	f.Modify(m)
	require.Equal(t, map[string]interface{}{"value2": int64(2)}, m.Fields())
}

func TestFilter_ApplyDeleteAllFields(t *testing.T) {
	f := Filter{
		FieldExclude: []string{"value*"},
	}
	require.NoError(t, f.Compile())
	require.NoError(t, f.Compile())
	require.True(t, f.IsActive())

	m := metric.New("m",
		map[string]string{},
		map[string]interface{}{
			"value":  int64(1),
			"value2": int64(2),
		},
		time.Now())
	selected, err := f.Select(m)
	require.NoError(t, err)
	require.True(t, selected)
	f.Modify(m)
	require.Empty(t, m.FieldList())
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
		"supercalifragilisticexpialidocious",
	}

	for _, measurement := range measurements {
		if !f.shouldNamePass(measurement) {
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

func TestFilter_NamePass_WithSeparator(t *testing.T) {
	f := Filter{
		NamePass:           []string{"foo.*.bar", "foo.*.abc.*.bar"},
		NamePassSeparators: ".,",
	}
	require.NoError(t, f.Compile())

	passes := []string{
		"foo..bar",
		"foo.abc.bar",
		"foo..abc..bar",
		"foo.xyz.abc.xyz-xyz.bar",
	}

	drops := []string{
		"foo.bar",
		"foo.abc,.bar", // "abc," is not considered under * as ',' is specified as a separator
		"foo..abc.bar", // ".abc" shall not be matched under * as '.' is specified as a separator
		"foo.abc.abc.bar",
		"foo.xyz.abc.xyz.xyz.bar",
		"foo.xyz.abc.xyz,xyz.bar",
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

func TestFilter_NameDrop_WithSeparator(t *testing.T) {
	f := Filter{
		NameDrop:           []string{"foo.*.bar", "foo.*.abc.*.bar"},
		NameDropSeparators: ".,",
	}
	require.NoError(t, f.Compile())

	drops := []string{
		"foo..bar",
		"foo.abc.bar",
		"foo..abc..bar",
		"foo.xyz.abc.xyz-xyz.bar",
	}

	passes := []string{
		"foo.bar",
		"foo.abc,.bar", // "abc," is not considered under * as ',' is specified as a separator
		"foo..abc.bar", // ".abc" shall not be matched under * as '.' is specified as a separator
		"foo.abc.abc.bar",
		"foo.xyz.abc.xyz.xyz.bar",
		"foo.xyz.abc.xyz,xyz.bar",
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

func TestFilter_FieldInclude(t *testing.T) {
	f := Filter{
		FieldInclude: []string{"foo*", "cpu_usage_idle"},
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

	for _, field := range passes {
		require.Truef(t, ShouldPassFilters(f.fieldIncludeFilter, f.fieldExcludeFilter, field), "Expected field %s to pass", field)
	}

	for _, field := range drops {
		require.Falsef(t, ShouldPassFilters(f.fieldIncludeFilter, f.fieldExcludeFilter, field), "Expected field %s to drop", field)
	}
}

func TestFilter_FieldExclude(t *testing.T) {
	f := Filter{
		FieldExclude: []string{"foo*", "cpu_usage_idle"},
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

	for _, field := range passes {
		require.Truef(t, ShouldPassFilters(f.fieldIncludeFilter, f.fieldExcludeFilter, field), "Expected field %s to pass", field)
	}

	for _, field := range drops {
		require.Falsef(t, ShouldPassFilters(f.fieldIncludeFilter, f.fieldExcludeFilter, field), "Expected field %s to drop", field)
	}
}

func TestFilter_TagPass(t *testing.T) {
	filters := []TagFilter{
		{
			Name:   "cpu",
			Values: []string{"cpu-*"},
		},
		{
			Name:   "mem",
			Values: []string{"mem_free"},
		}}
	f := Filter{
		TagPassFilters: filters,
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
			Values: []string{"cpu-*"},
		},
		{
			Name:   "mem",
			Values: []string{"mem_free"},
		}}
	f := Filter{
		TagDropFilters: filters,
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
	m := metric.New("m",
		map[string]string{
			"host":  "localhost",
			"mytag": "foobar",
		},
		map[string]interface{}{"value": int64(1)},
		time.Now())
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
	m := metric.New("m",
		map[string]string{
			"host":  "localhost",
			"mytag": "foobar",
		},
		map[string]interface{}{"value": int64(1)},
		time.Now())
	f := Filter{
		TagExclude: []string{"ho*"},
	}
	require.NoError(t, f.Compile())

	f.filterTags(m)
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, m.Tags())

	m = metric.New("m",
		map[string]string{
			"host":  "localhost",
			"mytag": "foobar",
		},
		map[string]interface{}{"value": int64(1)},
		time.Now())
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

// TestFilter_FieldIncludeAndExclude used for check case when
// both parameters were defined
// see: https://github.com/influxdata/telegraf/issues/2860
func TestFilter_FieldIncludeAndExclude(t *testing.T) {
	inputData := []string{"field1", "field2", "field3", "field4"}
	expectedResult := []bool{false, true, false, false}

	f := Filter{
		FieldInclude: []string{"field1", "field2"},
		FieldExclude: []string{"field1", "field3"},
	}

	require.NoError(t, f.Compile())

	for i, field := range inputData {
		require.Equal(t, ShouldPassFilters(f.fieldIncludeFilter, f.fieldExcludeFilter, field), expectedResult[i])
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
			Values: []string{"1", "4"},
		},
	}

	filterDrop := []TagFilter{
		{
			Name:   "tag1",
			Values: []string{"4"},
		},
		{
			Name:   "tag2",
			Values: []string{"3"},
		},
	}

	f := Filter{
		TagDropFilters: filterDrop,
		TagPassFilters: filterPass,
	}
	require.NoError(t, f.Compile())

	for i, tag := range inputData {
		require.Equal(t, f.shouldTagsPass(tag), expectedResult[i])
	}
}

func TestFilter_MetricPass(t *testing.T) {
	m := testutil.MustMetric("cpu",
		map[string]string{
			"host":   "Hugin",
			"source": "myserver@mycompany.com",
			"status": "ok",
		},
		map[string]interface{}{
			"value":  15.0,
			"id":     "24cxnwr3480k",
			"on":     true,
			"count":  18,
			"errors": 29,
			"total":  129,
		},
		time.Date(2023, time.April, 24, 23, 30, 15, 42, time.UTC),
	)

	var tests = []struct {
		name       string
		expression string
		expected   bool
	}{
		{
			name:     "empty",
			expected: true,
		},
		{
			name:       "exact name match (pass)",
			expression: `name == "cpu"`,
			expected:   true,
		},
		{
			name:       "exact name match (fail)",
			expression: `name == "test"`,
			expected:   false,
		},
		{
			name:       "case-insensitive tag match",
			expression: `tags.host.lowerAscii() == "hugin"`,
			expected:   true,
		},
		{
			name:       "regexp tag match",
			expression: `tags.source.matches("^[0-9a-zA-z-_]+@mycompany.com$")`,
			expected:   true,
		},
		{
			name:       "match field value",
			expression: `fields.count > 10`,
			expected:   true,
		},
		{
			name:       "match timestamp year",
			expression: `time.getFullYear() == 2023`,
			expected:   true,
		},
		{
			name:       "now",
			expression: `now() > time`,
			expected:   true,
		},
		{
			name:       "arithmetic",
			expression: `fields.count + fields.errors < fields.total`,
			expected:   true,
		},
		{
			name:       "logical expression",
			expression: `(name.startsWith("t") || fields.on) && "id" in fields && fields.id.contains("nwr")`,
			expected:   true,
		},
		{
			name:       "time arithmetic",
			expression: `time >= timestamp("2023-04-25T00:00:00Z") - duration("24h")`,
			expected:   true,
		},
		{
			name:       "complex field filtering",
			expression: `fields.exists(f, type(fields[f]) in [int, uint, double] && fields[f] > 20.0)`,
			expected:   true,
		},
		{
			name:       "complex field filtering (exactly one)",
			expression: `fields.exists_one(f, type(fields[f]) in [int, uint, double] && fields[f] > 20.0)`,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Filter{
				MetricPass: tt.expression,
			}
			require.NoError(t, f.Compile())
			selected, err := f.Select(m)
			require.NoError(t, err)
			require.Equal(t, tt.expected, selected)
		})
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
		{
			name: "metric filter exact name",
			filter: Filter{
				MetricPass: `name == "cpu"`,
			},
			metric: testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "metric filter regexp",
			filter: Filter{
				MetricPass: `name.matches("^c[a-z]*$")`,
			},
			metric: testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "metric filter time",
			filter: Filter{
				MetricPass: `time >= timestamp("2023-04-25T00:00:00Z") - duration("24h")`,
			},
			metric: testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "metric filter complex",
			filter: Filter{
				MetricPass: `"source" in tags` +
					` && fields.exists(f, type(fields[f]) in [int, uint, double] && fields[f] > 20.0)` +
					` && time >= timestamp("2023-04-25T00:00:00Z") - duration("24h")`,
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
				_, err := tt.filter.Select(tt.metric)
				require.NoError(b, err)
			}
		})
	}
}
