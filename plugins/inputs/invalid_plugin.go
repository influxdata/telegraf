package inputs

import (
	"log"

	"github.com/influxdata/telegraf"
)

// InvalidPlugin is used when a plugin may be invalid for a
// platform or other reason.
type InvalidPlugin struct {
	InvalidReason string
	OrigDesc      string
	OrigSampleCfg string
}

// Description should be inherited from the parent plugin.
func (i *InvalidPlugin) Description() string {
	return i.OrigDesc
}

// SampleConfig should be inherited from the parent plugin.
func (i *InvalidPlugin) SampleConfig() string {
	return i.OrigSampleCfg
}

// Gather does nothing, as the plugin is invalid.
func (i *InvalidPlugin) Gather(_a0 telegraf.Accumulator) error {
	return nil
}

// Start prints the reason the plugin was not valid.
func (i *InvalidPlugin) Start(telegraf.Accumulator) error {
	log.Printf("E! %s", i.InvalidReason)
	return nil
}

// Stop does nothing, as the plugin is invalid.
func (i *InvalidPlugin) Stop() { return }
