//go:generate ../../../tools/readme_config_includer/generator
package traceroute

import (
	_ "embed"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"sync"
)

//go:embed sample.conf
var sampleConfig string

type Traceroute struct {
	wg sync.WaitGroup

	Log telegraf.Logger `toml:"-"`

	// URLs to traceroute
	Urls []string
}

func (*Traceroute) SampleConfig() string {
	return sampleConfig
}

func (t *Traceroute) Init() error {
	for index, element := range t.Urls {
		fmt.Println("At index", index, "value is", element)
	}
	return nil
}

func (t *Traceroute) Gather(acc telegraf.Accumulator) error {
	println("Hola!")
	return nil
}

func init() {
	inputs.Add("traceroute", func() telegraf.Input {
		return &Traceroute{}
	})
}
