package modbus

import (
	"encoding/binary"
	"testing"
)

var test Modbus

func update(test *Modbus) {
	test.Client = "127.0.0.1:502"
	test.SlaveAddress = 1
	test.FunctionCode = 0
	test.Address = 0
	test.Quantity = 1
	test.TimeOut = 5
}

// Testing Code for TCP Client
func TestGetTCPdataReadCoils(t *testing.T) {
	// Note: Address may be offset by 1
	// Set bit @ address to query
	t.Log("Testing TCPClient Connections: Read Coils")
	update(&test)
	test.FunctionCode = 1

	test.err = test.getTCPdata()

	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if test.Results == nil {
		t.Errorf("Expected value of 1 @ modbus address 00001, but it was %d instead", test.Results[0])
	}
}

func TestGetTCPdataReadDiscreteInputs(t *testing.T) {
	// Note: Address may be offset by 1
	// Set bits @ address to query
	// Results will be in Byte format not binary
	t.Log("Testing TCPClient Connections: Read Discrete Inputs")
	update(&test)
	test.FunctionCode = 2
	test.Address = 0
	test.Quantity = 3

	test.err = test.getTCPdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if test.Results == nil {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", test.Results)
	}
}

func TestGetTCPdataReadHoldingRegister(t *testing.T) {
	// Note: Address may be offset by 1
	// Set HoldingRegister address to xFF FF or b1111 1111 1111 1111 or d65535
	t.Log("Testing TCPClient Connections: Read Holding Registers")
	update(&test)
	test.FunctionCode = 3

	test.err = test.getTCPdata()

	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 40001, but it was %d instead", binary.BigEndian.Uint16(test.Results))
	}
}

func TestGetTCPdataReadInputRegister(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InpuRegister address to xFF FF or b1111 1111 1111 1111 or d65535
	t.Log("Testing TCPClient Connections: Read Input Register")
	update(&test)
	test.FunctionCode = 4

	test.err = test.getTCPdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 30001, but it was %d instead", binary.BigEndian.Uint16(test.Results))
	}
}

