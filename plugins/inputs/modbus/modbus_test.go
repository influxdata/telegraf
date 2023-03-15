package modbus

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	mb "github.com/grid-x/modbus"
	"github.com/stretchr/testify/require"
	"github.com/tbrandon/mbserver"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestMain(m *testing.M) {
	telegraf.Debug = false
	os.Exit(m.Run())
}

func TestControllers(t *testing.T) {
	var tests = []struct {
		name       string
		controller string
		mode       string
		errmsg     string
	}{
		{
			name:       "TCP host",
			controller: "tcp://localhost:502",
		},
		{
			name:       "TCP mode auto",
			controller: "tcp://localhost:502",
			mode:       "auto",
		},
		{
			name:       "TCP mode TCP",
			controller: "tcp://localhost:502",
			mode:       "TCP",
		},
		{
			name:       "TCP mode RTUoverTCP",
			controller: "tcp://localhost:502",
			mode:       "RTUoverTCP",
		},
		{
			name:       "TCP mode ASCIIoverTCP",
			controller: "tcp://localhost:502",
			mode:       "ASCIIoverTCP",
		},
		{
			name:       "TCP invalid host",
			controller: "tcp://localhost",
			errmsg:     "address localhost: missing port in address",
		},
		{
			name:       "TCP invalid mode RTU",
			controller: "tcp://localhost:502",
			mode:       "RTU",
			errmsg:     "invalid transmission mode",
		},
		{
			name:       "TCP invalid mode ASCII",
			controller: "tcp://localhost:502",
			mode:       "ASCII",
			errmsg:     "invalid transmission mode",
		},
		{
			name:       "absolute file path",
			controller: "file:///dev/ttyUSB0",
		},
		{
			name:       "relative file path",
			controller: "file://dev/ttyUSB0",
		},
		{
			name:       "relative file path with dot",
			controller: "file://./dev/ttyUSB0",
		},
		{
			name:       "Windows COM-port",
			controller: "COM2",
		},
		{
			name:       "Windows COM-port file path",
			controller: "file://com2",
		},
		{
			name:       "serial mode auto",
			controller: "file:///dev/ttyUSB0",
			mode:       "auto",
		},
		{
			name:       "serial mode RTU",
			controller: "file:///dev/ttyUSB0",
			mode:       "RTU",
		},
		{
			name:       "serial mode ASCII",
			controller: "file:///dev/ttyUSB0",
			mode:       "ASCII",
		},
		{
			name:       "empty file path",
			controller: "file://",
			errmsg:     "invalid path for controller",
		},
		{
			name:       "empty controller",
			controller: "",
			errmsg:     "invalid path for controller",
		},
		{
			name:       "invalid scheme",
			controller: "foo://bar",
			errmsg:     "invalid controller",
		},
		{
			name:       "serial invalid mode TCP",
			controller: "file:///dev/ttyUSB0",
			mode:       "TCP",
			errmsg:     "invalid transmission mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Modbus{
				Name:             "dummy",
				Controller:       tt.controller,
				TransmissionMode: tt.mode,
				Log:              testutil.Logger{},
			}
			err := plugin.Init()
			if tt.errmsg != "" {
				require.ErrorContains(t, err, tt.errmsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCoils(t *testing.T) {
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

func TestRequestTypesCoil(t *testing.T) {
	tests := []struct {
		name        string
		address     uint16
		dataTypeOut string
		write       uint16
		read        interface{}
	}{
		{
			name:    "coil-1-off",
			address: 1,
			write:   0,
			read:    uint16(0),
		},
		{
			name:    "coil-2-on",
			address: 2,
			write:   0xFF00,
			read:    uint16(1),
		},
		{
			name:        "coil-3-false",
			address:     3,
			dataTypeOut: "BOOL",
			write:       0,
			read:        false,
		},
		{
			name:        "coil-4-true",
			address:     4,
			dataTypeOut: "BOOL",
			write:       0xFF00,
			read:        true,
		},
	}

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	for _, hrt := range tests {
		t.Run(hrt.name, func(t *testing.T) {
			_, err := client.WriteSingleCoil(hrt.address, hrt.write)
			require.NoError(t, err)

			modbus := Modbus{
				Name:              "TestRequestTypesCoil",
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
							Name:       hrt.name,
							OutputType: hrt.dataTypeOut,
							Address:    hrt.address,
						},
					},
				},
			}

			expected := []telegraf.Metric{
				testutil.MustMetric(
					"modbus",
					map[string]string{
						"type":     cCoils,
						"slave_id": "1",
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

func TestRequestTypesHoldingABCD(t *testing.T) {
	byteOrder := "ABCD"
	tests := []struct {
		name        string
		address     uint16
		byteOrder   string
		dataTypeIn  string
		dataTypeOut string
		scale       float64
		write       []byte
		read        interface{}
	}{
		{
			name:       "register10_uint8L",
			address:    10,
			dataTypeIn: "UINT8L",
			write:      []byte{0x18, 0x0d},
			read:       uint8(13),
		},
		{
			name:       "register10_uint8L-scale_.1",
			address:    10,
			dataTypeIn: "UINT8L",
			scale:      .1,
			write:      []byte{0x18, 0x0d},
			read:       float64(1.3),
		},
		{
			name:       "register10_uint8L_scale_10",
			address:    10,
			dataTypeIn: "UINT8L",
			scale:      10,
			write:      []byte{0x18, 0x0d},
			read:       float64(130),
		},
		{
			name:        "register10_uint8L_uint64",
			address:     10,
			dataTypeIn:  "UINT8L",
			dataTypeOut: "UINT64",
			write:       []byte{0x18, 0x0d},
			read:        uint64(13),
		},
		{
			name:        "register10_uint8L_int64",
			address:     10,
			dataTypeIn:  "UINT8L",
			dataTypeOut: "INT64",
			write:       []byte{0x18, 0x0d},
			read:        int64(13),
		},
		{
			name:        "register10_uint8L_float64",
			address:     10,
			dataTypeIn:  "UINT8L",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x18, 0x0d},
			read:        float64(13),
		},
		{
			name:       "register10_uint8L_float64_scale",
			address:    10,
			dataTypeIn: "UINT8L",
			scale:      1.0,
			write:      []byte{0x18, 0x0d},
			read:       float64(13),
		},
		{
			name:       "register15_int8L",
			address:    15,
			dataTypeIn: "UINT8L",
			write:      []byte{0x18, 0x0d},
			read:       uint8(13),
		},
		{
			name:       "register15_int8L-scale_.1",
			address:    15,
			dataTypeIn: "INT8L",
			scale:      .1,
			write:      []byte{0x18, 0x0d},
			read:       float64(1.3),
		},
		{
			name:       "register15_int8L_scale_10",
			address:    15,
			dataTypeIn: "INT8L",
			scale:      10,
			write:      []byte{0x18, 0x0d},
			read:       float64(130),
		},
		{
			name:        "register15_int8L_uint64",
			address:     15,
			dataTypeIn:  "INT8L",
			dataTypeOut: "UINT64",
			write:       []byte{0x18, 0x0d},
			read:        uint64(13),
		},
		{
			name:        "register15_int8L_int64",
			address:     15,
			dataTypeIn:  "INT8L",
			dataTypeOut: "INT64",
			write:       []byte{0x18, 0x0d},
			read:        int64(13),
		},
		{
			name:        "register15_int8L_float64",
			address:     15,
			dataTypeIn:  "INT8L",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x18, 0x0d},
			read:        float64(13),
		},
		{
			name:       "register15_int8L_float64_scale",
			address:    15,
			dataTypeIn: "INT8L",
			scale:      1.0,
			write:      []byte{0x18, 0x0d},
			read:       float64(13),
		},
		{
			name:       "register20_uint16",
			address:    20,
			dataTypeIn: "UINT16",
			write:      []byte{0x08, 0x98},
			read:       uint16(2200),
		},
		{
			name:       "register20_uint16-scale_.1",
			address:    20,
			dataTypeIn: "UINT16",
			scale:      .1,
			write:      []byte{0x08, 0x98},
			read:       float64(220),
		},
		{
			name:       "register20_uint16_scale_10",
			address:    20,
			dataTypeIn: "UINT16",
			scale:      10,
			write:      []byte{0x08, 0x98},
			read:       float64(22000),
		},
		{
			name:        "register20_uint16_uint64",
			address:     20,
			dataTypeIn:  "UINT16",
			dataTypeOut: "UINT64",
			write:       []byte{0x08, 0x98},
			read:        uint64(2200),
		},
		{
			name:        "register20_uint16_int64",
			address:     20,
			dataTypeIn:  "UINT16",
			dataTypeOut: "INT64",
			write:       []byte{0x08, 0x98},
			read:        int64(2200),
		},
		{
			name:        "register20_uint16_float64",
			address:     20,
			dataTypeIn:  "UINT16",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x08, 0x98},
			read:        float64(2200),
		},
		{
			name:       "register20_uint16_float64_scale",
			address:    20,
			dataTypeIn: "UINT16",
			scale:      1.0,
			write:      []byte{0x08, 0x98},
			read:       float64(2200),
		},
		{
			name:       "register30_int16",
			address:    30,
			dataTypeIn: "INT16",
			write:      []byte{0xf8, 0x98},
			read:       int16(-1896),
		},
		{
			name:       "register30_int16-scale_.1",
			address:    30,
			dataTypeIn: "INT16",
			scale:      .1,
			write:      []byte{0xf8, 0x98},
			read:       float64(-189.60000000000002),
		},
		{
			name:       "register30_int16_scale_10",
			address:    30,
			dataTypeIn: "INT16",
			scale:      10,
			write:      []byte{0xf8, 0x98},
			read:       float64(-18960),
		},
		{
			name:        "register30_int16_uint64",
			address:     30,
			dataTypeIn:  "INT16",
			dataTypeOut: "UINT64",
			write:       []byte{0xf8, 0x98},
			read:        uint64(18446744073709549720),
		},
		{
			name:        "register30_int16_int64",
			address:     30,
			dataTypeIn:  "INT16",
			dataTypeOut: "INT64",
			write:       []byte{0xf8, 0x98},
			read:        int64(-1896),
		},
		{
			name:        "register30_int16_float64",
			address:     30,
			dataTypeIn:  "INT16",
			dataTypeOut: "FLOAT64",
			write:       []byte{0xf8, 0x98},
			read:        float64(-1896),
		},
		{
			name:       "register30_int16_float64_scale",
			address:    30,
			dataTypeIn: "INT16",
			scale:      1.0,
			write:      []byte{0xf8, 0x98},
			read:       float64(-1896),
		},
		{
			name:       "register40_uint32",
			address:    40,
			dataTypeIn: "UINT32",
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       uint32(168496141),
		},
		{
			name:       "register40_uint32-scale_.1",
			address:    40,
			dataTypeIn: "UINT32",
			scale:      .1,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       float64(16849614.1),
		},
		{
			name:       "register40_uint32_scale_10",
			address:    40,
			dataTypeIn: "UINT32",
			scale:      10,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       float64(1684961410),
		},
		{
			name:        "register40_uint32_uint64",
			address:     40,
			dataTypeIn:  "UINT32",
			dataTypeOut: "UINT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:        uint64(168496141),
		},
		{
			name:        "register40_uint32_int64",
			address:     40,
			dataTypeIn:  "UINT32",
			dataTypeOut: "INT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:        int64(168496141),
		},
		{
			name:        "register40_uint32_float64",
			address:     40,
			dataTypeIn:  "UINT32",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:        float64(168496141),
		},
		{
			name:       "register40_uint32_float64_scale",
			address:    40,
			dataTypeIn: "UINT32",
			scale:      1.0,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       float64(168496141),
		},
		{
			name:       "register50_int32",
			address:    50,
			dataTypeIn: "INT32",
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       int32(-99939315),
		},
		{
			name:       "register50_int32-scale_.1",
			address:    50,
			dataTypeIn: "INT32",
			scale:      .1,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       float64(-9993931.5),
		},
		{
			name:       "register50_int32_scale_10",
			address:    50,
			dataTypeIn: "INT32",
			scale:      10,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       float64(-999393150),
		},
		{
			name:        "register50_int32_uint64",
			address:     50,
			dataTypeIn:  "INT32",
			dataTypeOut: "UINT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:        uint64(18446744073609612301),
		},
		{
			name:        "register50_int32_int64",
			address:     50,
			dataTypeIn:  "INT32",
			dataTypeOut: "INT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:        int64(-99939315),
		},
		{
			name:        "register50_int32_float64",
			address:     50,
			dataTypeIn:  "INT32",
			dataTypeOut: "FLOAT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:        float64(-99939315),
		},
		{
			name:       "register50_int32_float64_scale",
			address:    50,
			dataTypeIn: "INT32",
			scale:      1.0,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       float64(-99939315),
		},
		{
			name:       "register60_uint64",
			address:    60,
			dataTypeIn: "UINT64",
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       uint64(723685415333069058),
		},
		{
			name:       "register60_uint64-scale_.1",
			address:    60,
			dataTypeIn: "UINT64",
			scale:      .1,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(72368541533306905.8),
		},
		{
			name:       "register60_uint64_scale_10",
			address:    60,
			dataTypeIn: "UINT64",
			scale:      10,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(7236854153330690000), // quantization error
		},
		{
			name:        "register60_uint64_int64",
			address:     60,
			dataTypeIn:  "UINT64",
			dataTypeOut: "INT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        int64(723685415333069058),
		},
		{
			name:        "register60_uint64_float64",
			address:     60,
			dataTypeIn:  "UINT64",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        float64(723685415333069058),
		},
		{
			name:       "register60_uint64_float64_scale",
			address:    60,
			dataTypeIn: "UINT64",
			scale:      1.0,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(723685415333069058),
		},
		{
			name:       "register70_int64",
			address:    70,
			dataTypeIn: "INT64",
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       int64(-429236089273777918),
		},
		{
			name:       "register70_int64-scale_.1",
			address:    70,
			dataTypeIn: "INT64",
			scale:      .1,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(-42923608927377791.8),
		},
		{
			name:       "register70_int64_scale_10",
			address:    70,
			dataTypeIn: "INT64",
			scale:      10,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(-4292360892737779180),
		},
		{
			name:        "register70_int64_uint64",
			address:     70,
			dataTypeIn:  "INT64",
			dataTypeOut: "UINT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        uint64(18017507984435773698),
		},
		{
			name:        "register70_int64_float64",
			address:     70,
			dataTypeIn:  "INT64",
			dataTypeOut: "FLOAT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        float64(-429236089273777918),
		},
		{
			name:       "register70_int64_float64_scale",
			address:    70,
			dataTypeIn: "INT64",
			scale:      1.0,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(-429236089273777918),
		},
		{
			name:       "register80_float32",
			address:    80,
			dataTypeIn: "FLOAT32",
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float32(3.1415927410125732421875),
		},
		{
			name:       "register80_float32-scale_.1",
			address:    80,
			dataTypeIn: "FLOAT32",
			scale:      .1,
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float64(0.31415927410125732421875),
		},
		{
			name:       "register80_float32_scale_10",
			address:    80,
			dataTypeIn: "FLOAT32",
			scale:      10,
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float64(31.415927410125732421875),
		},
		{
			name:        "register80_float32_float64",
			address:     80,
			dataTypeIn:  "FLOAT32",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x40, 0x49, 0x0f, 0xdb},
			read:        float64(3.1415927410125732421875),
		},
		{
			name:       "register80_float32_float64_scale",
			address:    80,
			dataTypeIn: "FLOAT32",
			scale:      1.0,
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float64(3.1415927410125732421875),
		},
		{
			name:       "register90_float64",
			address:    90,
			dataTypeIn: "FLOAT64",
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(3.14159265359000006156975359772),
		},
		{
			name:       "register90_float64-scale_.1",
			address:    90,
			dataTypeIn: "FLOAT64",
			scale:      .1,
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(0.314159265359000006156975359772),
		},
		{
			name:       "register90_float64_scale_10",
			address:    90,
			dataTypeIn: "FLOAT64",
			scale:      10,
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(31.4159265359000006156975359772),
		},
		{
			name:       "register90_float64_float64_scale",
			address:    90,
			dataTypeIn: "FLOAT64",
			scale:      1.0,
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(3.14159265359000006156975359772),
		},
		{
			name:       "register100_float16",
			address:    100,
			dataTypeIn: "FLOAT16",
			write:      []byte{0xb8, 0x14},
			read:       float64(-0.509765625),
		},
		{
			name:       "register100_float16-scale_.1",
			address:    100,
			dataTypeIn: "FLOAT16",
			scale:      .1,
			write:      []byte{0xb8, 0x14},
			read:       float64(-0.0509765625),
		},
		{
			name:       "register100_float16_scale_10",
			address:    100,
			dataTypeIn: "FLOAT16",
			scale:      10,
			write:      []byte{0xb8, 0x14},
			read:       float64(-5.09765625),
		},
		{
			name:       "register100_float16_float64_scale",
			address:    100,
			dataTypeIn: "FLOAT16",
			scale:      1.0,
			write:      []byte{0xb8, 0x14},
			read:       float64(-0.509765625),
		},
	}

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	for _, hrt := range tests {
		t.Run(hrt.name, func(t *testing.T) {
			quantity := uint16(len(hrt.write) / 2)
			_, err := client.WriteMultipleRegisters(hrt.address, quantity, hrt.write)
			require.NoError(t, err)

			modbus := Modbus{
				Name:              "TestRequestTypesHoldingABCD",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			modbus.Requests = []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    byteOrder,
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:       hrt.name,
							InputType:  hrt.dataTypeIn,
							OutputType: hrt.dataTypeOut,
							Scale:      hrt.scale,
							Address:    hrt.address,
						},
					},
				},
			}

			expected := []telegraf.Metric{
				testutil.MustMetric(
					"modbus",
					map[string]string{
						"type":     cHoldingRegisters,
						"slave_id": "1",
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

func TestRequestTypesHoldingDCBA(t *testing.T) {
	byteOrder := "DCBA"
	tests := []struct {
		name        string
		address     uint16
		byteOrder   string
		dataTypeIn  string
		dataTypeOut string
		scale       float64
		write       []byte
		read        interface{}
	}{
		{
			name:       "register10_uint8L",
			address:    10,
			dataTypeIn: "UINT8L",
			write:      []byte{0x18, 0x0d},
			read:       uint8(13),
		},
		{
			name:       "register10_uint8L-scale_.1",
			address:    10,
			dataTypeIn: "UINT8L",
			scale:      .1,
			write:      []byte{0x18, 0x0d},
			read:       float64(1.3),
		},
		{
			name:       "register10_uint8L_scale_10",
			address:    10,
			dataTypeIn: "UINT8L",
			scale:      10,
			write:      []byte{0x18, 0x0d},
			read:       float64(130),
		},
		{
			name:        "register10_uint8L_uint64",
			address:     10,
			dataTypeIn:  "UINT8L",
			dataTypeOut: "UINT64",
			write:       []byte{0x18, 0x0d},
			read:        uint64(13),
		},
		{
			name:        "register10_uint8L_int64",
			address:     10,
			dataTypeIn:  "UINT8L",
			dataTypeOut: "INT64",
			write:       []byte{0x18, 0x0d},
			read:        int64(13),
		},
		{
			name:        "register10_uint8L_float64",
			address:     10,
			dataTypeIn:  "UINT8L",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x18, 0x0d},
			read:        float64(13),
		},
		{
			name:       "register10_uint8L_float64_scale",
			address:    10,
			dataTypeIn: "UINT8L",
			scale:      1.0,
			write:      []byte{0x18, 0x0d},
			read:       float64(13),
		},
		{
			name:       "register15_int8L",
			address:    15,
			dataTypeIn: "UINT8L",
			write:      []byte{0x18, 0x0d},
			read:       uint8(13),
		},
		{
			name:       "register15_int8L-scale_.1",
			address:    15,
			dataTypeIn: "INT8L",
			scale:      .1,
			write:      []byte{0x18, 0x0d},
			read:       float64(1.3),
		},
		{
			name:       "register15_int8L_scale_10",
			address:    15,
			dataTypeIn: "INT8L",
			scale:      10,
			write:      []byte{0x18, 0x0d},
			read:       float64(130),
		},
		{
			name:        "register15_int8L_uint64",
			address:     15,
			dataTypeIn:  "INT8L",
			dataTypeOut: "UINT64",
			write:       []byte{0x18, 0x0d},
			read:        uint64(13),
		},
		{
			name:        "register15_int8L_int64",
			address:     15,
			dataTypeIn:  "INT8L",
			dataTypeOut: "INT64",
			write:       []byte{0x18, 0x0d},
			read:        int64(13),
		},
		{
			name:        "register15_int8L_float64",
			address:     15,
			dataTypeIn:  "INT8L",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x18, 0x0d},
			read:        float64(13),
		},
		{
			name:       "register15_int8L_float64_scale",
			address:    15,
			dataTypeIn: "INT8L",
			scale:      1.0,
			write:      []byte{0x18, 0x0d},
			read:       float64(13),
		},
		{
			name:       "register20_uint16",
			address:    20,
			dataTypeIn: "UINT16",
			write:      []byte{0x08, 0x98},
			read:       uint16(2200),
		},
		{
			name:       "register20_uint16-scale_.1",
			address:    20,
			dataTypeIn: "UINT16",
			scale:      .1,
			write:      []byte{0x08, 0x98},
			read:       float64(220),
		},
		{
			name:       "register20_uint16_scale_10",
			address:    20,
			dataTypeIn: "UINT16",
			scale:      10,
			write:      []byte{0x08, 0x98},
			read:       float64(22000),
		},
		{
			name:        "register20_uint16_uint64",
			address:     20,
			dataTypeIn:  "UINT16",
			dataTypeOut: "UINT64",
			write:       []byte{0x08, 0x98},
			read:        uint64(2200),
		},
		{
			name:        "register20_uint16_int64",
			address:     20,
			dataTypeIn:  "UINT16",
			dataTypeOut: "INT64",
			write:       []byte{0x08, 0x98},
			read:        int64(2200),
		},
		{
			name:        "register20_uint16_float64",
			address:     20,
			dataTypeIn:  "UINT16",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x08, 0x98},
			read:        float64(2200),
		},
		{
			name:       "register20_uint16_float64_scale",
			address:    20,
			dataTypeIn: "UINT16",
			scale:      1.0,
			write:      []byte{0x08, 0x98},
			read:       float64(2200),
		},
		{
			name:       "register30_int16",
			address:    30,
			dataTypeIn: "INT16",
			write:      []byte{0xf8, 0x98},
			read:       int16(-1896),
		},
		{
			name:       "register30_int16-scale_.1",
			address:    30,
			dataTypeIn: "INT16",
			scale:      .1,
			write:      []byte{0xf8, 0x98},
			read:       float64(-189.60000000000002),
		},
		{
			name:       "register30_int16_scale_10",
			address:    30,
			dataTypeIn: "INT16",
			scale:      10,
			write:      []byte{0xf8, 0x98},
			read:       float64(-18960),
		},
		{
			name:        "register30_int16_uint64",
			address:     30,
			dataTypeIn:  "INT16",
			dataTypeOut: "UINT64",
			write:       []byte{0xf8, 0x98},
			read:        uint64(18446744073709549720),
		},
		{
			name:        "register30_int16_int64",
			address:     30,
			dataTypeIn:  "INT16",
			dataTypeOut: "INT64",
			write:       []byte{0xf8, 0x98},
			read:        int64(-1896),
		},
		{
			name:        "register30_int16_float64",
			address:     30,
			dataTypeIn:  "INT16",
			dataTypeOut: "FLOAT64",
			write:       []byte{0xf8, 0x98},
			read:        float64(-1896),
		},
		{
			name:       "register30_int16_float64_scale",
			address:    30,
			dataTypeIn: "INT16",
			scale:      1.0,
			write:      []byte{0xf8, 0x98},
			read:       float64(-1896),
		},
		{
			name:       "register40_uint32",
			address:    40,
			dataTypeIn: "UINT32",
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       uint32(168496141),
		},
		{
			name:       "register40_uint32-scale_.1",
			address:    40,
			dataTypeIn: "UINT32",
			scale:      .1,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       float64(16849614.1),
		},
		{
			name:       "register40_uint32_scale_10",
			address:    40,
			dataTypeIn: "UINT32",
			scale:      10,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       float64(1684961410),
		},
		{
			name:        "register40_uint32_uint64",
			address:     40,
			dataTypeIn:  "UINT32",
			dataTypeOut: "UINT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:        uint64(168496141),
		},
		{
			name:        "register40_uint32_int64",
			address:     40,
			dataTypeIn:  "UINT32",
			dataTypeOut: "INT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:        int64(168496141),
		},
		{
			name:        "register40_uint32_float64",
			address:     40,
			dataTypeIn:  "UINT32",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:        float64(168496141),
		},
		{
			name:       "register40_uint32_float64_scale",
			address:    40,
			dataTypeIn: "UINT32",
			scale:      1.0,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d},
			read:       float64(168496141),
		},
		{
			name:       "register50_int32",
			address:    50,
			dataTypeIn: "INT32",
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       int32(-99939315),
		},
		{
			name:       "register50_int32-scale_.1",
			address:    50,
			dataTypeIn: "INT32",
			scale:      .1,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       float64(-9993931.5),
		},
		{
			name:       "register50_int32_scale_10",
			address:    50,
			dataTypeIn: "INT32",
			scale:      10,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       float64(-999393150),
		},
		{
			name:        "register50_int32_uint64",
			address:     50,
			dataTypeIn:  "INT32",
			dataTypeOut: "UINT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:        uint64(18446744073609612301),
		},
		{
			name:        "register50_int32_int64",
			address:     50,
			dataTypeIn:  "INT32",
			dataTypeOut: "INT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:        int64(-99939315),
		},
		{
			name:        "register50_int32_float64",
			address:     50,
			dataTypeIn:  "INT32",
			dataTypeOut: "FLOAT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:        float64(-99939315),
		},
		{
			name:       "register50_int32_float64_scale",
			address:    50,
			dataTypeIn: "INT32",
			scale:      1.0,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d},
			read:       float64(-99939315),
		},
		{
			name:       "register60_uint64",
			address:    60,
			dataTypeIn: "UINT64",
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       uint64(723685415333069058),
		},
		{
			name:       "register60_uint64-scale_.1",
			address:    60,
			dataTypeIn: "UINT64",
			scale:      .1,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(72368541533306905.8),
		},
		{
			name:       "register60_uint64_scale_10",
			address:    60,
			dataTypeIn: "UINT64",
			scale:      10,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(7236854153330690000), // quantization error
		},
		{
			name:        "register60_uint64_int64",
			address:     60,
			dataTypeIn:  "UINT64",
			dataTypeOut: "INT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        int64(723685415333069058),
		},
		{
			name:        "register60_uint64_float64",
			address:     60,
			dataTypeIn:  "UINT64",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        float64(723685415333069058),
		},
		{
			name:       "register60_uint64_float64_scale",
			address:    60,
			dataTypeIn: "UINT64",
			scale:      1.0,
			write:      []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(723685415333069058),
		},
		{
			name:       "register70_int64",
			address:    70,
			dataTypeIn: "INT64",
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       int64(-429236089273777918),
		},
		{
			name:       "register70_int64-scale_.1",
			address:    70,
			dataTypeIn: "INT64",
			scale:      .1,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(-42923608927377791.8),
		},
		{
			name:       "register70_int64_scale_10",
			address:    70,
			dataTypeIn: "INT64",
			scale:      10,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(-4292360892737779180),
		},
		{
			name:        "register70_int64_uint64",
			address:     70,
			dataTypeIn:  "INT64",
			dataTypeOut: "UINT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        uint64(18017507984435773698),
		},
		{
			name:        "register70_int64_float64",
			address:     70,
			dataTypeIn:  "INT64",
			dataTypeOut: "FLOAT64",
			write:       []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:        float64(-429236089273777918),
		},
		{
			name:       "register70_int64_float64_scale",
			address:    70,
			dataTypeIn: "INT64",
			scale:      1.0,
			write:      []byte{0xfa, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02},
			read:       float64(-429236089273777918),
		},
		{
			name:       "register80_float32",
			address:    80,
			dataTypeIn: "FLOAT32",
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float32(3.1415927410125732421875),
		},
		{
			name:       "register80_float32-scale_.1",
			address:    80,
			dataTypeIn: "FLOAT32",
			scale:      .1,
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float64(0.31415927410125732421875),
		},
		{
			name:       "register80_float32_scale_10",
			address:    80,
			dataTypeIn: "FLOAT32",
			scale:      10,
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float64(31.415927410125732421875),
		},
		{
			name:        "register80_float32_float64",
			address:     80,
			dataTypeIn:  "FLOAT32",
			dataTypeOut: "FLOAT64",
			write:       []byte{0x40, 0x49, 0x0f, 0xdb},
			read:        float64(3.1415927410125732421875),
		},
		{
			name:       "register80_float32_float64_scale",
			address:    80,
			dataTypeIn: "FLOAT32",
			scale:      1.0,
			write:      []byte{0x40, 0x49, 0x0f, 0xdb},
			read:       float64(3.1415927410125732421875),
		},
		{
			name:       "register90_float64",
			address:    90,
			dataTypeIn: "FLOAT64",
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(3.14159265359000006156975359772),
		},
		{
			name:       "register90_float64-scale_.1",
			address:    90,
			dataTypeIn: "FLOAT64",
			scale:      .1,
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(0.314159265359000006156975359772),
		},
		{
			name:       "register90_float64_scale_10",
			address:    90,
			dataTypeIn: "FLOAT64",
			scale:      10,
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(31.4159265359000006156975359772),
		},
		{
			name:       "register90_float64_float64_scale",
			address:    90,
			dataTypeIn: "FLOAT64",
			scale:      1.0,
			write:      []byte{0x40, 0x09, 0x21, 0xfb, 0x54, 0x44, 0x2e, 0xea},
			read:       float64(3.14159265359000006156975359772),
		},
		{
			name:       "register100_float16",
			address:    100,
			dataTypeIn: "FLOAT16",
			write:      []byte{0xb8, 0x14},
			read:       float64(-0.509765625),
		},
		{
			name:       "register100_float16-scale_.1",
			address:    100,
			dataTypeIn: "FLOAT16",
			scale:      .1,
			write:      []byte{0xb8, 0x14},
			read:       float64(-0.0509765625),
		},
		{
			name:       "register100_float16_scale_10",
			address:    100,
			dataTypeIn: "FLOAT16",
			scale:      10,
			write:      []byte{0xb8, 0x14},
			read:       float64(-5.09765625),
		},
		{
			name:       "register100_float16_float64_scale",
			address:    100,
			dataTypeIn: "FLOAT16",
			scale:      1.0,
			write:      []byte{0xb8, 0x14},
			read:       float64(-0.509765625),
		},
	}

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)

	for _, hrt := range tests {
		t.Run(hrt.name, func(t *testing.T) {
			quantity := uint16(len(hrt.write) / 2)
			invert := make([]byte, 0, len(hrt.write))
			for i := len(hrt.write) - 1; i >= 0; i-- {
				invert = append(invert, hrt.write[i])
			}
			_, err := client.WriteMultipleRegisters(hrt.address, quantity, invert)
			require.NoError(t, err)

			modbus := Modbus{
				Name:              "TestRequestTypesHoldingDCBA",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			modbus.Requests = []requestDefinition{
				{
					SlaveID:      1,
					ByteOrder:    byteOrder,
					RegisterType: "holding",
					Fields: []requestFieldDefinition{
						{
							Name:       hrt.name,
							InputType:  hrt.dataTypeIn,
							OutputType: hrt.dataTypeOut,
							Scale:      hrt.scale,
							Address:    hrt.address,
						},
					},
				},
			}

			expected := []telegraf.Metric{
				testutil.MustMetric(
					"modbus",
					map[string]string{
						"type":     cHoldingRegisters,
						"slave_id": "1",
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

	require.NoError(t, modbus.Gather(&acc))
	require.Len(t, acc.Errors, 1)
	require.EqualError(t, acc.FirstError(), "slave 1: modbus: exception '6' (server device busy), function '129'")
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
		},
	)

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

	require.NoError(t, modbus.Gather(&acc))
	require.Len(t, acc.Errors, 1)
	require.EqualError(t, acc.FirstError(), "slave 1: modbus: exception '1' (illegal function), function '129'")
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
					OutputType:  "UINT16",
					Measurement: "modbus",
				},
				{
					Name:        "coil-3",
					Address:     uint16(3),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "BOOL",
					Measurement: "modbus",
				},
			},
		},
		{
			SlaveID:      1,
			RegisterType: "coil",
			Fields: []requestFieldDefinition{
				{
					Name:    "coil-4",
					Address: uint16(6),
				},
				{
					Name:    "coil-5",
					Address: uint16(7),
					Omit:    true,
				},
				{
					Name:        "coil-6",
					Address:     uint16(8),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "UINT16",
					Measurement: "modbus",
				},
				{
					Name:        "coil-7",
					Address:     uint16(9),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "BOOL",
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
					OutputType:  "UINT16",
					Measurement: "modbus",
				},
				{
					Name:        "discrete-3",
					Address:     uint16(3),
					InputType:   "INT64",
					Scale:       1.2,
					OutputType:  "BOOL",
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
					OutputType:  "UINT16",
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
					OutputType:  "UINT16",
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
					OutputType:  "UINT16",
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
			errormsg: "configuration invalid: empty field name in request for slave 1",
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
			errormsg: "configuration invalid: unknown byte-order \"AB\"",
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
			errormsg: "configuration invalid: field \"coil-0\" duplicated in measurement \"modbus\" (slave 1/\"coil\")",
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
			errormsg: "configuration invalid: field \"coil-0\" duplicated in measurement \"foo\" (slave 1/\"coil\")",
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
			errormsg: "configuration invalid: unknown byte-order \"AB\"",
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
			errormsg: "configuration invalid: field \"discrete-0\" duplicated in measurement \"modbus\" (slave 1/\"discrete\")",
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
			errormsg: "configuration invalid: field \"discrete-0\" duplicated in measurement \"foo\" (slave 1/\"discrete\")",
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
			errormsg: "configuration invalid: unknown byte-order \"AB\"",
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
			errormsg: "configuration invalid: empty field name in request for slave 1",
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
			errormsg: "cannot process configuration: initializing field \"holding-0\" failed: invalid input datatype \"\" for determining field length",
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
			errormsg: `configuration invalid: unknown output data-type "UINT8" for field "holding-0"`,
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
			errormsg: "configuration invalid: field \"holding-0\" duplicated in measurement \"modbus\" (slave 1/\"holding\")",
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
			errormsg: "configuration invalid: field \"holding-0\" duplicated in measurement \"foo\" (slave 1/\"holding\")",
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
			errormsg: "configuration invalid: unknown byte-order \"AB\"",
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
			errormsg: "configuration invalid: empty field name in request for slave 1",
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
			errormsg: "cannot process configuration: initializing field \"input-0\" failed: invalid input datatype \"\" for determining field length",
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
			errormsg: `configuration invalid: unknown output data-type "UINT8" for field "input-0"`,
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
			errormsg: "configuration invalid: field \"input-0\" duplicated in measurement \"modbus\" (slave 1/\"input\")",
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
			errormsg: "configuration invalid: field \"input-0\" duplicated in measurement \"foo\" (slave 1/\"input\")",
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

func TestConfigurationMaxExtraRegisterFail(t *testing.T) {
	tests := []struct {
		name     string
		requests []requestDefinition
		errormsg string
	}{{
		name: "MaxExtraRegister too large",
		requests: []requestDefinition{
			{
				SlaveID:           1,
				ByteOrder:         "ABCD",
				RegisterType:      "input",
				Optimization:      "max_insert",
				MaxExtraRegisters: 5000,
				Fields: []requestFieldDefinition{
					{
						Name:        "input-0",
						Address:     uint16(0),
						Measurement: "foo",
					},
				},
			},
		},
		errormsg: "configuration invalid: optimization_max_register_fill has to be between 1 and 125",
	},
		{
			name: "MaxExtraRegister too small",
			requests: []requestDefinition{
				{
					SlaveID:           1,
					ByteOrder:         "ABCD",
					RegisterType:      "input",
					Optimization:      "max_insert",
					MaxExtraRegisters: 0,
					Fields: []requestFieldDefinition{
						{
							Name:        "input-0",
							Address:     uint16(0),
							Measurement: "foo",
						},
					},
				},
			},
			errormsg: "configuration invalid: optimization_max_register_fill has to be between 1 and 125",
		}}

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

func TestRequestsStartingWithOmits(t *testing.T) {
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
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
					Omit:      true,
				},
				{
					Name:      "holding-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:      "holding-2",
					Address:   uint16(2),
					InputType: "INT16",
				},
			},
		},
	}
	require.NoError(t, modbus.Init())
	require.NotEmpty(t, modbus.requests)
	require.NotNil(t, modbus.requests[1])
	require.Equal(t, uint16(0), modbus.requests[1].holding[0].address)

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	handler := mb.NewTCPClientHandler("localhost:1502")
	require.NoError(t, handler.Connect())
	defer handler.Close()
	client := mb.NewClient(handler)
	_, err := client.WriteMultipleRegisters(uint16(0), 3, []byte{0x00, 0x01, 0x00, 0x02, 0x00, 0x03})
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cHoldingRegisters,
				"slave_id": strconv.Itoa(int(modbus.Requests[0].SlaveID)),
				"name":     modbus.Name,
			},
			map[string]interface{}{"holding-2": int16(3)},
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Gather(&acc))
	acc.Wait(len(expected))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestRequestsWithOmittedFieldsOnly(t *testing.T) {
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
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
					Omit:      true,
				},
				{
					Name:      "holding-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:      "holding-2",
					Address:   uint16(2),
					InputType: "INT16",
					Omit:      true,
				},
			},
		},
	}
	require.NoError(t, modbus.Init())
	require.Empty(t, modbus.requests)
}

func TestRequestsGroupWithOmittedFieldsOnly(t *testing.T) {
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
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
					Omit:      true,
				},
				{
					Name:      "holding-1",
					Address:   uint16(1),
					InputType: "UINT16",
					Omit:      true,
				},
				{
					Name:      "holding-2",
					Address:   uint16(2),
					InputType: "INT16",
					Omit:      true,
				},
				{
					Name:      "holding-8",
					Address:   uint16(8),
					InputType: "INT16",
				},
			},
		},
	}
	require.NoError(t, modbus.Init())
	require.Len(t, modbus.requests, 1)
	require.NotNil(t, modbus.requests[1])
	require.Len(t, modbus.requests[1].holding, 1)
	require.Equal(t, uint16(8), modbus.requests[1].holding[0].address)
	require.Equal(t, uint16(1), modbus.requests[1].holding[0].length)
}

