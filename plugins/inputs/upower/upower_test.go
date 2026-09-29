//go:build !windows && !darwin

package upower

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type fakeDevice struct {
	path  dbus.ObjectPath
	props map[string]any
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
		if d.path != path {
			continue
		}
		if d.err != nil {
			return nil, d.err
		}
		props := make(map[string]dbus.Variant, len(d.props))
		for k, v := range d.props {
			props[k] = dbus.MakeVariant(v)
		}
		return props, nil
	}
	return nil, errors.New("no such device")
}

var (
	battery = fakeDevice{
		path: "/org/freedesktop/UPower/devices/battery_BAT0",
		props: map[string]any{
			"NativePath":       "BAT0",
			"Vendor":           "LGC",
			"Model":            "5B10W13973",
			"Serial":           "1234",
			"Type":             uint32(2),
			"State":            uint32(2),
			"Technology":       uint32(2),
			"WarningLevel":     uint32(1),
			"Percentage":       float64(83.0),
			"Energy":           float64(41.57),
			"EnergyEmpty":      float64(0.0),
			"EnergyFull":       float64(50.08),
			"EnergyFullDesign": float64(57.0),
			"EnergyRate":       float64(7.891),
			"Voltage":          float64(11.98),
			"Temperature":      float64(0.0),
			"Capacity":         float64(87.86),
			"ChargeCycles":     int32(213),
			"TimeToEmpty":      int64(18966),
			"TimeToFull":       int64(0),
			"IsPresent":        true,
			"IsRechargeable":   true,
			"PowerSupply":      true,
		},
	}
	linePower = fakeDevice{
		path: "/org/freedesktop/UPower/devices/line_power_AC",
		props: map[string]any{
			"NativePath": "AC",
			"Type":       uint32(1),
			"Online":     false,
		},
	}
	mouse = fakeDevice{
		path: "/org/freedesktop/UPower/devices/mouse_hidpp_battery_0",
		props: map[string]any{
			"NativePath":     "hidpp_battery_0",
			"Vendor":         "Logitech",
			"Model":          "MX Master 3",
			"Type":           uint32(5),
			"State":          uint32(2),
			"WarningLevel":   uint32(1),
			"Percentage":     float64(55.0),
			"IsPresent":      true,
			"IsRechargeable": true,
		},
	}

	batteryMetric = metric.New(
		"upower",
		map[string]string{
			"native_path":   "BAT0",
			"type":          "battery",
			"vendor":        "LGC",
			"model":         "5B10W13973",
			"serial":        "1234",
			"state":         "discharging",
			"technology":    "lithium_polymer",
			"warning_level": "none",
		},
		map[string]any{
			"percentage":            83.0,
			"energy_wh":             41.57,
			"energy_empty_wh":       0.0,
			"energy_full_wh":        50.08,
			"energy_full_design_wh": 57.0,
			"energy_rate_w":         7.891,
			"voltage_v":             11.98,
			"temperature_c":         0.0,
			"capacity_percent":      87.86,
			"charge_cycles":         int32(213),
			"time_to_empty_sec":     int64(18966),
			"time_to_full_sec":      int64(0),
			"is_present":            true,
			"is_rechargeable":       true,
			"power_supply":          true,
		},
		time.Unix(0, 0),
		telegraf.Gauge,
	)
	linePowerMetric = metric.New(
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
	mouseMetric = metric.New(
		"upower",
		map[string]string{
			"native_path":   "hidpp_battery_0",
			"type":          "mouse",
			"vendor":        "Logitech",
			"model":         "MX Master 3",
			"state":         "discharging",
			"technology":    "unknown",
			"warning_level": "none",
		},
		map[string]any{
			"percentage":            55.0,
			"energy_wh":             0.0,
			"energy_empty_wh":       0.0,
			"energy_full_wh":        0.0,
			"energy_full_design_wh": 0.0,
			"energy_rate_w":         0.0,
			"voltage_v":             0.0,
			"temperature_c":         0.0,
			"capacity_percent":      0.0,
			"charge_cycles":         int32(0),
			"time_to_empty_sec":     int64(0),
			"time_to_full_sec":      int64(0),
			"is_present":            true,
			"is_rechargeable":       true,
			"power_supply":          false,
		},
		time.Unix(0, 0),
		telegraf.Gauge,
	)
)

func TestInitInvalidDeviceType(t *testing.T) {
	plugin := &UPower{DeviceTypes: []string{"battery", "toaster"}}
	require.ErrorContains(t, plugin.Init(), `invalid device type "toaster"`)
}

func TestInitDefaults(t *testing.T) {
	plugin := &UPower{}
	require.NoError(t, plugin.Init())
	require.Equal(t, config.Duration(5*time.Second), plugin.Timeout)
}

func TestGather(t *testing.T) {
	tests := []struct {
		name        string
		deviceTypes []string
		client      *fakeClient
		expected    []telegraf.Metric
		expectedErr string
	}{
		{
			name:     "all devices",
			client:   &fakeClient{devices: []fakeDevice{battery, linePower, mouse}},
			expected: []telegraf.Metric{batteryMetric, linePowerMetric, mouseMetric},
		},
		{
			name:        "filtered by type",
			deviceTypes: []string{"battery", "ups"},
			client:      &fakeClient{devices: []fakeDevice{battery, linePower, mouse}},
			expected:    []telegraf.Metric{batteryMetric},
		},
		{
			name:     "no devices",
			client:   &fakeClient{},
			expected: make([]telegraf.Metric, 0),
		},
		{
			name: "unknown enumeration values",
			client: &fakeClient{devices: []fakeDevice{{
				path:  "/org/freedesktop/UPower/devices/battery_X",
				props: map[string]any{"NativePath": "X", "Type": uint32(99), "State": uint32(42)},
			}}},
			expected: []telegraf.Metric{
				metric.New(
					"upower",
					map[string]string{
						"native_path":   "X",
						"type":          "unknown-99",
						"state":         "unknown-42",
						"technology":    "unknown",
						"warning_level": "unknown",
					},
					map[string]any{
						"percentage":            0.0,
						"energy_wh":             0.0,
						"energy_empty_wh":       0.0,
						"energy_full_wh":        0.0,
						"energy_full_design_wh": 0.0,
						"energy_rate_w":         0.0,
						"voltage_v":             0.0,
						"temperature_c":         0.0,
						"capacity_percent":      0.0,
						"charge_cycles":         int32(0),
						"time_to_empty_sec":     int64(0),
						"time_to_full_sec":      int64(0),
						"is_present":            false,
						"is_rechargeable":       false,
						"power_supply":          false,
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
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
			plugin := &UPower{
				DeviceTypes: tt.deviceTypes,
				Log:         testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			plugin.client = tt.client
			defer plugin.Stop()

			var acc testutil.Accumulator
			err := acc.GatherError(plugin.Gather)
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}
