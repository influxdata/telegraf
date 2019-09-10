package modbus

import (
	"encoding/binary"
	"errors"
	"math"
	"sort"
	"strconv"
	
	mb "github.com/goburrow/modbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Modbus struct {
	Type          string                   `toml:"type"`
	Controller    string                   `toml:"controller"`
	Port          int                      `toml:"port"` 
	BaudRate      int                      `toml:"baud_rate"`
	DataBits      int                      `toml:"data_bits"`
	Parity        string                   `toml:"parity"` 
	StopBits      int                      `toml:"stop_bits"`
	SlaveId       int                      `toml:"slave_id"` 
	Timeout       internal.Duration        `toml:"time_out"`
	Registers     registers                `toml:"registers"`
	isConnected   bool      
	isInitialized bool
	handlerTcp    *mb.TCPClientHandler
	handlerSerial *mb.RTUClientHandler
	Client        mb.Client
}

type registers struct {
	DiscreteInputs   register `toml:"discrete_inputs"`
	Coils            register `toml:"coils"`
	HoldingRegisters register `toml:"holding_registers"`
	InputRegisters   register `toml:"input_registers"`
}

type register struct {
	Tags      []tag    `toml:"tags"`
	chunks    []chunk  
	rawValues map[int]uint16
}

type tag struct {
	Name     string `toml:"name"`
	Order    string `toml:"order"`
	DataType string `toml:"data_type"`
	Scale    string `toml:"scale"`
	Address  []int  `toml:"address"`
	value    interface{}
}

type chunk struct {
	address int
	length  int
}

const (
	C_DIGITAL = "digital"
	C_ANALOG  = "analog"
)

var ModbusConfig = `
#TCP
type = "TCP"
controller="localhost"
port = 1502

#RTU
#type = "RTU"
#controller="/dev/ttyUSB0"
#baudRate = 9600
#dataBits = 8
#parity = "N"
#stopBits = 1

slave_id = 1
time_out = "1s"

 [[inputs.modbus.registers.holding_registers.tags]]
  name = "Voltage"
  order = "AB"
  datatype = "FLOAT32"
  scale = "/10"
  address = [
   0
  ]

 [[inputs.modbus.registers.holding_registers.tags]]
  name = "Current"
  order ="ABCD"
  datatype = "FLOAT32"
  scale = "/1000"
  address = [
   1,
   2
  ]

 [[inputs.modbus.registers.holding_registers.tags]]
   name = "Power"
   order = "ABCD"
   datatype = "FLOAT32"
   scale = "/10"
   address = [
	3,
	4
   ]

 [[inputs.modbus.registers.holding_registers.tags]]
   name = "Energy"
   order = "ABCD"
   datatype = "FLOAT32"	
   scale = "/1000"
   address = [
	5,
	6
   ]

 [[inputs.modbus.registers.holding_registers.tags]]
   name = "Frequency"
   order = "AB"	
   datatype = "FLOAT32"	
   scale = "/10"
   address = [
	7
   ]

 [[inputs.modbus.registers.holding_registers.tags]]
   name = "PowerFactor"
   order = "AB"
   datatype = "FLOAT32"
   scale = "/100"
   address = [
	8
   ]
`

func (s *Modbus) SampleConfig() string {
	return ModbusConfig
}

func (s *Modbus) Description() string {
	return "Modbus client"
}

func removeDuplicates(elements []int) []int {
	encountered := map[int]bool{}
	result := []int{}

	for v := range elements {
		if encountered[elements[v]] == true {
		} else {
			encountered[elements[v]] = true
			result = append(result, elements[v])
		}
	}

	return result
}

func createRawValueMap(r []tag, rawValue map[int]uint16) {
	addr := []int{}

	for _, element := range r {
		for _, a := range element.Address {
			addr = append(addr, a)
		}
	}

	addr = removeDuplicates(addr)
	sort.Ints(addr)

	for _, element := range addr {
		rawValue[element] = 0
	}
}

func createChunks(ch *[]chunk, rawValue map[int]uint16) {
	r := []int{}
	chunks := [][]int{}
	chunk_t := []int{}

	for k := range rawValue {
		r = append(r, k)
	}

	sort.Ints(r)

	for i, element := range r {
		n := 1
		if i+1 == len(r) {
			n = -1
			if len(r) == 1 {
				n = 0
			}
		}
		if element+n == r[i+n] {
			chunk_t = append(chunk_t, element)
			if i+1 == len(r) {
				chunks = append(chunks, chunk_t)
			}
		} else {
			chunk_t = append(chunk_t, element)
			chunks = append(chunks, chunk_t)
			chunk_t = []int{}
		}
	}

	for _, element := range chunks {
		*ch = append(*ch, chunk{element[0], len(element)})
	}
}

