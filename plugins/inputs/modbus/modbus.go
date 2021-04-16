package modbus

import (
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

// Modbus holds all data relevant to the plugin
type Modbus struct {
	Name             string          `toml:"name"`
	Controller       string          `toml:"controller"`
	TransmissionMode string          `toml:"transmission_mode"`
	BaudRate         int             `toml:"baud_rate"`
	DataBits         int             `toml:"data_bits"`
	Parity           string          `toml:"parity"`
	StopBits         int             `toml:"stop_bits"`
	Timeout          config.Duration `toml:"timeout"`
	Retries          int             `toml:"busy_retries"`
	RetriesWaitTime  config.Duration `toml:"busy_retries_wait"`
	Log              telegraf.Logger `toml:"-"`
	// Register configuration
	ConfigurationOriginal
	// Connection handling
	client      mb.Client
	handler     mb.ClientHandler
	isConnected bool
	// Request handling
	requests []request
}

type fieldConverterFunc func(bytes []byte) interface{}

type request struct {
	slaveID      byte
	registerType string
	address      uint16
	length       uint16
	fields       []field
}

type field struct {
	measurement string
	name        string
	scale       float64
	address     uint16
	length      uint16
	converter   fieldConverterFunc
	value       interface{}
}

const (
	cDiscreteInputs   = "discrete_input"
	cCoils            = "coil"
	cHoldingRegisters = "holding_register"
	cInputRegisters   = "input_register"
)

const description = `Retrieve data from MODBUS slave devices`
const sampleConfig = `
  ## Connection Configuration
  ##
  ## The plugin supports connections to PLCs via MODBUS/TCP or
  ## via serial line communication in binary (RTU) or readable (ASCII) encoding
  ##
  ## Device name
  name = "Device"

  ## Slave ID - addresses a MODBUS device on the bus
  ## Range: 0 - 255 [0 = broadcast; 248 - 255 = reserved]
  slave_id = 1

  ## Timeout for each request
  timeout = "1s"

  ## Maximum number of retries and the time to wait between retries
  ## when a slave-device is busy.
  # busy_retries = 0
  # busy_retries_wait = "100ms"

  # TCP - connect via Modbus/TCP
  controller = "tcp://localhost:502"

  ## Serial (RS485; RS232)
  # controller = "file:///dev/ttyUSB0"
  # baud_rate = 9600
  # data_bits = 8
  # parity = "N"
  # stop_bits = 1
  # transmission_mode = "RTU"


  ## Measurements
  ##

  ## Digital Variables, Discrete Inputs and Coils
  ## measurement - the (optional) measurement name, defaults to "modbus"
  ## name        - the variable name
  ## address     - variable address

  discrete_inputs = [
    { name = "start",          address = [0]},
    { name = "stop",           address = [1]},
    { name = "reset",          address = [2]},
    { name = "emergency_stop", address = [3]},
  ]
  coils = [
    { name = "motor1_run",     address = [0]},
    { name = "motor1_jog",     address = [1]},
    { name = "motor1_stop",    address = [2]},
  ]

  ## Analog Variables, Input Registers and Holding Registers
  ## measurement - the (optional) measurement name, defaults to "modbus"
  ## name        - the variable name
  ## byte_order  - the ordering of bytes
  ##  |---AB, ABCD   - Big Endian
  ##  |---BA, DCBA   - Little Endian
  ##  |---BADC       - Mid-Big Endian
  ##  |---CDAB       - Mid-Little Endian
  ## data_type  - INT16, UINT16, INT32, UINT32, INT64, UINT64,
  ##              FLOAT32-IEEE, FLOAT64-IEEE (the IEEE 754 binary representation)
  ##              FLOAT32, FIXED, UFIXED (fixed-point representation on input)
  ## scale      - the final numeric variable representation
  ## address    - variable address

  holding_registers = [
    { name = "power_factor", byte_order = "AB",   data_type = "FIXED", scale=0.01,  address = [8]},
    { name = "voltage",      byte_order = "AB",   data_type = "FIXED", scale=0.1,   address = [0]},
    { name = "energy",       byte_order = "ABCD", data_type = "FIXED", scale=0.001, address = [5,6]},
    { name = "current",      byte_order = "ABCD", data_type = "FIXED", scale=0.001, address = [1,2]},
    { name = "frequency",    byte_order = "AB",   data_type = "UFIXED", scale=0.1,  address = [7]},
    { name = "power",        byte_order = "ABCD", data_type = "UFIXED", scale=0.1,  address = [3,4]},
  ]
  input_registers = [
    { name = "tank_level",   byte_order = "AB",   data_type = "INT16",   scale=1.0,     address = [0]},
    { name = "tank_ph",      byte_order = "AB",   data_type = "INT16",   scale=1.0,     address = [1]},
    { name = "pump1_speed",  byte_order = "ABCD", data_type = "INT32",   scale=1.0,     address = [3,4]},
  ]
`

// SampleConfig returns a basic configuration for the plugin
func (m *Modbus) SampleConfig() string {
	return sampleConfig
}

// Description returns a short description of what the plugin does
func (m *Modbus) Description() string {
	return description
}

func (m *Modbus) Init() error {
	//check device name
	if m.Name == "" {
		return fmt.Errorf("device name is empty")
	}

	if m.Retries < 0 {
		return fmt.Errorf("retries cannot be negative")
	}

	// Check and process the configuration
	if err := m.ConfigurationOriginal.Check(); err != nil {
		return fmt.Errorf("original configuraton invalid: %v", err)
	}

	r, err := m.ConfigurationOriginal.Process()
	if err != nil {
		return fmt.Errorf("cannot process original configuraton: %v", err)
	}
	m.requests = append(m.requests, r...)

	// Setup client
	return m.initClient()
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
		handler := mb.NewTCPClientHandler(host + ":" + port)
		handler.Timeout = time.Duration(m.Timeout)
		m.handler = handler
	case "file":
		switch m.TransmissionMode {
		case "RTU":
			handler := mb.NewRTUClientHandler(u.Path)
			handler.Timeout = time.Duration(m.Timeout)
			handler.BaudRate = m.BaudRate
			handler.DataBits = m.DataBits
			handler.Parity = m.Parity
			handler.StopBits = m.StopBits
			m.handler = handler
		case "ASCII":
			handler := mb.NewASCIIClientHandler(u.Path)
			handler.Timeout = time.Duration(m.Timeout)
			handler.BaudRate = m.BaudRate
			handler.DataBits = m.DataBits
			handler.Parity = m.Parity
			handler.StopBits = m.StopBits
			m.handler = handler
		default:
			return fmt.Errorf("invalid protocol '%s' - '%s' ", u.Scheme, m.TransmissionMode)
		}
	default:
		return fmt.Errorf("invalid controller %q", m.Controller)
	}

	m.handler.SetSlave(m.SlaveID)
	m.client = mb.NewClient(m.handler)
	m.isConnected = false

	return nil
}

