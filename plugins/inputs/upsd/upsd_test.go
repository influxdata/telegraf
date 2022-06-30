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
			name   string
			err    bool
			tags   map[string]string
			fields map[string]interface{}
			out    func() []interaction
		}{
			{
				name: "test listening server with output",
				err:  false,
				tags: map[string]string{
					"serial":    "ABC123",
					"ups_name":  "fake",
					"model":     "Model 12345",
					"status_OL": "true",
				},
				fields: map[string]interface{}{
					"status_flags":            uint64(8),
					"ups.status":              "OL",
					"battery_charge_percent":  float64(100),
					"battery_voltage":         float64(13.4),
					"input_frequency":         nil,
					"input_voltage":           float64(242),
					"internal_temp":           nil,
					"load_percent":            float64(23),
					"output_voltage":          float64(230),
					"time_left_ns":            int64(600000000000),
					"nominal_input_voltage":   float64(230),
					"nominal_battery_voltage": float64(24),
					"nominal_power":           int64(700),
					"firmware":                "CUSTOM_FIRMWARE",
					"battery_date":            "2016-07-26",
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
VAR fake battery.runtime "600"
VAR fake output.voltage "230.0"
VAR fake battery.voltage "13.4"
VAR fake input.voltage.nominal "230.0"
VAR fake battery.voltage.nominal "24.0"
VAR fake ups.realpower.nominal "700"
VAR fake ups.firmware "CUSTOM_FIRMWARE"
VAR fake battery.mfr.date "2016-07-26"
VAR fake ups.status "OL"
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
	m = appendVariable(m, "ups.realpower.nominal", "NUMBER")
	m = appendVariable(m, "ups.firmware", "STRING:64")
	m = appendVariable(m, "battery.mfr.date", "STRING:64")
	m = appendVariable(m, "ups.status", "STRING:64")

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
