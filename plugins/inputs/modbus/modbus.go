package modbus

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"net/url"
	"sort"
	"time"

	mb "github.com/goburrow/modbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Modbus holds all data relevant to the plugin
type Modbus struct {
	Name             string            `toml:"name"`
	Controller       string            `toml:"controller"`
	TransmissionMode string            `toml:"transmission_mode"`
	BaudRate         int               `toml:"baud_rate"`
	DataBits         int               `toml:"data_bits"`
	Parity           string            `toml:"parity"`
	StopBits         int               `toml:"stop_bits"`
	SlaveID          int               `toml:"slave_id"`
	Timeout          internal.Duration `toml:"timeout"`
	Retries          int               `toml:"busy_retries"`
	RetriesWaitTime  internal.Duration `toml:"busy_retries_wait"`
	DiscreteInputs   []fieldContainer  `toml:"discrete_inputs"`
	Coils            []fieldContainer  `toml:"coils"`
	HoldingRegisters []fieldContainer  `toml:"holding_registers"`
	InputRegisters   []fieldContainer  `toml:"input_registers"`
	registers        []register
	isConnected      bool
	tcpHandler       *mb.TCPClientHandler
	rtuHandler       *mb.RTUClientHandler
	asciiHandler     *mb.ASCIIClientHandler
	client           mb.Client
}

type register struct {
	Type           string
	RegistersRange []registerRange
	Fields         []fieldContainer
}

