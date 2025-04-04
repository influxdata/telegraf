//go:generate ../../../tools/readme_config_includer/generator
package modbus_server

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/simonvetter/modbus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/modbus_server"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var config string

type MetricSchema struct {
	Register string  `toml:"register"`
	Address  uint16  `toml:"address"`
	Name     string  `toml:"name"`
	Type     string  `toml:"type,omitempty"`
	Bit      uint8   `toml:"bit,omitempty"`
	Scale    float64 `toml:"scale,omitempty"`
	Length   uint16  `toml:"length,omitempty"`
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
	server  *modbus.ModbusServer
	handler *modbus_server.Handler
	Log     telegraf.Logger `toml:"-"`
	ctx     context.Context
	cancel  context.CancelFunc
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
				Register: field.Register, Address: field.Address, Type: field.Type, Bit: field.Bit, Scale: field.Scale, Length: field.Length,
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

	for _, entry := range m.Metrics {
		err := checkMeasurement(entry)
		if err != nil {
			return nil, nil, err
		}

		for _, field := range entry.MetricSchema {
			memoryLayout = append(
				memoryLayout, modbus_server.MemoryEntry{
					Address: field.Address, Type: field.Type, Register: field.Register, Bit: field.Bit, Scale: field.Scale, Length: field.Length,
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

func (m *ModbusServer) getMetrics(timestamp time.Time) []telegraf.Metric {
	coils, coilOffset := m.handler.GetCoilsAndOffset()
	registers, registerOffset := m.handler.GetRegistersAndOffset()

	var metrics []telegraf.Metric
	metricFields := make(map[string]interface{})

	for _, entry := range m.Metrics {
		for _, field := range entry.MetricSchema {
			var err error
			metricFields[field.Name], err = modbus_server.ParseMemory(
				m.ByteOrder, modbus_server.MemoryEntry{
					Address:  field.Address,
					Type:     field.Type,
					Register: field.Register,
					Scale:    field.Scale,
					Bit:      field.Bit,
					Length:   field.Length,
				}, coilOffset, registerOffset, coils, registers,
			)

			if err != nil {
				m.Log.Errorf("Error parsing memory: %v", err)
				continue
			}
			metrics = append(metrics, metric.New(entry.Name, entry.Tags, metricFields, timestamp))
		}
	}
	return metrics
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

	// Initialize the handler
	m.handler, err = modbus_server.NewRequestHandler(uint16(len(coils)), coilOffset, uint16(len(registers)), registerOffset, m.Log)
	if err != nil {
		m.Log.Errorf("failed to create server: %v\n", err)
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
	// Create a cancellable context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	return nil
}

func (*ModbusServer) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (m *ModbusServer) Start(acc telegraf.Accumulator) error {
	err := m.server.Start()
	if err != nil {
		m.Log.Errorf("Error starting server: %v", err)
		return err
	}
	m.Log.Debug("Server started")
	go func() {
		for {
			select {
			case <-m.ctx.Done():
				return
				// Check if the channel is empty
			case lastEditTimestamp := <-m.handler.LastEdit:
				metrics := m.getMetrics(lastEditTimestamp)
				m.Log.Infof("Gathered %d metrics", len(metrics))
				for _, modbusMetric := range metrics {
					acc.AddMetric(modbusMetric)
				}
			}
		}
	}()
	return nil
}

func (m *ModbusServer) Stop() {
	err := m.server.Stop()
	if err != nil {
		m.Log.Errorf("Error stopping server: %v", err)
		return
	}
	m.cancel()
	close(m.handler.LastEdit)
	m.Log.Debug("Server stopped")
}

func init() {
	inputs.Add(
		"modbus_server", func() telegraf.Input {
			return &ModbusServer{}
		},
	)
}
