package modbus

import (
	"encoding/binary"
	"testing"
)

var test Modbus

func update(test *Modbus) {
	test.Client = "localhost:502"
	test.SlaveAddress = 1
	test.FunctionCode = 0
	test.Address = 1
	test.Quantity = 1
	test.TimeOut = 5
}

// Testing Code for TCP Client
func TestGetTCPdataReadCoils(t *testing.T) {
	t.Log("Testing TCPClient Connections: Read Coils")
	update(&test)
	test.FunctionCode = 1

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if results[0] != 1 {
		t.Errorf("Expected value of 1 @ modbus address 00001, but it was %d instead", results[0])
	}
}

func TestGetTCPdataReadDiscreteInputs(t *testing.T) {
	t.Log("Testing TCPClient Connections: Read Discrete Inputs")
	update(&test)
	test.FunctionCode = 2

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if results[0] != 1 {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", results[0])
	}
}

func TestGetTCPdataReadHoldingRegister(t *testing.T) {
	t.Log("Testing TCPClient Connections: Read Holding Registers")
	update(&test)
	test.FunctionCode = 3

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 40001, but it was %d instead", binary.BigEndian.Uint16(results))
	}
}

func TestGetTCPdataReadInputRegister(t *testing.T) {
	t.Log("Testing TCPClient Connections: Read Input Register")
	update(&test)
	test.FunctionCode = 4

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 30001, but it was %d instead", binary.BigEndian.Uint16(results))
	}
}

func TestGetTCPdataWriteSingleCoil(t *testing.T) {
	t.Log("Testing TCPClient Connections: Write Single Coil")
	update(&test)
	test.FunctionCode = 5
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.LittleEndian.Uint16(results) != 255 {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func TestGetTCPdataWriteSingleRegister(t *testing.T) {
	t.Log("Testing TCPClient Connections: Write Single Register")
	update(&test)
	test.FunctionCode = 6
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.LittleEndian.Uint16(results) != 255 {
		t.Errorf("Expected value of 255 @ modbus address 30001, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func TestGetTCPdataWriteMultipleCoils(t *testing.T) {
	t.Log("Testing TCPClient Connections: Write Single Register")
	update(&test)
	test.FunctionCode = 15

	test.Address = 0
	test.Quantity = 32
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55}

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != test.Quantity {
		t.Errorf("Expected value of 32 @ modbus address 10000, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func TestGetTCPdataWriteMultipleRegisters(t *testing.T) {
	t.Log("Testing TCPClient Connections: Write Multiple Register")
	update(&test)
	test.FunctionCode = 16
	test.Address = 0
	test.Quantity = 3
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55, 0xFF, 0xFF}

	results, err := getTCPdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != test.Quantity {
		t.Errorf("Expected value of 3 @ modbus address 40001, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func update2(test *Modbus) {
	test.Client = "COM3"
	test.SlaveAddress = 1
	test.FunctionCode = 0
	test.Address = 0
	test.Quantity = 1
	test.TimeOut = 5
	//serial connection
	test.Comm.BaudRate = 9600
	test.Comm.Databits = 8
	test.Comm.Parity = "N"
	test.Comm.Stopbits = 1
}

// Testing Code for RTU Client
func TestGetRTUdataReadCoils(t *testing.T) {
	t.Log("Testing RTUClient Connections: Read Coils")
	update2(&test)
	test.FunctionCode = 1

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if results[0] != 1 {
		t.Errorf("Expected value of 1 @ modbus address 00001, but it was %d instead", results[0])
	}
}

func TestGetRTUdataReadDiscreteInputs(t *testing.T) {
	t.Log("Testing RTUClient Connections: Read Discrete Inputs")
	update2(&test)
	test.FunctionCode = 2

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if results[0] != 1 {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", results[0])
	}
}

func TestGetRTUdataReadHoldingRegister(t *testing.T) {
	t.Log("Testing RTUClient Connections: Read Holding Registers")
	update2(&test)
	test.FunctionCode = 3

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 40001, but it was %d instead", binary.BigEndian.Uint16(results))
	}
}

func TestGetRTUdataReadInputRegister(t *testing.T) {
	t.Log("Testing RTUClient Connections: Read Input Register")
	update2(&test)
	test.FunctionCode = 4

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 30001, but it was %d instead", binary.BigEndian.Uint16(results))
	}
}

func TestGetRTUdataWriteSingleCoil(t *testing.T) {
	t.Log("Testing RTUClient Connections: Write Single Coil")
	update2(&test)
	test.FunctionCode = 5
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.LittleEndian.Uint16(results) != 255 {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func TestGetRTUdataWriteSingleRegister(t *testing.T) {
	t.Log("Testing RTUClient Connections: Write Single Register")
	update2(&test)
	test.FunctionCode = 6
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.LittleEndian.Uint16(results) != 255 {
		t.Errorf("Expected value of 255 @ modbus address 30001, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func TestGetRTUdataWriteMultipleCoils(t *testing.T) {
	t.Log("Testing RTUClient Connections: Write Single Register")
	update2(&test)
	test.FunctionCode = 15

	test.Address = 0
	test.Quantity = 32
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55}

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != test.Quantity {
		t.Errorf("Expected value of 32 @ modbus address 10000, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}

func TestGetRTUdataWriteMultipleRegisters(t *testing.T) {
	t.Log("Testing RTUClient Connections: Write Multiple Register")
	update2(&test)
	test.FunctionCode = 16
	test.Address = 0
	test.Quantity = 3
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55, 0xFF, 0xFF}

	results, err := getRTUdata(&test)
	t.Log(results)
	if err != nil || results == nil {
		t.Fatal(err, results)
	}
	if binary.BigEndian.Uint16(results) != test.Quantity {
		t.Errorf("Expected value of 3 @ modbus address 40001, but it was %d instead", binary.LittleEndian.Uint16(results))
	}
}
