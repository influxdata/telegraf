package modbus

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	modbus "github.com/goburrow/modbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Modbus library
type Modbus struct {
	Client       string
	SlaveAddress uint8
	FunctionCode uint8
	Address      uint16
	Quantity     uint16
	Values       []byte
	TimeOut      uint8
	Comm         serial
	Results      []byte
	err          error
	CX           modbusClients
}

type serial struct {
	BaudRate int
	Databits int
	Parity   string
	Stopbits int
}

type modbusClients struct {
	connected  bool
	TCPhandler *modbus.TCPClientHandler
	RTUhandler *modbus.RTUClientHandler
	client     modbus.Client
}

// ModbusConfig example
var sampleConfig = `
	## Set Modbus Config (Either TCP or RTU Client)
	## Modbust TCP Client
	## TCP Client = "localhost:502"
	Client = "localhost:502"

	## Modbus RTU Client
	## RTU Client = "/dev/ttyS0"
	## serial setup for RTUClient
	# serial = [11520,8,"N",1]

	## Call to device
	SlaveAddress = 1

	## Function Code to Device
	FunctionCode = 1

	## Device Memory Address
	Address = 1

	## Quantity of Values to read/write
	Quantity = 1

	## Array of values to write
	# Values = [0]

	# Timeout in seconds
	TimeOut = 5
`

// SampleConfig of modbus plugin
func (s *Modbus) SampleConfig() string {
	return sampleConfig
}

// Description of modbus plugin
func (s *Modbus) Description() string {
	return "Fault-tolerant, fail-fast implementation of Modbus protocol in Go."
}

// Gather TCP or RTU modbus parameters
func (s *Modbus) Gather(acc telegraf.Accumulator) (err error) {

	var inits bool
	// Modbus TCP
	if strings.Contains(s.Client, ":") {
		inits = true
		err = s.getTCPdata()

	} else if strings.ContainsAny(s.Client, "/") {
		inits = true
		err = s.getRTUdata()
	} else {
		inits = false
	}

	fields := make(map[string]interface{})

	if inits {
		field := string(s.FunctionCode)
		if s.Quantity == 1 {

			field += fmt.Sprintf("%04d", s.Address)
			fields[field] = s.Results[0]

		} else {
			for i := 1; i <= int(s.Quantity); i++ {

				offset := i - 1
				field1 := field + fmt.Sprintf("%04d", int(s.Address)+offset)
				fields[field1] = s.Results[offset]

			}
		}
	}

	tags := make(map[string]string)

	acc.AddFields("modbus", fields, tags)

	return nil
}

func (s *Modbus) createTCPClient() error {
	if s.CX.TCPhandler == nil && !s.CX.connected {
		s.CX.TCPhandler = modbus.NewTCPClientHandler(s.Client)
		s.CX.TCPhandler.Timeout = time.Duration(s.TimeOut) * time.Second
		s.CX.TCPhandler.SlaveId = s.SlaveAddress

		s.CX.TCPhandler.Logger = log.New(os.Stdout, "TCP: ", log.LstdFlags)

		s.err = s.CX.TCPhandler.Connect()
		if s.err != nil {
			s.CX.connected = false
			return s.err
		}

		defer s.CX.TCPhandler.Close()

		if s.CX.client == nil && !s.CX.connected {
			s.CX.client = modbus.NewClient(s.CX.TCPhandler)
			s.CX.connected = true
		}
	}

	return s.err
}

