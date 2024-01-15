//go:build linux
// +build linux

package temp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
)

const scalingFactor = float64(1000.0)

type TemperatureStat struct {
	Name        string
	Label       string
	Device      string
	Temperature float64
	Additional  map[string]interface{}
}

func (t *Temperature) Init() error {
	switch t.MetricFormat {
	case "":
		t.MetricFormat = "v2"
	case "v1", "v2":
		// Do nothing as those are valid
	default:
		return fmt.Errorf("invalid 'metric_format' %q", t.MetricFormat)
	}
	return nil
}

func (t *Temperature) Gather(acc telegraf.Accumulator) error {
	// Get all sensors and honor the HOST_SYS environment variable
	path := os.Getenv("HOST_SYS")
	if path == "" {
		path = "/sys"
	}

	// Try to use the hwmon interface
	temperatures, err := t.gatherHwmon(path)
	if err != nil {
		return fmt.Errorf("getting temperatures failed: %w", err)
	}

	if len(temperatures) == 0 {
		// There is no hwmon interface, fallback to thermal-zone parsing
		temperatures, err = t.gatherThermalZone(path)
		if err != nil {
			return fmt.Errorf("getting temperatures (via fallback) failed: %w", err)
		}
	}

	for _, temp := range temperatures {
		acc.AddFields(
			"temp",
			map[string]interface{}{"temp": temp.Temperature},
			t.getTagsForTemperature(temp, "_input"),
		)

		for measurement, value := range temp.Additional {
			fieldname := "temp"
			if measurement == "alarm" {
				fieldname = "active"
			}
			acc.AddFields(
				"temp",
				map[string]interface{}{fieldname: value},
				t.getTagsForTemperature(temp, "_"+measurement),
			)
		}
	}
	return nil
}

func (t *Temperature) gatherHwmon(syspath string) ([]TemperatureStat, error) {
	// Get all hwmon devices
	sensors, err := filepath.Glob(filepath.Join(syspath, "class", "hwmon", "hwmon*", "temp*_input"))
	if err != nil {
		return nil, fmt.Errorf("getting sensors failed: %w", err)
	}

	// Handle CentOS special path containing an additional "device" directory
	// see https://github.com/shirou/gopsutil/blob/master/host/host_linux.go
	if len(sensors) == 0 {
		sensors, err = filepath.Glob(filepath.Join(syspath, "class", "hwmon", "hwmon*", "device", "temp*_input"))
		if err != nil {
			return nil, fmt.Errorf("getting sensors on CentOS failed: %w", err)
		}
	}

	// Exit early if we cannot find any device
	if len(sensors) == 0 {
		return nil, nil
	}

	// Collect the sensor information
	stats := make([]TemperatureStat, 0, len(sensors))
	for _, s := range sensors {
		// Get the sensor directory and the temperature prefix from the path
		path := filepath.Dir(s)
		prefix := strings.SplitN(filepath.Base(s), "_", 2)[0]

		// Read the device name and fallback to the device name if we cannot get a sensible name.
		name, deviceName := getNameForPath(path)

		// Get the sensor label
		var label string
		if buf, err := os.ReadFile(filepath.Join(path, prefix+"_label")); err == nil {
			label = strings.TrimSpace(string(buf))
		}

		// Do the actual sensor readings
		temp := TemperatureStat{
			Name:       name,
			Label:      strings.ToLower(label),
			Device:     deviceName,
			Additional: make(map[string]interface{}),
		}

		// Temperature (mandatory)
		fn := filepath.Join(path, prefix+"_input")
		buf, err := os.ReadFile(fn)
		if err != nil {
			t.Log.Warnf("Couldn't read temperature from %q: %v", fn, err)
			continue
		}
		if v, err := strconv.ParseFloat(strings.TrimSpace(string(buf)), 64); err == nil {
			temp.Temperature = v / scalingFactor
		}

		// Alarm (optional)
		fn = filepath.Join(path, prefix+"_alarm")
		buf, err = os.ReadFile(fn)
		if err == nil {
			if a, err := strconv.ParseBool(strings.TrimSpace(string(buf))); err == nil {
				temp.Additional["alarm"] = a
			}
		}

		// Read all possible values of the sensor
		matches, err := filepath.Glob(filepath.Join(path, prefix+"_*"))
		if err != nil {
			t.Log.Warnf("Couldn't read files from %q: %v", filepath.Join(path, prefix+"_*"), err)
			continue
		}
		for _, fn := range matches {
			buf, err = os.ReadFile(fn)
			if err != nil {
				continue
			}
			parts := strings.SplitN(filepath.Base(fn), "_", 2)
			if len(parts) != 2 {
				continue
			}
			measurement := parts[1]

			// Skip already added values
			switch measurement {
			case "label", "input", "alarm":
				continue
			}

			v, err := strconv.ParseFloat(strings.TrimSpace(string(buf)), 64)
			if err != nil {
				continue
			}
			temp.Additional[measurement] = v / scalingFactor
		}

		stats = append(stats, temp)
	}

	return stats, nil
}

func (t *Temperature) gatherThermalZone(path string) ([]TemperatureStat, error) {
	return nil, errors.New("not implemented")
}

func (t *Temperature) getTagsForTemperature(temp TemperatureStat, suffix string) map[string]string {
	var sensor string
	switch t.MetricFormat {
	case "v1":
		sensor = temp.Name + "_" + strings.ReplaceAll(temp.Label, " ", "") + suffix
	case "v2":
		sensor = temp.Name + "_" + strings.ReplaceAll(temp.Label, " ", "_") + suffix
	}

	tags := map[string]string{"sensor": sensor}
	if t.DeviceTag {
		tags["device"] = temp.Device
	}
	return tags
}

func getNameForPath(path string) (string, string) {
	// Try to read the device link for fallback
	deviceName, err := os.Readlink(filepath.Join(path, "device"))
	if err == nil {
		deviceName = filepath.Base(deviceName)
	}

	// Read the device name and fallback to the device name if we cannot
	// get a sensible name.
	n, err := os.ReadFile(filepath.Join(path, "name"))
	if err != nil {
		return "", deviceName
	}
	return strings.TrimSpace(string(n)), deviceName
}
