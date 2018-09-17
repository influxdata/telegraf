package cloudwatch

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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

	zero := 0.0
	validMetrics := []telegraf.Metric{
		testutil.TestMetric(1),
		testutil.TestMetric(int32(1)),
		testutil.TestMetric(int64(1)),
		testutil.TestMetric(float64(1)),
		testutil.TestMetric(float64(0)),
		testutil.TestMetric(math.Copysign(zero, -1)), // the CW documentation does not call out -0 as rejected
		testutil.TestMetric(float64(8.515920e-109)),
		testutil.TestMetric(float64(1.174271e+108)), // largest should be 1.174271e+108
		testutil.TestMetric(true),
	}
	invalidMetrics := []telegraf.Metric{
		testutil.TestMetric("Foo"),
		testutil.TestMetric(math.Log(-1.0)),
		testutil.TestMetric(float64(8.515919e-109)), // smallest should be 8.515920e-109
		testutil.TestMetric(float64(1.174272e+108)), // largest should be 1.174271e+108
	}
	for _, point := range validMetrics {
		datums := BuildMetricDatum(false, point)
		assert.Equal(1, len(datums), fmt.Sprintf("Valid point should create a Datum {value: %v}", point))
	}
	for _, point := range invalidMetrics {
		datums := BuildMetricDatum(false, point)
		assert.Equal(0, len(datums), fmt.Sprintf("Valid point should not create a Datum {value: %v}", point))
	}

	statisticMetric, _ := metric.New(
		"test1",
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value_max": float64(10), "value_min": float64(0), "value_sum": float64(100), "value_count": float64(20)},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	datums := BuildMetricDatum(true, statisticMetric)
	assert.Equal(1, len(datums), fmt.Sprintf("Valid point should create a Datum {value: %v}", statisticMetric))

	multiFieldsMetric, _ := metric.New(
		"test1",
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"valueA": float64(10), "valueB": float64(0), "valueC": float64(100), "valueD": float64(20)},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	datums = BuildMetricDatum(true, multiFieldsMetric)
	assert.Equal(4, len(datums), fmt.Sprintf("Each field should create a Datum {value: %v}", multiFieldsMetric))

	multiStatisticMetric, _ := metric.New(
		"test1",
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"valueA_max": float64(10), "valueA_min": float64(0), "valueA_sum": float64(100), "valueA_count": float64(20),
			"valueB_max": float64(10), "valueB_min": float64(0), "valueB_sum": float64(100), "valueB_count": float64(20),
			"valueC_max": float64(10), "valueC_min": float64(0), "valueC_sum": float64(100),
			"valueD": float64(10), "valueE": float64(0),
		},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	datums = BuildMetricDatum(true, multiStatisticMetric)
	assert.Equal(7, len(datums), fmt.Sprintf("Valid point should create a Datum {value: %v}", multiStatisticMetric))
}

func TestPartitionDatums(t *testing.T) {

	assert := assert.New(t)

	testDatum := cloudwatch.MetricDatum{
		MetricName: aws.String("Foo"),
		Value:      aws.Float64(1),
	}

	zeroDatum := []*cloudwatch.MetricDatum{}
	oneDatum := []*cloudwatch.MetricDatum{&testDatum}
	twoDatum := []*cloudwatch.MetricDatum{&testDatum, &testDatum}
	threeDatum := []*cloudwatch.MetricDatum{&testDatum, &testDatum, &testDatum}

	assert.Equal([][]*cloudwatch.MetricDatum{}, PartitionDatums(2, zeroDatum))
	assert.Equal([][]*cloudwatch.MetricDatum{oneDatum}, PartitionDatums(2, oneDatum))
	assert.Equal([][]*cloudwatch.MetricDatum{oneDatum}, PartitionDatums(2, oneDatum))
	assert.Equal([][]*cloudwatch.MetricDatum{twoDatum}, PartitionDatums(2, twoDatum))
	assert.Equal([][]*cloudwatch.MetricDatum{twoDatum, oneDatum}, PartitionDatums(2, threeDatum))
}
