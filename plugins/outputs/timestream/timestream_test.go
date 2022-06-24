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

type mockTimestreamClient struct {
	WriteRecordsRequestCount int
}

func (m *mockTimestreamClient) CreateTable(context.Context, *timestreamwrite.CreateTableInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.CreateTableOutput, error) {
	return nil, nil
}
func (m *mockTimestreamClient) WriteRecords(context.Context, *timestreamwrite.WriteRecordsInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.WriteRecordsOutput, error) {
	m.WriteRecordsRequestCount++
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

	//multi-measure config validation multi table mode
	validConfigMultiMeasureMultiTableMode := Timestream{
		DatabaseName:                      tsDbName,
		MappingMode:                       MappingModeMultiTable,
		UseMultiMeasureRecords:            true,
		MeasureNameForMultiMeasureRecords: "multi-measure-name",
		Log:                               testutil.Logger{},
	}
	require.Nil(t, validConfigMultiMeasureMultiTableMode.Connect())

	invalidConfigMultiMeasureMultiTableMode := Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            MappingModeMultiTable,
		UseMultiMeasureRecords: true,
		// without MeasureNameForMultiMeasureRecords set we expect validation failure
		Log: testutil.Logger{},
	}
	require.Contains(t, invalidConfigMultiMeasureMultiTableMode.Connect().Error(), "MeasureNameForMultiMeasureRecords")

	// multi-measure config validation single table mode
	validConfigMultiMeasureSingleTableMode := Timestream{
		DatabaseName:           tsDbName,
		MappingMode:            MappingModeSingleTable,
		SingleTableName:        testSingleTableName,
		UseMultiMeasureRecords: true, // MeasureNameForMultiMeasureRecords is not needed as
		// measurement name (from telegraf metric) is used as multi-measure name in TS
		Log: testutil.Logger{},
	}
	require.Nil(t, validConfigMultiMeasureSingleTableMode.Connect())

	invalidConfigMultiMeasureSingleTableMode := Timestream{
		DatabaseName:                      tsDbName,
		MappingMode:                       MappingModeSingleTable,
		SingleTableName:                   testSingleTableName,
		UseMultiMeasureRecords:            true,
		MeasureNameForMultiMeasureRecords: "multi-measure-name",
		// value of MeasureNameForMultiMeasureRecords will be ignored and
		// measurement name (from telegraf metric) is used as multi-measure name in TS
		Log: testutil.Logger{},
	}
	require.Contains(t, invalidConfigMultiMeasureSingleTableMode.Connect().Error(), "MeasureNameForMultiMeasureRecords")

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

func TestWriteMultiMeasuresSingleTableMode(t *testing.T) {
	const recordCount = 100
	mockClient := &mockTimestreamClient{0}

	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return mockClient, nil
	}

	localTime, _ := strconv.Atoi(time1Epoch)

	var inputs []telegraf.Metric

	for i := 1; i <= recordCount+1; i++ {
		localTime++

		fieldName1 := "value_supported1" + strconv.Itoa(i)
		fieldName2 := "value_supported2" + strconv.Itoa(i)
		inputs = append(inputs, testutil.MustMetric(
			"multi_measure_name",
			map[string]string{"tag1": "value1"},
			map[string]interface{}{
				fieldName1: float64(10),
				fieldName2: float64(20),
			},
			time.Unix(int64(localTime), 0),
		))
	}

	plugin := Timestream{
		MappingMode:            MappingModeSingleTable,
		SingleTableName:        "test-multi-single-table-mode",
		DatabaseName:           tsDbName,
		UseMultiMeasureRecords: true, // use multi
		Log:                    testutil.Logger{},
	}

	// validate config correctness
	err := plugin.Connect()
	require.Nil(t, err)

	// validate multi-record generation
	result := plugin.TransformMetrics(inputs)
	// 'inputs' has a total of 101 metrics transformed to 2 writeRecord calls to TS
	require.Equal(t, 2, len(result), "Expected 2 WriteRecordsInput requests")

	var transformedRecords []types.Record
	for _, r := range result {
		transformedRecords = append(transformedRecords, r.Records...)
		// Assert that we use measure name from input
		require.Equal(t, *r.Records[0].MeasureName, "multi_measure_name")
	}
	// Expected 101 records
	require.Equal(t, recordCount+1, len(transformedRecords), "Expected 101 records after transforming")
	// validate write to TS
	err = plugin.Write(inputs)
	require.Nil(t, err, "Write to Timestream failed")
	require.Equal(t, mockClient.WriteRecordsRequestCount, 2, "Expected 2 WriteRecords calls")
}

