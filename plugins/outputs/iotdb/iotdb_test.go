//go:generate ../../../tools/readme_config_includer/generator
package iotdb

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/apache/iotdb-client-go/client"
	"github.com/docker/go-connections/nat"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

// newMetricWithOrderedFields creates new Metric and makes sure fields are in
// order. This is required to define the expected output where the field order
// needs to be defines.
func newMetricWithOrderedFields(
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

func TestInitInvalid(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *IoTDB
		expected string
	}{
		{
			name: "empty tag-conversion",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.TreatTagsAs = ""
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `unknown 'convert_tags_to' method ""`,
		},
		{
			name: "empty uint-conversion",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.ConvertUint64To = ""
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `unknown 'uint64_conversion' method ""`,
		},
		{
			name: "empty timestamp precision",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.TimeStampUnit = ""
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `unknown 'timestamp_precision' method ""`,
		},
		{
			name: "invalid tag-conversion",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.TreatTagsAs = "garbage"
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `unknown 'convert_tags_to' method "garbage"`,
		},
		{
			name: "invalid uint-conversion",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.ConvertUint64To = "garbage"
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `unknown 'uint64_conversion' method "garbage"`,
		},
		{
			name: "invalid timestamp precision",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.TimeStampUnit = "garbage"
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `unknown 'timestamp_precision' method "garbage"`,
		},
		{
			name: "negative timeout",
			plugin: func() *IoTDB {
				s := newIoTDB()
				s.Timeout = config.Duration(time.Second * -5)
				s.Log = &testutil.Logger{}
				return s
			}(),
			expected: `negative timeout`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualError(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestMetricConversion(t *testing.T) {
	var testTimestamp = time.Date(2022, time.July, 20, 12, 25, 33, 44, time.UTC)

	tests := []struct {
		name         string
		plugin       *IoTDB
		expected     recordsWithTags
		metrics      []telegraf.Metric
		doTestModify bool // if true, modify before comparing; if false, do not call `modifyRecordsWithTags`
	}{
		{
			name:   "default config",
			plugin: func() *IoTDB { s := newIoTDB(); return s }(),
			expected: recordsWithTags{
				DeviceIDList: []string{"root.computer.fan", "root.computer.fan", "root.computer.keyboard"},
				MeasurementsList: [][]string{
					{"temperature", "counter"},
					{"counter", "temperature"},
					{"temperature", "counter", "unsigned_big", "string", "bool", "int_text"},
				},
				ValuesList: [][]interface{}{
					{float64(42.55), int64(987654321)},
					{int64(123456789), float64(56.24)},
					{float64(30.33), int64(123456789), int64(math.MaxInt64), "Made in China.", bool(false), "123456789011"},
				},
				DataTypesList: [][]client.TSDataType{
					{client.DOUBLE, client.INT64},
					{client.INT64, client.DOUBLE},
					{client.DOUBLE, client.INT64, client.INT64, client.TEXT, client.BOOLEAN, client.TEXT},
				},
				TimestampList: []int64{testTimestamp.UnixNano(), testTimestamp.UnixNano(), testTimestamp.UnixNano()},
			},
			metrics: []telegraf.Metric{
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
					testTimestamp,
				),
				newMetricWithOrderedFields(
					"root.computer.fan",
					[]telegraf.Tag{ // same keys in different order
						{Key: "owner", Value: "gpu"},
						{Key: "price", Value: "cheap"},
					},
					[]telegraf.Field{ // same keys in different order
						{Key: "counter", Value: int64(123456789)},
						{Key: "temperature", Value: float64(56.24)},
					},
					testTimestamp,
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
					testTimestamp,
				),
			},
			doTestModify: false,
		},
		{
			name:   "unsigned int to text",
			plugin: func() *IoTDB { cli002 := newIoTDB(); cli002.ConvertUint64To = "text"; return cli002 }(),
			expected: recordsWithTags{
				DeviceIDList:     []string{"root.computer.uint_to_text"},
				MeasurementsList: [][]string{{"unsigned_big"}},
				ValuesList:       [][]interface{}{{strconv.FormatUint(uint64(math.MaxInt64+1000), 10)}},
				DataTypesList:    [][]client.TSDataType{{client.TEXT}},
				TimestampList:    []int64{testTimestamp.UnixNano()},
			},
			metrics: []telegraf.Metric{
				newMetricWithOrderedFields(
					"root.computer.uint_to_text",
					[]telegraf.Tag{},
					[]telegraf.Field{
						{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
					},
					testTimestamp,
				),
			},
			doTestModify: false,
		},
		{
			name:   "unsigned int to int with overflow",
			plugin: func() *IoTDB { cli002 := newIoTDB(); cli002.ConvertUint64To = "int64"; return cli002 }(),
			expected: recordsWithTags{
				DeviceIDList:     []string{"root.computer.overflow"},
				MeasurementsList: [][]string{{"unsigned_big"}},
				ValuesList:       [][]interface{}{{int64(-9223372036854774809)}},
				DataTypesList:    [][]client.TSDataType{{client.INT64}},
				TimestampList:    []int64{testTimestamp.UnixNano()},
			},
			metrics: []telegraf.Metric{
				newMetricWithOrderedFields(
					"root.computer.overflow",
					[]telegraf.Tag{},
					[]telegraf.Field{
						{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
					},
					testTimestamp,
				),
			},
			doTestModify: false,
		},
		{
			name:   "second timestamp precision",
			plugin: func() *IoTDB { s := newIoTDB(); s.TimeStampUnit = "second"; return s }(),
			expected: recordsWithTags{
				DeviceIDList:     []string{"root.computer.second"},
				MeasurementsList: [][]string{{"unsigned_big"}},
				ValuesList:       [][]interface{}{{int64(math.MaxInt64)}},
				DataTypesList:    [][]client.TSDataType{{client.INT64}},
				TimestampList:    []int64{testTimestamp.Unix()},
			},
			metrics: []telegraf.Metric{
				newMetricWithOrderedFields(
					"root.computer.second",
					[]telegraf.Tag{},
					[]telegraf.Field{
						{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
					},
					testTimestamp,
				),
			},
			doTestModify: false,
		},
		{ //treat tags as fields. And input Tags are NOT in order.
			name:   "treat tags as fields",
			plugin: func() *IoTDB { s := newIoTDB(); s.TreatTagsAs = "fields"; return s }(),
			expected: recordsWithTags{
				DeviceIDList:     []string{"root.computer.fields"},
				MeasurementsList: [][]string{{"temperature", "counter", "owner", "price"}},
				ValuesList: [][]interface{}{
					{float64(42.55), int64(987654321), "cpu", "expensive"},
				},
				DataTypesList: [][]client.TSDataType{
					{client.DOUBLE, client.INT64, client.TEXT, client.TEXT},
				},
				TimestampList: []int64{testTimestamp.UnixNano()},
			},
			metrics: []telegraf.Metric{
				newMetricWithOrderedFields(
					"root.computer.fields",
					[]telegraf.Tag{
						{Key: "price", Value: "expensive"},
						{Key: "owner", Value: "cpu"},
					},
					[]telegraf.Field{
						{Key: "temperature", Value: float64(42.55)},
						{Key: "counter", Value: int64(987654321)},
					},
					testTimestamp,
				),
			},
			doTestModify: true,
		},
		{ //treat tags as device IDs. And input Tags are in order.
			name:   "treat tags as device IDs",
			plugin: func() *IoTDB { s := newIoTDB(); s.TreatTagsAs = "device_id"; return s }(),
			expected: recordsWithTags{
				DeviceIDList:     []string{"root.computer.deviceID.cpu.expensive"},
				MeasurementsList: [][]string{{"temperature", "counter"}},
				ValuesList: [][]interface{}{
					{float64(42.55), int64(987654321)},
				},
				DataTypesList: [][]client.TSDataType{
					{client.DOUBLE, client.INT64},
				},
				TimestampList: []int64{testTimestamp.UnixNano()},
			},
			metrics: []telegraf.Metric{
				newMetricWithOrderedFields(
					"root.computer.deviceID",
					[]telegraf.Tag{
						{Key: "owner", Value: "cpu"},
						{Key: "price", Value: "expensive"},
					},
					[]telegraf.Field{
						{Key: "temperature", Value: float64(42.55)},
						{Key: "counter", Value: int64(987654321)},
					},
					testTimestamp,
				),
			},
			doTestModify: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = &testutil.Logger{}
			require.NoError(t, tt.plugin.Init())
			actual, err := tt.plugin.convertMetricsToRecordsWithTags(tt.metrics)
			require.NoError(t, err)
			if tt.doTestModify { // then `expected` is after modified; else `expected` is before modified.
				require.NoError(t, tt.plugin.modifyRecordsWithTags(actual))
			}
			// Ignore the tags-list for comparison
			actual.TagsList = nil
			require.EqualValues(t, &tt.expected, actual)
		})
	}
}

// Start a container and do integration test.
func TestIntegrationInserts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	const iotdbPort = "6667"

	container := testutil.Container{
		Image:        "apache/iotdb:0.13.0-node",
		ExposedPorts: []string{iotdbPort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port(iotdbPort)),
			wait.ForLog("IoTDB has started."),
		),
	}
	err := container.Start()
	require.NoError(t, err, "failed to start IoTDB container")
	defer func() {
		require.NoError(t, container.Terminate(), "terminating IoTDB container failed")
	}()

	t.Logf("Container Address:%q, ExposedPorts:[%v:%v]", container.Address, container.Ports[iotdbPort], iotdbPort)
	// create a client and tests two groups of insertion
	testClient := &IoTDB{
		Host:            container.Address,
		Port:            container.Ports[iotdbPort],
		User:            "root",
		Password:        "root",
		Timeout:         config.Duration(time.Second * 5),
		ConvertUint64To: "int64_clip",
		TimeStampUnit:   "nanosecond",
		TreatTagsAs:     "device_id",
	}
	testClient.Log = &testutil.Logger{}

	// generate Metrics to input
	metrics := []telegraf.Metric{
		newMetricWithOrderedFields(
			"root.computer.unsigned_big",
			[]telegraf.Tag{},
			[]telegraf.Field{
				{Key: "unsigned_big", Value: uint64(math.MaxInt64 + 1000)},
			},
			time.Now(),
		),
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
			time.Now(),
		),
		newMetricWithOrderedFields(
			"root.computer.fan",
			[]telegraf.Tag{ // same keys in different order
				{Key: "owner", Value: "gpu"},
				{Key: "price", Value: "cheap"},
			},
			[]telegraf.Field{ // same keys in different order
				{Key: "counter", Value: int64(123456789)},
				{Key: "temperature", Value: float64(56.24)},
			},
			time.Now(),
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
			time.Now(),
		),
	}

	require.NoError(t, testClient.Connect())
	require.NoError(t, testClient.Write(metrics))
	require.NoError(t, testClient.Close())
}
