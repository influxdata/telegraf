package util

import (
	"fmt"
	"github.com/vmware/govmomi/vim25/types"
	"strings"
)

var diskMappings = make(map[string]map[string]string)

func SaveDiskInfo(uuid string, hostname string, name string) {
	diskMappings[uuid] =
		map[string]string{"hostName": hostname, "diskName": name}
}

func LoadDiskMappings(hostMoRef types.ManagedObjectReference, hostname string) {
	for _, diskMapInfoEx := range *queryDiskMappings(hostMoRef) {
		// get cache tier mapping
		ssd := diskMapInfoEx.Mapping.Ssd
		diskUuid := ssd.VsanDiskInfo.VsanUuid
		diskPaths := strings.Split(ssd.DevicePath, "/")
		diskName := diskPaths[len(diskPaths)-1]
		SaveDiskInfo(diskUuid, hostname, diskName)

		// get capacity tier mapping
		for _, capacityDisk := range diskMapInfoEx.Mapping.NonSsd {
			diskUuid := capacityDisk.VsanDiskInfo.VsanUuid
			diskPaths := strings.Split(capacityDisk.DevicePath, "/")
			diskName := diskPaths[len(diskPaths)-1]
			SaveDiskInfo(diskUuid, hostname, diskName)
		}
	}
	fmt.Println(diskMappings)
}

func DiskMappingsQuery(diskUuid string) (string, string) {
	return diskMappings[diskUuid]["diskName"], diskMappings[diskUuid]["hostName"]
}

func init() {
	host := types.ManagedObjectReference{
		Type:  "HostSystem",
		Value: "host-15",
	}
	LoadDiskMappings(host, "10.172.47.149")
}
