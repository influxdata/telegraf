package systemd_timings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SystemdTimings is a telegraf plugin to gather systemd boot timing metrics.
type SystemdTimings struct {
}

// Measurement name.
const measurement = "systemd_timings"

// Record if we've already posted the system wide boot timestamps.
var postedBootTimestamps = false

// Map of a system wide boot metrics to their timestamps in microseconds, see:
// https://www.freedesktop.org/wiki/Software/systemd/dbus/ for more details.
var managerProps = map[string]string{
	"FirmwareTimestampMonotonic":               "",
	"LoaderTimestampMonotonic":                 "",
	"InitRDTimestampMonotonic":                 "",
	"UserspaceTimestampMonotonic":              "",
	"FinishTimestampMonotonic":                 "",
	"SecurityStartTimestampMonotonic":          "",
	"SecurityFinishTimestampMonotonic":         "",
	"GeneratorsStartTimestampMonotonic":        "",
	"GeneratorsFinishTimestampMonotonic":       "",
	"UnitsLoadStartTimestampMonotonic":         "",
	"UnitsLoadFinishTimestampMonotonic":        "",
	"InitRDSecurityStartTimestampMonotonic":    "",
	"InitRDSecurityFinishTimestampMonotonic":   "",
	"InitRDGeneratorsStartTimestampMonotonic":  "",
	"InitRDGeneratorsFinishTimestampMonotonic": "",
	"InitRDUnitsLoadStartTimestampMonotonic":   "",
	"InitRDUnitsLoadFinishTimestampMonotonic":  "",
}

// Group unit timestamps.
type unitTimestamps struct {
	Activating   uint64
	Activated    uint64
	Deactivating uint64
	Deactivated  uint64
}

// Keep track of the units previous timestamps, we only push if they have
// have changed against the previous set.
var allUnitTimestamps = map[string]*unitTimestamps{}

// stripType removes the dbus type from the string str to return only the value.
// See https://www.alteeve.com/w/List_of_DBus_data_types for dbus type
// information.
func stripType(str string) string {
	return strings.Split(str, " ")[1]
}

// getManagerProp retrieves the property value with name propName.
func getManagerProp(dbusConn *dbus.Conn, propName string) (string, error) {
	prop, err := dbusConn.GetManagerProperty(propName)
	if err != nil {
		return "", err
	}

	return stripType(prop), nil
}

// bootIsFinished returns true if systemd has completed all unit initialization.
func bootIsFinished() bool {
	// Connect to the systemd dbus.
	dbusConn, err := dbus.NewSystemConnection()
	if err != nil {
		return false
	}

	defer dbusConn.Close()

	// Read the "FinishTimestampMonotonic" manager property, this will be
	// non-zero if the system has finished initialization.
	progressStr, err := getManagerProp(dbusConn, "FinishTimestampMonotonic")
	if err != nil {
		return false
	}

	// Convert to an int for comparison.
	progressVal, err := strconv.ParseInt(progressStr, 10, 32)
	if err != nil {
		return false
	}

	return progressVal != 0
}

// postAllManagerProps reads all systemd manager properties and sends them to
// telegraf.
func postAllManagerProps(dbusConn *dbus.Conn, acc telegraf.Accumulator) error {

	// Read all properties and send non zero values to telegraf.
	for name := range managerProps {
		propVal, err := getManagerProp(dbusConn, name)
		if err != nil {
			continue
		} else {
			// Save since we might need the value later when computing per unit
			// time deltas.
			managerProps[name] = propVal
			if propVal == "" || propVal == "0" {
				// Skip zero valued properties, these indicate unset properties
				// in systemd.
				continue
			}

			value, err := strconv.ParseUint(propVal, 10, 64)
			if err != nil {
				acc.AddError(err)
				continue
			}

			// Build field and tag maps.
			tags := map[string]string{"SystemTimestamp": name}

			fields := map[string]interface{}{"SystemTimestampValue": value}

			// Send to telegraf.
			acc.AddFields(measurement, fields, tags)
		}
	}

	return nil
}