func TestRequestsEmptyFields(t *testing.T) {
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
			RegisterType: "holding",
		},
	}
	err := modbus.Init()
	require.EqualError(t, err, `configuration invalid: found request section without fields`)
}

func TestMultipleSlavesOneFail(t *testing.T) {
	telegraf.Debug = true
	modbus := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		Retries:           1,
		ConfigurationType: "request",
		Log:               testutil.Logger{},
	}
	modbus.Requests = []requestDefinition{
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
			},
		},
		{
			SlaveID:      2,
			ByteOrder:    "ABCD",
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
				},
			},
		},
		{
			SlaveID:      3,
			ByteOrder:    "ABCD",
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-0",
					Address:   uint16(0),
					InputType: "INT16",
				},
			},
		},
	}
	require.NoError(t, modbus.Init())

	serv := mbserver.NewServer()
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	serv.RegisterFunctionHandler(3,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			tcpframe, ok := frame.(*mbserver.TCPFrame)
			if !ok {
				return nil, &mbserver.IllegalFunction
			}

			if tcpframe.Device == 2 {
				// Simulate device 2 being unavailable
				return []byte{}, &mbserver.GatewayTargetDeviceFailedtoRespond
			}
			return []byte{0x02, 0x00, 0x42}, &mbserver.Success
		},
	)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cHoldingRegisters,
				"slave_id": "1",
				"name":     modbus.Name,
			},
			map[string]interface{}{"holding-0": int16(0x42)},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"modbus",
			map[string]string{
				"type":     cHoldingRegisters,
				"slave_id": "3",
				"name":     modbus.Name,
			},
			map[string]interface{}{"holding-0": int16(0x42)},
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, modbus.Gather(&acc))
	acc.Wait(len(expected))
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
	require.Len(t, acc.Errors, 1)
	require.EqualError(t, acc.FirstError(), "slave 2: modbus: exception '11' (gateway target device failed to respond), function '131'")
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	// Compare options
	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.SortMetrics(),
	}

	// Register the plugin
	inputs.Add("modbus", func() telegraf.Input { return &Modbus{} })

	// Define a function to return the register value as data
	readFunc := func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
		data := frame.GetData()
		register := binary.BigEndian.Uint16(data[0:2])
		numRegs := binary.BigEndian.Uint16(data[2:4])

		// Add the length in bytes and the register to the returned data
		buf := make([]byte, 2*numRegs+1)
		buf[0] = byte(2 * numRegs)
		switch numRegs {
		case 1: // 16-bit
			binary.BigEndian.PutUint16(buf[1:], register)
		case 2: // 32-bit
			binary.BigEndian.PutUint32(buf[1:], uint32(register))
		case 4: // 64-bit
			binary.BigEndian.PutUint64(buf[1:], uint64(register))
		}
		return buf, &mbserver.Success
	}

	// Setup a Modbus server to test against
	serv := mbserver.NewServer()
	serv.RegisterFunctionHandler(mb.FuncCodeReadInputRegisters, readFunc)
	serv.RegisterFunctionHandler(mb.FuncCodeReadHoldingRegisters, readFunc)
	require.NoError(t, serv.ListenTCP("localhost:1502"))
	defer serv.Close()

	// Run the test cases
	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedOutputFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")
		initErrorFilename := filepath.Join(testcasePath, "init.err")

		t.Run(f.Name(), func(t *testing.T) {
			// Read the expected error for the init call if any
			var expectedInitError string
			if _, err := os.Stat(initErrorFilename); err == nil {
				e, err := testutil.ParseLinesFromFile(initErrorFilename)
				require.NoError(t, err)
				require.Len(t, e, 1)
				expectedInitError = e[0]
			}

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedOutputFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedOutputFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected error if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				e, err := testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, e)
				expectedErrors = e
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Extract the plugin and make sure it connects to our dummy
			// server
			plugin := cfg.Inputs[0].Input.(*Modbus)
			plugin.Controller = "tcp://localhost:1502"

			// Init the plugin.
			err := plugin.Init()
			if expectedInitError != "" {
				require.ErrorContains(t, err, expectedInitError)
				return
			}
			require.NoError(t, err)

			// Gather data
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			if len(acc.Errors) > 0 {
				var actualErrorMsgs []string
				for _, err := range acc.Errors {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
				require.ElementsMatch(t, actualErrorMsgs, expectedErrors)
			}

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

type rangeDefinition struct {
	start     uint16
	count     uint16
	increment uint16
	length    uint16
	dtype     string
	omit      bool
}

type requestExpectation struct {
	fields []rangeDefinition
	req    request
}

func generateRequestDefinitions(ranges []rangeDefinition) []requestFieldDefinition {
	var fields []requestFieldDefinition

	id := 0
	for _, r := range ranges {
		if r.increment == 0 {
			r.increment = r.length
		}
		for i := uint16(0); i < r.count; i++ {
			f := requestFieldDefinition{
				Name:      fmt.Sprintf("holding-%d", id),
				Address:   r.start + i*r.increment,
				InputType: r.dtype,
				Omit:      r.omit,
			}
			fields = append(fields, f)
			id++
		}
	}
	return fields
}

func generateExpectation(defs []requestExpectation) []request {
	requests := make([]request, 0, len(defs))
	for _, def := range defs {
		r := def.req
		r.fields = make([]field, 0)
		for _, d := range def.fields {
			if d.increment == 0 {
				d.increment = d.length
			}
			for i := uint16(0); i < d.count; i++ {
				f := field{
					address: d.start + i*d.increment,
					length:  d.length,
				}
				r.fields = append(r.fields, f)
			}
		}
		requests = append(requests, r)
	}
	return requests
}

func requireEqualRequests(t *testing.T, expected, actual []request) {
	require.Equal(t, len(expected), len(actual), "request size mismatch")

	for i, e := range expected {
		a := actual[i]
		require.Equalf(t, e.address, a.address, "address mismatch in request %d", i)
		require.Equalf(t, e.length, a.length, "length mismatch in request %d", i)
		require.Equalf(t, len(e.fields), len(a.fields), "no. fields mismatch in request %d", i)
		for j, ef := range e.fields {
			af := a.fields[j]
			require.Equalf(t, ef.address, af.address, "address mismatch in field %d of request %d", j, i)
			require.Equalf(t, ef.length, af.length, "length mismatch in field %d of request %d", j, i)
		}
	}
}

func TestRequestOptimizationShrink(t *testing.T) {
	maxsize := maxQuantityHoldingRegisters
	tests := []struct {
		name     string
		inputs   []rangeDefinition
		expected []requestExpectation
	}{
		{
			name: "no omit",
			inputs: []rangeDefinition{
				{0, 2 * maxQuantityHoldingRegisters, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 0, count: maxsize, length: 1}},
					req:    request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{{start: maxsize, count: maxsize, length: 1}},
					req:    request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "borders",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, maxsize - 2, 1, 1, "INT16", true},
				{maxsize - 1, 2, 1, 1, "INT16", false},
				{maxsize + 1, maxsize - 2, 1, 1, "INT16", true},
				{2*maxsize - 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: maxsize - 1, count: 1, length: 1},
					},
					req: request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize, count: 1, length: 1},
						{start: 2*maxsize - 1, count: 1, length: 1},
					},
					req: request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "borders with gap",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, maxsize - 2, 1, 1, "INT16", true},
				{maxsize - 1, 2, 1, 1, "INT16", false},
				{maxsize + 1, 4, 1, 1, "INT16", true},
				{2*maxsize - 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: maxsize - 1, count: 1, length: 1},
					},
					req: request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{{start: maxsize, count: 1, length: 1}},
					req:    request{address: maxsize, length: 1},
				},
				{
					fields: []rangeDefinition{{start: 2*maxsize - 1, count: 1, length: 1}},
					req:    request{address: 2*maxsize - 1, length: 1},
				},
			},
		},
		{
			name: "large gaps",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{{start: maxsize - 2, count: 5, length: 1}},
					req:    request{address: maxsize - 2, length: 5},
				},
				{
					fields: []rangeDefinition{{start: maxsize + 42, count: 2, length: 1}},
					req:    request{address: maxsize + 42, length: 2},
				},
			},
		},
		{
			name: "large gaps filled",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, 17, 1, 1, "INT16", true},
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: 18, count: 3, length: 1},
						{start: maxsize - 2, count: 2, length: 1},
					},
					req: request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize, count: 3, length: 1},
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize, length: 44},
				},
			},
		},
		{
			name: "large gaps filled with offset",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 18, count: 3, length: 1},
						{start: maxsize - 2, count: 5, length: 1},
					},
					req: request{address: 18, length: 110},
				},
				{
					fields: []rangeDefinition{{start: maxsize + 42, count: 2, length: 1}},
					req:    request{address: maxsize + 42, length: 2},
				},
			},
		},
		{
			name: "worst case",
			inputs: []rangeDefinition{
				{0, maxsize, 2, 1, "INT16", false},
				{1, maxsize, 2, 1, "INT16", true},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 0, count: maxsize/2 + 1, increment: 2, length: 1}},
					req:    request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{{start: maxsize + 1, count: maxsize / 2, increment: 2, length: 1}},
					req:    request{address: maxsize + 1, length: maxsize - 2},
				},
			},
		},
		{
			name: "from PR #11106",
			inputs: []rangeDefinition{
				{0, 2, 1, 1, "INT16", true},
				{2, 1, 1, 1, "INT16", false},
				{3, 2*maxsize + 1, 1, 1, "INT16", true},
				{2*maxsize + 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 2, count: 1, length: 1}},
					req:    request{address: 2, length: 1},
				},
				{
					fields: []rangeDefinition{{start: 2*maxsize + 1, count: 1, length: 1}},
					req:    request{address: 2*maxsize + 1, length: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the input structure and the expectation
			requestFields := generateRequestDefinitions(tt.inputs)
			expected := generateExpectation(tt.expected)

			// Setup the plugin
			slaveID := byte(1)
			plugin := Modbus{
				Name:              "Test",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			plugin.Requests = []requestDefinition{
				{
					SlaveID:      slaveID,
					ByteOrder:    "ABCD",
					RegisterType: "holding",
					Optimization: "shrink",
					Fields:       requestFields,
				},
			}
			require.NoError(t, plugin.Init())
			require.NotEmpty(t, plugin.requests)
			require.Contains(t, plugin.requests, slaveID)
			requireEqualRequests(t, expected, plugin.requests[slaveID].holding)
		})
	}
}

