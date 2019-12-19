package form_urlencoded

import (
	"testing"

	"github.com/stretchr/testify/require"
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

	metric, err := parser.ParseLine(validFormData)
	require.NoError(t, err)
	require.Equal(t, "form_urlencoded_test", metric.Name())
	require.Equal(t, map[string]string{}, metric.Tags())
	require.Equal(t, map[string]interface{}{
		"field1": float64(42),
		"field2": float64(69),
	}, metric.Fields())
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
	require.Len(t, metrics, 0)
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
	require.Len(t, metrics, 0)
}
