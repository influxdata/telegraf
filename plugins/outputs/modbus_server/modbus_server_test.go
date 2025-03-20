package modbus_server

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/modbus_server"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitOpenClose(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Timeout:       60 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "measurement1",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 0, Name: "field1", Type: "UINT16"},
						{Register: "coil", Address: 1, Name: "field2", Type: "FLOAT32"},
					},
					Tags: map[string]string{"tag1": "value1"},
				},
			},
		},
		Log: testutil.Logger{},
	}

	require.NoError(t, m.Init())
	require.NoError(t, m.Connect())
	require.NoError(t, m.Close())
}

func TestSampleConfig(t *testing.T) {
	m := &ModbusServer{}
	require.NotEmpty(t, m.SampleConfig())
}

func TestCheckConfig(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1", Type: "BIT"},
						{Register: "register", Address: 2, Name: "field2", Type: "UINT16"},
					},
				},
			},
		},
	}
	memoryLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	require.NotEmpty(t, memoryLayout)
}

func TestCheckConfigAddressesOutOfRange(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1", Type: "BIT"},
						{Register: "register", Address: 2, Name: "field2", Type: "UINT16"},
						{Register: "register", Address: 3, Name: "field3", Type: "UINT16"},
						{Register: "register", Address: 4, Name: "field4", Type: "UINT16"},
						{Register: "register", Address: 10, Name: "field5", Type: "UINT16"},
						{Register: "register", Address: 11, Name: "field6", Type: "UINT16"},
						{Register: "register", Address: 12, Name: "field7", Type: "UINT16"},
					},
				},
			},
		},
	}
	memoryLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	require.Len(t, memoryLayout, 7)
}

func TestOverlappingEntries(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Timeout:       60 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 1, Name: "field1", Type: "UINT16"},
					},
				},
				{
					Name: "test_metric1",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 1, Name: "field1", Type: "UINT16"},
					},
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}

	require.NoError(t, m.Init())
	memMap, _, err := m.checkConfig()
	require.NoError(t, err)
	require.NotEmpty(t, memMap)
}

func TestDuplicateFields(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Timeout:       60 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 1, Name: "field1", Type: "UINT16"},
						{Register: "register", Address: 2, Name: "field1", Type: "UINT16"},
					},
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}
	require.Error(t, m.Init())
	memMap, _, err := m.checkConfig()
	require.Error(t, err)
	require.Empty(t, memMap)
}

func TestMemoryOverlap(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Timeout:       60 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 1, Name: "field1", Type: "UINT32"},
						{Register: "register", Address: 2, Name: "field2", Type: "UINT16"},
					},
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}
	require.Error(t, m.Init())
	memMap, _, err := m.checkConfig()
	require.Error(t, err)
	require.Empty(t, memMap)
}

