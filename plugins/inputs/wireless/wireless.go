package wireless

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embedd the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

// Wireless is used to store configuration values.
type Wireless struct {
	HostProc string          `toml:"host_proc"`
	Log      telegraf.Logger `toml:"-"`
}

func (*Wireless) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
