package timestream_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	ts "github.com/influxdata/telegraf/plugins/outputs/timestream"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

const tsDbName = "testDb"

const testSingleTableName = "SingleTableName"
const testSingleTableDim = "namespace"

var time1 = time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)

const time1Epoch = "1257890400"

var time2 = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

const time2Epoch = "1257894000"

const timeUnit = "SECONDS"

const metricName1 = "metricName1"
const metricName2 = "metricName2"

type mockTimestreamClient struct {
}

func (m *mockTimestreamClient) CreateTable(*timestreamwrite.CreateTableInput) (*timestreamwrite.CreateTableOutput, error) {
	return nil, nil
}
func (m *mockTimestreamClient) WriteRecords(*timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error) {
	return nil, nil
}
func (m *mockTimestreamClient) DescribeDatabase(*timestreamwrite.DescribeDatabaseInput) (*timestreamwrite.DescribeDatabaseOutput, error) {
	return nil, fmt.Errorf("hello from DescribeDatabase")
}

func TestConnectValidatesConfigParameters(t *testing.T) {
	assertions := assert.New(t)
	ts.WriteFactory = func(credentialConfig *internalaws.CredentialConfig) ts.WriteClient {
		return &mockTimestreamClient{}
	}

	// checking base arguments
	noDatabaseName := ts.Timestream{Log: testutil.Logger{}}
	assertions.Contains(noDatabaseName.Connect().Error(), "DatabaseName")

	noMappingMode := ts.Timestream{
		DatabaseName: tsDbName,
		Log:          testutil.Logger{},
	}
	assertions.Contains(noMappingMode.Connect().Error(), "MappingMode")

	incorrectMappingMode := ts.Timestream{
		DatabaseName: tsDbName,
		MappingMode:  "foo",
		Log:          testutil.Logger{},
	}
	assertions.Contains(incorrectMappingMode.Connect().Error(), "single-table")

	// multi-table arguments
	validMappingModeMultiTable := ts.Timestream{
		DatabaseName: tsDbName,
		MappingMode:  ts.MappingModeMultiTable,
		Log:          testutil.Logger{},
	}
	assertions.Nil(validMappingModeMultiTable.Connect())

	singleTableNameWithMultiTable := ts.Timestream{
		DatabaseName:    tsDbName,
		MappingMode:     ts.MappingModeMultiTable,
		SingleTableName: testSingleTableName,
		Log:             testutil.Logger{},
	}
	assertions.Contains(singleTableNameWithMultiTable.Connect().Error(), "SingleTableName")

	singleTableDimensionWithMultiTable := ts.Timestream{
		DatabaseName: tsDbName,
		MappingMode:  ts.MappingModeMultiTable,
		SingleTableDimensionNameForTelegrafMeasurementName: testSingleTableDim,
		Log: testutil.Logger{},
	}
	assertions.Contains(singleTableDimensionWithMultiTable.Connect().Error(),
		"SingleTableDimensionNameForTelegrafMeasurementName")

	// single-table arguments
	noTableNameMappingModeSingleTable := ts.Timestream{
		DatabaseName: tsDbName,
		MappingMode:  ts.MappingModeSingleTable,
		Log:          testutil.Logger{},
	}
	assertions.Contains(noTableNameMappingModeSingleTable.Connect().Error(), "SingleTableName")

	noDimensionNameMappingModeSingleTable := ts.Timestream{
		DatabaseName:    tsDbName,
		MappingMode:     ts.MappingModeSingleTable,
		SingleTableName: testSingleTableName,
		Log:             testutil.Logger{},
	}
	assertions.Contains(noDimensionNameMappingModeSingleTable.Connect().Error(),
		"SingleTableDimensionNameForTelegrafMeasurementName")

	validConfigurationMappingModeSingleTable := ts.Timestream{
		DatabaseName:    tsDbName,
		MappingMode:     ts.MappingModeSingleTable,
		SingleTableName: testSingleTableName,
		SingleTableDimensionNameForTelegrafMeasurementName: testSingleTableDim,
		Log: testutil.Logger{},
	}
	assertions.Nil(validConfigurationMappingModeSingleTable.Connect())

	// create table arguments
	createTableNoMagneticRetention := ts.Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            ts.MappingModeMultiTable,
		CreateTableIfNotExists: true,
		Log:                    testutil.Logger{},
	}
	assertions.Contains(createTableNoMagneticRetention.Connect().Error(),
		"CreateTableMagneticStoreRetentionPeriodInDays")

	createTableNoMemoryRetention := ts.Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            ts.MappingModeMultiTable,
		CreateTableIfNotExists: true,
		CreateTableMagneticStoreRetentionPeriodInDays: 3,
		Log: testutil.Logger{},
	}
	assertions.Contains(createTableNoMemoryRetention.Connect().Error(),
		"CreateTableMemoryStoreRetentionPeriodInHours")

	createTableValid := ts.Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            ts.MappingModeMultiTable,
		CreateTableIfNotExists: true,
		CreateTableMagneticStoreRetentionPeriodInDays: 3,
		CreateTableMemoryStoreRetentionPeriodInHours:  3,
		Log: testutil.Logger{},
	}
	assertions.Nil(createTableValid.Connect())

	// describe table on start arguments
	describeTableInvoked := ts.Timestream{
		DatabaseName:            tsDbName,
		MappingMode:             ts.MappingModeMultiTable,
		DescribeDatabaseOnStart: true,
		Log:                     testutil.Logger{},
	}
	assertions.Contains(describeTableInvoked.Connect().Error(), "hello from DescribeDatabase")
}

