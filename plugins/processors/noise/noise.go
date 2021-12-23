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

const sampleConfig = `
    ## Specified the type of the random distribution.
    ## Can be "laplacian", "gaussian" or "uniform".
    # type = "laplacian

    ## Center of the distribution.
    ## Only used for Laplacian and Gaussian distributions.
    # mu = 0.0

    ## Scale parameter for the Laplacian or Gaussian distribution
    # scale = 1.0

    ## Upper and lower bound of the Uniform distribution
    # min = -1.0
    # max = 1.0

    ## Apply the noise only to numeric fields matching the filter criteria below.
    ## Excludes takes precedence over includes.
    # include_fields = []
    # exclude_fields = []
`

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

func (p *Noise) SampleConfig() string {
	return sampleConfig
}

func (p *Noise) Description() string {
	return "Adds noise to numerical fields"
}

// generates a random noise value depending on the defined probability density
// function and adds that to the original value. If any integer overflows
// happen during the calculation, the result is set to MaxInt or 0 (for uint)
func (p *Noise) addNoise(value interface{}) interface{} {
	n := p.generator.Rand()
	switch v := value.(type) {
	case int:
		if v > 0 && (n > math.Nextafter(float64(math.MaxInt), 0) || int(n) > math.MaxInt-v) {
			p.Log.Debug("Int overflow, setting value to MaxInt")
			return int(math.MaxInt)
		}
		if v < 0 && (n < math.Nextafter(float64(math.MinInt), 0) || int(n) < math.MinInt-v) {
			p.Log.Debug("Int (negative) overflow, setting value to MinInt")
			return int(math.MinInt)
		}
		return v + int(n)
	case int8:
		if v > 0 && (n > math.Nextafter(float64(math.MaxInt8), 0) || int8(n) > math.MaxInt8-v) {
			p.Log.Debug("Int8 overflow, setting value to MaxInt8")
			return int8(math.MaxInt8)
		}
		if v < 0 && (n < math.Nextafter(float64(math.MinInt8), 0) || int8(n) < math.MinInt8-v) {
			p.Log.Debug("Int8 (negative) overflow, setting value to MinInt8")
			return int8(math.MinInt8)
		}
		return v + int8(n)
	case int16:
		if v > 0 && (n > math.Nextafter(float64(math.MaxInt16), 0) || int16(n) > math.MaxInt16-v) {
			p.Log.Debug("Int16 overflow, setting value to MaxInt16")
			return int16(math.MaxInt16)
		}
		if v < 0 && (n < math.Nextafter(float64(math.MinInt16), 0) || int16(n) < math.MinInt16-v) {
			p.Log.Debug("Int16 (negative) overflow, setting value to MinInt16")
			return int16(math.MinInt16)
		}
		return v + int16(n)
	case int32:
		if v > 0 && (n > math.Nextafter(float64(math.MaxInt32), 0) || int32(n) > math.MaxInt32-v) {
			p.Log.Debug("Int32 overflow, setting value to MaxInt32")
			return int32(math.MaxInt32)
		}
		if v < 0 && (n < math.Nextafter(float64(math.MinInt32), 0) || int32(n) < math.MinInt32-v) {
			p.Log.Debug("Int32 (negative) overflow, setting value to MinInt32")
			return int32(math.MinInt32)
		}
		return v + int32(n)
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
		if n < 0 {
			if uint(-n) > v {
				p.Log.Debug("Uint (negative) overflow, setting value to 0")
				return uint(0)
			}
			return v - uint(-n)
		}
		if n > math.Nextafter(float64(math.MaxUint), 0) || uint(n) > math.MaxUint-v {
			p.Log.Debug("Uint overflow, setting value to MaxUint")
			return uint(math.MaxUint)
		}
		return v + uint(n)
	case uint8:
		if n < 0 {
			if uint8(-n) > v {
				p.Log.Debug("Uint8 (negative) overflow, setting value to 0")
				return uint8(0)
			}
			return v - uint8(-n)
		}
		if n > math.Nextafter(float64(math.MaxUint8), 0) || uint8(n) > math.MaxUint8-v {
			p.Log.Debug("Uint8 overflow, setting value to MaxUint8")
			return uint8(math.MaxUint8)
		}
		return v + uint8(n)
	case uint16:
		if n < 0 {
			if uint16(-n) > v {
				p.Log.Debug("Uint16 (negative) overflow, setting value to 0")
				return uint16(0)
			}
			return v - uint16(-n)
		}
		if n > math.Nextafter(float64(math.MaxUint16), 0) || uint16(n) > math.MaxUint16-v {
			p.Log.Debug("Uint16 overflow, setting value to MaxUint16")
			return uint16(math.MaxUint16)
		}
		return v + uint16(n)
	case uint32:
		if n < 0 {
			if uint32(-n) > v {
				p.Log.Debug("Uint32 (negative) overflow, setting value to 0")
				return uint32(0)
			}
			return v - uint32(-n)
		}
		if n > math.Nextafter(float64(math.MaxUint32), 0) || uint32(n) > math.MaxUint32-v {
			p.Log.Debug("Uint32 overflow, setting value to MaxUint32")
			return uint32(math.MaxUint32)
		}
		return v + uint32(n)
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
