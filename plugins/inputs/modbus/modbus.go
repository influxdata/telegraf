//go:generate ../../../tools/readme_config_includer/generator
package modbus

import (
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	mb "github.com/grid-x/modbus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample_general_begin.conf
var sampleConfigStart string

//go:embed sample_general_end.conf
var sampleConfigEnd string

type ModbusWorkarounds struct {
	AfterConnectPause config.Duration `toml:"pause_after_connect"`
	PollPause         config.Duration `toml:"pause_between_requests"`
	CloseAfterGather  bool            `toml:"close_connection_after_gather"`
}

// Modbus holds all data relevant to the plugin
type Modbus struct {
	Name             string            `toml:"name"`
	Controller       string            `toml:"controller"`
	TransmissionMode string            `toml:"transmission_mode"`
	BaudRate         int               `toml:"baud_rate"`
	DataBits         int               `toml:"data_bits"`
	Parity           string            `toml:"parity"`
	StopBits         int               `toml:"stop_bits"`
	Timeout          config.Duration   `toml:"timeout"`
	Retries          int               `toml:"busy_retries"`
	RetriesWaitTime  config.Duration   `toml:"busy_retries_wait"`
	DebugConnection  bool              `toml:"debug_connection"`
	Workarounds      ModbusWorkarounds `toml:"workarounds"`
	Log              telegraf.Logger   `toml:"-"`
	// Register configuration
	ConfigurationType string `toml:"configuration_type"`
	ConfigurationOriginal
	ConfigurationPerRequest

	// Connection handling
	client      mb.Client
	handler     mb.ClientHandler
	isConnected bool
	// Request handling
	requests map[byte]requestSet
}

type fieldConverterFunc func(bytes []byte) interface{}

type requestSet struct {
	coil     []request
	discrete []request
	holding  []request
	input    []request
}

type field struct {
	measurement string
	name        string
	address     uint16
	length      uint16
	omit        bool
	converter   fieldConverterFunc
	value       interface{}
}

const (
	cDiscreteInputs   = "discrete_input"
	cCoils            = "coil"
	cHoldingRegisters = "holding_register"
	cInputRegisters   = "input_register"
)

// SampleConfig returns a basic configuration for the plugin
func (m *Modbus) SampleConfig() string {
	configs := []Configuration{}
	cfgOriginal := m.ConfigurationOriginal
	cfgPerRequest := m.ConfigurationPerRequest
	configs = append(configs, &cfgOriginal, &cfgPerRequest)

	totalConfig := sampleConfigStart
	for _, c := range configs {
		totalConfig += c.SampleConfigPart() + "\n"
	}
	totalConfig += "\n"
	totalConfig += sampleConfigEnd
	return totalConfig
}

func (m *Modbus) Init() error {
	//check device name
	if m.Name == "" {
		return fmt.Errorf("device name is empty")
	}

	if m.Retries < 0 {
		return fmt.Errorf("retries cannot be negative")
	}

	// Determine the configuration style
	var cfg Configuration
	switch m.ConfigurationType {
	case "", "register":
		cfg = &m.ConfigurationOriginal
	case "request":
		cfg = &m.ConfigurationPerRequest
	default:
		return fmt.Errorf("unknown configuration type %q", m.ConfigurationType)
	}

	// Check and process the configuration
	if err := cfg.Check(); err != nil {
		return fmt.Errorf("configuraton invalid: %v", err)
	}

	r, err := cfg.Process()
	if err != nil {
		return fmt.Errorf("cannot process configuraton: %v", err)
	}
	m.requests = r

	// Setup client
	if err := m.initClient(); err != nil {
		return fmt.Errorf("initializing client failed: %v", err)
	}

	return nil
}

// Gather implements the telegraf plugin interface method for data accumulation
func (m *Modbus) Gather(acc telegraf.Accumulator) error {
	if !m.isConnected {
		if err := m.connect(); err != nil {
			return err
		}
	}

	for slaveID, requests := range m.requests {
		m.Log.Debugf("Reading slave %d for %s...", slaveID, m.Controller)
		if err := m.readSlaveData(slaveID, requests); err != nil {
			acc.AddError(fmt.Errorf("slave %d: %w", slaveID, err))
			mberr, ok := err.(*mb.Error)
			if !ok || mberr.ExceptionCode != mb.ExceptionCodeServerDeviceBusy {
				m.Log.Debugf("Reconnecting to %s...", m.Controller)
				if err := m.disconnect(); err != nil {
					return fmt.Errorf("disconnecting failed: %w", err)
				}
				if err := m.connect(); err != nil {
					return fmt.Errorf("slave %d: connecting failed: %w", slaveID, err)
				}
			}
			continue
		}
		timestamp := time.Now()

		tags := map[string]string{
			"name":     m.Name,
			"type":     cCoils,
			"slave_id": strconv.Itoa(int(slaveID)),
		}
		m.collectFields(acc, timestamp, tags, requests.coil)

		tags["type"] = cDiscreteInputs
		m.collectFields(acc, timestamp, tags, requests.discrete)

		tags["type"] = cHoldingRegisters
		m.collectFields(acc, timestamp, tags, requests.holding)

		tags["type"] = cInputRegisters
		m.collectFields(acc, timestamp, tags, requests.input)
	}

	// Disconnect after read if configured
	if m.Workarounds.CloseAfterGather {
		return m.disconnect()
	}

	return nil
}

func (m *Modbus) initClient() error {
	u, err := url.Parse(m.Controller)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp":
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		switch m.TransmissionMode {
		case "RTUoverTCP":
			handler := mb.NewRTUOverTCPClientHandler(host + ":" + port)
			handler.Timeout = time.Duration(m.Timeout)
			if m.DebugConnection {
				handler.Logger = m
			}
			m.handler = handler
		case "ASCIIoverTCP":
			handler := mb.NewASCIIOverTCPClientHandler(host + ":" + port)
			handler.Timeout = time.Duration(m.Timeout)
			if m.DebugConnection {
				handler.Logger = m
			}
			m.handler = handler
		default:
			handler := mb.NewTCPClientHandler(host + ":" + port)
			handler.Timeout = time.Duration(m.Timeout)
			if m.DebugConnection {
				handler.Logger = m
			}
			m.handler = handler
		}
	case "file":
		switch m.TransmissionMode {
		case "RTU":
			handler := mb.NewRTUClientHandler(u.Path)
			handler.Timeout = time.Duration(m.Timeout)
			handler.BaudRate = m.BaudRate
			handler.DataBits = m.DataBits
			handler.Parity = m.Parity
			handler.StopBits = m.StopBits
			if m.DebugConnection {
				handler.Logger = m
			}
			m.handler = handler
		case "ASCII":
			handler := mb.NewASCIIClientHandler(u.Path)
			handler.Timeout = time.Duration(m.Timeout)
			handler.BaudRate = m.BaudRate
			handler.DataBits = m.DataBits
			handler.Parity = m.Parity
			handler.StopBits = m.StopBits
			if m.DebugConnection {
				handler.Logger = m
			}
			m.handler = handler
		default:
			return fmt.Errorf("invalid protocol '%s' - '%s' ", u.Scheme, m.TransmissionMode)
		}
	default:
		return fmt.Errorf("invalid controller %q", m.Controller)
	}

	m.client = mb.NewClient(m.handler)
	m.isConnected = false

	return nil
}

