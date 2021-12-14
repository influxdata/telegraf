package system

import (
	"os"

	"github.com/stretchr/testify/mock"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"

	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type MockPS struct {
	mock.Mock
	PSDiskDeps
}

type MockPSDisk struct {
	*SystemPS
	*mock.Mock
}

type MockDiskUsage struct {
	*mock.Mock
}

func (m *MockPS) LoadAvg() (*load.AvgStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*load.AvgStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) CPUTimes(_, _ bool) ([]cpu.TimesStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]cpu.TimesStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) DiskUsage(mountPointFilter []string, fstypeExclude []string) ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	ret := m.Called(mountPointFilter, fstypeExclude)

	r0 := ret.Get(0).([]*disk.UsageStat)
	r1 := ret.Get(1).([]*disk.PartitionStat)
	r2 := ret.Error(2)

	return r0, r1, r2
}

func (m *MockPS) NetIO() ([]net.IOCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.IOCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) NetProto() ([]net.ProtoCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.ProtoCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) DiskIO(_ []string) (map[string]disk.IOCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(map[string]disk.IOCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) VMStat() (*mem.VirtualMemoryStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*mem.VirtualMemoryStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) SwapStat() (*mem.SwapMemoryStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*mem.SwapMemoryStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) Temperature() ([]host.TemperatureStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]host.TemperatureStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) NetConnections() ([]net.ConnectionStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.ConnectionStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockDiskUsage) Partitions(all bool) ([]disk.PartitionStat, error) {
	ret := m.Called(all)

	r0 := ret.Get(0).([]disk.PartitionStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockDiskUsage) OSGetenv(key string) string {
	ret := m.Called(key)
	return ret.Get(0).(string)
}

func (m *MockDiskUsage) OSStat(name string) (os.FileInfo, error) {
	ret := m.Called(name)

	r0 := ret.Get(0).(os.FileInfo)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockDiskUsage) PSDiskUsage(path string) (*disk.UsageStat, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).(*disk.UsageStat)
	r1 := ret.Error(1)

	return r0, r1
}
