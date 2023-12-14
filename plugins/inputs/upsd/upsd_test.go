package upsd

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestUpsdGather(t *testing.T) {
	nut := &Upsd{}

	var (
		tests = []struct {
			name       string
			forceFloat bool
			err        bool
			tags       map[string]string
			fields     map[string]interface{}
			out        func() []interaction
		}{
			{
				name:       "test listening server with output",
				forceFloat: false,
				err:        false,
				tags: map[string]string{
					"serial":    "ABC123",
					"ups_name":  "fake",
					"model":     "Model 12345",
					"status_OL": "true",
				},
				fields: map[string]interface{}{
					"battery_charge_percent":  float64(100),
					"battery_date":            nil,
					"battery_mfr_date":        "2016-07-26",
					"battery_protection":      "yes",
					"battery_voltage":         float64(13.4),
					"battery_capacity":        float64(7.00),
					"battery_runtime":         int64(3873),
					"device_mfr":              "Eaton",
					"device_model":            "Eaton 2000",
					"driver_version":          "2.8.1",
					"driver_version_data":     "MGE HID 1.46",
					"driver_version_internal": float64(0.52),
					"driver_version_usb":      "libusb-1.0.26 (API: 0x1000108)",
					"device_type":             "ups",
					"firmware":                "CUSTOM_FIRMWARE",
					"ups_mfr":                 "Eaton",
					"ups_model":               "Eaton 2000",
					"ups_productid":           "ffff",
					"ups_test_result":         "Done and passed",
					"ups_type":                "online",
					"ups_vendorid":            int64(0463),
					"ups_test_interval":       int64(604800),
					"ups_beeper_status":       "enabled",
					"outlet_switchable":       "no",
					"input_frequency":         float64(49.9),
					"input_transfer_high":     int64(300),
					"input_transfer_low":      int64(100),
					"input_bypass_frequency":  float64(49.9),
					"input_bypass_voltage":    float64(234.0),
					"input_frequency_nominal": int64(50),
					"internal_temp":           float64(24.9),
					"ups_shutdown":            "enabled",
					"input_voltage":           float64(242),
					"load_percent":            float64(23),
					"nominal_battery_voltage": float64(24),
					"nominal_input_voltage":   float64(230),
					"nominal_power":           int64(700),
					"output_voltage":          float64(230),
					"output_current":          float64(1.60),
					"real_power":              float64(41),
					"status_flags":            uint64(8),
					"time_left_ns":            int64(600000000000),
					"ups_status":              "OL",
				},
				out: genOutput,
			},
			{
				name:       "test listening server with output & force floats",
				forceFloat: true,
				err:        false,
				tags: map[string]string{
					"serial":    "ABC123",
					"ups_name":  "fake",
					"model":     "Model 12345",
					"status_OL": "true",
				},
				fields: map[string]interface{}{
					"battery_charge_percent":  float64(100),
					"battery_date":            nil,
					"battery_mfr_date":        "2016-07-26",
					"battery_protection":      "yes",
					"battery_voltage":         float64(13.4),
					"battery_capacity":        float64(7.00),
					"battery_runtime":         int64(3873),
					"device_mfr":              "Eaton",
					"device_model":            "Eaton 2000",
					"driver_version":          "2.8.1",
					"driver_version_data":     "MGE HID 1.46",
					"driver_version_internal": float64(0.52),
					"driver_version_usb":      "libusb-1.0.26 (API: 0x1000108)",
					"device_type":             "ups",
					"firmware":                "CUSTOM_FIRMWARE",
					"ups_mfr":                 "Eaton",
					"ups_model":               "Eaton 2000",
					"ups_productid":           "ffff",
					"ups_test_result":         "Done and passed",
					"ups_type":                "online",
					"ups_vendorid":            "0463",
					"ups_test_interval":       int64(604800),
					"ups_beeper_status":       "enabled",
					"outlet_switchable":       "no",
					"input_frequency":         float64(49.9),
					"input_transfer_high":     int64(300),
					"input_transfer_low":      int64(100),
					"input_bypass_frequency":  float64(49.9),
					"input_bypass_voltage":    float64(234.0),
					"input_frequency_nominal": int64(50),
					"internal_temp":           float64(24.9),
					"ups_shutdown":            "enabled",
					"input_voltage":           float64(242),
					"load_percent":            float64(23),
					"nominal_battery_voltage": float64(24),
					"nominal_input_voltage":   float64(230),
					"nominal_power":           int64(700),
					"output_voltage":          float64(230),
					"output_current":          float64(1.60),
					"real_power":              float64(41),
					"status_flags":            uint64(8),
					"time_left_ns":            int64(600000000000),
					"ups_status":              "OL",
				},
				out: genOutput,
			},
		}

		acc testutil.Accumulator
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			lAddr, err := listen(ctx, t, tt.out())
			require.NoError(t, err)

			nut.Server = (lAddr.(*net.TCPAddr)).IP.String()
			nut.Port = (lAddr.(*net.TCPAddr)).Port
			nut.ForceFloat = tt.forceFloat

			err = nut.Gather(&acc)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				acc.AssertContainsFields(t, "upsd", tt.fields)
				acc.AssertContainsTaggedFields(t, "upsd", tt.fields, tt.tags)
			}
			cancel()
		})
	}
}

