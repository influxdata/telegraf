package noise

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
	"golang.org/x/exp/rand"
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
	defaultNoiseLog  = false
)

type Distribution struct {
	active   ActiveDistribution
	laplace  distuv.Laplace
	gaussian distuv.Normal
	uniform  distuv.Uniform
}

var noise float64
var fieldFilter filter.Filter
var sampleConfig = `
  [[processors.noise]]
    scale = 1.0
	min = 1.0
	max = 5.0
	mu = 0.0
	sigma = 2.0
	noise_type = "laplace"
	noise_log = false
    include_fields = []
	exclude_fields = []
`

type Noise struct {
	Generator     Distribution
	Scale         float64         `toml:"scale"`
	Min           float64         `toml:"min"`
	Max           float64         `toml:"max"`
	Mu            float64         `toml:"mu"`
	Sigma         float64         `toml:"sigma"`
	IncludeFields []string        `toml:"include_fields"`
	ExcludeFields []string        `toml:"exclude_fields"`
	NoiseType     string          `toml:"noise_type"`
	NoiseLog      bool            `toml:"noise_log"`
	Log           telegraf.Logger `toml:"-"`
}

// returns a random float value, with a probability densitity of the current
// active distribution
func (d Distribution) getNoise() float64 {
	switch d.active {
	case Uniform:
		return d.uniform.Rand()
	case Gaussian:
		return d.gaussian.Rand()
	default:
		return d.laplace.Rand()
	}
}

func (p *Noise) SampleConfig() string {
	return sampleConfig
}

func (p *Noise) Description() string {
	return "Adds noise to numerical fields"
}

// generates a random noise value depending on the defined probability density
// function and adds that to the original value
func (p *Noise) addNoise(value interface{}) interface{} {
	noise = p.Generator.getNoise()
	switch v := value.(type) {
	case int64:
		return int64(float64(v) + float64(v)*noise)
	case uint64:
		newV := float64(v) + float64(v)*noise
		if newV < 0 {
			return uint64(0)
		}
		return uint64(newV)
	case float64:
		return (v + v*noise)
	default:
		p.Log.Debugf("Value (%s) type invalid: [%s] is not an int64, uint64 or float64", v, value)
	}
	return value
}

// Creates a filter for Include and Exclude fields and sets the desired noise distribution
// BUG(wizarq): according to the logs, this function is called twice. Why?
func (p *Noise) Init() error {
	rand.Seed(uint64(time.Now().UnixNano()))
	fieldFilter, _ = filter.NewIncludeExcludeFilter(p.IncludeFields, p.ExcludeFields)
	p.Generator = Distribution{
		laplace:  distuv.Laplace{Mu: defaultMu, Scale: defaultScale, Src: nil},
		gaussian: distuv.Normal{Mu: defaultMu, Sigma: defaultSigma, Src: nil},
		uniform:  distuv.Uniform{Min: defaultMin, Max: defaultMax, Src: nil},
	}
	switch p.NoiseType {
	case "uniform":
		p.Generator.active = Uniform
	case "gaussian":
		p.Generator.active = Gaussian
	default:
		p.Generator.active = Laplace
	}
	return nil
}

func (p *Noise) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		p.Log.Debugf("Adding noise to [%s]", metric.Name())
		for key, value := range metric.Fields() {
			if !fieldFilter.Match(key) {
				continue
			}
			newVal := p.addNoise(value)
			metric.RemoveField(key)
			metric.AddField(key, newVal)
			if p.NoiseLog {
				metric.AddField("telegraf_noise", noise)
			}
		}
	}
	return metrics
}

func init() {
	processors.Add("noise", func() telegraf.Processor {
		return &Noise{
			NoiseType:     defaultNoiseType,
			NoiseLog:      defaultNoiseLog,
			Scale:         defaultScale,
			IncludeFields: []string{},
			ExcludeFields: []string{},
		}
	})
}
