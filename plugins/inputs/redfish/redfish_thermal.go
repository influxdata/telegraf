package redfish

import (
	"encoding/json"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/influxdata/telegraf"
)

func (r *Redfish) gatherThermal(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	thermalSubsys, err := chassis.ThermalSubsystem()
	if err != nil {
		return err
	}

	// The redfish version is not an indicator as to which of these api's has been implemented
	// We use the old endpoints only as a fallback as to not generate duplicates
	if thermalSubsys == nil {
		// Gather metrics via the legacy api
		err = r.gatherThermalMetrics(acc, address, system, chassis)
	} else {
		// Gather metrics via the current thermal subsys api
		err = r.gatherThermalSubsysMetrics(acc, address, system, thermalSubsys, chassis)
	}

	return err
}

func (r *Redfish) gatherThermalMetrics(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	thermal, err := chassis.Thermal()
	if err != nil || thermal == nil {
		return err
	}

	for _, j := range thermal.Temperatures {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["state"] = string(j.Status.State)
		tags["health"] = string(j.Status.Health)
		tags["name"] = j.Name
		tags["source"] = system.HostName
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
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

	for j := range thermal.Fans {
		tags := make(map[string]string, 20)
		fields := make(map[string]interface{}, 5)
		tags["member_id"] = thermal.Fans[j].MemberID
		tags["address"] = address

		if thermal.Fans[j].FanName != "" { //nolint:staticcheck // used for backwards compatibilty to ilo4
			tags["name"] = thermal.Fans[j].FanName //nolint:staticcheck // used for backwards compatibilty to ilo4
		} else {
			tags["name"] = thermal.Fans[j].Name
		}
		tags["source"] = system.HostName
		tags["state"] = string(thermal.Fans[j].Status.State)
		tags["health"] = string(thermal.Fans[j].Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
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
		err = json.Unmarshal(thermal.Fans[j].RawData, &ilo4ReadingPercent)
		if err != nil {
			return err
		}

		if ilo4ReadingPercent.CurrentReading != nil {
			fields["reading_percent"] = ilo4ReadingPercent.CurrentReading
		} else {
			if thermal.Fans[j].ReadingUnits == "RPM" {
				fields["upper_threshold_critical"] = thermal.Fans[j].UpperThresholdCritical
				fields["upper_threshold_fatal"] = thermal.Fans[j].UpperThresholdFatal
				fields["lower_threshold_critical"] = thermal.Fans[j].LowerThresholdCritical
				fields["lower_threshold_fatal"] = thermal.Fans[j].LowerThresholdFatal
				fields["reading_rpm"] = thermal.Fans[j].Reading
			} else if thermal.Fans[j].Reading != nil {
				fields["reading_percent"] = thermal.Fans[j].Reading
			}
		}
		acc.AddFields("redfish_thermal_fans", fields, tags)
	}

	return nil
}

func (r *Redfish) gatherThermalSubsysMetrics(
	acc telegraf.Accumulator,
	address string,
	system *schemas.ComputerSystem,
	thermalSubsys *schemas.ThermalSubsystem,
	chassis *schemas.Chassis) error {
	thermalMetrics, err := thermalSubsys.ThermalMetrics()
	if err != nil {
		return err
	}

	fans, err := thermalSubsys.Fans()
	if err != nil {
		return err
	}
	for _, j := range thermalMetrics.TemperatureReadingsCelsius {
		tags := make(map[string]string, 14)
		tags["name"] = j.DeviceName
		tags["source"] = system.HostName
		tags["address"] = address
		tags["state"] = string(thermalSubsys.Status.State)
		tags["health_rollup"] = string(thermalSubsys.Status.HealthRollup)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["reading_celsius"] = j.Reading
		acc.AddFields("redfish_thermalsubsys_temperatures", fields, tags)
	}

	for _, j := range fans {
		tags := make(map[string]string, 20)
		fields := make(map[string]interface{}, 5)
		tags["member_id"] = j.ID
		tags["name"] = j.Name
		tags["address"] = address
		tags["source"] = system.HostName
		tags["state"] = string(j.Status.State)
		tags["health"] = string(j.Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		if j.ReadingUnits == "RPM" {
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			fields["lower_threshold_critical"] = j.LowerThresholdCritical
			fields["lower_threshold_fatal"] = j.LowerThresholdFatal
			fields["reading_rpm"] = j.Reading
		} else if j.Reading != nil {
			fields["reading_percent"] = j.Reading
		} else {
			var speedPercentReading struct {
				SpeedPercent struct {
					Reading *float32
				}
			}
			err = json.Unmarshal(j.RawData, &speedPercentReading)
			if err != nil {
				return err
			}
			fields["reading_percent"] = speedPercentReading.SpeedPercent.Reading
		}

		acc.AddFields("redfish_thermalsubsys_fans", fields, tags)
	}

	return nil
}
