package timestream

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	"github.com/aws/smithy-go"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type (
	Timestream struct {
		MappingMode             string `toml:"mapping_mode"`
		DescribeDatabaseOnStart bool   `toml:"describe_database_on_start"`
		DatabaseName            string `toml:"database_name"`

		SingleTableName                                    string `toml:"single_table_name"`
		SingleTableDimensionNameForTelegrafMeasurementName string `toml:"single_table_dimension_name_for_telegraf_measurement_name"`

		CreateTableIfNotExists                        bool              `toml:"create_table_if_not_exists"`
		CreateTableMagneticStoreRetentionPeriodInDays int64             `toml:"create_table_magnetic_store_retention_period_in_days"`
		CreateTableMemoryStoreRetentionPeriodInHours  int64             `toml:"create_table_memory_store_retention_period_in_hours"`
		CreateTableTags                               map[string]string `toml:"create_table_tags"`
		MaxWriteGoRoutinesCount                       int               `toml:"max_write_go_routines"`

		Log telegraf.Logger
		svc WriteClient

		internalaws.CredentialConfig
	}

	WriteClient interface {
		CreateTable(context.Context, *timestreamwrite.CreateTableInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.CreateTableOutput, error)
		WriteRecords(context.Context, *timestreamwrite.WriteRecordsInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.WriteRecordsOutput, error)
		DescribeDatabase(context.Context, *timestreamwrite.DescribeDatabaseInput, ...func(*timestreamwrite.Options)) (*timestreamwrite.DescribeDatabaseOutput, error)
	}
)

// Mapping modes specify how Telegraf model should be represented in Timestream model.
// See sample config for more details.
const (
	MappingModeSingleTable = "single-table"
	MappingModeMultiTable  = "multi-table"
)

// MaxRecordsPerCall reflects Timestream limit of WriteRecords API call
const MaxRecordsPerCall = 100

// Default value for maximum number of parallel go routines to ingest/write data
// when max_write_go_routines is not specified in the config
const MaxWriteRoutinesDefault = 1

// WriteFactory function provides a way to mock the client instantiation for testing purposes.
var WriteFactory = func(credentialConfig *internalaws.CredentialConfig) (WriteClient, error) {
	cfg, err := credentialConfig.Credentials()
	if err != nil {
		return &timestreamwrite.Client{}, err
	}
	return timestreamwrite.NewFromConfig(cfg), nil
}

func (t *Timestream) Connect() error {
	if t.DatabaseName == "" {
		return fmt.Errorf("DatabaseName key is required")
	}

	if t.MappingMode == "" {
		return fmt.Errorf("MappingMode key is required")
	}

	if t.MappingMode != MappingModeSingleTable && t.MappingMode != MappingModeMultiTable {
		return fmt.Errorf("correct MappingMode key values are: '%s', '%s'",
			MappingModeSingleTable, MappingModeMultiTable)
	}

	if t.MappingMode == MappingModeSingleTable {
		if t.SingleTableName == "" {
			return fmt.Errorf("in '%s' mapping mode, SingleTableName key is required", MappingModeSingleTable)
		}

		if t.SingleTableDimensionNameForTelegrafMeasurementName == "" {
			return fmt.Errorf("in '%s' mapping mode, SingleTableDimensionNameForTelegrafMeasurementName key is required",
				MappingModeSingleTable)
		}
	}

	if t.MappingMode == MappingModeMultiTable {
		if t.SingleTableName != "" {
			return fmt.Errorf("in '%s' mapping mode, do not specify SingleTableName key", MappingModeMultiTable)
		}

		if t.SingleTableDimensionNameForTelegrafMeasurementName != "" {
			return fmt.Errorf("in '%s' mapping mode, do not specify SingleTableDimensionNameForTelegrafMeasurementName key", MappingModeMultiTable)
		}
	}

	if t.CreateTableIfNotExists {
		if t.CreateTableMagneticStoreRetentionPeriodInDays < 1 {
			return fmt.Errorf("if Telegraf should create tables, CreateTableMagneticStoreRetentionPeriodInDays key should have a value greater than 0")
		}

		if t.CreateTableMemoryStoreRetentionPeriodInHours < 1 {
			return fmt.Errorf("if Telegraf should create tables, CreateTableMemoryStoreRetentionPeriodInHours key should have a value greater than 0")
		}
	}

	if t.MaxWriteGoRoutinesCount <= 0 {
		t.MaxWriteGoRoutinesCount = MaxWriteRoutinesDefault
	}

	t.Log.Infof("Constructing Timestream client for '%s' mode", t.MappingMode)

	svc, err := WriteFactory(&t.CredentialConfig)
	if err != nil {
		return err
	}

	if t.DescribeDatabaseOnStart {
		t.Log.Infof("Describing database '%s' in region '%s'", t.DatabaseName, t.Region)

		describeDatabaseInput := &timestreamwrite.DescribeDatabaseInput{
			DatabaseName: aws.String(t.DatabaseName),
		}
		describeDatabaseOutput, err := svc.DescribeDatabase(context.Background(), describeDatabaseInput)
		if err != nil {
			t.Log.Errorf("Couldn't describe database '%s'. Check error, fix permissions, connectivity, create database.", t.DatabaseName)
			return err
		}
		t.Log.Infof("Describe database '%s' returned: '%s'.", t.DatabaseName, describeDatabaseOutput)
	}

	t.svc = svc
	return nil
}

