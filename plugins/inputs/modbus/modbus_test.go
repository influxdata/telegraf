package modbus

import (
	"fmt"
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
				Log: testutil.Logger{},
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
			name:      "register0_ab_float32_msb",
			address:   []uint16{0},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "FLOAT32",
			scale:     0.1,
			write:     []byte{0x89, 0x65},
			read:      float64(3517.3),
		},
		{
			name:      "register0_register1_ab_float32_msb",
			address:   []uint16{0, 1},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FLOAT32",
			scale:     0.001,
			write:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			read:      float64(4294967.295),
		},
		{
			name:      "register5_to_register8_abcdefgh_float32",
			address:   []uint16{5, 6, 7, 8},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "FLOAT32",
			scale:     0.000001,
			write:     []byte{0x00, 0x00, 0x00, 0x62, 0xC6, 0xD1, 0xA9, 0xB2},
			read:      float64(424242.424242),
		},
		{
			name:      "register6_to_register9_hgfedcba_float32_msb",
			address:   []uint16{6, 7, 8, 9},
			quantity:  4,
			byteOrder: "HGFEDCBA",
			dataType:  "FLOAT32",
			scale:     0.0000000001,
			write:     []byte{0xEA, 0x1E, 0x39, 0xEE, 0x8E, 0xA9, 0x54, 0xAB},
			read:      float64(1234567890.9876544),
		},
		{
			name:      "register0_ab_float",
			address:   []uint16{0},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "FIXED",
			scale:     0.1,
			write:     []byte{0xFF, 0xD6},
			read:      float64(-4.2),
		},
		{
			name:      "register1_ba_ufloat",
			address:   []uint16{1},
			quantity:  1,
			byteOrder: "BA",
			dataType:  "UFIXED",
			scale:     0.1,
			write:     []byte{0xD8, 0xFF},
			read:      float64(6549.6),
		},
		{
			name:      "register4_register5_abcd_float",
			address:   []uint16{4, 5},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FIXED",
			scale:     0.1,
			write:     []byte{0xFF, 0xFF, 0xFF, 0xD6},
			read:      float64(-4.2),
		},
		{
			name:      "register5_register6_dcba_ufloat",
			address:   []uint16{5, 6},
			quantity:  2,
			byteOrder: "DCBA",
			dataType:  "UFIXED",
			scale:     0.001,
			write:     []byte{0xD8, 0xFF, 0xFF, 0xFF},
			read:      float64(4294967.256),
		},
		{
			name:      "register5_to_register8_abcdefgh_float",
			address:   []uint16{5, 6, 7, 8},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "FIXED",
			scale:     0.000001,
			write:     []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xD6},
			read:      float64(-0.000042),
		},
		{
			name:      "register6_to_register9_hgfedcba_ufloat",
			address:   []uint16{6, 7, 8, 9},
			quantity:  4,
			byteOrder: "HGFEDCBA",
			dataType:  "UFIXED",
			scale:     0.000000001,
			write:     []byte{0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			read:      float64(18441921395.520346504),
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
			name:      "register50_register51_abcd_int32_scaled",
			address:   []uint16{50, 51},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "INT32",
			scale:     10,
			write:     []byte{0x00, 0x00, 0xAB, 0xCD},
			read:      int32(439810),
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
		{
			name:      "register130_register131_abcd_float32_ieee_scaled",
			address:   []uint16{130, 131},
			quantity:  2,
			byteOrder: "ABCD",
			dataType:  "FLOAT32-IEEE",
			scale:     10,
			write:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
			read:      float32(-3.3360025e-12),
		},
		{
			name:      "register140_to_register143_abcdefgh_int64_scaled",
			address:   []uint16{140, 141, 142, 143},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "INT64",
			scale:     10,
			write:     []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0xAB, 0xCD},
			read:      int64(10995116717570),
		},
		{
			name:      "register140_to_register143_abcdefgh_int64",
			address:   []uint16{140, 141, 142, 143},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "INT64",
			scale:     1,
			write:     []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0xAB, 0xCD},
			read:      int64(1099511671757),
		},
		{
			name:      "register150_to_register153_hgfedcba_int64",
			address:   []uint16{150, 151, 152, 153},
			quantity:  4,
			byteOrder: "HGFEDCBA",
			dataType:  "INT64",
			scale:     1,
			write:     []byte{0x84, 0xF6, 0x45, 0xF9, 0xBC, 0xFE, 0xFF, 0xFF},
			read:      int64(-1387387292028),
		},
		{
			name:      "register160_to_register163_badcfehg_int64",
			address:   []uint16{160, 161, 162, 163},
			quantity:  4,
			byteOrder: "BADCFEHG",
			dataType:  "INT64",
			scale:     1,
			write:     []byte{0xFF, 0xFF, 0xBC, 0xFE, 0x45, 0xF9, 0x84, 0xF6},
			read:      int64(-1387387292028),
		},
		{
			name:      "register170_to_register173_ghefcdab_int64",
			address:   []uint16{170, 171, 172, 173},
			quantity:  4,
			byteOrder: "GHEFCDAB",
			dataType:  "INT64",
			scale:     1,
			write:     []byte{0xF6, 0x84, 0xF9, 0x45, 0xFE, 0xBC, 0xFF, 0xFF},
			read:      int64(-1387387292028),
		},
		{
			name:      "register180_to_register183_abcdefgh_uint64_scaled",
			address:   []uint16{180, 181, 182, 183},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "UINT64",
			scale:     10,
			write:     []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0xAB, 0xCD},
			read:      uint64(10995116717570),
		},
		{
			name:      "register180_to_register183_abcdefgh_uint64",
			address:   []uint16{180, 181, 182, 183},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "UINT64",
			scale:     1,
			write:     []byte{0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0xAB, 0xCD},
			read:      uint64(1099511671757),
		},
		{
			name:      "register190_to_register193_hgfedcba_uint64",
			address:   []uint16{190, 191, 192, 193},
			quantity:  4,
			byteOrder: "HGFEDCBA",
			dataType:  "UINT64",
			scale:     1,
			write:     []byte{0x84, 0xF6, 0x45, 0xF9, 0xBC, 0xFE, 0xFF, 0xFF},
			read:      uint64(18446742686322259968),
		},
		{
			name:      "register200_to_register203_badcfehg_uint64",
			address:   []uint16{200, 201, 202, 203},
			quantity:  4,
			byteOrder: "BADCFEHG",
			dataType:  "UINT64",
			scale:     1,
			write:     []byte{0xFF, 0xFF, 0xBC, 0xFE, 0x45, 0xF9, 0x84, 0xF6},
			read:      uint64(18446742686322259968),
		},
		{
			name:      "register210_to_register213_ghefcdab_uint64",
			address:   []uint16{210, 211, 212, 213},
			quantity:  4,
			byteOrder: "GHEFCDAB",
			dataType:  "UINT64",
			scale:     1,
			write:     []byte{0xF6, 0x84, 0xF9, 0x45, 0xFE, 0xBC, 0xFF, 0xFF},
			read:      uint64(18446742686322259968),
		},
		{
			name:      "register214_to_register217_abcdefgh_float64_ieee",
			address:   []uint16{214, 215, 216, 217},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "FLOAT64-IEEE",
			scale:     1,
			write:     []byte{0xBF, 0x9C, 0x6A, 0x40, 0xC3, 0x47, 0x8F, 0x55},
			read:      float64(-0.02774907295123737),
		},
		{
			name:      "register214_to_register217_abcdefgh_float64_ieee_scaled",
			address:   []uint16{214, 215, 216, 217},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "FLOAT64-IEEE",
			scale:     0.1,
			write:     []byte{0xBF, 0x9C, 0x6A, 0x40, 0xC3, 0x47, 0x8F, 0x55},
			read:      float64(-0.002774907295123737),
		},
		{
			name:      "register218_to_register221_abcdefgh_float64_ieee_pos",
			address:   []uint16{218, 219, 220, 221},
			quantity:  4,
			byteOrder: "ABCDEFGH",
			dataType:  "FLOAT64-IEEE",
			scale:     1,
			write:     []byte{0x3F, 0x9C, 0x6A, 0x40, 0xC3, 0x47, 0x8F, 0x55},
			read:      float64(0.02774907295123737),
		},
		{
			name:      "register222_to_register225_hgfecdba_float64_ieee",
			address:   []uint16{222, 223, 224, 225},
			quantity:  4,
			byteOrder: "HGFEDCBA",
			dataType:  "FLOAT64-IEEE",
			scale:     1,
			write:     []byte{0x55, 0x8F, 0x47, 0xC3, 0x40, 0x6A, 0x9C, 0xBF},
			read:      float64(-0.02774907295123737),
		},
		{
			name:      "register226_to_register229_badcfehg_float64_ieee",
			address:   []uint16{226, 227, 228, 229},
			quantity:  4,
			byteOrder: "BADCFEHG",
			dataType:  "FLOAT64-IEEE",
			scale:     1,
			write:     []byte{0x9C, 0xBF, 0x40, 0x6A, 0x47, 0xC3, 0x55, 0x8F},
			read:      float64(-0.02774907295123737),
		},
		{
			name:      "register230_to_register233_ghefcdab_float64_ieee",
			address:   []uint16{230, 231, 232, 233},
			quantity:  4,
			byteOrder: "GHEFCDAB",
			dataType:  "FLOAT64-IEEE",
			scale:     1,
			write:     []byte{0x8F, 0x55, 0xC3, 0x47, 0x6A, 0x40, 0xBF, 0x9C},
			read:      float64(-0.02774907295123737),
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
				Log: testutil.Logger{},
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

func TestReadMultipleCoilLimit(t *testing.T) {
	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	assert.NoError(t, err)
	defer serv.Close()

	handler := m.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	assert.NoError(t, err)
	defer handler.Close()
	client := m.NewClient(handler)

	fcs := []fieldContainer{}
	writeValue := uint16(0)
	for i := 0; i <= 4000; i++ {
		fc := fieldContainer{}
		fc.Name = fmt.Sprintf("coil-%v", i)
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		t.Run(fc.Name, func(t *testing.T) {
			_, err = client.WriteSingleCoil(fc.Address[0], writeValue)
			assert.NoError(t, err)
		})

		writeValue = 65280 - writeValue
	}

	modbus := Modbus{
		Name:       "TestReadCoils",
		Controller: "tcp://localhost:1502",
		SlaveID:    1,
		Coils:      fcs,
	}

	err = modbus.Init()
	assert.NoError(t, err)
	var acc testutil.Accumulator
	err = modbus.Gather(&acc)
	assert.NoError(t, err)

	writeValue = 0
	for i := 0; i <= 4000; i++ {
		t.Run(modbus.registers[0].Fields[i].Name, func(t *testing.T) {
			assert.Equal(t, writeValue, modbus.registers[0].Fields[i].value)
			writeValue = 1 - writeValue
		})
	}
}

func TestReadMultipleHoldingRegisterLimit(t *testing.T) {
	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	assert.NoError(t, err)
	defer serv.Close()

	handler := m.NewTCPClientHandler("localhost:1502")
	err = handler.Connect()
	assert.NoError(t, err)
	defer handler.Close()
	client := m.NewClient(handler)

	fcs := []fieldContainer{}
	for i := 0; i <= 400; i++ {
		fc := fieldContainer{}
		fc.Name = fmt.Sprintf("HoldingRegister-%v", i)
		fc.ByteOrder = "AB"
		fc.DataType = "INT16"
		fc.Scale = 1.0
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		t.Run(fc.Name, func(t *testing.T) {
			_, err = client.WriteSingleRegister(fc.Address[0], uint16(i))
			assert.NoError(t, err)
		})
	}

	modbus := Modbus{
		Name:             "TestHoldingRegister",
		Controller:       "tcp://localhost:1502",
		SlaveID:          1,
		HoldingRegisters: fcs,
	}

	err = modbus.Init()
	assert.NoError(t, err)
	var acc testutil.Accumulator
	err = modbus.Gather(&acc)
	assert.NoError(t, err)

	for i := 0; i <= 400; i++ {
		assert.Equal(t, int16(i), modbus.registers[0].Fields[i].value)
	}
}

func TestRetrySuccessful(t *testing.T) {
	retries := 0
	maxretries := 2
	value := 1

	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	assert.NoError(t, err)
	defer serv.Close()

	// Make read on coil-registers fail for some trials by making the device
	// to appear busy
	serv.RegisterFunctionHandler(1,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			data := make([]byte, 2)
			data[0] = byte(1)
			data[1] = byte(value)

			except := &mbserver.SlaveDeviceBusy
			if retries >= maxretries {
				except = &mbserver.Success
			}
			retries++

			return data, except
		})

	t.Run("retry_success", func(t *testing.T) {
		modbus := Modbus{
			Name:       "TestRetry",
			Controller: "tcp://localhost:1502",
			SlaveID:    1,
			Retries:    maxretries,
			Coils: []fieldContainer{
				{
					Name:    "retry_success",
					Address: []uint16{0},
				},
			},
			Log: testutil.Logger{},
		}

		err = modbus.Init()
		assert.NoError(t, err)
		var acc testutil.Accumulator
		err = modbus.Gather(&acc)
		assert.NoError(t, err)
		assert.NotEmpty(t, modbus.registers)

		for _, coil := range modbus.registers {
			assert.Equal(t, uint16(value), coil.Fields[0].value)
		}
	})
}

