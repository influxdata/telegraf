package starlark

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestStarlark(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		input    telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "noop",
			source: `
def apply(metric):
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "iterate tags",
			source: `
def apply(metric):
	for k, v in metric.tags.items():
		metric.fields[k] = v
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
					"cpu":  "cpu0",
				},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"host": "example.org",
						"cpu":  "cpu0",
					},
					map[string]interface{}{
						"time_idle": 42,
						"host":      "example.org",
						"cpu":       "cpu0",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set name",
			source: `
def apply(metric):
	metric.name = "cpu2"
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu2",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set time",
			source: `
def apply(metric):
	metric.time -= metric.time % 100000000
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(42, 42).UTC(),
			),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(42, 0).UTC(),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Starlark{
				Source:  tt.source,
				OnError: "drop",
				Log:     testutil.Logger{},
			}
			err := plugin.Init()
			require.NoError(t, err)

			actual := plugin.Apply(tt.input)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func BenchmarkNoop(b *testing.B) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	for n := 0; n < b.N; n++ {
		_ = plugin.Apply(metrics...)
	}
}

func TestRename(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	metric.name = "howdy"
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"howdy",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(metrics...)

	testutil.RequireMetricsEqual(t, expected, actual)
}

func BenchmarkRename(b *testing.B) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	metric.name = "howdy"
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	for n := 0; n < b.N; n++ {
		plugin.Apply(metrics...)
	}
}

func TestSetTime(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	metric.time = 42
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 42),
		),
	}

	actual := plugin.Apply(metrics...)

	testutil.RequireMetricsEqual(t, expected, actual)
}

func TestGetTag(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	print(metric.tags['host'])
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"host": "example.org",
			},
			map[string]interface{}{
				"time_idle": 0,
			},
			time.Unix(0, 0),
		),
	}

	_ = plugin.Apply(metrics...)
}

func TestTagMapping(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	for k in metric.tags:
		print(k)
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"host": "example.org",
				"cpu":  "cpu0",
			},
			map[string]interface{}{
				"time_idle": 0,
			},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(metrics...)
	_ = actual
}

func TestTagMappingItems(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	print(dir(dict()))
	for k, v in {'x': 1}.items():
		print('items: ', k, v)
	print(type({'x': 1}.items))
	print(dir(metric.tags))
	for k in metric.tags.keys():
		print(k)
	for k,v in metric.tags.items():
		print('items:',k,v)
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"host": "example.org",
				"cpu":  "cpu0",
			},
			map[string]interface{}{
				"time_idle": 0,
			},
			time.Unix(0, 0),
		),
	}

	actual := plugin.Apply(metrics...)
	_ = actual
}

