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

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
const sFalse = 0x00000001

type Wmi struct {
	Host     string          `toml:"host"`
	Username config.Secret   `toml:"username"`
	Password config.Secret   `toml:"password"`
	Queries  []query         `toml:"query"`
	Methods  []method        `toml:"method"`
	Log      telegraf.Logger `toml:"-"`
}

func (*Wmi) SampleConfig() string {
	return sampleConfig
}

func (w *Wmi) Init() error {
	for i := range w.Queries {
		q := &w.Queries[i]
		if err := q.prepare(w.Host, w.Username, w.Password); err != nil {
			return fmt.Errorf("preparing query %q failed: %w", q.ClassName, err)
		}
	}

	for i := range w.Methods {
		m := &w.Methods[i]
		if err := m.prepare(w.Host, w.Username, w.Password); err != nil {
			return fmt.Errorf("preparing method %q failed: %w", m.Method, err)
		}
	}

	return nil
}

func (w *Wmi) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, q := range w.Queries {
		wg.Add(1)
		go func(q query) {
			defer wg.Done()
			acc.AddError(q.execute(acc))
		}(q)
	}

	for _, m := range w.Methods {
		wg.Add(1)
		go func(m method) {
			defer wg.Done()
			acc.AddError(m.execute(acc))
		}(m)
	}

	wg.Wait()

	return nil
}

func init() {
	inputs.Add("win_wmi", func() telegraf.Input { return &Wmi{} })
}
