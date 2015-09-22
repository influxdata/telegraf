// +build linux

package docker

import (
	"encoding/json"
        "os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	common "github.com/shirou/gopsutil/common"
	cpu "github.com/shirou/gopsutil/cpu"
)

// GetDockerIDList returnes a list of DockerID.
// This requires certain permission.
func GetDockerIDList() ([]string, error) {
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, ErrDockerNotAvailable
	}

	out, err := exec.Command(path, "ps", "-q", "--no-trunc").Output()
	if err != nil {
		return []string{}, err
	}
	lines := strings.Split(string(out), "\n")
	ret := make([]string, 0, len(lines))

	for _, l := range lines {
		if l == "" {
			continue
		}
		ret = append(ret, l)
	}

	return ret, nil
}

// CgroupCPU returnes specified cgroup id CPU status.
// containerid is same as docker id if you use docker.
// If you use container via systemd.slice, you could use
// containerid = docker-<container id>.scope and base=/sys/fs/cgroup/cpuacct/system.slice/
func CgroupCPU(containerid string, base string) (*cpu.CPUTimesStat, error) {
	if len(base) == 0 {
		base = "/sys/fs/cgroup/cpuacct/docker"
	}
	statfile := path.Join(base, containerid, "cpuacct.stat")

        if _, err := os.Stat(statfile); os.IsNotExist(err) {
            statfile = path.Join("/sys/fs/cgroup/cpuacct/system.slice",  "docker-" + containerid + ".scope", "cpuacct.stat")
        }

	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	// empty containerid means all cgroup
	if len(containerid) == 0 {
		containerid = "all"
	}
	ret := &cpu.CPUTimesStat{CPU: containerid}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if fields[0] == "user" {
			user, err := strconv.ParseFloat(fields[1], 64)
			if err == nil {
				ret.User = float64(user)
			}
		}
		if fields[0] == "system" {
			system, err := strconv.ParseFloat(fields[1], 64)
			if err == nil {
				ret.System = float64(system)
			}
		}
	}

	return ret, nil
}

func CgroupCPUDocker(containerid string) (*cpu.CPUTimesStat, error) {
	return CgroupCPU(containerid, "/sys/fs/cgroup/cpuacct/docker")
}

func CgroupMem(containerid string, base string) (*CgroupMemStat, error) {
	if len(base) == 0 {
		base = "/sys/fs/cgroup/memory/docker"
	}
	statfile := path.Join(base, containerid, "memory.stat")

        if _, err := os.Stat(statfile); os.IsNotExist(err) {
            statfile = path.Join("/sys/fs/cgroup/memory/system.slice",  "docker-" + containerid + ".scope", "memory.stat")
        }

	// empty containerid means all cgroup
	if len(containerid) == 0 {
		containerid = "all"
	}
	lines, err := common.ReadLines(statfile)
	if err != nil {
		return nil, err
	}
	ret := &CgroupMemStat{ContainerID: containerid}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		switch fields[0] {
		case "cache":
			ret.Cache = v
		case "rss":
			ret.RSS = v
		case "rss_huge":
			ret.RSSHuge = v
		case "mapped_file":
			ret.MappedFile = v
		case "pgpgin":
			ret.Pgpgin = v
		case "pgpgout":
			ret.Pgpgout = v
		case "pgfault":
			ret.Pgfault = v
		case "pgmajfault":
			ret.Pgmajfault = v
		case "inactive_anon":
			ret.InactiveAnon = v
		case "active_anon":
			ret.ActiveAnon = v
		case "inactive_file":
			ret.InactiveFile = v
		case "active_file":
			ret.ActiveFile = v
		case "unevictable":
			ret.Unevictable = v
		case "hierarchical_memory_limit":
			ret.HierarchicalMemoryLimit = v
		case "total_cache":
			ret.TotalCache = v
		case "total_rss":
			ret.TotalRSS = v
		case "total_rss_huge":
			ret.TotalRSSHuge = v
		case "total_mapped_file":
			ret.TotalMappedFile = v
		case "total_pgpgin":
			ret.TotalPgpgIn = v
		case "total_pgpgout":
			ret.TotalPgpgOut = v
		case "total_pgfault":
			ret.TotalPgFault = v
		case "total_pgmajfault":
			ret.TotalPgMajFault = v
		case "total_inactive_anon":
			ret.TotalInactiveAnon = v
		case "total_active_anon":
			ret.TotalActiveAnon = v
		case "total_inactive_file":
			ret.TotalInactiveFile = v
		case "total_active_file":
			ret.TotalActiveFile = v
		case "total_unevictable":
			ret.TotalUnevictable = v
		}
	}
	return ret, nil
}

func CgroupMemDocker(containerid string) (*CgroupMemStat, error) {
	return CgroupMem(containerid, "/sys/fs/cgroup/memory/docker")
}

func (m CgroupMemStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}