func TestWriteMultiMeasuresMultiTableMode(t *testing.T) {
	const recordCount = 100
	mockClient := &mockTimestreamClient{0}

	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return mockClient, nil
	}

	localTime, _ := strconv.Atoi(time1Epoch)

	var inputs []telegraf.Metric

	for i := 1; i <= recordCount; i++ {
		localTime++

		fieldName1 := "value_supported1" + strconv.Itoa(i)
		fieldName2 := "value_supported2" + strconv.Itoa(i)
		inputs = append(inputs, testutil.MustMetric(
			"multi_measure_name",
			map[string]string{"tag1": "value1"},
			map[string]interface{}{
				fieldName1: float64(10),
				fieldName2: float64(20),
			},
			time.Unix(int64(localTime), 0),
		))
	}

	plugin := Timestream{
		MappingMode:                       MappingModeMultiTable,
		DatabaseName:                      tsDbName,
		UseMultiMeasureRecords:            true, // use multi
		MeasureNameForMultiMeasureRecords: "config-multi-measure-name",
		Log:                               testutil.Logger{},
	}

	// validate config correctness
	err := plugin.Connect()
	require.Nil(t, err, "Invalid configuration")

	// validate multi-record generation
	result := plugin.TransformMetrics(inputs)
	// 'inputs' has a total of 101 metrics transformed to 2 writeRecord calls to TS
	require.Equal(t, 1, len(result), "Expected 1 WriteRecordsInput requests")

	// Assert that we use measure name from config
	require.Equal(t, *result[0].Records[0].MeasureName, "config-multi-measure-name")

	var transformedRecords []types.Record
	for _, r := range result {
		transformedRecords = append(transformedRecords, r.Records...)
	}
	// Expected 100 records
	require.Equal(t, recordCount, len(transformedRecords), "Expected 100 records after transforming")

	for _, input := range inputs {
		fmt.Println("Input", input)
		fmt.Println(*result[0].Records[0].MeasureName)
		break
	}

	// validate successful write to TS
	err = plugin.Write(inputs)
	require.Nil(t, err, "Write to Timestream failed")
	require.Equal(t, mockClient.WriteRecordsRequestCount, 1, "Expected 1 WriteRecords call")
}

func TestBuildMultiMeasuresInSingleAndMultiTableMode(t *testing.T) {
	input1 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag1": "value1"},
		map[string]interface{}{
			"measureDouble": aws.Float64(10),
		},
		time1,
	)

	input2 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag2": "value2"},
		map[string]interface{}{
			"measureBigint": aws.Int32(20),
		},
		time1,
	)

	input3 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag3": "value3"},
		map[string]interface{}{
			"measureVarchar": "DUMMY",
		},
		time1,
	)

	input4 := testutil.MustMetric(
		metricName1,
		map[string]string{"tag4": "value4"},
		map[string]interface{}{
			"measureBool": true,
		},
		time1,
	)

	expectedResultMultiTable := buildExpectedMultiRecords("config-multi-measure-name", metricName1)

	plugin := Timestream{
		MappingMode:                       MappingModeMultiTable,
		DatabaseName:                      tsDbName,
		UseMultiMeasureRecords:            true, // use multi
		MeasureNameForMultiMeasureRecords: "config-multi-measure-name",
		Log:                               testutil.Logger{},
	}

	// validate config correctness
	err := plugin.Connect()
	require.Nil(t, err, "Invalid configuration")

	// validate multi-record generation with MappingModeMultiTable
	result := plugin.TransformMetrics([]telegraf.Metric{input1, input2, input3, input4})
	require.Equal(t, 1, len(result), "Expected 1 WriteRecordsInput requests")

	require.EqualValues(t, result[0], expectedResultMultiTable)

	require.True(t, arrayContains(result, expectedResultMultiTable), "Expected that the list of requests to Timestream: %+v\n "+
		"will contain request: %+v\n\n", result, expectedResultMultiTable)

	// singleTableMode

	plugin = Timestream{
		MappingMode:            MappingModeSingleTable,
		SingleTableName:        "singleTableName",
		DatabaseName:           tsDbName,
		UseMultiMeasureRecords: true, // use multi
		Log:                    testutil.Logger{},
	}

	// validate config correctness
	err = plugin.Connect()
	require.Nil(t, err, "Invalid configuration")

	expectedResultSingleTable := buildExpectedMultiRecords(metricName1, "singleTableName")

	// validate multi-record generation with MappingModeSingleTable
	result = plugin.TransformMetrics([]telegraf.Metric{input1, input2, input3, input4})
	require.Equal(t, 1, len(result), "Expected 1 WriteRecordsInput requests")

	require.EqualValues(t, result[0], expectedResultSingleTable)

	require.True(t, arrayContains(result, expectedResultSingleTable), "Expected that the list of requests to Timestream: %+v\n "+
		"will contain request: %+v\n\n", result, expectedResultSingleTable)
}

