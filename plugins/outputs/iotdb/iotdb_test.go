//go:generate ../../../tools/readme_config_includer/generator
package iotdb

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/apache/iotdb-client-go/client"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const testPort = "6667"

func newTestClient() *IoTDB {
	testClient := newIoTDB()
	testClient.Host = "localhost"
	testClient.Port = testPort
	testClient.User = "root"
	testClient.Password = "root"
	testClient.Log = testutil.Logger{}
	return testClient
}

func newMetricWithOrderedFields(
	name string,
	tags []telegraf.Tag,
	fields []telegraf.Field,
	timestamp time.Time,
) telegraf.Metric {
	// This function creates new Metric and makes sure fields in order.
	// `metric.New()` uses map[string]interface{} so fields are NOT in order.
	// `AddField()` makes sure fields are in order, which is necessary for testing.
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
		newMetricWithOrderedFields(
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
		newMetricWithOrderedFields(
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
		newMetricWithOrderedFields(
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
		newMetricWithOrderedFields(
			"root.computer.mouse",
			[]telegraf.Tag{},
			[]telegraf.Field{
				{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
			},
			constTestTimestamp,
		),
	}
)

func TestMetricConversion(t *testing.T) {
	var (
		testRecordsWithTags001 = recordsWithTags{
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
		testRecordsWithTags002 = recordsWithTags{
			DeviceIDList:     []string{"root.computer.mouse"},
			MeasurementsList: [][]string{{"unsigned_big"}},
			ValuesList:       [][]interface{}{{fmt.Sprintf("%d", uint64(math.MaxInt64+1000))}},
			DataTypesList:    [][]client.TSDataType{{client.TEXT}},
			TimestampList:    []int64{constTestTimestamp.UnixNano()},
		}
		testRecordsWithTags003 = recordsWithTags{
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
		testRecordsWithTags004 = recordsWithTags{
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
		testRecordsWithTags005 = recordsWithTags{
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
	)
	cli001 := newTestClient()
	cli002 := newTestClient()
	cli002.ConvertUint64To = "text"
	cli003 := newTestClient()
	cli003.TimeStampUnit = "second"
	cli004 := newTestClient()
	cli004.TreatTagsAs = "fields"
	cli005 := newTestClient()
	cli005.TreatTagsAs = "device_id"
	tests := []struct {
		name        string
		cli         *IoTDB
		expectedRWT recordsWithTags
		metrics     []telegraf.Metric
		needModify  bool // if true, modify before comparing; if false, do not call `modifyRecordsWithTags`
	}{
		{
			name:        "01 test normal config",
			cli:         cli001,
			expectedRWT: testRecordsWithTags001,
			metrics:     testMetrics001,
			needModify:  false,
		},
		{
			name:        "02 test convert unsigned int to text",
			cli:         cli002,
			expectedRWT: testRecordsWithTags002,
			metrics:     testMetrics002,
			needModify:  false,
		},
		{
			name:        "03 test different timestamp precision",
			cli:         cli003,
			expectedRWT: testRecordsWithTags003,
			metrics:     testMetrics001,
			needModify:  false,
		},
		{
			name:        "04 test treat tags as fields",
			cli:         cli004,
			expectedRWT: testRecordsWithTags004,
			metrics:     testMetrics001,
			needModify:  true,
		},
		{
			name:        "05 test treat tags as device id",
			cli:         cli005,
			expectedRWT: testRecordsWithTags005,
			metrics:     testMetrics001,
			needModify:  true,
		},
	}
	for _, testCase := range tests {
		result, err := testCase.cli.convertMetricsToRecordsWithTags(testCase.metrics)
		require.NoError(t, err)
		if testCase.needModify {
			require.NoError(t, testCase.cli.modifyRecordsWithTags(result))
		}
		t.Logf("Testing case named %q", testCase.name)
		require.EqualValues(t, testCase.expectedRWT.DeviceIDList, result.DeviceIDList)
		require.EqualValues(t, testCase.expectedRWT.MeasurementsList, result.MeasurementsList)
		require.EqualValues(t, testCase.expectedRWT.ValuesList, result.ValuesList)
		require.EqualValues(t, testCase.expectedRWT.DataTypesList, result.DataTypesList)
		require.EqualValues(t, testCase.expectedRWT.TimestampList, result.TimestampList)
	}
}

func TestIntegrationLocalServerInserts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// require a local running instance
	t.Skip("No running instance of Apache IoTDB.")
	// create a client and tests two groups of insertion
	testClient := &IoTDB{
		Host:            "localhost",
		Port:            testPort,
		User:            "root",
		Password:        "root",
		Timeout:         config.Duration(time.Second * 5),
		ConvertUint64To: "int64_clip",
		TimeStampUnit:   "nanosecond",
		TreatTagsAs:     "device_id",
	}
	testClient.Log = testutil.Logger{}
	require.NoError(t, testClient.Connect())
	require.NoError(t, testClient.Write(testMetrics001))
	require.NoError(t, testClient.Write(testMetrics002))
	require.NoError(t, testClient.Close())
}
