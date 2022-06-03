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
		format   format
		expected string
	}{
		{
			format:   Carbon2FormatFieldSeparate,
			expected: fmt.Sprintf("metric=cpu field=usage_idle cpu=cpu0  91.5 %d\n", now.Unix()),
		},
		{
			format:   Carbon2FormatMetricIncludesField,
			expected: fmt.Sprintf("metric=cpu_usage_idle cpu=cpu0  91.5 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

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
		format   format
		expected string
	}{
		{
			format:   Carbon2FormatFieldSeparate,
			expected: fmt.Sprintf("metric=cpu field=usage_idle cpu=null  91.5 %d\n", now.Unix()),
		},
		{
			format:   Carbon2FormatMetricIncludesField,
			expected: fmt.Sprintf("metric=cpu_usage_idle cpu=null  91.5 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

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
		format   format
		expected string
	}{
		{
			format:   Carbon2FormatFieldSeparate,
			expected: fmt.Sprintf("metric=cpu_metric field=usage_idle_1 cpu_0=cpu_0  91.5 %d\n", now.Unix()),
		},
		{
			format:   Carbon2FormatMetricIncludesField,
			expected: fmt.Sprintf("metric=cpu_metric_usage_idle_1 cpu_0=cpu_0  91.5 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

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
		format   format
		expected string
	}{
		{
			format:   Carbon2FormatFieldSeparate,
			expected: fmt.Sprintf("metric=cpu field=usage_idle cpu=cpu0  90 %d\n", now.Unix()),
		},
		{
			format:   Carbon2FormatMetricIncludesField,
			expected: fmt.Sprintf("metric=cpu_usage_idle cpu=cpu0  90 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

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
		format   format
		expected string
	}{
		{
			format:   Carbon2FormatFieldSeparate,
			expected: "",
		},
		{
			format:   Carbon2FormatMetricIncludesField,
			expected: "",
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeMetricBool(t *testing.T) {
	requireMetric := func(t *testing.T, tim time.Time, value bool) telegraf.Metric {
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
		format   format
		expected string
	}{
		{
			metric:   requireMetric(t, now, false),
			format:   Carbon2FormatFieldSeparate,
			expected: fmt.Sprintf("metric=cpu field=java_lang_GarbageCollector_Valid tag_name=tag_value  0 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(t, now, false),
			format:   Carbon2FormatMetricIncludesField,
			expected: fmt.Sprintf("metric=cpu_java_lang_GarbageCollector_Valid tag_name=tag_value  0 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(t, now, true),
			format:   Carbon2FormatFieldSeparate,
			expected: fmt.Sprintf("metric=cpu field=java_lang_GarbageCollector_Valid tag_name=tag_value  1 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(t, now, true),
			format:   Carbon2FormatMetricIncludesField,
			expected: fmt.Sprintf("metric=cpu_java_lang_GarbageCollector_Valid tag_name=tag_value  1 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

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
		format   format
		expected string
	}{
		{
			format: Carbon2FormatFieldSeparate,
			expected: `metric=cpu field=value  42 0
metric=cpu field=value  42 0
`,
		},
		{
			format: Carbon2FormatMetricIncludesField,
			expected: `metric=cpu_value  42 0
metric=cpu_value  42 0
`,
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			s, err := NewSerializer(string(tc.format), DefaultSanitizeReplaceChar)
			require.NoError(t, err)

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
		format      format
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
			format:      Carbon2FormatFieldSeparate,
			expected:    fmt.Sprintf("metric=cpu:1 field=usage_idle  91.5 %d\n", now.Unix()),
			replaceChar: DefaultSanitizeReplaceChar,
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1", nil, fields, now)
			},
			format:      Carbon2FormatFieldSeparate,
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
			format:      Carbon2FormatFieldSeparate,
			expected:    fmt.Sprintf("metric=cpu:1:tmp:custom field=usage_idle  91.5 %d\n", now.Unix()),
			replaceChar: DefaultSanitizeReplaceChar,
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:      Carbon2FormatFieldSeparate,
			expected:    fmt.Sprintf("metric=cpu:1:tmp:custom:namespace field=usage_idle  91.5 %d\n", now.Unix()),
			replaceChar: DefaultSanitizeReplaceChar,
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:      Carbon2FormatMetricIncludesField,
			expected:    fmt.Sprintf("metric=cpu:1:tmp:custom:namespace_usage_idle  91.5 %d\n", now.Unix()),
			replaceChar: DefaultSanitizeReplaceChar,
		},
		{
			metricFunc: func() telegraf.Metric {
				fields := map[string]interface{}{
					"usage_idle": float64(91.5),
				}
				return metric.New("cpu=1=tmp$custom%namespace", nil, fields, now)
			},
			format:      Carbon2FormatMetricIncludesField,
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
			format:      Carbon2FormatMetricIncludesField,
			expectedErr: true,
			replaceChar: "___",
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.format), func(t *testing.T) {
			m := tc.metricFunc()

			s, err := NewSerializer(string(tc.format), tc.replaceChar)
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
