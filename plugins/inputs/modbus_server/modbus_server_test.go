package modbus_server

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/simonvetter/modbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/modbus_server"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Metrics:       nil,
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, m.Init())
	require.NoError(t, m.Start(acc))
	require.NoError(t, m.Gather(acc))

	m.Stop()
	require.NoError(t, acc.FirstError())
}

func TestSampleConfig(t *testing.T) {
	m := &ModbusServer{}
	require.NotEmpty(t, m.SampleConfig())
}

func TestCheckConfig(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1"},
						{Register: "register", Address: 2, Name: "field2", Type: "UINT16"},
						{Register: "register", Address: 40000, Name: "bit_field0", Type: "BIT", Bit: 0},
						{Register: "register", Address: 40000, Name: "bit_field1", Type: "BIT", Bit: 1},
						{Register: "register", Address: 40000, Name: "bit_field2", Type: "BIT", Bit: 2},
						{Register: "register", Address: 40000, Name: "bit_field3", Type: "BIT", Bit: 3},
						{Register: "register", Address: 40000, Name: "bit_field4", Type: "BIT", Bit: 4},
						{Register: "register", Address: 40000, Name: "bit_field5", Type: "BIT", Bit: 5},
						{Register: "register", Address: 40000, Name: "bit_field6", Type: "BIT", Bit: 6},
						{Register: "register", Address: 40000, Name: "bit_field7", Type: "BIT", Bit: 7},
						{Register: "register", Address: 40000, Name: "bit_field8", Type: "BIT", Bit: 8},
						{Register: "register", Address: 40000, Name: "bit_field9", Type: "BIT", Bit: 9},
						{Register: "register", Address: 40000, Name: "bit_field10", Type: "BIT", Bit: 10},
						{Register: "register", Address: 40000, Name: "bit_field11", Type: "BIT", Bit: 11},
						{Register: "register", Address: 40000, Name: "bit_field12", Type: "BIT", Bit: 12},
						{Register: "register", Address: 40000, Name: "bit_field13", Type: "BIT", Bit: 13},
						{Register: "register", Address: 40000, Name: "bit_field14", Type: "BIT", Bit: 14},
						{Register: "register", Address: 40000, Name: "bit_field15", Type: "BIT", Bit: 15},
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
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1"},
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

func TestGetMetrics(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
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
	m.handler, err = modbus_server.NewRequestHandler(2, 1, 4, 3, testutil.Logger{})
	require.NoError(t, err)
	_, err = m.handler.WriteHoldingRegisters(3, []uint16{123, 321, math.MaxUint16, math.MaxUint16})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(1, []bool{true, false})
	require.NoError(t, err)
	_, _, err = m.checkConfig()
	require.NoError(t, err)

	metrics := m.getMetrics(time.Now())
	require.Len(t, metrics, 5)
	require.Equal(t, "test_metric", metrics[0].Name())
	require.Equal(t, true, metrics[0].Fields()["field1"])

	require.Equal(t, "test_metric", metrics[1].Name())
	require.Equal(t, false, metrics[1].Fields()["field2"])

	require.Equal(t, "test_metric", metrics[2].Name())
	require.Equal(t, uint64(123), metrics[2].Fields()["field3"])

	require.Equal(t, "test_metric", metrics[3].Name())
	require.Equal(t, uint64(321), metrics[3].Fields()["field4"])

	require.Equal(t, "test_metric", metrics[4].Name())
	require.Equal(t, uint64(math.MaxUint32), metrics[4].Fields()["field5"])
}

func TestStartStop(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Timeout:       2 * time.Second,
			MaxClients:    5,
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, m.Init())
	require.NoError(t, m.Start(acc))
	m.Stop()
}