func TestRequestOptimizationRearrange(t *testing.T) {
	maxsize := maxQuantityHoldingRegisters
	tests := []struct {
		name     string
		inputs   []rangeDefinition
		expected []requestExpectation
	}{
		{
			name: "no omit",
			inputs: []rangeDefinition{
				{0, 2 * maxQuantityHoldingRegisters, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 0, count: maxsize, length: 1}},
					req:    request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{{start: maxsize, count: maxsize, length: 1}},
					req:    request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "borders",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, maxsize - 2, 1, 1, "INT16", true},
				{maxsize - 1, 2, 1, 1, "INT16", false},
				{maxsize + 1, maxsize - 2, 1, 1, "INT16", true},
				{2*maxsize - 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: maxsize - 1, count: 1, length: 1},
					},
					req: request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize, count: 1, length: 1},
						{start: 2*maxsize - 1, count: 1, length: 1},
					},
					req: request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "borders with gap",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, maxsize - 2, 1, 1, "INT16", true},
				{maxsize - 1, 2, 1, 1, "INT16", false},
				{maxsize + 1, 4, 1, 1, "INT16", true},
				{2*maxsize - 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 0, count: 1, length: 1}},
					req:    request{address: 0, length: 1},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 1, count: 1, length: 1},
						{start: maxsize, count: 1, length: 1},
					},
					req: request{address: maxsize - 1, length: 2},
				},
				{
					fields: []rangeDefinition{{start: 2*maxsize - 1, count: 1, length: 1}},
					req:    request{address: 2*maxsize - 1, length: 1},
				},
			},
		},
		{
			name: "large gaps",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{{start: maxsize - 2, count: 5, length: 1}},
					req:    request{address: maxsize - 2, length: 5},
				},
				{
					fields: []rangeDefinition{{start: maxsize + 42, count: 2, length: 1}},
					req:    request{address: maxsize + 42, length: 2},
				},
			},
		},
		{
			name: "large gaps filled",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, 17, 1, 1, "INT16", true},
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: 18, count: 3, length: 1},
					},
					req: request{address: 0, length: 21},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize - 2, length: 46},
				},
			},
		},
		{
			name: "large gaps filled with offset",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize - 2, length: 46},
				},
			},
		},
		{
			name: "from PR #11106",
			inputs: []rangeDefinition{
				{0, 2, 1, 1, "INT16", true},
				{2, 1, 1, 1, "INT16", false},
				{3, 2*maxsize + 1, 1, 1, "INT16", true},
				{2*maxsize + 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 2, count: 1, length: 1}},
					req:    request{address: 2, length: 1},
				},
				{
					fields: []rangeDefinition{{start: 2*maxsize + 1, count: 1, length: 1}},
					req:    request{address: 2*maxsize + 1, length: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the input structure and the expectation
			requestFields := generateRequestDefinitions(tt.inputs)
			expected := generateExpectation(tt.expected)

			// Setup the plugin
			slaveID := byte(1)
			plugin := Modbus{
				Name:              "Test",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			plugin.Requests = []requestDefinition{
				{
					SlaveID:      slaveID,
					ByteOrder:    "ABCD",
					RegisterType: "holding",
					Optimization: "rearrange",
					Fields:       requestFields,
				},
			}
			require.NoError(t, plugin.Init())
			require.NotEmpty(t, plugin.requests)
			require.Contains(t, plugin.requests, slaveID)
			requireEqualRequests(t, expected, plugin.requests[slaveID].holding)
		})
	}
}

