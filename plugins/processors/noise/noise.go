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

type ActiveDistribution int

const (
	Laplace ActiveDistribution = iota
	Gaussian
	Uniform
)

const (
	defaultScale     = 5.0
	defaultMin       = -1.0
	defaultMax       = 1.0
	defaultMu        = 0.0
	defaultSigma     = 0.1
	defaultNoiseType = "laplace"
)

const sampleConfig = `
  [[processors.noise]]
    ## Specified the type of the random distribution.
    ## Can be "laplace", "gaussian" or "uniform".
    # type = "laplace

    ## Center of the distribution.
    ## Only used for "laplacian" and "gaussian" distributions.
    # mu = 0.0

    ## Scale parameter of the Laplacian distribution
    # scale = 1.0

    ## Standard deviation of the Gaussian distribution
    # sigma = 0.1

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
	Sigma         float64         `toml:"sigma"`
	IncludeFields []string        `toml:"include_fields"`
	ExcludeFields []string        `toml:"exclude_fields"`
	NoiseType     string          `toml:"type"`
	Log           telegraf.Logger `toml:"-"`
	Generator     distuv.Rander
	FieldFilter   filter.Filter
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
	noise := p.Generator.Rand()
	if value == 0 {
		return noise
	}
	switch v := value.(type) {
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		n := float64(v) + float64(v)*noise
		if n > float64(math.MaxInt64) {
			p.Log.Debug("Int64 overflow, setting value to MaxInt64")
			return int64(math.MaxInt64)
		}
		if n < float64(math.MinInt64) {
			p.Log.Debug("Int64 (negative) overflow, setting value to MinInt64")
			return int64(math.MinInt64)
		}
		return int64(n)
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		n := float64(v) + float64(v)*noise
		if n > float64(math.MaxUint64) {
			p.Log.Debug("UInt64 overflow, setting value to MaxInt64")
			return uint64(math.MaxUint64)
		}
		if n < 0 {
			p.Log.Debug("UInt64 (negative) overflow, setting value to 0")
			return uint64(0)
		}
		return uint64(n)
	case float32:
	case float64:
		return (v + v*noise)
	default:
		p.Log.Debugf("Value (%v) type invalid: [%v] is not an int, uint or float", v, reflect.TypeOf(value))
	}
	return value
}

// Creates a filter for Include and Exclude fields and sets the desired noise
// distribution
// BUG(wizarq): according to the logs, this function is called twice. Why?
func (p *Noise) Init() error {
	fieldFilter, err := filter.NewIncludeExcludeFilter(p.IncludeFields, p.ExcludeFields)
	if err != nil {
		return fmt.Errorf("creating fieldFilter failed: %v", err)
	}
	p.FieldFilter = fieldFilter

	switch p.NoiseType {
	case "", "laplace":
		p.Generator = &distuv.Laplace{Mu: p.Mu, Scale: p.Scale}
	case "uniform":
		p.Generator = &distuv.Uniform{Min: p.Min, Max: p.Max}
	case "gaussian":
		p.Generator = &distuv.Normal{Mu: p.Mu, Sigma: p.Sigma}
	default:
		return fmt.Errorf("unknown distribution type %q", p.NoiseType)
	}
	return nil
}

func (p *Noise) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		p.Log.Debugf("Adding noise to [%s]", metric.Name())
		for _, field := range metric.FieldList() {
			if !p.FieldFilter.Match(field.Key) {
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
			Sigma:     defaultSigma,
			Min:       defaultMin,
			Max:       defaultMax,
		}
	})
}