func (s *Modbus) getTCPdata() (err error) {
	// create TCPClient connection if not already done
	s.err = s.createTCPClient()
	if s.err != nil {
		return s.err
	}
	// Function Codes
	switch s.FunctionCode {
	// FC01 = Read Coil Status
	case 1:
		s.Results, s.err = s.CX.client.ReadCoils(s.Address, s.Quantity)

	// FC02 = Read Input Status
	case 2:
		s.Results, s.err = s.CX.client.ReadDiscreteInputs(s.Address, s.Quantity)

	// FC03 = Read Holding Registers
	case 3:
		s.Results, s.err = s.CX.client.ReadHoldingRegisters(s.Address, s.Quantity)

	// FC04 = Read Input Registers
	case 4:
		s.Results, s.err = s.CX.client.ReadInputRegisters(s.Address, s.Quantity)

	// FC05 = Write Single Coil
	case 5:
		s.Results, s.err = s.CX.client.WriteSingleCoil(s.Address, binary.BigEndian.Uint16(s.Values))

		// FC06	= Write Single Register
	case 6:
		s.Results, s.err = s.CX.client.WriteSingleRegister(s.Address, binary.BigEndian.Uint16(s.Values))

	// FC15 = Write Multiple Coils
	case 15:
		s.Results, s.err = s.CX.client.WriteMultipleCoils(s.Address, s.Quantity, s.Values)

	// FC16 = Write Multiple Registers
	case 16:
		s.Results, s.err = s.CX.client.WriteMultipleRegisters(s.Address, s.Quantity, s.Values)

	default:
		//do nothing
	}

	return s.err
}

func (s *Modbus) createRTUClient() (err error) {

	if s.CX.RTUhandler == nil && !s.CX.connected {
		s.CX.RTUhandler = modbus.NewRTUClientHandler(s.Client)
		s.CX.RTUhandler.BaudRate = s.Comm.BaudRate
		s.CX.RTUhandler.DataBits = s.Comm.Databits
		s.CX.RTUhandler.Parity = s.Comm.Parity
		s.CX.RTUhandler.StopBits = s.Comm.Stopbits
		s.CX.RTUhandler.SlaveId = s.SlaveAddress
		s.CX.RTUhandler.Timeout = time.Duration(s.TimeOut) * time.Second

		s.CX.RTUhandler.Logger = log.New(os.Stdout, "RTU: ", log.LstdFlags)

		s.err = s.CX.RTUhandler.Connect()
		if s.err != nil {
			s.CX.connected = false
			return s.err
		}
		defer s.CX.RTUhandler.Close()

		if s.CX.client == nil && !s.CX.connected {
			s.CX.client = modbus.NewClient(s.CX.RTUhandler)
			s.CX.connected = true
		}
	}

	return nil
}

func (s *Modbus) getRTUdata() (err error) {
	// create RTUClient connection if not already done
	s.err = s.createRTUClient()
	if s.err != nil {
		return s.err
	}
	// Function Codes
	switch s.FunctionCode {
	// FC01 = Read Coil Status
	case 1:
		s.Results, s.err = s.CX.client.ReadCoils(s.Address, s.Quantity)

	// FC02 = Read Input Status
	case 2:
		s.Results, s.err = s.CX.client.ReadDiscreteInputs(s.Address, s.Quantity)

	// FC03 = Read Holding Registers
	case 3:
		s.Results, s.err = s.CX.client.ReadHoldingRegisters(s.Address, s.Quantity)

	// FC04 = Read Input Registers
	case 4:
		s.Results, s.err = s.CX.client.ReadInputRegisters(s.Address, s.Quantity)

	// FC05 = Write Single Coil
	case 5:
		s.Results, s.err = s.CX.client.WriteSingleCoil(s.Address, binary.BigEndian.Uint16(s.Values))

		// FC06	= Write Single Register
	case 6:
		s.Results, s.err = s.CX.client.WriteSingleRegister(s.Address, binary.BigEndian.Uint16(s.Values))

	// FC15 = Write Multiple Coils
	case 15:
		s.Results, s.err = s.CX.client.WriteMultipleCoils(s.Address, s.Quantity, s.Values)

	// FC16 = Write Multiple Registers
	case 16:
		s.Results, s.err = s.CX.client.WriteMultipleRegisters(s.Address, s.Quantity, s.Values)

	default:
		//do nothing
	}

	return err
}

func init() {

	inputs.Add("modbus", func() telegraf.Input {
		return &Modbus{}
	})
}