func (t *Timestream) Close() error {
	return nil
}

func init() {
	outputs.Add("timestream", func() telegraf.Output {
		return &Timestream{}
	})
}

func (t *Timestream) Write(metrics []telegraf.Metric) error {
	writeRecordsInputs := t.TransformMetrics(metrics)

	maxWriteJobs := t.MaxWriteGoRoutinesCount
	numberOfWriteRecordsInputs := len(writeRecordsInputs)

	if numberOfWriteRecordsInputs < maxWriteJobs {
		maxWriteJobs = numberOfWriteRecordsInputs
	}

	var wg sync.WaitGroup
	errs := make(chan error, numberOfWriteRecordsInputs)
	writeJobs := make(chan *timestreamwrite.WriteRecordsInput, maxWriteJobs)

	start := time.Now()

	for i := 0; i < maxWriteJobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for writeJob := range writeJobs {
				if err := t.writeToTimestream(writeJob, true); err != nil {
					errs <- err
				}
			}
		}()
	}

	for i := range writeRecordsInputs {
		writeJobs <- writeRecordsInputs[i]
	}

	// Close channel once all jobs are added
	close(writeJobs)

	wg.Wait()
	elapsed := time.Since(start)

	close(errs)

	t.Log.Infof("##WriteToTimestream - Metrics size: %d request size: %d time(ms): %d",
		len(metrics), len(writeRecordsInputs), elapsed.Milliseconds())

	// On partial failures, Telegraf will reject the entire batch of metrics and
	// retry. writeToTimestream will return retryable exceptions only.
	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Timestream) writeToTimestream(writeRecordsInput *timestreamwrite.WriteRecordsInput, resourceNotFoundRetry bool) error {
	t.Log.Debugf("Writing to Timestream: '%v' with ResourceNotFoundRetry: '%t'", writeRecordsInput, resourceNotFoundRetry)

	_, err := t.svc.WriteRecords(context.Background(), writeRecordsInput)
	if err != nil {
		// Telegraf will retry ingesting the metrics if an error is returned from the plugin.
		// Therefore, return error only for retryable exceptions: ThrottlingException and 5xx exceptions.
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			if resourceNotFoundRetry {
				t.Log.Warnf("Failed to write to Timestream database '%s' table '%s'. Error: '%s'",
					t.DatabaseName, *writeRecordsInput.TableName, notFound)
				return t.createTableAndRetry(writeRecordsInput)
			}
			t.logWriteToTimestreamError(notFound, writeRecordsInput.TableName)
		}

		var rejected *types.RejectedRecordsException
		if errors.As(err, &rejected) {
			t.logWriteToTimestreamError(err, writeRecordsInput.TableName)
			return nil
		}

		var throttling *types.ThrottlingException
		if errors.As(err, &throttling) {
			return fmt.Errorf("unable to write to Timestream database '%s' table '%s'. Error: %s",
				t.DatabaseName, *writeRecordsInput.TableName, throttling)
		}

		var internal *types.InternalServerException
		if errors.As(err, &internal) {
			return fmt.Errorf("unable to write to Timestream database '%s' table '%s'. Error: %s",
				t.DatabaseName, *writeRecordsInput.TableName, internal)
		}

		var operation *smithy.OperationError
		if !errors.As(err, &operation) {
			// Retry other, non-aws errors.
			return fmt.Errorf("unable to write to Timestream database '%s' table '%s'. Error: %s",
				t.DatabaseName, *writeRecordsInput.TableName, err)
		}
		t.logWriteToTimestreamError(err, writeRecordsInput.TableName)
	}
	return nil
}

