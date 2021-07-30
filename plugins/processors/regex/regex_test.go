package regex

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
	m1 := metric.New("access_log",
		map[string]string{
			"verb":      "GET",
			"resp_code": "200",
		},
		map[string]interface{}{
			"request": "/users/42/",
		},
		time.Now(),
	)
	return m1
}

func newM2() telegraf.Metric {
	m2 := metric.New("access_log",
		map[string]string{
			"verb":      "GET",
			"resp_code": "200",
		},
		map[string]interface{}{
			"request":       "/api/search/?category=plugins&q=regex&sort=asc",
			"ignore_number": int64(200),
			"ignore_bool":   true,
		},
		time.Now(),
	)
	return m2
}

func TestFieldConversions(t *testing.T) {
	tests := []struct {
		message        string
		converter      converter
		expectedFields map[string]interface{}
	}{
		{
			message: "Should change existing field",
			converter: converter{
				Key:         "request",
				Pattern:     "^/users/\\d+/$",
				Replacement: "/users/{id}/",
			},
			expectedFields: map[string]interface{}{
				"request": "/users/{id}/",
			},
		},
		{
			message: "Should add new field",
			converter: converter{
				Key:         "request",
				Pattern:     "^/users/\\d+/$",
				Replacement: "/users/{id}/",
				ResultKey:   "normalized_request",
			},
			expectedFields: map[string]interface{}{
				"request":            "/users/42/",
				"normalized_request": "/users/{id}/",
			},
		},
	}

	for _, test := range tests {
		regex := Regex{
			Fields: []converter{
				test.converter,
			},
		}
		require.NoError(t, regex.Init())

		processed := regex.Apply(newM1())

		expectedTags := map[string]string{
			"verb":      "GET",
			"resp_code": "200",
		}

		assert.Equal(t, test.expectedFields, processed[0].Fields(), test.message)
		assert.Equal(t, expectedTags, processed[0].Tags(), "Should not change tags")
		assert.Equal(t, "access_log", processed[0].Name(), "Should not change name")
	}
}

func TestTagConversions(t *testing.T) {
	tests := []struct {
		message      string
		converter    converter
		expectedTags map[string]string
	}{
		{
			message: "Should change existing tag",
			converter: converter{
				Key:         "resp_code",
				Pattern:     "^(\\d)\\d\\d$",
				Replacement: "${1}xx",
			},
			expectedTags: map[string]string{
				"verb":      "GET",
				"resp_code": "2xx",
			},
		},
		{
			message: "Should append to existing tag",
			converter: converter{
				Key:         "verb",
				Pattern:     "^(.*)$",
				Replacement: " (${1})",
				ResultKey:   "resp_code",
				Append:      true,
			},
			expectedTags: map[string]string{
				"verb":      "GET",
				"resp_code": "200 (GET)",
			},
		},
		{
			message: "Should add new tag",
			converter: converter{
				Key:         "resp_code",
				Pattern:     "^(\\d)\\d\\d$",
				Replacement: "${1}xx",
				ResultKey:   "resp_code_group",
			},
			expectedTags: map[string]string{
				"verb":            "GET",
				"resp_code":       "200",
				"resp_code_group": "2xx",
			},
		},
	}

	for _, test := range tests {
		regex := Regex{
			Tags: []converter{
				test.converter,
			},
		}
		require.NoError(t, regex.Init())

		processed := regex.Apply(newM1())

		expectedFields := map[string]interface{}{
			"request": "/users/42/",
		}

		assert.Equal(t, expectedFields, processed[0].Fields(), test.message, "Should not change fields")
		assert.Equal(t, test.expectedTags, processed[0].Tags(), test.message)
		assert.Equal(t, "access_log", processed[0].Name(), "Should not change name")
	}
}

