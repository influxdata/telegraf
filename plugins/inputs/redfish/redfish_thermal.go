package redfish

import (
	"encoding/json"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/stmcginnis/gofish/schemas"
)

func (r *Redfish) gatherThermal(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	// Gather metrics via the legacy api
	// TODO: Add thermal metric gatering via the new thermalSubsys API
	return r.gatherThermalMetrics(acc, address, system, chassis)
}

func (r *Redfish) gatherThermalMetrics(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	thermal, err := chassis.Thermal()
	if err != nil || thermal == nil {
		return fmt.Errorf("error parsing input from %s: %w", address, err)
	}

	for _, j := range thermal.Temperatures {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.HostName
		tags["state"] = string(j.Status.State)
		tags["health"] = string(j.Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["datacenter"] = "" // Not in the standard, keeping for backward compatibility
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["reading_celsius"] = j.ReadingCelsius
		fields["upper_threshold_critical"] = j.UpperThresholdCritical
		fields["upper_threshold_fatal"] = j.UpperThresholdFatal
		fields["lower_threshold_critical"] = j.LowerThresholdCritical
		fields["lower_threshold_fatal"] = j.LowerThresholdFatal
		acc.AddFields("redfish_thermal_temperatures", fields, tags)
	}

	for _, j := range thermal.Fans {
		tags := make(map[string]string, 20)
		fields := make(map[string]interface{}, 5)
		tags["member_id"] = j.MemberID
		tags["address"] = address

		// FanName is Deprecated but kept around for ilo4 support
		if j.Name == "" {
			tags["name"] = j.FanName
		} else {
			tags["name"] = j.Name
		}
		tags["source"] = system.HostName
		tags["state"] = string(j.Status.State)
		tags["health"] = string(j.Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["datacenter"] = "" // Not in the standard, keeping for backward compatibility
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		// Due to ILO4 not being fully readfish compatible we have to do this parsing manually
		var ilo4ReadingPercent struct {
			CurrentReading *int64
		}
		json.Unmarshal(j.RawData, &ilo4ReadingPercent)

		if ilo4ReadingPercent.CurrentReading != nil {
			fields["reading_percent"] = ilo4ReadingPercent.CurrentReading
		} else {
			if j.ReadingUnits == "RPM" {
				fields["upper_threshold_critical"] = j.UpperThresholdCritical
				fields["upper_threshold_fatal"] = j.UpperThresholdFatal
				fields["lower_threshold_critical"] = j.LowerThresholdCritical
				fields["lower_threshold_fatal"] = j.LowerThresholdFatal
				fields["reading_rpm"] = j.Reading
			} else if j.Reading != nil {
				fields["reading_percent"] = j.Reading
			} else {
				fields["reading_percent"] = j.Reading
			}
		}
		acc.AddFields("redfish_thermal_fans", fields, tags)
	}

	return nil
}