// Connect to a MODBUS Slave device via Modbus/[TCP|RTU|ASCII]
func (m *Modbus) connect() error {
	err := m.handler.Connect()
	m.isConnected = err == nil
	if m.isConnected && m.Workarounds.AfterConnectPause != 0 {
		nextRequest := time.Now().Add(time.Duration(m.Workarounds.AfterConnectPause))
		time.Sleep(time.Until(nextRequest))
	}
	return err
}

func (m *Modbus) disconnect() error {
	err := m.handler.Close()
	m.isConnected = false
	return err
}

func (m *Modbus) readSlaveData(slaveID byte, requests requestSet) error {
	m.handler.SetSlave(slaveID)

	for retry := 0; retry < m.Retries; retry++ {
		err := m.gatherFields(requests)
		if err == nil {
			// Reading was successful
			return nil
		}

		// Exit in case a non-recoverable error occurred
		mberr, ok := err.(*mb.Error)
		if !ok || mberr.ExceptionCode != mb.ExceptionCodeServerDeviceBusy {
			return err
		}

		// Wait some time and try again reading the slave.
		m.Log.Infof("Device busy! Retrying %d more time(s)...", m.Retries-retry)
		time.Sleep(time.Duration(m.RetriesWaitTime))
	}
	return m.gatherFields(requests)
}

func (m *Modbus) gatherFields(requests requestSet) error {
	if err := m.gatherRequestsCoil(requests.coil); err != nil {
		return err
	}
	if err := m.gatherRequestsDiscrete(requests.discrete); err != nil {
		return err
	}
	if err := m.gatherRequestsHolding(requests.holding); err != nil {
		return err
	}
	return m.gatherRequestsInput(requests.input)
}

func (m *Modbus) gatherRequestsCoil(requests []request) error {
	for _, request := range requests {
		m.Log.Debugf("trying to read coil@%v[%v]...", request.address, request.length)
		bytes, err := m.client.ReadCoils(request.address, request.length)
		if err != nil {
			return err
		}
		nextRequest := time.Now().Add(time.Duration(m.Workarounds.PollPause))
		m.Log.Debugf("got coil@%v[%v]: %v", request.address, request.length, bytes)

		// Bit value handling
		for i, field := range request.fields {
			offset := field.address - request.address
			idx := offset / 8
			bit := offset % 8

			request.fields[i].value = uint16((bytes[idx] >> bit) & 0x01)
			m.Log.Debugf("  field %s with bit %d @ byte %d: %v --> %v", field.name, bit, idx, (bytes[idx]>>bit)&0x01, request.fields[i].value)
		}

		// Some (serial) devices require a pause between requests...
		time.Sleep(time.Until(nextRequest))
	}
	return nil
}

