//go:generate ../../../tools/readme_config_includer/generator
package synproxy

import (
	_ "embed"
	"os"
	"path"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embedd the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type Synproxy struct {
	Log telegraf.Logger `toml:"-"`

	// Synproxy stats filename (proc filesystem)
	statFile string
}

func (*Synproxy) SampleConfig() string {
	return sampleConfig
}

func getHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func init() {
	inputs.Add("synproxy", func() telegraf.Input {
		return &Synproxy{
			statFile: path.Join(getHostProc(), "/net/stat/synproxy"),
		}
	})
}
