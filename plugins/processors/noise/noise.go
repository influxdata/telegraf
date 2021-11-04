package noise

import (
	"time"

	"golang.org/x/exp/rand"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	defaultScale = 1.0
)

var fieldFilter filter.Filter
var sampleConfig = `
  [[processors.noise]]
    scale = 1.0
    include_fields = []
	exclude_fields = []
`

type Noise struct {
	Scale         float64         `toml:"scale"`
	IncludeFields []string        `toml:"include_fields"`
	ExcludeFields []string        `toml:"exclude_fields"`
	Log           telegraf.Logger `toml:"-"`
}

func (p *Noise) SampleConfig() string {
	return sampleConfig
}

func (p *Noise) Description() string {
	return "Adds noise to numerical fields"
}

// Gets a random, positive Integer from a Laplace distribution and adds it to a
// given value
func (p *Noise) addLaplaceUInt64(value uint64) uint64 {
	return value + value*uint64(p.getRandomLaplaceNoise())
}

// Gets a random Integer from a Laplace distribution and adds it to a
// given value
func (p *Noise) addLaplaceInt64(value int64) int64 {
	return value + value*int64(p.getRandomLaplaceNoise())
}

// Gets a random, positive float from a Laplace distribution and adds it to a
// given value
func (p *Noise) addLaplaceFloat64(value float64) float64 {
	return value + value*p.getRandomLaplaceNoise()
}

// Returns a random float from a Laplace distribution.
func (p *Noise) getRandomLaplaceNoise() float64 {
	rand.Seed(uint64(time.Now().UnixNano()))
	l := distuv.Laplace{
		Mu:    0,
		Scale: p.Scale,
		Src:   nil,
	}
	return l.Rand()
}

// Takes a value as interface and adds laplace noise to any given numerical type
func (p *Noise) addNoiseToValue(value interface{}) interface{} {
	switch v := value.(type) {
	case int64:
		value = p.addLaplaceInt64(v)
	case uint64:
		value = p.addLaplaceUInt64(v)
	case float64:
		value = p.addLaplaceFloat64(v)
	default:
		p.Log.Debugf("Value (%s) type invalid: [%s] is not an int64, uint64 or float64", v, value)
	}
	return value
}

// Creates a filter for Include and Exclude fields
// BUG(wizarq): according to the logs, this function is called twice. Why?
func (p *Noise) Init() error {
	fieldFilter, _ = filter.NewIncludeExcludeFilter(p.IncludeFields, p.ExcludeFields)
	return nil
}

func (p *Noise) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		p.Log.Debugf("Adding noise to [%s]", metric.Name())
		for key, value := range metric.Fields() {
			if !fieldFilter.Match(key) {
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, p.addNoiseToValue(value))
		}
	}
	return metrics
}

func init() {
	processors.Add("noise", func() telegraf.Processor {
		return &Noise{
			Scale:         defaultScale,
			IncludeFields: []string{},
			ExcludeFields: []string{},
		}
	})
}
