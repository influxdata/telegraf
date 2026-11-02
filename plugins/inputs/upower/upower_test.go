//go:build !windows && !darwin

package upower

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitInvalidDeviceType(t *testing.T) {
	plugin := &UPower{DeviceTypes: []string{"battery", "toaster"}}
	require.ErrorContains(t, plugin.Init(), `invalid device type "toaster"`)
}

func TestInitInvalidTimeout(t *testing.T) {
	plugin := &UPower{Timeout: config.Duration(-1 * time.Second)}
	require.ErrorContains(t, plugin.Init(), "timeout must not be negative")
}

func TestCases(t *testing.T) {
	// Get all testcase directories
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			inputFilename := filepath.Join(testcasePath, "devices.json")
			expectedFilename := filepath.Join(testcasePath, "expected.out")

			// Load the input data
			devices, err := readDevices(inputFilename)
			require.NoError(t, err)

			// Read the expected output if any
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Configure and setup the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			plugin := cfg.Inputs[0].Input.(*UPower)
			plugin.Log = testutil.Logger{}
			require.NoError(t, plugin.Init())
			plugin.client = &fakeClient{devices: devices}
			defer plugin.Stop()

			// Gather and check the metrics
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreType(), testutil.SortMetrics())
		})
	}
}

func TestGatherErrors(t *testing.T) {
	linePower := fakeDevice{
		path: "/org/freedesktop/UPower/devices/line_power_AC",
		props: map[string]dbus.Variant{
			"NativePath": dbus.MakeVariant("AC"),
			"Type":       dbus.MakeVariant(uint32(1)),
			"Online":     dbus.MakeVariant(false),
		},
	}
	linePowerMetric := metric.New(
		"upower",
		map[string]string{
			"native_path": "AC",
			"type":        "line_power",
		},
		map[string]any{
			"online": false,
		},
		time.Unix(0, 0),
		telegraf.Gauge,
	)

	tests := []struct {
		name        string
		client      *fakeClient
		expected    []telegraf.Metric
		expectedErr string
	}{
		{
			name: "device error does not stop collection",
			client: &fakeClient{devices: []fakeDevice{
				{path: "/org/freedesktop/UPower/devices/broken", err: errors.New("boom")},
				linePower,
			}},
			expected:    []telegraf.Metric{linePowerMetric},
			expectedErr: `getting properties of device "/org/freedesktop/UPower/devices/broken" failed: boom`,
		},
		{
			name:        "enumeration error",
			client:      &fakeClient{enumerateErr: errors.New("boom")},
			expectedErr: "enumerating devices failed: boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &UPower{Log: testutil.Logger{}}
			require.NoError(t, plugin.Init())
			plugin.client = tt.client
			defer plugin.Stop()

			var acc testutil.Accumulator
			require.ErrorContains(t, acc.GatherError(plugin.Gather), tt.expectedErr)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

// Fake implementation of the D-Bus client
type fakeDevice struct {
	path  dbus.ObjectPath
	props map[string]dbus.Variant
	err   error
}

type fakeClient struct {
	devices      []fakeDevice
	enumerateErr error
}

func (*fakeClient) connected() bool { return true }
func (*fakeClient) close() error    { return nil }

func (c *fakeClient) enumerateDevices(context.Context) ([]dbus.ObjectPath, error) {
	if c.enumerateErr != nil {
		return nil, c.enumerateErr
	}
	paths := make([]dbus.ObjectPath, 0, len(c.devices))
	for _, d := range c.devices {
		paths = append(paths, d.path)
	}
	return paths, nil
}

func (c *fakeClient) deviceProperties(_ context.Context, path dbus.ObjectPath) (map[string]dbus.Variant, error) {
	for _, d := range c.devices {
		if d.path == path {
			return d.props, d.err
		}
	}
	return nil, errors.New("no such device")
}

func readDevices(filename string) ([]fakeDevice, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var entries []struct {
		Path  string `json:"path"`
		Props []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"props"`
	}
	if err := json.Unmarshal(buf, &entries); err != nil {
		return nil, err
	}

	devices := make([]fakeDevice, 0, len(entries))
	for _, e := range entries {
		props := make(map[string]dbus.Variant, len(e.Props))
		for _, p := range e.Props {
			v, err := parseValue(p.Type, p.Value)
			if err != nil {
				return nil, fmt.Errorf("parsing property %q of %q failed: %w", p.Name, e.Path, err)
			}
			props[p.Name] = dbus.MakeVariant(v)
		}
		devices = append(devices, fakeDevice{path: dbus.ObjectPath(e.Path), props: props})
	}
	return devices, nil
}

func parseValue(typ, value string) (any, error) {
	switch typ {
	case "string":
		return value, nil
	case "boolean":
		return strconv.ParseBool(value)
	case "double":
		return strconv.ParseFloat(value, 64)
	case "int32":
		v, err := strconv.ParseInt(value, 10, 32)
		return int32(v), err
	case "int64":
		return strconv.ParseInt(value, 10, 64)
	case "uint32":
		v, err := strconv.ParseUint(value, 10, 32)
		return uint32(v), err
	}
	return nil, fmt.Errorf("unknown type %q", typ)
}
