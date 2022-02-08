package timestream

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/testutil"
)

const tsDbName = "testDb"

const testSingleTableName = "SingleTableName"
const testSingleTableDim = "namespace"

var time1 = time.Date(2009, time.November, 10, 22, 0, 0, 0, time.UTC)

const time1Epoch = "1257890400"
const timeUnit = types.TimeUnitSeconds

var time2 = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

const time2Epoch = "1257894000"

const metricName1 = "metricName1"
const metricName2 = "metricName2"

type mockTimestreamClient struct{}

func (m *mockTimestreamClient) CreateTable(context.Context, *timestreamwrite.CreateTableInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.CreateTableOutput, error) {
	return nil, nil
}
func (m *mockTimestreamClient) WriteRecords(context.Context, *timestreamwrite.WriteRecordsInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.WriteRecordsOutput, error) {
	return nil, nil
}
func (m *mockTimestreamClient) DescribeDatabase(context.Context, *timestreamwrite.DescribeDatabaseInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.DescribeDatabaseOutput, error) {
	return nil, fmt.Errorf("hello from DescribeDatabase")
}

func TestConnectValidatesConfigParameters(t *testing.T) {
	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return &mockTimestreamClient{}, nil
	}
	// checking base arguments
	noDatabaseName := Timestream{Log: testutil.Logger{}}
	require.Contains(t, noDatabaseName.Connect().Error(), "DatabaseName")

	noMappingMode := Timestream{
		DatabaseName: tsDbName,
		Log:          testutil.Logger{},
	}
	require.Contains(t, noMappingMode.Connect().Error(), "MappingMode")

	incorrectMappingMode := Timestream{
		DatabaseName: tsDbName,
		MappingMode:  "foo",
		Log:          testutil.Logger{},
	}
	require.Contains(t, incorrectMappingMode.Connect().Error(), "single-table")

	// multi-table arguments
	validMappingModeMultiTable := Timestream{
		DatabaseName: tsDbName,
		MappingMode:  MappingModeMultiTable,
		Log:          testutil.Logger{},
	}
	require.Nil(t, validMappingModeMultiTable.Connect())

	singleTableNameWithMultiTable := Timestream{
		DatabaseName:    tsDbName,
		MappingMode:     MappingModeMultiTable,
		SingleTableName: testSingleTableName,
		Log:             testutil.Logger{},
	}
	require.Contains(t, singleTableNameWithMultiTable.Connect().Error(), "SingleTableName")

	singleTableDimensionWithMultiTable := Timestream{
		DatabaseName: tsDbName,
		MappingMode:  MappingModeMultiTable,
		SingleTableDimensionNameForTelegrafMeasurementName: testSingleTableDim,
		Log: testutil.Logger{},
	}
	require.Contains(t, singleTableDimensionWithMultiTable.Connect().Error(),
		"SingleTableDimensionNameForTelegrafMeasurementName")

	// single-table arguments
	noTableNameMappingModeSingleTable := Timestream{
		DatabaseName: tsDbName,
		MappingMode:  MappingModeSingleTable,
		Log:          testutil.Logger{},
	}
	require.Contains(t, noTableNameMappingModeSingleTable.Connect().Error(), "SingleTableName")

	noDimensionNameMappingModeSingleTable := Timestream{
		DatabaseName:    tsDbName,
		MappingMode:     MappingModeSingleTable,
		SingleTableName: testSingleTableName,
		Log:             testutil.Logger{},
	}
	require.Contains(t, noDimensionNameMappingModeSingleTable.Connect().Error(),
		"SingleTableDimensionNameForTelegrafMeasurementName")

	validConfigurationMappingModeSingleTable := Timestream{
		DatabaseName:    tsDbName,
		MappingMode:     MappingModeSingleTable,
		SingleTableName: testSingleTableName,
		SingleTableDimensionNameForTelegrafMeasurementName: testSingleTableDim,
		Log: testutil.Logger{},
	}
	require.Nil(t, validConfigurationMappingModeSingleTable.Connect())

	// create table arguments
	createTableNoMagneticRetention := Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            MappingModeMultiTable,
		CreateTableIfNotExists: true,
		Log:                    testutil.Logger{},
	}
	require.Contains(t, createTableNoMagneticRetention.Connect().Error(),
		"CreateTableMagneticStoreRetentionPeriodInDays")

	createTableNoMemoryRetention := Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            MappingModeMultiTable,
		CreateTableIfNotExists: true,
		CreateTableMagneticStoreRetentionPeriodInDays: 3,
		Log: testutil.Logger{},
	}
	require.Contains(t, createTableNoMemoryRetention.Connect().Error(),
		"CreateTableMemoryStoreRetentionPeriodInHours")

	createTableValid := Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            MappingModeMultiTable,
		CreateTableIfNotExists: true,
		CreateTableMagneticStoreRetentionPeriodInDays: 3,
		CreateTableMemoryStoreRetentionPeriodInHours:  3,
		Log: testutil.Logger{},
	}
	require.Nil(t, createTableValid.Connect())

	// describe table on start arguments
	describeTableInvoked := Timestream{
		DatabaseName:            tsDbName,
		MappingMode:             MappingModeMultiTable,
		DescribeDatabaseOnStart: true,
		Log:                     testutil.Logger{},
	}
	require.Contains(t, describeTableInvoked.Connect().Error(), "hello from DescribeDatabase")
}

