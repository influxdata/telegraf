package redfish

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestDellApis(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "test", "test", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/dell/systems.json")
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell/thermal.json")
		case "/redfish/v1/Chassis/System.Embedded.1/Power":
			http.ServeFile(w, r, "testdata/dell/power.json")
		case "/redfish/v1/Chassis/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell/chassis.json")
		case "/redfish/v1/Systems/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell/systems_1.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	address, _, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)

	expectedMetrics := []telegraf.Metric{
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "CPU1 Temp",
				"member_id": "iDRAC.Embedded.1#CPU1Temp",
				"source":    "tpa-hostname",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 3.0,
				"lower_threshold_fatal":    3.0,
				"reading_celsius":          40.0,
				"upper_threshold_critical": 93.0,
				"upper_threshold_fatal":    93.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan1A",
				"member_id": "0x17||Fan.Embedded.1A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_rpm":              17760,
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan1B",
				"member_id": "0x17||Fan.Embedded.1B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15360,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan2A",
				"member_id": "0x17||Fan.Embedded.2A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              17880,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan2B",
				"member_id": "0x17||Fan.Embedded.2B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15120,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan3A",
				"member_id": "0x17||Fan.Embedded.3A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              18000,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan3B",
				"member_id": "0x17||Fan.Embedded.3B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15600,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan4A",
				"member_id": "0x17||Fan.Embedded.4A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              17280,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan4B",
				"member_id": "0x17||Fan.Embedded.4B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15360,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan5A",
				"member_id": "0x17||Fan.Embedded.5A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              17640,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan5B",
				"member_id": "0x17||Fan.Embedded.5B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15600,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan6A",
				"member_id": "0x17||Fan.Embedded.6A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              17760,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan6B",
				"member_id": "0x17||Fan.Embedded.6B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15600,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan7A",
				"member_id": "0x17||Fan.Embedded.7A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              17400,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan7B",
				"member_id": "0x17||Fan.Embedded.7B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15720,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan8A",
				"member_id": "0x17||Fan.Embedded.8A",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              18000,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board Fan8B",
				"member_id": "0x17||Fan.Embedded.8B",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"lower_threshold_critical": 600,
				"lower_threshold_fatal":    600,
				"reading_rpm":              15840,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powercontrol",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Power Control",
				"member_id": "PowerControl",
				"address":   address,
			},
			map[string]interface{}{
				"average_consumed_watts": 426.0,
				"interval_in_min":        int64(1),
				"max_consumed_watts":     436.0,
				"min_consumed_watts":     425.0,
				"power_allocated_watts":  1628.0,
				"power_available_watts":  0.0,
				"power_capacity_watts":   1628.0,
				"power_consumed_watts":   429.0,
				"power_requested_watts":  704.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powersupplies",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "PS1 Status",
				"member_id":  "PSU.Slot.1",
				"address":    address,
				"health":     "OK",
				"state":      "Enabled",
				"serial_num": "PHARP0079G0049",
			},
			map[string]interface{}{
				"power_capacity_watts": 750.00,
				"power_input_watts":    900.0,
				"power_output_watts":   203.0,
				"line_input_voltage":   206.00,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_voltages",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board DIMM PG",
				"member_id": "iDRAC.Embedded.1#SystemBoardDIMMPG",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_volts": 1.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_voltages",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board NDC PG",
				"member_id": "iDRAC.Embedded.1#SystemBoardNDCPG",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_volts": 1.0,
			},
			time.Unix(0, 0),
		),

		metric.New(
			"redfish_power_voltages",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "System Board PS1 PG FAIL",
				"member_id": "iDRAC.Embedded.1#SystemBoardPS1PGFAIL",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_volts": 1.0,
			},
			time.Unix(0, 0),
		),
	}
	plugin := &Redfish{
		Address:          ts.URL,
		Username:         config.NewSecret([]byte("test")),
		Password:         config.NewSecret([]byte("test")),
		ComputerSystemID: "System.Embedded.1",
		IncludeMetrics:   []string{"thermal", "power"},
	}
	require.NoError(t, plugin.Init())
	var acc testutil.Accumulator

	err = plugin.Gather(&acc)
	require.NoError(t, err)
	require.True(t, acc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expectedMetrics, acc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func TestHPApis(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "test", "test", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/hp/systems.json")
		case "/redfish/v1/Chassis/1/Thermal":
			http.ServeFile(w, r, "testdata/hp/thermal.json")
		case "/redfish/v1/Chassis/1/Power":
			http.ServeFile(w, r, "testdata/hp/power.json")
		case "/redfish/v1/Chassis/DE043000/Drives/0":
			http.ServeFile(w, r, "testdata/hp/storage_drive0.json")
		case "/redfish/v1/Systems/1":
			http.ServeFile(w, r, "testdata/hp/systems_1.json")
		case "/redfish/v1/Systems/1/Storage/":
			http.ServeFile(w, r, "testdata/hp/storage.json")
		case "/redfish/v1/Systems/1/Storage/DE043000":
			http.ServeFile(w, r, "testdata/hp/storage_de043000.json")
		case "/redfish/v1/Chassis/1/":
			http.ServeFile(w, r, "testdata/hp/chassis.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	address, _, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)

	expectedMetricsHp := []telegraf.Metric{
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "01-Inlet Ambient",
				"member_id": "0",
				"source":    "tpa-hostname",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          19.0,
				"upper_threshold_critical": 42.0,
				"upper_threshold_fatal":    47.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "44-P/S 2 Zone",
				"source":    "tpa-hostname",
				"member_id": "42",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          34.0,
				"upper_threshold_critical": 75.0,
				"upper_threshold_fatal":    80.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 1",
				"member_id": "0",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 2",
				"member_id": "1",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 3",
				"member_id": "2",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powercontrol",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "",
				"member_id": "0",
				"address":   address,
			},
			map[string]interface{}{
				"average_consumed_watts": 221.0,
				"interval_in_min":        int64(20),
				"max_consumed_watts":     252.0,
				"min_consumed_watts":     220.0,
				"power_capacity_watts":   1600.0,
				"power_consumed_watts":   221.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powersupplies",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "HpeServerPowerSupply",
				"member_id":  "0",
				"address":    address,
				"health":     "OK",
				"state":      "Enabled",
				"serial_num": "5WEBP0B8JAQ2K9",
			},
			map[string]interface{}{
				"power_capacity_watts":    800.0,
				"line_input_voltage":      205.0,
				"last_power_output_watts": 0.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powersupplies",
			map[string]string{
				"source":     "tpa-hostname",
				"name":       "HpeServerPowerSupply",
				"member_id":  "1",
				"address":    address,
				"health":     "OK",
				"state":      "Enabled",
				"serial_num": "5WEBP0B8JAQ2KL",
			},
			map[string]interface{}{
				"power_capacity_watts":    800.0,
				"line_input_voltage":      205.0,
				"last_power_output_watts": 90.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_storage",
			map[string]string{
				"address":       address,
				"disk_health":   "OK",
				"disk_state":    "Enabled",
				"health_rollup": "OK",
				"location":      "Slot=22:Port=2:Box=1:Bay=1",
				"manufacturer":  "HPE",
				"media_type":    "SSD",
				"model":         "Modelname",
				"protocol":      "SAS",
				"serial_number": "SERIALNUMBER",
				"source":        "tpa-hostname",
				"state":         "Enabled",
			},
			map[string]interface{}{
				"capacity_bytes": 1600321314816,
				"speed_gbs":      22.0,
			},
			time.Unix(0, 0),
		),
	}

	hpPlugin := &Redfish{
		Address:          ts.URL,
		Username:         config.NewSecret([]byte("test")),
		Password:         config.NewSecret([]byte("test")),
		ComputerSystemID: "1",
		IncludeMetrics:   []string{"thermal", "power", "storage"},
	}
	require.NoError(t, hpPlugin.Init())
	var hpAcc testutil.Accumulator

	err = hpPlugin.Gather(&hpAcc)
	require.NoError(t, err)
	require.True(t, hpAcc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expectedMetricsHp, hpAcc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func TestHPilo4Apis(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "test", "test", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/hp/systems.json")
		case "/redfish/v1/Chassis/1/Thermal":
			http.ServeFile(w, r, "testdata/hp/thermal_ilo4.json")
		case "/redfish/v1/Chassis/1/Power":
			http.ServeFile(w, r, "testdata/hp/power.json")
		case "/redfish/v1/Systems/1":
			http.ServeFile(w, r, "testdata/hp/systems_1.json")
		case "/redfish/v1/Chassis/1/":
			http.ServeFile(w, r, "testdata/hp/chassis.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	address, _, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)

	expectedMetricsHp := []telegraf.Metric{
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "01-Inlet Ambient",
				"member_id": "0",
				"source":    "tpa-hostname",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          19.0,
				"upper_threshold_critical": 42.0,
				"upper_threshold_fatal":    47.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":      "44-P/S 2 Zone",
				"member_id": "42",
				"source":    "tpa-hostname",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_celsius":          34.0,
				"upper_threshold_critical": 75.0,
				"upper_threshold_fatal":    80.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"address":   address,
				"health":    "OK",
				"member_id": "",
				"name":      "Fan 1",
				"source":    "tpa-hostname",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 17,
			},
			time.Unix(0, 0),
		),
	}

	hpPlugin := &Redfish{
		Address:          ts.URL,
		Username:         config.NewSecret([]byte("test")),
		Password:         config.NewSecret([]byte("test")),
		ComputerSystemID: "1",
		IncludeMetrics:   []string{"thermal"},
	}
	require.NoError(t, hpPlugin.Init())
	var hpAcc testutil.Accumulator

	err = hpPlugin.Gather(&hpAcc)
	require.NoError(t, err)
	require.True(t, hpAcc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expectedMetricsHp, hpAcc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func checkAuth(r *http.Request, username, password, token string) bool {
	// The base path requires not auth
	if r.URL.Path == "/redfish/v1/" {
		return true
	}

	authHeader := r.Header.Get("X-Auth-Token")
	if authHeader != "" {
		return authHeader == token
	}
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return user == username && pass == password
}

func TestBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "testing", "testing", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/dell/systems.json")
		case "/redfish/v1/Systems/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell/systems_1.json")
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell/thermal.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	r := &Redfish{
		Address:          ts.URL,
		Username:         config.NewSecret([]byte("test")),
		Password:         config.NewSecret([]byte("test")),
		ComputerSystemID: "System.Embedded.1",
		IncludeMetrics:   []string{"thermal", "power"},
	}

	var acc testutil.Accumulator
	require.NoError(t, r.Init())
	_, err := url.Parse(ts.URL)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.EqualError(t, err, "failed to retrieve some items: [{\"link\":\"/redfish/v1/Systems/\",\"error\":\"401: Unauthorized.\\n\"}]")
}

func TestTokenAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "testing", "testing", "faulty-token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/dell/systems.json")
		case "/redfish/v1/Systems/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell/systems_1.json")
		case "/redfish/v1/Chassis/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell/systems_1.json")
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell/thermal.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	r := &Redfish{
		Address:          ts.URL,
		Token:            config.NewSecret([]byte("token")),
		ComputerSystemID: "System.Embedded.1",
		IncludeMetrics:   []string{"thermal", "power"},
	}

	var acc testutil.Accumulator
	require.NoError(t, r.Init())
	_, err := url.Parse(ts.URL)
	require.NoError(t, err)
	err = r.Gather(&acc)
	require.EqualError(t, err, "failed to retrieve some items: [{\"link\":\"/redfish/v1/Systems/\",\"error\":\"401: Unauthorized.\\n\"}]")
}

