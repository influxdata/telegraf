//go:generate ../../../tools/readme_config_includer/generator
package round

import (
	_ "embed"
	"fmt"
	"math"
	"reflect"

	"golang.org/x/exp/constraints"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Round struct {
	Precision     int             `toml:"precision"`
	IncludeFields []string        `toml:"include_fields"`
	ExcludeFields []string        `toml:"exclude_fields"`
	Log           telegraf.Logger `toml:"-"`

	factor float64
	fields filter.Filter
}

func (*Round) SampleConfig() string {
	return sampleConfig
}

// Creates a filter for Include and Exclude fields
func (p *Round) Init() error {
	fieldFilter, err := filter.NewIncludeExcludeFilter(p.IncludeFields, p.ExcludeFields)
	if err != nil {
		return fmt.Errorf("creating fieldFilter failed: %w", err)
	}
	p.fields = fieldFilter

	p.factor = math.Pow10(p.Precision * -1)

	return nil
}

func (p *Round) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			if !p.fields.Match(field.Key) {
				continue
			}
			field.Value = p.round(field.Value)
		}
	}
	return metrics
}

// rounds the provided value to Precision.
func (p *Round) round(value interface{}) interface{} {
	switch v := value.(type) {
	case int:
		return roundInt(v, int64(p.factor))
	case int8:
		return roundInt(v, int64(p.factor))
	case int16:
		return roundInt(v, int64(p.factor))
	case int32:
		return roundInt(v, int64(p.factor))
	case int64:
		return roundInt(v, int64(p.factor))
	case uint:
		return roundInt(v, int64(p.factor))
	case uint8:
		return roundInt(v, int64(p.factor))
	case uint16:
		return roundInt(v, int64(p.factor))
	case uint32:
		return roundInt(v, int64(p.factor))
	case uint64:
		return roundInt(v, int64(p.factor))
	case float32:
		return roundFloat(v, p.factor)
	case float64:
		return roundFloat(v, p.factor)
	default:
		p.Log.Debugf("Value (%v) type invalid: [%v] is not an int, uint or float", v, reflect.TypeOf(value))
	}
	return value
}

func roundInt[V constraints.Integer](value V, factor int64) V {
	f := V(factor)

	if factor < 10 {
		return value
	}

	v := value / f // -12.55

	remainder := (value / (f / 10)) % 10
	if remainder < 5 {
		return v * f
	}

	return (v + 1) * f
}

func roundFloat[V constraints.Integer | constraints.Float](value V, factor float64) V {
	return V(math.Round(float64(value)/factor) * factor)
}

func init() {
	processors.Add("round", func() telegraf.Processor {
		return &Round{
			Precision: 0,
		}
	})
}
