package inputs

import (
	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/mock"
)

type MockPlugin struct {
	mock.Mock
}

func (m *MockPlugin) Gather(_a0 telegraf.Accumulator) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