func TestRequestOptimizationAggressive(t *testing.T) {
	maxsize := maxQuantityHoldingRegisters
	tests := []struct {
		name     string
		inputs   []rangeDefinition
		expected []requestExpectation
	}{
		{
			name: "no omit",
			inputs: []rangeDefinition{
				{0, 2 * maxQuantityHoldingRegisters, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 0, count: maxsize, length: 1}},
					req:    request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{{start: maxsize, count: maxsize, length: 1}},
					req:    request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "borders",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, maxsize - 2, 1, 1, "INT16", true},
				{maxsize - 1, 2, 1, 1, "INT16", false},
				{maxsize + 1, maxsize - 2, 1, 1, "INT16", true},
				{2*maxsize - 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: maxsize - 1, count: 1, length: 1},
					},
					req: request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize, count: 1, length: 1},
						{start: 2*maxsize - 1, count: 1, length: 1},
					},
					req: request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "borders with gap",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, maxsize - 2, 1, 1, "INT16", true},
				{maxsize - 1, 2, 1, 1, "INT16", false},
				{maxsize + 1, 4, 1, 1, "INT16", true},
				{2*maxsize - 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: maxsize - 1, count: 1, length: 1},
					},
					req: request{address: 0, length: maxsize},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize, count: 1, length: 1},
						{start: 2*maxsize - 1, count: 1, length: 1},
					},
					req: request{address: maxsize, length: maxsize},
				},
			},
		},
		{
			name: "large gaps",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize - 2, length: 46},
				},
			},
		},
		{
			name: "large gaps filled",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, 17, 1, 1, "INT16", true},
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
						{start: 18, count: 3, length: 1},
					},
					req: request{address: 0, length: 21},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize - 2, length: 46},
				},
			},
		},
		{
			name: "large gaps filled with offset",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize - 2, length: 46},
				},
			},
		},
		{
			name: "from PR #11106",
			inputs: []rangeDefinition{
				{0, 2, 1, 1, "INT16", true},
				{2, 1, 1, 1, "INT16", false},
				{3, 2*maxsize + 1, 1, 1, "INT16", true},
				{2*maxsize + 1, 1, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 2, count: 1, length: 1}},
					req:    request{address: 2, length: 1},
				},
				{
					fields: []rangeDefinition{{start: 2*maxsize + 1, count: 1, length: 1}},
					req:    request{address: 2*maxsize + 1, length: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the input structure and the expectation
			requestFields := generateRequestDefinitions(tt.inputs)
			expected := generateExpectation(tt.expected)

			// Setup the plugin
			slaveID := byte(1)
			plugin := Modbus{
				Name:              "Test",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			plugin.Requests = []requestDefinition{
				{
					SlaveID:      slaveID,
					ByteOrder:    "ABCD",
					RegisterType: "holding",
					Optimization: "aggressive",
					Fields:       requestFields,
				},
			}
			require.NoError(t, plugin.Init())
			require.NotEmpty(t, plugin.requests)
			require.Contains(t, plugin.requests, slaveID)
			requireEqualRequests(t, expected, plugin.requests[slaveID].holding)
		})
	}
}

