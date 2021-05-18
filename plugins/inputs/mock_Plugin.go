package inputs

import (
	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/mock"
)

// MockPlugin struct should be named the same as the Plugin
type MockPlugin struct {
	mock.Mock

	constructedVariable string
}

// Description will appear directly above the plugin definition in the config file
func (m *MockPlugin) Description() string {
	return `This is an example plugin`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (m *MockPlugin) SampleConfig() string {
	return `  sampleVar = 'foo'`
}

// Init can be implemented to do one-time processing stuff like initializing variables
func (m *MockPlugin) Init() error {
	m.constructedVariable = "I'm initialized now."
	return nil
}

// Gather defines what data the plugin will gather.
func (m *MockPlugin) Gather(_a0 telegraf.Accumulator) error {
	ret := m.Called(_a0)

	r0 := ret.Error(0)

	return r0
}