func TestCheckMeasurement(t *testing.T) {
	tests := []struct {
		measurement MetricDefinition
		expectError bool
	}{
		{
			measurement: MetricDefinition{
				Name: "measurement1",
				MetricSchema: []MetricSchema{
					{Register: "coil", Address: 0, Name: "field1", Type: "UINT16"},
					{Register: "coil", Address: 1, Name: "field2", Type: "FLOAT32"},
				},
				Tags: map[string]string{"tag1": "value1"},
			},
			expectError: false,
		},
		{
			measurement: MetricDefinition{
				Name: "measurement2",
				MetricSchema: []MetricSchema{
					{Register: "coil", Address: 0, Name: "field1", Type: "UINT16"},
					{Register: "coil", Address: 0, Name: "field1", Type: "FLOAT32"},
				},
				Tags: map[string]string{"tag1": "value1"},
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		err := checkMeasurement(test.measurement)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestMetricsConversion(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1"},
						{Register: "coil", Address: 2, Name: "field2"},

						{Register: "register", Address: 3, Name: "field3", Type: "UINT16"},
						{Register: "register", Address: 4, Name: "field4", Type: "UINT16"},
						{Register: "register", Address: 5, Name: "field5", Type: "UINT32"},
					},
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}

	var err error
	m.handler, err = modbus_server.NewRequestHandler(2, 1, 5, 3, testutil.Logger{})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(1, []bool{false, true})
	require.NoError(t, err)
	_, err = m.handler.WriteHoldingRegisters(3, []uint16{123, 321, math.MaxUint16, math.MaxUint16})
	require.NoError(t, err)

	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	m.MemoryMap, err = memLayout.GetMemoryMappedByHashID()

	require.NoError(t, err)
	// generate metrics
	var metrics []telegraf.Metric

	// Define sample fields for the metric
	field1 := map[string]interface{}{
		"field1": true,
	}
	field2 := map[string]interface{}{
		"field2": false,
	}
	field3 := map[string]interface{}{
		"field3": uint16(3),
	}
	field4 := map[string]interface{}{
		"field4": uint16(4),
	}
	field5 := map[string]interface{}{
		"field5": uint32(5),
	}
	// Define tags for the metric
	tags := map[string]string{}

	// Create the metric
	metrics = append(
		metrics,
		metric.New("test_metric", tags, field1, time.Now()),
		metric.New("test_metric", tags, field2, time.Now()),
		metric.New("test_metric", tags, field3, time.Now()),
		metric.New("test_metric", tags, field4, time.Now()),
		metric.New("test_metric", tags, field5, time.Now()),
	)

	require.NoError(t, m.Write(metrics))

	coils, err := m.handler.ReadCoils(1, 2)
	require.NoError(t, err)
	registers, err := m.handler.ReadHoldingRegisters(3, 5)
	require.NoError(t, err)
	// check if the values are updated
	require.True(t, coils[0])
	require.False(t, coils[1])
	require.Equal(t, uint16(3), registers[0])
	require.Equal(t, uint16(4), registers[1])
	require.Equal(t, uint16(0), registers[2])
	require.Equal(t, uint16(5), registers[3])
}

func TestMetricsConversionTwoMeasurements(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric1",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1", Type: "BIT"},
						{Register: "coil", Address: 2, Name: "field2", Type: "BIT"},
						{Register: "register", Address: 3, Name: "field3", Type: "UINT16"},
					},
					Tags: map[string]string{},
				},
				{
					Name: "test_metric2",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 4, Name: "field4", Type: "UINT16"},
						{Register: "register", Address: 5, Name: "field5", Type: "UINT32"},
					},
					Tags: map[string]string{},
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}

	var err error
	m.handler, err = modbus_server.NewRequestHandler(2, 1, 5, 3, testutil.Logger{})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(1, []bool{false, true})
	require.NoError(t, err)
	_, err = m.handler.WriteHoldingRegisters(3, []uint16{123, 321, math.MaxUint16, math.MaxUint16})
	require.NoError(t, err)
	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	m.MemoryMap, err = memLayout.GetMemoryMappedByHashID()

	require.NoError(t, err)
	// generate metrics
	var metrics []telegraf.Metric

	// Define sample fields for the metrics
	field1 := map[string]interface{}{
		"field1": true,
	}
	field2 := map[string]interface{}{
		"field2": false,
	}
	field3 := map[string]interface{}{
		"field3": uint16(3),
	}
	field4 := map[string]interface{}{
		"field4": uint16(4),
	}
	field5 := map[string]interface{}{
		"field5": uint32(5),
	}
	tags := map[string]string{}

	// Create the metrics
	metrics = append(
		metrics,
		metric.New("test_metric1", tags, field1, time.Now()),
		metric.New("test_metric1", tags, field2, time.Now()),
		metric.New("test_metric1", tags, field3, time.Now()),
		metric.New("test_metric2", tags, field4, time.Now()),
		metric.New("test_metric2", tags, field5, time.Now()),
	)

	require.NoError(t, m.Write(metrics))

	coils, err := m.handler.ReadCoils(1, 2)
	require.NoError(t, err)
	registers, err := m.handler.ReadHoldingRegisters(3, 4)
	require.NoError(t, err)
	// check if the values are updated
	require.True(t, coils[0])
	require.False(t, coils[1])
	require.Equal(t, uint16(3), registers[0])
	require.Equal(t, uint16(4), registers[1])
	require.Equal(t, uint16(0), registers[2])
	require.Equal(t, uint16(5), registers[3])
}

func TestSameMetricsDifferentTags(t *testing.T) {
	tags1 := map[string]string{"tag1": "value1"}
	tags2 := map[string]string{"tag2": "value2"}

	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 0, Name: "field0", Type: "UINT16"},
						{Register: "register", Address: 1, Name: "field1", Type: "UINT16"},
					},
					Tags: tags1,
				},
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 0, Name: "field0", Type: "UINT16"},
						{Register: "register", Address: 1, Name: "field1", Type: "UINT16"},
					},
					Tags: tags2,
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}

	var err error
	m.handler, err = modbus_server.NewRequestHandler(0, 0, 2, 0, testutil.Logger{})
	require.NoError(t, err)
	_, err = m.handler.WriteHoldingRegisters(0, []uint16{123, 321})
	require.NoError(t, err)
	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	m.MemoryMap, err = memLayout.GetMemoryMappedByHashID()

	require.NoError(t, err)
	// generate metrics
	var metrics []telegraf.Metric

	// Define sample fields for the metrics
	field0 := map[string]interface{}{
		"field0": uint16(3),
	}
	field1 := map[string]interface{}{
		"field1": uint32(4),
	}

	// Write the first series
	metrics = append(
		metrics,
		metric.New("test_metric", map[string]string{"tag1": "value1"}, field0, time.Now()),
		metric.New("test_metric", map[string]string{"tag1": "value1"}, field1, time.Now()),
	)

	require.NoError(t, m.Write(metrics))

	res, err := m.handler.ReadHoldingRegisters(0, 2)
	require.NoError(t, err)
	require.Equal(t, uint16(3), res[0])
	require.Equal(t, uint16(4), res[1])

	// Reset registers
	field0 = map[string]interface{}{
		"field0": uint16(0),
	}
	field1 = map[string]interface{}{
		"field1": uint32(0),
	}

	// Write the second series
	metrics = make([]telegraf.Metric, 0)
	metrics = append(
		metrics,
		metric.New("test_metric", map[string]string{"tag2": "value2"}, field0, time.Now()),
		metric.New("test_metric", map[string]string{"tag2": "value2"}, field1, time.Now()),
	)

	require.NoError(t, m.Write(metrics))
	res, err = m.handler.ReadHoldingRegisters(0, 2)
	require.NoError(t, err)
	require.Equal(t, uint16(0), res[0])
	require.Equal(t, uint16(0), res[1])
}