func initialization(m *Modbus) {
	m.Registers.DiscreteInputs.rawValues = make(map[int]uint16)
	m.Registers.Coils.rawValues = make(map[int]uint16)
	m.Registers.HoldingRegisters.rawValues = make(map[int]uint16)
	m.Registers.InputRegisters.rawValues = make(map[int]uint16)

	createRawValueMap(m.Registers.DiscreteInputs.Tags, m.Registers.DiscreteInputs.rawValues)
	createRawValueMap(m.Registers.Coils.Tags, m.Registers.Coils.rawValues)
	createRawValueMap(m.Registers.HoldingRegisters.Tags, m.Registers.HoldingRegisters.rawValues)
	createRawValueMap(m.Registers.InputRegisters.Tags, m.Registers.InputRegisters.rawValues)

	createChunks(&m.Registers.DiscreteInputs.chunks, m.Registers.DiscreteInputs.rawValues)
	createChunks(&m.Registers.Coils.chunks, m.Registers.Coils.rawValues)
	createChunks(&m.Registers.HoldingRegisters.chunks, m.Registers.HoldingRegisters.rawValues)
	createChunks(&m.Registers.InputRegisters.chunks, m.Registers.InputRegisters.rawValues)
}

func connect(m *Modbus) error {
	switch m.Type {
	case "TCP":
		m.handlerTcp = mb.NewTCPClientHandler(m.Controller + ":" + strconv.Itoa(m.Port))		
		m.handlerTcp.Timeout = m.Timeout.Duration
		m.handlerTcp.SlaveId = byte(m.SlaveId)
		m.Client = mb.NewClient(m.handlerTcp)
		err := m.handlerTcp.Connect()		
		if err != nil {
			return err
		}
		m.isConnected = true
		return nil
	case "RTU":
		m.handlerSerial = mb.NewRTUClientHandler(m.Controller)
		m.handlerSerial.Timeout = m.Timeout.Duration
		m.handlerSerial.SlaveId = byte(m.SlaveId)
		m.handlerSerial.BaudRate = m.BaudRate
		m.handlerSerial.DataBits = m.DataBits
		m.handlerSerial.Parity = m.Parity
		m.handlerSerial.StopBits = m.StopBits
		m.Client = mb.NewClient(m.handlerSerial)
		err := m.handlerSerial.Connect()		
		if err != nil {
			return err
		}
		m.isConnected = true
		return nil
	default:
		return errors.New("Not valid type")
	}
}

type fn func(uint16, uint16) ([]byte, error)

func getRawValue(t string, f fn, ch []chunk, r map[int]uint16) error {
	for _, chunk_t := range ch {
		results, err := f(uint16(chunk_t.address), uint16(chunk_t.length))
		if err != nil {
			return err
		}

		if t == C_DIGITAL {
			for i := 0; i < len(results); i++ {
				for b := 0; b < chunk_t.length; b++ {
					r[chunk_t.address+b] = uint16(results[i] >> uint(b) & 0x01)
				}

			}
		}

		if t == C_ANALOG {
			for i := 0; i < len(results); i += 2 {
				register := uint16(results[i])<<8 | uint16(results[i+1])
				r[chunk_t.address+i/2] = uint16(register)
			}
		}
	}

	return nil
}

func convertEndianness16(o string, r []byte) interface{} {
	switch o {
	case "AB":
		return binary.BigEndian.Uint16(r)
	case "BA":
		return binary.LittleEndian.Uint16(r)
	default:
		return r
	}
}

