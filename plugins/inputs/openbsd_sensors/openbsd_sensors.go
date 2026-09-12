//go:generate ../../../tools/readme_config_includer/generator
package openbsd_sensors

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "openbsd_sensors"

var (
	defaultBinary  = "/sbin/sysctl"
	defaultTimeout = config.Duration(5 * time.Second)

	// Sensor status strings as printed by print_sensor() in OpenBSD's
	// sbin/sysctl/sysctl.c; sensors with an unspecified status omit it.
	// Numeric codes match enum sensor_status (SENSOR_S_OK = 1, ...).
	sensorStatuses = map[string]float64{
		"OK":       1,
		"WARNING":  2,
		"CRITICAL": 3,
		"UNKNOWN":  4,
	}

	// Drive state strings as printed by print_sensor() and the matching
	// SENSOR_DRIVE_* values from sys/sys/sensors.h. value is emitted as well
	// as state because Prometheus cannot store string-only metrics.
	driveStates = map[string]float64{
		"empty":         1,
		"ready":         2,
		"powering up":   3,
		"online":        4,
		"idle":          5,
		"active":        6,
		"rebuilding":    7,
		"powering down": 8,
		"failed":        9,
		"degraded":      10,
	}
)

type OpenbsdSensors struct {
	Binary  string          `toml:"binary"`
	Devices []string        `toml:"devices"`
	Timeout config.Duration `toml:"timeout"`

	deviceFilter filter.Filter
	run          runner
}

type runner func(binary string, timeout config.Duration) (*bytes.Buffer, error)

func (*OpenbsdSensors) SampleConfig() string {
	return sampleConfig
}

func (s *OpenbsdSensors) Init() error {
	if len(s.Devices) > 0 {
		f, err := filter.Compile(s.Devices)
		if err != nil {
			return fmt.Errorf("compiling device filter failed: %w", err)
		}
		s.deviceFilter = f
	}

	return nil
}

func (s *OpenbsdSensors) Gather(acc telegraf.Accumulator) error {
	out, err := s.run(s.Binary, s.Timeout)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %w", err)
	}

	// Empty output is valid: hosts without any sensor device print nothing
	// for the hw.sensors tree, so no metrics are emitted in that case.
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		s.parseLine(line, acc)
	}

	return nil
}

// parseLine parses a single line of the sysctl output which, according to
// print_sensor() in OpenBSD's sbin/sysctl/sysctl.c, has the form
//
//	hw.sensors.<device>.<type><index>=<value>[ <unit>][ (<description>)][, <STATUS>][, <timestamp>]
//
// e.g.
//
//	hw.sensors.cpu0.temp0=43.00 degC
//	hw.sensors.softraid0.drive0=online (sd2), OK
//	hw.sensors.nmea0.timedelta0=-0.000104 secs (GPS differential), OK, Sun Jul 19 23:16:01.999
func (s *OpenbsdSensors) parseLine(line string, acc telegraf.Accumulator) {
	name, value, found := strings.Cut(line, "=")
	if !found {
		acc.AddError(fmt.Errorf("unexpected line %q", line))
		return
	}

	rest, found := strings.CutPrefix(name, "hw.sensors.")
	if !found {
		acc.AddError(fmt.Errorf("unexpected sensor name %q", name))
		return
	}
	device, sensor, found := strings.Cut(rest, ".")
	if !found {
		acc.AddError(fmt.Errorf("unexpected sensor name %q", name))
		return
	}

	if s.deviceFilter != nil && !s.deviceFilter.Match(device) {
		return
	}

	// The sensor type is the sensor name without the trailing index digits,
	// e.g. "temp0" -> "temp"
	sensorType := strings.TrimRightFunc(sensor, unicode.IsDigit)

	// Split off the optional ", <STATUS>" and ", <timestamp>" suffixes. The
	// timestamp (last value change, printed e.g. for timedelta sensors) is
	// deliberately discarded as it is no metric.
	parts := strings.Split(value, ", ")
	payload := parts[0]

	var status string
	for _, part := range parts[1:] {
		if _, ok := sensorStatuses[part]; ok {
			status = part
			break
		}
	}

	// Split off the optional trailing "(<description>)", e.g.
	// "27.80 degC (zone temperature)"
	var description string
	if strings.HasSuffix(payload, ")") {
		if idx := strings.LastIndex(payload, " ("); idx >= 0 {
			description = payload[idx+2 : len(payload)-1]
			payload = payload[:idx]
		}
	}

	fields := make(map[string]interface{}, 3)
	switch {
	case payload == "unknown":
		// SENSOR_FUNKNOWN: the kernel has no value. Do not invent one.
		// If a status is present, emit status_code so Prometheus has a
		// numeric field (string-only metrics are dropped).
		if code, ok := sensorStatuses[status]; ok {
			fields["status_code"] = code
		}
	case sensorType == "drive":
		fields["state"] = payload
		if v, ok := driveStates[payload]; ok {
			fields["value"] = v
		} else {
			acc.AddError(fmt.Errorf("unrecognized drive state %q of sensor %q", payload, name))
		}
	case sensorType == "indicator":
		// Boolean indicator printed as "On" or "Off"
		if payload == "On" {
			fields["value"] = float64(1)
		} else {
			fields["value"] = float64(0)
		}
	default:
		// Numeric value optionally followed by a unit (e.g. "43.00 degC");
		// percent and humidity sensors have the unit attached (e.g. "49.50%")
		number, _, _ := strings.Cut(payload, " ")
		number = strings.TrimSuffix(number, "%")
		v, err := strconv.ParseFloat(number, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("cannot parse value %q of sensor %q: %w", payload, name, err))
			return
		}
		fields["value"] = v
	}

	if status != "" {
		fields["status"] = status
	}
	if len(fields) == 0 {
		// Unknown value without status: nothing numeric or named to report
		return
	}

	tags := make(map[string]string, 4)
	tags["device"] = device
	tags["sensor"] = sensor
	tags["type"] = sensorType
	if description != "" {
		tags["description"] = description
	}

	acc.AddFields(measurement, fields, tags)
}

// sysctlRunner executes sysctl to query the hw.sensors tree and returns its
// output
func sysctlRunner(binary string, timeout config.Duration) (*bytes.Buffer, error) {
	cmd := exec.Command(binary, "hw.sensors")

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := internal.RunTimeout(cmd, time.Duration(timeout)); err != nil {
		return nil, fmt.Errorf("error running sysctl: %w", err)
	}

	return &out, nil
}

func init() {
	inputs.Add("openbsd_sensors", func() telegraf.Input {
		return &OpenbsdSensors{
			run:     sysctlRunner,
			Binary:  defaultBinary,
			Timeout: defaultTimeout,
		}
	})
}
