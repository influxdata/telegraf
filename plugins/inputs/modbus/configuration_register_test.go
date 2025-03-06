package modbus

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	mb "github.com/grid-x/modbus"
	"github.com/stretchr/testify/require"
	"github.com/tbrandon/mbserver"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestRegister(t *testing.T) {
	modbus := Modbus{
		Name:              "TestRetryFailExhausted",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "register",
		Log:               testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.Coils = []fieldDefinition{
		{
			Name:    "coil",
			Address: []uint16{0},
		},
	}
	modbus.DiscreteInputs = []fieldDefinition{
		{
			Name:    "discrete",
			Address: []uint16{0},
		},
	}
	modbus.HoldingRegisters = []fieldDefinition{
		{
			Name:      "holding",
			Address:   []uint16{0},
			DataType:  "INT16",
			ByteOrder: "AB",
			Scale:     1.0,
		},
	}
	modbus.InputRegisters = []fieldDefinition{
		{
			Name:      "input",
			Address:   []uint16{0},
			DataType:  "INT16",
			ByteOrder: "AB",
			Scale:     1.0,
		},
	}

	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NotNil(t, modbus.requests[1])
	require.Len(t, modbus.requests[1].coil, len(modbus.Coils))
	require.Len(t, modbus.requests[1].discrete, len(modbus.DiscreteInputs))
	require.Len(t, modbus.requests[1].holding, len(modbus.HoldingRegisters))
	require.Len(t, modbus.requests[1].input, len(modbus.InputRegisters))
}

func TestRegisterCoils(t *testing.T) {
	var coilTests = []struct {
		name     string
		address  uint16
		dtype    string
		quantity uint16
		write    []byte
		read     interface{}
	}{
		{
			name:     "coil0_turn_off",
			address:  0,
			quantity: 1,
			write:    []byte{0x00},
			read:     uint16(0),
		},
		{
			name:     "coil0_turn_on",
			address:  0,
			quantity: 1,
			write:    []byte{0x01},
			read:     uint16(1),
		},
		{
			name:     "coil1_turn_on",
			address:  1,
			quantity: 1,
			write:    []byte{0x01},
			read:     uint16(1),
		},
		{
			name:     "coil2_turn_on",
			address:  2,
			quantity: 1,
			write:    []byte{0x01},
			read:     uint16(1),
		},
		{
			name:     "coil3_turn_on",
			address:  3,
			quantity: 1,
			write:    []byte{0x01},
			read:     uint16(1),
		},
		{
			name:     "coil1_turn_off",
			address:  1,
			quantity: 1,
			write:    []byte{0x00},
			read:     uint16(0),
		},
		{
			name:     "coil2_turn_off",
			address:  2,
			quantity: 1,
			write:    []byte{0x00},
			read:     uint16(0),
		},
		{
			name:     "coil3_turn_off",
			address:  3,
			quantity: 1,
			write:    []byte{0x00},
			read:     uint16(0),
		},
		{
			name:     "coil4_turn_off",
			address:  4,
			quantity: 1,
			write:    []byte{0x00},
			read:     uint16(0),
		},
		{
			name:     "coil4_turn_on",
			address:  4,
			quantity: 1,
			write:    []byte{0x01},
			read:     uint16(1),
		},
		{
			name:     "coil4_turn_off_bool",
			address:  4,
			quantity: 1,
			dtype:    "BOOL",
			write:    []byte{0x00},
			read:     false,
		},
		{
			name:     "coil4_turn_on_bool",
			address:  4,
			quantity: 1,
			dtype:    "BOOL",
			write:    []byte{0x01},
			read:     true,
		},
	}

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	for _, ct := range coilTests {
		t.Run(ct.name, func(t *testing.T) {
			_, err := client.WriteMultipleCoils(ct.address, ct.quantity, ct.write)
			require.NoError(t, err)

			modbus := Modbus{
				Name:       "TestCoils",
				Controller: "tcp://localhost:1502",
				Log:        testutil.Logger{},
			}
			modbus.SlaveID = 1
			modbus.Coils = []fieldDefinition{
				{
					Name:     ct.name,
					Address:  []uint16{ct.address},
					DataType: ct.dtype,
				},
			}

			expected := []telegraf.Metric{
				testutil.MustMetric(
					"modbus",
					map[string]string{
						"type":     cCoils,
						"slave_id": strconv.Itoa(int(modbus.SlaveID)),
						"name":     modbus.Name,
					},
					map[string]interface{}{ct.name: ct.read},
					time.Unix(0, 0),
				),
			}

			var acc testutil.Accumulator
			require.NoError(t, modbus.Init())
			require.NotEmpty(t, modbus.requests)
			require.NoError(t, modbus.Gather(&acc))
			acc.Wait(len(expected))

			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestRegisterHoldingRegisters(t *testing.T) {
	var holdingRegisterTests = []struct {
		name      string
		address   []uint16
		quantity  uint16
		bit       uint8
		byteOrder string
		dataType  string
		scale     float64
		write     []byte
		read      interface{}
	}{
		{
			name:      "register5_bit3",
			address:   []uint16{5},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "BIT",
			bit:       3,
			write:     []byte{0x18, 0x0d},
			read:      uint8(1),
		},
		{
			name:      "register5_bit14",
			address:   []uint16{5},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "BIT",
			bit:       14,
			write:     []byte{0x18, 0x0d},
			read:      uint8(0),
		},
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
			name:      "register20_uint16",
			address:   []uint16{10},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT8L",
			scale:     1,
			write:     []byte{0x18, 0x0D},
			read:      uint8(13),
		},
		{
			name:      "register20_uint16-scale_.1",
			address:   []uint16{10},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT8L",
			scale:     .1,
			write:     []byte{0x18, 0x0D},
			read:      uint8(1),
		},
		{
			name:      "register20_uint16_scale_10",
			address:   []uint16{10},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT8L",
			scale:     10,
			write:     []byte{0x18, 0x0D},
			read:      uint8(130),
		},
		{
			name:      "register11_uint8H",
			address:   []uint16{11},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT8H",
			scale:     1,
			write:     []byte{0x18, 0x0D},
			read:      uint8(24),
		},
		{
			name:      "register11_uint8L-scale_.1",
			address:   []uint16{11},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT8H",
			scale:     .1,
			write:     []byte{0x18, 0x0D},
			read:      uint8(2),
		},
		{
			name:      "register11_uint8L_scale_10",
			address:   []uint16{11},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT8H",
			scale:     10,
			write:     []byte{0x18, 0x0D},
			read:      uint8(240),
		},
		{
			name:      "register12_int8L",
			address:   []uint16{12},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT8L",
			scale:     1,
			write:     []byte{0x98, 0x8D},
			read:      int8(-115),
		},
		{
			name:      "register12_int8L-scale_.1",
			address:   []uint16{12},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT8L",
			scale:     .1,
			write:     []byte{0x98, 0x8D},
			read:      int8(-11),
		},
		{
			name:      "register12_int8L_scale_10",
			address:   []uint16{12},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT8L",
			scale:     10,
			write:     []byte{0x98, 0xF8},
			read:      int8(-80),
		},
		{
			name:      "register13_int8H",
			address:   []uint16{13},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT8H",
			scale:     1,
			write:     []byte{0x98, 0x8D},
			read:      int8(-104),
		},
		{
			name:      "register13_int8H-scale_.1",
			address:   []uint16{13},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT8H",
			scale:     .1,
			write:     []byte{0x98, 0x8D},
			read:      int8(-10),
		},
		{
			name:      "register13_int8H_scale_10",
			address:   []uint16{13},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "INT8H",
			scale:     10,
			write:     []byte{0xFD, 0x8D},
			read:      int8(-30),
		},
		{
			name:      "register15_ab_uint16",
			address:   []uint16{15},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT16",
			scale:     1,
			write:     []byte{0xAB, 0xCD},
			read:      uint16(43981),
		},
		{
			name:      "register15_ab_uint16-scale_.1",
			address:   []uint16{15},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "UINT16",
			scale:     .1,
			write:     []byte{0xAB, 0xCD},
			read:      uint16(4398),
		},
		{
			name:      "register15_ab_uint16_scale_10",
			address:   []uint16{15},
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
		{
			name:      "register240_abcd_float16",
			address:   []uint16{240},
			quantity:  1,
			byteOrder: "AB",
			dataType:  "FLOAT16-IEEE",
			scale:     1,
			write:     []byte{0xb8, 0x14},
			read:      float64(-0.509765625),
		},
		{
			name:      "register240_dcba_float16",
			address:   []uint16{240},
			quantity:  1,
			byteOrder: "BA",
			dataType:  "FLOAT16-IEEE",
			scale:     1,
			write:     []byte{0x14, 0xb8},
			read:      float64(-0.509765625),
		},
		{
			name:      "register250_abcd_string",
			address:   []uint16{250, 251, 252, 253, 254, 255, 256},
			quantity:  7,
			byteOrder: "AB",
			dataType:  "STRING",
			write:     []byte{0x4d, 0x6f, 0x64, 0x62, 0x75, 0x73, 0x20, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x00},
			read:      "Modbus String",
		},
		{
			name:      "register250_dcba_string",
			address:   []uint16{250, 251, 252, 253, 254, 255, 256},
			quantity:  7,
			byteOrder: "BA",
			dataType:  "STRING",
			write:     []byte{0x6f, 0x4d, 0x62, 0x64, 0x73, 0x75, 0x53, 0x20, 0x72, 0x74, 0x6e, 0x69, 0x00, 0x67},
			read:      "Modbus String",
		},
	}

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	for _, hrt := range holdingRegisterTests {
		t.Run(hrt.name, func(t *testing.T) {
			_, err := client.WriteMultipleRegisters(hrt.address[0], hrt.quantity, hrt.write)
			require.NoError(t, err)

			modbus := Modbus{
				Name:       "TestHoldingRegisters",
				Controller: "tcp://localhost:1502",
				Log:        testutil.Logger{},
			}
			modbus.SlaveID = 1
			modbus.HoldingRegisters = []fieldDefinition{
				{
					Name:      hrt.name,
					ByteOrder: hrt.byteOrder,
					DataType:  hrt.dataType,
					Scale:     hrt.scale,
					Address:   hrt.address,
					Bit:       hrt.bit,
				},
			}

			expected := []telegraf.Metric{
				testutil.MustMetric(
					"modbus",
					map[string]string{
						"type":     cHoldingRegisters,
						"slave_id": strconv.Itoa(int(modbus.SlaveID)),
						"name":     modbus.Name,
					},
					map[string]interface{}{hrt.name: hrt.read},
					time.Unix(0, 0),
				),
			}

			var acc testutil.Accumulator
			require.NoError(t, modbus.Init())
			require.NotEmpty(t, modbus.requests)
			require.NoError(t, modbus.Gather(&acc))
			acc.Wait(len(expected))

			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestRegisterReadMultipleCoilWithHole(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := make([]fieldDefinition, 0, 26)
	expectedFields := make(map[string]interface{})
	writeValue := uint16(0)
	readValue := uint16(0)
	for i := 0; i < 14; i++ {
		fc := fieldDefinition{}
		fc.Name = fmt.Sprintf("coil-%v", i)
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleCoil(fc.Address[0], writeValue)
		require.NoError(t, err)

		expectedFields[fc.Name] = readValue
		writeValue = 65280 - writeValue
		readValue = 1 - readValue
	}
	for i := 15; i < 18; i++ {
		fc := fieldDefinition{}
		fc.Name = fmt.Sprintf("coil-%v", i)
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleCoil(fc.Address[0], writeValue)
		require.NoError(t, err)

		expectedFields[fc.Name] = readValue
		writeValue = 65280 - writeValue
		readValue = 1 - readValue
	}
	for i := 24; i < 33; i++ {
		fc := fieldDefinition{}
		fc.Name = fmt.Sprintf("coil-%v", i)
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleCoil(fc.Address[0], writeValue)
		require.NoError(t, err)

		expectedFields[fc.Name] = readValue
		writeValue = 65280 - writeValue
		readValue = 1 - readValue
	}
	require.Len(t, expectedFields, len(fcs))

	modbus := Modbus{
		Name:       "TestReadMultipleCoilWithHole",
		Controller: "tcp://localhost:1502",
		Log:        testutil.Logger{Name: "modbus:MultipleCoilWithHole"},
	}
	modbus.SlaveID = 1
	modbus.Coils = fcs

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cCoils,
				"slave_id": strconv.Itoa(int(modbus.SlaveID)),
				"name":     modbus.Name,
			},
			expectedFields,
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NoError(t, modbus.Gather(&acc))
	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestRegisterReadMultipleCoilLimit(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := make([]fieldDefinition, 0, 4000)
	expectedFields := make(map[string]interface{})
	writeValue := uint16(0)
	readValue := uint16(0)
	for i := 0; i < 4000; i++ {
		fc := fieldDefinition{}
		fc.Name = fmt.Sprintf("coil-%v", i)
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleCoil(fc.Address[0], writeValue)
		require.NoError(t, err)

		expectedFields[fc.Name] = readValue
		writeValue = 65280 - writeValue
		readValue = 1 - readValue
	}
	require.Len(t, expectedFields, len(fcs))

	modbus := Modbus{
		Name:       "TestReadCoils",
		Controller: "tcp://localhost:1502",
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.Coils = fcs

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cCoils,
				"slave_id": strconv.Itoa(int(modbus.SlaveID)),
				"name":     modbus.Name,
			},
			expectedFields,
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NoError(t, modbus.Gather(&acc))
	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestRegisterReadMultipleHoldingRegisterWithHole(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := make([]fieldDefinition, 0, 20)
	expectedFields := make(map[string]interface{})
	for i := 0; i < 10; i++ {
		fc := fieldDefinition{
			Name:      fmt.Sprintf("HoldingRegister-%v", i),
			ByteOrder: "AB",
			DataType:  "INT16",
			Scale:     1.0,
			Address:   []uint16{uint16(i)},
		}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleRegister(fc.Address[0], uint16(i))
		require.NoError(t, err)

		expectedFields[fc.Name] = int64(i)
	}
	for i := 20; i < 30; i++ {
		fc := fieldDefinition{
			Name:      fmt.Sprintf("HoldingRegister-%v", i),
			ByteOrder: "AB",
			DataType:  "INT16",
			Scale:     1.0,
			Address:   []uint16{uint16(i)},
		}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleRegister(fc.Address[0], uint16(i))
		require.NoError(t, err)

		expectedFields[fc.Name] = int64(i)
	}
	require.Len(t, expectedFields, len(fcs))

	modbus := Modbus{
		Name:       "TestHoldingRegister",
		Controller: "tcp://localhost:1502",
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.HoldingRegisters = fcs

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cHoldingRegisters,
				"slave_id": strconv.Itoa(int(modbus.SlaveID)),
				"name":     modbus.Name,
			},
			expectedFields,
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NoError(t, modbus.Gather(&acc))
	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestRegisterReadMultipleHoldingRegisterLimit(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := make([]fieldDefinition, 0, 401)
	expectedFields := make(map[string]interface{})
	for i := 0; i <= 400; i++ {
		fc := fieldDefinition{}
		fc.Name = fmt.Sprintf("HoldingRegister-%v", i)
		fc.ByteOrder = "AB"
		fc.DataType = "INT16"
		fc.Scale = 1.0
		fc.Address = []uint16{uint16(i)}
		fcs = append(fcs, fc)

		_, err := client.WriteSingleRegister(fc.Address[0], uint16(i))
		require.NoError(t, err)

		expectedFields[fc.Name] = int64(i)
	}

	modbus := Modbus{
		Name:       "TestHoldingRegister",
		Controller: "tcp://localhost:1502",
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.HoldingRegisters = fcs

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cHoldingRegisters,
				"slave_id": strconv.Itoa(int(modbus.SlaveID)),
				"name":     modbus.Name,
			},
			expectedFields,
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NoError(t, modbus.Gather(&acc))
	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestRegisterHighAddresses(t *testing.T) {
	// Test case for issue https://github.com/influxdata/telegraf/issues/15138

	// Setup a server
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	// Write the register values
	data := []byte{
		0x4d, 0x6f, 0x64, 0x62, 0x75, 0x73, 0x20, 0x53,
		0x74, 0x72, 0x69, 0x6e, 0x67, 0x20, 0x48, 0x65,
		0x6c, 0x6c, 0x6f, 0x00,
	}
	_, err := client.WriteMultipleRegisters(65524, 10, data)
	require.NoError(t, err)
	_, err = client.WriteMultipleRegisters(65534, 1, []byte{0x10, 0x92})
	require.NoError(t, err)

	modbus := Modbus{
		Name:       "Issue-15138",
		Controller: "tcp://localhost:1502",
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.HoldingRegisters = []fieldDefinition{
		{
			Name:      "DeviceName",
			ByteOrder: "AB",
			DataType:  "STRING",
			Address:   []uint16{65524, 65525, 65526, 65527, 65528, 65529, 65530, 65531, 65532, 65533},
		},
		{
			Name:      "DeviceConnectionStatus",
			ByteOrder: "AB",
			DataType:  "UINT16",
			Address:   []uint16{65534},
			Scale:     1,
		},
	}

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cHoldingRegisters,
				"slave_id": strconv.Itoa(int(modbus.SlaveID)),
				"name":     modbus.Name,
			},
			map[string]interface{}{
				"DeviceName":             "Modbus String Hello",
				"DeviceConnectionStatus": uint16(4242),
			},
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.Len(t, modbus.requests[1].holding, 1)
	require.NoError(t, modbus.Gather(&acc))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
