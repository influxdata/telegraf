package modbus

import (
	"encoding/binary"
	"errors"
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

type Modbus struct {
	Controller        string            `toml:"controller"`
	Protocol          string            `toml:"protocol"`
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
	ascii_handler  *mb.ASCIIClientHandler
	client         mb.Client
}

type register struct {
	Type            string
	registers_range []register_range
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
	C_DISCRETE_INPUTS   = "Discrete_Inputs"
	C_COILS             = "Coils"
	C_HOLDING_REGISTERS = "Holding_Registers"
	C_INPUT_REGISTERS   = "Input_Registers"
)

var ModbusConfig = `
 slave_id = 1
 time_out = "1s"
 #protocol = "RTU"
 
 #TCP 
 controller = "tcp://localhost:1502"
 
 #RTU
 #controller = "file:///dev/ttyUSB0"
 #baud_rate = 9600
 #data_bits = 8
 #parity = "N"
 #stop_bits = 1
 
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
   { name = "PowerFactor", byte_order = "AB",   data_type = "FLOAT32", scale="0.01" ,  address = [8]},
   { name = "Voltage",     byte_order = "AB",   data_type = "FLOAT32", scale="0.1" ,   address = [0]},   
   { name = "Energy",      byte_order = "ABCD", data_type = "FLOAT32", scale="0.001" , address = [5,6]},
   { name = "Current",     byte_order = "ABCD", data_type = "FLOAT32", scale="0.001" , address = [1, 2]},
   { name = "Frequency",   byte_order = "AB",   data_type = "FLOAT32", scale="0.1" ,   address = [7]},
   { name = "Power",       byte_order = "ABCD", data_type = "FLOAT32", scale="0.1" ,   address = [3,4]},      
 ] 
 input_registers = [
   { name = "TankLevel",   byte_order = "AB",   data_type = "INT16",   scale="1" ,     address = [0]},
   { name = "TankPH",      byte_order = "AB",   data_type = "INT16",  scale="1" ,     address = [1]},   
   { name = "Pump1-Speed", byte_order = "ABCD", data_type = "INT32",   scale="1" ,     address = [3,4]},
 ]
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
		host, port, _err := net.SplitHostPort(u.Host)
		if _err != nil {
			return _err
		}
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
		if m.Protocol == "RTU" {
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
		} else if m.Protocol == "ASCII" {
			m.ascii_handler = mb.NewASCIIClientHandler(u.Path)
			m.ascii_handler.Timeout = m.Time_out.Duration
			m.ascii_handler.SlaveId = byte(m.Slave_Id)
			m.ascii_handler.BaudRate = m.Baud_Rate
			m.ascii_handler.DataBits = m.Data_Bits
			m.ascii_handler.Parity = m.Parity
			m.ascii_handler.StopBits = m.Stop_Bits
			m.client = mb.NewClient(m.ascii_handler)
			err := m.ascii_handler.Connect()
			if err != nil {
				return err
			}
			m.is_connected = true
			return nil
		} else {
			return errors.New(fmt.Sprintf("Not valid protcol [%s] - [%s] ", u.Scheme, m.Protocol))
		}
	default:
		return errors.New("Not valid Controller")
	}
}

func initialization(m *Modbus) error {
	r := reflect.ValueOf(m).Elem()
	for i := 0; i < r.NumField(); i++ {
		f := r.Field(i)

		if f.Type().String() == "[]modbus.tag" {
			tags := f.Interface().([]tag)
			name := r.Type().Field(i).Name

			if len(tags) == 0 {
				continue
			}

			err := validateTags(tags, name)
			if err != nil {
				return err
			}

			addrs := []uint16{}
			for _, tag := range tags {
				for _, a := range tag.Address {
					addrs = append(addrs, a)
				}
			}

			addrs = removeDuplicates(addrs)
			sort.Slice(addrs, func(i, j int) bool { return addrs[i] < addrs[j] })

			ii := 0
			var registers_range []register_range

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
					registers_range = append(registers_range, register_range{start, end - start + 1})
				}
			}

			var fn func(uint16, uint16) ([]byte, error)

			if name == C_DISCRETE_INPUTS {
				fn = m.client.ReadDiscreteInputs
			} else if name == C_COILS {
				fn = m.client.ReadCoils
			} else if name == C_INPUT_REGISTERS {
				fn = m.client.ReadInputRegisters
			} else if name == C_HOLDING_REGISTERS {
				fn = m.client.ReadHoldingRegisters
			} else {
				return errors.New("Not Valid function")
			}

			m.registers = append(m.registers, register{name, registers_range, fn, tags})
		}
	}
	m.is_initialized = true

	return nil
}

func validateTags(t []tag, n string) error {
	byte_order := []string{"AB", "BA", "ABCD", "CDAB", "BADC", "DCBA"}
	data_type := []string{"UINT16", "INT16", "UINT32", "INT32", "FLOAT32-IEEE", "FLOAT32"}

	name_encountered := map[string]bool{}
	for i := range t {
		//check empty name
		if t[i].Name == "" {
			return errors.New(fmt.Sprintf("Empty Name in %s", n))
		}

		//search name duplicate
		if name_encountered[t[i].Name] {
			return errors.New(fmt.Sprintf("Name [%s] in %s is Duplicated", t[i].Name, n))
		} else {
			name_encountered[t[i].Name] = true
		}

		if n == C_INPUT_REGISTERS || n == C_HOLDING_REGISTERS {
			// search byte order
			byte_order_encountered := false
			for j := range byte_order {
				if byte_order[j] == t[i].Byte_Order {
					byte_order_encountered = true
					break
				}
			}

			if !byte_order_encountered {
				return errors.New(fmt.Sprintf("Not valid Byte Order [%s] in %s", t[i].Byte_Order, n))
			}

			// search data type
			data_type_encountered := false
			for j := range byte_order {
				if data_type[j] == t[i].Data_Type {
					data_type_encountered = true
					break
				}
			}

			if !data_type_encountered {
				return errors.New(fmt.Sprintf("Not valid Data Type [%s] in %s", t[i].Data_Type, n))
			}

			// check scale
			_, err := strconv.ParseFloat(t[i].Scale, 32)
			if err != nil {
				return errors.New(fmt.Sprintf("Not valid Scale [%s] in %s", t[i].Scale, n))
			}
		}

		// check address
		if len(t[i].Address) == 0 || len(t[i].Address) > 2 {
			return errors.New(fmt.Sprintf("Not valid address [%v] length [%v] in %s", t[i].Address, len(t[i].Address), n))
		} else if n == C_INPUT_REGISTERS || n == C_HOLDING_REGISTERS {
			if (len(t[i].Address) == 1 && len(t[i].Byte_Order) != 2) || (len(t[i].Address) == 2 && len(t[i].Byte_Order) != 4) {
				return errors.New(fmt.Sprintf("Not valid byte order [%s] and address [%v]  in %s", t[i].Byte_Order, t[i].Address, n))
			}

			// search duplicated
			if len(t[i].Address) > len(removeDuplicates(t[i].Address)) {
				return errors.New(fmt.Sprintf("Duplicate address [%v]  in %s", t[i].Address, n))
			}

		} else if len(t[i].Address) > 1 || (n == C_INPUT_REGISTERS || n == C_HOLDING_REGISTERS) {
			return errors.New(fmt.Sprintf("Not valid address [%v] length [%v] in %s", t[i].Address, len(t[i].Address), n))
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
		raw_values := make(map[uint16]uint16)		
		for _, rr := range r.registers_range {			
			res, err := r.ReadValue(uint16(rr.address), uint16(rr.length))
			if err != nil {
				return err
			}

			if r.Type == C_DISCRETE_INPUTS || r.Type == C_COILS {
				for i := 0; i < len(res); i++ {
					for j := uint16(0); j < rr.length; j++ {
						raw_values[rr.address+j] = uint16(res[i] >> uint(j) & 0x01)
					}
				}
				continue
			}	

			if r.Type == C_INPUT_REGISTERS || r.Type == C_HOLDING_REGISTERS {
				for i := 0; i < len(res); i += 2 {
					raw_values[rr.address+uint16(i)/2] = uint16(res[i])<<8 | uint16(res[i+1])
				}				
			}					
		}

		if r.Type == C_DISCRETE_INPUTS || r.Type == C_COILS {
			setDigitalValue(r, raw_values)
		}

		if r.Type == C_INPUT_REGISTERS || r.Type == C_HOLDING_REGISTERS {
			setAnalogValue(r, raw_values)
		}		
	}

	return nil
}

func setDigitalValue(r register, raw_values map[uint16]uint16) error {		
	for i := 0; i < len(r.Tags); i++ {
		r.Tags[i].value = raw_values[r.Tags[i].Address[0]]
	}

	return nil
}

func setAnalogValue(r register, raw_values map[uint16]uint16) error {	
	for i := 0; i < len(r.Tags); i++ {
		bytes := []byte{}
		for _, rv := range r.Tags[i].Address {
			bytes = append(bytes, byte(raw_values[rv]>>8))
			bytes = append(bytes, byte(raw_values[rv]&255))
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
		err := initialization(m)
		if err != nil {
			return err
		}
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