func buildExpectedMultiRecords(multiMeasureName string, tableName string) *timestreamwrite.WriteRecordsInput {
	var recordsMultiTableMode []types.Record
	recordDouble := buildMultiRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag1": "value1"},
			measureValues: map[string]string{"measureDouble": "10"},
		}}, multiMeasureName, types.MeasureValueTypeDouble)

	recordsMultiTableMode = append(recordsMultiTableMode, recordDouble...)

	recordBigint := buildMultiRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag2": "value2"},
			measureValues: map[string]string{"measureBigint": "20"},
		}}, multiMeasureName, types.MeasureValueTypeBigint)

	recordsMultiTableMode = append(recordsMultiTableMode, recordBigint...)

	recordVarchar := buildMultiRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag3": "value3"},
			measureValues: map[string]string{"measureVarchar": "DUMMY"},
		}}, multiMeasureName, types.MeasureValueTypeVarchar)

	recordsMultiTableMode = append(recordsMultiTableMode, recordVarchar...)

	recordBool := buildMultiRecords([]SimpleInput{
		{
			t:             time1Epoch,
			tableName:     metricName1,
			dimensions:    map[string]string{"tag4": "value4"},
			measureValues: map[string]string{"measureBool": "true"},
		},
	}, multiMeasureName, types.MeasureValueTypeBoolean)

	recordsMultiTableMode = append(recordsMultiTableMode, recordBool...)

	expectedResultMultiTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(tableName),
		Records:          recordsMultiTableMode,
		CommonAttributes: &types.Record{},
	}
	return expectedResultMultiTable
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
func TestWriteWhenRequestsGreaterThanMaxWriteGoRoutinesCount(t *testing.T) {
	const maxWriteRecordsCalls = 5
	const maxRecordsInWriteRecordsCall = 100
	const totalRecords = maxWriteRecordsCalls * maxRecordsInWriteRecordsCall
	mockClient := &mockTimestreamClient{0}

	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return mockClient, nil
	}

	plugin := Timestream{
		MappingMode:  MappingModeMultiTable,
		DatabaseName: tsDbName,
		// Spawn only one go routine to serve all 5 write requests
		MaxWriteGoRoutinesCount: 2,
		Log:                     testutil.Logger{},
	}

	require.NoError(t, plugin.Connect())

	var inputs []telegraf.Metric

	for i := 1; i <= totalRecords; i++ {
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

	err := plugin.Write(inputs)
	require.Nil(t, err, "Expected to write without any errors ")
	require.Equal(t, mockClient.WriteRecordsRequestCount, maxWriteRecordsCalls, "Expected 5 calls to WriteRecords")
}