type mockTimestreamErrorClient struct {
	ErrorToReturnOnWriteRecords error
}

func (m *mockTimestreamErrorClient) CreateTable(*timestreamwrite.CreateTableInput) (*timestreamwrite.CreateTableOutput, error) {
	return nil, nil
}
func (m *mockTimestreamErrorClient) WriteRecords(*timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error) {
	return nil, m.ErrorToReturnOnWriteRecords
}
func (m *mockTimestreamErrorClient) DescribeDatabase(*timestreamwrite.DescribeDatabaseInput) (*timestreamwrite.DescribeDatabaseOutput, error) {
	return nil, nil
}

func TestThrottlingErrorIsReturnedToTelegraf(t *testing.T) {
	assertions := assert.New(t)

	ts.WriteFactory = func(credentialConfig *internalaws.CredentialConfig) ts.WriteClient {
		return &mockTimestreamErrorClient{
			awserr.New(timestreamwrite.ErrCodeThrottlingException,
				"Throttling Test", nil),
		}
	}
	plugin := ts.Timestream{
		MappingMode:  ts.MappingModeMultiTable,
		DatabaseName: tsDbName,
		Log:          testutil.Logger{},
	}
	plugin.Connect()
	input := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": float64(1)},
		time1,
	)

	err := plugin.Write([]telegraf.Metric{input})

	assertions.NotNil(err, "Expected an error to be returned to Telegraf, "+
		"so that the write will be retried by Telegraf later.")
}

func TestRejectedRecordsErrorResultsInMetricsBeingSkipped(t *testing.T) {
	assertions := assert.New(t)

	ts.WriteFactory = func(credentialConfig *internalaws.CredentialConfig) ts.WriteClient {
		return &mockTimestreamErrorClient{
			awserr.New(timestreamwrite.ErrCodeRejectedRecordsException,
				"RejectedRecords Test", nil),
		}
	}
	plugin := ts.Timestream{
		MappingMode:  ts.MappingModeMultiTable,
		DatabaseName: tsDbName,
		Log:          testutil.Logger{},
	}
	plugin.Connect()
	input := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": float64(1)},
		time1,
	)

	err := plugin.Write([]telegraf.Metric{input})

	assertions.Nil(err, "Expected to silently swallow the RejectedRecordsException, "+
		"as retrying this error doesn't make sense.")
}

