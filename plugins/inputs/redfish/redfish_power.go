package redfish

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/influxdata/telegraf"
)

func (r *Redfish) gatherPower(acc telegraf.Accumulator, system *schemas.ComputerSystem, chassis *schemas.Chassis, chassisTags map[string]string) error {
	power, err := chassis.Power()
	if err != nil {
		return fmt.Errorf("parsing power data from %q failed: %w", r.host, err)
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
			"address":   r.host,
			"name":      j.Name,
			"source":    system.HostName,
		}
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			// Location.PostalAddress.DataCenter is nowhere in the redfish standard
			// We do some manual parsing in order to not break existing code
			var datacenter datacenterTag
			// Try to parse the datacenter from the raw data. As this data is not defined
			// in the standard, it might not exist in the raw data and parsing may fail. In this
			// case we will get an "" for the datacenter which is perfectly fine, so ignore the
			// marshaling error.
			//nolint:errcheck // See above
			json.Unmarshal(chassis.RawData, &datacenter)

			tags["datacenter"] = datacenter.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			maps.Copy(tags, chassisTags)
		}

		fields := map[string]any{
			"power_allocated_watts":  toFloat64(j.PowerAllocatedWatts),
			"power_available_watts":  toFloat64(j.PowerAvailableWatts),
			"power_capacity_watts":   toFloat64(j.PowerCapacityWatts),
			"power_consumed_watts":   toFloat64(j.PowerConsumedWatts),
			"power_requested_watts":  j.PowerRequestedWatts,
			"average_consumed_watts": toFloat64(j.PowerMetrics.AverageConsumedWatts),
			"max_consumed_watts":     toFloat64(j.PowerMetrics.MaxConsumedWatts),
			"min_consumed_watts":     toFloat64(j.PowerMetrics.MinConsumedWatts),
		}
		if j.PowerMetrics.IntervalInMin != nil {
			fields["interval_in_min"] = int64(*j.PowerMetrics.IntervalInMin)
		}

		acc.AddFields("redfish_power_powercontrol", fields, tags)
	}

	// each iteration would copy a large ammount of data, therefore we use indexing
	for i := range len(power.PowerSupplies) {
		tags := make(map[string]string, 19)
		tags["member_id"] = power.PowerSupplies[i].MemberID
		tags["address"] = r.host
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

			tags["datacenter"] = datacenter.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			maps.Copy(tags, chassisTags)
		}

		fields := make(map[string]any)
		fields["power_input_watts"] = toFloat64(power.PowerSupplies[i].PowerInputWatts)
		fields["power_output_watts"] = toFloat64(power.PowerSupplies[i].PowerOutputWatts)
		fields["line_input_voltage"] = toFloat64(power.PowerSupplies[i].LineInputVoltage)
		fields["last_power_output_watts"] = toFloat64(power.PowerSupplies[i].LastPowerOutputWatts)
		fields["power_capacity_watts"] = toFloat64(power.PowerSupplies[i].PowerCapacityWatts)
		acc.AddFields("redfish_power_powersupplies", fields, tags)
	}

	for _, j := range power.Voltages {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = r.host
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
			maps.Copy(tags, chassisTags)
		}

		fields := make(map[string]any)
		fields["reading_volts"] = toFloat64(j.ReadingVolts)
		fields["upper_threshold_critical"] = toFloat64(j.UpperThresholdCritical)
		fields["upper_threshold_fatal"] = toFloat64(j.UpperThresholdFatal)
		fields["lower_threshold_critical"] = toFloat64(j.LowerThresholdCritical)
		fields["lower_threshold_fatal"] = toFloat64(j.LowerThresholdFatal)
		acc.AddFields("redfish_power_voltages", fields, tags)
	}

	return nil
}

// gofish decodes the power values as float32, so convert them with float32
// precision to get 229.9 instead of 229.89999389648438
func toFloat64(v *float32) any {
	if v == nil {
		return nil
	}
	f, err := strconv.ParseFloat(strconv.FormatFloat(float64(*v), 'g', -1, 32), 64)
	if err != nil {
		return nil
	}
	return f
}
