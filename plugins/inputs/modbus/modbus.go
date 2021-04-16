package modbus

import (
	"fmt"
	"net"
	"net/url"
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
	client       mb.Client
	isConnected  bool
	tcpHandler   *mb.TCPClientHandler
	rtuHandler   *mb.RTUClientHandler
	asciiHandler *mb.ASCIIClientHandler
	// Request handling
	requests []request
}

type fieldConverterFunc func(bytes []byte) interface{}

type request struct {
	SlaveID        int
	Type           string
	RegistersRange []registerRange
	Fields         []field
}

type field struct {
	Measurement string
	Name        string
	Scale       float64
	Address     []uint16
	converter   fieldConverterFunc
	value       interface{}
}

type registerRange struct {
	address uint16
	length  uint16
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

	return m.ConfigurationOriginal.Process(m)
}

// Connect to a MODBUS Slave device via Modbus/[TCP|RTU|ASCII]
func connect(m *Modbus) error {
	u, err := url.Parse(m.Controller)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp":
		var host, port string
		host, port, err = net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		m.tcpHandler = mb.NewTCPClientHandler(host + ":" + port)
		m.tcpHandler.Timeout = time.Duration(m.Timeout)
		m.tcpHandler.SlaveID = byte(m.SlaveID)
		m.client = mb.NewClient(m.tcpHandler)
		err := m.tcpHandler.Connect()
		if err != nil {
			return err
		}
		m.isConnected = true
		return nil
	case "file":
		if m.TransmissionMode == "RTU" {
			m.rtuHandler = mb.NewRTUClientHandler(u.Path)
			m.rtuHandler.Timeout = time.Duration(m.Timeout)
			m.rtuHandler.SlaveID = byte(m.SlaveID)
			m.rtuHandler.BaudRate = m.BaudRate
			m.rtuHandler.DataBits = m.DataBits
			m.rtuHandler.Parity = m.Parity
			m.rtuHandler.StopBits = m.StopBits
			m.client = mb.NewClient(m.rtuHandler)
			err := m.rtuHandler.Connect()
			if err != nil {
				return err
			}
			m.isConnected = true
			return nil
		} else if m.TransmissionMode == "ASCII" {
			m.asciiHandler = mb.NewASCIIClientHandler(u.Path)
			m.asciiHandler.Timeout = time.Duration(m.Timeout)
			m.asciiHandler.SlaveID = byte(m.SlaveID)
			m.asciiHandler.BaudRate = m.BaudRate
			m.asciiHandler.DataBits = m.DataBits
			m.asciiHandler.Parity = m.Parity
			m.asciiHandler.StopBits = m.StopBits
			m.client = mb.NewClient(m.asciiHandler)
			err := m.asciiHandler.Connect()
			if err != nil {
				return err
			}
			m.isConnected = true
			return nil
		} else {
			return fmt.Errorf("invalid protocol '%s' - '%s' ", u.Scheme, m.TransmissionMode)
		}
	default:
		return fmt.Errorf("invalid controller")
	}
}

func disconnect(m *Modbus) error {
	u, err := url.Parse(m.Controller)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp":
		m.tcpHandler.Close()
		return nil
	case "file":
		if m.TransmissionMode == "RTU" {
			m.rtuHandler.Close()
			return nil
		} else if m.TransmissionMode == "ASCII" {
			m.asciiHandler.Close()
			return nil
		} else {
			return fmt.Errorf("invalid protocol '%s' - '%s' ", u.Scheme, m.TransmissionMode)
		}
	default:
		return fmt.Errorf("invalid controller")
	}
}

func readRegisterValues(m *Modbus, rt string, rr registerRange) ([]byte, error) {
	if rt == cDiscreteInputs {
		return m.client.ReadDiscreteInputs(rr.address, rr.length)
	} else if rt == cCoils {
		return m.client.ReadCoils(rr.address, rr.length)
	} else if rt == cInputRegisters {
		return m.client.ReadInputRegisters(rr.address, rr.length)
	} else if rt == cHoldingRegisters {
		return m.client.ReadHoldingRegisters(rr.address, rr.length)
	} else {
		return []byte{}, fmt.Errorf("not Valid function")
	}
}

func (m *Modbus) getFields() error {
	for _, request := range m.requests {
		rawValues := make(map[uint16][]byte)
		bitRawValues := make(map[uint16]uint16)
		for _, rr := range request.RegistersRange {
			address := rr.address
			readValues, err := readRegisterValues(m, request.Type, rr)
			if err != nil {
				return err
			}

			// Raw Values
			if request.Type == cDiscreteInputs || request.Type == cCoils {
				for _, readValue := range readValues {
					for bitPosition := uint(0); bitPosition < 8; bitPosition++ {
						bitRawValues[address] = getBitValue(readValue, bitPosition)
						address = address + 1
						if address > rr.address+rr.length {
							break
						}
					}
				}
			}

			// Raw Values
			if request.Type == cInputRegisters || request.Type == cHoldingRegisters {
				batchSize := 2
				for batchSize < len(readValues) {
					rawValues[address] = readValues[0:batchSize:batchSize]
					address = address + 1
					readValues = readValues[batchSize:]
				}

				rawValues[address] = readValues[0:batchSize:batchSize]
			}
		}

		if request.Type == cDiscreteInputs || request.Type == cCoils {
			for i := 0; i < len(request.Fields); i++ {
				request.Fields[i].value = bitRawValues[request.Fields[i].Address[0]]
			}
		}

		if request.Type == cInputRegisters || request.Type == cHoldingRegisters {
			for i := 0; i < len(request.Fields); i++ {
				var buf []byte

				for j := 0; j < len(request.Fields[i].Address); j++ {
					tempArray := rawValues[request.Fields[i].Address[j]]
					for x := 0; x < len(tempArray); x++ {
						buf = append(buf, tempArray[x])
					}
				}

				request.Fields[i].value = request.Fields[i].converter(buf)
			}
		}
	}

	return nil
}

// Gather implements the telegraf plugin interface method for data accumulation
func (m *Modbus) Gather(acc telegraf.Accumulator) error {
	if !m.isConnected {
		err := connect(m)
		if err != nil {
			m.isConnected = false
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
			disconnect(m)
			m.isConnected = false
			return err
		}
		// Reading was successful, leave the retry loop
		break
	}

	grouper := metric.NewSeriesGrouper()
	for _, reg := range m.requests {
		tags := map[string]string{
			"name": m.Name,
			"type": reg.Type,
		}

		for _, field := range reg.Fields {
			// In case no measurement was specified we use "modbus" as default
			measurement := "modbus"
			if field.Measurement != "" {
				measurement = field.Measurement
			}

			// Group the data by series
			if err := grouper.Add(measurement, tags, timestamp, field.Name, field.value); err != nil {
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
