package redfish

import (
	"net/url"

	"github.com/influxdata/telegraf"
)

type thermal struct {
	Fans []struct {
		Name                   string
		MemberID               string
		FanName                string
		CurrentReading         *int64
		Reading                *int64
		ReadingUnits           *string
		UpperThresholdCritical *int64
		UpperThresholdFatal    *int64
		LowerThresholdCritical *int64
		LowerThresholdFatal    *int64
		Status                 status
	}
	Temperatures []struct {
		Name                   string
		MemberID               string
		ReadingCelsius         *float64
		UpperThresholdCritical *float64
		UpperThresholdFatal    *float64
		LowerThresholdCritical *float64
		LowerThresholdFatal    *float64
		Status                 status
	}
}

func (r *Redfish) gatherThermal(acc telegraf.Accumulator, address string, system *system, chassis *chassis) error {
	thermal, err := r.getThermal(chassis.Thermal.Ref)
	if err != nil {
		return err
	}

	for _, j := range thermal.Temperatures {
		tags := make(map[string]string, 19)
		tags["member_id"] = j.MemberID
		tags["address"] = address
		tags["name"] = j.Name
		tags["source"] = system.Hostname
		tags["state"] = j.Status.State
		tags["health"] = j.Status.Health
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
		tags["name"] = j.Name
		if j.FanName != "" {
			tags["name"] = j.FanName
		}
		tags["source"] = system.Hostname
		tags["state"] = j.Status.State
		tags["health"] = j.Status.Health
		if _, ok := r.tagSet[tagSetChassisLocation]; ok && chassis.Location != nil {
			tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		if j.ReadingUnits != nil && *j.ReadingUnits == "RPM" {
			fields["upper_threshold_critical"] = j.UpperThresholdCritical
			fields["upper_threshold_fatal"] = j.UpperThresholdFatal
			fields["lower_threshold_critical"] = j.LowerThresholdCritical
			fields["lower_threshold_fatal"] = j.LowerThresholdFatal
			fields["reading_rpm"] = j.Reading
		} else if j.CurrentReading != nil {
			fields["reading_percent"] = j.CurrentReading
		} else {
			fields["reading_percent"] = j.Reading
		}
		acc.AddFields("redfish_thermal_fans", fields, tags)
	}

	return nil
}

func (r *Redfish) getThermal(ref string) (*thermal, error) {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	thermal := &thermal{}
	err := r.getData(loc.String(), thermal)
	if err != nil {
		return nil, err
	}
	return thermal, nil
}