func (t *Timestream) logWriteToTimestreamError(err error, tableName *string) {
	t.Log.Errorf("Failed to write to Timestream database '%s' table '%s'. Skipping metric! Error: '%s'",
		t.DatabaseName, *tableName, err)
}

func (t *Timestream) createTableAndRetry(writeRecordsInput *timestreamwrite.WriteRecordsInput) error {
	if t.CreateTableIfNotExists {
		t.Log.Infof("Trying to create table '%s' in database '%s', as 'CreateTableIfNotExists' config key is 'true'.", *writeRecordsInput.TableName, t.DatabaseName)
		err := t.createTable(writeRecordsInput.TableName)
		if err == nil {
			t.Log.Infof("Table '%s' in database '%s' created. Retrying writing.", *writeRecordsInput.TableName, t.DatabaseName)
			return t.writeToTimestream(writeRecordsInput, false)
		}
		t.Log.Errorf("Failed to create table '%s' in database '%s': %s. Skipping metric!", *writeRecordsInput.TableName, t.DatabaseName, err)
	} else {
		t.Log.Errorf("Not trying to create table '%s' in database '%s', as 'CreateTableIfNotExists' config key is 'false'. Skipping metric!", *writeRecordsInput.TableName, t.DatabaseName)
	}
	return nil
}

// createTable creates a Timestream table according to the configuration.
func (t *Timestream) createTable(tableName *string) error {
	createTableInput := &timestreamwrite.CreateTableInput{
		DatabaseName: aws.String(t.DatabaseName),
		TableName:    aws.String(*tableName),
		RetentionProperties: &types.RetentionProperties{
			MagneticStoreRetentionPeriodInDays: t.CreateTableMagneticStoreRetentionPeriodInDays,
			MemoryStoreRetentionPeriodInHours:  t.CreateTableMemoryStoreRetentionPeriodInHours,
		},
	}
	var tags []types.Tag
	for key, val := range t.CreateTableTags {
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		})
	}
	createTableInput.Tags = tags

	_, err := t.svc.CreateTable(context.Background(), createTableInput)
	if err != nil {
		if _, ok := err.(*types.ConflictException); ok {
			// if the table was created in the meantime, it's ok.
			return nil
		}
		return err
	}
	return nil
}

// TransformMetrics transforms a collection of Telegraf Metrics into write requests to Timestream.
// Telegraf Metrics are grouped by Name, Tag Keys and Time to use Timestream CommonAttributes.
// Returns collection of write requests to be performed to Timestream.
func (t *Timestream) TransformMetrics(metrics []telegraf.Metric) []*timestreamwrite.WriteRecordsInput {
	writeRequests := make(map[string]*timestreamwrite.WriteRecordsInput, len(metrics))
	for _, m := range metrics {
		// build MeasureName, MeasureValue, MeasureValueType
		records := t.buildWriteRecords(m)
		if len(records) == 0 {
			continue
		}

		var tableName string

		if t.MappingMode == MappingModeSingleTable {
			tableName = t.SingleTableName
		}

		if t.MappingMode == MappingModeMultiTable {
			tableName = m.Name()
		}

		if curr, ok := writeRequests[tableName]; !ok {
			newWriteRecord := &timestreamwrite.WriteRecordsInput{
				DatabaseName:     aws.String(t.DatabaseName),
				TableName:        aws.String(tableName),
				Records:          records,
				CommonAttributes: &types.Record{},
			}

			writeRequests[tableName] = newWriteRecord
		} else {
			curr.Records = append(curr.Records, records...)
		}
	}

	// Create result as array of WriteRecordsInput. Split requests over records count limit to smaller requests.
	var result []*timestreamwrite.WriteRecordsInput
	for _, writeRequest := range writeRequests {
		if len(writeRequest.Records) > MaxRecordsPerCall {
			for _, recordsPartition := range partitionRecords(MaxRecordsPerCall, writeRequest.Records) {
				newWriteRecord := &timestreamwrite.WriteRecordsInput{
					DatabaseName:     writeRequest.DatabaseName,
					TableName:        writeRequest.TableName,
					Records:          recordsPartition,
					CommonAttributes: writeRequest.CommonAttributes,
				}
				result = append(result, newWriteRecord)
			}
		} else {
			result = append(result, writeRequest)
		}
	}
	return result
}

