package redfish

import (
	"net/url"

	"github.com/influxdata/telegraf"
)

type Members []struct {
	Ref string `json:"@odata.id"`
}

// The data returned by system.Storage.Ref
type StorageCollection struct {
	Members Members
	Storage []Storage
}

type Storage struct {
	Name   string
	Status struct {
		State        string
		HealthRollup string
	}
	Drives []Drive
}

// Data from Storage.Drives[x].Ref
type Drive struct {
	Ref              string `json:"@odata.id"`
	Status           status
	PhysicalLocation struct {
		PartLocation struct {
			ServiceLabel string
			LocationType string
		}
	}
	CapableSpeedGbs float32
	CapacityBytes   uint64
	MediaType       string
	Manufacturer    string
	Model           string
	SerialNumber    string
	Protocol        string
}

type MembersList struct {
	Members []struct {
		Ref string `json:"@odata.id"`
	}
}

func (r *Redfish) gatherStorage(acc telegraf.Accumulator, address string, system *system, chassis *chassis) error {
	storage := StorageCollection{}
	err := r.getStorage(system.Storage.Ref, &storage)
	if err != nil {
		return err
	}

	for _, j := range storage.Storage {
		tags := make(map[string]string, 20)
		tags["address"] = address
		tags["source"] = system.Hostname
		tags["state"] = j.Status.State
		tags["health_rollup"] = j.Status.HealthRollup
		if _, ok := r.tagSet[tagSetChassisLocation]; ok && chassis.Location != nil {
			tags["datacenter"] = chassis.Location.PostalAddress.DataCenter
			tags["room"] = chassis.Location.PostalAddress.Room
			tags["rack"] = chassis.Location.Placement.Rack
			tags["row"] = chassis.Location.Placement.Row
		}
		if _, ok := r.tagSet[tagSetChassis]; ok {
			setChassisTags(chassis, tags)
		}

		for _, drive := range j.Drives {
			fields := make(map[string]interface{})
			tags["manufacturer"] = drive.Manufacturer
			tags["media_type"] = drive.MediaType
			tags["model"] = drive.Model
			tags["location"] = drive.PhysicalLocation.PartLocation.ServiceLabel
			tags["protocol"] = drive.Protocol
			tags["serial_number"] = drive.SerialNumber
			fields["speed_gbs"] = drive.CapableSpeedGbs
			fields["capacity_bytes"] = drive.CapacityBytes
			fields["disk_health"] = drive.Status.Health
			fields["disk_state"] = drive.Status.State

			acc.AddFields("redfish_storage", fields, tags)
		}

	}

	return nil
}

func (r *Redfish) getStorage(ref string, totalStorage *StorageCollection) error {
	loc := r.baseURL.ResolveReference(&url.URL{Path: ref})
	err := r.getData(loc.String(), totalStorage)

	if err != nil {
		return err
	}

	// If any data gathering fails, just continue as there might be further endpoints that work.
	// Otherwise its just an empty array that wont produce any metrics.
	for _, member := range totalStorage.Members {
		loc := r.baseURL.ResolveReference(&url.URL{Path: member.Ref})
		storage := &Storage{}
		err = r.getData(loc.String(), storage)
		if err != nil {
			err = nil
			continue
		}

		for i, storageDisk := range storage.Drives {
			loc := r.baseURL.ResolveReference(&url.URL{Path: storageDisk.Ref})
			err = r.getData(loc.String(), &storage.Drives[i])
			if err != nil {
				err = nil
				continue
			}
		}

		totalStorage.Storage = append(totalStorage.Storage, *storage)
	}

	return err
}
