package system

import (
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/stretchr/testify/mock"
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