func (t *Timestream) buildDimensions(point telegraf.Metric) []types.Dimension {
	var dimensions []types.Dimension
	for tagName, tagValue := range point.Tags() {
		dimension := types.Dimension{
			Name:  aws.String(tagName),
			Value: aws.String(tagValue),
		}
		dimensions = append(dimensions, dimension)
	}
	if t.MappingMode == MappingModeSingleTable {
		dimension := types.Dimension{
			Name:  aws.String(t.SingleTableDimensionNameForTelegrafMeasurementName),
			Value: aws.String(point.Name()),
		}
		dimensions = append(dimensions, dimension)
	}
	return dimensions
}

// buildWriteRecords builds the Timestream write records from Metric Fields only.
// Tags and time are not included - common attributes are built separately.
// Records with unsupported Metric Field type are skipped.
// It returns an array of Timestream write records.
func (t *Timestream) buildWriteRecords(point telegraf.Metric) []types.Record {
	var records []types.Record

	dimensions := t.buildDimensions(point)

	for fieldName, fieldValue := range point.Fields() {
		stringFieldValue, stringFieldValueType, ok := convertValue(fieldValue)
		if !ok {
			t.Log.Errorf("Skipping field '%s'. The type '%s' is not supported in Timestream as MeasureValue. "+
				"Supported values are: [int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool]",
				fieldName, reflect.TypeOf(fieldValue))
			continue
		}

		timeUnit, timeValue := getTimestreamTime(point.Time())

		record := types.Record{
			MeasureName:      aws.String(fieldName),
			MeasureValueType: stringFieldValueType,
			MeasureValue:     aws.String(stringFieldValue),
			Dimensions:       dimensions,
			Time:             aws.String(timeValue),
			TimeUnit:         timeUnit,
		}
		records = append(records, record)
	}
	return records
}

// partitionRecords splits the Timestream records into smaller slices of a max size
// so that are under the limit for the Timestream API call.
// It returns the array of array of records.
func partitionRecords(size int, records []types.Record) [][]types.Record {
	numberOfPartitions := len(records) / size
	if len(records)%size != 0 {
		numberOfPartitions++
	}

	partitions := make([][]types.Record, numberOfPartitions)

	for i := 0; i < numberOfPartitions; i++ {
		start := size * i
		end := size * (i + 1)
		if end > len(records) {
			end = len(records)
		}

		partitions[i] = records[start:end]
	}

	return partitions
}

// getTimestreamTime produces Timestream TimeUnit and TimeValue with minimum possible granularity
// while maintaining the same information.
func getTimestreamTime(t time.Time) (timeUnit types.TimeUnit, timeValue string) {
	nanosTime := t.UnixNano()
	if nanosTime%1e9 == 0 {
		timeUnit = types.TimeUnitSeconds
		timeValue = strconv.FormatInt(nanosTime/1e9, 10)
	} else if nanosTime%1e6 == 0 {
		timeUnit = types.TimeUnitMilliseconds
		timeValue = strconv.FormatInt(nanosTime/1e6, 10)
	} else if nanosTime%1e3 == 0 {
		timeUnit = types.TimeUnitMicroseconds
		timeValue = strconv.FormatInt(nanosTime/1e3, 10)
	} else {
		timeUnit = types.TimeUnitNanoseconds
		timeValue = strconv.FormatInt(nanosTime, 10)
	}
	return timeUnit, timeValue
}

// convertValue converts single Field value from Telegraf Metric and produces
// value, valueType Timestream representation.
func convertValue(v interface{}) (value string, valueType types.MeasureValueType, ok bool) {
	ok = true

	switch t := v.(type) {
	case int:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatInt(int64(t), 10)
	case int8:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatInt(int64(t), 10)
	case int16:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatInt(int64(t), 10)
	case int32:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatInt(int64(t), 10)
	case int64:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatInt(t, 10)
	case uint:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatUint(uint64(t), 10)
	case uint8:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatUint(uint64(t), 10)
	case uint16:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatUint(uint64(t), 10)
	case uint32:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatUint(uint64(t), 10)
	case uint64:
		valueType = types.MeasureValueTypeBigint
		value = strconv.FormatUint(t, 10)
	case float32:
		valueType = types.MeasureValueTypeDouble
		value = strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		valueType = types.MeasureValueTypeDouble
		value = strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		valueType = types.MeasureValueTypeBoolean
		if t {
			value = "true"
		} else {
			value = "false"
		}
	case string:
		valueType = types.MeasureValueTypeVarchar
		value = t
	default:
		// Skip unsupported type.
		ok = false
		return value, valueType, ok
	}
	return value, valueType, ok
}
