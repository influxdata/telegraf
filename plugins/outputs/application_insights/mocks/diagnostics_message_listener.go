package mocks

import "github.com/stretchr/testify/mock"

type DiagnosticsMessageListener struct {
	mock.Mock
}

func (_m *DiagnosticsMessageListener) Remove() {
	_m.Called()
}
