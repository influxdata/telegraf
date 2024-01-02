package upsd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
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
					"battery_voltage":         float64(13.4),
					"firmware":                "CUSTOM_FIRMWARE",
					"input_voltage":           float64(242),
					"load_percent":            float64(23),
					"nominal_battery_voltage": float64(24),
					"nominal_input_voltage":   float64(230),
					"nominal_power":           int64(700),
					"output_voltage":          float64(230),
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
					"battery_voltage":         float64(13.4),
					"firmware":                "CUSTOM_FIRMWARE",
					"input_voltage":           float64(242),
					"load_percent":            float64(23),
					"nominal_battery_voltage": float64(24),
					"nominal_input_voltage":   float64(230),
					"nominal_power":           int64(700),
					"output_voltage":          float64(230),
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

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("upsd", func() telegraf.Input {
		return &Upsd{}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")

		t.Run(f.Name(), func(t *testing.T) {
			// Prepare the influx parser for expectations
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Setup a server from the input data
			server, err := setupServer(testcasePath)
			require.NoError(t, err)

			// Start the server
			ctx, cancel := context.WithCancel(context.Background())
			addr, err := server.listen(ctx)
			require.NoError(t, err)
			defer cancel()

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Upsd)
			plugin.Server = (addr.(*net.TCPAddr)).IP.String()
			plugin.Port = (addr.(*net.TCPAddr)).Port
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.PrintMetrics(actual)
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}

type interaction struct {
	Expected string
	Response string
}

type variable struct {
	Name  string
	Value string
}

type mockServer struct {
	protocol []interaction
}

func (s *mockServer) init() {
	s.protocol = []interaction{
		{
			Expected: "VER\n",
			Response: "1\n",
		},
		{
			Expected: "NETVER\n",
			Response: "1\n",
		},
		{
			Expected: "LIST UPS\n",
			Response: "BEGIN LIST UPS\nUPS fake \"fake UPS\"\nEND LIST UPS\n",
		},
		{
			Expected: "LIST CLIENT fake\n",
			Response: "BEGIN LIST CLIENT fake\nCLIENT fake 127.0.0.1\nEND LIST CLIENT fake\n",
		},
		{
			Expected: "LIST CMD fake\n",
			Response: "BEGIN LIST CMD fake\nEND LIST CMD fake\n",
		},
		{
			Expected: "GET UPSDESC fake\n",
			Response: "UPSDESC fake \"stub-ups-description\"\n",
		},
		{
			Expected: "GET NUMLOGINS fake\n",
			Response: "NUMLOGINS fake 1\n",
		},
	}
}

func (s *mockServer) addVariables(variables []variable, types map[string]string) error {
	// Add a VAR entries for the variables
	values := make([]string, 0, len(variables))
	for _, v := range variables {
		values = append(values, fmt.Sprintf("VAR fake %s %q", v.Name, v.Value))
	}

	s.protocol = append(s.protocol, interaction{
		Expected: "LIST VAR fake\n",
		Response: "BEGIN LIST VAR fake\n" + strings.Join(values, "\n") + "\nEND LIST VAR fake\n",
	})

	// Add a description and type interaction for the variable
	for _, v := range variables {
		variableType, found := types[v.Name]
		if !found {
			return fmt.Errorf("type for variable %q not found", v.Name)
		}

		s.protocol = append(s.protocol, interaction{
			Expected: "GET DESC fake " + v.Name + "\n",
			Response: "DESC fake" + v.Name + " \"No description here\"\n",
		})
		s.protocol = append(s.protocol, interaction{
			Expected: "GET TYPE fake " + v.Name + "\n",
			Response: "TYPE fake " + v.Name + " " + variableType + "\n",
		})
	}

	return nil
}

func (s *mockServer) listen(ctx context.Context) (net.Addr, error) {
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
				_ = conn.SetReadDeadline(time.Now().Add(time.Minute))

				in := make([]byte, 128)
				for _, interaction := range s.protocol {
					n, err := conn.Read(in)
					if err != nil {
						fmt.Printf("Failed to read from connection: %v\n", err)
						return
					}

					request := in[:n]
					if !bytes.Equal([]byte(interaction.Expected), request) {
						fmt.Printf("Unexpected request %q, expected %q\n", string(request), interaction.Expected)
						return
					}

					if _, err := conn.Write([]byte(interaction.Response)); err != nil {
						fmt.Printf("Cannot write answer for request %q: %v\n", string(request), err)
						return
					}
				}

				// Append EOF to end of output bytes
				if _, err := conn.Write([]byte{0, 0}); err != nil {
					fmt.Printf("Cannot write EOF: %v\n", err)
					return
				}
			}()
		}
	}()

	return ln.Addr(), nil
}

func setupServer(path string) (*mockServer, error) {
	// Read the variables
	varbuf, err := os.ReadFile(filepath.Join(path, "variables.dev"))
	if err != nil {
		return nil, fmt.Errorf("reading variables failed: %w", err)
	}

	// Parse the information into variable names and values (upsc format)
	variables := make([]variable, 0)
	scanner := bufio.NewScanner(bytes.NewBuffer(varbuf))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("cannot parse line %s", line)
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		variables = append(variables, variable{name, value})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("processing variables failed: %w", err)
	}

	// Read the variable-type mapping
	typebuf, err := os.ReadFile(filepath.Join(path, "types.dev"))
	if err != nil {
		return nil, fmt.Errorf("reading variables failed: %w", err)
	}

	// Parse the information into variable names and values (upsc format)
	types := make(map[string]string, 0)
	scanner = bufio.NewScanner(bytes.NewBuffer(typebuf))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("cannot parse line %s", line)
		}
		name := strings.TrimSpace(parts[0])
		vartype := strings.TrimSpace(parts[1])
		types[name] = vartype
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("processing variables failed: %w", err)
	}

	// Setup the server and add the device information
	server := &mockServer{}
	server.init()
	err = server.addVariables(variables, types)
	return server, err
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
VAR fake battery.runtime "600.00"
VAR fake output.voltage "230.0"
VAR fake battery.voltage "13.4"
VAR fake input.voltage.nominal "230.0"
VAR fake battery.voltage.nominal "24.0"
VAR fake ups.realpower "41.0"
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
	m = appendVariable(m, "ups.realpower", "NUMBER")
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
