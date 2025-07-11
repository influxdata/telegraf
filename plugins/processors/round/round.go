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
		return round(v, p.factor)
	case int8:
		return round(v, p.factor)
	case int16:
		return round(v, p.factor)
	case int32:
		return round(v, p.factor)
	case int64:
		return round(v, p.factor)
	case uint:
		return round(v, p.factor)
	case uint8:
		return round(v, p.factor)
	case uint16:
		return round(v, p.factor)
	case uint32:
		return round(v, p.factor)
	case uint64:
		return round(v, p.factor)
	case float32:
		return round(v, p.factor)
	case float64:
		return round(v, p.factor)
	default:
		p.Log.Debugf("Value (%v) type invalid: [%v] is not an int, uint or float", v, reflect.TypeOf(value))
	}
	return value
}

func round[V constraints.Integer | constraints.Float](value V, factor float64) V {
	return V(math.Round(float64(value)/factor) * factor)
}

func init() {
	processors.Add("round", func() telegraf.Processor {
		return &Round{
			Precision: 3,
		}
	})
}
