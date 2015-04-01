package tivan

import "github.com/stretchr/testify/mock"

import "github.com/vektra/cypress"

type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) Receive(_a0 *cypress.Message) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