func TestMetricConversions(t *testing.T) {
	inputTemplate := []telegraf.Metric{
		testutil.MustMetric(
			"access_log",
			map[string]string{
				"verb":      "GET",
				"resp_code": "200",
			},
			map[string]interface{}{
				"request": "/users/42/",
			},
			time.Unix(1627646243, 0),
		),
		testutil.MustMetric(
			"access_log",
			map[string]string{
				"verb":      "GET",
				"resp_code": "200",
			},
			map[string]interface{}{
				"request":       "/api/search/?category=plugins&q=regex&sort=asc",
				"ignore_number": int64(200),
				"ignore_bool":   true,
			},
			time.Unix(1627646253, 0),
		),
		testutil.MustMetric(
			"error_log",
			map[string]string{
				"verb":      "GET",
				"resp_code": "404",
			},
			map[string]interface{}{
				"request":       "/api/search/?category=plugins&q=regex&sort=asc",
				"ignore_number": int64(404),
				"ignore_flag":   true,
				"error_message": "request too silly",
			},
			time.Unix(1627646263, 0),
		),
	}

	tests := []struct {
		name      string
		converter converter
		expected  []telegraf.Metric
	}{
		{
			name: "Should change metric name",
			converter: converter{
				Key:         "measurement",
				Pattern:     "^(\\w+)_log$",
				Replacement: "${1}",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(200),
						"ignore_bool":   true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error",
					map[string]string{
						"verb":      "GET",
						"resp_code": "404",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(404),
						"ignore_flag":   true,
						"error_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
		{
			name: "Should change field name",
			converter: converter{
				Key:         "fields",
				Pattern:     "^(?:ignore|error)_(\\w+)$",
				Replacement: "result_${1}",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"result_number": int64(200),
						"result_bool":   true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "404",
					},
					map[string]interface{}{
						"request":        "/api/search/?category=plugins&q=regex&sort=asc",
						"result_number":  int64(404),
						"result_flag":    true,
						"result_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
		{
			name: "Should change tag name",
			converter: converter{
				Key:         "tags",
				Pattern:     "^resp_(\\w+)$",
				Replacement: "${1}",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb": "GET",
						"code": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb": "GET",
						"code": "200",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(200),
						"ignore_bool":   true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error_log",
					map[string]string{
						"verb": "GET",
						"code": "404",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(404),
						"ignore_flag":   true,
						"error_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
		{
			name: "Should keep existing field name",
			converter: converter{
				Key:         "fields",
				Pattern:     "^(?:ignore|error)_(\\w+)$",
				Replacement: "request",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(200),
						"ignore_bool":   true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "404",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(404),
						"ignore_flag":   true,
						"error_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
		{
			name: "Should keep existing tag name",
			converter: converter{
				Key:         "tags",
				Pattern:     "^resp_(\\w+)$",
				Replacement: "verb",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(200),
						"ignore_bool":   true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "404",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(404),
						"ignore_flag":   true,
						"error_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
		{
			name: "Should overwrite existing field name",
			converter: converter{
				Key:         "fields",
				Pattern:     "^ignore_bool$",
				Replacement: "request",
				ResultKey:   "overwrite",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "200",
					},
					map[string]interface{}{
						"ignore_number": int64(200),
						"request":       true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error_log",
					map[string]string{
						"verb":      "GET",
						"resp_code": "404",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(404),
						"ignore_flag":   true,
						"error_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
		{
			name: "Should overwrite existing tag name",
			converter: converter{
				Key:         "tags",
				Pattern:     "^resp_(\\w+)$",
				Replacement: "verb",
				ResultKey:   "overwrite",
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb": "200",
					},
					map[string]interface{}{
						"request": "/users/42/",
					},
					time.Unix(1627646243, 0),
				),
				testutil.MustMetric(
					"access_log",
					map[string]string{
						"verb": "200",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(200),
						"ignore_bool":   true,
					},
					time.Unix(1627646253, 0),
				),
				testutil.MustMetric(
					"error_log",
					map[string]string{
						"verb": "404",
					},
					map[string]interface{}{
						"request":       "/api/search/?category=plugins&q=regex&sort=asc",
						"ignore_number": int64(404),
						"ignore_flag":   true,
						"error_message": "request too silly",
					},
					time.Unix(1627646263, 0),
				),
			},
		},
	}

	for _, test := range tests {
		// Copy the inputs as they will be modified by the processor
		input := make([]telegraf.Metric, len(inputTemplate))
		for i, m := range inputTemplate {
			input[i] = m.Copy()
		}

		t.Run(test.name, func(t *testing.T) {
			regex := Regex{Metrics: []converter{test.converter}}
			require.NoError(t, regex.Init())

			actual := regex.Apply(input...)
			testutil.RequireMetricsEqual(t, test.expected, actual)
		})
	}
}

func TestMultipleConversions(t *testing.T) {
	regex := Regex{
		Tags: []converter{
			{
				Key:         "resp_code",
				Pattern:     "^(\\d)\\d\\d$",
				Replacement: "${1}xx",
				ResultKey:   "resp_code_group",
			},
			{
				Key:         "resp_code_group",
				Pattern:     "2xx",
				Replacement: "OK",
				ResultKey:   "resp_code_text",
			},
		},
		Fields: []converter{
			{
				Key:         "request",
				Pattern:     "^/api(?P<method>/[\\w/]+)\\S*",
				Replacement: "${method}",
				ResultKey:   "method",
			},
			{
				Key:         "request",
				Pattern:     ".*category=(\\w+).*",
				Replacement: "${1}",
				ResultKey:   "search_category",
			},
		},
	}
	require.NoError(t, regex.Init())

	processed := regex.Apply(newM2())

	expectedFields := map[string]interface{}{
		"request":         "/api/search/?category=plugins&q=regex&sort=asc",
		"method":          "/search/",
		"search_category": "plugins",
		"ignore_number":   int64(200),
		"ignore_bool":     true,
	}
	expectedTags := map[string]string{
		"verb":            "GET",
		"resp_code":       "200",
		"resp_code_group": "2xx",
		"resp_code_text":  "OK",
	}

	assert.Equal(t, expectedFields, processed[0].Fields())
	assert.Equal(t, expectedTags, processed[0].Tags())
}

func TestNoMatches(t *testing.T) {
	tests := []struct {
		message        string
		converter      converter
		expectedFields map[string]interface{}
	}{
		{
			message: "Should not change anything if there is no field with given key",
			converter: converter{
				Key:         "not_exists",
				Pattern:     "\\.*",
				Replacement: "x",
			},
			expectedFields: map[string]interface{}{
				"request": "/users/42/",
			},
		},
		{
			message: "Should not change anything if regex doesn't match",
			converter: converter{
				Key:         "request",
				Pattern:     "not_match",
				Replacement: "x",
			},
			expectedFields: map[string]interface{}{
				"request": "/users/42/",
			},
		},
		{
			message: "Should not emit new tag/field when result_key given but regex doesn't match",
			converter: converter{
				Key:         "request",
				Pattern:     "not_match",
				Replacement: "x",
				ResultKey:   "new_field",
			},
			expectedFields: map[string]interface{}{
				"request": "/users/42/",
			},
		},
	}

	for _, test := range tests {
		regex := Regex{
			Fields: []converter{
				test.converter,
			},
		}
		require.NoError(t, regex.Init())

		processed := regex.Apply(newM1())

		assert.Equal(t, test.expectedFields, processed[0].Fields(), test.message)
	}
}

func BenchmarkConversions(b *testing.B) {
	regex := Regex{
		Tags: []converter{
			{
				Key:         "resp_code",
				Pattern:     "^(\\d)\\d\\d$",
				Replacement: "${1}xx",
				ResultKey:   "resp_code_group",
			},
		},
		Fields: []converter{
			{
				Key:         "request",
				Pattern:     "^/users/\\d+/$",
				Replacement: "/users/{id}/",
			},
		},
	}
	require.NoError(b, regex.Init())

	for n := 0; n < b.N; n++ {
		processed := regex.Apply(newM1())
		_ = processed
	}
}