func convertEndianness32(o string, r []byte) interface{} {
	switch o {
	case "ABCD":
		return binary.BigEndian.Uint32(r)
	case "DCBA":
		return binary.LittleEndian.Uint32(r)
	case "BADC":
		return uint32(binary.LittleEndian.Uint16(r[0:]))<<16 | uint32(binary.LittleEndian.Uint16(r[2:]))
	case "CDAB":
		return uint32(binary.BigEndian.Uint16(r[2:]))<<16 | uint32(binary.BigEndian.Uint16(r[0:]))
	default:
		return r
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
	operator := s[0]
	switch operator {
	case '/':
		div, err := strconv.Atoi(s[1:])
		if err == nil {
			return float32(v) / float32(div)
		}
		return 0
	case '*':
		div, err := strconv.Atoi(s[1:])
		if err == nil {
			return v * uint16(div)
		}
		return 0
	default:
		return 0
	}
}

func scale32(s string, v uint32) interface{} {
	if len(s) == 0 {
		return v
	}
	operator := s[0]
	switch operator {
	case '/':
		div, err := strconv.Atoi(s[1:])
		if err == nil {
			return float32(v) / float32(div)
		}
		return 0
	case '*':
		div, err := strconv.Atoi(s[1:])
		if err == nil {
			return v * uint32(div)
		}
		return 0
	default:
		return 0
	}
}

func setTag(s string, t []tag, r map[int]uint16) {
	if s == C_DIGITAL {
		for i := 0; i < len(t); i++ {
			t[i].value = r[t[i].Address[0]]
		}
	}

	if s == C_ANALOG {
		for i := 0; i < len(t); i++ {
			rawValues := []byte{}
			for _, rv := range t[i].Address {
				rawValues = append(rawValues, byte(r[rv]>>8))
				rawValues = append(rawValues, byte(r[rv]&0x00FF))
			}

			switch t[i].DataType {
			case "UINT16":
				e := convertEndianness16(t[i].Order, rawValues)
				e16, _ := e.(uint16)
				f := format16(t[i].DataType, e16)
				f16 := f.(uint16)
				s := scale16(t[i].Scale, f16)
				t[i].value = s
			case "INT16":
				e := convertEndianness16(t[i].Order, rawValues)
				e16, _ := e.(uint16)
				f := format16(t[i].DataType, e16)
				t[i].value = f
			case "UINT32":
				e := convertEndianness32(t[i].Order, rawValues)
				e32, _ := e.(uint32)
				f := format32(t[i].DataType, e32)
				f32 := f.(uint32)
				s := scale32(t[i].Scale, f32)
				t[i].value = s
			case "INT32":
				e := convertEndianness32(t[i].Order, rawValues)
				e32, _ := e.(uint32)
				f := format32(t[i].DataType, e32)
				t[i].value = f
			case "FLOAT32-IEEE":
				e := convertEndianness32(t[i].Order, rawValues)
				e32, _ := e.(uint32)
				f := format32(t[i].DataType, e32)
				t[i].value = f
			case "FLOAT32":
				if len(rawValues) == 2 {
					e := convertEndianness16(t[i].Order, rawValues)
					e16, _ := e.(uint16)
					f := format16(t[i].DataType, e16)
					f16 := f.(uint16)
					s := scale16(t[i].Scale, f16)
					t[i].value = s
				} else {
					e := convertEndianness32(t[i].Order, rawValues)
					e32, _ := e.(uint32)
					s := scale32(t[i].Scale, e32)
					t[i].value = s
				}

			default:
				t[i].value = 0
			}
		}
	}
}

func addFields(t []tag) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(t); i++ {
		if len(t[i].Name) > 0 {
			fields[t[i].Name] = t[i].value
		} else {
			name := ""
			for _, e := range t[i].Address {
				name = name + "-" + strconv.Itoa(e)
			}
			name = name[:len(name)-1]
			fields[name] = t[i]
		}
	}

	return fields
}

func (m *Modbus) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	//Init
	if m.isInitialized == false {
		initialization(m)
		m.isInitialized = true
	}

	// Connect
	if m.isConnected == false {
		err := connect(m)
		if err != nil {
			m.isConnected = false
			return err
		}
	}

	// Get Raw Values
	err := getRawValue(C_DIGITAL,
		m.Client.ReadDiscreteInputs,
		m.Registers.DiscreteInputs.chunks,
		m.Registers.DiscreteInputs.rawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_DIGITAL,
		m.Client.ReadCoils,
		m.Registers.Coils.chunks,
		m.Registers.Coils.rawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_ANALOG,
		m.Client.ReadHoldingRegisters,
		m.Registers.HoldingRegisters.chunks,
		m.Registers.HoldingRegisters.rawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_ANALOG,
		m.Client.ReadInputRegisters,
		m.Registers.InputRegisters.chunks,
		m.Registers.InputRegisters.rawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	// Set tags
	setTag(C_DIGITAL, m.Registers.DiscreteInputs.Tags, m.Registers.DiscreteInputs.rawValues)
	setTag(C_DIGITAL, m.Registers.Coils.Tags, m.Registers.Coils.rawValues)
	setTag(C_ANALOG, m.Registers.HoldingRegisters.Tags, m.Registers.HoldingRegisters.rawValues)
	setTag(C_ANALOG, m.Registers.InputRegisters.Tags, m.Registers.InputRegisters.rawValues)

	// Add Fields
	fields = addFields(m.Registers.DiscreteInputs.Tags)
	acc.AddFields("modbus.inputs", fields, tags)
	fields = addFields(m.Registers.Coils.Tags)
	acc.AddFields("modbus.coils", fields, tags)
	fields = addFields(m.Registers.HoldingRegisters.Tags)
	acc.AddFields("modbus.HoldingRegister", fields, tags)
	fields = addFields(m.Registers.InputRegisters.Tags)
	acc.AddFields("modbus.InputRegisters", fields, tags)

	return nil
}

func init() {
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })
}
