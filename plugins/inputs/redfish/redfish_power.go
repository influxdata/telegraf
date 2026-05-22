package redfish

import (
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/stmcginnis/gofish/schemas"
)

func (r *Redfish) gatherPower(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	powerSubsys, err := chassis.PowerSubsystem()
	if err != nil {
		return err
	}

	// The redfish version is not an indicator as to which of these api's has been implemented
	// We use the old endpoints only as a fallback as to not generate duplicates
	if powerSubsys == nil {
		// Gather metrics via the legacy api
		err = r.gatherPowerMetrics(acc, address, system, chassis)
	} else {
		// Gather metrics via the current thermal subsys api
		err = r.gatherPowerSubsysMetrics(acc, address, system, powerSubsys, chassis)
	}

	return err
}

func (r *Redfish) gatherPowerMetrics(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	power, err := chassis.Power()
	if err != nil || power == nil {
		return err
	}

	for _, j := range power.PowerControl {
		tags := map[string]string{
			"member_id": j.MemberID,
			"address":   address,
			"name":      j.Name,
			"source":    system.HostName,
		}
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := map[string]interface{}{
			"power_allocated_watts":  j.PowerAllocatedWatts,
			"power_available_watts":  j.PowerAvailableWatts,
			"power_capacity_watts":   j.PowerCapacityWatts,
			"power_consumed_watts":   j.PowerConsumedWatts,
			"power_requested_watts":  j.PowerRequestedWatts,
			"average_consumed_watts": float64(*j.PowerMetrics.AverageConsumedWatts),
			"interval_in_min":        int64(*j.PowerMetrics.IntervalInMin),
			"max_consumed_watts":     j.PowerMetrics.MaxConsumedWatts,
			"min_consumed_watts":     j.PowerMetrics.MinConsumedWatts,
		}

		acc.AddFields("redfish_power_powercontrol", fields, tags)
	}

	for _, j := range power.PowerSupplies {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.HostName
		tags["state"] = string(j.Status.State)
		tags["serial_num"] = j.SerialNumber
		tags["health"] = string(j.Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["power_input_watts"] = j.PowerInputWatts
		fields["power_output_watts"] = j.PowerOutputWatts
		fields["line_input_voltage"] = j.LineInputVoltage
		fields["last_power_output_watts"] = j.LastPowerOutputWatts
		fields["power_capacity_watts"] = j.PowerCapacityWatts
		acc.AddFields("redfish_power_powersupplies", fields, tags)
	}

	for _, j := range power.Voltages {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.HostName
		tags["state"] = string(j.Status.State)
		tags["health"] = string(j.Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["reading_volts"] = j.ReadingVolts
		fields["upper_threshold_critical"] = j.UpperThresholdCritical
		fields["upper_threshold_fatal"] = j.UpperThresholdFatal
		fields["lower_threshold_critical"] = j.LowerThresholdCritical
		fields["lower_threshold_fatal"] = j.LowerThresholdFatal
		acc.AddFields("redfish_power_voltages", fields, tags)
	}

	return nil
}

func (r *Redfish) gatherPowerSubsysMetrics(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, powerSubsys *schemas.PowerSubsystem, chassis *schemas.Chassis) error {

	for _, redundGroup := range powerSubsys.PowerSupplyRedundancy {
		tags := map[string]string{
			"name":   redundGroup.GroupName,
			"source": system.HostName,
		}
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := map[string]interface{}{
			"type":   redundGroup.RedundancyType,
			"health": redundGroup.Status.Health,
			"state":  redundGroup.Status.State,
		}

		acc.AddFields("redfish_powersubsys_redundancy", fields, tags)
	}

	psu, err := powerSubsys.PowerSupplies()
	if err != nil || psu == nil {
		return err
	}

	// Contains Voltage and wattage info
	for _, j := range psu {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.HostName
		tags["state"] = string(j.Status.State)
		tags["serial_num"] = j.SerialNumber
		tags["hotpluggable"] = strconv.FormatBool(j.HotPluggable)
		tags["line_input_voltage_type"] = string(j.LineInputVoltageType)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["health"] = j.Status.Health
		fields["power_input_watts"] = j.PowerInputWatts
		fields["power_output_watts"] = j.PowerOutputWatts
		fields["line_input_voltage"] = j.LineInputVoltage
		fields["last_power_output_watts"] = j.LastPowerOutputWatts
		fields["power_capacity_watts"] = j.PowerCapacityWatts
		fields["firmware_version"] = j.FirmwareVersion
		acc.AddFields("redfish_powersubsys_powersupplies", fields, tags)
	}

	return nil
}
