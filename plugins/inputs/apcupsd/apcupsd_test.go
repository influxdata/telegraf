package apcupsd

import (
	"context"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestApcupsdDocs(_ *testing.T) {
	apc := &ApcUpsd{}
	apc.SampleConfig()
}

func TestApcupsdInit(t *testing.T) {
	input, ok := inputs.Inputs["apcupsd"]
	if !ok {
		t.Fatal("Input not defined")
	}

	_ = input().(*ApcUpsd)
}

func listen(ctx context.Context, t *testing.T, out [][]byte) (string, error) {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp4", "127.0.0.1:0")
	if err != nil {
		return "", err
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
				require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Second)))

				in := make([]byte, 128)
				n, err := conn.Read(in)
				require.NoError(t, err, "failed to read from connection")

				status := []byte{0, 6, 's', 't', 'a', 't', 'u', 's'}
				want, got := status, in[:n]
				require.Equal(t, want, got)

				// Run against test function and append EOF to end of output bytes
				out = append(out, []byte{0, 0})

				for _, o := range out {
					_, err := conn.Write(o)
					require.NoError(t, err, "failed to write to connection")
				}
			}()
		}
	}()

	return ln.Addr().String(), nil
}

func TestConfig(t *testing.T) {
	apc := &ApcUpsd{Timeout: defaultTimeout}

	var (
		tests = []struct {
			name    string
			servers []string
			err     bool
		}{
			{
				name:    "test listen address no scheme",
				servers: []string{"127.0.0.1:1234"},
				err:     true,
			},
			{
				name:    "test no port",
				servers: []string{"127.0.0.3"},
				err:     true,
			},
		}

		acc testutil.Accumulator
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apc.Servers = tt.servers

			err := apc.Gather(&acc)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApcupsdGather(t *testing.T) {
	apc := &ApcUpsd{Timeout: defaultTimeout}

	var (
		tests = []struct {
			name   string
			err    bool
			tags   map[string]string
			fields map[string]interface{}
			out    func() [][]byte
		}{
			{
				name: "test listening server with output",
				err:  false,
				tags: map[string]string{
					"serial":   "ABC123",
					"status":   "ONLINE",
					"ups_name": "BERTHA",
					"model":    "Model 12345",
				},
				fields: map[string]interface{}{
					"status_flags":            uint64(8),
					"battery_charge_percent":  float64(0),
					"battery_voltage":         float64(0),
					"input_frequency":         float64(0),
					"input_voltage":           float64(0),
					"internal_temp":           float64(0),
					"load_percent":            float64(13),
					"output_voltage":          float64(0),
					"time_left_ns":            int64(2790000000000),
					"time_on_battery_ns":      int64(0),
					"nominal_input_voltage":   float64(230),
					"nominal_battery_voltage": float64(12),
					"nominal_power":           865,
					"firmware":                "857.L3 .I USB FW:L3",
					"battery_date":            "2016-09-06",
				},
				out: genOutput,
			},
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
			if err != nil {
				t.Fatal(err)
			}

			apc.Servers = []string{"tcp://" + lAddr}

			err = apc.Gather(&acc)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				acc.AssertContainsTaggedFields(t, "apcupsd", tt.fields, tt.tags)
			}
			cancel()
		})
	}
}

// The following functionality is straight from apcupsd tests.

// kvBytes is a helper to generate length and key/value byte buffers.
func kvBytes(kv string) ([]byte, []byte) {
	lenb := make([]byte, 2)
	binary.BigEndian.PutUint16(lenb, uint16(len(kv)))

	return lenb, []byte(kv)
}

func genOutput() [][]byte {
	kvs := []string{
		"SERIALNO : ABC123",
		"STATUS   : ONLINE",
		"STATFLAG : 0x08 Status Flag",
		"UPSNAME  : BERTHA",
		"MODEL    : Model 12345",
		"DATE     : 2016-09-06 22:13:28 -0400",
		"HOSTNAME : example",
		"LOADPCT  :  13.0 Percent Load Capacity",
		"BATTDATE : 2016-09-06",
		"TIMELEFT :  46.5 Minutes",
		"TONBATT  : 0 seconds",
		"NUMXFERS : 0",
		"SELFTEST : NO",
		"NOMINV   : 230 Volts",
		"NOMBATTV : 12.0 Volts",
		"NOMPOWER : 865 Watts",
		"FIRMWARE : 857.L3 .I USB FW:L3",
		"ALARMDEL : Low Battery",
	}

	var out [][]byte
	for _, kv := range kvs {
		lenb, kvb := kvBytes(kv)
		out = append(out, lenb)
		out = append(out, kvb)
	}

	return out
}

func genBadOutput() [][]byte {
	kvs := []string{
		"STATFLAG : 0x08Status Flag",
	}

	var out [][]byte
	for _, kv := range kvs {
		lenb, kvb := kvBytes(kv)
		out = append(out, lenb)
		out = append(out, kvb)
	}

	return out
}
