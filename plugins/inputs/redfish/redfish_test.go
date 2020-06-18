package redfish

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestApis(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !checkAuth(r, "test", "test") {
			http.Error(w, "Unauthorized.", 401)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell_thermal.json")
		case "/redfish/v1/Chassis/System.Embedded.1/Power":
			http.ServeFile(w, r, "testdata/dell_power.json")
		case "/redfish/v1/Chassis/System.Embedded.1/":
			http.ServeFile(w, r, "testdata/dell_chassis.json")
		case "/redfish/v1/Systems/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell_systems.json")
		case "/redfish/v1/Chassis/System.Embedded.2/Thermal":
			http.ServeFile(w, r, "testdata/hp_thermal.json")
		case "/redfish/v1/Chassis/System.Embedded.2/Power":
			http.ServeFile(w, r, "testdata/hp_power.json")
		case "/redfish/v1/Systems/System.Embedded.2":
			http.ServeFile(w, r, "testdata/hp_systems.json")
		case "/redfish/v1/Chassis/System.Embedded.2/":
			http.ServeFile(w, r, "testdata/hp_power.json")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	expected_metrics_hp := []telegraf.Metric{
		testutil.MustMetric(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":    "01-Inlet Ambient",
				"source":  "tpa-hostname",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          19,
				"upper_threshold_critical": 42,
				"upper_threshold_fatal":    47,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":    "44-P/S 2 Zone",
				"source":  "tpa-hostname",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          34,
				"upper_threshold_critical": 75,
				"upper_threshold_fatal":    80,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":  "tpa-hostname",
				"name":    "Fan 1",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":  "tpa-hostname",
				"name":    "Fan 2",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":  "tpa-hostname",
				"name":    "Fan 3",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_powersupplies",
			map[string]string{
				"source":  "tpa-hostname",
				"name":    "HpeServerPowerSupply",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"power_capacity_watts":    800.0,
				"line_input_voltage":      205.0,
				"last_power_output_watts": 0.0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_powersupplies",
			map[string]string{
				"source":  "tpa-hostname",
				"name":    "HpeServerPowerSupply",
				"address": ts.URL,
				"health":  "OK",
				"state":   "Enabled",
			},
			map[string]interface{}{
				"power_capacity_watts":    800.0,
				"line_input_voltage":      205.0,
				"last_power_output_watts": 90.0,
			},
			time.Unix(0, 0),
		),
	}

	expected_metrics := []telegraf.Metric{
		testutil.MustMetric(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":       "CPU1 Temp",
				"source":     "tpa-hostname",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          40,
				"upper_threshold_critical": 93,
				"upper_threshold_fatal":    93,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan1A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17760,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan1B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15360,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan2A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17880,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan2B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15120,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan3A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              18000,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan3B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15600,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan4A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17280,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan4B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15360,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan5A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17640,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan5B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15600,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan6A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17760,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan6B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15600,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan7A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17400,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan7B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15720,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan8A",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              18000,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan8B",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              15840,
				"upper_threshold_critical": 0,
				"upper_threshold_fatal":    0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_powersupplies",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "PS1 Status",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"power_capacity_watts":    750.00,
				"power_input_watts":       900.0,
				"power_output_watts":      203.0,
				"last_power_output_watts": 0.0,
				"line_input_voltage":      206.00,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_voltages",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board DIMM PG",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_volts":            1.0,
				"upper_threshold_critical": 0.0,
				"upper_threshold_fatal":    0.0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_voltages",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board NDC PG",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_volts":            1.0,
				"upper_threshold_critical": 0.0,
				"upper_threshold_fatal":    0.0,
			},
			time.Unix(0, 0),
		),

		testutil.MustMetric(
			"redfish_power_voltages",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board PS1 PG FAIL",
				"address":    ts.URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"reading_volts":            1.0,
				"upper_threshold_critical": 0.0,
				"upper_threshold_fatal":    0.0,
			},
			time.Unix(0, 0),
		),
	}
	plugin := &Redfish{
		Address:           ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		ComputerSystemId:  "System.Embedded.1",
	}
	plugin.Init()
	var acc testutil.Accumulator

	err := plugin.Gather(&acc)
	require.NoError(t, err)
	require.True(t, acc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expected_metrics, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())

	hp_plugin := &Redfish{
		Address:           ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		ComputerSystemId:  "System.Embedded.2",
	}
	hp_plugin.Init()
	var hp_acc testutil.Accumulator

	err = hp_plugin.Gather(&hp_acc)
	require.NoError(t, err)
	require.True(t, hp_acc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expected_metrics_hp, hp_acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func checkAuth(r *http.Request, username, password string) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return user == username && pass == password
}