func TestNoUsernameorPasswordConfiguration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "testing", "testing", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/dell/systems.json")
		case "/redfish/v1/Systems/System.Embedded.1":
			http.ServeFile(w, r, "testdata/dell/systems.json")
		case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
			http.ServeFile(w, r, "testdata/dell/thermal.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	r := &Redfish{
		Address:          ts.URL,
		ComputerSystemID: "System.Embedded.1",
		IncludeMetrics:   []string{"thermal", "power"},
	}

	err := r.Init()
	require.Error(t, err)
	require.EqualError(t, err, "Empty token or username or password. Provide either a token or user and password")
}

func TestInvalidDellJSON(t *testing.T) {
	tests := []struct {
		name             string
		thermalfilename  string
		powerfilename    string
		chassisfilename  string
		hostnamefilename string
	}{
		{
			name:             "check Thermal",
			thermalfilename:  "testdata/dell/thermal_invalid.json",
			powerfilename:    "testdata/dell/power.json",
			chassisfilename:  "testdata/dell/chassis.json",
			hostnamefilename: "testdata/dell/systems_1.json",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/dell/thermal.json",
			powerfilename:    "testdata/dell/power_invalid.json",
			chassisfilename:  "testdata/dell/chassis.json",
			hostnamefilename: "testdata/dell/systems_1.json",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/dell/thermal.json",
			powerfilename:    "testdata/dell/power.json",
			chassisfilename:  "testdata/dell/chassis.json",
			hostnamefilename: "testdata/dell/systems_invalid.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !checkAuth(r, "test", "test", "token") {
					http.Error(w, "Unauthorized.", http.StatusUnauthorized)
					return
				}

				switch r.URL.Path {
				case "/redfish/v1/":
					http.ServeFile(w, r, "testdata/base.json")
				case "/redfish/v1/Systems/":
					http.ServeFile(w, r, "testdata/dell/systems.json")
				case "/redfish/v1/Chassis/System.Embedded.1/Thermal":
					http.ServeFile(w, r, tt.thermalfilename)
				case "/redfish/v1/Chassis/System.Embedded.1/Power":
					http.ServeFile(w, r, tt.powerfilename)
				case "/redfish/v1/Chassis/System.Embedded.1":
					http.ServeFile(w, r, tt.chassisfilename)
				case "/redfish/v1/Systems/System.Embedded.1":
					http.ServeFile(w, r, tt.hostnamefilename)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer ts.Close()

			plugin := &Redfish{
				Address:          ts.URL,
				Username:         config.NewSecret([]byte("test")),
				Password:         config.NewSecret([]byte("test")),
				ComputerSystemID: "System.Embedded.1",
				IncludeMetrics:   []string{"thermal", "power"},
			}

			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			err := plugin.Gather(&acc)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid character")
		})
	}
}

