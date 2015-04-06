package system

import "github.com/stretchr/testify/mock"

import "github.com/influxdb/tivan/plugins/system/ps/cpu"
import "github.com/influxdb/tivan/plugins/system/ps/disk"
import "github.com/influxdb/tivan/plugins/system/ps/load"
import "github.com/influxdb/tivan/plugins/system/ps/mem"
import "github.com/influxdb/tivan/plugins/system/ps/net"

type MockPS struct {
	mock.Mock
}

func (m *MockPS) LoadAvg() (*load.LoadAvgStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).(*load.LoadAvgStat)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *MockPS) CPUTimes() ([]cpu.CPUTimesStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]cpu.CPUTimesStat)
	r1 := ret.Error(1)

	return r0, r1
}
func (m *MockPS) DiskUsage() ([]*disk.DiskUsageStat, error) {
	ret := m.Called()

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
