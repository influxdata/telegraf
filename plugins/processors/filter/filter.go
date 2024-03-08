//go:generate ../../../tools/readme_config_includer/generator
package filter

import (
	_ "embed"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Filter struct {
	Rules         []rule          `toml:"rule"`
	DefaultAction string          `toml:"default"`
	Log           telegraf.Logger `toml:"-"`
	defaultPass   bool
}

func (*Filter) SampleConfig() string {
	return sampleConfig
}

func (f *Filter) Init() error {
	// Check the default-action setting
	switch f.DefaultAction {
	case "", "pass":
		f.defaultPass = true
	case "drop":
		// Do nothing, those options are valid
		if len(f.Rules) == 0 {
			f.Log.Warn("dropping all metrics as no rule is provided")
		}
	default:
		return fmt.Errorf("invalid default action %q", f.DefaultAction)
	}

	// Check and initialize rules
	for i := range f.Rules {
		if err := f.Rules[i].init(); err != nil {
			return fmt.Errorf("initialization of rule %d failed: %w", i+1, err)
		}
	}

	return nil
}

func (f *Filter) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, m := range in {
		if f.applyRules(m) {
			out = append(out, m)
		} else {
			m.Drop()
		}
	}
	return out
}

func (f *Filter) applyRules(m telegraf.Metric) bool {
	for _, r := range f.Rules {
		if pass, applies := r.apply(m); applies {
			return pass
		}
	}
	return f.defaultPass
}

func init() {
	processors.Add("filter", func() telegraf.Processor {
		return &Filter{}
	})
}