// query dbus to access unit startup timing data, all time measurements here
// are measured in microseconds.
func getUnitTimingData(dbusConn *dbus.Conn,
	unitName string,
	userSpaceStart uint64) (uint64, uint64, uint64, uint64, uint64, error) {

	// Retrieve all timing properties for this unit.
	activatingProp, err := dbusConn.GetUnitProperty(unitName,
		"InactiveExitTimestampMonotonic")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	activatedProp, err := dbusConn.GetUnitProperty(unitName,
		"ActiveEnterTimestampMonotonic")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	deactivatingProp, err := dbusConn.GetUnitProperty(unitName,
		"ActiveExitTimestampMonotonic")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	deactivatedProp, err := dbusConn.GetUnitProperty(unitName,
		"InactiveEnterTimestampMonotonic")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	// Convert all to uint64 types and subtract the user space start time
	// stamp to give us relative startup times.
	activating, err := strconv.ParseUint(
		stripType(activatingProp.Value.String()), 10, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	activated, err := strconv.ParseUint(
		stripType(activatedProp.Value.String()), 10, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	deactivating, err := strconv.ParseUint(
		stripType(deactivatingProp.Value.String()), 10, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	deactivated, err := strconv.ParseUint(
		stripType(deactivatedProp.Value.String()), 10, 64)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	if activating > 0 {
		activating -= userSpaceStart
	}

	if activated > 0 {
		activated -= userSpaceStart
	}

	if deactivating > 0 {
		deactivating -= userSpaceStart
	}

	if deactivated > 0 {
		deactivated -= userSpaceStart
	}

	runtime := uint64(0)
	if activated >= activating {
		runtime = activated - activating
	} else if deactivated >= activating {
		runtime = deactivated - activating
	}

	// Return the timing data for this unit, converted to seconds.
	return activating, activated, deactivating, deactivated, runtime, nil
}

// postAllUnitTimingData
func postAllUnitTimingData(dbusConn *dbus.Conn, acc telegraf.Accumulator) error {
	statusList, err := dbusConn.ListUnits()
	if err != nil {
		acc.AddError(err)
		return err
	}

	// Get the user space start timestamp so we can subtract it from all
	// unit timestamps to give us a relative offset from user space start.
	userTs, found := managerProps["UserspaceTimestampMonotonic"]
	if !found {
		return fmt.Errorf(`UserspaceTimestampMonotonic not found, cannot
						  compute unit timestamps`)
	}

	// Convert UserspaceTimestampMonotonic to a uint64
	userStartTs, err := strconv.ParseUint(userTs, 10, 64)
	if err != nil {
		acc.AddError(err)
		return err
	}

	// For each unit query timing data, don't stop on failure.
	for _, unitStatus := range statusList {
		activating, activated, deactivating, deactivated, runtime, err :=
			getUnitTimingData(dbusConn, unitStatus.Name, userStartTs)
		if err != nil {
			acc.AddError(err)
		} else {
			if runtime == 0 {
				// Don't post results for services which were never started
				// or stopped.
				continue
			}

			// Only post these to telegraf if they are different that the
			// last time they were posted.  This ensures that only restarted
			// or manually stopped services would have new metrics posted.
			entry, exists := allUnitTimestamps[unitStatus.Name]
			if exists {
				if entry.Activating == activating &&
					entry.Activated == activated &&
					entry.Deactivating == deactivating &&
					entry.Deactivated == deactivated {
					// No change since the last time we collected, so don't
					// post to telegraf.
					continue
				}
			} else {
				entry = new(unitTimestamps)
			}

			entry.Activating = activating
			entry.Activated = activated
			entry.Deactivating = deactivating
			entry.Deactivated = deactivated
			allUnitTimestamps[unitStatus.Name] = entry

			// These are per unit wide timestamps, so tag them as such.
			tags := map[string]string{"UnitName": unitStatus.Name}

			// Construct fields map.
			fields := map[string]interface{}{
				"ActivatingTimestamp":   activating,
				"ActivatedTimestamp":    activated,
				"DeactivatingTimestamp": deactivating,
				"DeactivatedTimestamp":  deactivated,
				"RunDuration":           runtime,
			}

			// Send to telegraf.
			acc.AddFields(measurement, fields, tags)
		}
	}

	return nil
}

// Description returns a short description of the plugin
func (s *SystemdTimings) Description() string {
	return "Gather systemd boot and unit timing data"
}

// SampleConfig returns sample configuration options.
func (s *SystemdTimings) SampleConfig() string {
	return ""
}

// Gather reads timestamp metrics from systemd via dbus and sends them to
// telegraf.
func (s *SystemdTimings) Gather(acc telegraf.Accumulator) error {
	if !bootIsFinished() {
		// We are not ready to collect yet, telegraf will call us later to try
		// again.
		return nil
	}

	// Connect to the systemd dbus.
	dbusConn, err := dbus.NewSystemConnection()
	if err != nil {
		return err
	}

	defer dbusConn.Close()

	// Only read system wide "manager" properties once per telegraf lifetime.
	if postedBootTimestamps == false {
		err = postAllManagerProps(dbusConn, acc)
		if err != nil {
			acc.AddError(err)
			return err
		}

		postedBootTimestamps = true
	}

	// Read all unit timing data, this will only post metrics which have changed
	// value.
	err = postAllUnitTimingData(dbusConn, acc)
	if err != nil {
		acc.AddError(err)
		return err
	}

	return nil
}

func init() {
	inputs.Add("systemd_timings", func() telegraf.Input {
		return &SystemdTimings{}
	})
}
