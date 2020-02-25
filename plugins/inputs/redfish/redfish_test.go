package redfish

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestApis(t *testing.T) {

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	expected_metrics_hp := []telegraf.Metric{
		testutil.MustMetric(
			"cputemperature",
			map[string]string{
				"Name":     "01-Inlet Ambient",
				"Hostname": "tpa-hostname",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Health":      "OK",
				"State":       "Enabled",
				"Temperature": "19",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"cputemperature",
			map[string]string{
				"Name":     "44-P/S 2 Zone",
				"Hostname": "tpa-hostname",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Health":      "OK",
				"State":       "Enabled",
				"Temperature": "34",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "Fan 1",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Fanspeed": "23",
				"Health":   "OK",
				"State":    "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "Fan 2",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Fanspeed": "23",
				"Health":   "OK",
				"State":    "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "Fan 3",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Fanspeed": "23",
				"Health":   "OK",
				"State":    "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"powersupply",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "HpeServerPowerSupply",
				"MemberId": "0",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"PowerCapacityWatts":   "800",
				"LineInputVoltage":     "205",
				"LastPowerOutputWatts": "0",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"powersupply",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "HpeServerPowerSupply",
				"MemberId": "1",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"PowerCapacityWatts":   "800",
				"LineInputVoltage":     "205",
				"LastPowerOutputWatts": "90",
			},
			time.Unix(0, 0),
		),
	}

	expected_metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cputemperature",
			map[string]string{
				"Name":     "CPU1 Temp",
				"Hostname": "tpa-hostname",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter":  "",
				"Health":      "OK",
				"Rack":        "",
				"Room":        "",
				"Row":         "",
				"State":       "Enabled",
				"Temperature": "40",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan1A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "17760",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan1B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15360",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan2A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "17880",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan2B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15120",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan3A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "18000",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan3B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15600",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan4A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "17280",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan4B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15360",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan5A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "17640",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan5B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15600",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan6A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "17760",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan6B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15600",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan7A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "17400",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan7B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15720",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan8A",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "18000",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board Fan8B",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Fanspeed":   "15840",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"powersupply",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "PS1 Status",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter":         "",
				"Health":             "OK",
				"Rack":               "",
				"Room":               "",
				"Row":                "",
				"State":              "Enabled",
				"PowerCapacityWatts": "750",
				"PowerInputWatts":    "900",
				"PowerOutputWatts":   "203",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"voltages",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board DIMM PG",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
				"Voltage":    "1",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"voltages",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board NDC PG",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
				"Voltage":    "1",
			},
			time.Unix(0, 0),
		),

		testutil.MustMetric(
			"voltages",
			map[string]string{
				"Hostname": "tpa-hostname",
				"Name":     "System Board PS1 PG FAIL",
				"OOBIP":    ts.URL,
			},
			map[string]interface{}{
				"Datacenter": "",
				"Health":     "OK",
				"Rack":       "",
				"Room":       "",
				"Row":        "",
				"State":      "Enabled",
				"Voltage":    "1",
			},
			time.Unix(0, 0),
		),
	}
	plugin := &Redfish{
		Host:              ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
		Server:            "dell",
	}
	plugin.Init()
	var acc testutil.Accumulator

	_ = plugin.Gather(&acc)
	assert.True(t, acc.HasMeasurement("cputemperature"))
	testutil.RequireMetricsEqual(t, expected_metrics, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())

	hp_plugin := &Redfish{
		Host:              ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.2",
		Server:            "hp",
	}
	hp_plugin.Init()
	var hp_acc testutil.Accumulator

	_ = hp_plugin.Gather(&hp_acc)
	assert.True(t, hp_acc.HasMeasurement("cputemperature"))
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
		Host:              "https://127.0.0.1",
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
		Server:            "dell",
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "connect: connection refused")
	}
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
		Host:              ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
		Server:            "dell",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	assert.EqualError(t, err, "received status code 401 (Unauthorized), expected 200")
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
		Host:   ts.URL,
		Id:     "System.Embedded.1",
		Server: "dell",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	assert.EqualError(t, err, "Did not provide all the mandatory fields in the configuration")
}

func TestInvalidServerVarConfiguration(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !checkAuth(r, "test", "test") {
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
		Host:              ts.URL,
		BasicAuthUsername: "test",
		BasicAuthPassword: "test",
		Id:                "System.Embedded.1",
		Server:            "wtu",
	}

	var acc testutil.Accumulator
	r.Init()
	err := r.Gather(&acc)
	assert.EqualError(t, err, "Did not provide all the mandatory fields in the configuration")
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
			locationfilename: "testdata/dell_location.json",
			hostnamefilename: "testdata/dell_hostname.json",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_powerinvalid.json",
			locationfilename: "testdata/dell_location.json",
			hostnamefilename: "testdata/dell_hostname.json",
		},
		{
			name:             "check Location",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_locationinvalid.json",
			hostnamefilename: "testdata/dell_hostname.json",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/dell_thermal.json",
			powerfilename:    "testdata/dell_power.json",
			locationfilename: "testdata/dell_location.json",
			hostnamefilename: "testdata/dell_hostnameinvalid.json",
		},
	}
	for _, tt := range tests {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
			Host:              ts.URL,
			BasicAuthUsername: "test",
			BasicAuthPassword: "test",
			Id:                "System.Embedded.1",
			Server:            "dell",
		}

		plugin.Init()

		var acc testutil.Accumulator
		err := plugin.Gather(&acc)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "error parsing input:")
		}
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
			hostnamefilename: "testdata/hp_hostname.json",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/hp_thermal.json",
			powerfilename:    "testdata/hp_powerinvalid.json",
			hostnamefilename: "testdata/hp_hostname.json",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/hp_thermal.json",
			powerfilename:    "testdata/hp_power.json",
			hostnamefilename: "testdata/hp_hostnameinvalid.json",
		},
	}
	for _, tt := range tests {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
			Host:              ts.URL,
			BasicAuthUsername: "test",
			BasicAuthPassword: "test",
			Id:                "System.Embedded.2",
			Server:            "hp",
		}

		plugin.Init()

		var acc testutil.Accumulator
		err := plugin.Gather(&acc)

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "error parsing input:")
		}
	}
}
