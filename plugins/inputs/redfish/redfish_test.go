// Copyright 2020, Verizon
//Licensed under the terms of the MIT License. SEE LICENSE file in project root for terms.

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
			"cpu_temperature",
			map[string]string{
				"name":     "01-Inlet Ambient",
				"hostname": "tpa-hostname",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"health":      "OK",
				"state":       "Enabled",
				"temperature": "19",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"cpu_temperature",
			map[string]string{
				"name":     "44-P/S 2 Zone",
				"hostname": "tpa-hostname",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"health":      "OK",
				"state":       "Enabled",
				"temperature": "34",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "Fan 1",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"fanspeed": "23",
				"health":   "OK",
				"state":    "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "Fan 2",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"fanspeed": "23",
				"health":   "OK",
				"state":    "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "Fan 3",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"fanspeed": "23",
				"health":   "OK",
				"state":    "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"powersupply",
			map[string]string{
				"hostname":  "tpa-hostname",
				"name":      "HpeServerPowerSupply",
				"member_id": "0",
				"oob_ip":    ts.URL,
			},
			map[string]interface{}{
				"power_capacity_watts":    "800",
				"line_input_voltage":      "205",
				"last_power_output_watts": "0",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"powersupply",
			map[string]string{
				"hostname":  "tpa-hostname",
				"name":      "HpeServerPowerSupply",
				"member_id": "1",
				"oob_ip":    ts.URL,
			},
			map[string]interface{}{
				"power_capacity_watts":    "800",
				"line_input_voltage":      "205",
				"last_power_output_watts": "90",
			},
			time.Unix(0, 0),
		),
	}

	expected_metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu_temperature",
			map[string]string{
				"name":     "CPU1 Temp",
				"hostname": "tpa-hostname",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter":  "",
				"health":      "OK",
				"rack":        "",
				"room":        "",
				"row":         "",
				"state":       "Enabled",
				"temperature": "40",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan1A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "17760",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan1B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15360",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan2A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "17880",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan2B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15120",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan3A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "18000",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan3B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15600",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan4A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "17280",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan4B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15360",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan5A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "17640",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan5B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15600",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan6A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "17760",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan6B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15600",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan7A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "17400",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan7B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15720",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan8A",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "18000",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fans",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board Fan8B",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"fanspeed":   "15840",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"powersupply",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "PS1 Status",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter":           "",
				"health":               "OK",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"state":                "Enabled",
				"power_capacity_watts": "750",
				"power_input_watts":    "900",
				"power_output_watts":   "203",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"voltages",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board DIMM PG",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
				"voltage":    "1",
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"voltages",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board NDC PG",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
				"voltage":    "1",
			},
			time.Unix(0, 0),
		),

		testutil.MustMetric(
			"voltages",
			map[string]string{
				"hostname": "tpa-hostname",
				"name":     "System Board PS1 PG FAIL",
				"oob_ip":   ts.URL,
			},
			map[string]interface{}{
				"datacenter": "",
				"health":     "OK",
				"rack":       "",
				"room":       "",
				"row":        "",
				"state":      "Enabled",
				"voltage":    "1",
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
		//	insecure_skip_verify : "true",
	}
	plugin.Init()
	var acc testutil.Accumulator

	_ = plugin.Gather(&acc)
	assert.True(t, acc.HasMeasurement("cpu_temperature"))
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
	assert.True(t, hp_acc.HasMeasurement("cpu_temperature"))
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
		//insecure_skip_verify : "true",
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
	assert.EqualError(t, err, "Did not provide IP or username and password")
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
	assert.EqualError(t, err, "Did not provide correct server information, supported server details are dell or hp")
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
