package redfish

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestApis(t *testing.T) {

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
			http.ServeFile(w, r, "testdata/dell_location.json")
		case "/redfish/v1/Systems/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell_hostname.json")
		case "/redfish/v1/Chassis/System.Embedded.2/Thermal":
			http.ServeFile(w, r, "testdata/hp_thermal.json")
		case "/redfish/v1/Chassis/System.Embedded.2/Power":
			http.ServeFile(w, r, "testdata/hp_power.json")
		case "/redfish/v1/Systems/System.Embedded.2":
			http.ServeFile(w, r, "testdata/hp_hostname.json")
		case "/redfish/v1/Chassis/System.Embedded.2/":
			http.ServeFile(w, r, "testdata/hp_power.json")
		default:
			panic("Cannot handle request")
		}
	}))

	CUSTOM_URL := "127.0.0.1:3458"
	l, _ := net.Listen("tcp", CUSTOM_URL)
	ts.Listener = l
	ts.StartTLS()
	defer ts.Close()

	expected_metrics_hp := []telegraf.Metric{
		testutil.MustMetric(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "01-Inlet Ambient",
				"source":    "tpa-hostname",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"temperature": 19,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "44-P/S 2 Zone",
				"source":    "tpa-hostname",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"temperature": 34,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 1",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 23,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 2",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 23,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 3",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 23,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_powersupplies",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "HpeServerPowerSupply",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
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
				"source":    "tpa-hostname",
				"name":      "HpeServerPowerSupply",
				"source_ip": CUSTOM_URL,
				"health":    "OK",
				"state":     "Enabled",
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
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"temperature": 40,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan1A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 17760,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan1B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15360,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan2A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 17880,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan2B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15120,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan3A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 18000,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan3B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15600,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan4A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 17280,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan4B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15360,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan5A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 17640,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan5B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15600,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan6A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 17760,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan6B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15600,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan7A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 17400,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan7B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15720,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan8A",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 18000,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_thermal_fans",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board Fan8B",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"fanspeed": 15840,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_powersupplies",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "PS1 Status",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"power_capacity_watts":    750.0,
				"power_input_watts":       900.0,
				"power_output_watts":      203.0,
				"last_power_output_watts": 0.0,
				"line_input_voltage":      206.0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_voltages",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board DIMM PG",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"voltage": 1.0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"redfish_power_voltages",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board NDC PG",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"voltage": 1.0,
			},
			time.Unix(0, 0),
		),

		testutil.MustMetric(
			"redfish_power_voltages",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "System Board PS1 PG FAIL",
				"source_ip":  CUSTOM_URL,
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			map[string]interface{}{
				"voltage": 1.0,
			},
			time.Unix(0, 0),
		),
	}
	plugin := &Redfish{
		Host:              "127.0.0.1:3458",
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
	}
	plugin.Init()
	var acc testutil.Accumulator

	err := plugin.Gather(&acc)
	require.NoError(t, err)
	require.True(t, acc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expected_metrics, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())

	hp_plugin := &Redfish{
		Host:              "127.0.0.1:3458",
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.2",
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
		Host:              "127.0.0.1",
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect: connection refused")
}

func TestInvalidUsernameorPassword(t *testing.T) {

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
	CUSTOM_URL := "127.0.0.1:3458"
	l, _ := net.Listen("tcp", CUSTOM_URL)
	ts.Listener = l
	ts.StartTLS()
	defer ts.Close()

	r := &Redfish{
		Host:              CUSTOM_URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	require.Error(t, err)
	require.EqualError(t, err, "received status code 401 (Unauthorized), expected 200")
}
func TestNoUsernameorPasswordConfiguration(t *testing.T) {

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
	CUSTOM_URL := "127.0.0.1:3458"
	l, _ := net.Listen("tcp", CUSTOM_URL)
	ts.Listener = l
	ts.StartTLS()
	defer ts.Close()

	r := &Redfish{
		Host: CUSTOM_URL,
		Id:   "System.Embedded.1",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	require.Error(t, err)
	require.EqualError(t, err, "Did not provide IP or username and password")
}

func TestInvalidDellJSON(t *testing.T) {

	tests := []struct {
		name             string
		thermalfilename  string
		powerfilename    string
		locationfilename string
		hostnamefilename string
		CUSTOM_URL       string
	}{
		{
			name:             "check Thermal",
			thermalfilename:  "testdata/dell_thermalinvalid.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_location.json",
			hostnamefilename: "testdata/dell_hostname.json",
			CUSTOM_URL:       "127.0.0.1:3459",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_powerinvalid.json",
			locationfilename: "testdata/dell_location.json",
			hostnamefilename: "testdata/dell_hostname.json",
			CUSTOM_URL:       "127.0.0.1:3451",
		},
		{
			name:             "check Location",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_locationinvalid.json",
			hostnamefilename: "testdata/dell_hostname.json",
			CUSTOM_URL:       "127.0.0.1:3452",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_location.json",
			hostnamefilename: "testdata/dell_hostnameinvalid.json",
			CUSTOM_URL:       "127.0.0.1:3453",
		},
	}
	for _, tt := range tests {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
		l, _ := net.Listen("tcp", tt.CUSTOM_URL)
		ts.Listener = l
		ts.StartTLS()
		defer ts.Close()

		plugin := &Redfish{
			Host:              tt.CUSTOM_URL,
			BasicAuthUsername: "test",
			BasicAuthPassword: "test",
			Id:                "System.Embedded.1",
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
		CUSTOM_URL       string
	}{
		{
			name:             "check Thermal",
			thermalfilename:  "testdata/hp_thermalinvalid.json",
			powerfilename:    "testdata/hp_power.json",
			hostnamefilename: "testdata/hp_hostname.json",
			CUSTOM_URL:       "127.0.0.1:3278",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/hp_thermal.json",
			powerfilename:    "testdata/hp_powerinvalid.json",
			hostnamefilename: "testdata/hp_hostname.json",
			CUSTOM_URL:       "127.0.0.1:3289",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/hp_thermal.json",
			powerfilename:    "testdata/hp_power.json",
			hostnamefilename: "testdata/hp_hostnameinvalid.json",
			CUSTOM_URL:       "127.0.0.1:3290",
		},
	}
	for _, tt := range tests {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
		l, _ := net.Listen("tcp", tt.CUSTOM_URL)
		ts.Listener = l
		ts.StartTLS()
		defer ts.Close()

		plugin := &Redfish{
			Host:              tt.CUSTOM_URL,
			BasicAuthUsername: "test",
			BasicAuthPassword: "test",
			Id:                "System.Embedded.2",
		}

		plugin.Init()

		var acc testutil.Accumulator
		err := plugin.Gather(&acc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "error parsing input:")
	}
}
