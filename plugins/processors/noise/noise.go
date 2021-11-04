package noise

import (
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"gonum.org/v1/gonum/stat/distuv"
)

const (
	defaultSensitivity = 1.0
	defaultEpsilon     = 1.0
)

var sampleConfig = `
  [[processors.noise]]
    sensitivity = 1.0
    epsilon = 1.0
    ignore_fields = []
    ignore_measurements = []
`

var laplaceValue = 0.0
var fieldExcludeSet = make(map[string]bool)
var measurementExcludeSet = make(map[string]bool)

type Noise struct {
	Sensitivity        float64
	Epsilon            float64
	IgnoreFields       []string        `toml:"ignore_fields"`
	IgnoreMeasurements []string        `toml:"ignore_measurements"`
	Log                telegraf.Logger `toml:"-"`
}

func (p *Noise) SampleConfig() string {
	return sampleConfig
}

func (p *Noise) Description() string {
	return "Generates noise based on the Laplace distribution and add that to " +
		"all numerical field values. Exclusions can be listed in the config file."
}

// Gets a random, positive Integer from a Laplace distribution and adds it to a
// given value
func (p *Noise) addLaplaceUInt64(value uint64) uint64 {
	return value + value*uint64(math.Round(math.Abs(p.getRandomLaplaceNoise())))
}

// Gets a random Integer from a Laplace distribution and adds it to a
// given value
func (p *Noise) addLaplaceInt64(value int64) int64 {
	return value + value*int64(math.Round(math.Abs(p.getRandomLaplaceNoise())))
}

// Gets a random, positive float from a Laplace distribution and adds it to a
// given value
func (p *Noise) addLaplaceFloat64(value float64) float64 {
	return value + value*math.Abs(p.getRandomLaplaceNoise())
}

// Returns a random float from a Laplace distribution. If nil is passed, a
// random seed will be generated internally
func (p *Noise) getRandomLaplaceNoise() float64 {
	l := distuv.Laplace{
		Mu:    0,
		Scale: p.Sensitivity / p.Epsilon,
		Src:   nil,
	}
	laplaceValue = l.Rand()
	return laplaceValue
}

// Takes a value as interface and adds laplace noise to any given numerical type
func (p *Noise) addNoiseToValue(value interface{}) interface{} {
	switch v := value.(type) {
	case int64:
		value = p.addLaplaceInt64(value.(int64))
	case uint64:
		value = p.addLaplaceUInt64(value.(uint64))
	case float64:
		value = p.addLaplaceFloat64(value.(float64))
	default:
		p.Log.Debugf("Value (%s) type invalid: [%s] is not an int64, uint64 or float64", v, value)
	}
	return value
}

// Iterates through all fields and calls addNoiseToValue() for each value not
// in the exclude set.
func (p *Noise) addNoiseToMetric(metric telegraf.Metric) {
	for key, value := range metric.Fields() {
		/* check for ignore fields */
		if _, ok := fieldExcludeSet[key]; ok {
			continue
		}
		value = p.addNoiseToValue(value)
		metric.RemoveField(key)
		metric.AddField(key, value)
	}
}

// Convertes the IgnoreFields and IgnoreMeasurements to maps, so no looping is necessary later on.
// BUG(wizarq): according to the logs, this function is called twice. Why?
func (p *Noise) Init() error {
	p.Log.Debugf("Creating filters for fields and measurements")
	for _, filter := range p.IgnoreFields {
		fieldExcludeSet[filter] = true
	}
	for _, filter := range p.IgnoreMeasurements {
		p.Log.Debugf("Adding [%s] to IgnoreMeasurementFilter", filter)
		measurementExcludeSet[filter] = true
	}
	return nil
}

func (p *Noise) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		/* check for ignore fields */
		if _, ok := measurementExcludeSet[metric.Name()]; ok {
			continue
		}
		p.addNoiseToMetric(metric)
	}
	return metrics
}

func init() {
	processors.Add("noise", func() telegraf.Processor {
		return &Noise{
			Sensitivity: defaultSensitivity,
			Epsilon:     defaultEpsilon,
		}
	})
}