func TestOverlappingEntries(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Timeout:       2 * time.Second,
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
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Timeout:       2 * time.Second,
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

func TestUpdateMemory(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 10, Name: "field1"},
						{Register: "coil", Address: 11, Name: "field2"},

						{Register: "register", Address: 30, Name: "field3", Type: "UINT16"},
						{Register: "register", Address: 31, Name: "field4", Type: "UINT16"},
						{Register: "register", Address: 32, Name: "field5", Type: "UINT32"},
					},
				},
			},
		},
		Log: testutil.Logger{
			Name:  "",
			Quiet: false,
		},
	}
	// init
	memLayout, _, err := m.checkConfig()
	require.NoError(t, err)
	coils, registers := memLayout.GetCoilsAndRegisters()
	coilOffset, registerOffset := memLayout.GetMemoryOffsets()
	m.handler, err = modbus_server.NewRequestHandler(uint16(len(coils)), coilOffset, uint16(len(registers)), registerOffset, testutil.Logger{})
	require.NoError(t, err)

	_, err = m.handler.WriteHoldingRegisters(30, []uint16{123, 321, math.MaxUint16, math.MaxUint16})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(10, []bool{true, false})
	require.NoError(t, err)

	readCoils, err := m.handler.ReadCoils(10, 2)
	require.NoError(t, err)
	require.Equal(t, []bool{true, false}, readCoils)

	readRegisters, err := m.handler.ReadHoldingRegisters(30, 4)
	require.NoError(t, err)
	require.Equal(t, []uint16{123, 321, math.MaxUint16, math.MaxUint16}, readRegisters)

	// get metrics
	require.NoError(t, err)
	metrics := m.getMetrics(time.Now())
	require.Len(t, metrics, 5)
	require.Equal(t, "test_metric", metrics[0].Name())
	require.Equal(t, true, metrics[0].Fields()["field1"])

	require.Equal(t, "test_metric", metrics[1].Name())
	require.Equal(t, false, metrics[1].Fields()["field2"])

	require.Equal(t, "test_metric", metrics[2].Name())
	require.Equal(t, uint64(123), metrics[2].Fields()["field3"])

	require.Equal(t, "test_metric", metrics[3].Name())
	require.Equal(t, uint64(321), metrics[3].Fields()["field4"])

	require.Equal(t, "test_metric", metrics[4].Name())
	require.Equal(t, uint64(math.MaxUint32), metrics[4].Fields()["field5"])

	// update memory
	_, err = m.handler.WriteHoldingRegisters(30, []uint16{111})
	require.NoError(t, err)
	_, err = m.handler.WriteHoldingRegisters(31, []uint16{222, 0, 333})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(10, []bool{false, false})
	require.NoError(t, err)

	// check metrics update
	metrics = m.getMetrics(time.Now())
	require.Len(t, metrics, 5)
	require.Equal(t, "test_metric", metrics[0].Name())
	require.Equal(t, false, metrics[0].Fields()["field1"])

	require.Equal(t, "test_metric", metrics[1].Name())
	require.Equal(t, false, metrics[1].Fields()["field2"])

	require.Equal(t, "test_metric", metrics[2].Name())
	require.Equal(t, uint64(111), metrics[2].Fields()["field3"])

	require.Equal(t, "test_metric", metrics[3].Name())
	require.Equal(t, uint64(222), metrics[3].Fields()["field4"])

	require.Equal(t, "test_metric", metrics[4].Name())
	require.Equal(t, uint64(333), metrics[4].Fields()["field5"])
}

func TestMemoryOverlap(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Timeout:       2 * time.Second,
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

func TestAccumulatedMetrics(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Timeout:       2 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1"},
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
	var err error
	m.handler, err = modbus_server.NewRequestHandler(1, 1, 1, 2, testutil.Logger{})
	require.NoError(t, err)

	_, err = m.handler.WriteHoldingRegisters(2, []uint16{123})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(1, []bool{true})
	require.NoError(t, err)
	acc := &testutil.Accumulator{}
	require.NoError(t, m.Init())
	require.NoError(t, m.Start(acc))
	require.Empty(t, acc.Metrics)
	// Update last edit time to trigger more metrics
	_, err = m.handler.WriteCoils(1, []bool{true})
	require.NoError(t, err)

	acc.Wait(2)
	require.Len(t, acc.Metrics, 2)
	// Update last edit time to trigger new metrics
	_, err = m.handler.WriteCoils(1, []bool{false})
	require.NoError(t, err)

	acc.Wait(4)
	require.Len(t, acc.Metrics, 4)

	m.Stop()
}

func TestAccumulatedMetricsNoNewUpdates(t *testing.T) {
	m := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: "tcp://localhost:10502",
			ByteOrder:     "ABCD",
			Timeout:       2 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "test_metric",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 1, Name: "field1"},
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
	var err error
	m.handler, err = modbus_server.NewRequestHandler(1, 1, 1, 2, testutil.Logger{})
	require.NoError(t, err)

	_, err = m.handler.WriteHoldingRegisters(2, []uint16{123})
	require.NoError(t, err)
	_, err = m.handler.WriteCoils(1, []bool{true})
	require.NoError(t, err)

	acc := &testutil.Accumulator{}
	require.NoError(t, m.Init())
	require.NoError(t, m.Start(acc))
	require.Empty(t, acc.Metrics)
	// Update last edit time to trigger more metrics
	m.handler.LastEdit <- time.Now()
	acc.Wait(2)
	require.Len(t, acc.Metrics, 2)
	// No new updates
	time.Sleep(1 * time.Millisecond)
	require.Len(t, acc.Metrics, 2)

	m.Stop()
}

