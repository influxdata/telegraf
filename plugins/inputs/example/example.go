package example

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Example struct should be named the same as the Plugin
type Example struct {
	// Example for a mandatory option to set a tag
	DeviceName string `toml:"device_name"`

	// Config options are converted to the correct type automatically
	NumberFields int64 `toml:"number_fields"`

	// We can also use booleans and have diverging names between user-configuration options and struct members
	EnableRandomVariable bool `toml:"enable_random"`

	// Example of passing a duration option allowing the format of e.g. "100ms", "5m" or "1h"
	Timeout config.Duration `toml:"timeout"`

	// Telegraf logging facility
	// The exact name is important to allow automatic initialization by telegraf.
	Log telegraf.Logger `toml:"-"`

	// This is a non-exported internal state.
	count int64
}

// Usually the default (example) configuration is contained in this constant.
// Please use '## '' to denote comments and '# ' to specify default settings and start each line with two spaces.
const sampleConfig = `
  ## Device name used as a tag
  ## This is a mandatory option that needs to be set by the user, so we do not
  ## comment it.
  device_name = ""

  ## Number of fields contained in the output
  ## This should be greater than zero and less then ten.
	## Here, two is the default, so we comment the option with the default value shown.
  # number_fields = 2

  ## Enable setting the field(s) to random values
  ## By default, the field values are set to zero.
  # enable_random = false

  ## Specify a duration allowing time-unit suffixes ('ns','ms', 's', 'm', etc.)
	# timeout = "100ms"
`

// Description will appear directly above the plugin definition in the config file
func (m *Example) Description() string {
	return `This is an example plugin`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (m *Example) SampleConfig() string {
	return sampleConfig
}

// Init can be implemented to do one-time processing stuff like initializing variables
func (m *Example) Init() error {
	// Check your options according to your requirements
	if m.DeviceName == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	// Set your defaults.
	// Please note: In golang all fields are initialzed to their nil value, so you should not
	// set these fields if the nil value is what you want (e.g. for booleans).
	if m.NumberFields < 1 {
		m.Log.Debugf("Setting number of fields to default from invalid value %d", m.NumberFields)
		m.NumberFields = 2
	}

	// Initialze your internal states
	m.count = 1

	return nil
}

// Gather defines what data the plugin will gather.
func (m *Example) Gather(acc telegraf.Accumulator) error {
	// Imagine some completely arbitrary error occuring here
	if m.NumberFields > 10 {
		return fmt.Errorf("too many fields")
	}

	// For illustration we gather three metrics in one go
	for run := 0; run < 3; run++ {
		// Imagine an error occurs here but you want to keep the other
		// metrics, then you cannot simply return, as this would drop
		// all later metrics. Simply accumulate errors in this case
		// and ignore the metric.
		if m.EnableRandomVariable && m.DeviceName == "flappy" && run > 1 {
			acc.AddError(fmt.Errorf("too many runs for random values"))
			continue
		}

		// Construct the fields
		fields := map[string]interface{}{"count": m.count}
		for i := int64(1); i < m.NumberFields; i++ {
			name := fmt.Sprintf("field%d", i)
			value := 0.0
			if m.EnableRandomVariable {
				value = rand.Float64()
			}
			fields[name] = value
		}

		// Construct the tags
		tags := map[string]string{"device": m.DeviceName}

		// Add the metric with the current timestamp
		acc.AddFields("example", fields, tags)

		m.count++
	}

	return nil
}

// Register the plugin
func init() {
	inputs.Add("example", func() telegraf.Input {
		return &Example{
			// Set the default timeout here to distinguish it from the user setting it to zero
			Timeout: config.Duration(100 * time.Millisecond),
		}
	})
}
