//go:build linux

package systemd_units

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	sdbus "github.com/coreos/go-systemd/v22/dbus"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type properties struct {
	uf         string
	utype      string
	state      sdbus.UnitStatus
	ufPreset   string
	ufState    string
	properties map[string]interface{}
}

func TestDefaultPattern(t *testing.T) {
	plugin := &SystemdUnits{}
	require.NoError(t, plugin.Init())
	require.Equal(t, "*", plugin.Pattern)
}

func TestListFiles(t *testing.T) {
	tests := []struct {
		name        string
		properties  map[string]properties
		expected    []telegraf.Metric
		expectedErr string
	}{
		{
			name: "example loaded active running",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "loaded",
						ActiveState: "active",
						SubState:    "running",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":   "example.service",
						"load":   "loaded",
						"active": "active",
						"sub":    "running",
					},
					map[string]interface{}{
						"load_code":   0,
						"active_code": 0,
						"sub_code":    0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example loaded active exited",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "loaded",
						ActiveState: "active",
						SubState:    "exited",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":   "example.service",
						"load":   "loaded",
						"active": "active",
						"sub":    "exited",
					},
					map[string]interface{}{
						"load_code":   0,
						"active_code": 0,
						"sub_code":    4,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example loaded failed failed",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "loaded",
						ActiveState: "failed",
						SubState:    "failed",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":   "example.service",
						"load":   "loaded",
						"active": "failed",
						"sub":    "failed",
					},
					map[string]interface{}{
						"load_code":   0,
						"active_code": 3,
						"sub_code":    12,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example not-found inactive dead",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "not-found",
						ActiveState: "inactive",
						SubState:    "dead",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":   "example.service",
						"load":   "not-found",
						"active": "inactive",
						"sub":    "dead",
					},
					map[string]interface{}{
						"load_code":   2,
						"active_code": 2,
						"sub_code":    1,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example unknown unknown unknown",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "unknown",
						ActiveState: "unknown",
						SubState:    "unknown",
					},
				},
			},
			expectedErr: "parsing field 'load' failed, value not in map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin. Do NOT call Start() as this would connect to
			// the real systemd daemon.
			plugin := &SystemdUnits{
				Pattern: "examp*",
				Timeout: config.Duration(time.Second),
			}
			require.NoError(t, plugin.Init())

			// Create a fake client to inject data
			plugin.client = &fakeClient{
				units:     tt.properties,
				connected: true,
			}
			defer plugin.Stop()

			// Run gather
			var acc testutil.Accumulator
			err := acc.GatherError(plugin.Gather)
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)

			// Do the comparison
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestShow(t *testing.T) {
	tests := []struct {
		name        string
		properties  map[string]properties
		expected    []telegraf.Metric
		expectedErr string
	}{
		{
			name: "example loaded active running",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "loaded",
						ActiveState: "active",
						SubState:    "running",
					},
					ufPreset: "disabled",
					ufState:  "enabled",
					properties: map[string]interface{}{
						"Id":                "example.service",
						"StatusErrno":       0,
						"NRestarts":         1,
						"MemoryCurrent":     1000,
						"MemoryPeak":        2000,
						"MemorySwapCurrent": 3000,
						"MemorySwapPeak":    4000,
						"MemoryAvailable":   5000,
						"MainPID":           9999,
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":      "example.service",
						"load":      "loaded",
						"active":    "active",
						"sub":       "running",
						"uf_state":  "enabled",
						"uf_preset": "disabled",
					},
					map[string]interface{}{
						"load_code":    0,
						"active_code":  0,
						"sub_code":     0,
						"status_errno": 0,
						"restarts":     1,
						"mem_current":  1000,
						"mem_peak":     2000,
						"swap_current": 3000,
						"swap_peak":    4000,
						"mem_avail":    5000,
						"pid":          9999,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example loaded active exited",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "loaded",
						ActiveState: "active",
						SubState:    "exited",
					},
					ufPreset: "disabled",
					ufState:  "enabled",
					properties: map[string]interface{}{
						"Id":          "example.service",
						"StatusErrno": 0,
						"NRestarts":   0,
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":      "example.service",
						"load":      "loaded",
						"active":    "active",
						"sub":       "exited",
						"uf_state":  "enabled",
						"uf_preset": "disabled",
					},
					map[string]interface{}{
						"load_code":    0,
						"active_code":  0,
						"sub_code":     4,
						"status_errno": 0,
						"restarts":     0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example loaded failed failed",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "loaded",
						ActiveState: "failed",
						SubState:    "failed",
					},
					ufPreset: "disabled",
					ufState:  "enabled",
					properties: map[string]interface{}{
						"Id":                "example.service",
						"StatusErrno":       10,
						"NRestarts":         1,
						"MemoryCurrent":     1000,
						"MemoryPeak":        2000,
						"MemorySwapCurrent": 3000,
						"MemorySwapPeak":    4000,
						"MemoryAvailable":   5000,
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":      "example.service",
						"load":      "loaded",
						"active":    "failed",
						"sub":       "failed",
						"uf_state":  "enabled",
						"uf_preset": "disabled",
					},
					map[string]interface{}{
						"load_code":    0,
						"active_code":  3,
						"sub_code":     12,
						"status_errno": 10,
						"restarts":     1,
						"mem_current":  1000,
						"mem_peak":     2000,
						"swap_current": 3000,
						"swap_peak":    4000,
						"mem_avail":    5000,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example not-found inactive dead",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "not-found",
						ActiveState: "inactive",
						SubState:    "dead",
					},
					ufPreset: "disabled",
					ufState:  "enabled",
					properties: map[string]interface{}{
						"Id": "example.service",
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"systemd_units",
					map[string]string{
						"name":      "example.service",
						"load":      "not-found",
						"active":    "inactive",
						"sub":       "dead",
						"uf_state":  "enabled",
						"uf_preset": "disabled",
					},
					map[string]interface{}{
						"load_code":   2,
						"active_code": 2,
						"sub_code":    1,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "example unknown unknown unknown",
			properties: map[string]properties{
				"example.service": {
					utype: "Service",
					state: sdbus.UnitStatus{
						Name:        "example.service",
						LoadState:   "unknown",
						ActiveState: "unknown",
						SubState:    "unknown",
					},
					ufPreset: "unknown",
					ufState:  "unknown",
					properties: map[string]interface{}{
						"Id": "example.service",
					},
				},
			},
			expectedErr: "parsing field 'load' failed, value not in map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin. Do NOT call Start() as this would connect to
			// the real systemd daemon.
			plugin := &SystemdUnits{
				Pattern:    "examp*",
				SubCommand: "show",
				Timeout:    config.Duration(time.Second),
			}
			require.NoError(t, plugin.Init())

			// Create a fake client to inject data
			plugin.client = &fakeClient{
				units:     tt.properties,
				connected: true,
			}
			defer plugin.Stop()

			// Run gather
			var acc testutil.Accumulator
			err := acc.GatherError(plugin.Gather)
			if tt.expectedErr != "" {
				require.ErrorContains(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)

			// Do the comparison
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestMultiInstance(t *testing.T) {
	// Setup plugin. Do NOT call Start() as this would connect to
	// the real systemd daemon.
	plugin := &SystemdUnits{
		Pattern: "examp* user@*",
		Timeout: config.Duration(time.Second),
	}
	require.NoError(t, plugin.Init())

	// Create a fake client to inject data
	plugin.client = &fakeClient{
		units: map[string]properties{
			"example.service": {
				utype: "Service",
				state: sdbus.UnitStatus{
					Name:        "example.service",
					LoadState:   "loaded",
					ActiveState: "active",
					SubState:    "running",
				},
			},
			"user-runtime-dir@1000.service": {
				uf:    "user-runtime-dir@.service",
				utype: "Service",
				state: sdbus.UnitStatus{
					Name:        "user-runtime-dir@1000.service",
					LoadState:   "loaded",
					ActiveState: "active",
					SubState:    "exited",
				},
			},
			"user@1000.service": {
				uf:    "user@.service",
				utype: "Service",
				state: sdbus.UnitStatus{
					Name:        "user@1000.service",
					LoadState:   "loaded",
					ActiveState: "active",
					SubState:    "running",
				},
			},
			"user@1001.service": {
				uf:    "user@.service",
				utype: "Service",
				state: sdbus.UnitStatus{
					Name:        "user@1001.service",
					LoadState:   "loaded",
					ActiveState: "active",
					SubState:    "exited",
				},
			},
		},
		connected: true,
	}
	defer plugin.Stop()

	// Run gather
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	// Define expectation
	expected := []telegraf.Metric{
		metric.New(
			"systemd_units",
			map[string]string{
				"name":   "example.service",
				"load":   "loaded",
				"active": "active",
				"sub":    "running",
			},
			map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"systemd_units",
			map[string]string{
				"name":   "user@1000.service",
				"load":   "loaded",
				"active": "active",
				"sub":    "running",
			},
			map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"systemd_units",
			map[string]string{
				"name":   "user@1001.service",
				"load":   "loaded",
				"active": "active",
				"sub":    "exited",
			},
			map[string]interface{}{
				"load_code":   0,
				"active_code": 0,
				"sub_code":    4,
			},
			time.Unix(0, 0),
		),
	}

	// Do the comparison
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
}

// Fake client implementation
type fakeClient struct {
	units     map[string]properties
	connected bool
}

func (c *fakeClient) Connected() bool {
	return !c.connected
}

func (c *fakeClient) Close() {
	c.connected = false
}

func (c *fakeClient) ListUnitFilesByPatternsContext(_ context.Context, _, pattern []string) ([]sdbus.UnitFile, error) {
	f := filter.MustCompile(pattern)

	var files []sdbus.UnitFile
	seen := make(map[string]bool)
	for name, props := range c.units {
		var uf string
		if props.uf != "" && f.Match(props.uf) {
			uf = "/usr/lib/systemd/system/" + props.uf
		} else if props.uf == "" && f.Match(name) {
			uf = "/usr/lib/systemd/system/" + name
		} else {
			continue
		}

		if !seen[uf] {
			files = append(files, sdbus.UnitFile{Path: uf, Type: "unknown"})
		}
		seen[uf] = true
	}

	return files, nil
}

func (c *fakeClient) ListUnitsByNamesContext(_ context.Context, units []string) ([]sdbus.UnitStatus, error) {
	var states []sdbus.UnitStatus
	for name, u := range c.units {
		for _, requestedName := range units {
			if name == requestedName {
				states = append(states, u.state)
				break
			}
		}
	}

	return states, nil
}

func (c *fakeClient) GetUnitTypePropertiesContext(_ context.Context, unit, unitType string) (map[string]interface{}, error) {
	u, found := c.units[unit]
	if !found {
		return nil, nil
	}
	if u.utype != unitType {
		return nil, fmt.Errorf("Unknown interface 'org.freedesktop.systemd1.%s", unitType)
	}
	return u.properties, nil
}

func (c *fakeClient) GetUnitPropertyContext(_ context.Context, unit, propertyName string) (*sdbus.Property, error) {
	u, found := c.units[unit]
	if !found {
		return nil, nil
	}

	switch propertyName {
	case "UnitFileState":
		return &sdbus.Property{Name: propertyName, Value: dbus.MakeVariant(u.ufState)}, nil
	case "UnitFilePreset":
		return &sdbus.Property{Name: propertyName, Value: dbus.MakeVariant(u.ufPreset)}, nil
	}
	return nil, errors.New("unknown property")
}

func (c *fakeClient) ListUnitsContext(_ context.Context) ([]sdbus.UnitStatus, error) {
	units := make([]sdbus.UnitStatus, 0, len(c.units))
	for _, u := range c.units {
		units = append(units, u.state)
	}
	return units, nil
}
