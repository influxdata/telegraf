package systemd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/coreos/go-systemd/dbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type dbusConn interface {
	GetUnitProperty(string, string) (*dbus.Property, error)
	GetUnitTypeProperty(string, string, string) (*dbus.Property, error)
}

type SystemD struct {
	UnitPatterns []string
}

func (_ *SystemD) Description() string {
	return "Read unit metrics of systemd using dbus"
}

const sampleConfig = `
  ## List of unit regex pattern
  unit_patterns = [".*"]
`

func (_ *SystemD) SampleConfig() string {
	return sampleConfig
}

func (s *SystemD) Gather(acc telegraf.Accumulator) error {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		return fmt.Errorf("failed to get dbus connection: %s", err)
	}
	defer conn.Close()

	units, err := conn.ListUnits()
	if err != nil {
		return err
	}

	units = s.filterUnits(acc, units)

	for _, unit := range units {
		tags := map[string]string{
			"unitName": unit.Name,
			"unitType": "",
		}
		fields := map[string]interface{}{}

		collectActiveState(unit, conn, fields)

		if strings.HasSuffix(unit.Name, ".timer") {
			collectTimerUnit(unit, conn, tags, fields)
		} else if strings.HasSuffix(unit.Name, ".service") {
			collectServiceUnit(unit, conn, tags, fields)
		} else if strings.HasSuffix(unit.Name, ".socket") {
			collectSocketUnit(unit, conn, tags, fields)
		} else {
			continue
		}

		acc.AddFields("systemd", fields, tags)
	}

	return nil
}

func (s *SystemD) filterUnits(acc telegraf.Accumulator, units []dbus.UnitStatus) []dbus.UnitStatus {
	filtered := []dbus.UnitStatus{}
	for _, unit := range units {
		for _, pattern := range s.UnitPatterns {
			matched, err := regexp.MatchString(pattern, unit.Name)
			if err != nil {
				acc.AddError(err)
				continue
			}
			if matched {
				filtered = append(filtered, unit)
			}
		}
	}
	return filtered
}

func collectActiveState(unit dbus.UnitStatus, conn dbusConn, fields map[string]interface{}) {
	fields["isActive"] = 0
	fields["activeEnterTimestamp"] = 0

	if unit.ActiveState == "active" {
		fields["isActive"] = 1

		timestampValue, err := conn.GetUnitProperty(unit.Name, "ActiveEnterTimestamp")
		if err == nil {
			fields["activeEnterTimestamp"] = timestampValue.Value.Value().(uint64)
		}
	}
}

func collectTimerUnit(unit dbus.UnitStatus, conn dbusConn, tags map[string]string, fields map[string]interface{}) {
	tags["unitType"] = "Timer"

	lastTriggerValue, err := conn.GetUnitTypeProperty(unit.Name, "Timer", "LastTriggerUSec")
	if err == nil {
		fields["lastTriggerUSec"] = lastTriggerValue.Value.Value().(uint64)
	}
}

func collectServiceUnit(unit dbus.UnitStatus, conn dbusConn, tags map[string]string, fields map[string]interface{}) {
	tags["unitType"] = "Service"

	restartsCount, err := conn.GetUnitTypeProperty(unit.Name, "Service", "NRestarts")
	if err == nil {
		fields["nRestarts"] = restartsCount.Value.Value().(uint32)
	}
}

func collectSocketUnit(unit dbus.UnitStatus, conn dbusConn, tags map[string]string, fields map[string]interface{}) {
	tags["unitType"] = "Socket"

	acceptedConnectionCount, err := conn.GetUnitTypeProperty(unit.Name, "Socket", "NAccepted")
	if err == nil {
		fields["nAccepted"] = acceptedConnectionCount.Value.Value().(uint32)
	}

	currentConnectionCount, err := conn.GetUnitTypeProperty(unit.Name, "Socket", "NConnection")
	if err == nil {
		fields["nConnection"] = currentConnectionCount.Value.Value().(uint32)
	}

	refusedConnectionCount, err := conn.GetUnitTypeProperty(unit.Name, "Socket", "NRefused")
	if err == nil {
		fields["nRefused"] = refusedConnectionCount.Value.Value().(uint32)
	}
}

func init() {
	inputs.Add("systemd", func() telegraf.Input {
		return &SystemD{}
	})
}
