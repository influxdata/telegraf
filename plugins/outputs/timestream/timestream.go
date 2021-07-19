package timestream

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"golang.org/x/net/http2"
	"net"
	"net/http"
)

type (
	Timestream struct {
		Region      string `toml:"region"`
		AccessKey   string `toml:"access_key"`
		SecretKey   string `toml:"secret_key"`
		RoleARN     string `toml:"role_arn"`
		Profile     string `toml:"profile"`
		Filename    string `toml:"shared_credential_file"`
		Token       string `toml:"token"`
		EndpointURL string `toml:"endpoint_url"`

		MappingMode             string `toml:"mapping_mode"`
		DescribeDatabaseOnStart bool   `toml:"describe_database_on_start"`
		DatabaseName            string `toml:"database_name"`

		SingleTableName                                    string `toml:"single_table_name"`
		SingleTableDimensionNameForTelegrafMeasurementName string `toml:"single_table_dimension_name_for_telegraf_measurement_name"`

		CreateTableIfNotExists                        bool              `toml:"create_table_if_not_exists"`
		CreateTableMagneticStoreRetentionPeriodInDays int64             `toml:"create_table_magnetic_store_retention_period_in_days"`
		CreateTableMemoryStoreRetentionPeriodInHours  int64             `toml:"create_table_memory_store_retention_period_in_hours"`
		CreateTableTags                               map[string]string `toml:"create_table_tags"`

		Log telegraf.Logger
		svc WriteClient
	}

	WriteClient interface {
		CreateTable(*timestreamwrite.CreateTableInput) (*timestreamwrite.CreateTableOutput, error)
		WriteRecords(*timestreamwrite.WriteRecordsInput) (*timestreamwrite.WriteRecordsOutput, error)
		DescribeDatabase(*timestreamwrite.DescribeDatabaseInput) (*timestreamwrite.DescribeDatabaseOutput, error)
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

var sampleConfig = `
  ## Amazon Region
  region = "us-east-1"
  
  ## Amazon Credentials
  ## Credentials are loaded in the following order:
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) Explicit credentials from 'access_key' and 'secret_key'
  ## 3) Shared profile from 'profile'
  ## 4) Environment variables
  ## 5) Shared credentials file
  ## 6) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #profile = ""
  #shared_credential_file = ""
  
  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Timestream database where the metrics will be inserted.
  ## The database must exist prior to starting Telegraf.
  database_name = "yourDatabaseNameHere"

  ## Specifies if the plugin should describe the Timestream database upon starting
  ## to validate if it has access necessary permissions, connection, etc., as a safety check.
  ## If the describe operation fails, the plugin will not start 
  ## and therefore the Telegraf agent will not start.
  describe_database_on_start = false

  ## The mapping mode specifies how Telegraf records are represented in Timestream.
  ## Valid values are: single-table, multi-table.
  ## For example, consider the following data in line protocol format:
  ## weather,location=us-midwest,season=summer temperature=82,humidity=71 1465839830100400200
  ## airquality,location=us-west no2=5,pm25=16 1465839830100400200
  ## where weather and airquality are the measurement names, location and season are tags, 
  ## and temperature, humidity, no2, pm25 are fields.
  ## In multi-table mode:
  ##  - first line will be ingested to table named weather
  ##  - second line will be ingested to table named airquality
  ##  - the tags will be represented as dimensions
  ##  - first table (weather) will have two records:
  ##      one with measurement name equals to temperature, 
  ##      another with measurement name equals to humidity
  ##  - second table (airquality) will have two records:
  ##      one with measurement name equals to no2, 
  ##      another with measurement name equals to pm25
  ##  - the Timestream tables from the example will look like this:
  ##      TABLE "weather":
  ##        time | location | season | measure_name | measure_value::bigint
  ##        2016-06-13 17:43:50 | us-midwest | summer | temperature | 82
  ##        2016-06-13 17:43:50 | us-midwest | summer | humidity | 71
  ##      TABLE "airquality":
  ##        time | location | measure_name | measure_value::bigint
  ##        2016-06-13 17:43:50 | us-west | no2 | 5
  ##        2016-06-13 17:43:50 | us-west | pm25 | 16
  ## In single-table mode:
  ##  - the data will be ingested to a single table, which name will be valueOf(single_table_name)
  ##  - measurement name will stored in dimension named valueOf(single_table_dimension_name_for_telegraf_measurement_name)
  ##  - location and season will be represented as dimensions
  ##  - temperature, humidity, no2, pm25 will be represented as measurement name
  ##  - the Timestream table from the example will look like this:
  ##      Assuming:
  ##        - single_table_name = "my_readings"
  ##        - single_table_dimension_name_for_telegraf_measurement_name = "namespace"
  ##      TABLE "my_readings":
  ##        time | location | season | namespace | measure_name | measure_value::bigint
  ##        2016-06-13 17:43:50 | us-midwest | summer | weather | temperature | 82
  ##        2016-06-13 17:43:50 | us-midwest | summer | weather | humidity | 71
  ##        2016-06-13 17:43:50 | us-west | NULL | airquality | no2 | 5
  ##        2016-06-13 17:43:50 | us-west | NULL | airquality | pm25 | 16
  ## In most cases, using multi-table mapping mode is recommended.
  ## However, you can consider using single-table in situations when you have thousands of measurement names.
  mapping_mode = "multi-table"

  ## Only valid and required for mapping_mode = "single-table"
  ## Specifies the Timestream table where the metrics will be uploaded.
  # single_table_name = "yourTableNameHere"

  ## Only valid and required for mapping_mode = "single-table" 
  ## Describes what will be the Timestream dimension name for the Telegraf
  ## measurement name.
  # single_table_dimension_name_for_telegraf_measurement_name = "namespace"

  ## Specifies if the plugin should create the table, if the table do not exist.
  ## The plugin writes the data without prior checking if the table exists.
  ## When the table does not exist, the error returned from Timestream will cause
  ## the plugin to create the table, if this parameter is set to true.
  create_table_if_not_exists = true

  ## Only valid and required if create_table_if_not_exists = true
  ## Specifies the Timestream table magnetic store retention period in days.
  ## Check Timestream documentation for more details.
  create_table_magnetic_store_retention_period_in_days = 365

  ## Only valid and required if create_table_if_not_exists = true
  ## Specifies the Timestream table memory store retention period in hours.
  ## Check Timestream documentation for more details.
  create_table_memory_store_retention_period_in_hours = 24

  ## Only valid and optional if create_table_if_not_exists = true
  ## Specifies the Timestream table tags.
  ## Check Timestream documentation for more details
  # create_table_tags = { "foo" = "bar", "environment" = "dev"}
`

// WriteFactory function provides a way to mock the client instantiation for testing purposes.
var WriteFactory = func(credentialConfig *internalaws.CredentialConfig) WriteClient {
	/**
	* Recommended Timestream write client SDK configuration:
	*  - Use SDK DEFAULT_BACKOFF_STRATEGY
	*  - Request timeout of 20 seconds
	 */

	// Setting 20 seconds for timeout
	tr := &http.Transport{
		ResponseHeaderTimeout: 20 * time.Second,
		// Using DefaultTransport values for other parameters: https://golang.org/pkg/net/http/#RoundTripper
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Timeout:   30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// So client makes HTTP/2 requests
	http2.ConfigureTransport(tr)
	sess := timestreamSession(credentialConfig, tr)
	return timestreamwrite.New(sess)
}

// This is motivated from config/aws/credentials.go with additional settings that are timestream related
var timestreamSession = func(c *internalaws.CredentialConfig, tr *http.Transport) *session.Session {
	if c.RoleARN != "" {
		return assumeCredentials(c, tr)
	} else {
		return rootCredentials(c, tr)
	}
}

var rootCredentials = func(c *internalaws.CredentialConfig, tr *http.Transport) *session.Session {
	config := &aws.Config{
		Region: aws.String(c.Region),
	}
	if c.EndpointURL != "" {
		config.Endpoint = &c.EndpointURL
	}
	if c.AccessKey != "" || c.SecretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, c.Token)
	} else if c.Profile != "" || c.Filename != "" {
		config.Credentials = credentials.NewSharedCredentials(c.Filename, c.Profile)
	}

	config.HTTPClient = &http.Client{Transport: tr}
	return session.New(config)
}

var assumeCredentials = func(c *internalaws.CredentialConfig, tr *http.Transport) *session.Session {
	rootCredentials := rootCredentials(c, tr)
	config := &aws.Config{
		Region:   aws.String(c.Region),
		Endpoint: &c.EndpointURL,
	}
	config.Credentials = stscreds.NewCredentials(rootCredentials, c.RoleARN)
	config.HTTPClient = &http.Client{Transport: tr}
	return session.New(config)
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

	t.Log.Infof("Constructing Timestream client for '%s' mode", t.MappingMode)

	credentialConfig := &internalaws.CredentialConfig{
		Region:      t.Region,
		AccessKey:   t.AccessKey,
		SecretKey:   t.SecretKey,
		RoleARN:     t.RoleARN,
		Profile:     t.Profile,
		Filename:    t.Filename,
		Token:       t.Token,
		EndpointURL: t.EndpointURL,
	}
	svc := WriteFactory(credentialConfig)

	if t.DescribeDatabaseOnStart {
		t.Log.Infof("Describing database '%s' in region '%s'", t.DatabaseName, t.Region)

		describeDatabaseInput := &timestreamwrite.DescribeDatabaseInput{
			DatabaseName: aws.String(t.DatabaseName),
		}
		describeDatabaseOutput, err := svc.DescribeDatabase(describeDatabaseInput)
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

func (t *Timestream) SampleConfig() string {
	return sampleConfig
}

func (t *Timestream) Description() string {
	return "Configuration for Amazon Timestream output."
}

func init() {
	outputs.Add("timestream", func() telegraf.Output {
		return &Timestream{}
	})
}

func (t *Timestream) Write(metrics []telegraf.Metric) error {
	writeRecordsInputs := t.TransformMetrics(metrics)

	errs := make(chan error, len(writeRecordsInputs))
	var wg sync.WaitGroup
	wg.Add(len(writeRecordsInputs))

	start := time.Now()
	for _, writeRecordsInput := range writeRecordsInputs {
		go func(inp *timestreamwrite.WriteRecordsInput) {
			defer wg.Done()
			if err := t.writeToTimestream(inp, true); err != nil {
				errs <- err
			}

		}(writeRecordsInput)
	}

	wg.Wait()
	now := time.Now()
	elapsed := now.Sub(start)

	close(errs)

	t.Log.Infof("##WriteToTimestream - Metrics size: %d request size: %d time(ms): %d",
		len(metrics), len(writeRecordsInputs), elapsed.Milliseconds())

	// On partial failures, Telegraf will reject the entire batch of metrics and
	// retry. writeToTimestream will return retryable exceptions only.
	err, _ := <-errs

	if err != nil {
		return err
	}

	return nil
}

func (t *Timestream) writeToTimestream(writeRecordsInput *timestreamwrite.WriteRecordsInput, resourceNotFoundRetry bool) error {
	t.Log.Debugf("Writing to Timestream: '%v' with ResourceNotFoundRetry: '%t'", writeRecordsInput, resourceNotFoundRetry)

	_, err := t.svc.WriteRecords(writeRecordsInput)
	if err != nil {
		// Telegraf will retry ingesting the metrics if an error is returned from the plugin.
		// Therefore, return error only for retryable exceptions: ThrottlingException and 5xx exceptions.
		if e, ok := err.(awserr.Error); ok {
			switch e.Code() {
			case timestreamwrite.ErrCodeResourceNotFoundException:
				if resourceNotFoundRetry {
					t.Log.Warnf("Failed to write to Timestream database '%s' table '%s'. Error: '%s'",
						t.DatabaseName, *writeRecordsInput.TableName, e)
					return t.createTableAndRetry(writeRecordsInput)
				}
				t.logWriteToTimestreamError(err, writeRecordsInput.TableName)
			case timestreamwrite.ErrCodeThrottlingException:
				return fmt.Errorf("unable to write to Timestream database '%s' table '%s'. Error: %s",
					t.DatabaseName, *writeRecordsInput.TableName, err)
			case timestreamwrite.ErrCodeInternalServerException:
				return fmt.Errorf("unable to write to Timestream database '%s' table '%s'. Error: %s",
					t.DatabaseName, *writeRecordsInput.TableName, err)
			default:
				t.logWriteToTimestreamError(err, writeRecordsInput.TableName)
			}
		} else {
			// Retry other, non-aws errors.
			return fmt.Errorf("unable to write to Timestream database '%s' table '%s'. Error: %s",
				t.DatabaseName, *writeRecordsInput.TableName, err)
		}
	}
	return nil
}

func (t *Timestream) logWriteToTimestreamError(err error, tableName *string) {
	t.Log.Errorf("Failed to write to Timestream database '%s' table '%s'. Skipping metric! Error: '%s'",
		t.DatabaseName, *tableName, err)
}

func (t *Timestream) createTableAndRetry(writeRecordsInput *timestreamwrite.WriteRecordsInput) error {
	var tableName = *writeRecordsInput.TableName
	if t.CreateTableIfNotExists {
		t.Log.Infof("Trying to create table '%s' in database '%s', as 'CreateTableIfNotExists' config key is 'true'.", tableName, t.DatabaseName)
		if err := t.createTable(tableName); err != nil {
			t.Log.Errorf("Failed to create table '%s' in database '%s': %s. Skipping metric!", tableName, t.DatabaseName, err)
		} else {
			t.Log.Infof("Table '%s' in database '%s' created. Retrying writing.", tableName, t.DatabaseName)
			return t.writeToTimestream(writeRecordsInput, false)
		}
	} else {
		t.Log.Errorf("Not trying to create table '%s' in database '%s', as 'CreateTableIfNotExists' config key is 'false'. Skipping metric!", tableName, t.DatabaseName)
	}
	return nil
}

// createTable creates a Timestream table according to the configuration.
func (t *Timestream) createTable(tableName string) error {
	createTableInput := &timestreamwrite.CreateTableInput{
		DatabaseName: aws.String(t.DatabaseName),
		TableName:    aws.String(tableName),
		RetentionProperties: &timestreamwrite.RetentionProperties{
			MagneticStoreRetentionPeriodInDays: aws.Int64(t.CreateTableMagneticStoreRetentionPeriodInDays),
			MemoryStoreRetentionPeriodInHours:  aws.Int64(t.CreateTableMemoryStoreRetentionPeriodInHours),
		},
	}
	var tags []*timestreamwrite.Tag
	for key, val := range t.CreateTableTags {
		tags = append(tags, &timestreamwrite.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		})
	}
	createTableInput.SetTags(tags)

	_, err := t.svc.CreateTable(createTableInput)
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			// if the table was created in the meantime, it's ok.
			if e.Code() == timestreamwrite.ErrCodeConflictException {
				return nil
			}
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
				CommonAttributes: &timestreamwrite.Record{},
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

func hashFromMetricTimeNameTagKeys(m telegraf.Metric) uint64 {
	h := fnv.New64a()
	h.Write([]byte(m.Name()))
	h.Write([]byte("\n"))
	for _, tag := range m.TagList() {
		if tag.Key == "" {
			continue
		}

		h.Write([]byte(tag.Key))
		h.Write([]byte("\n"))
		h.Write([]byte(tag.Value))
		h.Write([]byte("\n"))
	}
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(b, uint64(m.Time().UnixNano()))
	h.Write(b[:n])
	h.Write([]byte("\n"))
	return h.Sum64()
}

func (t *Timestream) buildDimensions(point telegraf.Metric) []*timestreamwrite.Dimension {
	var dimensions []*timestreamwrite.Dimension
	for tagName, tagValue := range point.Tags() {
		dimension := &timestreamwrite.Dimension{
			Name:  aws.String(tagName),
			Value: aws.String(tagValue),
		}
		dimensions = append(dimensions, dimension)
	}
	if t.MappingMode == MappingModeSingleTable {
		dimension := &timestreamwrite.Dimension{
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
func (t *Timestream) buildWriteRecords(point telegraf.Metric) []*timestreamwrite.Record {
	var records []*timestreamwrite.Record
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

		record := &timestreamwrite.Record{
			MeasureName:      aws.String(fieldName),
			MeasureValueType: aws.String(stringFieldValueType),
			MeasureValue:     aws.String(stringFieldValue),
			Dimensions:       dimensions,
			Time:             aws.String(timeValue),
			TimeUnit:         aws.String(timeUnit),
		}

		records = append(records, record)
	}

	return records
}

// partitionRecords splits the Timestream records into smaller slices of a max size
// so that are under the limit for the Timestream API call.
// It returns the array of array of records.
func partitionRecords(size int, records []*timestreamwrite.Record) [][]*timestreamwrite.Record {
	numberOfPartitions := len(records) / size
	if len(records)%size != 0 {
		numberOfPartitions++
	}

	partitions := make([][]*timestreamwrite.Record, numberOfPartitions)

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
func getTimestreamTime(time time.Time) (timeUnit string, timeValue string) {
	const (
		TimeUnitS  = "SECONDS"
		TimeUnitMS = "MILLISECONDS"
		TimeUnitUS = "MICROSECONDS"
		TimeUnitNS = "NANOSECONDS"
	)
	nanosTime := time.UnixNano()
	if nanosTime%1e9 == 0 {
		timeUnit = TimeUnitS
		timeValue = strconv.FormatInt(nanosTime/1e9, 10)
	} else if nanosTime%1e6 == 0 {
		timeUnit = TimeUnitMS
		timeValue = strconv.FormatInt(nanosTime/1e6, 10)
	} else if nanosTime%1e3 == 0 {
		timeUnit = TimeUnitUS
		timeValue = strconv.FormatInt(nanosTime/1e3, 10)
	} else {
		timeUnit = TimeUnitNS
		timeValue = strconv.FormatInt(nanosTime, 10)
	}
	return
}

// convertValue converts single Field value from Telegraf Metric and produces
// value, valueType Timestream representation.
func convertValue(v interface{}) (value string, valueType string, ok bool) {
	const (
		TypeBigInt  = "BIGINT"
		TypeDouble  = "DOUBLE"
		TypeBoolean = "BOOLEAN"
		TypeVarchar = "VARCHAR"
	)
	ok = true

	switch t := v.(type) {
	case int:
		valueType = TypeBigInt
		value = strconv.FormatInt(int64(t), 10)
	case int8:
		valueType = TypeBigInt
		value = strconv.FormatInt(int64(t), 10)
	case int16:
		valueType = TypeBigInt
		value = strconv.FormatInt(int64(t), 10)
	case int32:
		valueType = TypeBigInt
		value = strconv.FormatInt(int64(t), 10)
	case int64:
		valueType = TypeBigInt
		value = strconv.FormatInt(t, 10)
	case uint:
		valueType = TypeBigInt
		value = strconv.FormatUint(uint64(t), 10)
	case uint8:
		valueType = TypeBigInt
		value = strconv.FormatUint(uint64(t), 10)
	case uint16:
		valueType = TypeBigInt
		value = strconv.FormatUint(uint64(t), 10)
	case uint32:
		valueType = TypeBigInt
		value = strconv.FormatUint(uint64(t), 10)
	case uint64:
		valueType = TypeBigInt
		value = strconv.FormatUint(t, 10)
	case float32:
		valueType = TypeDouble
		value = strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		valueType = TypeDouble
		value = strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		valueType = TypeBoolean
		if t {
			value = "true"
		} else {
			value = "false"
		}
	case string:
		valueType = TypeVarchar
		value = t
	default:
		// Skip unsupported type.
		ok = false
		return
	}
	return
}