func TestWriteWhenRequestsLesserThanMaxWriteGoRoutinesCount(t *testing.T) {
	t.Skip("Skipping test due to data race, will be re-visited")
	const maxWriteRecordsCalls = 2
	const maxRecordsInWriteRecordsCall = 100
	const totalRecords = maxWriteRecordsCalls * maxRecordsInWriteRecordsCall
	mockClient := &mockTimestreamClient{0}

	WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
		return mockClient, nil
	}

	plugin := Timestream{
		MappingMode:  MappingModeMultiTable,
		DatabaseName: tsDbName,
		// Spawn 5 parallel go routines to serve 2 write requests
		// In this case only 2 of the 5 go routines will process the write requests
		MaxWriteGoRoutinesCount: 5,
		Log:                     testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	var inputs []telegraf.Metric

	for i := 1; i <= totalRecords; i++ {
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

	err := plugin.Write(inputs)
	require.Nil(t, err, "Expected to write without any errors ")
	require.Equal(t, mockClient.WriteRecordsRequestCount, maxWriteRecordsCalls, "Expected 5 calls to WriteRecords")
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

func TestTransformMetricsRequestsAboveLimitAreSplitSingleTable(t *testing.T) {
	const maxRecordsInWriteRecordsCall = 100

	localTime, _ := strconv.Atoi(time1Epoch)

	var inputs []telegraf.Metric

	for i := 1; i <= maxRecordsInWriteRecordsCall+1; i++ {
		localTime++

		fieldName := "value_supported" + strconv.Itoa(i)
		inputs = append(inputs, testutil.MustMetric(
			metricName1,
			map[string]string{"tag1": "value1"},
			map[string]interface{}{
				fieldName: float64(10),
			},
			time.Unix(int64(localTime), 0),
		))
	}

	localTime, _ = strconv.Atoi(time1Epoch)

	var recordsFirstReq []types.Record

	for i := 1; i <= maxRecordsInWriteRecordsCall; i++ {
		localTime++

		recordsFirstReq = append(recordsFirstReq, buildRecord(SimpleInput{
			t:             strconv.Itoa(localTime),
			tableName:     testSingleTableName,
			dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
			measureValues: map[string]string{"value_supported" + strconv.Itoa(i): "10"},
		})...)
	}

	expectedResult1SingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          recordsFirstReq,
		CommonAttributes: &types.Record{},
	}

	var recordsSecondReq []types.Record

	localTime++

	recordsSecondReq = append(recordsSecondReq, buildRecord(SimpleInput{
		t:             strconv.Itoa(localTime),
		tableName:     testSingleTableName,
		dimensions:    map[string]string{"tag1": "value1", testSingleTableDim: metricName1},
		measureValues: map[string]string{"value_supported" + strconv.Itoa(maxRecordsInWriteRecordsCall+1): "10"},
	})...)

	expectedResult2SingleTable := &timestreamwrite.WriteRecordsInput{
		DatabaseName:     aws.String(tsDbName),
		TableName:        aws.String(testSingleTableName),
		Records:          recordsSecondReq,
		CommonAttributes: &types.Record{},
	}

	comparisonTest(t, MappingModeSingleTable,
		inputs,
		[]*timestreamwrite.WriteRecordsInput{expectedResult1SingleTable, expectedResult2SingleTable})
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

func buildMultiRecords(inputs []SimpleInput, multiMeasureName string, measureType types.MeasureValueType) []types.Record {
	var tsRecords []types.Record
	for _, input := range inputs {
		var multiMeasures []types.MeasureValue
		var tsDimensions []types.Dimension

		for k, v := range input.dimensions {
			tsDimensions = append(tsDimensions, types.Dimension{
				Name:  aws.String(k),
				Value: aws.String(v),
			})
		}

		for k, v := range input.measureValues {
			multiMeasures = append(multiMeasures, types.MeasureValue{
				Name:  aws.String(k),
				Value: aws.String(v),
				Type:  measureType,
			})
		}

		tsRecords = append(tsRecords, types.Record{
			MeasureName:      aws.String(multiMeasureName),
			MeasureValueType: "MULTI",
			MeasureValues:    multiMeasures,
			Dimensions:       tsDimensions,
			Time:             aws.String(input.t),
			TimeUnit:         timeUnit,
		})
	}

	return tsRecords
}
