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
			name:     "coil0_turn_off",
			address:  0,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
		{
			name:     "coil0_turn_on",
			address:  0,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "coil1_turn_on",
			address:  1,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "coil2_turn_on",
			address:  2,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "coil3_turn_on",
			address:  3,
			quantity: 1,
			write:    []byte{0x01},
			read:     1,
		},
		{
			name:     "coil1_turn_off",
			address:  1,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
		{
			name:     "coil2_turn_off",
			address:  2,
			quantity: 1,
			write:    []byte{0x00},
			read:     0,
		},
		{
			name:     "coil3_turn_off",
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
				Name:       "TestCoils",
				Controller: "tcp://localhost:1502",
				SlaveID:    1,
				Coils: []fieldContainer{
					{
						Name:    ct.name,
						Address: []uint16{ct.address},
					},
				},
			}

			err = modbus.Init()
			assert.NoError(t, err)
			var acc testutil.Accumulator
			err = modbus.Gather(&acc)
			assert.NoError(t, err)
			assert.NotEmpty(t, modbus.registers)

			for _, coil := range modbus.registers {
				assert.Equal(t, ct.read, coil.Fields[0].value)
			}
		})
	}
}

func TestHoldingRegisters(t *testing.T) {
	var holdingRegisterTests = []struct {
		name      string
		address   []uint16
		quantity  uint16
		byteOrder string
		dataType  string
		scale     float64
		write     []byte
		read      interface{}
	}{
		{
			name:      "register0_ab_float32",
			address:   []uint16{0},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "FLOAT32",
			scale:     0.1,
			write:     []byte{0x08, 0x98},
			read:      float64(220),
		},
		{
			name:      "register0_register1_ab_float32",
			address:   []uint16{0, 1},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FLOAT32",
			scale:     0.001,
			write:     []byte{0x00, 0x00, 0x03, 0xE8},
			read:      float64(1),
		},
		{
			name:      "register1_register2_abcd_float32",
			address:   []uint16{1, 2},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FLOAT32",
			scale:     0.1,
			write:     []byte{0x00, 0x00, 0x08, 0x98},
			read:      float64(220),
		},
		{
			name:      "register3_register4_abcd_float32",
			address:   []uint16{3, 4},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FLOAT32",
			scale:     0.1,
			write:     []byte{0x00, 0x00, 0x08, 0x98},
			read:      float64(220),
		},
		{
			name:      "register7_ab_float32",
			address:   []uint16{7},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "FLOAT32",
			scale:     0.1,
			write:     []byte{0x01, 0xF4},
			read:      float64(50),
		},
		{
			name:      "register10_ab_uint16",
			address:   []uint16{10},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT16",
			scale:     1,
			write:     []byte{0xAB, 0xCD},
			read:      uint16(43981),
		},
		{
			name:      "register10_ab_uint16-scale_.1",
			address:   []uint16{10},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT16",
			scale:     .1,
			write:     []byte{0xAB, 0xCD},
			read:      uint16(4398),
		},
		{
			name:      "register10_ab_uint16_scale_10",
			address:   []uint16{10},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT16",
			scale:     10,
			write:     []byte{0x00, 0x2A},
			read:      uint16(420),
		},
		{
			name:      "register20_ba_uint16",
			address:   []uint16{20},
			quantity:  1,
			byteOrder: "BA",
			dataType:  "UINT16",
			scale:     1,
			write:     []byte{0xAB, 0xCD},
			read:      uint16(52651),
		},
		{
			name:      "register30_ab_int16",
			address:   []uint16{20},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT16",
			scale:     1,
			write:     []byte{0xAB, 0xCD},
			read:      int16(-21555),
		},
		{
			name:      "register40_ba_int16",
			address:   []uint16{40},
			quantity:  1,
			byteOrder: "BA",
			dataType:  "INT16",
			scale:     1,
			write:     []byte{0xAB, 0xCD},
			read:      int16(-12885),
		},
		{
			name:      "register50_register51_abcd_int32",
			address:   []uint16{50, 51},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "INT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      int32(-1430532899),
		},
		{
			name:      "register60_register61_dcba_int32",
			address:   []uint16{60, 61},
			quantity:  2,
			byteOrder: "DCBA",
			dataType:  "INT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      int32(-573785174),
		},
		{
			name:      "register70_register71_badc_int32",
			address:   []uint16{70, 71},
			quantity:  2,
			byteOrder: "BADC",
			dataType:  "INT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      int32(-1146430004),
		},
		{
			name:      "register80_register81_cdab_int32",
			address:   []uint16{80, 81},
			quantity:  2,
			byteOrder: "CDAB",
			dataType:  "INT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      int32(-857888069),
		},
		{
			name:      "register90_register91_abcd_uint32",
			address:   []uint16{90, 91},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "UINT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      uint32(2864434397),
		},
		{
			name:      "register100_register101_dcba_uint32",
			address:   []uint16{100, 101},
			quantity:  2,
			byteOrder: "DCBA",
			dataType:  "UINT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      uint32(3721182122),
		},
		{
			name:      "register110_register111_badc_uint32",
			address:   []uint16{110, 111},
			quantity:  2,
			byteOrder: "BADC",
			dataType:  "UINT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      uint32(3148537292),
		},
		{
			name:      "register120_register121_cdab_uint32",
			address:   []uint16{120, 121},
			quantity:  2,
			byteOrder: "CDAB",
			dataType:  "UINT32",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      uint32(3437079227),
		},
		{
			name:      "register130_register131_abcd_float32_ieee",
			address:   []uint16{130, 131},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FLOAT32-IEEE",
			scale:     1,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      float32(-3.3360025e-13),
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
				Name:       "TestHoldingRegisters",
				Controller: "tcp://localhost:1502",
				SlaveID:    1,
				HoldingRegisters: []fieldContainer{
					{
						Name:      hrt.name,
						ByteOrder: hrt.byteOrder,
						DataType:  hrt.dataType,
						Scale:     hrt.scale,
						Address:   hrt.address,
					},
				},
			}

			err = modbus.Init()
			assert.NoError(t, err)
			var acc testutil.Accumulator
			modbus.Gather(&acc)
			assert.NotEmpty(t, modbus.registers)

			for _, coil := range modbus.registers {
				assert.Equal(t, hrt.read, coil.Fields[0].value)
			}
		})
	}
}