type mockTimestreamErrorClient struct {
	ErrorToReturnOnWriteRecords error
}

func (m *mockTimestreamErrorClient) CreateTable(context.Context, *timestreamwrite.CreateTableInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.CreateTableOutput, error) {
	return nil, nil
}
func (m *mockTimestreamErrorClient) WriteRecords(context.Context, *timestreamwrite.WriteRecordsInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.WriteRecordsOutput, error) {
	return nil, m.ErrorToReturnOnWriteRecords
}
func (m *mockTimestreamErrorClient) DescribeDatabase(context.Context, *timestreamwrite.DescribeDatabaseInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.DescribeDatabaseOutput, error) {
	return nil, nil
}

func TestThrottlingErrorIsReturnedToTelegraf(t *testing.T) {
	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return &mockTimestreamErrorClient{
			ErrorToReturnOnWriteRecords: &types.ThrottlingException{Message: aws.String("Throttling Test")},
		}, nil
	}

	plugin := Timestream{
		MappingMode:  MappingModeMultiTable,
		DatabaseName: tsDbName,
		Log:          testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())
	input := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": float64(1)},
		time1,
	)

	err := plugin.Write([]telegraf.Metric{input})

	require.NotNil(t, err, "Expected an error to be returned to Telegraf, "+
		"so that the write will be retried by Telegraf later.")
}

func TestRejectedRecordsErrorResultsInMetricsBeingSkipped(t *testing.T) {
	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return &mockTimestreamErrorClient{
			ErrorToReturnOnWriteRecords: &types.RejectedRecordsException{Message: aws.String("RejectedRecords Test")},
		}, nil
	}

	plugin := Timestream{
		MappingMode:  MappingModeMultiTable,
		DatabaseName: tsDbName,
		Log:          testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())
	input := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{"value": float64(1)},
		time1,
	)

	err := plugin.Write([]telegraf.Metric{input})

	require.Nil(t, err, "Expected to silently swallow the RejectedRecordsException, "+
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

	records := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag2": "value2", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value": "10"},
		},

		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{testSingleTableDim: metricName1},
			measureValues: map[string]string{"value": "20"},
		},
	})

	expectedResultSingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          records,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{input1, input2, input3},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	recordsMulti := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag2": "value2"},
			measureValues: map[string]string{"value": "10"},
		},
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{},
			measureValues: map[string]string{"value": "20"},
		},
	})

	expectedResultMultiTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(metricName1),
		Records:          recordsMulti,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeMultiTable,
		[]telegraf.Metric{input1, input2, input3},
		[]*timestreamwrite.WriteRecordsInput{expectedResultMultiTable})
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

	expectedResult1SingleTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: resultFields,
	})
	expectedResult2SingleTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported" + strconv.Itoa(maxRecordsInWriteRecordsCall+1): "10"},
	})
	comparisonTest(t, MappingModeSingleTable,
		inputs,
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})

	expectedResult1MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: resultFields,
	})
	expectedResult2MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported" + strconv.Itoa(maxRecordsInWriteRecordsCall+1): "10"},
	})
	comparisonTest(t, MappingModeMultiTable,
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
		metricName2,
		map[string]string{"tag2": "value2"},
		map[string]interface{}{
			"value_supported3": float64(30),
		},
		time1,
	)

	recordsSingle := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
		},
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag2": "value2", testSingleTableDim: metricName2},
			measureValues: map[string]string{"value_supported3": "30"},
		},
	})

	expectedResultSingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          recordsSingle,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResult1MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})

	expectedResult2MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName2,
		dimensions:    map[string]string{"tag2": "value2"},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, MappingModeMultiTable,
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
		metricName2,
		map[string]string{"tag1": "value2"},
		map[string]interface{}{
			"value_supported1": float64(20),
		},
		time1,
	)

	recordsSingle := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value_supported1": "10"},
		},
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value2", testSingleTableDim: metricName2},
			measureValues: map[string]string{"value_supported1": "20"},
		},
	})

	expectedResultSingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          recordsSingle,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResult1MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10"},
	})
	expectedResult2MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName2,
		dimensions:    map[string]string{"tag1": "value2"},
		measureValues: map[string]string{"value_supported1": "20"},
	})

	comparisonTest(t, MappingModeMultiTable,
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

	recordsSingle := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
		},
		{
			t:             time2Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value_supported3": "30"},
		},
	})

	expectedResultSingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          recordsSingle,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	recordsMultiTable := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag1": "value1"},
			measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
		},
		{
			t:             time2Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag1": "value1"},
			measureValues: map[string]string{"value_supported3": "30"},
		},
	})

	expectedResultMultiTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(metricName1),
		Records:          recordsMultiTable,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeMultiTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultMultiTable})
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

	expectedResultSingleTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20", "value_supported3": "30"},
	})

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResultMultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20", "value_supported3": "30"},
	})

	comparisonTest(t, MappingModeMultiTable,
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

	recordsSingle := buildRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
		},
		{
			t:             time1Epoch,
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName2},
			measureValues: map[string]string{"value_supported3": "30"},
		},
	})

	expectedResultSingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          recordsSingle,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{input1, input2},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResult1MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10", "value_supported2": "20"},
	})
	expectedResult2MultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName2,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported3": "30"},
	})

	comparisonTest(t, MappingModeMultiTable,
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
	expectedResultSingleTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported1": "10"},
	})

	comparisonTest(t, MappingModeSingleTable,
		[]telegraf.Metric{metricWithUnsupportedField},
		[]*timestreamwrite.WriteRecordsInput{expectedResultSingleTable})

	expectedResultMultiTable := buildExpectedInput(SimpleInput{
		t:             time1Epoch,
		tableName:     metricName1,
		dimensions:    map[string]string{"tag1": "value1"},
		measureValues: map[string]string{"value_supported1": "10"},
	})

	comparisonTest(t, MappingModeMultiTable,
		[]telegraf.Metric{metricWithUnsupportedField},
		[]*timestreamwrite.WriteRecordsInput{expectedResultMultiTable})
}

