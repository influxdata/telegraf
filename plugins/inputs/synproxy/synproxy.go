package synproxy

import (
	"path"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Synproxy struct {
	Log telegraf.Logger `toml:"-"`

	// Synproxy stats filename (proc filesystem)
	statFile string
}

func (k *Synproxy) Description() string {
	return "Get synproxy counter statistics from procfs"
}

func (k *Synproxy) SampleConfig() string {
	return ""
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{
			statFile: path.Join(getHostProc(), "/net/stat/synproxy"),
		}
	})
}