func TestRetryFail(t *testing.T) {
	maxretries := 2

	serv := mbserver.NewServer()
	err := serv.ListenTCP("localhost:1502")
	assert.NoError(t, err)
	defer serv.Close()

	// Make the read on coils fail with busy
	serv.RegisterFunctionHandler(1,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			data := make([]byte, 2)
			data[0] = byte(1)
			data[1] = byte(0)

			return data, &mbserver.SlaveDeviceBusy
		})

	t.Run("retry_fail", func(t *testing.T) {
		modbus := Modbus{
			Name:       "TestRetryFail",
			Controller: "tcp://localhost:1502",
			SlaveID:    1,
			Retries:    maxretries,
			Coils: []fieldContainer{
				{
					Name:    "retry_fail",
					Address: []uint16{0},
				},
			},
			Log: testutil.Logger{},
		}

		err = modbus.Init()
		assert.NoError(t, err)
		var acc testutil.Accumulator
		err = modbus.Gather(&acc)
		assert.Error(t, err)
	})

	// Make the read on coils fail with illegal function preventing retry
	counter := 0
	serv.RegisterFunctionHandler(1,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			counter++
			data := make([]byte, 2)
			data[0] = byte(1)
			data[1] = byte(0)

			return data, &mbserver.IllegalFunction
		})

	t.Run("retry_fail", func(t *testing.T) {
		modbus := Modbus{
			Name:       "TestRetryFail",
			Controller: "tcp://localhost:1502",
			SlaveID:    1,
			Retries:    maxretries,
			Coils: []fieldContainer{
				{
					Name:    "retry_fail",
					Address: []uint16{0},
				},
			},
			Log: testutil.Logger{},
		}

		err = modbus.Init()
		assert.NoError(t, err)
		var acc testutil.Accumulator
		err = modbus.Gather(&acc)
		assert.Error(t, err)
		assert.Equal(t, counter, 1)
	})
}
