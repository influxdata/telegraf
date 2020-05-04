package starlark

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestStarlarkInitError(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Starlark
	}{
		{
			name: "source must define apply",
			plugin: &Starlark{
				Source:  "",
				OnError: "drop",
				Log:     testutil.Logger{},
			},
		},
		{
			name: "apply must be a function",
			plugin: &Starlark{
				Source: `
apply = 42
`,
				OnError: "drop",
				Log:     testutil.Logger{},
			},
		},
		{
			name: "apply function must take one arg",
			plugin: &Starlark{
				Source: `
def apply():
	pass
`,
				OnError: "drop",
				Log:     testutil.Logger{},
			},
		},
		{
			name: "package scope must have valid syntax",
			plugin: &Starlark{
				Source: `
for
`,
				OnError: "drop",
				Log:     testutil.Logger{},
			},
		},
		{
			name: "OnError choice is unsupported value",
			plugin: &Starlark{
				Source: `
def apply(metric):
	pass
`,
				OnError: "foo",
				Log:     testutil.Logger{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Init()
			require.Error(t, err)
		})
	}
}

var tests = []struct {
	name     string
	source   string
	input    []telegraf.Metric
	expected []telegraf.Metric
}{
	{
		name: "drop",
		source: `
def apply(metric):
	return None
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{"time_idle": 42},
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
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{"time_idle": 42},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{"time_idle": 42},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "new",
		source: `
def apply(metric):
	m = Metric('cpu')
	m.fields['time_guest'] = 2.0
	m.time = 0
	return m
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_guest": 2.0,
				},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "illegal metric duplicate reference", // FIXME: detect and error?
		source: `
def apply(metric):
	return [metric, metric]
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "deepcopy", // TODO: refcount
		source: `
def apply(metric):
	return [metric, deepcopy(metric)]
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "save to global scope", // FIXME
		source: `
cache = []

def apply(metric):
	# is not None not allowed in starlark
	result = None
	if cache:
		result = cache.pop()
		print(result.fields)
	cache.append(metric)
	return result
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 1.0,
				},
				time.Unix(0, 0),
			),
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 2.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 1.0,
				},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "save to global scope 2", // FIXME
		source: `
cache = []

def apply(metric):
	cache.append(metric)
	print(len(cache))
	return None
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 1.0,
				},
				time.Unix(0, 0),
			),
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 2.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{},
	},
	{
		name: "read from global scope",
		source: `
name = "howdy"

def apply(metric):
	metric.name = name
	return metric
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("howdy",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "set metric name",
		source: `
def apply(metric):
	metric.name = "howdy"
	return metric
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{"time_idle": 42.0},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("howdy",
				map[string]string{},
				map[string]interface{}{"time_idle": 42.0},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "getattr tags",
		source: `
def apply(metric):
	metric.tags
	return metric
`,
		input: []telegraf.Metric{
			testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "iterate tags",
		source: `
def apply(metric):
	for k, v in metric.tags:
		pass
	return metric
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"host": "example.org",
					"cpu":  "cpu0",
					"foo":  "bar",
				},
				map[string]interface{}{"time_idle": 42.0},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"host": "example.org",
					"cpu":  "cpu0",
					"foo":  "bar",
				},
				map[string]interface{}{"time_idle": 42.0},
				time.Unix(0, 0),
			),
		},
	},
	{
		name: "copy tags to fields",
		source: `
def apply(metric):
	for k, v in metric.tags.items():
		metric.fields[k] = v
	return metric
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{
					"host": "example.org",
					"cpu":  "cpu0",
				},
				map[string]interface{}{"time_idle": 42},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
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
		name: "iterate fields",
		source: `
def apply(metric):
	for k, v in metric.fields:
		pass
	return metric
`,
		input: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"cores":     4,
					"load":      42.5,
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
		expected: []telegraf.Metric{
			testutil.MustMetric("cpu",
				map[string]string{},
				map[string]interface{}{
					"cores":     4,
					"load":      42.5,
					"time_idle": 42.0,
				},
				time.Unix(0, 0),
			),
		},
	},
}

func TestStarlark(t *testing.T) {
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

func BenchmarkStarlark(b *testing.B) {
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			plugin := &Starlark{
				Source:  tt.source,
				OnError: "drop",
				Log:     testutil.Logger{},
			}

			err := plugin.Init()
			if err != nil {
				panic(err)
			}

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				plugin.Apply(tt.input...)
			}
		})
	}
}

