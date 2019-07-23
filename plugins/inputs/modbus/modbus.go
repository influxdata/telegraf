package modbus

import (
	"errors"
	"sort"
	"strconv"
	"time"
	"math"
	"encoding/binary"	
	
	mb "github.com/goburrow/modbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Modbus struct {
	Type           string
	Controller     string
	Port           int
	BaudRate       int
	DataBits       int
	Parity         string
	StopBits       int
	SlaveId        int
	Timeout        int
	Registers      registers
	isConnected    bool
	isInitialized  bool
	HandlerTcp    *mb.TCPClientHandler
	HandlerSerial *mb.RTUClientHandler
	Client         mb.Client
}

type registers struct {
	DiscreteInputs 	 	register
	Coils          	 	register
	HoldingRegisters 	register
	InputRegisters  	register
}

type register struct {
	Tags      []tag
	Chunks    []chunk
	RawValues map[int]uint16
}

type tag struct {
	Name     string
	Order    string
	DataType string
	Scale    string
	Address  []int
	Value    interface{}
}

type chunk struct {
	Address int
	Length  int
}

const (
    C_DIGITAL	= "digital"
    C_ANALOG	= "analog"    
)

var ModbusConfig = `
 #TCP
 #type = "TCP"
 #controller="192.168.0.9"
 #port = 502

 #RTU
 type = "RTU"
 controller="/dev/ttyUSB0"
 baudRate = 9600
 dataBits = 8
 parity = "N"
 stopBits = 1
 
 slaveId = 1
 timeout = 1

  [[inputs.modbus.Registers.InputRegisters.Tags]]
   name = "Voltage"
   order = "AB"
   datatype = "FLOAT32"
   scale = "/10"
   address = [
    0      
   ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
   name = "Current"
   order ="CDAB"
   datatype = "FLOAT32"
   scale = "/1000"
   address = [
    1,
    2
   ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Power"
    order = "CDAB"
    datatype = "FLOAT32"
    scale = "/10"
    address = [
     3,
     4      
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Energy"
    order = "CDAB"
    datatype = "FLOAT32"	
    scale = "/1000"
    address = [
     5,
     6      
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Frequency"
	order = "AB"	
	datatype = "FLOAT32"	    
    scale = "/10"
    address = [
     7
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
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
		if i + 1 == len(r) {
			n = -1
			if len(r) == 1 {
				n = 0
			}
		}
		if element+n == r[i+n] {
			chunk_t = append(chunk_t, element)
			if i + 1 == len(r) {
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
	m.Registers.DiscreteInputs.RawValues = make(map[int]uint16)
	m.Registers.Coils.RawValues = make(map[int]uint16)
	m.Registers.HoldingRegisters.RawValues = make(map[int]uint16)
	m.Registers.InputRegisters.RawValues = make(map[int]uint16)

	createRawValueMap(m.Registers.DiscreteInputs.Tags, m.Registers.DiscreteInputs.RawValues)
	createRawValueMap(m.Registers.Coils.Tags, m.Registers.Coils.RawValues)
	createRawValueMap(m.Registers.HoldingRegisters.Tags, m.Registers.HoldingRegisters.RawValues)
	createRawValueMap(m.Registers.InputRegisters.Tags, m.Registers.InputRegisters.RawValues)

	createChunks(&m.Registers.DiscreteInputs.Chunks, m.Registers.DiscreteInputs.RawValues)
	createChunks(&m.Registers.Coils.Chunks, m.Registers.Coils.RawValues)
	createChunks(&m.Registers.HoldingRegisters.Chunks, m.Registers.HoldingRegisters.RawValues)
	createChunks(&m.Registers.InputRegisters.Chunks, m.Registers.InputRegisters.RawValues)	
}

func connect(m *Modbus) error {
	switch m.Type {
	case "TCP":	
		m.HandlerTcp = mb.NewTCPClientHandler(m.Controller + ":" + strconv.Itoa(m.Port))
		m.HandlerTcp.Timeout = time.Duration(m.Timeout) * time.Second
		m.HandlerTcp.SlaveId = byte(m.SlaveId)
		m.Client = mb.NewClient(m.HandlerTcp)
		err := m.HandlerTcp.Connect()
		defer m.HandlerTcp.Close()
		if err != nil {
			return err
		}
		m.isConnected = true
		return nil
	case "RTU":		
		m.HandlerSerial = mb.NewRTUClientHandler(m.Controller)
		m.HandlerSerial.Timeout = time.Duration(m.Timeout) * time.Second
		m.HandlerSerial.SlaveId = byte(m.SlaveId)
		m.HandlerSerial.BaudRate = m.BaudRate
		m.HandlerSerial.DataBits = m.DataBits
		m.HandlerSerial.Parity = m.Parity
		m.HandlerSerial.StopBits = m.StopBits
		m.Client = mb.NewClient(m.HandlerSerial)
		err := m.HandlerSerial.Connect()
		defer m.HandlerSerial.Close()
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
		results, err := f(uint16(chunk_t.Address), uint16(chunk_t.Length))
		if err != nil {
			return err
		}
		
		if t == C_DIGITAL {
			for i := 0; i < len(results); i++ {				
				for b := 0; b < chunk_t.Length; b++ {				
					r[chunk_t.Address + b] = uint16(results[i] >> uint(b) & 0x01)
				}

			}
		}

		if t == C_ANALOG {
			for i := 0; i < len(results); i += 2 {
				register := uint16(results[i]) << 8 | uint16(results[i + 1])
				r[chunk_t.Address + i / 2] = uint16(register)				
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
		return uint32(binary.LittleEndian.Uint16(r[0:])) << 16 | uint32(binary.LittleEndian.Uint16(r[2:]))
	case "CDAB":
		return uint32(binary.BigEndian.Uint16(r[2:])) << 16 | uint32(binary.BigEndian.Uint16(r[0:]))
	default:
		return r
	}
}

func format16(f string, r uint16)  interface{}{
	switch f{
	case "UINT16":
		return r
	case "INT16":
		return int16(r)
	default:
		return r
	}
}

func format32(f string, r uint32)  interface{}{
	switch f{
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
		div, err := strconv.Atoi(s[1:len(s)])
		if err == nil {		
			return float32(v) / float32(div)
		}
		return 0
	case '*':
		div, err := strconv.Atoi(s[1:len(s)])
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
		div, err := strconv.Atoi(s[1:len(s)])
		if err == nil {		
			return float32(v) / float32(div)
		}
		return 0
	case '*':
		div, err := strconv.Atoi(s[1:len(s)])
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
			t[i].Value = r[t[i].Address[0]]
		}		
	}

	if s == C_ANALOG {
		for i := 0; i < len(t); i++ {		
			rawValues := []byte{}
			for _, rv := range t[i].Address {								
				rawValues = append(rawValues, byte(r[rv] >> 8) )
				rawValues = append(rawValues, byte(r[rv] & 0x00FF) )
			}
			
			switch t[i].DataType {
			case "UINT16":							
				e := convertEndianness16(t[i].Order, rawValues)
				e16, _ := e.(uint16)
				f := format16(t[i].DataType, e16)	
				f16 := f.(uint16)			
				s := scale16(t[i].Scale, f16)
				t[i].Value = s
			case "INT16":							
				e := convertEndianness16(t[i].Order, rawValues)
				e16, _ := e.(uint16)
				f := format16(t[i].DataType, e16)					
				t[i].Value = f
			case "UINT32":							
				e := convertEndianness32(t[i].Order, rawValues)
				e32, _ := e.(uint32)
				f := format32(t[i].DataType, e32)
				f32 := f.(uint32)
				s := scale32(t[i].Scale, f32)
				t[i].Value = s	
			case "INT32":							
				e := convertEndianness32(t[i].Order, rawValues)
				e32, _ := e.(uint32)
				f := format32(t[i].DataType, e32)				
				t[i].Value = f
			case "FLOAT32-IEEE":
				e := convertEndianness32(t[i].Order, rawValues)
				e32, _ := e.(uint32)
				f := format32(t[i].DataType, e32)				
				t[i].Value = f
			case "FLOAT32":
				if len(rawValues) == 2 {					
					e := convertEndianness16(t[i].Order, rawValues)
					e16, _ := e.(uint16)
					f := format16(t[i].DataType, e16)	
					f16 := f.(uint16)			
					s := scale16(t[i].Scale, f16)
					t[i].Value = s
				} else {					
					e := convertEndianness32(t[i].Order, rawValues)
					e32, _ := e.(uint32)										
					s := scale32(t[i].Scale, e32)
					t[i].Value = s	
				}
				
			default:
				t[i].Value = 0				
			}
		}
	}
}

func addFields(t []tag) map[string]interface{} {
	fields := make(map[string]interface{})
	for i:=0 ; i < len(t); i++ {
		if len(t[i].Name) > 0 {
			fields[t[i].Name] = t[i].Value
		} else {
			name := ""
			for _, e := range t[i].Address {
				name = name + "-" + strconv.Itoa(e)
			}
			name = name[:len(name) - 1]
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
			   m.Registers.DiscreteInputs.Chunks, 
			   m.Registers.DiscreteInputs.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_DIGITAL, 
			  m.Client.ReadCoils, 
			  m.Registers.Coils.Chunks, 
			  m.Registers.Coils.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_ANALOG, 
			  m.Client.ReadHoldingRegisters, 
			  m.Registers.HoldingRegisters.Chunks, 
			  m.Registers.HoldingRegisters.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_ANALOG, 
			  m.Client.ReadInputRegisters, 
			  m.Registers.InputRegisters.Chunks, 
			  m.Registers.InputRegisters.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	// Set tags
	setTag(C_DIGITAL, m.Registers.DiscreteInputs.Tags, m.Registers.DiscreteInputs.RawValues)
	setTag(C_DIGITAL, m.Registers.Coils.Tags, m.Registers.Coils.RawValues)
	setTag(C_ANALOG, m.Registers.HoldingRegisters.Tags, m.Registers.HoldingRegisters.RawValues)
	setTag(C_ANALOG, m.Registers.InputRegisters.Tags, m.Registers.InputRegisters.RawValues)

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
