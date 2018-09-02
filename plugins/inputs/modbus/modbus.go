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
}

type serial struct {
	BaudRate int
	Databits int
	Parity   string
	Stopbits int
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

func (s *Modbus) SampleConfig() string {
	return sampleConfig
}

func (s *Modbus) Description() string {
	return "Fault-tolerant, fail-fast implementation of Modbus protocol in Go."
}

func (s *Modbus) Gather(acc telegraf.Accumulator) error {

	var inits bool
	// Modbus TCP
	if strings.Contains(s.Client, ":") {
		inits = true
		s.Results, s.err = getTCPdata(s)

	} else if strings.ContainsAny(s.Client, "/") {
		inits = true
		s.Results, s.err = getRTUdata(s)
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

func getTCPdata(s *Modbus) (results []byte, err error) {

	handler := modbus.NewTCPClientHandler(s.Client)
	handler.Timeout = time.Duration(s.TimeOut) * time.Second
	handler.SlaveId = s.SlaveAddress
	handler.Logger = log.New(os.Stdout, "TCP: ", log.LstdFlags)

	err = handler.Connect()
	if err != nil {
		return []byte{}, err
	}
	defer handler.Close()

	client := modbus.NewClient(handler)
	// Function Codes
	switch s.FunctionCode {
	// FC01 = Read Coil Status
	case 1:
		results, err = client.ReadCoils(s.Address, s.Quantity)

	// FC02 = Read Input Status
	case 2:
		results, err = client.ReadDiscreteInputs(s.Address, s.Quantity)

	// FC03 = Read Holding Registers
	case 3:
		results, err = client.ReadHoldingRegisters(s.Address, s.Quantity)

	// FC04 = Read Input Registers
	case 4:
		results, err = client.ReadInputRegisters(s.Address, s.Quantity)

	// FC05 = Write Single Coil
	case 5:
		results, err = client.WriteSingleCoil(s.Address, binary.BigEndian.Uint16(s.Values))

		// FC06	= Write Single Register
	case 6:
		results, err = client.WriteSingleRegister(s.Address, binary.BigEndian.Uint16(s.Values))

	// FC15 = Write Multiple Coils
	case 15:
		results, err = client.WriteMultipleCoils(s.Address, s.Quantity, s.Values)

	// FC16 = Write Multiple Registers
	case 16:
		results, err = client.WriteMultipleRegisters(s.Address, s.Quantity, s.Values)

	default:
		//do nothing
	}

	return results, err
}

func getRTUdata(s *Modbus) (results []byte, err error) {

	handler := modbus.NewRTUClientHandler(s.Client)
	handler.BaudRate = s.Comm.BaudRate
	handler.DataBits = s.Comm.Databits
	handler.Parity = s.Comm.Parity
	handler.StopBits = s.Comm.Stopbits
	handler.SlaveId = s.SlaveAddress
	handler.Timeout = time.Duration(s.TimeOut) * time.Second
	
	handler.Logger = log.New(os.Stdout, "RTU: ", log.LstdFlags)

	err = handler.Connect()
	if err != nil {
		return []byte{}, err
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	// Function Codes
	switch s.FunctionCode {
	// FC01 = Read Coil Status
	case 1:
		results, err = client.ReadCoils(s.Address, s.Quantity)

	// FC02 = Read Input Status
	case 2:
		results, err = client.ReadDiscreteInputs(s.Address, s.Quantity)

	// FC03 = Read Holding Registers
	case 3:
		results, err = client.ReadHoldingRegisters(s.Address, s.Quantity)

	// FC04 = Read Input Registers
	case 4:
		results, err = client.ReadInputRegisters(s.Address, s.Quantity)

	// FC05 = Write Single Coil
	case 5:
		results, err = client.WriteSingleCoil(s.Address, binary.BigEndian.Uint16(s.Values))

		// FC06	= Write Single Register
	case 6:
		results, err = client.WriteSingleRegister(s.Address, binary.BigEndian.Uint16(s.Values))

	// FC15 = Write Multiple Coils
	case 15:
		results, err = client.WriteMultipleCoils(s.Address, s.Quantity, s.Values)

	// FC16 = Write Multiple Registers
	case 16:
		results, err = client.WriteMultipleRegisters(s.Address, s.Quantity, s.Values)

	default:
		//do nothing
	}

	return results, err
}

func init() {

	inputs.Add("modbus", func() telegraf.Input {
		return &Modbus{}
	})
}
