//go:generate ../../../tools/readme_config_includer/generator
package modbus_server

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/simonvetter/modbus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/modbus_server"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var config string

type MetricSchema struct {
	Register         string  `toml:"register"`
	Address          uint16  `toml:"address"`
	Name             string  `toml:"name"`
	CoilInitialValue bool    `toml:"coil_initial_value,omitempty"`
	Type             string  `toml:"type,omitempty"`
	Bit              uint8   `toml:"bit,omitempty"`
	Scale            float64 `toml:"scale,omitempty"`
	Length           uint16  `toml:"length,omitempty"`
}

type MetricDefinition struct {
	Name         string            `toml:"measurement"`
	MetricSchema []MetricSchema    `toml:"fields"`
	Tags         map[string]string `toml:"tags"`
}

type ModbusServerConfig struct {
	ServerAddress string             `toml:"server_address"`
	ByteOrder     string             `toml:"byte_order"`
	Timeout       time.Duration      `toml:"timeout"`
	MaxClients    uint               `toml:"max_clients"`
	Metrics       []MetricDefinition `toml:"metrics"`
}

type ModbusServer struct {
	ModbusServerConfig
	server      *modbus.ModbusServer
	handler     *modbus_server.Handler
	MemoryMap   map[uint64]map[string]modbus_server.MemoryEntry
	Log         telegraf.Logger `toml:"-"`
	ctx         context.Context
	cancel      context.CancelFunc
	hashGrouper *HashIDGenerator
}

func (*ModbusServer) SampleConfig() string {
	return config
}

func checkMeasurement(measurement MetricDefinition) error {
	memoryLayout := modbus_server.MemoryLayout{}
	fields := make(map[string]bool)

	for _, field := range measurement.MetricSchema {
		// check for duplicate field names
		if _, ok := fields[field.Name]; ok {
			return fmt.Errorf("duplicate field name: %v", field.Name)
		}
		fields[field.Name] = true
		memoryLayout = append(
			memoryLayout, modbus_server.MemoryEntry{
				Address: field.Address, Type: field.Type, Measurement: measurement.Name, Field: field.Name, Register: field.Register, Bit: field.Bit,
				Scale: field.Scale, Length: field.Length,
			},
		)
	}

	_, overlaps, err := memoryLayout.HasOverlap()
	if err != nil {
		return err
	}

	if len(overlaps) > 0 {
		return fmt.Errorf("overlapping addresses: %v in measurement: %v", measurement.Name, overlaps)
	}

	return nil
}

func (m *ModbusServer) checkConfig() (modbus_server.MemoryLayout, []string, error) {
	memoryLayout := modbus_server.MemoryLayout{}
	m.hashGrouper = NewHashIDGenerator()
	for _, entry := range m.Metrics {
		err := checkMeasurement(entry)
		if err != nil {
			return nil, nil, err
		}
		for _, field := range entry.MetricSchema {
			hashID := m.hashGrouper.GetID(entry.Name, entry.Tags)
			memoryLayout = append(
				memoryLayout, modbus_server.MemoryEntry{
					Address: field.Address, Type: field.Type, HashID: hashID, Field: field.Name, Register: field.Register, Bit: field.Bit, Scale: field.Scale,
					Length: field.Length, CoilInitialValue: field.CoilInitialValue,
				},
			)
		}
	}
	_, overlaps, err := memoryLayout.HasOverlap()
	if err != nil {
		return nil, overlaps, err
	}

	return memoryLayout, overlaps, nil
}

func (m *ModbusServer) InitCoilValues(memory modbus_server.MemoryLayout) {
	for _, entry := range memory {
		if entry.CoilInitialValue {
			if entry.Register == "coil" {
				_, err := m.handler.WriteCoils(entry.Address, []bool{entry.CoilInitialValue})
				if err != nil {
					m.Log.Errorf("failed to init coil value: %v", err)
				}
			}
		}
	}
}

