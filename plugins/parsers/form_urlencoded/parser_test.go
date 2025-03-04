package form_urlencoded

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

const (
	validFormData              = "tag1=foo&tag2=bar&tag3=baz&field1=42&field2=69"
	encodedFormData            = "tag1=%24%24%24&field1=1e%2B3"
	notEscapedProperlyFormData = "invalid=%Y5"
	blankKeyFormData           = "=42&field2=69"
	emptyFormData              = ""
)

func TestParseValidFormData(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
	}

	metrics, err := parser.Parse([]byte(validFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "form_urlencoded_test", metrics[0].Name())
	require.Equal(t, map[string]string{}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(42),
		"field2": float64(69),
	}, metrics[0].Fields())
}

func TestParseLineValidFormData(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
	}

	metrics, err := parser.ParseLine(validFormData)
	require.NoError(t, err)
	require.Equal(t, "form_urlencoded_test", metrics.Name())
	require.Equal(t, map[string]string{}, metrics.Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(42),
		"field2": float64(69),
	}, metrics.Fields())
}

func TestParseValidFormDataWithTags(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
		TagKeys:    []string{"tag1", "tag2"},
	}

	metrics, err := parser.Parse([]byte(validFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "form_urlencoded_test", metrics[0].Name())
	require.Equal(t, map[string]string{
		"tag1": "foo",
		"tag2": "bar",
	}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(42),
		"field2": float64(69),
	}, metrics[0].Fields())
}

func TestParseValidFormDataDefaultTags(t *testing.T) {
	parser := Parser{
		MetricName:  "form_urlencoded_test",
		TagKeys:     []string{"tag1", "tag2"},
		DefaultTags: map[string]string{"tag4": "default"},
	}

	metrics, err := parser.Parse([]byte(validFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "form_urlencoded_test", metrics[0].Name())
	require.Equal(t, map[string]string{
		"tag1": "foo",
		"tag2": "bar",
		"tag4": "default",
	}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(42),
		"field2": float64(69),
	}, metrics[0].Fields())
}

func TestParseValidFormDataDefaultTagsOverride(t *testing.T) {
	parser := Parser{
		MetricName:  "form_urlencoded_test",
		TagKeys:     []string{"tag1", "tag2"},
		DefaultTags: map[string]string{"tag1": "default"},
	}

	metrics, err := parser.Parse([]byte(validFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "form_urlencoded_test", metrics[0].Name())
	require.Equal(t, map[string]string{
		"tag1": "default",
		"tag2": "bar",
	}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(42),
		"field2": float64(69),
	}, metrics[0].Fields())
}

func TestParseEncodedFormData(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
		TagKeys:    []string{"tag1"},
	}

	metrics, err := parser.Parse([]byte(encodedFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "form_urlencoded_test", metrics[0].Name())
	require.Equal(t, map[string]string{
		"tag1": "$$$",
	}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(1000),
	}, metrics[0].Fields())
}

func TestParseInvalidFormDataError(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
	}

	metrics, err := parser.Parse([]byte(notEscapedProperlyFormData))
	require.Error(t, err)
	require.Empty(t, metrics)
}

func TestParseInvalidFormDataEmptyKey(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
	}

	// Empty key for field
	metrics, err := parser.Parse([]byte(blankKeyFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, map[string]string{}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field2": float64(69),
	}, metrics[0].Fields())

	// Empty key for tag
	parser.TagKeys = []string{""}
	metrics, err = parser.Parse([]byte(blankKeyFormData))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, map[string]string{}, metrics[0].Tags())
	require.Equal(t, map[string]interface{}{
		"field2": float64(69),
	}, metrics[0].Fields())
}

func TestParseInvalidFormDataEmptyString(t *testing.T) {
	parser := Parser{
		MetricName: "form_urlencoded_test",
	}

	metrics, err := parser.Parse([]byte(emptyFormData))
	require.NoError(t, err)
	require.Empty(t, metrics)
}

const benchmarkData = `tags_host=myhost&tags_platform=python&tags_sdkver=3.11.5&value=5`

func TestBenchmarkData(t *testing.T) {
	plugin := &Parser{
		MetricName: "benchmark",
		TagKeys:    []string{"tags_host", "tags_platform", "tags_sdkver"},
	}

	expected := []telegraf.Metric{
		metric.New(
			"benchmark",
			map[string]string{
				"tags_host":     "myhost",
				"tags_platform": "python",
				"tags_sdkver":   "3.11.5",
			},
			map[string]interface{}{
				"value": 5.0,
			},
			time.Unix(0, 0),
		),
	}

	actual, err := plugin.Parse([]byte(benchmarkData))
	require.NoError(t, err)
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

func BenchmarkParsing(b *testing.B) {
	plugin := &Parser{
		MetricName: "benchmark",
		TagKeys:    []string{"source", "tags_platform", "tags_sdkver"},
	}

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse([]byte(benchmarkData))
	}
}
