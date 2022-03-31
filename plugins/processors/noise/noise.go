package noise

import (
	"fmt"
	"math"
	"reflect"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	defaultScale     = 1.0
	defaultMin       = -1.0
	defaultMax       = 1.0
	defaultMu        = 0.0
	defaultNoiseType = "laplacian"
)

type Noise struct {
	Scale         float64         `toml:"scale"`
	Min           float64         `toml:"min"`
	Max           float64         `toml:"max"`
	Mu            float64         `toml:"mu"`
	IncludeFields []string        `toml:"include_fields"`
	ExcludeFields []string        `toml:"exclude_fields"`
	NoiseType     string          `toml:"type"`
	Log           telegraf.Logger `toml:"-"`
	generator     distuv.Rander
	fieldFilter   filter.Filter
}

// generates a random noise value depending on the defined probability density
// function and adds that to the original value. If any integer overflows
// happen during the calculation, the result is set to MaxInt or 0 (for uint)
func (p *Noise) addNoise(value interface{}) interface{} {
	n := p.generator.Rand()
	switch v := value.(type) {
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		if v > 0 && (n > math.Nextafter(float64(math.MaxInt64), 0) || int64(n) > math.MaxInt64-v) {
			p.Log.Debug("Int64 overflow, setting value to MaxInt64")
			return int64(math.MaxInt64)
		}
		if v < 0 && (n < math.Nextafter(float64(math.MinInt64), 0) || int64(n) < math.MinInt64-v) {
			p.Log.Debug("Int64 (negative) overflow, setting value to MinInt64")
			return int64(math.MinInt64)
		}
		return v + int64(n)
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		if n < 0 {
			if uint64(-n) > v {
				p.Log.Debug("Uint64 (negative) overflow, setting value to 0")
				return uint64(0)
			}
			return v - uint64(-n)
		}
		if n > math.Nextafter(float64(math.MaxUint64), 0) || uint64(n) > math.MaxUint64-v {
			p.Log.Debug("Uint64 overflow, setting value to MaxUint64")
			return uint64(math.MaxUint64)
		}
		return v + uint64(n)
	case float32:
		return v + float32(n)
	case float64:
		return v + n
	default:
		p.Log.Debugf("Value (%v) type invalid: [%v] is not an int, uint or float", v, reflect.TypeOf(value))
	}
	return value
}

// Creates a filter for Include and Exclude fields and sets the desired noise
// distribution
func (p *Noise) Init() error {
	fieldFilter, err := filter.NewIncludeExcludeFilter(p.IncludeFields, p.ExcludeFields)
	if err != nil {
		return fmt.Errorf("creating fieldFilter failed: %v", err)
	}
	p.fieldFilter = fieldFilter

	switch p.NoiseType {
	case "", "laplacian":
		p.generator = &distuv.Laplace{Mu: p.Mu, Scale: p.Scale}
	case "uniform":
		p.generator = &distuv.Uniform{Min: p.Min, Max: p.Max}
	case "gaussian":
		p.generator = &distuv.Normal{Mu: p.Mu, Sigma: p.Scale}
	default:
		return fmt.Errorf("unknown distribution type %q", p.NoiseType)
	}
	return nil
}

func (p *Noise) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			if !p.fieldFilter.Match(field.Key) {
				continue
			}
			field.Value = p.addNoise(field.Value)
		}
	}
	return metrics
}

func init() {
	processors.Add("noise", func() telegraf.Processor {
		return &Noise{
			NoiseType: defaultNoiseType,
			Mu:        defaultMu,
			Scale:     defaultScale,
			Min:       defaultMin,
			Max:       defaultMax,
		}
	})
}