func TestInvalidHPJSON(t *testing.T) {
	tests := []struct {
		name             string
		thermalfilename  string
		powerfilename    string
		hostnamefilename string
		chassisfilename  string
	}{
		{
			name:             "check Thermal",
			thermalfilename:  "testdata/hp/thermal_invalid.json",
			powerfilename:    "testdata/hp/power.json",
			hostnamefilename: "testdata/hp/systems_1.json",
			chassisfilename:  "testdata/hp/chassis.json",
		},
		{
			name:             "check Power",
			thermalfilename:  "testdata/hp/thermal.json",
			powerfilename:    "testdata/hp/power_invalid.json",
			hostnamefilename: "testdata/hp/systems_1.json",
			chassisfilename:  "testdata/hp/chassis.json",
		},
		{
			name:             "check Hostname",
			thermalfilename:  "testdata/hp/thermal.json",
			powerfilename:    "testdata/hp/power.json",
			hostnamefilename: "testdata/hp/systems_invalid.json",
			chassisfilename:  "testdata/hp/chassis.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !checkAuth(r, "test", "test", "token") {
					http.Error(w, "Unauthorized.", http.StatusUnauthorized)
					return
				}

				switch r.URL.Path {
				case "/redfish/v1/":
					http.ServeFile(w, r, "testdata/base.json")
				case "/redfish/v1/Systems/":
					http.ServeFile(w, r, "testdata/hp/systems.json")
				case "/redfish/v1/Systems/1":
					http.ServeFile(w, r, tt.hostnamefilename)
				case "/redfish/v1/Chassis/1/Thermal":
					http.ServeFile(w, r, tt.thermalfilename)
				case "/redfish/v1/Chassis/1/Power":
					http.ServeFile(w, r, tt.powerfilename)
				case "/redfish/v1/Chassis/1/":
					http.ServeFile(w, r, tt.chassisfilename)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer ts.Close()

			plugin := &Redfish{
				Address:          ts.URL,
				Username:         config.NewSecret([]byte("test")),
				Password:         config.NewSecret([]byte("test")),
				ComputerSystemID: "1",
				IncludeMetrics:   []string{"thermal", "power"},
			}

			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			err := plugin.Gather(&acc)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid character '{' looking for beginning of object key string")
		})
	}
}

