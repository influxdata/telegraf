//go:build windows

package win_wmi

import (
	_ "embed"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Wmi struct
type Wmi struct {
	Queries []Query `toml:"query"`
	Log     telegraf.Logger
}

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
const sFalse = 0x00000001

// Init function
func (s *Wmi) Init() error {
	return compileInputs(s)
}

// SampleConfig function
func (s *Wmi) SampleConfig() string {
	return sampleConfig
}

// Gather function
func (s *Wmi) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, query := range s.Queries {
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			err := q.execute(acc)
			if err != nil {
				acc.AddError(err)
			}
		}(query)
	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
