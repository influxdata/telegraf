package modbus

import (
	"testing"

	m "github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
	"github.com/tbrandon/mbserver"

	"github.com/influxdata/telegraf/testutil"
)

func TestCoils(t *testing.T) {
	var coilTests = []struct {
		name     string
		address  uint16
		quantity uint16
		write    []byte
		read     uint16
	}{
		{
			name:     "Coil0-TurnOff",
			address:  0,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
		{
			name:     "Coil0-TurnOn",
			address:  0,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "Coil1-TurnOn",
			address:  1,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "Coil2-TurnOn",
			address:  2,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "Coil3-TurnOn",
			address:  3,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "Coil1-TurnOff",
			address:  1,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
		{
			name:     "Coil2-TurnOff",
			address:  2,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
		{
			name:     "Coil3-TurnOff",
			address:  3,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
	}

	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	defer serv.Close()
	assert.NoError(t, err)

	handler := m.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	assert.NoError(t, err)
	defer handler.Close()
	client := m.NewClient(handler)

	for _, ct := range coilTests {
		t.Run(ct.name, func(t *testing.T) {
			_, err = client.WriteMultipleCoils(ct.address, ct.quantity, ct.write)
			assert.NoError(t, err)

			modbus := Modbus{
				Controller: "tcp://localhost:1502",
				Slave_Id:   1,
				Coils: []tag{
					{
						Address: []uint16{ct.address},
					},
				},
			}

			var acc testutil.Accumulator
			modbus.Gather(&acc)

			for _, coil := range modbus.registers {
				assert.Equal(t, ct.read, coil.Tags[0].value)
			}
		})
	}
}

func TestHoldingRegisters(t *testing.T) {
	var holdingRegisterTests = []struct {
		name       string
		address    []uint16
		quantity   uint16
		byte_order string
		data_type  string
		scale      string
		write      []byte
		read       interface{}
	}{
		{
			name:       "Register0-AB-FLOAT32",
			address:    []uint16{0},
			quantity:   1,
			byte_order: "AB",
			data_type:  "FLOAT32",
			scale:      "0.1",
			write:      []byte{0x08, 0x98},
			read:       float32(220),
		},
		{
			name:       "Register0-Register1-AB-FLOAT32",
			address:    []uint16{0, 1},
			quantity:   2,
			byte_order: "ABCD",
			data_type:  "FLOAT32",
			scale:      "0.001",
			write:      []byte{0x00, 0x00, 0x03, 0xE8},
			read:       float32(1),
		},
		{
			name:       "Register1-Register2-ABCD-FLOAT32",
			address:    []uint16{1, 2},
			quantity:   2,
			byte_order: "ABCD",
			data_type:  "FLOAT32",
			scale:      "0.1",
			write:      []byte{0x00, 0x00, 0x08, 0x98},
			read:       float32(220),
		},
		{
			name:       "Register3-Register4-ABCD-FLOAT32",
			address:    []uint16{3, 4},
			quantity:   2,
			byte_order: "ABCD",
			data_type:  "FLOAT32",
			scale:      "0.1",
			write:      []byte{0x00, 0x00, 0x08, 0x98},
			read:       float32(220),
		},
		{
			name:       "Register7-AB-FLOAT32",
			address:    []uint16{7},
			quantity:   1,
			byte_order: "AB",
			data_type:  "FLOAT32",
			scale:      "0.1",
			write:      []byte{0x01, 0xF4},
			read:       float32(50),
		},
		{
			name:       "Register10-AB-UINT16",
			address:    []uint16{10},
			quantity:   1,
			byte_order: "AB",
			data_type:  "UINT16",
			write:      []byte{0xAB, 0xCD},
			read:       uint16(43981),
		},
		{
			name:       "Register20-BA-UINT16",
			address:    []uint16{20},
			quantity:   1,
			byte_order: "BA",
			data_type:  "UINT16",
			write:      []byte{0xAB, 0xCD},
			read:       uint16(52651),
		},
		{
			name:       "Register30-AB-INT16",
			address:    []uint16{20},
			quantity:   1,
			byte_order: "AB",
			data_type:  "INT16",
			write:      []byte{0xAB, 0xCD},
			read:       int16(-21555),
		},
		{
			name:       "Register40-BA-INT16",
			address:    []uint16{40},
			quantity:   1,
			byte_order: "BA",
			data_type:  "INT16",
			write:      []byte{0xAB, 0xCD},
			read:       int16(-12885),
		},
		{
			name:       "Register50-Register51-ABCD-INT32",
			address:    []uint16{50, 51},
			quantity:   2,
			byte_order: "ABCD",
			data_type:  "INT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       int32(-1430532899),
		},
		{
			name:       "Register60-Register61-DCBA-INT32",
			address:    []uint16{60, 61},
			quantity:   2,
			byte_order: "DCBA",
			data_type:  "INT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       int32(-573785174),
		},
		{
			name:       "Register70-Register71-BADC-INT32",
			address:    []uint16{70, 71},
			quantity:   2,
			byte_order: "BADC",
			data_type:  "INT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       int32(-1146430004),
		},
		{
			name:       "Register80-Register81-CDAB-INT32",
			address:    []uint16{80, 81},
			quantity:   2,
			byte_order: "CDAB",
			data_type:  "INT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       int32(-857888069),
		},
		{
			name:       "Register90-Register91-ABCD-UINT32",
			address:    []uint16{90, 91},
			quantity:   2,
			byte_order: "ABCD",
			data_type:  "UINT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       uint32(2864434397),
		},
		{
			name:       "Register100-Register101-DCBA-UINT32",
			address:    []uint16{100, 101},
			quantity:   2,
			byte_order: "DCBA",
			data_type:  "UINT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       uint32(3721182122),
		},
		{
			name:       "Register110-Register111-BADC-UINT32",
			address:    []uint16{110, 111},
			quantity:   2,
			byte_order: "BADC",
			data_type:  "UINT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       uint32(3148537292),
		},
		{
			name:       "Register120-Register121-CDAB-UINT32",
			address:    []uint16{120, 121},
			quantity:   2,
			byte_order: "CDAB",
			data_type:  "UINT32",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       uint32(3437079227),
		},
		{
			name:       "Register130-Register131-ABCD-FLOAT32-IEEE",
			address:    []uint16{130, 131},
			quantity:   2,
			byte_order: "ABCD",
			data_type:  "FLOAT32-IEEE",
			write:      []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:       float32(-3.3360025e-13),
		},
	}

	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	defer serv.Close()
	assert.NoError(t, err)

	handler := m.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	assert.NoError(t, err)
	defer handler.Close()
	client := m.NewClient(handler)

	for _, hrt := range holdingRegisterTests {
		t.Run(hrt.name, func(t *testing.T) {
			_, err = client.WriteMultipleRegisters(hrt.address[0], hrt.quantity, hrt.write)
			assert.NoError(t, err)

			modbus := Modbus{
				Controller: "tcp://localhost:1502",
				Slave_Id:   1,
				Holding_Registers: []tag{
					{
						Byte_Order: hrt.byte_order,
						Data_Type:  hrt.data_type,
						Scale:      hrt.scale,
						Address:    hrt.address,
					},
				},
			}

			var acc testutil.Accumulator
			modbus.Gather(&acc)

			for _, coil := range modbus.registers {
				assert.Equal(t, hrt.read, coil.Tags[0].value)
			}
		})
	}
}
