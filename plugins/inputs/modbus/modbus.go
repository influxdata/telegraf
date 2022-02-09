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

type ModbusWorkarounds struct {
	PollPause        config.Duration `toml:"pause_between_requests"`
	CloseAfterGather bool            `toml:"close_connection_after_gather"`
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

const description = `Retrieve data from MODBUS slave devices`
const sampleConfigStart = `
  ## Connection Configuration
  ##
  ## The plugin supports connections to PLCs via MODBUS/TCP, RTU over TCP, ASCII over TCP or
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

  ## Trace the connection to the modbus device as debug messages
  ## Note: You have to enable telegraf's debug mode to see those messages!
  # debug_connection = false

  ## For Modbus over TCP you can choose between "TCP", "RTUoverTCP" and "ASCIIoverTCP"
  ## default behaviour is "TCP" if the controller is TCP
  ## For Serial you can choose between "RTU" and "ASCII"
  # transmission_mode = "RTU"

	## Define the configuration schema
  ##  |---register -- define fields per register type in the original style (only supports one slave ID)
  ##  |---request  -- define fields on a requests base
  configuration_type = "register"
`
const sampleConfigEnd = `
  ## Enable workarounds required by some devices to work correctly
  # [inputs.modbus.workarounds]
    ## Pause between read requests sent to the device. This might be necessary for (slow) serial devices.
    # pause_between_requests = "0ms"
    ## Close the connection after every gather cycle. Usually the plugin closes the connection after a certain
    ## idle-timeout, however, if you query a device with limited simultaneous connectivity (e.g. serial devices)
    ## from multiple instances you might want to only stay connected during gather and disconnect afterwards.
    # close_connection_after_gather = false
`

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

	timestamp := time.Now()
	for retry := 0; retry <= m.Retries; retry++ {
		timestamp = time.Now()
		if err := m.gatherFields(); err != nil {
			if mberr, ok := err.(*mb.Error); ok && mberr.ExceptionCode == mb.ExceptionCodeServerDeviceBusy && retry < m.Retries {
				m.Log.Infof("Device busy! Retrying %d more time(s)...", m.Retries-retry)
				time.Sleep(time.Duration(m.RetriesWaitTime))
				continue
			}
			// Show the disconnect error this way to not shadow the initial error
			if discerr := m.disconnect(); discerr != nil {
				m.Log.Errorf("Disconnecting failed: %v", discerr)
			}
			return err
		}
		// Reading was successful, leave the retry loop
		break
	}

	for slaveID, requests := range m.requests {
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
	return err
}

func (m *Modbus) disconnect() error {
	err := m.handler.Close()
	m.isConnected = false
	return err
}

func (m *Modbus) gatherFields() error {
	for slaveID, requests := range m.requests {
		m.handler.SetSlave(slaveID)
		if err := m.gatherRequestsCoil(requests.coil); err != nil {
			return err
		}
		if err := m.gatherRequestsDiscrete(requests.discrete); err != nil {
			return err
		}
		if err := m.gatherRequestsHolding(requests.holding); err != nil {
			return err
		}
		if err := m.gatherRequestsInput(requests.input); err != nil {
			return err
		}
	}

	return nil
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
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })
}
