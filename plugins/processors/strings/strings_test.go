package strings

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
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

func newM2() telegraf.Metric {
	m1, _ := metric.New("IIS_log",
		map[string]string{
			"verb":           "GET",
			"S-ComputerName": "MIXEDCASE_hostname",
		},
		map[string]interface{}{
			"Request":      "/mixed/CASE/paTH/?from=-1D&to=now",
			"req/sec":      5,
			" whitespace ": "  whitespace\t",
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
			name: "Should change existing field to titlecase",
			plugin: &Strings{
				Titlecase: []converter{
					{
						Field: "request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/Mixed/CASE/PaTH/?From=-1D&To=Now", fv)
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

func TestFieldKeyConversions(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Strings
		check  func(t *testing.T, actual telegraf.Metric)
	}{
		{
			name: "Should change existing field key to lowercase",
			plugin: &Strings{
				Lowercase: []converter{
					{
						FieldKey: "Request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should change existing field key to uppercase",
			plugin: &Strings{
				Uppercase: []converter{
					{
						FieldKey: "Request",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("Request")
				require.False(t, ok)

				fv, ok = actual.GetField("REQUEST")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should trim from both sides",
			plugin: &Strings{
				Trim: []converter{
					{
						FieldKey: "Request",
						Cutset:   "eR",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("quest")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should trim from both sides but not make lowercase",
			plugin: &Strings{
				// Tag/field key multiple executions occur in the following order: (initOnce)
				//   Lowercase
				//   Uppercase
				//   Titlecase
				//   Trim
				//   TrimLeft
				//   TrimRight
				//   TrimPrefix
				//   TrimSuffix
				//   Replace
				Lowercase: []converter{
					{
						FieldKey: "Request",
					},
				},
				Trim: []converter{
					{
						FieldKey: "request",
						Cutset:   "tse",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("requ")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should trim from left side",
			plugin: &Strings{
				TrimLeft: []converter{
					{
						FieldKey: "req/sec",
						Cutset:   "req/",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("sec")
				require.True(t, ok)
				require.Equal(t, int64(5), fv)
			},
		},
		{
			name: "Should trim from right side",
			plugin: &Strings{
				TrimRight: []converter{
					{
						FieldKey: "req/sec",
						Cutset:   "req/",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("req/sec")
				require.True(t, ok)
				require.Equal(t, int64(5), fv)
			},
		},
		{
			name: "Should trim prefix 'req/'",
			plugin: &Strings{
				TrimPrefix: []converter{
					{
						FieldKey: "req/sec",
						Prefix:   "req/",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("sec")
				require.True(t, ok)
				require.Equal(t, int64(5), fv)
			},
		},
		{
			name: "Should trim suffix '/sec'",
			plugin: &Strings{
				TrimSuffix: []converter{
					{
						FieldKey: "req/sec",
						Suffix:   "/sec",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("req")
				require.True(t, ok)
				require.Equal(t, int64(5), fv)
			},
		},
		{
			name: "Trim without cutset removes whitespace",
			plugin: &Strings{
				Trim: []converter{
					{
						FieldKey: " whitespace ",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("whitespace")
				require.True(t, ok)
				require.Equal(t, "  whitespace\t", fv)
			},
		},
		{
			name: "Trim left without cutset removes whitespace",
			plugin: &Strings{
				TrimLeft: []converter{
					{
						FieldKey: " whitespace ",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("whitespace ")
				require.True(t, ok)
				require.Equal(t, "  whitespace\t", fv)
			},
		},
		{
			name: "Trim right without cutset removes whitespace",
			plugin: &Strings{
				TrimRight: []converter{
					{
						FieldKey: " whitespace ",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField(" whitespace")
				require.True(t, ok)
				require.Equal(t, "  whitespace\t", fv)
			},
		},
		{
			name: "No change if field missing",
			plugin: &Strings{
				Lowercase: []converter{
					{
						FieldKey: "xyzzy",
						Suffix:   "-1D&to=now",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("Request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
		{
			name: "Should trim the existing field to 6 characters",
			plugin: &Strings{
				Left: []converter{
					{
						Field: "Request",
						Width: 6,
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("Request")
				require.True(t, ok)
				require.Equal(t, "/mixed", fv)
			},
		},
		{
			name: "Should do nothing to the string",
			plugin: &Strings{
				Left: []converter{
					{
						Field: "Request",
						Width: 600,
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				fv, ok := actual.GetField("Request")
				require.True(t, ok)
				require.Equal(t, "/mixed/CASE/paTH/?from=-1D&to=now", fv)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.plugin.Apply(newM2())
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
		{
			name: "Should add new titlecase tag",
			plugin: &Strings{
				Titlecase: []converter{
					{
						Tag:  "s-computername",
						Dest: "s-computername_titlecase",
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

				tv, ok = actual.GetTag("s-computername_titlecase")
				require.True(t, ok)
				require.Equal(t, "MIXEDCASE_hostname", tv)
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

func TestTagKeyConversions(t *testing.T) {
	tests := []struct {
		name   string
		plugin *Strings
		check  func(t *testing.T, actual telegraf.Metric)
	}{
		{
			name: "Should change existing tag key to lowercase",
			plugin: &Strings{
				Lowercase: []converter{
					{
						Tag:    "S-ComputerName",
						TagKey: "S-ComputerName",
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
			name: "Should add new lowercase tag key",
			plugin: &Strings{
				Lowercase: []converter{
					{
						TagKey: "S-ComputerName",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				tv, ok := actual.GetTag("verb")
				require.True(t, ok)
				require.Equal(t, "GET", tv)

				tv, ok = actual.GetTag("S-ComputerName")
				require.False(t, ok)

				tv, ok = actual.GetTag("s-computername")
				require.True(t, ok)
				require.Equal(t, "MIXEDCASE_hostname", tv)
			},
		},
		{
			name: "Should add new uppercase tag key",
			plugin: &Strings{
				Uppercase: []converter{
					{
						TagKey: "S-ComputerName",
					},
				},
			},
			check: func(t *testing.T, actual telegraf.Metric) {
				tv, ok := actual.GetTag("verb")
				require.True(t, ok)
				require.Equal(t, "GET", tv)

				tv, ok = actual.GetTag("S-ComputerName")
				require.False(t, ok)

				tv, ok = actual.GetTag("S-COMPUTERNAME")
				require.True(t, ok)
				require.Equal(t, "MIXEDCASE_hostname", tv)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := tt.plugin.Apply(newM2())
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
		Titlecase: []converter{
			{
				Field: "status",
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
			"status":        "green",
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
		"status":            "Green",
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

func TestBase64Decode(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *Strings
		metric   []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "base64decode success",
			plugin: &Strings{
				Base64Decode: []converter{
					{
						Field: "message",
					},
				},
			},
			metric: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"message": "aG93ZHk=",
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"message": "howdy",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "base64decode not valid base64 returns original string",
			plugin: &Strings{
				Base64Decode: []converter{
					{
						Field: "message",
					},
				},
			},
			metric: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"message": "_not_base64_",
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"message": "_not_base64_",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "base64decode not valid utf-8 returns original string",
			plugin: &Strings{
				Base64Decode: []converter{
					{
						Field: "message",
					},
				},
			},
			metric: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"message": "//5oAG8AdwBkAHkA",
					},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"cpu",
					map[string]string{},
					map[string]interface{}{
						"message": "//5oAG8AdwBkAHkA",
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.plugin.Apply(tt.metric...)
			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}
