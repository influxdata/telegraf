package podman

import "github.com/containers/podman/v3/libpod/define"

var info = define.Info{
	Host: &define.HostInfo{
		Distribution: define.DistributionInfo{
			Distribution: "fedora",
		},
		CPUs: 8,
	},
	Store: &define.StoreInfo{
		ContainerStore: define.ContainerStore{
			Number:  2,
			Paused:  0,
			Running: 1,
			Stopped: 1,
		},
		ImageStore: define.ImageStore{
			Number: 10,
		},
	},
	Version: define.Version{
		Version: "3.2.0",
	},
}
