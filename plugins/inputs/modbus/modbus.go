package modbus

import (
	"encoding/binary"
	"errors"
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

type Modbus struct {
	Controller        string            `toml:"controller"`
	Baud_Rate         int               `toml:"baud_rate"`
	Data_Bits         int               `toml:"data_bits"`
	Parity            string            `toml:"parity"`
	Stop_Bits         int               `toml:"stop_bits"`
	Slave_Id          int               `toml:"slave_id"`
	Time_out          internal.Duration `toml:"time_out"`
	Discrete_Inputs   []tag             `toml:"discrete_inputs"`
	Coils             []tag             `toml:"coils"`
	Holding_Registers []tag             `toml:"holding_registers"`
	Input_Registers   []tag             `toml:"input_registers"`
	registers         []register

	is_connected   bool
	is_initialized bool
	tcp_handler    *mb.TCPClientHandler
	serial_handler *mb.RTUClientHandler
	client         mb.Client
}

type register struct {
	Type            string
	registers_range []register_range
	raw_values      map[uint16]uint16
	ReadValue       func(uint16, uint16) ([]byte, error)
	Tags            []tag
}

type tag struct {
	Name       string   `toml:"name"`
	Byte_Order string   `toml:"byte_order"`
	Data_Type  string   `toml:"data_type"`
	Scale      string   `toml:"scale"`
	Address    []uint16 `toml:"address"`
	value      interface{}
}

type register_range struct {
	address uint16
	length  uint16
}

const (
	C_DIGITAL           = "digital"
	C_ANALOG            = "analog"
	C_DISCRETE_INPUTS   = "Discrete_Inputs"
	C_COILS             = "Coils"
	C_HOLDING_REGISTERS = "Holding_Registers"
	C_INPUT_REGISTERS   = "Input_Registers"
)

var ModbusConfig = `
slave_id = 1
 time_out = "1s"
 
 #TCP 
 controller="tcp://localhost:1502"
 
 #RTU
 #controller="file:///dev/ttyUSB0"
 #baudRate = 9600
 #dataBits = 8
 #parity = "N"
 #stopBits = 1

 discrete_inputs = []
 coils = [] 
 holding_registers = [
   { name = "Voltage",     byte_order = "AB",   data_type = "FLOAT32", scale="0.1" ,   address = [0]},
   { name = "Current",     byte_order = "ABCD", data_type = "FLOAT32", scale="0.001" , address = [1, 2]},
   { name = "Power",       byte_order = "ABCD", data_type = "FLOAT32", scale="0.1" ,   address = [3,4]},
   { name = "Energy",      byte_order = "ABCD", data_type = "FLOAT32", scale="0.001" , address = [5,6]},
   { name = "Frequency",   byte_order = "AB",   data_type = "FLOAT32", scale="0.1" ,   address = [7]},
   { name = "PowerFactor", byte_order = "AB",   data_type = "FLOAT32", scale="0.01" ,  address = [8]},
 ] 
 input_registers = []
`

func (s *Modbus) SampleConfig() string {
	return ModbusConfig
}

func (s *Modbus) Description() string {
	return "Modbus client"
}

func connect(m *Modbus) error {
	u, err := url.Parse(m.Controller)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "tcp":
		host, port, _ := net.SplitHostPort(u.Host)
		m.tcp_handler = mb.NewTCPClientHandler(host + ":" + port)
		m.tcp_handler.Timeout = m.Time_out.Duration
		m.tcp_handler.SlaveId = byte(m.Slave_Id)
		m.client = mb.NewClient(m.tcp_handler)
		err := m.tcp_handler.Connect()
		if err != nil {
			return err
		}
		m.is_connected = true
		return nil
	case "file":
		m.serial_handler = mb.NewRTUClientHandler(u.Path)
		m.serial_handler.Timeout = m.Time_out.Duration
		m.serial_handler.SlaveId = byte(m.Slave_Id)
		m.serial_handler.BaudRate = m.Baud_Rate
		m.serial_handler.DataBits = m.Data_Bits
		m.serial_handler.Parity = m.Parity
		m.serial_handler.StopBits = m.Stop_Bits
		m.client = mb.NewClient(m.serial_handler)
		err := m.serial_handler.Connect()
		if err != nil {
			return err
		}
		m.is_connected = true
		return nil
	default:
		return errors.New("Not valid Protocol")
	}
}

