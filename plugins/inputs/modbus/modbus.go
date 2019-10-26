package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"net/url"
	"reflect"
	"sort"
	"strconv"

	mb "github.com/goburrow/modbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Modbus holds all data relevant to the plugin
type Modbus struct {
	Controller       string            `toml:"controller"`
	TransmissionMode string            `toml:"transmission_mode"`
	BaudRate         int               `toml:"baud_rate"`
	DataBits         int               `toml:"data_bits"`
	Parity           string            `toml:"parity"`
	StopBits         int               `toml:"stop_bits"`
	SlaveID          int               `toml:"slave_id"`
	Timeout          internal.Duration `toml:"timeout"`
	DiscreteInputs   []fieldContainer  `toml:"discrete_inputs"`
	Coils            []fieldContainer  `toml:"coils"`
	HoldingRegisters []fieldContainer  `toml:"holding_registers"`
	InputRegisters   []fieldContainer  `toml:"input_registers"`
	registers        []register
	isConnected      bool
	isInitialized    bool
	tcpHandler       *mb.TCPClientHandler
	rtuHandler       *mb.RTUClientHandler
	asciiHandler     *mb.ASCIIClientHandler
	client           mb.Client
}

type register struct {
	Type           string
	RegistersRange []registerRange
	ReadValue      func(uint16, uint16) ([]byte, error)
	Fields         []fieldContainer
}

type fieldContainer struct {
	Name      string   `toml:"name"`
	ByteOrder string   `toml:"byte_order"`
	DataType  string   `toml:"data_type"`
	Scale     string   `toml:"scale"`
	Address   []uint16 `toml:"address"`
	value     interface{}
}

type registerRange struct {
	address uint16
	length  uint16
}

const (
	cDiscreteInputs   = "DiscreteInputs"
	cCoils            = "Coils"
	cHoldingRegisters = "HoldingRegisters"
	cInputRegisters   = "InputRegisters"
)

