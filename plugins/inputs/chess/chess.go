package chess

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// chess.go

// chess is the plugin type
type Chess struct {
	Ok  bool            `toml:"ok"`
	Log telegraf.Logger `toml:"-"`
}

func (c *Chess) Description() string {
	return "A description of the chess plugin..."
}

func (c *Chess) SampleConfig() string {
	return `
	## Return a sample configuration for plugin...
	ok = true
`
}

// Init is a method that sets up and validates the config
func (c *Chess) Init() error {
	return nil
}

func (c *Chess) Gather(acc telegraf.Accumulator) error {
	if c.Ok {
		acc.AddFields("state", map[string]interface{}{"value": "pretty good"}, nil)
	} else {
		acc.AddFields("state", map[string]interface{}{"value": "not great"}, nil)
	}
	return nil
}

func init() {
	inputs.Add("chess", func() telegraf.Input { return &Chess{} })
}
