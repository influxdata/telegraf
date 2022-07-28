//go:generate ../../../tools/readme_config_includer/generator
package iotdb

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/apache/iotdb-client-go/client"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var (
	testHost     = "localhost" // The server's (ip) address that you want to connect to.
	testPort     = "6667"      // The server's port that you want to connect to.
	testUser     = "root"
	testPassword = "root"
)

func TestConnectAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testClient := &IoTDB{
		Host:     testHost,
		Port:     testPort,
		User:     testUser,
		Password: testPassword,
	}
	testClient.Log = testutil.Logger{}

	var err error
	err = testClient.Connect()
	require.NoError(t, err)
	err = testClient.Close()
	require.NoError(t, err)
}

func TestInitAndConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	var testClient = &IoTDB{
		Host:     testHost,
		Port:     testPort,
		User:     testUser,
		Password: testPassword,
	}
	testClient.Log = testutil.Logger{}

	var err error
	err = testClient.Init()
	require.NoError(t, err)
	err = testClient.Connect()
	require.NoError(t, err)
	err = testClient.Close()
	require.NoError(t, err)
}

func generateTestMetric(
	name string,
	tags []telegraf.Tag,
	fields []telegraf.Field,
	timestamp time.Time,
) telegraf.Metric {
	m := metric.New(name, map[string]string{}, map[string]interface{}{}, timestamp)
	for _, tag := range tags {
		m.AddTag(tag.Key, tag.Value)
	}
	for _, field := range fields {
		m.AddField(field.Key, field.Value)
	}
	return m
}

