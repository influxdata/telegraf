package dmcache

import (
	"os/exec"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type DMCache struct {
	PerDevice bool `toml:"per_device"`
	rawStatus func() ([]string, error)
}

var sampleConfig = `
	## Whether to report per-device stats or not
	per_device = true
`

func (c *DMCache) SampleConfig() string {
	return sampleConfig
}

func (c *DMCache) Description() string {
	return "Provide a native collection for dmsetup based statistics for dm-cache"
}

func init() {
	inputs.Add("dmcache", func() telegraf.Input {
		return &DMCache{
			PerDevice: true,
			rawStatus: func() ([]string, error) {
				out, err := exec.Command("/bin/sh", "-c", "sudo /sbin/dmsetup status --target cache").Output()
				if err != nil {
					return nil, err
				}
				if string(out) == "No devices found\n" {
					return nil, nil
				}

				status := strings.Split(string(out), "\n")
				status = status[:len(status)-1] // removing last empty line

				return status, nil
			},
		}
	})
}