func TestTransformMetricsSkipEmptyMetric(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{}, //no fields here
		time1,
	)
	input2 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag2": "value2"},
		map[string]interface{}{
			"value": float64(10),
		},
		time1,
	)
	input3 := testutil.MustMetric(
		metricName1,
		map[string]string{}, //record with no dimensions should appear in the results
		map[string]interface{}{
			"value": float64(20),
		},
		time1,
	)

	expectedResult1SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag2": "value2", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value": "10"},
	})
	expectedResult2SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{testSingleTableDim: metricName1},
		measureValues: map[string]string{"value": "20"},
	})
	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{input1, input2, input3},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag2": "value2"},
		measureValues: map[string]string{"value": "10"},
	})
	expectedResult2MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{},
		measureValues: map[string]string{"value": "20"},
	})
	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{input1, input2, input3},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1MultiTable, expectedResult2MultiTable})
}

func TestTransformMetricsRequestsAboveLimitAreSplit(t *testing.T) {
	const maxRecordsInWriteRecordsCall = 100

	var inputs []telegraf.Metric
	for i := 1; i <= maxRecordsInWriteRecordsCall+1; i++ {
		fieldName := "value_supported" + strconv.Itoa(i)
		inputs = append(inputs, testutil.MustMetric(
			metricName1,
			map[string]string{"tag1": "value1"},
			map[string]interface{}{
				fieldName: float64(10),
			},
			time1,
		))
	}

	resultFields := make(map[string]string)
	for i := 1; i <= maxRecordsInWriteRecordsCall; i++ {
		fieldName := "value_supported" + strconv.Itoa(i)
		resultFields[fieldName] = "10"
	}

	expectedResult1SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: resultFields,
	})
	expectedResult2SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported" + strconv.Itoa(maxRecordsInWriteRecordsCall+1): "10"},
	})
	comparisonTest(t, ts.MappingModeSingleTable,
		inputs,
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: resultFields,
	})
	expectedResult2MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported" + strconv.Itoa(maxRecordsInWriteRecordsCall+1): "10"},
	})
	comparisonTest(t, ts.MappingModeMultiTable,
		inputs,
		[]*timestreamwrite.WriteRecordsInput{expectedResult1MultiTable, expectedResult2MultiTable})
}

func TestTransformMetricsDifferentDimensionsSameTimestampsAreWrittenSeparate(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported1": float64(10), "value_supported2": float64(20),
		},
		time1,
	)
	input2 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag2": "value2"},
		map[string]interface{}{
			"value_supported3": float64(30),
		},
		time1,
	)

	expectedResult1SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag2": "value2", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag2": "value2"},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1MultiTable, expectedResult2MultiTable})
}

func TestTransformMetricsSameDimensionsDifferentDimensionValuesAreWrittenSeparate(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported1": float64(10),
		},
		time1,
	)
	input2 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value2"},
		map[string]interface{}{
			"value_supported1": float64(20),
		},
		time1,
	)

	expectedResult1SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10"},
	})
	expectedResult2SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value2", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "20"},
	})

	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10"},
	})
	expectedResult2MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value2"},
		measureValues: map[string]string{"value_supported1": "20"},
	})

	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1MultiTable, expectedResult2MultiTable})
}

func TestTransformMetricsSameDimensionsDifferentTimestampsAreWrittenSeparate(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported1": float64(10), "value_supported2": float64(20),
		},
		time1,
	)
	input2 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported3": float64(30),
		},
		time2,
	)

	expectedResult1SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2SingleTable := buildExpectedRecords(SimpleInput{
		t:             time2Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2MultiTable := buildExpectedRecords(SimpleInput{
		t:             time2Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1MultiTable, expectedResult2MultiTable})
}

func TestTransformMetricsSameDimensionsSameTimestampsAreWrittenTogether(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported1": float64(10), "value_supported2": float64(20),
		},
		time1,
	)
	input2 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported3": float64(30),
		},
		time1,
	)

	expectedResultSingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20", "value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResultMultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20", "value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultMultiTable})
}

func TestTransformMetricsDifferentMetricsAreWrittenToDifferentTablesInMultiTableMapping(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported1": float64(10), "value_supported2": float64(20),
		},
		time1,
	)
	input2 := testutil.MustMetric(
		metricName2,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported3": float64(30),
		},
		time1,
	)

	expectedResult1SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2SingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName2},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2MultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName2,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResult1MultiTable, expectedResult2MultiTable})
}