func TestRequestOptimizationMaxInsertSmall(t *testing.T) {
	maxsize := maxQuantityHoldingRegisters
	maxExtraRegisters := uint16(5)
	tests := []struct {
		name     string
		inputs   []rangeDefinition
		expected []requestExpectation
	}{
		{
			name: "large gaps",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
					},
					req: request{address: maxsize - 2, length: 5},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize + 42, length: 2},
				},
			},
		},
		{
			name: "large gaps filled",
			inputs: []rangeDefinition{
				{0, 1, 1, 1, "INT16", false},
				{1, 17, 1, 1, "INT16", true},
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{
						{start: 0, count: 1, length: 1},
					},
					req: request{address: 0, length: 1},
				},
				{
					fields: []rangeDefinition{
						{start: 18, count: 3, length: 1},
					},
					req: request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
					},
					req: request{address: maxsize - 2, length: 5},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize + 42, length: 2},
				},
			},
		},
		{
			name: "large gaps filled with offset",
			inputs: []rangeDefinition{
				{18, 3, 1, 1, "INT16", false},
				{21, maxsize - 23, 1, 1, "INT16", true},
				{maxsize - 2, 5, 1, 1, "INT16", false},
				{maxsize + 3, 39, 1, 1, "INT16", true},
				{maxsize + 42, 2, 1, 1, "INT16", false},
			},
			expected: []requestExpectation{
				{
					fields: []rangeDefinition{{start: 18, count: 3, length: 1}},
					req:    request{address: 18, length: 3},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize - 2, count: 5, length: 1},
					},
					req: request{address: maxsize - 2, length: 5},
				},
				{
					fields: []rangeDefinition{
						{start: maxsize + 42, count: 2, length: 1},
					},
					req: request{address: maxsize + 42, length: 2},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate the input structure and the expectation
			requestFields := generateRequestDefinitions(tt.inputs)
			expected := generateExpectation(tt.expected)

			// Setup the plugin
			slaveID := byte(1)
			plugin := Modbus{
				Name:              "Test",
				Controller:        "tcp://localhost:1502",
				ConfigurationType: "request",
				Log:               testutil.Logger{},
			}
			plugin.Requests = []requestDefinition{
				{
					SlaveID:           slaveID,
					ByteOrder:         "ABCD",
					RegisterType:      "holding",
					Optimization:      "max_insert",
					MaxExtraRegisters: maxExtraRegisters,
					Fields:            requestFields,
				},
			}
			require.NoError(t, plugin.Init())
			require.NotEmpty(t, plugin.requests)
			require.Contains(t, plugin.requests, slaveID)
			requireEqualRequests(t, expected, plugin.requests[slaveID].holding)
		})
	}
}
func TestRequestsWorkaroundsOneRequestPerField(t *testing.T) {
	plugin := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "request",
		Log:               testutil.Logger{},
		Workarounds:       ModbusWorkarounds{OnRequestPerField: true},
	}
	plugin.Requests = []requestDefinition{
		{
			SlaveID:      1,
			ByteOrder:    "ABCD",
			RegisterType: "holding",
			Fields: []requestFieldDefinition{
				{
					Name:      "holding-1",
					Address:   uint16(1),
					InputType: "INT16",
				},
				{
					Name:      "holding-2",
					Address:   uint16(2),
					InputType: "INT16",
				},
				{
					Name:      "holding-3",
					Address:   uint16(3),
					InputType: "INT16",
				},
				{
					Name:      "holding-4",
					Address:   uint16(4),
					InputType: "INT16",
				},
				{
					Name:      "holding-5",
					Address:   uint16(5),
					InputType: "INT16",
				},
			},
		},
	}
	require.NoError(t, plugin.Init())
	require.Len(t, plugin.requests[1].holding, len(plugin.Requests[0].Fields))
}

