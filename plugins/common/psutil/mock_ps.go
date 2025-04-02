package psutil

import (
	"os"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/stretchr/testify/mock"
)

// MockPS is a mock implementation of the PS interface for testing purposes.
type MockPS struct {
	mock.Mock
	PSDiskDeps
}

// MockPSDisk is a mock implementation of the PSDiskDeps interface for testing purposes.
type MockPSDisk struct {
	*SystemPS
	*mock.Mock
}

// MockDiskUsage is a mock implementation for disk usage operations.
type MockDiskUsage struct {
	*mock.Mock
}

// CPUTimes returns the CPU times statistics.
func (m *MockPS) CPUTimes(_, _ bool) ([]cpu.TimesStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]cpu.TimesStat)
	r1 := ret.Error(1)

	return r0, r1
}

// DiskUsage returns the disk usage statistics.
func (m *MockPS) DiskUsage(mountPointFilter, mountOptsExclude, fstypeExclude []string) ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	ret := m.Called(mountPointFilter, mountOptsExclude, fstypeExclude)

	r0 := ret.Get(0).([]*disk.UsageStat)
	r1 := ret.Get(1).([]*disk.PartitionStat)
	r2 := ret.Error(2)

	return r0, r1, r2
}

// NetIO returns network I/O statistics for every network interface installed on the system.
func (m *MockPS) NetIO() ([]net.IOCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.IOCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

// NetProto returns network statistics for the entire system.
func (m *MockPS) NetProto() ([]net.ProtoCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.ProtoCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

// DiskIO returns the disk I/O statistics.
func (m *MockPS) DiskIO(_ []string) (map[string]disk.IOCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(map[string]disk.IOCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

// VMStat returns the virtual memory statistics.
func (m *MockPS) VMStat() (*mem.VirtualMemoryStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*mem.VirtualMemoryStat)
	r1 := ret.Error(1)

	return r0, r1
}

// SwapStat returns the swap memory statistics.
func (m *MockPS) SwapStat() (*mem.SwapMemoryStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*mem.SwapMemoryStat)
	r1 := ret.Error(1)

	return r0, r1
}

// NetConnections returns a list of network connections opened.
func (m *MockPS) NetConnections() ([]net.ConnectionStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.ConnectionStat)
	r1 := ret.Error(1)

	return r0, r1
}

// NetConntrack returns more detailed info about the conntrack table.
func (m *MockPS) NetConntrack(perCPU bool) ([]net.ConntrackStat, error) {
	ret := m.Called(perCPU)

	r0 := ret.Get(0).([]net.ConntrackStat)
	r1 := ret.Error(1)

	return r0, r1
}

// Partitions returns the disk partition statistics.
func (m *MockDiskUsage) Partitions(all bool) ([]disk.PartitionStat, error) {
	ret := m.Called(all)

	r0 := ret.Get(0).([]disk.PartitionStat)
	r1 := ret.Error(1)

	return r0, r1
}

// OSGetenv returns the value of the environment variable named by the key.
func (m *MockDiskUsage) OSGetenv(key string) string {
	ret := m.Called(key)
	return ret.Get(0).(string)
}

// OSStat returns the FileInfo structure describing the named file.
func (m *MockDiskUsage) OSStat(name string) (os.FileInfo, error) {
	ret := m.Called(name)

	r0 := ret.Get(0).(os.FileInfo)
	r1 := ret.Error(1)

	return r0, r1
}

// PSDiskUsage returns a file system usage for the specified path.
func (m *MockDiskUsage) PSDiskUsage(path string) (*disk.UsageStat, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).(*disk.UsageStat)
	r1 := ret.Error(1)

	return r0, r1
}
