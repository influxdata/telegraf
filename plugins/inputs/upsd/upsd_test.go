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

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestBadServer(t *testing.T) {
	// Create and start a server without interactions
	server := &mockServer{}
	ctx, cancel := context.WithCancel(context.Background())
	addr, err := server.listen(ctx)
	require.NoError(t, err)
	defer cancel()

	// Setup the plugin
	plugin := &Upsd{
		Server: addr.IP.String(),
		Port:   addr.Port,
	}
	require.NoError(t, plugin.Init())

	// Do the query
	var acc testutil.Accumulator
	require.Error(t, plugin.Gather(&acc))
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
			plugin.Server = addr.IP.String()
			plugin.Port = addr.Port
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
			acc.Lock()
			defer acc.Unlock()
			require.Empty(t, acc.Errors)
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

		s.protocol = append(s.protocol,
			interaction{
				Expected: "GET DESC fake " + v.Name + "\n",
				Response: "DESC fake" + v.Name + " \"No description here\"\n",
			},
			interaction{
				Expected: "GET TYPE fake " + v.Name + "\n",
				Response: "TYPE fake " + v.Name + " " + variableType + "\n",
			},
		)
	}

	return nil
}

func (s *mockServer) listen(ctx context.Context) (*net.TCPAddr, error) {
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

	return ln.Addr().(*net.TCPAddr), nil
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