func (m *Modbus) gatherRequestsDiscrete(requests []request) error {
	for _, request := range requests {
		m.Log.Debugf("trying to read discrete@%v[%v]...", request.address, request.length)
		bytes, err := m.client.ReadDiscreteInputs(request.address, request.length)
		if err != nil {
			return err
		}
		nextRequest := time.Now().Add(time.Duration(m.Workarounds.PollPause))
		m.Log.Debugf("got discrete@%v[%v]: %v", request.address, request.length, bytes)

		// Bit value handling
		for i, field := range request.fields {
			offset := field.address - request.address
			idx := offset / 8
			bit := offset % 8

			request.fields[i].value = uint16((bytes[idx] >> bit) & 0x01)
			m.Log.Debugf("  field %s with bit %d @ byte %d: %v --> %v", field.name, bit, idx, (bytes[idx]>>bit)&0x01, request.fields[i].value)
		}

		// Some (serial) devices require a pause between requests...
		time.Sleep(time.Until(nextRequest))
	}
	return nil
}

func (m *Modbus) gatherRequestsHolding(requests []request) error {
	for _, request := range requests {
		m.Log.Debugf("trying to read holding@%v[%v]...", request.address, request.length)
		bytes, err := m.client.ReadHoldingRegisters(request.address, request.length)
		if err != nil {
			return err
		}
		nextRequest := time.Now().Add(time.Duration(m.Workarounds.PollPause))
		m.Log.Debugf("got holding@%v[%v]: %v", request.address, request.length, bytes)

		// Non-bit value handling
		for i, field := range request.fields {
			// Determine the offset of the field values in the read array
			offset := 2 * (field.address - request.address) // registers are 16bit = 2 byte
			length := 2 * field.length                      // field length is in registers a 16bit

			// Convert the actual value
			request.fields[i].value = field.converter(bytes[offset : offset+length])
			m.Log.Debugf("  field %s with offset %d with len %d: %v --> %v", field.name, offset, length, bytes[offset:offset+length], request.fields[i].value)
		}

		// Some (serial) devices require a pause between requests...
		time.Sleep(time.Until(nextRequest))
	}
	return nil
}

func (m *Modbus) gatherRequestsInput(requests []request) error {
	for _, request := range requests {
		m.Log.Debugf("trying to read input@%v[%v]...", request.address, request.length)
		bytes, err := m.client.ReadInputRegisters(request.address, request.length)
		if err != nil {
			return err
		}
		nextRequest := time.Now().Add(time.Duration(m.Workarounds.PollPause))
		m.Log.Debugf("got input@%v[%v]: %v", request.address, request.length, bytes)

		// Non-bit value handling
		for i, field := range request.fields {
			// Determine the offset of the field values in the read array
			offset := 2 * (field.address - request.address) // registers are 16bit = 2 byte
			length := 2 * field.length                      // field length is in registers a 16bit

			// Convert the actual value
			request.fields[i].value = field.converter(bytes[offset : offset+length])
			m.Log.Debugf("  field %s with offset %d with len %d: %v --> %v", field.name, offset, length, bytes[offset:offset+length], request.fields[i].value)
		}

		// Some (serial) devices require a pause between requests...
		time.Sleep(time.Until(nextRequest))
	}
	return nil
}

func (m *Modbus) collectFields(acc telegraf.Accumulator, timestamp time.Time, tags map[string]string, requests []request) {
	grouper := metric.NewSeriesGrouper()
	for _, request := range requests {
		// Collect tags from global and per-request
		rtags := map[string]string{}
		for k, v := range tags {
			rtags[k] = v
		}
		for k, v := range request.tags {
			rtags[k] = v
		}

		for _, field := range request.fields {
			// In case no measurement was specified we use "modbus" as default
			measurement := "modbus"
			if field.measurement != "" {
				measurement = field.measurement
			}

			// Group the data by series
			if err := grouper.Add(measurement, rtags, timestamp, field.name, field.value); err != nil {
				acc.AddError(fmt.Errorf("cannot add field %q for measurement %q: %v", field.name, measurement, err))
				continue
			}
		}
	}

	// Add the metrics grouped by series to the accumulator
	for _, x := range grouper.Metrics() {
		acc.AddMetric(x)
	}
}

// Implement the logger interface of the modbus client
func (m *Modbus) Printf(format string, v ...interface{}) {
	m.Log.Debugf(format, v...)
}

// Add this plugin to telegraf
func init() {
	inputs.Add("modbus", func() telegraf.Input {
		return &Modbus{
			Name:              "Device",
			Controller:        "tcp://localhost:502",
			TransmissionMode:  "RTU",
			Timeout:           1,
			BaudRate:          9600,
			DataBits:          8,
			Parity:            "N",
			StopBits:          1,
			Retries:           0,
			RetriesWaitTime:   100,
			DebugConnection:   false,
			ConfigurationType: "register",
		}
	})
}
