// +build !darwin,!freebsd,!linux,!netbsd,!openbsd

package prometheus_sockets

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Gather the measurements
func (p *PrometheusSocketWalker) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("prometheus_sockets", func() telegraf.Input {
		return &PrometheusSocketWalker{}
	})
}