func TestConnection(t *testing.T) {

	r := &Redfish{
		Address:           "http://127.0.0.1",
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		ComputerSystemId:  "System.Embedded.1",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect: connection refused")
}

func TestInvalidUsernameorPassword(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !checkAuth(r, "testing", "testing") {
			http.Error(w, "Unauthorized.", 401)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell_thermal.json")
		default:
			panic("Cannot handle request")
		}
	}))
	defer ts.Close()

	r := &Redfish{
		Address:           ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		ComputerSystemId:  "System.Embedded.1",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	require.Error(t, err)
	require.EqualError(t, err, "received status code 401 (Unauthorized), expected 200")
}
func TestNoUsernameorPasswordConfiguration(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !checkAuth(r, "testing", "testing") {
			http.Error(w, "Unauthorized.", 401)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell_thermal.json")
		default:
			panic("Cannot handle request")
		}
	}))
	defer ts.Close()

	r := &Redfish{
		Address:          ts.URL,
		ComputerSystemId: "System.Embedded.1",
	}

	err := r.Init()
	require.Error(t, err)
	require.EqualError(t, err, "did not provide IP or username and password")
}

func TestInvalidDellJSON(t *testing.T) {

	tests := []struct {
		name             string
		thermalfilename  string
		powerfilename    string
		locationfilename string
		hostnamefilename string
	}{
		{
			name:             "check Thermal",
			thermalfilename:  "testdata/dell_thermalinvalid.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_chassis.json",
			hostnamefilename: "testdata/dell_systems.json",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_powerinvalid.json",
			locationfilename: "testdata/dell_chassis.json",
			hostnamefilename: "testdata/dell_systems.json",
		},
		{
			name:             "check Location",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_chassisinvalid.json",
			hostnamefilename: "testdata/dell_systems.json",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_chassis.json",
			hostnamefilename: "testdata/dell_systemsinvalid.json",
		},
	}
	for _, tt := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !checkAuth(r, "test", "test") {
				http.Error(w, "Unauthorized.", 401)
				return
			}

			switch r.URL.Path {
			case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
				http.ServeFile(w, r, tt.thermalfilename)
			case "/redfish/v1/Chassis/System.Embedded.1/Power":
				http.ServeFile(w, r, tt.powerfilename)
			case "/redfish/v1/Chassis/System.Embedded.1/":
				http.ServeFile(w, r, tt.locationfilename)
			case "/redfish/v1/Systems/System.Embedded.1":
				http.ServeFile(w, r, tt.hostnamefilename)
			default:
				panic("Cannot handle request")
			}
		}))
		defer ts.Close()

		plugin := &Redfish{
			Address:           ts.URL,
			BasicAuthUsername: "test",
			BasicAuthPassword: "test",
			ComputerSystemId:  "System.Embedded.1",
		}

		plugin.Init()

		var acc testutil.Accumulator
		err := plugin.Gather(&acc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error parsing input:")
	}
}

func TestInvalidHPJSON(t *testing.T) {

	tests := []struct {
		name             string
		thermalfilename  string
		powerfilename    string
		hostnamefilename string
	}{
		{
			name:             "check Thermal",
			thermalfilename:  "testdata/hp_thermalinvalid.json",
			powerfilename:    "testdata/hp_power.json",
			hostnamefilename: "testdata/hp_systems.json",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/hp_thermal.json",
			powerfilename:    "testdata/hp_powerinvalid.json",
			hostnamefilename: "testdata/hp_systems.json",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/hp_thermal.json",
			powerfilename:    "testdata/hp_power.json",
			hostnamefilename: "testdata/hp_systemsinvalid.json",
		},
	}
	for _, tt := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if !checkAuth(r, "test", "test") {
				http.Error(w, "Unauthorized.", 401)
				return
			}

			switch r.URL.Path {
			case "/redfish/v1/Chassis/System.Embedded.2/Thermal":
				http.ServeFile(w, r, tt.thermalfilename)
			case "/redfish/v1/Chassis/System.Embedded.2/Power":
				http.ServeFile(w, r, tt.powerfilename)
			case "/redfish/v1/Systems/System.Embedded.2":
				http.ServeFile(w, r, tt.hostnamefilename)
			default:
				panic("Cannot handle request")
			}
		}))
		defer ts.Close()

		plugin := &Redfish{
			Address:           ts.URL,
			BasicAuthUsername: "test",
			BasicAuthPassword: "test",
			ComputerSystemId:  "System.Embedded.2",
		}

		plugin.Init()

		var acc testutil.Accumulator
		err := plugin.Gather(&acc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error parsing input:")
	}
}