func comparisonTest(t *testing.T,
	mappingMode string,
	telegrafMetrics []telegraf.Metric,
	timestreamRecords []*timestreamwrite.WriteRecordsInput,
) {
	var plugin Timestream
	switch mappingMode {
	case MappingModeSingleTable:
		plugin = Timestream{
			MappingMode:  mappingMode,
			DatabaseName: tsDbName,

			SingleTableName: testSingleTableName,
			SingleTableDimensionNameForTelegrafMeasurementName: testSingleTableDim,
			Log: testutil.Logger{},
		}
	case MappingModeMultiTable:
		plugin = Timestream{
			MappingMode:  mappingMode,
			DatabaseName: tsDbName,
			Log:          testutil.Logger{},
		}
	}

	comparison(t, plugin, mappingMode, telegrafMetrics, timestreamRecords)
}

func comparison(t *testing.T,
	plugin Timestream,
	mappingMode string,
	telegrafMetrics []telegraf.Metric,
	timestreamRecords []*timestreamwrite.WriteRecordsInput) {
	result := plugin.TransformMetrics(telegrafMetrics)

	require.Equal(t, len(timestreamRecords), len(result), "The number of transformed records was expected to be different")
	for _, tsRecord := range timestreamRecords {
		require.True(t, arrayContains(result, tsRecord), "Expected that the list of requests to Timestream: \n%s\n\n "+
			"will contain request: \n%s\n\nUsed MappingMode: %s", result, tsRecord, mappingMode)
	}
}

func arrayContains(
	array []*timestreamwrite.WriteRecordsInput,
	element *timestreamwrite.WriteRecordsInput,
) bool {
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

func buildExpectedInput(i SimpleInput) *timestreamwrite.WriteRecordsInput {
	var tsDimensions []types.Dimension
	for k, v := range i.dimensions {
		tsDimensions = append(tsDimensions, types.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	var tsRecords []types.Record
	for k, v := range i.measureValues {
		tsRecords = append(tsRecords, types.Record{
			MeasureName:      aws.String(k),
			MeasureValue:     aws.String(v),
			MeasureValueType: types.MeasureValueTypeDouble,
			Dimensions:       tsDimensions,
			Time:             aws.String(i.t),
			TimeUnit:         timeUnit,
		})
	}

	result := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(i.tableName),
		Records:          tsRecords,
		CommonAttributes: &types.Record{},
	}

	return result
}

func buildRecords(inputs []SimpleInput) []types.Record {
	var tsRecords []types.Record

	for _, inp := range inputs {
		tsRecords = append(tsRecords, buildRecord(inp)...)
	}

	return tsRecords
}

func buildRecord(input SimpleInput) []types.Record {
	var tsRecords []types.Record

	var tsDimensions []types.Dimension

	for k, v := range input.dimensions {
		tsDimensions = append(tsDimensions, types.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	for k, v := range input.measureValues {
		tsRecords = append(tsRecords, types.Record{
			MeasureName:      aws.String(k),
			MeasureValue:     aws.String(v),
			MeasureValueType: types.MeasureValueTypeDouble,
			Dimensions:       tsDimensions,
			Time:             aws.String(input.t),
			TimeUnit:         timeUnit,
		})
	}

	return tsRecords
}