var (
	constTestTimestamp = time.Date(2022, time.July, 20, 12, 25, 33, 44, time.UTC)
	testMetrics001     = []telegraf.Metric{
		generateTestMetric(
			"root.computer.fan",
			[]telegraf.Tag{
				{Key: "price", Value: "expensive"},
				{Key: "owner", Value: "cpu"},
			},
			[]telegraf.Field{
				{Key: "temperature", Value: float64(42.55)},
				{Key: "counter", Value: int64(987654321)},
			},
			constTestTimestamp,
		),
		generateTestMetric(
			"root.computer.fan",
			[]telegraf.Tag{ // same keys in different order
				{Key: "owner", Value: "gpu"},
				{Key: "price", Value: "cheap"},
			},
			[]telegraf.Field{
				{Key: "temperature", Value: float64(56.24)},
				{Key: "counter", Value: int64(123456789)},
			},
			constTestTimestamp,
		),
		generateTestMetric(
			"root.computer.keyboard",
			[]telegraf.Tag{},
			[]telegraf.Field{
				{Key: "temperature", Value: float64(30.33)},
				{Key: "counter", Value: int64(123456789)},
				{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
				{Key: "string", Value: "Made in China."},
				{Key: "bool", Value: bool(false)},
				{Key: "int_text", Value: "123456789011"},
			},
			constTestTimestamp,
		),
	}
	testMetrics002 = []telegraf.Metric{
		generateTestMetric(
			"root.computer.mouse",
			[]telegraf.Tag{},
			[]telegraf.Field{
				{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
			},
			constTestTimestamp,
		),
	}
)

// compare two RecordsWithTags, returns True if and only if they are the same.
func compareRecords(rwt1 *RecordsWithTags, rwt2 *RecordsWithTags, log telegraf.Logger) bool {
	if !(len(rwt1.DeviceIDList) == len(rwt2.DeviceIDList) &&
		len(rwt1.MeasurementsList) == len(rwt2.MeasurementsList) &&
		len(rwt1.ValuesList) == len(rwt2.ValuesList) &&
		len(rwt1.DataTypesList) == len(rwt2.DataTypesList) &&
		len(rwt1.TimestampList) == len(rwt2.TimestampList)) {
		// length not match
		log.Errorf("compareRecords Cechk failed. Two RecordsWithTags has different shape.")
		return false
	}
	for index, deviceID := range rwt1.DeviceIDList {
		if !(deviceID == rwt2.DeviceIDList[index]) {
			log.Errorf("compareRecords Cechk failed. rwt1.DeviceIDList[%d]=%v, rwt2.DeviceIDList[%d]=%v.",
				index, deviceID, index, rwt2.DeviceIDList[index])
			return false
		}
	}
	for index, mList := range rwt1.MeasurementsList {
		if !(len(mList) == len(rwt2.MeasurementsList[index])) {
			log.Errorf("compareRecords Cechk failed. Two MeasurementsList has different shape. %d : %d",
				len(mList), len(rwt2.MeasurementsList[index]))
			return false
		}
		for index002, m := range rwt1.MeasurementsList[index] {
			if !(m == rwt2.MeasurementsList[index][index002]) {
				log.Errorf("compareRecords Cechk failed. rwt1.MeasurementsList[%d][%d]=%v, rwt2.MeasurementsList[%d][%d]=%v.",
					index, index002, m, index, index002, rwt2.MeasurementsList[index][index002])
				return false
			}
		}
	}
	for index, mList := range rwt1.ValuesList {
		if !(len(mList) == len(rwt2.ValuesList[index])) {
			log.Errorf("compareRecords Cechk failed. Two ValuesList has different shape. %d : %d",
				len(mList), len(rwt2.ValuesList[index]))
			return false
		}
		for index002, m := range rwt1.ValuesList[index] {
			if !(m == rwt2.ValuesList[index][index002]) {
				log.Errorf("compareRecords Cechk failed. rwt1.ValuesList[%d][%d]=%v, rwt2.ValuesList[%d][%d]=%v.",
					index, index002, m, index, index002, rwt2.ValuesList[index][index002])
				return false
			}
		}
	}
	for index, mList := range rwt1.DataTypesList {
		if !(len(mList) == len(rwt2.DataTypesList[index])) {
			log.Errorf("compareRecords Cechk failed. Two DataTypesList has different shape. %d : %d",
				len(mList), len(rwt2.DataTypesList[index]))
			return false
		}
		for index002, m := range rwt1.DataTypesList[index] {
			if !(m == rwt2.DataTypesList[index][index002]) {
				log.Errorf("compareRecords Cechk failed. rwt1.DataTypesList[%d][%d]=%v, rwt2.DataTypesList[%d][%d]=%v.",
					index, index002, m, index, index002, rwt2.DataTypesList[index][index002])
				return false
			}
		}
	}
	for index, timestamp := range rwt1.TimestampList {
		if !(timestamp == rwt2.TimestampList[index]) {
			log.Errorf("compareRecords Cechk failed. rwt1.DeviceIDList[%d]=%v, rwt2.DeviceIDList[%d]=%v.",
				index, timestamp, index, rwt2.TimestampList[index])
			return false
		}
	}
	return true
}

// util function, test 'Write' with given session and config
func testConnectWriteMetricInThisConf(s *IoTDB, metrics []telegraf.Metric) error {
	connError := s.Connect()
	if connError != nil {
		return connError
	}
	writeError := s.Write(metrics)
	if writeError != nil {
		return writeError
	}
	closeError := s.Close()
	if closeError != nil {
		return closeError
	}
	return nil
}

// Test defualt configuration, uint64 -> int64
func TestMetricConvertion001(t *testing.T) {
	var testClient = &IoTDB{
		Host:            testHost,
		Port:            testPort,
		User:            testUser,
		Password:        testPassword,
		ConvertUint64To: "ToInt64",
		TimeStampUnit:   "nanosecond",
		TreateTagsAs:    "Measurements",
	}
	testClient.Log = testutil.Logger{}

	result, err := testClient.ConvertMetricsToRecordsWithTags(testMetrics001)
	require.NoError(t, err)
	var testRecordsWithTags001 = RecordsWithTags{
		DeviceIDList: []string{"root.computer.fan", "root.computer.fan", "root.computer.keyboard"},
		MeasurementsList: [][]string{
			{"temperature", "counter"}, {"temperature", "counter"},
			{"temperature", "counter", "unsigned_big", "string", "bool", "int_text"},
		},
		ValuesList: [][]interface{}{
			{float64(42.55), int64(987654321)},
			{float64(56.24), int64(123456789)},
			{float64(30.33), int64(123456789), int64(math.MaxInt64), "Made in China.", bool(false), "123456789011"},
		},
		DataTypesList: [][]client.TSDataType{
			{client.DOUBLE, client.INT64},
			{client.DOUBLE, client.INT64},
			{client.DOUBLE, client.INT64, client.INT64, client.TEXT, client.BOOLEAN, client.TEXT},
		},
		TimestampList: []int64{
			constTestTimestamp.UnixNano(), constTestTimestamp.UnixNano(), constTestTimestamp.UnixNano(),
		},
	}
	require.True(t, compareRecords(result, &testRecordsWithTags001, testClient.Log))
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	require.NoError(t, testConnectWriteMetricInThisConf(testClient, testMetrics001))
}

// Test converting uint64 to text.
func TestMetricConvertion002(t *testing.T) {
	var testClient = &IoTDB{
		Host:            testHost,
		Port:            testPort,
		User:            testUser,
		Password:        testPassword,
		ConvertUint64To: "Text",
		TimeStampUnit:   "nanosecond",
		TreateTagsAs:    "Measurements",
	}
	testClient.Log = testutil.Logger{}

	result, err := testClient.ConvertMetricsToRecordsWithTags(testMetrics002)
	require.NoError(t, err)
	testRecordsWithTags002 := RecordsWithTags{
		DeviceIDList:     []string{"root.computer.mouse"},
		MeasurementsList: [][]string{{"unsigned_big"}},
		ValuesList:       [][]interface{}{{fmt.Sprintf("%d", uint64(math.MaxInt64+1000))}},
		DataTypesList:    [][]client.TSDataType{{client.TEXT}},
		TimestampList:    []int64{constTestTimestamp.UnixNano()},
	}
	require.True(t, compareRecords(result, &testRecordsWithTags002, testClient.Log))
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	require.NoError(t, testConnectWriteMetricInThisConf(testClient, testMetrics002))
}

// Test time unit second.
func TestTagsConvertion003(t *testing.T) {
	var testClient = &IoTDB{
		Host:            testHost,
		Port:            testPort,
		User:            testUser,
		Password:        testPassword,
		ConvertUint64To: "ToInt64",
		TimeStampUnit:   "second",
		TreateTagsAs:    "Measurements",
	}
	testClient.Log = testutil.Logger{}

	result, err := testClient.ConvertMetricsToRecordsWithTags(testMetrics001)
	require.NoError(t, err)
	var testRecordsWithTags003 = RecordsWithTags{
		DeviceIDList: []string{"root.computer.fan", "root.computer.fan", "root.computer.keyboard"},
		MeasurementsList: [][]string{
			{"temperature", "counter"}, {"temperature", "counter"},
			{"temperature", "counter", "unsigned_big", "string", "bool", "int_text"},
		},
		ValuesList: [][]interface{}{
			{float64(42.55), int64(987654321)},
			{float64(56.24), int64(123456789)},
			{float64(30.33), int64(123456789), int64(math.MaxInt64), "Made in China.", bool(false), "123456789011"},
		},
		DataTypesList: [][]client.TSDataType{
			{client.DOUBLE, client.INT64},
			{client.DOUBLE, client.INT64},
			{client.DOUBLE, client.INT64, client.INT64, client.TEXT, client.BOOLEAN, client.TEXT},
		},
		TimestampList: []int64{
			constTestTimestamp.Unix(), constTestTimestamp.Unix(), constTestTimestamp.Unix(),
		},
	}
	require.True(t, compareRecords(result, &testRecordsWithTags003, testClient.Log))
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	require.NoError(t, testConnectWriteMetricInThisConf(testClient, testMetrics001))
}

// Test Tags modification in method 'Measurements'
func TestTagsConvertion004(t *testing.T) {
	var testClient = &IoTDB{
		Host:            testHost,
		Port:            testPort,
		User:            testUser,
		Password:        testPassword,
		ConvertUint64To: "ToInt64",
		TimeStampUnit:   "nanosecond",
		TreateTagsAs:    "Measurements",
	}
	testClient.Log = testutil.Logger{}

	result, err := testClient.ConvertMetricsToRecordsWithTags(testMetrics001)
	require.NoError(t, err)
	err = testClient.ModifiyRecordsWithTags(result)
	require.NoError(t, err)
	testRecordsWithTags004 := RecordsWithTags{
		DeviceIDList: []string{"root.computer.fan", "root.computer.fan", "root.computer.keyboard"},
		MeasurementsList: [][]string{
			{"temperature", "counter", "owner", "price"}, {"temperature", "counter", "owner", "price"},
			{"temperature", "counter", "unsigned_big", "string", "bool", "int_text"},
		},
		ValuesList: [][]interface{}{
			{float64(42.55), int64(987654321), "cpu", "expensive"},
			{float64(56.24), int64(123456789), "gpu", "cheap"},
			{float64(30.33), int64(123456789), int64(math.MaxInt64), "Made in China.", bool(false), "123456789011"},
		},
		DataTypesList: [][]client.TSDataType{
			{client.DOUBLE, client.INT64, client.TEXT, client.TEXT},
			{client.DOUBLE, client.INT64, client.TEXT, client.TEXT},
			{client.DOUBLE, client.INT64, client.INT64, client.TEXT, client.BOOLEAN, client.TEXT},
		},
		TimestampList: []int64{
			constTestTimestamp.UnixNano(), constTestTimestamp.UnixNano(), constTestTimestamp.UnixNano(),
		},
	}
	require.True(t, compareRecords(result, &testRecordsWithTags004, testClient.Log))
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	require.NoError(t, testConnectWriteMetricInThisConf(testClient, testMetrics001))
}

// Test Tags modification in method 'DeviceID_subtree'
func TestTagsConvertion005(t *testing.T) {
	var testClient = &IoTDB{
		Host:            testHost,
		Port:            testPort,
		User:            testUser,
		Password:        testPassword,
		ConvertUint64To: "ToInt64",
		TimeStampUnit:   "nanosecond",
		TreateTagsAs:    "DeviceID_subtree",
	}
	testClient.Log = testutil.Logger{}

	result, err := testClient.ConvertMetricsToRecordsWithTags(testMetrics001)
	require.NoError(t, err)
	err = testClient.ModifiyRecordsWithTags(result)
	require.NoError(t, err)
	testRecordsWithTags005 := RecordsWithTags{
		DeviceIDList: []string{"root.computer.fan.cpu.expensive", "root.computer.fan.gpu.cheap", "root.computer.keyboard"},
		MeasurementsList: [][]string{
			{"temperature", "counter"}, {"temperature", "counter"},
			{"temperature", "counter", "unsigned_big", "string", "bool", "int_text"},
		},
		ValuesList: [][]interface{}{
			{float64(42.55), int64(987654321)},
			{float64(56.24), int64(123456789)},
			{float64(30.33), int64(123456789), int64(math.MaxInt64), "Made in China.", bool(false), "123456789011"},
		},
		DataTypesList: [][]client.TSDataType{
			{client.DOUBLE, client.INT64},
			{client.DOUBLE, client.INT64},
			{client.DOUBLE, client.INT64, client.INT64, client.TEXT, client.BOOLEAN, client.TEXT},
		},
		TimestampList: []int64{
			constTestTimestamp.UnixNano(), constTestTimestamp.UnixNano(), constTestTimestamp.UnixNano(),
		},
	}
	require.True(t, compareRecords(result, &testRecordsWithTags005, testClient.Log))
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	require.NoError(t, testConnectWriteMetricInThisConf(testClient, testMetrics001))
}
