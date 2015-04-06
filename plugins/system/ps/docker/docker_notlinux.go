// +build !linux

package docker

import (
	"encoding/json"

	"github.com/influxdb/tivan/plugins/system/ps/common"
	"github.com/influxdb/tivan/plugins/system/ps/cpu"
)

// GetDockerIDList returnes a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	return nil, common.NotImplementedError
}

// CgroupCPU returnes specified cgroup id CPU status.
// containerid is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerid = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
func CgroupCPU(containerid string, base string) (*cpu.CPUTimesStat, error) {
	return nil, common.NotImplementedError
}

func CgroupCPUDocker(containerid string) (*cpu.CPUTimesStat, error) {
	return CgroupCPU(containerid, "/sys/fs/cgroup/cpuacct/docker")
}

func CgroupMem(containerid string, base string) (*CgroupMemStat, error) {
	return nil, common.NotImplementedError
}

func CgroupMemDocker(containerid string) (*CgroupMemStat, error) {
	return CgroupMem(containerid, "/sys/fs/cgroup/memory/docker")
}

func (m CgroupMemStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}