type fieldContainer struct {
	Measurement string   `toml:"measurement"`
	Name        string   `toml:"name"`
	ByteOrder   string   `toml:"byte_order"`
	DataType    string   `toml:"data_type"`
	Scale       float64  `toml:"scale"`
	Address     []uint16 `toml:"address"`
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
  ## data_type  - INT16, UINT16, INT32, UINT32, INT64, UINT64, FLOAT32, FLOAT32-IEEE (the IEEE 754 binary representation)
  ## scale      - the final numeric variable representation
  ## address    - variable address

  holding_registers = [
    { name = "power_factor", byte_order = "AB",   data_type = "FLOAT32", scale=0.01,  address = [8]},
    { name = "voltage",      byte_order = "AB",   data_type = "FLOAT32", scale=0.1,   address = [0]},
    { name = "energy",       byte_order = "ABCD", data_type = "FLOAT32", scale=0.001, address = [5,6]},
    { name = "current",      byte_order = "ABCD", data_type = "FLOAT32", scale=0.001, address = [1,2]},
    { name = "frequency",    byte_order = "AB",   data_type = "FLOAT32", scale=0.1,   address = [7]},
    { name = "power",        byte_order = "ABCD", data_type = "FLOAT32", scale=0.1,   address = [3,4]},
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

	err := m.InitRegister(m.DiscreteInputs, cDiscreteInputs)
	if err != nil {
		return err
	}

	err = m.InitRegister(m.Coils, cCoils)
	if err != nil {
		return err
	}

	err = m.InitRegister(m.HoldingRegisters, cHoldingRegisters)
	if err != nil {
		return err
	}

	err = m.InitRegister(m.InputRegisters, cInputRegisters)
	if err != nil {
		return err
	}

	return nil
}

func (m *Modbus) InitRegister(fields []fieldContainer, name string) error {
	if len(fields) == 0 {
		return nil
	}

	err := validateFieldContainers(fields, name)
	if err != nil {
		return err
	}

	addrs := []uint16{}
	for _, field := range fields {
		for _, a := range field.Address {
			addrs = append(addrs, a)
		}
	}

	addrs = removeDuplicates(addrs)
	sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

	ii := 0
	var registersRange []registerRange

	// Get range of consecutive integers
	// [1, 2, 3, 5, 6, 10, 11, 12, 14]
	// (1, 3) , (5, 2) , (10, 3), (14 , 1)
	for range addrs {
		if ii < len(addrs) {
			start := addrs[ii]
			end := start

			for ii < len(addrs)-1 && addrs[ii+1]-addrs[ii] == 1 {
				end = addrs[ii+1]
				ii++
			}
			ii++
			registersRange = append(registersRange, registerRange{start, end - start + 1})
		}
	}

	m.registers = append(m.registers, register{name, registersRange, fields})

	return nil
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
		m.tcpHandler.Timeout = m.Timeout.Duration
		m.tcpHandler.SlaveId = byte(m.SlaveID)
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
			m.rtuHandler.Timeout = m.Timeout.Duration
			m.rtuHandler.SlaveId = byte(m.SlaveID)
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
			m.asciiHandler.Timeout = m.Timeout.Duration
			m.asciiHandler.SlaveId = byte(m.SlaveID)
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

func validateFieldContainers(t []fieldContainer, n string) error {
	nameEncountered := map[string]bool{}
	for _, item := range t {
		//check empty name
		if item.Name == "" {
			return fmt.Errorf("empty name in '%s'", n)
		}

		//search name duplicate
		canonical_name := item.Measurement + "." + item.Name
		if nameEncountered[canonical_name] {
			return fmt.Errorf("name '%s' is duplicated in measurement '%s' '%s' - '%s'", item.Name, item.Measurement, n, item.Name)
		} else {
			nameEncountered[canonical_name] = true
		}

		if n == cInputRegisters || n == cHoldingRegisters {
			// search byte order
			switch item.ByteOrder {
			case "AB", "BA", "ABCD", "CDAB", "BADC", "DCBA", "ABCDEFGH", "HGFEDCBA", "BADCFEHG", "GHEFCDAB":
				break
			default:
				return fmt.Errorf("invalid byte order '%s' in '%s' - '%s'", item.ByteOrder, n, item.Name)
			}

			// search data type
			switch item.DataType {
			case "UINT16", "INT16", "UINT32", "INT32", "UINT64", "INT64", "FLOAT32-IEEE", "FLOAT32":
				break
			default:
				return fmt.Errorf("invalid data type '%s' in '%s' - '%s'", item.DataType, n, item.Name)
			}

			// check scale
			if item.Scale == 0.0 {
				return fmt.Errorf("invalid scale '%f' in '%s' - '%s'", item.Scale, n, item.Name)
			}
		}

		// check address
		if len(item.Address) != 1 && len(item.Address) != 2 && len(item.Address) != 4 {
			return fmt.Errorf("invalid address '%v' length '%v' in '%s' - '%s'", item.Address, len(item.Address), n, item.Name)
		}

		if n == cInputRegisters || n == cHoldingRegisters {
			if 2*len(item.Address) != len(item.ByteOrder) {
				return fmt.Errorf("invalid byte order '%s' and address '%v'  in '%s' - '%s'", item.ByteOrder, item.Address, n, item.Name)
			}

			// search duplicated
			if len(item.Address) > len(removeDuplicates(item.Address)) {
				return fmt.Errorf("duplicate address '%v'  in '%s' - '%s'", item.Address, n, item.Name)
			}
		} else if len(item.Address) != 1 {
			return fmt.Errorf("invalid address'%v' length'%v' in '%s' - '%s'", item.Address, len(item.Address), n, item.Name)
		}
	}
	return nil
}

func removeDuplicates(elements []uint16) []uint16 {
	encountered := map[uint16]bool{}
	result := []uint16{}

	for v := range elements {
		if encountered[elements[v]] {
		} else {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func readRegisterValues(m *Modbus, rt string, rr registerRange) ([]byte, error) {
	if rt == cDiscreteInputs {
		return m.client.ReadDiscreteInputs(uint16(rr.address), uint16(rr.length))
	} else if rt == cCoils {
		return m.client.ReadCoils(uint16(rr.address), uint16(rr.length))
	} else if rt == cInputRegisters {
		return m.client.ReadInputRegisters(uint16(rr.address), uint16(rr.length))
	} else if rt == cHoldingRegisters {
		return m.client.ReadHoldingRegisters(uint16(rr.address), uint16(rr.length))
	} else {
		return []byte{}, fmt.Errorf("not Valid function")
	}
}

func (m *Modbus) getFields() error {
	for _, register := range m.registers {
		rawValues := make(map[uint16][]byte)
		bitRawValues := make(map[uint16]uint16)
		for _, rr := range register.RegistersRange {
			address := rr.address
			readValues, err := readRegisterValues(m, register.Type, rr)
			if err != nil {
				return err
			}

			// Raw Values
			if register.Type == cDiscreteInputs || register.Type == cCoils {
				for _, readValue := range readValues {
					for bitPosition := 0; bitPosition < 8; bitPosition++ {
						bitRawValues[address] = getBitValue(readValue, bitPosition)
						address = address + 1
						if address+1 > rr.length {
							break
						}
					}
				}
			}

			// Raw Values
			if register.Type == cInputRegisters || register.Type == cHoldingRegisters {
				batchSize := 2
				for batchSize < len(readValues) {
					rawValues[address] = readValues[0:batchSize:batchSize]
					address = address + 1
					readValues = readValues[batchSize:]
				}

				rawValues[address] = readValues[0:batchSize:batchSize]
			}
		}

		if register.Type == cDiscreteInputs || register.Type == cCoils {
			for i := 0; i < len(register.Fields); i++ {
				register.Fields[i].value = bitRawValues[register.Fields[i].Address[0]]
			}
		}

		if register.Type == cInputRegisters || register.Type == cHoldingRegisters {
			for i := 0; i < len(register.Fields); i++ {
				var values_t []byte

				for j := 0; j < len(register.Fields[i].Address); j++ {
					tempArray := rawValues[register.Fields[i].Address[j]]
					for x := 0; x < len(tempArray); x++ {
						values_t = append(values_t, tempArray[x])
					}
				}

				register.Fields[i].value = convertDataType(register.Fields[i], values_t)
			}

		}
	}

	return nil
}

func getBitValue(n byte, pos int) uint16 {
	return uint16(n >> uint(pos) & 0x01)
}

func convertDataType(t fieldContainer, bytes []byte) interface{} {
	switch t.DataType {
	case "UINT16":
		e16 := convertEndianness16(t.ByteOrder, bytes)
		return scaleUint16(t.Scale, e16)
	case "INT16":
		e16 := convertEndianness16(t.ByteOrder, bytes)
		f16 := int16(e16)
		return scaleInt16(t.Scale, f16)
	case "UINT32":
		e32 := convertEndianness32(t.ByteOrder, bytes)
		return scaleUint32(t.Scale, e32)
	case "INT32":
		e32 := convertEndianness32(t.ByteOrder, bytes)
		f32 := int32(e32)
		return scaleInt32(t.Scale, f32)
	case "UINT64":
		e64 := convertEndianness64(t.ByteOrder, bytes)
		f64 := format64(t.DataType, e64).(uint64)
		return scaleUint64(t.Scale, f64)
	case "INT64":
		e64 := convertEndianness64(t.ByteOrder, bytes)
		f64 := format64(t.DataType, e64).(int64)
		return scaleInt64(t.Scale, f64)
	case "FLOAT32-IEEE":
		e32 := convertEndianness32(t.ByteOrder, bytes)
		f32 := math.Float32frombits(e32)
		return scaleFloat32(t.Scale, f32)
	case "FLOAT32":
		if len(bytes) == 2 {
			e16 := convertEndianness16(t.ByteOrder, bytes)
			return scale16toFloat32(t.Scale, e16)
		} else if len(bytes) == 4 {
			e32 := convertEndianness32(t.ByteOrder, bytes)
			return scale32toFloat32(t.Scale, e32)
		} else {
			e64 := convertEndianness64(t.ByteOrder, bytes)
			return scale64toFloat32(t.Scale, e64)
		}
	default:
		return 0
	}
}

func convertEndianness16(o string, b []byte) uint16 {
	switch o {
	case "AB":
		return binary.BigEndian.Uint16(b)
	case "BA":
		return binary.LittleEndian.Uint16(b)
	default:
		return 0
	}
}

func convertEndianness32(o string, b []byte) uint32 {
	switch o {
	case "ABCD":
		return binary.BigEndian.Uint32(b)
	case "DCBA":
		return binary.LittleEndian.Uint32(b)
	case "BADC":
		return uint32(binary.LittleEndian.Uint16(b[0:]))<<16 | uint32(binary.LittleEndian.Uint16(b[2:]))
	case "CDAB":
		return uint32(binary.BigEndian.Uint16(b[2:]))<<16 | uint32(binary.BigEndian.Uint16(b[0:]))
	default:
		return 0
	}
}

func convertEndianness64(o string, b []byte) uint64 {
	switch o {
	case "ABCDEFGH":
		return binary.BigEndian.Uint64(b)
	case "HGFEDCBA":
		return binary.LittleEndian.Uint64(b)
	case "BADCFEHG":
		return uint64(binary.LittleEndian.Uint16(b[0:]))<<48 | uint64(binary.LittleEndian.Uint16(b[2:]))<<32 | uint64(binary.LittleEndian.Uint16(b[4:]))<<16 | uint64(binary.LittleEndian.Uint16(b[6:]))
	case "GHEFCDAB":
		return uint64(binary.BigEndian.Uint16(b[6:]))<<48 | uint64(binary.BigEndian.Uint16(b[4:]))<<32 | uint64(binary.BigEndian.Uint16(b[2:]))<<16 | uint64(binary.BigEndian.Uint16(b[0:]))
	default:
		return 0
	}
}

func format16(f string, r uint16) interface{} {
	switch f {
	case "UINT16":
		return r
	case "INT16":
		return int16(r)
	default:
		return r
	}
}

func format32(f string, r uint32) interface{} {
	switch f {
	case "UINT32":
		return r
	case "INT32":
		return int32(r)
	case "FLOAT32-IEEE":
		return math.Float32frombits(r)
	default:
		return r
	}
}

func format64(f string, r uint64) interface{} {
	switch f {
	case "UINT64":
		return r
	case "INT64":
		return int64(r)
	default:
		return r
	}
}

func scale16toFloat32(s float64, v uint16) float64 {
	return float64(v) * s
}

func scale32toFloat32(s float64, v uint32) float64 {
	return float64(float64(v) * float64(s))
}

func scale64toFloat32(s float64, v uint64) float64 {
	return float64(float64(v) * float64(s))
}

func scaleInt16(s float64, v int16) int16 {
	return int16(float64(v) * s)
}

func scaleUint16(s float64, v uint16) uint16 {
	return uint16(float64(v) * s)
}

func scaleUint32(s float64, v uint32) uint32 {
	return uint32(float64(v) * float64(s))
}

func scaleInt32(s float64, v int32) int32 {
	return int32(float64(v) * float64(s))
}

func scaleFloat32(s float64, v float32) float32 {
	return float32(float64(v) * s)
}

func scaleUint64(s float64, v uint64) uint64 {
	return uint64(float64(v) * float64(s))
}

func scaleInt64(s float64, v int64) int64 {
	return int64(float64(v) * float64(s))
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
	for retry := 0; retry <= m.Retries; retry += 1 {
		timestamp = time.Now()
		err := m.getFields()
		if err != nil {
			mberr, ok := err.(*mb.ModbusError)
			if ok && mberr.ExceptionCode == mb.ExceptionCodeServerDeviceBusy && retry < m.Retries {
				log.Printf("I! [inputs.modbus] device busy! Retrying %d more time(s)...", m.Retries-retry)
				time.Sleep(m.RetriesWaitTime.Duration)
				continue
			}
			disconnect(m)
			m.isConnected = false
			return err
		}
		// Reading was successful, leave the retry loop
		break
	}

	grouper := metric.NewSeriesGrouper()
	for _, reg := range m.registers {
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
			grouper.Add(measurement, tags, timestamp, field.Name, field.value)
		}

		// Add the metrics grouped by series to the accumulator
		for _, metric := range grouper.Metrics() {
			acc.AddMetric(metric)
		}
	}

	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })
}
