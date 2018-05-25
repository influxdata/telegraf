package crypto

import (
	//"errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// http://forum.hiveos.farm/discussion/192/hive-api
type hiveos struct {
}

var hiveosSampleConf = `
`

func (*hiveos) Description() string {
	return "Read hiveos's mining status"
}

func (*hiveos) SampleConfig() string {
	return hiveosSampleConf
}

func (e *hiveos) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("hiveos", func() telegraf.Input { return &hiveos{} })
}
