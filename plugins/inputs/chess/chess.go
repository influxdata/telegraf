package chess

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// chess.go

// chess is the plugin type
type Chess struct {
	Profiles []string        `toml:"profiles"`
	Log      telegraf.Logger `toml:"-"`
}

const SampleConfig = `
  # A list of profiles for monotoring 
  profiles = ["username1", "username2"]
`

func (c *Chess) Description() string {
	return "Monitor profiles from chess.com"
}

func (c *Chess) SampleConfig() string {
	return SampleConfig
}

// Init is a method that sets up and validates the config
func (c *Chess) Init() error {
	if c.Profiles == nil {
		return fmt.Errorf("no profiles listed in the config")
	}
	return nil
}

func (c *Chess) Gather(acc telegraf.Accumulator) error {
	// if c.Ok {
	// 	acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
	// } else {
	// 	acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
	// }
	return nil
}

func init() {
	inputs.Add("chess", func() telegraf.Input { return &Chess{} })
}
