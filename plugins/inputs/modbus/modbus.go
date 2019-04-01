package modbus

import (
	"errors"
	"sort"
	"strconv"
	"time"
	
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
	RawValues map[int]int
}

type tag struct {
	Name    string
	Order   string	
	Scale   string
	Address []int
	Value   interface{}
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
   order ="AB"	
   scale = "/10"
   address = [
    0      
   ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
   name = "Current"
   order ="CDAB"	
   scale = "/1000"
   address = [
    1,
    2
   ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Power"
    order ="CDAB"	
    scale = "/10"
    address = [
     3,
     4      
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Energy"
    order ="CDAB"	
    scale = "/1000"
    address = [
     5,
     6      
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "Frecuency"
    order ="AB"	
    scale = "/10"
    address = [
     7
    ]

  [[inputs.modbus.Registers.InputRegisters.Tags]]
    name = "PowerFactor"
    order ="AB"	
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

func createRawValueMap(r []tag, rawValue map[int]int) {
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

func createChunks(ch *[]chunk, rawValue map[int]int) {
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
	m.Registers.DiscreteInputs.RawValues = make(map[int]int)
	m.Registers.Coils.RawValues = make(map[int]int)
	m.Registers.HoldingRegisters.RawValues = make(map[int]int)
	m.Registers.InputRegisters.RawValues = make(map[int]int)

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
func getRawValue(t string, f fn, ch []chunk, r map[int]int) error {
	for _, chunk_t := range ch {		
		results, err := f(uint16(chunk_t.Address), uint16(chunk_t.Length))
		if err != nil {
			return err
		}

		if t == C_DIGITAL {
			for i := 0; i < len(results); i++ {				
				for b := 0; b < chunk_t.Length; b++ {				
					r[chunk_t.Address + b] = int(results[i] >> uint(b) & 0x01)
				}

			}
		}

		if t == C_ANALOG {
			for i := 0; i < len(results); i += 2 {
				register := uint16(results[i]) << 8 | uint16(results[i + 1])
				r[chunk_t.Address + i / 2] = int(register)
			}
		}
	}

	return nil
}

func BA(x int) int {
	return int(x & 0xFF00 >> 8 | x & 0x00FF << 8)
}

func joinBytes(x int, s uint, y int) int {
	return int(x << s | y)
}

func convertEndianness(o string, r []int) int {
	switch o {
	case "AB":
		return int(r[0])
	case "BA":
		return BA(r[0])
	case "ABCD":
		return joinBytes(int(r[0]), 16, int(r[1]))
	case "CDAB":
		return joinBytes(int(r[1]), 16, int(r[0]))
	case "BADC":
		return joinBytes(BA(r[0]), 16, BA(r[1]))
	case "DCBA":
		return joinBytes(BA(r[1]), 16, BA(r[0]))
	default:
		return int(r[0])
	}
}

func scale(s string, v int) interface{} {
	if len(s) == 0 {
		return 0
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
			return v * int(div)
		}
		return 0
	default:
		return 0
	}
}

func setTag(s string, t []tag, r map[int]int) {
	if s == C_DIGITAL {
		for i := 0; i < len(t); i++ {
			t[i].Value = r[t[i].Address[0]]
		}		
	}

	if s == C_ANALOG {
		for i := 0; i < len(t); i++ {		
			rawValues := []int{}
			for _, rv := range t[i].Address {				
				rawValues = append(rawValues, r[rv])
			}

			t[i].Value = scale(t[i].Scale, convertEndianness(t[i].Order, rawValues))
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
	err := getRawValue(C_DIGITAL, m.Client.ReadDiscreteInputs, m.Registers.DiscreteInputs.Chunks, m.Registers.DiscreteInputs.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_DIGITAL, m.Client.ReadCoils, m.Registers.Coils.Chunks, m.Registers.Coils.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_ANALOG, m.Client.ReadHoldingRegisters, m.Registers.HoldingRegisters.Chunks, m.Registers.HoldingRegisters.RawValues)
	if err != nil {
		m.isConnected = false
		return err
	}

	err = getRawValue(C_ANALOG, m.Client.ReadInputRegisters, m.Registers.InputRegisters.Chunks, m.Registers.InputRegisters.RawValues)
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