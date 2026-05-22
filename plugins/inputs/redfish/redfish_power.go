package redfish

import (
	"net/url"

	"github.com/influxdata/telegraf"
)

type power struct {
	PowerControl []struct {
		Name                string
		MemberID            string
		PowerAllocatedWatts *float64
		PowerAvailableWatts *float64
		PowerCapacityWatts  *float64
		PowerConsumedWatts  *float64
		PowerRequestedWatts *float64
		PowerMetrics        struct {
			AverageConsumedWatts *float64
			IntervalInMin        int
			MaxConsumedWatts     *float64
			MinConsumedWatts     *float64
		}
	}
	PowerSupplies []struct {
		Name                 string
		MemberID             string
		PowerInputWatts      *float64
		PowerCapacityWatts   *float64
		PowerOutputWatts     *float64
		LastPowerOutputWatts *float64
		Status               status
		LineInputVoltage     *float64
		SerialNumber         string
	}
	Voltages []struct {
		Name                   string
		MemberID               string
		ReadingVolts           *float64
		UpperThresholdCritical *float64
		UpperThresholdFatal    *float64
		LowerThresholdCritical *float64
		LowerThresholdFatal    *float64
		Status                 status
	}
}

func (r *Redfish) gatherPower(acc telegraf.Accumulator, address string, system *system, chassis *chassis) error {
	power, err := r.getPower(chassis.Power.Ref)
	if err != nil {
		return err
	}

	for _, j := range power.PowerControl {
		tags := map[string]string{
			"member_id": j.MemberID,
			"address":   address,
			"name":      j.Name,
			"source":    system.Hostname,
		}
		if _, ok := r.tagSet[tagSetChassisLocation]; ok && chassis.Location != nil {
			tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
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
			"interval_in_min":        j.PowerMetrics.IntervalInMin,
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
		tags["source"] = system.Hostname
		tags["state"] = j.Status.State
		tags["serial_num"] = j.SerialNumber
		if _, ok := r.tagSet[tagSetChassisLocation]; ok && chassis.Location != nil {
			tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
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
		acc.AddFields("redfish_power_powersupplies", fields, tags)
	}

	for _, j := range power.Voltages {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.Hostname
		if _, ok := r.tagSet[tagSetChassisLocation]; ok && chassis.Location != nil {
			tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
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
		fields["reading_volts"] = j.ReadingVolts
		fields["upper_threshold_critical"] = j.UpperThresholdCritical
		fields["upper_threshold_fatal"] = j.UpperThresholdFatal
		fields["lower_threshold_critical"] = j.LowerThresholdCritical
		fields["lower_threshold_fatal"] = j.LowerThresholdFatal
		acc.AddFields("redfish_power_voltages", fields, tags)
	}

	return nil
}

func (r *Redfish) getPower(ref string) (*power, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	power := &power{}
	err := r.getData(loc.String(), power)
	if err != nil {
		return nil, err
	}
	return power, nil
}