func TestModbusServerIntegration(t *testing.T) {
	// Create a ModbusServer instance
	serverAddr := "tcp://localhost:10502"
	server := &ModbusServer{
		ModbusServerConfig: ModbusServerConfig{
			ServerAddress: serverAddr,
			ByteOrder:     "ABCD",
			Timeout:       5 * 60 * time.Second,
			MaxClients:    5,
			Metrics: []MetricDefinition{
				{
					Name: "measurement1",
					MetricSchema: []MetricSchema{
						{Register: "coil", Address: 0, Name: "field1"},
						{Register: "coil", Address: 1, Name: "field2"},
						{Register: "coil", Address: 2, Name: "field3"},
						{Register: "holding", Address: 40001, Name: "float_field", Type: "FLOAT32"},
					},
					Tags: map[string]string{
						"tag1": "value1",
						"tag2": "value2",
					},
				},
			},
		},
		Log: testutil.Logger{},
	}

	// Initialize the server
	require.NoError(t, server.Init())

	// Create a test accumulator
	acc := &testutil.Accumulator{}

	// Start the server
	require.NoError(t, server.Start(acc))
	defer server.Stop()

	serverCoils, _ := server.handler.GetCoilsAndOffset()
	serverRegisters, _ := server.handler.GetRegistersAndOffset()
	require.Equal(t, []bool{false, false, false}, serverCoils)
	require.Equal(t, []uint16{0, 0}, serverRegisters)

	_, err := server.handler.WriteCoils(0, []bool{false, false, false})
	require.NoError(t, err)
	_, err = server.handler.WriteHoldingRegisters(40001, []uint16{0, 0})
	require.NoError(t, err)

	// Add a delay to ensure the server is fully up and running
	client, err := modbus.NewClient(
		&modbus.ClientConfiguration{
			URL:     serverAddr,
			Timeout: 10 * time.Second,
		},
	)

	require.NoError(t, err)

	// Create wait group to wait for the client operations to complete
	var wg sync.WaitGroup
	wg.Add(1)
	// Run the client operations in a separate goroutine
	go func() {
		defer wg.Done()
		// Open the client connection
		err := client.Open()
		assert.NoError(t, err)

		defer func(client *modbus.ModbusClient) {
			err := client.Close()
			assert.NoError(t, err)
		}(client)
		// Read 1 coil
		coil, err := client.ReadCoil(0)
		assert.NoError(t, err)
		assert.False(t, coil)

		// Read 1 register
		register, err := client.ReadRegister(40001, 0)
		assert.NoError(t, err)
		assert.Equal(t, uint16(0), register)

		// Write coils
		err = client.WriteCoils(0, []bool{true, false, true})
		assert.NoError(t, err)

		// Read coils
		coils, err := client.ReadCoils(0, 3)
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, false, true}, coils)

		// Write holding registers
		err = client.WriteRegisters(40001, []uint16{0x3f80, 0x0000})
		assert.NoError(t, err)

		// Read holding registers
		registers, err := client.ReadRegisters(40001, 2, 0)
		assert.NoError(t, err)
		assert.Equal(t, []uint16{0x3f80, 0x0000}, registers)
	}()

	// Wait for the client operations to complete
	wg.Wait()
}
