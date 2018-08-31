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

func newM2() telegraf.Metric {
	m2, _ := metric.New("IIS_log",
		map[string]string{
			"verb":           "GET",
			"resp_code":      "200",
			"s-computername": "MIXEDCASE_hostname",
		},
		map[string]interface{}{
			"request":       "/mixed/CASE/paTH/?from=-1D&to=now",
			"cs-host":       "AAAbbb",
			"ignore_number": int64(200),
			"ignore_bool":   true,
		},
		time.Now(),
	)
	return m2
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
						Field:  "request",
						Cutset: "/w",
					},
				},
				Lowercase: []converter{
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
					converter{
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
			converter{
				Tag: "s-computername",
			},
			converter{
				Field: "request",
			},
			converter{
				Field: "cs-host",
				Dest:  "cs-host_lowercase",
			},
		},
		Uppercase: []converter{
			converter{
				Tag: "verb",
			},
		},
	}

	processed := plugin.Apply(newM2())

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
	}

	assert.Equal(t, expectedFields, processed[0].Fields())
	assert.Equal(t, expectedTags, processed[0].Tags())
}

func TestReadmeExample(t *testing.T) {
	plugin := &Strings{
		Lowercase: []converter{
			converter{
				Tag: "uri_stem",
			},
		},
		TrimPrefix: []converter{
			converter{
				Tag:    "uri_stem",
				Prefix: "/api/",
			},
		},
		Uppercase: []converter{
			converter{
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
