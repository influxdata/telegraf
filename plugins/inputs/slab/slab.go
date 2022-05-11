package slab

import (
	"os"
	"path"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type SlabStats struct {
	Log telegraf.Logger `toml:"-"`

	statFile string
	UseSudo  bool `toml:"use_sudo"`
}

func getHostProc() string {
	procPath := "/proc"
	if os.Getenv("HOST_PROC") != "" {
		procPath = os.Getenv("HOST_PROC")
	}
	return procPath
}

func init() {
	inputs.Add("slab", func() telegraf.Input {
		return &SlabStats{
			statFile: path.Join(getHostProc(), "/slabinfo"),
		}
	})
}