func TestGetTCPdataWriteSingleCoil(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputCoils address to xFF or b1111 1111  or d255
	t.Log("Testing TCPClient Connections: Write Single Coil")
	update(&test)
	test.FunctionCode = 5
	test.Address = 0
	test.Values = []byte{0xff, 0x00}

	test.err = test.getTCPdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.LittleEndian.Uint16(test.Results) != 255 {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}

func TestGetTCPdataWriteSingleRegister(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputRegister address to xFF or b1111 1111  or d255
	t.Log("Testing TCPClient Connections: Write Single Register")
	update(&test)
	test.FunctionCode = 6
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	test.err = test.getTCPdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.LittleEndian.Uint16(test.Results) != 255 {
		t.Errorf("Expected value of 255 @ modbus address 30001, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}

func TestGetTCPdataWriteMultipleCoils(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputRegister addresses 0-15  to xAA AA or b1010 1010 1010 1010  or d43690
	// Set InputRegister addresses 16-32 to x55 55 or b0101 0101 0101 0101  or d21845
	t.Log("Testing TCPClient Connections: Write Single Register")
	update(&test)
	test.FunctionCode = 15

	test.Address = 0
	test.Quantity = 32
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55}

	test.err = test.getTCPdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != test.Quantity {
		t.Errorf("Expected value of 32 @ modbus address 10000, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}

func TestGetTCPdataWriteMultipleRegisters(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputRegister addresses 0 to xAA AA or b1010 1010 1010 1010  or d43690
	// Set InputRegister addresses 1 to x55 55 or b0101 0101 0101 0101  or d21845
	// Set InputRegister addresses 2 to xFF FF or b1111 1111 1111 1111  or d65535
	t.Log("Testing TCPClient Connections: Write Multiple Register")
	update(&test)
	test.FunctionCode = 16
	test.Address = 0
	test.Quantity = 3
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55, 0xFF, 0xFF}

	test.err = test.getTCPdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != test.Quantity {
		t.Errorf("Expected value of 3 @ modbus address 40001, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
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
	// Note: Address may be offset by 1
	// Set bit @ address to query
	t.Log("Testing RTUClient Connections: Read Coils")
	update2(&test)
	test.FunctionCode = 1

	test.err = test.getRTUdata()

	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if test.Results == nil {
		t.Errorf("Expected value of 1 @ modbus address 00001, but it was %d instead", test.Results[0])
	}
}

func TestGetRTUdataReadDiscreteInputs(t *testing.T) {
	// Note: Address may be offset by 1
	// Set bits @ address to query
	// Results will be in Byte format not binary
	t.Log("Testing RTUClient Connections: Read Discrete Inputs")
	update2(&test)
	test.FunctionCode = 2
	test.Address = 0
	test.Quantity = 3

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if test.Results == nil {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", test.Results[0])
	}
}

func TestGetRTUdataReadHoldingRegister(t *testing.T) {
	// Note: Address may be offset by 1
	// Set HoldingRegister address to xFF FF or b1111 1111 1111 1111 or d65535
	t.Log("Testing RTUClient Connections: Read Holding Registers")
	update2(&test)
	test.FunctionCode = 3

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 40001, but it was %d instead", binary.BigEndian.Uint16(test.Results))
	}
}

func TestGetRTUdataReadInputRegister(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InpuRegister address to xFF FF or b1111 1111 1111 1111 or d65535
	t.Log("Testing RTUClient Connections: Read Input Register")
	update2(&test)
	test.FunctionCode = 4

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != 65535 {
		t.Errorf("Expected value of 65535 @ modbus address 30001, but it was %d instead", binary.BigEndian.Uint16(test.Results))
	}
}

func TestGetRTUdataWriteSingleCoil(t *testing.T) {
		// Note: Address may be offset by 1
	// Set InputCoils address to xFF or b1111 1111  or d255
	t.Log("Testing RTUClient Connections: Write Single Coil")
	update2(&test)
	test.FunctionCode = 5
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.LittleEndian.Uint16(test.Results) != 255 {
		t.Errorf("Expected value of 1 @ modbus address 10001, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}

func TestGetRTUdataWriteSingleRegister(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputRegister address to xFF or b1111 1111  or d255
	t.Log("Testing RTUClient Connections: Write Single Register")
	update2(&test)
	test.FunctionCode = 6
	test.Address = 1
	test.Values = []byte{0xff, 0x00}

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.LittleEndian.Uint16(test.Results) != 255 {
		t.Errorf("Expected value of 255 @ modbus address 30001, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}

func TestGetRTUdataWriteMultipleCoils(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputRegister addresses 0-15  to xAA AA or b1010 1010 1010 1010  or d43690
	// Set InputRegister addresses 16-32 to x55 55 or b0101 0101 0101 0101  or d21845
	t.Log("Testing RTUClient Connections: Write Single Register")
	update2(&test)
	test.FunctionCode = 15

	test.Address = 0
	test.Quantity = 32
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55}

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != test.Quantity {
		t.Errorf("Expected value of 32 @ modbus address 10000, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}

func TestGetRTUdataWriteMultipleRegisters(t *testing.T) {
	// Note: Address may be offset by 1
	// Set InputRegister addresses 0 to xAA AA or b1010 1010 1010 1010  or d43690
	// Set InputRegister addresses 1 to x55 55 or b0101 0101 0101 0101  or d21845
	// Set InputRegister addresses 2 to xFF FF or b1111 1111 1111 1111  or d65535
	t.Log("Testing RTUClient Connections: Write Multiple Register")
	update2(&test)
	test.FunctionCode = 16
	test.Address = 0
	test.Quantity = 3
	//BigEndian = two bytes per register
	test.Values = []byte{0xAA, 0xAA, 0x55, 0x55, 0xFF, 0xFF}

	test.err = test.getRTUdata()
	t.Log(test.Results)
	if test.err != nil || test.Results == nil {
		t.Fatal(test.err, test.Results)
	}
	if binary.BigEndian.Uint16(test.Results) != test.Quantity {
		t.Errorf("Expected value of 3 @ modbus address 40001, but it was %d instead", binary.LittleEndian.Uint16(test.Results))
	}
}