// --- Metric implementation testing --
func TestMetricAccess(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		input    telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "get/set name",
			source: `
def apply(metric):
	metric.name = metric.name + "2"
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
			name: "get/set time",
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
		{
			name: "get tags",
			source: `
def apply(metric):
	print("[metric/get tags] "+metric.tags['host'])
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
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
					},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set tags", // This should fail
			source: `
def apply(metric):
	metric.tags = {}
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{},
		},
		{
			name: "get fields",
			source: `
def apply(metric):
	print("[metric/get fields] {}".format(metric.fields['time_idle']))
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
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
					},
					map[string]interface{}{
						"time_idle": 42,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "set fields", // This should fail
			source: `
def apply(metric):
	metric.fields = {}
	return metric
			`,
			input: testutil.MustMetric(
				"cpu",
				map[string]string{
					"host": "example.org",
				},
				map[string]interface{}{
					"time_idle": 42,
				},
				time.Unix(0, 0),
			),
			expected: []telegraf.Metric{},
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

// --- Fieldset implementation testing --
func TestFieldsGetSet(t *testing.T) {
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

func TestFieldsBuiltins(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "clear",
			source: `
def apply(metric):
	metric.fields.clear()
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
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "get",
			source: `
def apply(metric):
	metric.fields["default"] = metric.fields.get("load", 42.5)
	metric.fields["actual"]  = metric.fields.get("time_idle", 4)
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
						"default":   42.5,
						"actual":    0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "items",
			source: `
def apply(metric):
	items = ['{}={}'.format(k,v) for k,v in metric.fields.items()]
	metric.fields['result'] = ','.join(sorted(items))
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"active":    true,
						"cores":     4,
						"load":      42.5,
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
						"active":    true,
						"cores":     4,
						"load":      42.5,
						"result":    "active=True,cores=4,load=42.5,time_idle=0",
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
	metric.fields['result'] = ','.join(sorted(metric.fields.keys()))
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"active":    true,
						"cores":     4,
						"load":      42.5,
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
						"active":    true,
						"cores":     4,
						"load":      42.5,
						"result":    "active,cores,load,time_idle",
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
	metric.fields['result'] = metric.fields.pop('load')
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"active":    true,
						"cores":     4,
						"load":      42.5,
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
						"active":    true,
						"cores":     4,
						"time_idle": 0,
						"result":    42.5,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "popitem",
			source: `
def apply(metric):
	k,v = metric.fields.popitem()
	metric.fields['result'] = '{}={}'.format(k,v)
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"time_idle": 42,
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
						"result": "time_idle=42",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "setdefault key not set",
			source: `
def apply(metric):
	metric.fields['result'] = metric.fields.setdefault('age', 10)
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
						"age":       10,
						"result":    10,
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
	metric.fields['result'] = metric.fields.setdefault('age', 10)
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"age":       1,
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
						"age":       1,
						"result":    1,
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update list only",
			source: `
def apply(metric):
	metric.fields.update([('a', 1), ('b', 42.5)])
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
						"a":         1,
						"b":         42.5,
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update dict only",
			source: `
def apply(metric):
	metric.fields.update([('a', 1), ('b', 42.5)])
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
						"a":         1,
						"b":         42.5,
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update kwargs only",
			source: `
def apply(metric):
	metric.fields.update(None, c=True, d='zz')
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
						"c":         true,
						"d":         "zz",
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update list and kwargs",
			source: `
def apply(metric):
	metric.fields.update([('a', 1), ('b', 42.5)], c=True, d='zz')
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
						"a":         1,
						"b":         42.5,
						"c":         true,
						"d":         "zz",
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "update dict and kwargs",
			source: `
def apply(metric):
	values = dict([('a', 1), ('b', 42.5)])
	metric.fields.update(values, c=True, d='zz')
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
						"a":         1,
						"b":         42.5,
						"c":         true,
						"d":         "zz",
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "values",
			source: `
def apply(metric):
	metric.tags['result'] = ','.join(sorted([str(v) for v in metric.fields.values()]))
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"active":    true,
						"cores":     4,
						"load":      42.5,
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
						"result": "0,4,42.5,True",
					},
					map[string]interface{}{
						"active":    true,
						"cores":     4,
						"load":      42.5,
						"time_idle": 0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "length",
			source: `
def apply(metric):
	metric.fields['result'] = len(metric.fields.values())
	return metric
`,
			input: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{
						"cpu": "cpu0",
					},
					map[string]interface{}{
						"active":    true,
						"cores":     4,
						"load":      42.5,
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
						"active":    true,
						"cores":     4,
						"load":      42.5,
						"time_idle": 0,
						"result":    4,
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

func TestFieldsKeyError(t *testing.T) {
	plugin := &Starlark{
		Source: `
def apply(metric):
	metric.fields['foo']
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

// --- Tagset implementation testing --
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
