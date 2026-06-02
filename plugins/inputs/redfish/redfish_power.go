package redfish

import (
	"encoding/json"
	"strconv"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/influxdata/telegraf"
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

	for j := range power.PowerSupplies {
		tags := make(map[string]string, 19)
		tags["member_id"] = power.PowerSupplies[j].MemberID
		tags["address"] = address
		tags["name"] = power.PowerSupplies[j].Name
		tags["source"] = system.HostName
		tags["state"] = string(power.PowerSupplies[j].Status.State)
		tags["serial_num"] = power.PowerSupplies[j].SerialNumber
		tags["health"] = string(power.PowerSupplies[j].Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["power_input_watts"] = power.PowerSupplies[j].PowerInputWatts
		fields["power_output_watts"] = power.PowerSupplies[j].PowerOutputWatts
		fields["line_input_voltage"] = power.PowerSupplies[j].LineInputVoltage
		fields["last_power_output_watts"] = power.PowerSupplies[j].LastPowerOutputWatts
		fields["power_capacity_watts"] = power.PowerSupplies[j].PowerCapacityWatts
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

func (r *Redfish) gatherPowerSubsysMetrics(
	acc telegraf.Accumulator,
	address string,
	system *schemas.ComputerSystem,
	powerSubsys *schemas.PowerSubsystem,
	chassis *schemas.Chassis) error {
	for _, redundGroup := range powerSubsys.PowerSupplyRedundancy {
		tags := map[string]string{
			"name":    redundGroup.GroupName,
			"address": address,
			"source":  system.HostName,
			"type":    string(redundGroup.RedundancyType),
			"health":  string(redundGroup.Status.Health),
			"state":   string(redundGroup.Status.State),
		}
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := map[string]interface{}{
			"redund_group_count": redundGroup.RedundancyGroupCount,
		}

		acc.AddFields("redfish_powersubsys_redundancy", fields, tags)
	}

	psu, err := powerSubsys.PowerSupplies()
	if err != nil || psu == nil {
		return err
	}

	for _, j := range psu {
		// Due to Gofish not having implemented a PowerSupplyUnit getter (PowerSupplies() is for the old API)
		// this manual parsing is required.
		// The Type definitions exist thou, since they are generated from the official Standard
		powerSupply := schemas.PowerSupplyUnit{}
		err = json.Unmarshal(j.RawData, &powerSupply)
		if err != nil {
			return err
		}

		powerSupply.SetClient(powerSubsys.GetClient())
		psuMetrics, err := powerSupply.Metrics()
		if err != nil {
			return err
		}

		tags := make(map[string]string, 19)
		tags["address"] = address
		tags["name"] = powerSupply.Name
		tags["source"] = system.HostName
		tags["state"] = string(powerSupply.Status.State)
		tags["serial_num"] = powerSupply.SerialNumber
		tags["hotpluggable"] = strconv.FormatBool(powerSupply.HotPluggable)
		tags["health"] = string(powerSupply.Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["power_input_watts"] = psuMetrics.InputPowerWatts.Reading
		fields["power_output_watts"] = psuMetrics.OutputPowerWatts.Reading
		fields["line_input_voltage"] = psuMetrics.InputVoltage.Reading
		fields["power_capacity_watts"] = powerSupply.PowerCapacityWatts
		fields["firmware_version"] = powerSupply.FirmwareVersion
		acc.AddFields("redfish_powersubsys_powersupplies", fields, tags)
	}

	return nil
}
