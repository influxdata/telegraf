//go:build openbsd

package sensors

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
)

// Sensor status strings as printed by print_sensor() in OpenBSD's
// sbin/sysctl/sysctl.c; sensors with an unspecified status omit it.
// Numeric codes match enum sensor_status (SENSOR_S_OK = 1, ...).
var sensorStatuses = map[string]float64{
	"OK":       1,
	"WARNING":  2,
	"CRITICAL": 3,
	"UNKNOWN":  4,
}

func (s *Sensors) Init() error {
	if s.path == "" {
		path, err := exec.LookPath("sysctl")
		if err != nil {
			return fmt.Errorf("looking up \"sysctl\" failed: %w", err)
		}
		s.path = path
	}

	f, err := filter.Compile(s.Devices)
	if err != nil {
		return fmt.Errorf("compiling device filter failed: %w", err)
	}
	s.deviceFilter = f

	return nil
}

func (s *Sensors) Gather(acc telegraf.Accumulator) error {
	cmd := exec.Command(s.path, "hw.sensors")
	out, err := internal.StdOutputTimeout(cmd, time.Duration(s.Timeout))
	if err != nil {
		return fmt.Errorf("failed to run command %q: %w - %s", strings.Join(cmd.Args, " "), err, string(out))
	}

	// Empty output is valid: hosts without any sensor device print nothing
	// for the hw.sensors tree, so no metrics are emitted in that case.
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		tags, fields, err := parseLine(line)
		if err != nil {
			return err
		}
		if len(fields) == 0 {
			continue
		}
		if s.deviceFilter != nil && !s.deviceFilter.Match(tags["device"]) {
			continue
		}
		acc.AddFields("sensors", fields, tags)
	}

	return scanner.Err()
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
func parseLine(line string) (map[string]string, map[string]any, error) {
	name, value, found := strings.Cut(line, "=")
	if !found {
		return nil, nil, fmt.Errorf("unexpected line %q", line)
	}

	rest, found := strings.CutPrefix(name, "hw.sensors.")
	if !found {
		return nil, nil, fmt.Errorf("unexpected sensor name %q", name)
	}
	device, sensor, found := strings.Cut(rest, ".")
	if !found {
		return nil, nil, fmt.Errorf("unexpected sensor name %q", name)
	}

	// The sensor type is the sensor name without the trailing index digits,
	// e.g. "temp0" -> "temp"
	sensorType := strings.TrimRightFunc(sensor, unicode.IsDigit)

	// Peel off the optional ", <STATUS>" and ", <timestamp>" suffixes from the
	// right. Neither of them contains ", " and both follow the description, so
	// a description is present exactly when what remains ends in ")". The
	// timestamp records the last value change and is no metric.
	var status string
	for range 2 {
		if strings.HasSuffix(value, ")") {
			break
		}
		idx := strings.LastIndex(value, ", ")
		if idx < 0 {
			break
		}
		if _, ok := sensorStatuses[value[idx+2:]]; ok {
			status = value[idx+2:]
		}
		value = value[:idx]
	}

	// Split off the optional trailing "(<description>)", e.g.
	// "27.80 degC (zone temperature)". Descriptions are printed unescaped and
	// may contain parentheses themselves, e.g. "+1.5V (Vccp)", so they start
	// at the first " (" and not the last one.
	payload, description := value, ""
	if strings.HasSuffix(payload, ")") {
		if idx := strings.Index(payload, " ("); idx >= 0 {
			description = payload[idx+2 : len(payload)-1]
			payload = payload[:idx]
		}
	}

	tags := map[string]string{
		"device": device,
		"sensor": sensor,
		"type":   sensorType,
	}
	if description != "" {
		tags["description"] = description
	}

	// print_sensor() prints "unknown" for any sensor flagged SENSOR_FUNKNOWN,
	// whatever its type, and for a drive state outside SENSOR_DRIVE_*. It must
	// therefore be handled per type instead of ahead of the type switch.
	fields := make(map[string]any, 4)
	switch sensorType {
	case "drive":
		// Drive state strings as printed by print_sensor() and the matching
		// SENSOR_DRIVE_* values from sys/sys/sensors.h. value is emitted as
		// well as state because Prometheus cannot store string-only metrics.
		// "unknown" is kept as a state, so a drive that cannot be read stays
		// visible.
		fields["state"] = payload
		switch payload {
		case "empty":
			fields["value"] = float64(1)
		case "ready":
			fields["value"] = float64(2)
		case "powering up":
			fields["value"] = float64(3)
		case "online":
			fields["value"] = float64(4)
		case "idle":
			fields["value"] = float64(5)
		case "active":
			fields["value"] = float64(6)
		case "rebuilding":
			fields["value"] = float64(7)
		case "powering down":
			fields["value"] = float64(8)
		case "failed":
			fields["value"] = float64(9)
		case "degraded":
			fields["value"] = float64(10)
		}
	case "indicator":
		// Boolean indicator printed as "On" or "Off"
		fields["state"] = payload
		switch payload {
		case "On":
			fields["value"] = true
		case "Off":
			fields["value"] = false
		}
	default:
		if payload != "unknown" {
			// Numeric value optionally followed by a unit, e.g. "43.00 degC".
			// Percent and humidity sensors print the unit attached to the
			// number instead, e.g. "49.50%".
			number, unit, _ := strings.Cut(payload, " ")
			if strings.HasSuffix(number, "%") {
				number, unit = strings.TrimSuffix(number, "%"), "%"
			}
			v, err := strconv.ParseFloat(number, 64)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot parse value %q of sensor %q: %w", payload, name, err)
			}
			fields["value"] = v
			if unit != "" { // generic integers print no unit
				tags["unit"] = unit
			}
		}
	}

	if status != "" {
		// status_code is emitted as well because Prometheus drops string-only
		// fields, leaving no way to alert on the status otherwise.
		fields["status"] = status
		fields["status_code"] = sensorStatuses[status]
	}
	if len(fields) == 0 {
		// Unknown value without status: nothing numeric or named to report
		return nil, nil, nil
	}

	return tags, fields, nil
}