func TestTransformMetricsUnsupportedFieldsAreSkipped(t *testing.T) {
	metricWithUnsupportedField := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"value_supported1": float64(10), "value_unsupported": time.Now(),
		},
		time1,
	)
	expectedResultSingleTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10"},
	})

	comparisonTest(t, ts.MappingModeSingleTable,
		[]telegraf.Metric{metricWithUnsupportedField},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResultMultiTable := buildExpectedRecords(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10"},
	})

	comparisonTest(t, ts.MappingModeMultiTable,
		[]telegraf.Metric{metricWithUnsupportedField},
		[]*timestreamwrite.WriteRecordsInput{expectedResultMultiTable})
}

func comparisonTest(t *testing.T,
	mappingMode string,
	telegrafMetrics []telegraf.Metric,
	timestreamRecords []*timestreamwrite.WriteRecordsInput) {

	var plugin ts.Timestream
	switch mappingMode {
	case ts.MappingModeSingleTable:
		plugin = ts.Timestream{
			MappingMode:  mappingMode,
			DatabaseName: tsDbName,

			SingleTableName: testSingleTableName,
			SingleTableDimensionNameForTelegrafMeasurementName: testSingleTableDim,
			Log: testutil.Logger{},
		}
	case ts.MappingModeMultiTable:
		plugin = ts.Timestream{
			MappingMode:  mappingMode,
			DatabaseName: tsDbName,
			Log:          testutil.Logger{},
		}
	}
	assertions := assert.New(t)

	result := plugin.TransformMetrics(telegrafMetrics)

	assertions.Equal(len(timestreamRecords), len(result), "The number of transformed records was expected to be different")
	for _, tsRecord := range timestreamRecords {
		assertions.True(arrayContains(result, tsRecord), "Expected that the list of requests to Timestream: \n%s\n\n "+
			"will contain request: \n%s\n\nUsed MappingMode: %s", result, tsRecord, mappingMode)
	}
}

func arrayContains(
	array []*timestreamwrite.WriteRecordsInput,
	element *timestreamwrite.WriteRecordsInput) bool {

	sortWriteInputForComparison(*element)

	for _, a := range array {
		sortWriteInputForComparison(*a)

		if reflect.DeepEqual(a, element) {
			return true
		}
	}
	return false
}

func sortWriteInputForComparison(element timestreamwrite.WriteRecordsInput) {
	// sort the records by MeasureName, as they are kept in an array, but the order of records doesn't matter
	sort.Slice(element.Records, func(i, j int) bool {
		return strings.Compare(*element.Records[i].MeasureName, *element.Records[j].MeasureName) < 0
	})
	// sort the dimensions in CommonAttributes
	if element.CommonAttributes != nil {
		sort.Slice(element.CommonAttributes.Dimensions, func(i, j int) bool {
			return strings.Compare(*element.CommonAttributes.Dimensions[i].Name,
				*element.CommonAttributes.Dimensions[j].Name) < 0
		})
	}
	// sort the dimensions in Records
	for _, r := range element.Records {
		sort.Slice(r.Dimensions, func(i, j int) bool {
			return strings.Compare(*r.Dimensions[i].Name, *r.Dimensions[j].Name) < 0
		})
	}
}

type SimpleInput struct {
	t             string
	tableName     string
	dimensions    map[string]string
	measureValues map[string]string
}

func buildExpectedRecords(i SimpleInput) *timestreamwrite.WriteRecordsInput {
	var tsDimensions []*timestreamwrite.Dimension
	for k, v := range i.dimensions {
		tsDimensions = append(tsDimensions, &timestreamwrite.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	var tsRecords []*timestreamwrite.Record
	for k, v := range i.measureValues {
		tsRecords = append(tsRecords, &timestreamwrite.Record{
			MeasureName:      aws.String(k),
			MeasureValue:     aws.String(v),
			MeasureValueType: aws.String("DOUBLE"),
		})
	}

	result := &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(tsDbName),
		TableName:    aws.String(i.tableName),
		Records:      tsRecords,
		CommonAttributes: &timestreamwrite.Record{
			Dimensions: tsDimensions,
			Time:       aws.String(i.t),
			TimeUnit:   aws.String(timeUnit),
		},
	}

	return result
}
