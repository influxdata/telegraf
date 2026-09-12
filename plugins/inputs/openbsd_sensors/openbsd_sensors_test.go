package openbsd_sensors

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func sysctlOutput(output string) runner {
	return func(string, config.Duration) (*bytes.Buffer, error) {
		return bytes.NewBufferString(output), nil
	}
}

func TestParseFullOutput(t *testing.T) {
	plugin := &OpenbsdSensors{
		run: sysctlOutput(fullOutput),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu0",
				"sensor": "temp0",
				"type":   "temp",
			},
			map[string]interface{}{
				"value": float64(43.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu0",
				"sensor": "frequency0",
				"type":   "frequency",
			},
			map[string]interface{}{
				"value": float64(2250000000.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu1",
				"sensor": "frequency0",
				"type":   "frequency",
			},
			map[string]interface{}{
				"value": float64(1400000000.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "acpitz0",
				"sensor":      "temp0",
				"type":        "temp",
				"description": "zone temperature",
			},
			map[string]interface{}{
				"value": float64(27.80),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "pchtemp0",
				"sensor": "temp0",
				"type":   "temp",
			},
			map[string]interface{}{
				"value": float64(41.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "softraid0",
				"sensor":      "drive0",
				"type":        "drive",
				"description": "sd2",
			},
			map[string]interface{}{
				"value":  float64(4),
				"state":  "online",
				"status": "OK",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "softraid0",
				"sensor":      "drive1",
				"type":        "drive",
				"description": "sd3",
			},
			map[string]interface{}{
				"value":  float64(7),
				"state":  "rebuilding",
				"status": "WARNING",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "nmea0",
				"sensor":      "indicator0",
				"type":        "indicator",
				"description": "Signal",
			},
			map[string]interface{}{
				"value":  float64(1),
				"status": "OK",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "nmea0",
				"sensor":      "timedelta0",
				"type":        "timedelta",
				"description": "GPS differential",
			},
			map[string]interface{}{
				"value":  float64(-0.000104),
				"status": "OK",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "nmea0",
				"sensor":      "angle0",
				"type":        "angle",
				"description": "Latitude",
			},
			map[string]interface{}{
				"value":  float64(22.7643),
				"status": "OK",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "nmea0",
				"sensor":      "distance0",
				"type":        "distance",
				"description": "Altitude",
			},
			map[string]interface{}{
				"value":  float64(227.200),
				"status": "OK",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "nmea0",
				"sensor":      "velocity0",
				"type":        "velocity",
				"description": "Ground speed",
			},
			map[string]interface{}{
				"value":  float64(0.000),
				"status": "OK",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "upd0",
				"sensor": "percent0",
				"type":   "percent",
			},
			map[string]interface{}{
				"value":  float64(100.00),
				"status": "OK",
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestDeviceFilter(t *testing.T) {
	plugin := &OpenbsdSensors{
		Devices: []string{"cpu*", "acpitz0"},
		run:     sysctlOutput(fullOutput),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu0",
				"sensor": "temp0",
				"type":   "temp",
			},
			map[string]interface{}{
				"value": float64(43.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu0",
				"sensor": "frequency0",
				"type":   "frequency",
			},
			map[string]interface{}{
				"value": float64(2250000000.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu1",
				"sensor": "frequency0",
				"type":   "frequency",
			},
			map[string]interface{}{
				"value": float64(1400000000.00),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device":      "acpitz0",
				"sensor":      "temp0",
				"type":        "temp",
				"description": "zone temperature",
			},
			map[string]interface{}{
				"value": float64(27.80),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestUnknownValue(t *testing.T) {
	// SENSOR_FUNKNOWN prints "unknown". Without a status there is nothing
	// to report; with a status, status_code provides a numeric field.
	output := `hw.sensors.foo0.temp0=unknown
hw.sensors.foo0.temp1=unknown, UNKNOWN
`
	plugin := &OpenbsdSensors{
		run: sysctlOutput(output),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	expected := []telegraf.Metric{
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "foo0",
				"sensor": "temp1",
				"type":   "temp",
			},
			map[string]interface{}{
				"status":      "UNKNOWN",
				"status_code": float64(4),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestDriveStates(t *testing.T) {
	output := `hw.sensors.softraid0.drive0=empty
hw.sensors.softraid0.drive1=ready
hw.sensors.softraid0.drive2=powering up (sd2)
hw.sensors.softraid0.drive3=online (sd3), OK
hw.sensors.softraid0.drive4=idle
hw.sensors.softraid0.drive5=active
hw.sensors.softraid0.drive6=rebuilding (sd4), WARNING
hw.sensors.softraid0.drive7=powering down
hw.sensors.softraid0.drive8=failed, CRITICAL
hw.sensors.softraid0.drive9=degraded, WARNING
hw.sensors.softraid0.drive10=bogus
`
	plugin := &OpenbsdSensors{
		run: sysctlOutput(output),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.Errors, 1)

	expected := []telegraf.Metric{
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive0", "type": "drive"},
			map[string]interface{}{"state": "empty", "value": float64(1)},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive1", "type": "drive"},
			map[string]interface{}{"state": "ready", "value": float64(2)},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive2", "type": "drive", "description": "sd2"},
			map[string]interface{}{"state": "powering up", "value": float64(3)},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive3", "type": "drive", "description": "sd3"},
			map[string]interface{}{"state": "online", "value": float64(4), "status": "OK"},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive4", "type": "drive"},
			map[string]interface{}{"state": "idle", "value": float64(5)},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive5", "type": "drive"},
			map[string]interface{}{"state": "active", "value": float64(6)},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive6", "type": "drive", "description": "sd4"},
			map[string]interface{}{"state": "rebuilding", "value": float64(7), "status": "WARNING"},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive7", "type": "drive"},
			map[string]interface{}{"state": "powering down", "value": float64(8)},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive8", "type": "drive"},
			map[string]interface{}{"state": "failed", "value": float64(9), "status": "CRITICAL"},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive9", "type": "drive"},
			map[string]interface{}{"state": "degraded", "value": float64(10), "status": "WARNING"},
			time.Unix(0, 0),
		),
		metric.New("openbsd_sensors",
			map[string]string{"device": "softraid0", "sensor": "drive10", "type": "drive"},
			map[string]interface{}{"state": "bogus"},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestMalformedLines(t *testing.T) {
	output := `no equal sign
hw.sensors.incomplete0=1.00
hw.sensors.cpu0.temp0=not-a-number degC
hw.cpuspeed=2200
hw.sensors.cpu0.temp1=43.00 degC
`
	plugin := &OpenbsdSensors{
		run: sysctlOutput(output),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.Errors, 4)

	// The valid line is still gathered
	expected := []telegraf.Metric{
		metric.New(
			"openbsd_sensors",
			map[string]string{
				"device": "cpu0",
				"sensor": "temp1",
				"type":   "temp",
			},
			map[string]interface{}{
				"value": float64(43.00),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestEmptyOutput(t *testing.T) {
	// Hosts without any sensor device produce no output at all
	plugin := &OpenbsdSensors{
		run: sysctlOutput(""),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.GetTelegrafMetrics())
	require.Empty(t, acc.Errors)
}

var fullOutput = `hw.sensors.cpu0.temp0=43.00 degC
hw.sensors.cpu0.frequency0=2250000000.00 Hz
hw.sensors.cpu1.frequency0=1400000000.00 Hz
hw.sensors.acpitz0.temp0=27.80 degC (zone temperature)
hw.sensors.pchtemp0.temp0=41.00 degC
hw.sensors.softraid0.drive0=online (sd2), OK
hw.sensors.softraid0.drive1=rebuilding (sd3), WARNING
hw.sensors.nmea0.indicator0=On (Signal), OK
hw.sensors.nmea0.timedelta0=-0.000104 secs (GPS differential), OK, Sun Jul 19 23:16:01.999
hw.sensors.nmea0.angle0=22.7643 degrees (Latitude), OK
hw.sensors.nmea0.distance0=227.200 m (Altitude), OK
hw.sensors.nmea0.velocity0=0.000 m/s (Ground speed), OK
hw.sensors.upd0.percent0=100.00%, OK
`