func (m *ModbusServer) Init() error {
	// create the server object
	memLayout, overlaps, err := m.checkConfig()
	if err != nil {
		m.Log.Errorf("failed to create server: %v\n", err)
		return err
	}

	if len(overlaps) > 0 {
		m.Log.Warnf("Overlapping addresses: %v", overlaps)
	}

	coils, registers := memLayout.GetCoilsAndRegisters()
	coilOffset, registerOffset := memLayout.GetMemoryOffsets()
	m.MemoryMap, err = memLayout.GetMemoryMappedByHashID()
	if err != nil {
		return err
	}

	m.Log.Debugf("MemoryLayout: %v", memLayout)
	m.Log.Debugf("Metrics: %v", m.Metrics)

	m.Log.Debugf("Coils: %v, Registers: %v, CoilOffset: %v, RegisterOffset: %v", coils, registers, coilOffset, registerOffset)
	m.Log.Debugf("MemoryMap: %v", m.MemoryMap)

	m.handler, err = modbus_server.NewRequestHandler(uint16(len(coils)), coilOffset, uint16(len(registers)), registerOffset, m.Log)
	if err != nil {
		m.Log.Errorf("failed to create server: %v", err)
		return err
	}

	m.server, err = modbus.NewServer(
		&modbus.ServerConfiguration{
			URL:        m.ServerAddress,
			Timeout:    m.Timeout * time.Second,
			MaxClients: m.MaxClients,
		}, m.handler,
	)

	if err != nil {
		m.Log.Errorf("failed to create server: %v\n", err)
		return err
	}

	m.InitCoilValues(memLayout)

	// Create a cancellable context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	return nil
}

func (m *ModbusServer) Write(metrics []telegraf.Metric) error {
	for _, metr := range metrics {
		metr.Accept()
		m.Log.Debugf("--------------------metric: %v------------------\n", metr.Name())
		m.Log.Debugf("tags: %v\n", metr.Tags())
		m.Log.Debugf("fields: %v\n", metr.FieldList())
		m.Log.Debugf("time: %v\n", metr.Time())

		hashID := m.hashGrouper.GetID(metr.Name(), metr.Tags())
		memMap, ok := m.MemoryMap[hashID]
		if !ok {
			m.Log.Errorf("failed to find metric: %v, id %v", metr.Name(), hashID)
			return fmt.Errorf("failed to find metric: %v, id %v", metr.Name(), hashID)
		}
		for _, field := range metr.FieldList() {
			memEntry := memMap[field.Key]
			if memEntry.Register == "coil" {
				_, err := m.handler.WriteCoils(memEntry.Address, []bool{field.Value.(bool)})
				if err != nil {
					m.Log.Errorf("failed to write metric: %v", err)
					return err
				}
			} else if memEntry.Type == "BIT" {
				bitIndex := memEntry.Bit
				bitValue := field.Value.(bool)
				_, err := m.handler.WriteBitToHoldingRegister(memEntry.Address, bitValue, bitIndex)
				if err != nil {
					m.Log.Errorf("failed to write metric: %v", err)
					return err
				}
			} else {
				registerValues, err := modbus_server.ParseMetric(m.ByteOrder, field.Value, memEntry.Type, memEntry.Scale)
				if err != nil {
					m.Log.Errorf("failed to parse metric: %v", err)
					return err
				}
				_, err = m.handler.WriteHoldingRegisters(memEntry.Address, registerValues)
				if err != nil {
					m.Log.Errorf("failed to write metric: %v", err)
					return err
				}
			}
		}
	}
	return nil
}

func (m *ModbusServer) Connect() error {
	// Create a Modbus TCP client configuration
	err := m.server.Start()
	if err != nil {
		return err
	}
	return nil
}

func (m *ModbusServer) Close() error {
	err := m.server.Stop()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	outputs.Add(
		"modbus_server", func() telegraf.Output {
			return &ModbusServer{}
		},
	)
}