const description = `Retrieve data from MODBUS slave devices`
const sampleConfig = `
 ## Connection Configuration
 ##
 ## The plugin supports connections to PLCs via MODBUS/TCP or
 ## via serial line communication in binary (RTU) or readable (ASCII) encoding
 ##
 
 ## Slave ID - addresses a MODBUS device on the bus
 ## Range: 0 - 255 [0 = broadcast; 248 - 255 = reserved]
 slave_id = 1
 
 ## Timeout for each request
 timeout = "1s"
 
 # TCP - connect via Modbus/TCP
 controller = "tcp://localhost:502"
 
 # Serial (RS485; RS232)
 #controller = "file:///dev/ttyUSB0"
 #baud_rate = 9600
 #data_bits = 8
 #parity = "N"
 #stop_bits = 1
 #transmission_mode = "RTU"
 
 
 ## Measurements
 ##
 
 ## Digital Variables, Discrete Inputs and Coils
 ## name    - the variable name
 ## address - variable address
 
 discrete_inputs = [
   { name = "Start",          address = [0]},
   { name = "Stop",           address = [1]},
   { name = "Reset",          address = [2]},
   { name = "EmergencyStop",  address = [3]},
 ]
 coils = [
   { name = "Motor1-Run",     address = [0]},
   { name = "Motor1-Jog",     address = [1]},
   { name = "Motor1-Stop",    address = [2]},
 ]
 
 ## Analog Variables, Input Registers and Holding Registers
 ## name       - the variable name
 ## byte_order - the ordering of bytes
 ##  |---AB, ABCD   - Big Endian
 ##  |---BA, DCBA   - Little Endian
 ##  |---BADC       - Mid-Big Endian
 ##  |---CDAB       - Mid-Little Endian
 ## data_type  - UINT16, INT16, INT32, UINT32, FLOAT32, FLOAT32-IEEE (the IEEE 754 binary representation)
 ## scale      - the final numeric variable representation
 ## address    - variable address
 
 holding_registers = [
   { name = "PowerFactor", byte_order = "AB",   data_type = "FLOAT32", scale="0.01",  address = [8]},
   { name = "Voltage",     byte_order = "AB",   data_type = "FLOAT32", scale="0.1",   address = [0]},
   { name = "Energy",      byte_order = "ABCD", data_type = "FLOAT32", scale="0.001", address = [5,6]},
   { name = "Current",     byte_order = "ABCD", data_type = "FLOAT32", scale="0.001", address = [1, 2]},
   { name = "Frequency",   byte_order = "AB",   data_type = "FLOAT32", scale="0.1",   address = [7]},
   { name = "Power",       byte_order = "ABCD", data_type = "FLOAT32", scale="0.1",   address = [3,4]},
 ]
 input_registers = [
   { name = "TankLevel",   byte_order = "AB",   data_type = "INT16",   scale="1",     address = [0]},
   { name = "TankPH",      byte_order = "AB",   data_type = "INT16",   scale="1",     address = [1]},
   { name = "Pump1-Speed", byte_order = "ABCD", data_type = "INT32",   scale="1",     address = [3,4]},
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

// Connect to a MODBUS Slave device via Modbus/[TCP|RTU|ASCII]
func connect(m *Modbus) error {
	u, err := url.Parse(m.Controller)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp":
		host, port, _err := net.SplitHostPort(u.Host)
		if _err != nil {
			return _err
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
			return fmt.Errorf("Not valid protcol [%s] - [%s] ", u.Scheme, m.TransmissionMode)
		}
	default:
		return fmt.Errorf("Not valid Controller")
	}
}

func initialization(m *Modbus) error {
	r := reflect.ValueOf(m).Elem()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)

		if f.Type().String() == "[]modbus.fieldContainer" {
			fields := f.Interface().([]fieldContainer)
			name := r.Type().Field(i).Name

			if len(fields) == 0 {
				continue
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

			var fn func(uint16, uint16) ([]byte, error)

			if name == cDiscreteInputs {
				fn = m.client.ReadDiscreteInputs
			} else if name == cCoils {
				fn = m.client.ReadCoils
			} else if name == cInputRegisters {
				fn = m.client.ReadInputRegisters
			} else if name == cHoldingRegisters {
				fn = m.client.ReadHoldingRegisters
			} else {
				return fmt.Errorf("Not Valid function")
			}

			m.registers = append(m.registers, register{name, registersRange, fn, fields})
		}
	}
	m.isInitialized = true

	return nil
}

func validateFieldContainers(t []fieldContainer, n string) error {
	byteOrder := []string{"AB", "BA", "ABCD", "CDAB", "BADC", "DCBA"}
	dataType := []string{"UINT16", "INT16", "UINT32", "INT32", "FLOAT32-IEEE", "FLOAT32"}

	nameEncountered := map[string]bool{}
	for i := range t {
		//check empty name
		if t[i].Name == "" {
			return fmt.Errorf("Empty Name in %s", n)
		}

		//search name duplicate
		if nameEncountered[t[i].Name] {
			return fmt.Errorf("Name [%s] in %s is Duplicated", t[i].Name, n)
		} else {
			nameEncountered[t[i].Name] = true
		}

		if n == cInputRegisters || n == cHoldingRegisters {
			// search byte order
			byteOrderEncountered := false
			for j := range byteOrder {
				if byteOrder[j] == t[i].ByteOrder {
					byteOrderEncountered = true
					break
				}
			}

			if !byteOrderEncountered {
				return fmt.Errorf("Not valid Byte Order [%s] in %s", t[i].ByteOrder, n)
			}

			// search data type
			dataTypeEncountered := false
			for j := range byteOrder {
				if dataType[j] == t[i].DataType {
					dataTypeEncountered = true
					break
				}
			}

			if !dataTypeEncountered {
				return fmt.Errorf("Not valid Data Type [%s] in %s", t[i].DataType, n)
			}

			// check scale
			_, err := strconv.ParseFloat(t[i].Scale, 32)
			if err != nil {
				return fmt.Errorf("Not valid Scale [%s] in %s", t[i].Scale, n)
			}
		}

		// check address
		if len(t[i].Address) == 0 || len(t[i].Address) > 2 {
			return fmt.Errorf("Not valid address [%v] length [%v] in %s", t[i].Address, len(t[i].Address), n)
		} else if n == cInputRegisters || n == cHoldingRegisters {
			if (len(t[i].Address) == 1 && len(t[i].ByteOrder) != 2) || (len(t[i].Address) == 2 && len(t[i].ByteOrder) != 4) {
				return fmt.Errorf("Not valid byte order [%s] and address [%v]  in %s", t[i].ByteOrder, t[i].Address, n)
			}

			// search duplicated
			if len(t[i].Address) > len(removeDuplicates(t[i].Address)) {
				return fmt.Errorf("Duplicate address [%v]  in %s", t[i].Address, n)
			}

		} else if len(t[i].Address) > 1 || (n == cInputRegisters || n == cHoldingRegisters) {
			return fmt.Errorf("Not valid address [%v] length [%v] in %s", t[i].Address, len(t[i].Address), n)
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

func addFields(t []fieldContainer) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(t); i++ {
		if len(t[i].Name) > 0 {
			fields[t[i].Name] = t[i].value
		} else {
			name := ""
			for _, e := range t[i].Address {
				name = name + "-" + strconv.Itoa(int(e))
			}
			name = name[:len(name)-1]
			fields[name] = t[i]
		}
	}

	return fields
}

func (m *Modbus) getFields() error {
	for _, r := range m.registers {
		rawValues := make(map[uint16]uint16)
		for _, rr := range r.RegistersRange {
			res, err := r.ReadValue(uint16(rr.address), uint16(rr.length))
			if err != nil {
				return err
			}

			if r.Type == cDiscreteInputs || r.Type == cCoils {
				for i := 0; i < len(res); i++ {
					for j := uint16(0); j < rr.length; j++ {
						rawValues[rr.address+j] = uint16(res[i] >> uint(j) & 0x01)
					}
				}
				continue
			}

			if r.Type == cInputRegisters || r.Type == cHoldingRegisters {
				for i := 0; i < len(res); i += 2 {
					rawValues[rr.address+uint16(i)/2] = uint16(res[i])<<8 | uint16(res[i+1])
				}
			}
		}

		if r.Type == cDiscreteInputs || r.Type == cCoils {
			setDigitalValue(r, rawValues)
		}

		if r.Type == cInputRegisters || r.Type == cHoldingRegisters {
			setAnalogValue(r, rawValues)
		}
	}

	return nil
}

func setDigitalValue(r register, rawValues map[uint16]uint16) error {
	for i := 0; i < len(r.Fields); i++ {
		r.Fields[i].value = rawValues[r.Fields[i].Address[0]]
	}

	return nil
}

func setAnalogValue(r register, rawValues map[uint16]uint16) error {
	for i := 0; i < len(r.Fields); i++ {
		bytes := []byte{}
		for _, rv := range r.Fields[i].Address {
			bytes = append(bytes, byte(rawValues[rv]>>8))
			bytes = append(bytes, byte(rawValues[rv]&255))
		}

		r.Fields[i].value = convertDataType(r.Fields[i], bytes)
	}

	return nil
}

func convertDataType(t fieldContainer, bytes []byte) interface{} {
	switch t.DataType {
	case "UINT16":
		e16, _ := convertEndianness16(t.ByteOrder, bytes).(uint16)
		f16 := format16(t.DataType, e16).(uint16)
		return scale16(t.Scale, f16)
	case "INT16":
		e16, _ := convertEndianness16(t.ByteOrder, bytes).(uint16)
		return format16(t.DataType, e16)
	case "UINT32":
		e32, _ := convertEndianness32(t.ByteOrder, bytes).(uint32)
		f32 := format32(t.DataType, e32).(uint32)
		return scale32(t.Scale, f32)
	case "INT32":
		e32, _ := convertEndianness32(t.ByteOrder, bytes).(uint32)
		return format32(t.DataType, e32)
	case "FLOAT32-IEEE":
		e32, _ := convertEndianness32(t.ByteOrder, bytes).(uint32)
		return format32(t.DataType, e32)
	case "FLOAT32":
		if len(bytes) == 2 {
			e16, _ := convertEndianness16(t.ByteOrder, bytes).(uint16)
			f16 := format16(t.DataType, e16).(uint16)
			return scale16(t.Scale, f16)
		} else {
			e32, _ := convertEndianness32(t.ByteOrder, bytes).(uint32)
			return scale32(t.Scale, e32)
		}
	default:
		return 0
	}
}

func convertEndianness16(o string, b []byte) interface{} {
	switch o {
	case "AB":
		return binary.BigEndian.Uint16(b)
	case "BA":
		return binary.LittleEndian.Uint16(b)
	default:
		return b
	}
}

func convertEndianness32(o string, b []byte) interface{} {
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
		return b
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

func scale16(s string, v uint16) interface{} {
	if len(s) == 0 {
		return v
	}

	n, err := strconv.ParseFloat(s, 32)
	if err == nil {
		return float32(v) * float32(n)
	}

	return 0
}

func scale32(s string, v uint32) interface{} {
	if len(s) == 0 {
		return v
	}

	n, err := strconv.ParseFloat(s, 32)
	if err == nil {
		return float32(v) * float32(n)
	}

	return 0
}

// Gather implements the telegraf plugin interface method for data accumulation
func (m *Modbus) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	if !m.isConnected {
		err := connect(m)
		if err != nil {
			m.isConnected = false
			return err
		}
	}

	if !m.isInitialized {
		err := initialization(m)
		if err != nil {
			return err
		}
	}

	err := m.getFields()
	if err != nil {
		m.isConnected = false
		return err
	}

	for _, reg := range m.registers {
		fields = addFields(reg.Fields)
		// telegraf.Accumulator.AddFields(Name, Fields, Tags, Time)
		acc.AddFields("modbus."+reg.Type, fields, tags)
	}

	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })
}
