package modbus

import (
	"testing"

	m "github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
	"github.com/tbrandon/mbserver"

	"github.com/influxdata/telegraf/testutil"
)

type mbTest struct {
	Name        string
	Address     uint16
	Quantity    uint16
	WriteValue  []byte
	ExpectValue interface{}
}

func TestCoils(t *testing.T) {
	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	defer serv.Close()
	assert.NoError(t, err)

	handler := m.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	assert.NoError(t, err)
	defer handler.Close()
	client := m.NewClient(handler)

	modbus := Modbus{
		Type:       "TCP",
		Controller: "localhost",
		Port:       1502,
		SlaveId:    1,
		//Timeout:    0,
		Registers: registers{
			Coils: register{
				Tags: []tag{
					{Name: "Coil0", Address: []int{0}},
					{Name: "Coil1", Address: []int{1}},
				},
			},
		},
	}

	coilTests := []mbTest{}
	coilTest := mbTest{Address: 0, Quantity: 2, WriteValue: []byte{0x03}, ExpectValue: []byte{0x03}}
	coilTests = append(coilTests, coilTest)
	coilTest = mbTest{Address: 0, Quantity: 2, WriteValue: []byte{0x02}, ExpectValue: []byte{0x02}}
	coilTests = append(coilTests, coilTest)
	coilTest = mbTest{Address: 0, Quantity: 2, WriteValue: []byte{0x01}, ExpectValue: []byte{0x01}}
	coilTests = append(coilTests, coilTest)
	coilTest = mbTest{Address: 0, Quantity: 2, WriteValue: []byte{0x00}, ExpectValue: []byte{0x00}}
	coilTests = append(coilTests, coilTest)

	for _, coilTest := range coilTests {
		_, err = client.WriteMultipleCoils(coilTest.Address, coilTest.Quantity, coilTest.WriteValue)
		assert.NoError(t, err)

		var acc testutil.Accumulator
		modbus.Gather(&acc)

		res := []byte{0x00}
		for i, t := range modbus.Registers.Coils.Tags {
			v, _ := t.value.(uint16)
			res[0] = res[0] | byte(v<<uint(i))
		}

		assert.Equal(t, coilTest.ExpectValue, res)
	}
}

func TestRegisters(t *testing.T) {
	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	defer serv.Close()
	assert.NoError(t, err)

	handler := m.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	assert.NoError(t, err)
	defer handler.Close()
	client := m.NewClient(handler)

	modbus := Modbus{
		Type:       "TCP",
		Controller: "localhost",
		Port:       1502,
		SlaveId:    1,
		//Timeout:    0,
		Registers: registers{
			HoldingRegisters: register{
				Tags: []tag{
					{
						Name:     "Register0",
						Order:    "AB",
						DataType: "FLOAT32",
						Address:  []int{0},
						Scale:    "/10",
					},
					{
						Name:     "Register1-2",
						Order:    "ABCD",
						DataType: "FLOAT32",
						Address:  []int{1, 2},
						Scale:    "/1000",
					},
					{
						Name:     "Register3-4",
						Order:    "ABCD",
						DataType: "FLOAT32",
						Address:  []int{3, 4},
						Scale:    "/10",
					},
					{
						Name:     "Register7",
						Order:    "AB",
						DataType: "FLOAT32",
						Address:  []int{7},
						Scale:    "/10",
					},
					{
						Name:     "Uint16AB",
						Order:    "AB",
						DataType: "UINT16",
						Address:  []int{10},
					},
					{
						Name:     "Uint16BA",
						Order:    "BA",
						DataType: "UINT16",
						Address:  []int{10},
					},
					{
						Name:     "Int16AB",
						Order:    "AB",
						DataType: "INT16",
						Address:  []int{10},
					},
					{
						Name:     "Int16BA",
						Order:    "BA",
						DataType: "INT16",
						Address:  []int{10},
					},
					{
						Name:     "Int32ABCD",
						Order:    "ABCD",
						DataType: "INT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Int32DCBA",
						Order:    "DCBA",
						DataType: "INT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Int32BADC",
						Order:    "BADC",
						DataType: "INT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Int32CDAB",
						Order:    "CDAB",
						DataType: "INT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Uint32ABCD",
						Order:    "ABCD",
						DataType: "UINT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Uint32DCBA",
						Order:    "DCBA",
						DataType: "UINT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Uint32BADC",
						Order:    "BADC",
						DataType: "UINT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "Uint32CDAB",
						Order:    "CDAB",
						DataType: "UINT32",
						Address:  []int{10, 11},
					},
					{
						Name:     "FLOAT32-IEEE",
						Order:    "ABCD",
						DataType: "FLOAT32-IEEE",
						Address:  []int{10, 11},
					},
				},
			},
		},
	}

	regTests := []mbTest{}
	regTest := mbTest{Name: "Register0", Address: 0, Quantity: 1, WriteValue: []byte{0x08, 0x98}, ExpectValue: float32(220)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Register1-2", Address: 1, Quantity: 2, WriteValue: []byte{0x00, 0x00, 0x03, 0xE8}, ExpectValue: float32(1)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Register3-4", Address: 3, Quantity: 2, WriteValue: []byte{0x00, 0x00, 0x08, 0x98}, ExpectValue: float32(220)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Register7", Address: 7, Quantity: 1, WriteValue: []byte{0x01, 0xF4}, ExpectValue: float32(50)}
	regTests = append(regTests, regTest)

	regTest = mbTest{Name: "Uint16AB", Address: 10, Quantity: 1, WriteValue: []byte{0xAB, 0xCD}, ExpectValue: uint16(43981)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Uint16BA", Address: 10, Quantity: 1, WriteValue: []byte{0xAB, 0xCD}, ExpectValue: uint16(52651)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Int16AB", Address: 10, Quantity: 1, WriteValue: []byte{0xAB, 0xCD}, ExpectValue: int16(-21555)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Int16BA", Address: 10, Quantity: 1, WriteValue: []byte{0xAB, 0xCD}, ExpectValue: int16(-12885)}
	regTests = append(regTests, regTest)

	regTest = mbTest{Name: "Int32ABCD", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: int32(-1430532899)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Int32DCBA", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: int32(-573785174)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Int32BADC", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: int32(-1146430004)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Int32CDAB", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: int32(-857888069)}
	regTests = append(regTests, regTest)

	regTest = mbTest{Name: "Uint32ABCD", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: uint32(2864434397)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Uint32DCBA", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: uint32(3721182122)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Uint32BADC", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: uint32(3148537292)}
	regTests = append(regTests, regTest)
	regTest = mbTest{Name: "Uint32CDAB", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: uint32(3437079227)}
	regTests = append(regTests, regTest)

	regTest = mbTest{Name: "FLOAT32-IEEE", Address: 10, Quantity: 2, WriteValue: []byte{0xAA, 0xBB, 0xCC, 0xDD}, ExpectValue: float32(-3.3360025e-13)}
	regTests = append(regTests, regTest)

	for _, regTest := range regTests {
		_, err = client.WriteMultipleRegisters(regTest.Address, regTest.Quantity, regTest.WriteValue)
		assert.NoError(t, err)

		var acc testutil.Accumulator
		modbus.Gather(&acc)

		for _, tags := range modbus.Registers.HoldingRegisters.Tags {
			if tags.Name == regTest.Name {
				assert.Equal(t, regTest.ExpectValue, tags.value)
			}
		}
	}
}
