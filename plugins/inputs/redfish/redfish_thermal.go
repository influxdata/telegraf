package redfish

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/influxdata/telegraf"
)

func (r *Redfish) gatherThermal(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	// Gather metrics via the legacy api
	// TODO: Add thermal metric gatering via the new thermalSubsys API
	return r.gatherThermalMetrics(acc, address, system, chassis)
}

func (r *Redfish) gatherThermalMetrics(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	thermal, err := chassis.Thermal()
	if err != nil {
		return fmt.Errorf("parsing thermal data from %q failed: %w", address, err)
	}
	// thermal is nil when the legacy thermal endpoints are not available
	// The newer subsys endpoints should be used as a form of retry
	if thermal == nil {
		r.Log.Warnf("Skipping thermal data of chassis %q. Only the legacy thermal API is supported at the moment", chassis.ID)
		return nil
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
			maps.Copy(tags, r.chassisTags)
		}

		fields := make(map[string]interface{})
		fields["reading_celsius"] = j.ReadingCelsius
		fields["upper_threshold_critical"] = j.UpperThresholdCritical
		fields["upper_threshold_fatal"] = j.UpperThresholdFatal
		fields["lower_threshold_critical"] = j.LowerThresholdCritical
		fields["lower_threshold_fatal"] = j.LowerThresholdFatal
		acc.AddFields("redfish_thermal_temperatures", fields, tags)
	}

	for i := range thermal.Fans {
		tags := make(map[string]string, 20)
		fields := make(map[string]interface{}, 5)
		tags["member_id"] = thermal.Fans[i].MemberID
		tags["address"] = address

		if thermal.Fans[i].Name == "" {
			tags["name"] = thermal.Fans[i].FanName //nolint:staticcheck // FanName is Deprecated but kept around for ilo4 support
		} else {
			tags["name"] = thermal.Fans[i].Name
		}
		tags["source"] = system.HostName
		tags["state"] = string(thermal.Fans[i].Status.State)
		tags["health"] = string(thermal.Fans[i].Status.Health)
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
			maps.Copy(tags, r.chassisTags)
		}

		// Due to ILO4 not being fully readfish compatible we have to do this parsing manually
		var ilo4ReadingPercent struct {
			CurrentReading *int64
		}
		json.Unmarshal(thermal.Fans[i].RawData, &ilo4ReadingPercent) //nolint:errcheck // Ignore if the marshalling fails as this legacy block should be removed

		if ilo4ReadingPercent.CurrentReading != nil {
			fields["reading_percent"] = ilo4ReadingPercent.CurrentReading
		} else {
			if thermal.Fans[i].ReadingUnits == "RPM" {
				fields["upper_threshold_critical"] = thermal.Fans[i].UpperThresholdCritical
				fields["upper_threshold_fatal"] = thermal.Fans[i].UpperThresholdFatal
				fields["lower_threshold_critical"] = thermal.Fans[i].LowerThresholdCritical
				fields["lower_threshold_fatal"] = thermal.Fans[i].LowerThresholdFatal
				fields["reading_rpm"] = thermal.Fans[i].Reading
			} else {
				fields["reading_percent"] = thermal.Fans[i].Reading
			}
		}
		acc.AddFields("redfish_thermal_fans", fields, tags)
	}

	return nil
}
