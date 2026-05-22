package redfish

import (
	"github.com/influxdata/telegraf"
	"github.com/stmcginnis/gofish/schemas"
)

func (r *Redfish) gatherThermal(acc telegraf.Accumulator, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {

	thermalSubsys, err := chassis.ThermalSubsystem()
	if err != nil {
		return err
	}

	// The redfish version is not an indicator as to which of these api's has been implemented
	// We use the old endpoints only as a fallback as to not generate duplicates
	if thermalSubsys == nil {
		// Gather metrics via the legacy api
		r.gatherThermalMetrics(acc, system, chassis)
	} else {
		// Gather metrics via the current thermal subsys api
		r.gatherThermalSubsysMetrics(acc, system, thermalSubsys, chassis)
	}

	return nil
}

func (r *Redfish) gatherThermalMetrics(acc telegraf.Accumulator, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	thermal, err := chassis.Thermal()
	if err != nil || thermal == nil {
		return err
	}

	for _, j := range thermal.Temperatures {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		// tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.HostName
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// tags["datacenter"] = chassis.Location.PostalAddress.DataCenter // Not in the standard
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["state"] = j.Status.State
		fields["health"] = j.Status.Health
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
		//tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.HostName
		fields["state"] = j.Status.State
		fields["health"] = j.Status.Health
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
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
			fields["reading_percent"] = j.Reading
		}
		acc.AddFields("redfish_thermal_fans", fields, tags)
	}

	return nil
}

func (r *Redfish) gatherThermalSubsysMetrics(acc telegraf.Accumulator, system *schemas.ComputerSystem, thermalSubsys *schemas.ThermalSubsystem, chassis *schemas.Chassis) error {
	thermalMetrics, _ := thermalSubsys.ThermalMetrics()
	fans, _ := thermalSubsys.Fans()
	for _, j := range thermalMetrics.TemperatureReadingsCelsius {
		tags := make(map[string]string, 14)
		tags["name"] = j.DeviceName
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
		fields["reading_celsius"] = j.Reading
		acc.AddFields("redfish_thermalsubsys_temperatures", fields, tags)
	}

	for _, j := range fans {
		tags := make(map[string]string, 20)
		fields := make(map[string]interface{}, 5)
		tags["member_id"] = j.MemberID
		tags["name"] = j.Name
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
		} else {
			fields["reading_percent"] = j.Reading
		}
		acc.AddFields("redfish_thermalsubsys_fans", fields, tags)
	}

	return nil
}
