package carbon2

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func MustMetric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestSerializeMetricFloat(t *testing.T) {
	now := time.Now()
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(91.5),
	}
	m, err := metric.New("cpu", tags, fields, now)
	require.NoError(t, err)

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
			s, err := NewSerializer(string(tc.format))
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
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
	m, err := metric.New("cpu", tags, fields, now)
	require.NoError(t, err)

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
			s, err := NewSerializer(string(tc.format))
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
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
	m, err := metric.New("cpu metric", tags, fields, now)
	require.NoError(t, err)

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
			s, err := NewSerializer(string(tc.format))
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
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
	m, err := metric.New("cpu", tags, fields, now)
	require.NoError(t, err)

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
			s, err := NewSerializer(string(tc.format))
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
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
	m, err := metric.New("cpu", tags, fields, now)
	assert.NoError(t, err)

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
			s, err := NewSerializer(string(tc.format))
			require.NoError(t, err)

			buf, err := s.Serialize(m)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
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

		m, err := metric.New("cpu", tags, fields, tim)
		require.NoError(t, err)

		return m
	}

	now := time.Now()

	testcases := []struct {
		metric   telegraf.Metric
		format   string
		expected string
	}{
		{
			metric:   requireMetric(t, now, false),
			format:   string(Carbon2FormatFieldSeparate),
			expected: fmt.Sprintf("metric=cpu field=java_lang_GarbageCollector_Valid tag_name=tag_value  0 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(t, now, false),
			format:   string(Carbon2FormatMetricIncludesField),
			expected: fmt.Sprintf("metric=cpu_java_lang_GarbageCollector_Valid tag_name=tag_value  0 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(t, now, true),
			format:   string(Carbon2FormatFieldSeparate),
			expected: fmt.Sprintf("metric=cpu field=java_lang_GarbageCollector_Valid tag_name=tag_value  1 %d\n", now.Unix()),
		},
		{
			metric:   requireMetric(t, now, true),
			format:   string(Carbon2FormatMetricIncludesField),
			expected: fmt.Sprintf("metric=cpu_java_lang_GarbageCollector_Valid tag_name=tag_value  1 %d\n", now.Unix()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.format, func(t *testing.T) {
			s, err := NewSerializer(tc.format)
			require.NoError(t, err)

			buf, err := s.Serialize(tc.metric)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
		})
	}
}

func TestSerializeBatch(t *testing.T) {
	m := MustMetric(
		metric.New(
			"cpu",
			map[string]string{},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(0, 0),
		),
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
			s, err := NewSerializer(string(tc.format))
			require.NoError(t, err)

			buf, err := s.SerializeBatch(metrics)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, string(buf))
		})
	}
}
