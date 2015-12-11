package system

import (
	"fmt"
	gonet "net"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/influxdb/telegraf/internal"
	"github.com/influxdb/telegraf/plugins"
	"github.com/influxdb/telegraf/testutil"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/docker"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	"github.com/stretchr/testify/assert"
)

type DockerContainerStat struct {
	Id      string
	Name    string
	Command string
	Labels  map[string]string
	CPU     *cpu.CPUTimesStat
	Mem     *docker.CgroupMemStat
}

type PS interface {
	CPUTimes(perCPU, totalCPU bool) ([]cpu.CPUTimesStat, error)
	DiskUsage() ([]*disk.DiskUsageStat, error)
	NetIO() ([]net.NetIOCountersStat, error)
	NetProto() ([]net.NetProtoCountersStat, error)
	DiskIO() (map[string]disk.DiskIOCountersStat, error)
	VMStat() (*mem.VirtualMemoryStat, error)
	SwapStat() (*mem.SwapMemoryStat, error)
	DockerStat() ([]*DockerContainerStat, error)
	NetConnections() ([]net.NetConnectionStat, error)
}

func add(acc plugins.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.Add(name, val, tags)
	}
}

type systemPS struct {
	dockerClient *dc.Client
}

func (s *systemPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.CPUTimesStat, error) {
	var cpuTimes []cpu.CPUTimesStat
	if perCPU {
		if perCPUTimes, err := cpu.CPUTimes(true); err == nil {
			cpuTimes = append(cpuTimes, perCPUTimes...)
		} else {
			return nil, err
		}
	}
	if totalCPU {
		if totalCPUTimes, err := cpu.CPUTimes(false); err == nil {
			cpuTimes = append(cpuTimes, totalCPUTimes...)
		} else {
			return nil, err
		}
	}
	return cpuTimes, nil
}

func (s *systemPS) DiskUsage() ([]*disk.DiskUsageStat, error) {
	parts, err := disk.DiskPartitions(true)
	if err != nil {
		return nil, err
	}

	var usage []*disk.DiskUsageStat

	for _, p := range parts {
		if _, err := os.Stat(p.Mountpoint); err == nil {
			du, err := disk.DiskUsage(p.Mountpoint)
			if err != nil {
				return nil, err
			}
			du.Fstype = p.Fstype
			usage = append(usage, du)
		}
	}

	return usage, nil
}

func (s *systemPS) NetProto() ([]net.NetProtoCountersStat, error) {
	return net.NetProtoCounters(nil)
}

func (s *systemPS) NetIO() ([]net.NetIOCountersStat, error) {
	return net.NetIOCounters(true)
}

func (s *systemPS) NetConnections() ([]net.NetConnectionStat, error) {
	return net.NetConnections("all")
}

func (s *systemPS) DiskIO() (map[string]disk.DiskIOCountersStat, error) {
	m, err := disk.DiskIOCounters()
	if err == internal.NotImplementedError {
		return nil, nil
	}

	return m, err
}

func (s *systemPS) VMStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (s *systemPS) SwapStat() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}

func (s *systemPS) DockerStat() ([]*DockerContainerStat, error) {
	if s.dockerClient == nil {
		c, err := dc.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			return nil, err
		}

		s.dockerClient = c
	}

	opts := dc.ListContainersOptions{}

	containers, err := s.dockerClient.ListContainers(opts)
	if err != nil {
		if _, ok := err.(*gonet.OpError); ok {
			return nil, nil
		}

		return nil, err
	}

	var stats []*DockerContainerStat

	for _, container := range containers {
		ctu, err := docker.CgroupCPUDocker(container.ID)
		if err != nil {
			return nil, err
		}

		mem, err := docker.CgroupMemDocker(container.ID)
		if err != nil {
			return nil, err
		}

		name := strings.Join(container.Names, " ")

		stats = append(stats, &DockerContainerStat{
			Id:      container.ID,
			Name:    name,
			Command: container.Command,
			Labels:  container.Labels,
			CPU:     ctu,
			Mem:     mem,
		})
	}

	return stats, nil
}

// Asserts that a given accumulator contains a measurment of type float64 with
// specific tags within a certain distance of a given expected value. Asserts a failure
// if the measurement is of the wrong type, or if no matching measurements are found
//
// Paramaters:
//     t *testing.T            : Testing object to use
//     acc testutil.Accumulator: Accumulator to examine
//     measurement string      : Name of the measurement to examine
//     expectedValue float64   : Value to search for within the measurement
//     delta float64           : Maximum acceptable distance of an accumulated value
//                               from the expectedValue parameter. Useful when
//                               floating-point arithmatic imprecision makes looking
//                               for an exact match impractical
//     tags map[string]string  : Tag set the found measurement must have. Set to nil to
//                               ignore the tag set.
func assertContainsTaggedFloat(
	t *testing.T,
	acc *testutil.Accumulator,
	measurement string,
	expectedValue float64,
	delta float64,
	tags map[string]string,
) {
	var actualValue float64
	for _, pt := range acc.Points {
		if pt.Measurement == measurement {
			if (tags == nil) || reflect.DeepEqual(pt.Tags, tags) {
				if value, ok := pt.Fields["value"].(float64); ok {
					actualValue = value
					if (value >= expectedValue-delta) && (value <= expectedValue+delta) {
						// Found the point, return without failing
						return
					}
				} else {
					assert.Fail(t, fmt.Sprintf("Measurement \"%s\" does not have type float64",
						measurement))
				}

			}
		}
	}
	msg := fmt.Sprintf("Could not find measurement \"%s\" with requested tags within %f of %f, Actual: %f",
		measurement, delta, expectedValue, actualValue)
	assert.Fail(t, msg)
}