func initialization(m *Modbus) {
	r := reflect.ValueOf(m).Elem()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)

		if f.Type().String() == "[]modbus.tag" {
			tags := f.Interface().([]tag)

			if len(tags) == 0 {
				continue
			}

			addrs := []uint16{}
			for _, tag := range tags {
				for _, a := range tag.Address {
					addrs = append(addrs, a)
				}
			}

			addrs = removeDuplicates(addrs)
			sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

			raw_values := make(map[uint16]uint16)
			for _, j := range addrs {
				raw_values[j] = 0
			}

			ii := 0
			var registers_range []register_range

			name := r.Type().Field(i).Name
			for range addrs {
				if ii < len(addrs) {
					start := addrs[ii]
					end := start

					for ii < len(addrs)-1 && addrs[ii+1]-addrs[ii] == 1 {
						end = addrs[ii+1]
						ii++
					}
					ii++
					registers_range = append(registers_range, register_range{start, end - start + 1})
				}
			}

			var fn func(uint16, uint16) ([]byte, error)

			if name == C_DISCRETE_INPUTS {
				fn = m.client.ReadDiscreteInputs
			}

			if name == C_COILS {
				fn = m.client.ReadCoils
			}

			if name == C_INPUT_REGISTERS {
				fn = m.client.ReadInputRegisters
			}

			if name == C_HOLDING_REGISTERS {
				fn = m.client.ReadHoldingRegisters
			}

			m.registers = append(m.registers, register{name, registers_range, raw_values, fn, tags})
		}
	}
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

func addFields(t []tag) map[string]interface{} {
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

func (m *Modbus) GetTags() error {
	for _, r := range m.registers {
		for _, rr := range r.registers_range {
			res, err := r.ReadValue(uint16(rr.address), uint16(rr.length))
			if err != nil {
				return err
			}

			if r.Type == C_DISCRETE_INPUTS || r.Type == C_COILS {
				getDigitalValue(r, rr, res)
			}

			if r.Type == C_INPUT_REGISTERS || r.Type == C_HOLDING_REGISTERS {
				getAnalogValue(r, rr, res)
			}
		}
	}

	return nil
}

func getDigitalValue(r register, rr register_range, results []uint8) error {
	for i := 0; i < len(results); i++ {
		for j := uint16(0); j < rr.length; j++ {
			r.raw_values[rr.address+j] = uint16(results[i] >> uint(j) & 0x01)

		}
	}

	for i := 0; i < len(r.Tags); i++ {
		r.Tags[i].value = r.raw_values[r.Tags[i].Address[0]]
	}

	return nil
}

func getAnalogValue(r register, rr register_range, res []uint8) error {
	for i := 0; i < len(res); i += 2 {
		r.raw_values[rr.address+uint16(i)/2] = uint16(res[i])<<8 | uint16(res[i+1])
	}

	for i := 0; i < len(r.Tags); i++ {
		bytes := []byte{}
		for _, rv := range r.Tags[i].Address {
			bytes = append(bytes, byte(r.raw_values[rv]>>8))
			bytes = append(bytes, byte(r.raw_values[rv]&255))
		}

		r.Tags[i].value = convertDataType(r.Tags[i], bytes)
	}

	return nil
}

func convertDataType(t tag, bytes []byte) interface{} {
	switch t.Data_Type {
	case "UINT16":
		e16, _ := convertEndianness16(t.Byte_Order, bytes).(uint16)
		f16 := format16(t.Data_Type, e16).(uint16)
		return scale16(t.Scale, f16)
	case "INT16":
		e16, _ := convertEndianness16(t.Byte_Order, bytes).(uint16)
		return format16(t.Data_Type, e16)
	case "UINT32":
		e32, _ := convertEndianness32(t.Byte_Order, bytes).(uint32)
		f32 := format32(t.Data_Type, e32).(uint32)
		return scale32(t.Scale, f32)
	case "INT32":
		e32, _ := convertEndianness32(t.Byte_Order, bytes).(uint32)
		return format32(t.Data_Type, e32)
	case "FLOAT32-IEEE":
		e32, _ := convertEndianness32(t.Byte_Order, bytes).(uint32)
		return format32(t.Data_Type, e32)
	case "FLOAT32":
		if len(bytes) == 2 {
			e16, _ := convertEndianness16(t.Byte_Order, bytes).(uint16)
			f16 := format16(t.Data_Type, e16).(uint16)
			return scale16(t.Scale, f16)
		} else {
			e32, _ := convertEndianness32(t.Byte_Order, bytes).(uint32)
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

func (m *Modbus) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	if !m.is_connected {
		err := connect(m)
		if err != nil {
			m.is_connected = false
			return err
		}
	}

	if !m.is_initialized {
		initialization(m)
		m.is_initialized = true
	}

	err := m.GetTags()
	if err != nil {
		m.is_connected = false
		return err
	}

	for _, reg := range m.registers {
		fields = addFields(reg.Tags)
		acc.AddFields("modbus."+reg.Type, fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })
}
