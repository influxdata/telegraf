package cloudwatch

import (
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

// Test that each tag becomes one dimension
func TestBuildDimensions(t *testing.T) {
	const MaxDimensions = 10

	assert := assert.New(t)

	testPoint := testutil.TestMetric(1)
	dimensions := BuildDimensions(testPoint.Tags())

	tagKeys := make([]string, len(testPoint.Tags()))
	i := 0
	for k, _ := range testPoint.Tags() {
		tagKeys[i] = k
		i += 1
	}

	sort.Strings(tagKeys)

	if len(testPoint.Tags()) >= MaxDimensions {
		assert.Equal(MaxDimensions, len(dimensions), "Number of dimensions should be less than MaxDimensions")
	} else {
		assert.Equal(len(testPoint.Tags()), len(dimensions), "Number of dimensions should be equal to number of tags")
	}

	for i, key := range tagKeys {
		if i >= 10 {
			break
		}
		assert.Equal(key, *dimensions[i].Name, "Key should be equal")
		assert.Equal(testPoint.Tags()[key], *dimensions[i].Value, "Value should be equal")
	}
}

// Test that metrics with valid values have a MetricDatum created where as non valid do not.
// Skips "time.Time" type as something is converting the value to string.
func TestBuildMetricDatums(t *testing.T) {
	assert := assert.New(t)

	validMetrics := []telegraf.Metric{
		testutil.TestMetric(1),
		testutil.TestMetric(int32(1)),
		testutil.TestMetric(int64(1)),
		testutil.TestMetric(float64(1)),
		testutil.TestMetric(true),
	}

	for _, point := range validMetrics {
		datums := BuildMetricDatum(point)
		assert.Equal(1, len(datums), "Valid type should create a Datum")
	}

	nonValidPoint := testutil.TestMetric("Foo")

	assert.Equal(0, len(BuildMetricDatum(nonValidPoint)), "Invalid type should not create a Datum")
}

func TestPartitionDatums(t *testing.T) {

	assert := assert.New(t)

	testDatum := cloudwatch.MetricDatum{
		MetricName: aws.String("Foo"),
		Value:      aws.Float64(1),
	}

	oneDatum := []*cloudwatch.MetricDatum{&testDatum}
	twoDatum := []*cloudwatch.MetricDatum{&testDatum, &testDatum}
	threeDatum := []*cloudwatch.MetricDatum{&testDatum, &testDatum, &testDatum}

	assert.Equal([][]*cloudwatch.MetricDatum{oneDatum}, PartitionDatums(2, oneDatum))
	assert.Equal([][]*cloudwatch.MetricDatum{twoDatum}, PartitionDatums(2, twoDatum))
	assert.Equal([][]*cloudwatch.MetricDatum{twoDatum, oneDatum}, PartitionDatums(2, threeDatum))
}