func TestRegisterWorkaroundsOneRequestPerField(t *testing.T) {
	plugin := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "register",
		Log:               testutil.Logger{},
		Workarounds:       ModbusWorkarounds{OnRequestPerField: true},
	}
	plugin.SlaveID = 1
	plugin.HoldingRegisters = []fieldDefinition{
		{
			ByteOrder: "AB",
			DataType:  "INT16",
			Name:      "holding-1",
			Address:   []uint16{1},
			Scale:     1.0,
		},
		{
			ByteOrder: "AB",
			DataType:  "INT16",
			Name:      "holding-2",
			Address:   []uint16{2},
			Scale:     1.0,
		},
		{
			ByteOrder: "AB",
			DataType:  "INT16",
			Name:      "holding-3",
			Address:   []uint16{3},
			Scale:     1.0,
		},
		{
			ByteOrder: "AB",
			DataType:  "INT16",
			Name:      "holding-4",
			Address:   []uint16{4},
			Scale:     1.0,
		},
		{
			ByteOrder: "AB",
			DataType:  "INT16",
			Name:      "holding-5",
			Address:   []uint16{5},
			Scale:     1.0,
		},
	}
	require.NoError(t, plugin.Init())
	require.Len(t, plugin.requests[1].holding, len(plugin.HoldingRegisters))
}

