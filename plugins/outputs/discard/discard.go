//go:generate go run ../../../scripts/generate_plugindata/main.go
//go:generate go run ../../../scripts/generate_plugindata/main.go --clean
package discard

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Discard struct{}

func (d *Discard) Connect() error { return nil }
func (d *Discard) Close() error   { return nil }
func (d *Discard) SampleConfig() string {
	return `{{ .SampleConfig }}`
}
func (d *Discard) Write(_ []telegraf.Metric) error {
	return nil
}

func init() {
	outputs.Add("discard", func() telegraf.Output { return &Discard{} })
}
