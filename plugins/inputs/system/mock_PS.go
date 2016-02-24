package system

import (
	"github.com/stretchr/testify/mock"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"

	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type MockPS struct {
	mock.Mock
}

func (m *MockPS) LoadAvg() (*load.LoadAvgStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*load.LoadAvgStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.CPUTimesStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]cpu.CPUTimesStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) DiskUsage(mountPointFilter []string, fstypeExclude []string) ([]*disk.DiskUsageStat, error) {
	ret := m.Called(mountPointFilter, fstypeExclude)

	r0 := ret.Get(0).([]*disk.DiskUsageStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) NetIO() ([]net.NetIOCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.NetIOCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) NetProto() ([]net.NetProtoCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.NetProtoCountersStat)
	r1 := ret.Error(1)

	return r0, r1
}

func (m *MockPS) DiskIO() (map[string]disk.DiskIOCountersStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(map[string]disk.DiskIOCountersStat)
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

func (m *MockPS) NetConnections() ([]net.NetConnectionStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]net.NetConnectionStat)
	r1 := ret.Error(1)

	return r0, r1
}
