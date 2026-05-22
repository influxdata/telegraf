package redfish

import (
	"github.com/influxdata/telegraf"
	"github.com/stmcginnis/gofish/schemas"
)

type MembersList struct {
	Members []struct {
		Ref string `json:"@odata.id"`
	}
}

func (r *Redfish) gatherStorage(acc telegraf.Accumulator, address string, system *schemas.ComputerSystem, chassis *schemas.Chassis) error {
	storage, err := system.Storage()
	if err != nil && len(storage) == 0 {
		return err
	}

	for _, j := range storage {
		tags := make(map[string]string, 20)
		tags["source"] = system.HostName
		tags["address"] = address
		tags["state"] = string(j.Status.State)
		tags["health_rollup"] = string(j.Status.HealthRollup)
		if _, ok := r.tagSet[tagSetChassisLocation]; ok {
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		drives, err := j.Drives()
		if err != nil {
			return err
		}

		for _, drive := range drives {
			fields := make(map[string]interface{}, 5)
			tags["manufacturer"] = drive.Manufacturer
			tags["media_type"] = string(drive.MediaType)
			tags["model"] = drive.Model
			tags["location"] = drive.PhysicalLocation.PartLocation.ServiceLabel
			tags["protocol"] = string(drive.Protocol)
			tags["serial_number"] = drive.SerialNumber
			tags["disk_health"] = string(drive.Status.Health)
			tags["disk_state"] = string(drive.Status.State)
			fields["speed_gbs"] = drive.CapableSpeedGbs
			fields["capacity_bytes"] = drive.CapacityBytes

			acc.AddFields("redfish_storage", fields, tags)
		}

	}

	return nil
}