func TestWriteBitType(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					Tags: map[string]string{},
					MetricSchema: func() []MetricSchema {
						schema := make([]MetricSchema, 16)
						for i := 0; i < 16; i++ {
							schema[i] = MetricSchema{Register: "register", Address: 0, Name: fmt.Sprintf("field%d", i+1), Type: "BIT", Bit: uint8(i)}
						}
						return schema
					}(),
				},
			},
		},
		Log: testutil.Logger{},
	}

	var err error
	m.handler, err = modbus_server.NewRequestHandler(0, 0, 1, 0, testutil.Logger{})
	require.NoError(t, err)
	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	m.MemoryMap, err = memLayout.GetMemoryMappedByHashID()
	require.NoError(t, err)

	res, err := m.handler.ReadHoldingRegisters(0, 1)
	require.NoError(t, err)
	require.Equal(t, uint16(0), res[0])

	// Define sample fields for the metric
	fields := map[string]interface{}{
		"field1": true, "field2": false, "field3": true, "field4": false,
		"field5": true, "field6": false, "field7": true, "field8": false,
		"field9": true, "field10": false, "field11": true, "field12": false,
		"field13": true, "field14": false, "field15": true, "field16": false,
	}
	tags := map[string]string{}

	// Create the metric
	metrics := make([]telegraf.Metric, 0, len(fields))
	for k, v := range fields {
		metrics = append(metrics, metric.New("test_metric", tags, map[string]interface{}{k: v}, time.Now()))
	}

	require.NoError(t, m.Write(metrics))

	res, err = m.handler.ReadHoldingRegisters(0, 1)
	require.NoError(t, err)
	// Check if the values are updated 21845 = 0b0101010101010101
	require.Equal(t, uint16(21845), res[0])

	// Define sample fields for the metric
	fields = map[string]interface{}{
		"field1": false, "field2": true, "field3": false, "field4": true,
		"field5": false, "field6": true, "field7": false, "field8": true,
		"field9": false, "field10": true, "field11": false, "field12": true,
		"field13": false, "field14": true, "field15": false, "field16": true,
	}
	// Create the metric
	metrics = make([]telegraf.Metric, 0, len(fields))
	for k, v := range fields {
		metrics = append(metrics, metric.New("test_metric", tags, map[string]interface{}{k: v}, time.Now()))
	}

	require.NoError(t, m.Write(metrics))

	res, err = m.handler.ReadHoldingRegisters(0, 1)
	require.NoError(t, err)
	// Check if the values are updated 43690 = 0b1010101010101010
	require.Equal(t, uint16(43690), res[0])
}

func TestWriteStringMetric(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_string_metric",
					MetricSchema: []MetricSchema{
						{Register: "register", Address: 0, Name: "field1", Type: "STRING", Length: 3},
					},
				},
			},
		},
		Log: testutil.Logger{},
	}

	var err error
	m.handler, err = modbus_server.NewRequestHandler(0, 0, 3, 0, testutil.Logger{})
	require.NoError(t, err)
	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	m.MemoryMap, err = memLayout.GetMemoryMappedByHashID()
	require.NoError(t, err)

	// Define sample fields for the metric
	field1 := map[string]interface{}{
		"field1": "Hello",
	}
	// Define tags for the metric
	tags := make(map[string]string)

	// Create the metric
	metrics := []telegraf.Metric{
		metric.New("test_string_metric", tags, field1, time.Now()),
	}

	require.NoError(t, m.Write(metrics))

	// Check if the values are updated
	expectedRegisters := []uint16{0x4865, 0x6c6c, 0x6f00} // "Hello" in UTF-16

	res, err := m.handler.ReadHoldingRegisters(0, 3)
	require.NoError(t, err)
	require.Equal(t, expectedRegisters, res)
}

func TestInitCoilValues(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:5502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 0, Name: "field1", CoilInitialValue: true},
						{Register: "coil", Address: 1, Name: "field2", CoilInitialValue: false},
						{Register: "coil", Address: 2, Name: "field3"},
					},
				},
			},
		},
		Log: testutil.Logger{},
	}

	var err error
	m.handler, err = modbus_server.NewRequestHandler(3, 0, 0, 0, testutil.Logger{})
	require.NoError(t, err)
	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)

	m.InitCoilValues(memLayout)

	// Check if the coil value is initialized correctly
	coils, err := m.handler.ReadCoils(0, 3)
	require.NoError(t, err)
	require.True(t, coils[0])
	require.False(t, coils[1])
	require.False(t, coils[2])
}
