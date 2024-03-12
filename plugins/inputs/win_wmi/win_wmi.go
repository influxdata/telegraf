//go:generate ../../../tools/readme_config_includer/generator
//go:build windows

package win_wmi

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Wmi struct
type Wmi struct {
	Host     string          `toml:"host"`
	Username config.Secret   `toml:"username"`
	Password config.Secret   `toml:"password"`
	Queries  []Query         `toml:"query"`
	Log      telegraf.Logger `toml:"-"`
}

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
const sFalse = 0x00000001

// Init function
func (w *Wmi) Init() error {
	for i := range w.Queries {
		q := &w.Queries[i]
		if err := q.prepare(w.Host, w.Username, w.Password); err != nil {
			return fmt.Errorf("preparing query %q failed: %w", q.ClassName, err)
		}
	}

	return nil
}

// SampleConfig function
func (*Wmi) SampleConfig() string {
	return sampleConfig
}

// Gather function
func (w *Wmi) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, query := range w.Queries {
		wg.Add(1)
		go func(q Query) {
			defer wg.Done()
			acc.AddError(q.execute(acc))
		}(query)
	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