// Connect to a MODBUS Slave device via Modbus/[TCP|RTU|ASCII]
func (m *Modbus) connect() error {
	err := m.handler.Connect()
	m.isConnected = err == nil
	return err
}

func (m *Modbus) disconnect() error {
	err := m.handler.Close()
	m.isConnected = false
	return err
}

func (m *Modbus) readRegisterValues(registerType string, address, length uint16) ([]byte, error) {
	switch registerType {
	case cDiscreteInputs:
		return m.client.ReadDiscreteInputs(address, length)
	case cCoils:
		return m.client.ReadCoils(address, length)
	case cInputRegisters:
		return m.client.ReadInputRegisters(address, length)
	case cHoldingRegisters:
		return m.client.ReadHoldingRegisters(address, length)
	}
	return nil, fmt.Errorf("invalid register type %q", registerType)
}

func (m *Modbus) getFields() error {
	for _, request := range m.requests {
		bytes, err := m.readRegisterValues(request.registerType, request.address, request.length)
		if err != nil {
			return err
		}

		switch request.registerType {
		case cDiscreteInputs, cCoils:
			// Bit value handling
			for i, field := range request.fields {
				offset := field.address - request.address
				idx := offset / 8
				bit := offset % 8

				request.fields[i].value = uint16((bytes[idx] >> bit) & 0x01)
			}
		case cInputRegisters, cHoldingRegisters:
			// Non-bit value handling
			for i, field := range request.fields {
				// Determine the offset of the field values in the read array
				offset := 2 * (field.address - request.address) // registers are 16bit = 2 byte
				length := 2 * field.length                      // field length is in registers a 16bit

				// Convert the actual value
				request.fields[i].value = field.converter(bytes[offset : offset+length])
			}
		}
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

	timestamp := time.Now()
	for retry := 0; retry <= m.Retries; retry++ {
		timestamp = time.Now()
		err := m.getFields()
		if err != nil {
			mberr, ok := err.(*mb.Error)
			if ok && mberr.ExceptionCode == mb.ExceptionCodeServerDeviceBusy && retry < m.Retries {
				m.Log.Infof("Device busy! Retrying %d more time(s)...", m.Retries-retry)
				time.Sleep(time.Duration(m.RetriesWaitTime))
				continue
			}
			// Ignore return error to not shadow the initial error
			//nolint:errcheck,revive
			m.disconnect()
			return err
		}
		// Reading was successful, leave the retry loop
		break
	}

	grouper := metric.NewSeriesGrouper()
	for _, request := range m.requests {
		tags := map[string]string{
			"name":     m.Name,
			"type":     request.registerType,
			"slave_id": strconv.Itoa(int(request.slaveID)),
		}

		for _, field := range request.fields {
			// In case no measurement was specified we use "modbus" as default
			measurement := "modbus"
			if field.measurement != "" {
				measurement = field.measurement
			}

			// Group the data by series
			if err := grouper.Add(measurement, tags, timestamp, field.name, field.value); err != nil {
				return err
			}
		}

		// Add the metrics grouped by series to the accumulator
		for _, x := range grouper.Metrics() {
			acc.AddMetric(x)
		}
	}

	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })
}