func TestUpsdGatherFail(t *testing.T) {
	nut := &Upsd{}

	var (
		tests = []struct {
			name   string
			err    bool
			tags   map[string]string
			fields map[string]interface{}
			out    func() []interaction
		}{
			{
				name: "test with bad output",
				err:  true,
				out:  genBadOutput,
			},
		}

		acc testutil.Accumulator
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			lAddr, err := listen(ctx, t, tt.out())
			require.NoError(t, err)

			nut.Server = (lAddr.(*net.TCPAddr)).IP.String()
			nut.Port = (lAddr.(*net.TCPAddr)).Port

			err = nut.Gather(&acc)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				acc.AssertContainsTaggedFields(t, "upsd", tt.fields, tt.tags)
			}
			cancel()
		})
	}
}

func listen(ctx context.Context, t *testing.T, out []interaction) (net.Addr, error) {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	go func() {
		defer ln.Close()

		for ctx.Err() == nil {
			func() {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Minute)))

				in := make([]byte, 128)
				for _, interaction := range out {
					n, err := conn.Read(in)
					require.NoError(t, err, "failed to read from connection")

					expectedBytes := []byte(interaction.Expected)
					want, got := expectedBytes, in[:n]
					require.Equal(t, want, got)

					_, err = conn.Write([]byte(interaction.Response))
					require.NoError(t, err, "failed to respond to LIST UPS")
				}

				// Append EOF to end of output bytes
				_, err = conn.Write([]byte{0, 0})
				require.NoError(t, err, "failed to write EOF")
			}()
		}
	}()

	return ln.Addr(), nil
}

type interaction struct {
	Expected string
	Response string
}

