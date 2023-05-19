package carbon2

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	testcases := []struct {
		format   string
		expected string
	}{
		{
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu field=usage_idle cpu=cpu0  91.5 %d\n", now.Unix()),
		},
		{
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu_usage_idle cpu=cpu0  91.5 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeMetricWithEmptyStringTag(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m := metric.New("cpu", tags, fields, now)

	testcases := []struct {
		format   string
		expected string
	}{
		{
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu field=usage_idle cpu=null  91.5 %d\n", now.Unix()),
		},
		{
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu_usage_idle cpu=null  91.5 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeWithSpaces(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu 0": "cpu 0",
	}
	fields := map[string]interface{}{
		"usage_idle 1": float64(91.5),
	}
	m := metric.New("cpu metric", tags, fields, now)

	testcases := []struct {
		format   string
		expected string
	}{
		{
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu_metric field=usage_idle_1 cpu_0=cpu_0  91.5 %d\n", now.Unix()),
		},
		{
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu_metric_usage_idle_1 cpu_0=cpu_0  91.5 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeMetricInt(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": int64(90),
	}
	m := metric.New("cpu", tags, fields, now)

	testcases := []struct {
		format   string
		expected string
	}{
		{
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu field=usage_idle cpu=cpu0  90 %d\n", now.Unix()),
		},
		{
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu_usage_idle cpu=cpu0  90 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeMetricString(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": "foobar",
	}
	m := metric.New("cpu", tags, fields, now)

	testcases := []struct {
		format   string
		expected string
	}{
		{
			format:   "field_separate",
			expected: "",
		},
		{
			format:   "metric_includes_field",
			expected: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeMetricBool(t *testing.T) {
	requireMetric := func(tim time.Time, value bool) telegraf.Metric {
		tags := map[string]string{
			"tag_name": "tag_value",
		}
		fields := map[string]interface{}{
			"java_lang_GarbageCollector_Valid": value,
		}

		m := metric.New("cpu", tags, fields, tim)

		return m
	}

	now := time.Now()

	testcases := []struct {
		metric   telegraf.Metric
		format   string
		expected string
	}{
		{
			metric:   requireMetric(now, false),
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu field=java_lang_GarbageCollector_Valid tag_name=tag_value  0 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(now, false),
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu_java_lang_GarbageCollector_Valid tag_name=tag_value  0 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(now, true),
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu field=java_lang_GarbageCollector_Valid tag_name=tag_value  1 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(now, true),
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu_java_lang_GarbageCollector_Valid tag_name=tag_value  1 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.Serialize(tc.metric)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeBatch(t *testing.T) {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42,
		},
		time.Unix(0, 0),
	)

	metrics := []telegraf.Metric{m, m}

	testcases := []struct {
		format   string
		expected string
	}{
		{
			format: "field_separate",
			expected: `metric=cpu field=value  42 0
metric=cpu field=value  42 0
`,
		},
		{
			format: "metric_includes_field",
			expected: `metric=cpu_value  42 0
metric=cpu_value  42 0
`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s := &Serializer{
				Format: tc.format,
			}
			require.NoError(t, s.Init())

			buf, err := s.SerializeBatch(metrics)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeMetricIsProperlySanitized(t *testing.T) {
	now := time.Now()

	testcases := []struct {
		metricFunc  func() telegraf.Metric
		format      string
		expected    string
		replaceChar string
		expectedErr bool
	}{
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1", nil, fields, now)
			},
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu:1 field=usage_idle  91.5 %d\n", now.Unix()),
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1", nil, fields, now)
			},
			format:      "field_separate",
			expected:    fmt.Sprintf("metric=cpu_1 field=usage_idle  91.5 %d\n", now.Unix()),
			replaceChar: "_",
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom", nil, fields, now)
			},
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu:1:tmp:custom field=usage_idle  91.5 %d\n", now.Unix()),
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:   "field_separate",
			expected: fmt.Sprintf("metric=cpu:1:tmp:custom:namespace field=usage_idle  91.5 %d\n", now.Unix()),
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:   "metric_includes_field",
			expected: fmt.Sprintf("metric=cpu:1:tmp:custom:namespace_usage_idle  91.5 %d\n", now.Unix()),
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:      "metric_includes_field",
			expected:    fmt.Sprintf("metric=cpu_1_tmp_custom_namespace_usage_idle  91.5 %d\n", now.Unix()),
			replaceChar: "_",
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:      "metric_includes_field",
			expectedErr: true,
			replaceChar: "___",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			m := tc.metricFunc()

			s := &Serializer{
				Format:              tc.format,
				SanitizeReplaceChar: tc.replaceChar,
			}
			err := s.Init()
			if tc.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}
