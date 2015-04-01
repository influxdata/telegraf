package tivan

import "github.com/stretchr/testify/mock"

import "github.com/vektra/cypress"

type MockPlugin struct {
	mock.Mock
}

func (m *MockPlugin) Read() ([]*cypress.Message, error) {
	ret := m.Called()

	r0 := ret.Get(0).([]*cypress.Message)
	r1 := ret.Error(1)

	return r0, r1
}