func TestTags(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "drop",
			source: `
def apply(metric):
	pass
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{},
		},
		{
			name: "passthrough",
			source: `
def apply(metric):
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set tag",
			source: `
def apply(metric):
	metric.tags['host'] = 'example.org'
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "get tag",
			source: `
def apply(metric):
	metric.tags['set'] = metric.tags['cpu']
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
						"set": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "clear",
			source: `
def apply(metric):
	metric.tags.clear()
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "iterate",
			source: `
def apply(metric):
	metric.tags['result'] = ','.join(['%s=%s' % (k, v) for k, v in metric.tags])
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":    "cpu0",
						"host":   "example.org",
						"result": "cpu=cpu0,host=example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "items",
			source: `
def apply(metric):
	metric.tags['result'] = ','.join(['%s=%s' % item for item in metric.tags.items()])
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":    "cpu0",
						"host":   "example.org",
						"result": "cpu=cpu0,host=example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "keys",
			source: `
def apply(metric):
	metric.tags['result'] = ','.join(metric.tags.keys())
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":    "cpu0",
						"host":   "example.org",
						"result": "cpu,host",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "pop",
			source: `
def apply(metric):
	metric.tags['c'] = metric.tags.pop('a')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"b": "y",
						"c": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "popitem",
			source: `
def apply(metric):
	metric.tags['c'] = '='.join(metric.tags.popitem())
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"b": "y",
						"c": "a=x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "setdefault key not set",
			source: `
def apply(metric):
	metric.tags.setdefault('c', 'z')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
						"c": "z",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "setdefault key set",
			source: `
def apply(metric):
	metric.tags['c'] = metric.tags.setdefault('a', 'z')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"c": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update",
			source: `
def apply(metric):
	metric.tags.update([('b', 'y'), ('c', 'z')], d='zz')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
						"c": "z",
						"d": "zz",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Starlark{
				Source:  tt.source,
				OnError: "drop",
				Log:     testutil.Logger{},
			}
			err := plugin.Init()
			require.NoError(t, err)

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func BenchmarkCheckTags(b *testing.B) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	metric.tags
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		plugin.Apply(metrics...)
	}
}

func BenchmarkIterateTags(b *testing.B) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	for k, v in metric.tags:
		pass
	return metric
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	if err != nil {
		panic(err)
	}

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"host": "example.org",
				"cpu":  "cpu0",
				"foo":  "bar",
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		plugin.Apply(metrics...)
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "set string field",
			source: `
def apply(metric):
	metric.fields['host'] = 'example.org'
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"host": "example.org",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "get string field",
			source: `
def apply(metric):
	value = metric.fields['value']
	if value != "xyzzy" and type(value) != "str":
		return
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": "xyzzy",
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": "xyzzy",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set integer field",
			source: `
def apply(metric):
	metric.fields['value'] = 42
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "get integer field",
			source: `
def apply(metric):
	value = metric.fields['value']
	if value != 42 and type(value) != "int":
		return
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set bool field",
			source: `
def apply(metric):
	metric.fields['value'] = True
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": true,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "get bool field",
			source: `
def apply(metric):
	value = metric.fields['value']
	if value and type(value) != "bool":
		return
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": true,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": true,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set float field",
			source: `
def apply(metric):
	metric.fields['value'] = 42.5
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.5,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "get float field",
			source: `
def apply(metric):
	value = metric.fields['value']
	if value == 42.5 and type(value) != "float":
		return
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.5,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"value": 42.5,
					},
					time.Unix(0, 0),
				),
			},
		},
		{ ///////
			name: "clear",
			source: `
def apply(metric):
	metric.tags.clear()
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "items",
			source: `
def apply(metric):
	metric.tags['result'] = ','.join(['%s=%s' % item for item in metric.tags.items()])
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":    "cpu0",
						"host":   "example.org",
						"result": "cpu=cpu0,host=example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "keys",
			source: `
def apply(metric):
	metric.tags['result'] = ','.join(metric.tags.keys())
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":  "cpu0",
						"host": "example.org",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu":    "cpu0",
						"host":   "example.org",
						"result": "cpu,host",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "pop",
			source: `
def apply(metric):
	metric.tags['c'] = metric.tags.pop('a')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"b": "y",
						"c": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "popitem",
			source: `
def apply(metric):
	metric.tags['c'] = '='.join(metric.tags.popitem())
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"b": "y",
						"c": "a=x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "setdefault key not set",
			source: `
def apply(metric):
	metric.tags.setdefault('c', 'z')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
						"c": "z",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "setdefault key set",
			source: `
def apply(metric):
	metric.tags['c'] = metric.tags.setdefault('a', 'z')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"c": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update",
			source: `
def apply(metric):
	metric.tags.update([('b', 'y'), ('c', 'z')], d='zz')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"a": "x",
						"b": "y",
						"c": "z",
						"d": "zz",
					},
					map[string]interface{}{
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Starlark{
				Source:  tt.source,
				OnError: "drop",
				Log:     testutil.Logger{},
			}
			err := plugin.Init()
			require.NoError(t, err)

			actual := plugin.Apply(tt.input...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestEmptySource(t *testing.T) {
	plugin := &Starlark{
		Source:  "",
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	require.Error(t, err)
}

func TestKeyError(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	metric.tags['foo']
`,
		OnError: "drop",
		Log:     testutil.Logger{},
	}
	err := plugin.Init()
	require.NoError(t, err)

	input := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"cpu": "cpu0",
			},
			map[string]interface{}{
				"time_idle": 0,
			},
			time.Unix(0, 0),
		),
	}
	expected := []telegraf.Metric{}

	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)
}