func TestIncludeTagSetsConfiguration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "test", "test", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/hp/systems.json")
		case "/redfish/v1/Chassis/1/Thermal":
			http.ServeFile(w, r, "testdata/hp/thermal.json")
		case "/redfish/v1/Chassis/1/Power":
			http.ServeFile(w, r, "testdata/hp/power.json")
		case "/redfish/v1/Systems/1":
			http.ServeFile(w, r, "testdata/hp/systems_1.json")
		case "/redfish/v1/Chassis/1/":
			http.ServeFile(w, r, "testdata/hp/chassis.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	address, _, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)

	expectedMetricsHp := []telegraf.Metric{
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":                 "01-Inlet Ambient",
				"member_id":            "0",
				"source":               "tpa-hostname",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
			},
			map[string]interface{}{
				"reading_celsius":          19.0,
				"upper_threshold_critical": 42.0,
				"upper_threshold_fatal":    47.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_temperatures",
			map[string]string{
				"name":                 "44-P/S 2 Zone",
				"source":               "tpa-hostname",
				"member_id":            "42",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
			},
			map[string]interface{}{
				"reading_celsius":          34.0,
				"upper_threshold_critical": 75.0,
				"upper_threshold_fatal":    80.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":               "tpa-hostname",
				"name":                 "Fan 1",
				"member_id":            "0",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":               "tpa-hostname",
				"name":                 "Fan 2",
				"member_id":            "1",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermal_fans",
			map[string]string{
				"source":               "tpa-hostname",
				"name":                 "Fan 3",
				"member_id":            "2",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
			},
			map[string]interface{}{
				"reading_percent": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powercontrol",
			map[string]string{
				"source":               "tpa-hostname",
				"name":                 "",
				"member_id":            "0",
				"address":              address,
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
			},
			map[string]interface{}{
				"average_consumed_watts": 221.0,
				"interval_in_min":        int64(20),
				"max_consumed_watts":     252.0,
				"min_consumed_watts":     220.0,
				"power_capacity_watts":   1600.0,
				"power_consumed_watts":   221.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powersupplies",
			map[string]string{
				"source":               "tpa-hostname",
				"name":                 "HpeServerPowerSupply",
				"member_id":            "0",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
				"serial_num":           "5WEBP0B8JAQ2K9",
			},
			map[string]interface{}{
				"power_capacity_watts":    800.0,
				"line_input_voltage":      205.0,
				"last_power_output_watts": 0.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_power_powersupplies",
			map[string]string{
				"source":               "tpa-hostname",
				"name":                 "HpeServerPowerSupply",
				"member_id":            "1",
				"address":              address,
				"health":               "OK",
				"state":                "Enabled",
				"rack":                 "",
				"room":                 "",
				"row":                  "",
				"chassis_chassistype":  "RackMount",
				"chassis_manufacturer": "HP",
				"chassis_model":        "Proliant Gen10",
				"chassis_partnumber":   "CT6NWPYZ",
				"chassis_powerstate":   "On",
				"chassis_sku":          "CLFYTTWP",
				"chassis_serialnumber": "QWEVC007C99803",
				"chassis_state":        "Enabled",
				"chassis_health":       "OK",
				"serial_num":           "5WEBP0B8JAQ2KL",
			},
			map[string]interface{}{
				"power_capacity_watts":    800.0,
				"line_input_voltage":      205.0,
				"last_power_output_watts": 90.0,
			},
			time.Unix(0, 0),
		),
	}

	hpPlugin := &Redfish{
		Address:          ts.URL,
		Username:         config.NewSecret([]byte("test")),
		Password:         config.NewSecret([]byte("test")),
		ComputerSystemID: "1",
		IncludeTagSets:   []string{"chassis", "chassis.location"},
		IncludeMetrics:   []string{"thermal", "power"},
	}
	require.NoError(t, hpPlugin.Init())
	var hpAcc testutil.Accumulator

	err = hpPlugin.Gather(&hpAcc)
	require.NoError(t, err)
	require.True(t, hpAcc.HasMeasurement("redfish_thermal_temperatures"))
	testutil.RequireMetricsEqual(t, expectedMetricsHp, hpAcc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}

func TestSubsystemsApi(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r, "test", "test", "token") {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/redfish/v1/":
			http.ServeFile(w, r, "testdata/base.json")
		case "/redfish/v1/Systems/":
			http.ServeFile(w, r, "testdata/hp/systems.json")
		case "/redfish/v1/Chassis/1/ThermalSubsystem":
			http.ServeFile(w, r, "testdata/hp/thermal_subsys.json")
		case "/redfish/v1/Chassis/1/ThermalSubsystem/ThermalMetrics":
			http.ServeFile(w, r, "testdata/hp/thermal_metrics.json")
		case "/redfish/v1/Chassis/1/ThermalSubsystem/Fans":
			http.ServeFile(w, r, "testdata/hp/thermal_subsys_fans.json")
		case "/redfish/v1/Chassis/1/ThermalSubsystem/Fans/0":
			http.ServeFile(w, r, "testdata/hp/thermal_subsys_fans_0.json")
		case "/redfish/v1/Chassis/1/PowerSubsystem":
			http.ServeFile(w, r, "testdata/hp/power_subsys.json")
		case "/redfish/v1/Chassis/1/PowerSubsystem/PowerSupplies":
			http.ServeFile(w, r, "testdata/hp/power_subsys_psu.json")
		case "/redfish/v1/Chassis/1/PowerSubsystem/PowerSupplies/PowerSupply1":
			http.ServeFile(w, r, "testdata/hp/power_subsys_psu_1.json")
		case "/redfish/v1/Chassis/1/PowerSubsystem/PowerSupplies/PowerSupply1/Metrics":
			http.ServeFile(w, r, "testdata/hp/power_subsys_metrics.json")
		case "/redfish/v1/Systems/1":
			http.ServeFile(w, r, "testdata/hp/systems_1.json")
		case "/redfish/v1/Chassis/1/":
			http.ServeFile(w, r, "testdata/hp/chassis_subsys.json")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer ts.Close()

	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	address, _, err := net.SplitHostPort(u.Host)
	require.NoError(t, err)

	expectedMetricsHp := []telegraf.Metric{
		metric.New(
			"redfish_thermalsubsys_temperatures",
			map[string]string{
				"name":          "01-Inlet Ambient",
				"source":        "tpa-hostname",
				"address":       address,
				"health_rollup": "OK",
				"state":         "Enabled",
			},
			map[string]interface{}{
				"reading_celsius": 19.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermalsubsys_temperatures",
			map[string]string{
				"name":          "02-CPU 1 PkgTmp",
				"source":        "tpa-hostname",
				"address":       address,
				"health_rollup": "OK",
				"state":         "Enabled",
			},
			map[string]interface{}{
				"reading_celsius": 42.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_thermalsubsys_fans",
			map[string]string{
				"source":    "tpa-hostname",
				"name":      "Fan 1",
				"member_id": "0",
				"address":   address,
				"health":    "OK",
				"state":     "Enabled",
			},
			map[string]interface{}{
				"reading_percent": 20.0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_powersubsys_redundancy",
			map[string]string{
				"name":    "",
				"source":  "tpa-hostname",
				"address": address,
				"type":    "Failover",
				"health":  "OK",
				"state":   "UnavailableOffline",
			},
			map[string]interface{}{
				"redund_group_count": 0,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"redfish_powersubsys_powersupplies",
			map[string]string{
				"source":       "tpa-hostname",
				"name":         "Power Supply Bay 1",
				"hotpluggable": "false",
				"address":      address,
				"health":       "Warning",
				"state":        "Enabled",
				"serial_num":   "3488247",
			},
			map[string]interface{}{
				"power_capacity_watts": 400.0,
				"line_input_voltage":   230.2,
				"power_input_watts":    937.4,
				"power_output_watts":   937.4,
				"firmware_version":     "1.00",
			},
			time.Unix(0, 0),
		),
	}

	hpPlugin := &Redfish{
		Address:          ts.URL,
		Username:         config.NewSecret([]byte("test")),
		Password:         config.NewSecret([]byte("test")),
		ComputerSystemID: "1",
		IncludeMetrics:   []string{"thermal", "power"},
	}
	require.NoError(t, hpPlugin.Init())
	var hpAcc testutil.Accumulator

	err = hpPlugin.Gather(&hpAcc)
	require.NoError(t, err)
	require.True(t, hpAcc.HasMeasurement("redfish_thermalsubsys_temperatures"))
	testutil.RequireMetricsEqual(t, expectedMetricsHp, hpAcc.GetTelegrafMetrics(),
		testutil.IgnoreTime())
}