func genOutput() []interaction {
	m := make([]interaction, 0)
	m = append(m, interaction{
		Expected: "VER\n",
		Response: "1\n",
	})
	m = append(m, interaction{
		Expected: "NETVER\n",
		Response: "1\n",
	})
	m = append(m, interaction{
		Expected: "LIST UPS\n",
		Response: `BEGIN LIST UPS
UPS fake "fakescription"
END LIST UPS
`,
	})
	m = append(m, interaction{
		Expected: "LIST CLIENT fake\n",
		Response: `BEGIN LIST CLIENT fake
CLIENT fake 192.168.1.1
END LIST CLIENT fake
`,
	})
	m = append(m, interaction{
		Expected: "LIST CMD fake\n",
		Response: `BEGIN LIST CMD fake
END LIST CMD fake
`,
	})
	m = append(m, interaction{
		Expected: "GET UPSDESC fake\n",
		Response: "UPSDESC fake \"stub-ups-description\"\n",
	})
	m = append(m, interaction{
		Expected: "GET NUMLOGINS fake\n",
		Response: "NUMLOGINS fake 1\n",
	})
	m = append(m, interaction{
		Expected: "LIST VAR fake\n",
		Response: `BEGIN LIST VAR fake
VAR fake device.serial "ABC123"
VAR fake device.model "Model 12345"
VAR fake input.voltage "242.0"
VAR fake ups.load "23.0"
VAR fake battery.charge "100.0"
VAR fake battery.runtime "3873"
VAR fake output.voltage "230.0"
VAR fake battery.voltage "13.4"
VAR fake input.voltage.nominal "230.0"
VAR fake battery.voltage.nominal "24.0"
VAR fake ups.realpower "41.0"
VAR fake ups.realpower.nominal "700"
VAR fake ups.firmware "CUSTOM_FIRMWARE"
VAR fake battery.mfr.date "2016-07-26"
VAR fake ups.status "OL"
VAR fake battery.capacity "7.00"
VAR fake device.mfr "Eaton"
VAR fake device.model "Eaton 2000"
VAR fake driver.version "2.8.1"
VAR fake driver.version.data "MGE HID 1.46"
VAR fake driver.version.internal "0.52"
VAR fake driver.version.usb "libusb-1.0.26 (API: 0x1000108)"
VAR fake device.type "ups"
VAR fake ups.firmware "1.0"
VAR fake ups.mfr "Eaton"
VAR fake ups.model "Eaton 2000"
VAR fake ups.productid "ffff"
VAR fake ups.test.result "Done and passed"
VAR fake ups.type "online"
VAR fake ups.vendorid "0463"
VAR fake ups.test.interval "604800"
VAR fake ups.beeper.status "enabled"
VAR fake ups.shutdown "enabled"
VAR fake outlet.switchable "no"
VAR fake input.bypass.frequency "49.9"
VAR fake input.bypass.voltage "234.0"
VAR fake input.frequency.nominal "50"
VAR fake output.current "1.60"
VAR fake output.frequency "49.9"
VAR fake output.frequency.nominal "50"
VAR fake output.voltage.nominal "230"
VAR fake ups.power "389"
VAR fake ups.power.nominal "2000"
VAR fake ups.temperature "24.9"
END LIST VAR fake
`,
	})
	m = appendVariable(m, "device.serial", "STRING:64")
	m = appendVariable(m, "device.model", "STRING:64")
	m = appendVariable(m, "input.voltage", "NUMBER")
	m = appendVariable(m, "ups.load", "NUMBER")
	m = appendVariable(m, "battery.charge", "NUMBER")
	m = appendVariable(m, "battery.runtime", "NUMBER")
	m = appendVariable(m, "output.voltage", "NUMBER")
	m = appendVariable(m, "battery.voltage", "NUMBER")
	m = appendVariable(m, "input.voltage.nominal", "NUMBER")
	m = appendVariable(m, "battery.voltage.nominal", "NUMBER")
	m = appendVariable(m, "ups.realpower", "NUMBER")
	m = appendVariable(m, "ups.realpower.nominal", "NUMBER")
	m = appendVariable(m, "ups.firmware", "STRING:64")
	m = appendVariable(m, "battery.mfr.date", "STRING:64")
	m = appendVariable(m, "ups.status", "STRING:64")
	m = appendVariable(m, "battery.capacity", "NUMBER")
	m = appendVariable(m, "device.mfr", "STRING:64")
	m = appendVariable(m, "device.model", "STRING:64")
	m = appendVariable(m, "driver.version", "STRING:64")
	m = appendVariable(m, "driver.version.data", "STRING:64")
	m = appendVariable(m, "driver.version.internal", "NUMBER")
	m = appendVariable(m, "driver.version.usb", "STRING:64")
	m = appendVariable(m, "device.type", "STRING:64")
	m = appendVariable(m, "ups.firmware", "STRING:64")
	m = appendVariable(m, "ups.mfr", "STRING:64")
	m = appendVariable(m, "ups.model", "STRING:64")
	m = appendVariable(m, "ups.productid", "STRING:64")
	m = appendVariable(m, "ups.test.result", "STRING:64")
	m = appendVariable(m, "ups.type", "STRING:64")
	m = appendVariable(m, "ups.vendorid", "NUMBER")
	m = appendVariable(m, "ups.test.interval", "NUMBER")
	m = appendVariable(m, "ups.beeper.status", "STRING:64")
	m = appendVariable(m, "ups.shutdown", "STRING:64")
	m = appendVariable(m, "outlet.switchable", "STRING:64")
	m = appendVariable(m, "input.bypass.frequency", "NUMBER")
	m = appendVariable(m, "input.bypass.voltage", "NUMBER")
	m = appendVariable(m, "input.frequency.nominal", "NUMBER")
	m = appendVariable(m, "output.current", "NUMBER")
	m = appendVariable(m, "output.frequency", "NUMBER")
	m = appendVariable(m, "output.frequency.nominal", "NUMBER")
	m = appendVariable(m, "output.voltage.nominal", "NUMBER")
	m = appendVariable(m, "ups.power", "NUMBER")
	m = appendVariable(m, "ups.power.nominal", "NUMBER")
	m = appendVariable(m, "ups.temperature", "NUMBER")

	return m
}

func appendVariable(m []interaction, name string, typ string) []interaction {
	m = append(m, interaction{
		Expected: "GET DESC fake " + name + "\n",
		Response: "DESC fake" + name + " \"No description here\"\n",
	})
	m = append(m, interaction{
		Expected: "GET TYPE fake " + name + "\n",
		Response: "TYPE fake " + name + " " + typ + "\n",
	})
	return m
}

func genBadOutput() []interaction {
	m := make([]interaction, 0)
	return m
}