func TestRequestsWorkaroundsReadCoilsStartingAtZeroRequest(t *testing.T) {
	plugin := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "request",
		Log:               testutil.Logger{},
		Workarounds:       ModbusWorkarounds{ReadCoilsStartingAtZero: true},
	}
	plugin.SlaveID = 1
	plugin.Requests = []requestDefinition{
		{
			SlaveID:      1,
			RegisterType: "coil",
			Fields: []requestFieldDefinition{
				{
					Name:    "coil-8",
					Address: uint16(8),
				},
				{
					Name:    "coil-new-group",
					Address: maxQuantityCoils,
				},
			},
		},
	}
	require.NoError(t, plugin.Init())
	require.Len(t, plugin.requests[1].coil, 2)

	// First group should now start at zero and have the cumulated length
	require.Equal(t, uint16(0), plugin.requests[1].coil[0].address)
	require.Equal(t, uint16(9), plugin.requests[1].coil[0].length)

	// The second field should form a new group as the previous request
	// is now too large (beyond max-coils-per-read) after zero enforcement.
	require.Equal(t, maxQuantityCoils, plugin.requests[1].coil[1].address)
	require.Equal(t, uint16(1), plugin.requests[1].coil[1].length)
}

func TestRequestsWorkaroundsReadCoilsStartingAtZeroRegister(t *testing.T) {
	plugin := Modbus{
		Name:              "Test",
		Controller:        "tcp://localhost:1502",
		ConfigurationType: "register",
		Log:               testutil.Logger{},
		Workarounds:       ModbusWorkarounds{ReadCoilsStartingAtZero: true},
	}
	plugin.SlaveID = 1
	plugin.Coils = []fieldDefinition{
		{
			Name:    "coil-8",
			Address: []uint16{8},
		},
		{
			Name:    "coil-new-group",
			Address: []uint16{maxQuantityCoils},
		},
	}
	require.NoError(t, plugin.Init())
	require.Len(t, plugin.requests[1].coil, 2)

	// First group should now start at zero and have the cumulated length
	require.Equal(t, uint16(0), plugin.requests[1].coil[0].address)
	require.Equal(t, uint16(9), plugin.requests[1].coil[0].length)

	// The second field should form a new group as the previous request
	// is now too large (beyond max-coils-per-read) after zero enforcement.
	require.Equal(t, maxQuantityCoils, plugin.requests[1].coil[1].address)
	require.Equal(t, uint16(1), plugin.requests[1].coil[1].length)
}
