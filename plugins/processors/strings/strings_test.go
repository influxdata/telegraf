package strings

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newM1() telegraf.Metric {
	m1, _ := metric.New("IIS_log",
		map[string]string{
			"verb":           "GET",
			"s-computername": "MIXEDCASE_hostname",
		},
		map[string]interface{}{
			"request":    "/mixed/CASE/paTH/?from=-1D&to=now",
			"whitespace": "  whitespace\t",
		},
		time.Now(),
	)
	return m1
}

func TestFieldConversions(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Strings
		check  func(t *testing.T, actual telegraf.Metric)
	}{
		{
			name: "Should change existing field to lowercase",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Field: "request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/mixed/case/path/?from=-1d&to=now", fv)
			},
		},
		{
			name: "Should change existing field to uppercase",
			plugin: &Strings{
				Uppercase: []converter{
					{
						Field: "request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/MIXED/CASE/PATH/?FROM=-1D&TO=NOW", fv)
			},
		},
		{
			name: "Should add new lowercase field",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Field: "request",
						Dest:  "lowercase_request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)

				fv, ok = actual.GetField("lowercase_request")
				require.True(t, ok)
				require.Equal(t, "/mixed/case/path/?from=-1d&to=now", fv)
			},
		},
		{
			name: "Should trim from both sides",
			plugin: &Strings{
				Trim: []converter{
					{
						Field:  "request",
						Cutset: "/w",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "mixed/CASE/paTH/?from=-1D&to=no", fv)
			},
		},
		{
			name: "Should trim from both sides and make lowercase",
			plugin: &Strings{
				Trim: []converter{
					{
						Field:  "request",
						Cutset: "/w",
					},
				},
				Lowercase: []converter{
					{
						Field: "request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "mixed/case/path/?from=-1d&to=no", fv)
			},
		},
		{
			name: "Should trim from left side",
			plugin: &Strings{
				TrimLeft: []converter{
					{
						Field:  "request",
						Cutset: "/w",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should trim from right side",
			plugin: &Strings{
				TrimRight: []converter{
					{
						Field:  "request",
						Cutset: "/w",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=no", fv)
			},
		},
		{
			name: "Should trim prefix '/mixed'",
			plugin: &Strings{
				TrimPrefix: []converter{
					{
						Field:  "request",
						Prefix: "/mixed",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should trim suffix '-1D&to=now'",
			plugin: &Strings{
				TrimSuffix: []converter{
					{
						Field:  "request",
						Suffix: "-1D&to=now",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=", fv)
			},
		},
		{
			name: "Trim without cutset removes whitespace",
			plugin: &Strings{
				Trim: []converter{
					{
						Field: "whitespace",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("whitespace")
				require.True(t, ok)
				require.Equal(t, "whitespace", fv)
			},
		},
		{
			name: "Trim left without cutset removes whitespace",
			plugin: &Strings{
				TrimLeft: []converter{
					{
						Field: "whitespace",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("whitespace")
				require.True(t, ok)
				require.Equal(t, "whitespace\t", fv)
			},
		},
		{
			name: "Trim right without cutset removes whitespace",
			plugin: &Strings{
				TrimRight: []converter{
					{
						Field: "whitespace",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("whitespace")
				require.True(t, ok)
				require.Equal(t, "  whitespace", fv)
			},
		},
		{
			name: "No change if field missing",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Field:  "xyzzy",
						Suffix: "-1D&to=now",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.plugin.Apply(newM1())
			require.Len(t, metrics, 1)
			tt.check(t, metrics[0])
		})
	}
}

func TestTagConversions(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Strings
		check  func(t *testing.T, actual telegraf.Metric)
	}{
		{
			name: "Should change existing tag to lowercase",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Tag: "s-computername",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				tv, ok := actual.GetTag("verb")
				require.True(t, ok)
				require.Equal(t, "GET", tv)

				tv, ok = actual.GetTag("s-computername")
				require.True(t, ok)
				require.Equal(t, "mixedcase_hostname", tv)
			},
		},
		{
			name: "Should add new lowercase tag",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Tag:  "s-computername",
						Dest: "s-computername_lowercase",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				tv, ok := actual.GetTag("verb")
				require.True(t, ok)
				require.Equal(t, "GET", tv)

				tv, ok = actual.GetTag("s-computername")
				require.True(t, ok)
				require.Equal(t, "MIXEDCASE_hostname", tv)

				tv, ok = actual.GetTag("s-computername_lowercase")
				require.True(t, ok)
				require.Equal(t, "mixedcase_hostname", tv)
			},
		},
		{
			name: "Should add new uppercase tag",
			plugin: &Strings{
				Uppercase: []converter{
					{
						Tag:  "s-computername",
						Dest: "s-computername_uppercase",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				tv, ok := actual.GetTag("verb")
				require.True(t, ok)
				require.Equal(t, "GET", tv)

				tv, ok = actual.GetTag("s-computername")
				require.True(t, ok)
				require.Equal(t, "MIXEDCASE_hostname", tv)

				tv, ok = actual.GetTag("s-computername_uppercase")
				require.True(t, ok)
				require.Equal(t, "MIXEDCASE_HOSTNAME", tv)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.plugin.Apply(newM1())
			require.Len(t, metrics, 1)
			tt.check(t, metrics[0])
		})
	}
}

func TestMeasurementConversions(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Strings
		check  func(t *testing.T, actual telegraf.Metric)
	}{
		{
			name: "lowercase measurement",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Measurement: "IIS_log",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				name := actual.Name()
				require.Equal(t, "iis_log", name)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.plugin.Apply(newM1())
			require.Len(t, metrics, 1)
			tt.check(t, metrics[0])
		})
	}
}

func TestMultipleConversions(t *testing.T) {
	plugin := &Strings{
		Lowercase: []converter{
			{
				Tag: "s-computername",
			},
			{
				Field: "request",
			},
			{
				Field: "cs-host",
				Dest:  "cs-host_lowercase",
			},
		},
		Uppercase: []converter{
			{
				Tag: "verb",
			},
		},
		Replace: []converter{
			{
				Tag: "foo",
				Old: "a",
				New: "x",
			},
			{
				Tag: "bar",
				Old: "b",
				New: "y",
			},
		},
	}

	m, _ := metric.New("IIS_log",
		map[string]string{
			"verb":           "GET",
			"resp_code":      "200",
			"s-computername": "MIXEDCASE_hostname",
			"foo":            "a",
			"bar":            "b",
		},
		map[string]interface{}{
			"request":       "/mixed/CASE/paTH/?from=-1D&to=now",
			"cs-host":       "AAAbbb",
			"ignore_number": int64(200),
			"ignore_bool":   true,
		},
		time.Now(),
	)

	processed := plugin.Apply(m)

	expectedFields := map[string]interface{}{
		"request":           "/mixed/case/path/?from=-1d&to=now",
		"ignore_number":     int64(200),
		"ignore_bool":       true,
		"cs-host":           "AAAbbb",
		"cs-host_lowercase": "aaabbb",
	}
	expectedTags := map[string]string{
		"verb":           "GET",
		"resp_code":      "200",
		"s-computername": "mixedcase_hostname",
		"foo":            "x",
		"bar":            "y",
	}

	assert.Equal(t, expectedFields, processed[0].Fields())
	assert.Equal(t, expectedTags, processed[0].Tags())
}

func TestReadmeExample(t *testing.T) {
	plugin := &Strings{
		Lowercase: []converter{
			{
				Tag: "uri_stem",
			},
		},
		TrimPrefix: []converter{
			{
				Tag:    "uri_stem",
				Prefix: "/api/",
			},
		},
		Uppercase: []converter{
			{
				Field: "cs-host",
				Dest:  "cs-host_normalised",
			},
		},
	}

	m, _ := metric.New("iis_log",
		map[string]string{
			"verb":     "get",
			"uri_stem": "/API/HealthCheck",
		},
		map[string]interface{}{
			"cs-host":      "MIXEDCASE_host",
			"referrer":     "-",
			"ident":        "-",
			"http_version": "1.1",
			"agent":        "UserAgent",
			"resp_bytes":   int64(270),
		},
		time.Now(),
	)

	processed := plugin.Apply(m)

	expectedTags := map[string]string{
		"verb":     "get",
		"uri_stem": "healthcheck",
	}
	expectedFields := map[string]interface{}{
		"cs-host":            "MIXEDCASE_host",
		"cs-host_normalised": "MIXEDCASE_HOST",
		"referrer":           "-",
		"ident":              "-",
		"http_version":       "1.1",
		"agent":              "UserAgent",
		"resp_bytes":         int64(270),
	}

	assert.Equal(t, expectedFields, processed[0].Fields())
	assert.Equal(t, expectedTags, processed[0].Tags())
}

func newMetric(name string) telegraf.Metric {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	m, _ := metric.New(name, tags, fields, time.Now())
	return m
}

func TestMeasurementReplace(t *testing.T) {
	plugin := &Strings{
		Replace: []converter{
			{
				Old:         "_",
				New:         "-",
				Measurement: "*",
			},
		},
	}
	metrics := []telegraf.Metric{
		newMetric("foo:some_value:bar"),
		newMetric("average:cpu:usage"),
		newMetric("average_cpu_usage"),
	}
	results := plugin.Apply(metrics...)
	assert.Equal(t, "foo:some-value:bar", results[0].Name(), "`_` was not changed to `-`")
	assert.Equal(t, "average:cpu:usage", results[1].Name(), "Input name should have been unchanged")
	assert.Equal(t, "average-cpu-usage", results[2].Name(), "All instances of `_` should have been changed to `-`")
}

func TestMeasurementCharDeletion(t *testing.T) {
	plugin := &Strings{
		Replace: []converter{
			{
				Old:         "foo",
				New:         "",
				Measurement: "*",
			},
		},
	}
	metrics := []telegraf.Metric{
		newMetric("foo:bar:baz"),
		newMetric("foofoofoo"),
		newMetric("barbarbar"),
	}
	results := plugin.Apply(metrics...)
	assert.Equal(t, ":bar:baz", results[0].Name(), "Should have deleted the initial `foo`")
	assert.Equal(t, "foofoofoo", results[1].Name(), "Should have refused to delete the whole string")
	assert.Equal(t, "barbarbar", results[2].Name(), "Should not have changed the input")
}
