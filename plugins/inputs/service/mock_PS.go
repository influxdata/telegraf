package service

import (
	"github.com/shirou/gopsutil/process"
	"github.com/stretchr/testify/mock"
)

type MockPs struct {
	mock.Mock
}

func (m *MockPs) MemInfo(processName string) ([]*process.MemoryInfoStat, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]*process.MemoryInfoStat)
	r1 := ret.Error(1)

	return r0, r1
}
