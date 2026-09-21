package redfish

import (
	"encoding/json"
	"fmt"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/influxdata/telegraf"
)

func (r *Redfish) gatherPower(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	return r.gatherPowerMetrics(acc, address, system, chassis)
}

func (r *Redfish) gatherPowerMetrics(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	power, err := chassis.Power()
	if err != nil {
		return fmt.Errorf("parsing power data from %q failed: %w", address, err)
	}

	// power is nil when the legacy power endpoints are not available
	// The newer subsys endpoints should be used as a form of retry
	if power == nil {
		r.Log.Warnf("Skipping power data of chassis %q. Only the legacy power API is supported at the moment", chassis.ID)
		return nil
	}

	for _, j := range power.PowerControl {
		tags := map[string]string{
			"member_id": j.MemberID,
			"address":   address,
			"name":      j.Name,
			"source":    system.HostName,
		}
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// Location.PostalAddress.DataCenter is nowhere in the redfish standard
			// We do some manual parsing in order to not break existing code
			var datacenter datacenterTag
			//nolint:errcheck // Ignore if the marshalling fails as this datapoint should not exist
			json.Unmarshal(chassis.RawData, &datacenter)

			tags["datacenter"] = datacenter.Location.PostalAddress.DataCenter
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
			"average_consumed_watts": j.PowerMetrics.AverageConsumedWatts,
			"interval_in_min":        int64(*j.PowerMetrics.IntervalInMin),
			"max_consumed_watts":     j.PowerMetrics.MaxConsumedWatts,
			"min_consumed_watts":     j.PowerMetrics.MinConsumedWatts,
		}

		acc.AddFields("redfish_power_powercontrol", fields, tags)
	}

	for i := range len(power.PowerSupplies) {
		tags := make(map[string]string, 19)
		tags["member_id"] = power.PowerSupplies[i].MemberID
		tags["address"] = address
		tags["name"] = power.PowerSupplies[i].Name
		tags["source"] = system.HostName
		tags["state"] = string(power.PowerSupplies[i].Status.State)
		tags["health"] = string(power.PowerSupplies[i].Status.Health)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// Location.PostalAddress.DataCenter is nowhere in the redfish standard
			// We do some manual parsing in order to not break existing code
			var datacenter datacenterTag
			//nolint:errcheck // Ignore if the marshalling fails as this datapoint should not exist
			json.Unmarshal(chassis.RawData, &datacenter)
			r.Log.Warnf("The datacenter tag will be removed in a future version as it does not conform to DTMF's standard.")

			tags["datacenter"] = datacenter.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		fields := make(map[string]interface{})
		fields["power_input_watts"] = power.PowerSupplies[i].PowerInputWatts
		fields["power_output_watts"] = power.PowerSupplies[i].PowerOutputWatts
		fields["line_input_voltage"] = power.PowerSupplies[i].LineInputVoltage
		fields["last_power_output_watts"] = power.PowerSupplies[i].LastPowerOutputWatts
		fields["power_capacity_watts"] = power.PowerSupplies[i].PowerCapacityWatts
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
			// Location.PostalAddress.DataCenter is nowhere in the redfish standard
			// We do some manual parsing in order to not break existing code
			var datacenter datacenterTag
			//nolint:errcheck // Ignore if the marshalling fails as this datapoint should not exist
			json.Unmarshal(chassis.RawData, &datacenter)

			tags["datacenter"] = datacenter.Location.PostalAddress.DataCenter
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
