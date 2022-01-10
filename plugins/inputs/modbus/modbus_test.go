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
					Name:    ct.name,
					Address: []uint16{ct.address},
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

func TestReadMultipleCoilWithHole(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := []fieldDefinition{}
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

func TestReadMultipleCoilLimit(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := []fieldDefinition{}
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

func TestReadMultipleHoldingRegisterWithHole(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := []fieldDefinition{}
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

func TestReadMultipleHoldingRegisterLimit(t *testing.T) {
	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	fcs := []fieldDefinition{}
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

func TestRetrySuccessful(t *testing.T) {
	retries := 0
	maxretries := 2
	value := 1

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
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

	modbus := Modbus{
		Name:       "TestRetry",
		Controller: "tcp://localhost:1502",
		Retries:    maxretries,
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.Coils = []fieldDefinition{
		{
			Name:    "retry_success",
			Address: []uint16{0},
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
			map[string]interface{}{"retry_success": uint16(value)},
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

func TestRetryFailExhausted(t *testing.T) {
	maxretries := 2

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	// Make the read on coils fail with busy
	serv.RegisterFunctionHandler(1,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			data := make([]byte, 2)
			data[0] = byte(1)
			data[1] = byte(0)

			return data, &mbserver.SlaveDeviceBusy
		})

	modbus := Modbus{
		Name:       "TestRetryFailExhausted",
		Controller: "tcp://localhost:1502",
		Retries:    maxretries,
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.Coils = []fieldDefinition{
		{
			Name:    "retry_fail",
			Address: []uint16{0},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)

	err := modbus.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "modbus: exception '6' (server device busy), function '129'", err.Error())
}

func TestRetryFailIllegal(t *testing.T) {
	maxretries := 2

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

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

	modbus := Modbus{
		Name:       "TestRetryFailExhausted",
		Controller: "tcp://localhost:1502",
		Retries:    maxretries,
		Log:        testutil.Logger{},
	}
	modbus.SlaveID = 1
	modbus.Coils = []fieldDefinition{
		{
			Name:    "retry_fail",
			Address: []uint16{0},
		},
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)

	err := modbus.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "modbus: exception '1' (illegal function), function '129'", err.Error())
	require.Equal(t, counter, 1)
}

func TestConfigurationRegister(t *testing.T) {
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

func TestConfigurationPerRequest(t *testing.T) {
	modbus := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "request",
		Log:               testutil.Logger{},
	}
	modbus.Requests = []requestDefinition{
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "coil",
			Fields: []requestFieldDefinition{
				{
					Name:    "coil-0",
					Address: uint16(0),
				},
				{
					Name:    "coil-1",
					Address: uint16(1),
					Omit:    true,
				},
				{
					Name:        "coil-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
		},
		{
			SlaveID:      1,
			RegisterType: "coil",
			Fields: []requestFieldDefinition{
				{
					Name:    "coil-3",
					Address: uint16(6),
				},
				{
					Name:    "coil-4",
					Address: uint16(7),
					Omit:    true,
				},
				{
					Name:        "coil-5",
					Address:     uint16(8),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
		},
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "discrete",
			Fields: []requestFieldDefinition{
				{
					Name:    "discrete-0",
					Address: uint16(0),
				},
				{
					Name:    "discrete-1",
					Address: uint16(1),
					Omit:    true,
				},
				{
					Name:        "discrete-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
		},
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
				},
				{
					Name:      "holding-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:        "holding-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
		},
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "input",
			Fields: []requestFieldDefinition{
				{
					Name:      "input-0",
					Address:   uint16(0),
					InputType: "INT16",
				},
				{
					Name:      "input-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:        "input-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
		},
	}

	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NotNil(t, modbus.requests[1])
	require.Len(t, modbus.requests[1].coil, 2)
	require.Len(t, modbus.requests[1].discrete, 1)
	require.Len(t, modbus.requests[1].holding, 1)
	require.Len(t, modbus.requests[1].input, 1)
}

func TestConfigurationPerRequestWithTags(t *testing.T) {
	modbus := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "request",
		Log:               testutil.Logger{},
	}
	modbus.Requests = []requestDefinition{
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "coil",
			Fields: []requestFieldDefinition{
				{
					Name:    "coil-0",
					Address: uint16(0),
				},
				{
					Name:    "coil-1",
					Address: uint16(1),
					Omit:    true,
				},
				{
					Name:        "coil-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
			Tags: map[string]string{
				"first":  "a",
				"second": "bb",
				"third":  "ccc",
			},
		},
		{
			SlaveID:      1,
			RegisterType: "coil",
			Fields: []requestFieldDefinition{
				{
					Name:    "coil-3",
					Address: uint16(6),
				},
				{
					Name:    "coil-4",
					Address: uint16(7),
					Omit:    true,
				},
				{
					Name:        "coil-5",
					Address:     uint16(8),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
			Tags: map[string]string{
				"first":  "a",
				"second": "bb",
				"third":  "ccc",
			},
		},
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "discrete",
			Fields: []requestFieldDefinition{
				{
					Name:    "discrete-0",
					Address: uint16(0),
				},
				{
					Name:    "discrete-1",
					Address: uint16(1),
					Omit:    true,
				},
				{
					Name:        "discrete-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
			Tags: map[string]string{
				"first":  "a",
				"second": "bb",
				"third":  "ccc",
			},
		},
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
				},
				{
					Name:      "holding-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:        "holding-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
			Tags: map[string]string{
				"first":  "a",
				"second": "bb",
				"third":  "ccc",
			},
		},
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "input",
			Fields: []requestFieldDefinition{
				{
					Name:      "input-0",
					Address:   uint16(0),
					InputType: "INT16",
				},
				{
					Name:      "input-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:        "input-2",
					Address:     uint16(2),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "FLOAT64",
					Measurement: "modbus",
				},
			},
			Tags: map[string]string{
				"first":  "a",
				"second": "bb",
				"third":  "ccc",
			},
		},
	}

	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NotNil(t, modbus.requests[1])
	require.Len(t, modbus.requests[1].coil, 2)
	require.Len(t, modbus.requests[1].discrete, 1)
	require.Len(t, modbus.requests[1].holding, 1)
	require.Len(t, modbus.requests[1].input, 1)

	expectedTags := map[string]string{
		"first":  "a",
		"second": "bb",
		"third":  "ccc",
	}
	require.Equal(t, expectedTags, modbus.requests[1].coil[0].tags)
	require.Equal(t, expectedTags, modbus.requests[1].coil[1].tags)
	require.Equal(t, expectedTags, modbus.requests[1].discrete[0].tags)
	require.Equal(t, expectedTags, modbus.requests[1].holding[0].tags)
	require.Equal(t, expectedTags, modbus.requests[1].input[0].tags)
}

func TestConfigurationPerRequestFail(t *testing.T) {
	tests := []struct {
		name     string
		requests []requestDefinition
		errormsg string
	}{
		{
			name: "empty field name (coil)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "coil",
					Fields: []requestFieldDefinition{
						{
							Address: uint16(15),
						},
					},
				},
			},
			errormsg: "configuraton invalid: empty field name in request for slave 1",
		},
		{
			name: "invalid byte-order (coil)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "AB",
					RegisterType: "coil",
					Fields:       []requestFieldDefinition{},
				},
			},
			errormsg: "configuraton invalid: unknown byte-order \"AB\"",
		},
		{
			name: "duplicate fields (coil)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "coil",
					Fields: []requestFieldDefinition{
						{
							Name:    "coil-0",
							Address: uint16(0),
						},
						{
							Name:    "coil-0",
							Address: uint16(1),
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"coil-0\" duplicated in measurement \"modbus\" (slave 1/\"coil\")",
		},
		{
			name: "duplicate fields multiple requests (coil)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "coil",
					Fields: []requestFieldDefinition{
						{
							Name:        "coil-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "coil",
					Fields: []requestFieldDefinition{
						{
							Name:        "coil-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"coil-0\" duplicated in measurement \"foo\" (slave 1/\"coil\")",
		},
		{
			name: "invalid byte-order (discrete)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "AB",
					RegisterType: "discrete",
					Fields:       []requestFieldDefinition{},
				},
			},
			errormsg: "configuraton invalid: unknown byte-order \"AB\"",
		},
		{
			name: "duplicate fields (discrete)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "discrete",
					Fields: []requestFieldDefinition{
						{
							Name:    "discrete-0",
							Address: uint16(0),
						},
						{
							Name:    "discrete-0",
							Address: uint16(1),
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"discrete-0\" duplicated in measurement \"modbus\" (slave 1/\"discrete\")",
		},
		{
			name: "duplicate fields multiple requests (discrete)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "discrete",
					Fields: []requestFieldDefinition{
						{
							Name:        "discrete-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "discrete",
					Fields: []requestFieldDefinition{
						{
							Name:        "discrete-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"discrete-0\" duplicated in measurement \"foo\" (slave 1/\"discrete\")",
		},
		{
			name: "invalid byte-order (holding)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "AB",
					RegisterType: "holding",
					Fields:       []requestFieldDefinition{},
				},
			},
			errormsg: "configuraton invalid: unknown byte-order \"AB\"",
		},
		{
			name: "invalid field name (holding)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Address: uint16(0),
						},
					},
				},
			},
			errormsg: "configuraton invalid: empty field name in request for slave 1",
		},
		{
			name: "invalid field input type (holding)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:    "holding-0",
							Address: uint16(0),
						},
					},
				},
			},
			errormsg: "cannot process configuraton: initializing field \"holding-0\" failed: invalid input datatype \"\" for determining field length",
		},
		{
			name: "invalid field output type (holding)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:       "holding-0",
							Address:    uint16(0),
							InputType:  "UINT16",
							OutputType: "UINT8",
						},
					},
				},
			},
			errormsg: "cannot process configuraton: initializing field \"holding-0\" failed: unknown output type \"UINT8\"",
		},
		{
			name: "duplicate fields (holding)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:    "holding-0",
							Address: uint16(0),
						},
						{
							Name:    "holding-0",
							Address: uint16(1),
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"holding-0\" duplicated in measurement \"modbus\" (slave 1/\"holding\")",
		},
		{
			name: "duplicate fields multiple requests (holding)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:        "holding-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:        "holding-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"holding-0\" duplicated in measurement \"foo\" (slave 1/\"holding\")",
		},
		{
			name: "invalid byte-order (input)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "AB",
					RegisterType: "input",
					Fields:       []requestFieldDefinition{},
				},
			},
			errormsg: "configuraton invalid: unknown byte-order \"AB\"",
		},
		{
			name: "invalid field name (input)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					RegisterType: "input",
					Fields: []requestFieldDefinition{
						{
							Address: uint16(0),
						},
					},
				},
			},
			errormsg: "configuraton invalid: empty field name in request for slave 1",
		},
		{
			name: "invalid field input type (input)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					RegisterType: "input",
					Fields: []requestFieldDefinition{
						{
							Name:    "input-0",
							Address: uint16(0),
						},
					},
				},
			},
			errormsg: "cannot process configuraton: initializing field \"input-0\" failed: invalid input datatype \"\" for determining field length",
		},
		{
			name: "invalid field output type (input)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					RegisterType: "input",
					Fields: []requestFieldDefinition{
						{
							Name:       "input-0",
							Address:    uint16(0),
							InputType:  "UINT16",
							OutputType: "UINT8",
						},
					},
				},
			},
			errormsg: "cannot process configuraton: initializing field \"input-0\" failed: unknown output type \"UINT8\"",
		},
		{
			name: "duplicate fields (input)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "input",
					Fields: []requestFieldDefinition{
						{
							Name:    "input-0",
							Address: uint16(0),
						},
						{
							Name:    "input-0",
							Address: uint16(1),
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"input-0\" duplicated in measurement \"modbus\" (slave 1/\"input\")",
		},
		{
			name: "duplicate fields multiple requests (input)",
			requests: []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "input",
					Fields: []requestFieldDefinition{
						{
							Name:        "input-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
				{
					SlaveID:      1,
					ByteOrder:    "ABCD",
					RegisterType: "input",
					Fields: []requestFieldDefinition{
						{
							Name:        "input-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
			},
			errormsg: "configuraton invalid: field \"input-0\" duplicated in measurement \"foo\" (slave 1/\"input\")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Modbus{
				Name:              "Test",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			plugin.Requests = tt.requests

			err := plugin.Init()
			require.Error(t, err)
			require.Equal(t, tt.errormsg, err.Error())
			require.Empty(t, plugin.requests)
		})
	}
}
