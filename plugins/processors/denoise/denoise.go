//go:generate ../../../tools/readme_config_includer/generator
package noise

import (
	_ "embed"
	"fmt"
	"math"
	"reflect"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultSignificantFigures = 3
)

type Denoise struct {
	SignificantFigures int             `toml:"sf"`
	IncludeFields      []string        `toml:"include_fields"`
	ExcludeFields      []string        `toml:"exclude_fields"`
	Log                telegraf.Logger `toml:"-"`
	fieldFilter        filter.Filter
}

func roundToSignificantFigures(f float64, sf int) float64 {
	magnitude := math.Pow(10, float64(sf-1)-math.Floor(math.Log10(math.Abs(f))))
	rounded := math.Round(f * magnitude)
	return rounded / magnitude
}

// denoises the provided value to the specific number of significant figures.
func (p *Denoise) denoise(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		if v == 0 {
			return int64(0)
		}

		return roundToSignificantFigures(float64(v), p.SignificantFigures)
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		if v == 0 {
			return uint64(0)
		}

		return roundToSignificantFigures(float64(v), p.SignificantFigures)
	case float32:
	case float64:
		if v == 0 {
			return float64(0)
		}

		return roundToSignificantFigures(v, p.SignificantFigures)
	default:
		p.Log.Debugf("Value (%v) type invalid: [%v] is not an int, uint or float", v, reflect.TypeOf(value))
	}
	return value
}

func (*Denoise) SampleConfig() string {
	return sampleConfig
}

// Creates a filter for Include and Exclude fields
func (p *Denoise) Init() error {
	fieldFilter, err := filter.NewIncludeExcludeFilter(p.IncludeFields, p.ExcludeFields)
	if err != nil {
		return fmt.Errorf("creating fieldFilter failed: %w", err)
	}
	p.fieldFilter = fieldFilter

	if p.SignificantFigures < 1 {
		return fmt.Errorf("significant figures must be at least 1, got %d", p.SignificantFigures)
	}

	return nil
}

func (p *Denoise) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			if !p.fieldFilter.Match(field.Key) {
				continue
			}
			field.Value = p.denoise(field.Value)
		}
	}
	return metrics
}

func init() {
	processors.Add("denoise", func() telegraf.Processor {
		return &Denoise{
			SignificantFigures: defaultSignificantFigures,
		}
	})
}
