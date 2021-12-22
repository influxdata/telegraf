package mock

import (
	"math"
	"math/rand"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Mock struct {
	counter int64

	MetricName string            `json:"metric_name"`
	Tags       map[string]string `json:"tags"`

	Random   []*Random   `json:"random_float"`
	Step     []*Step     `json:"step"`
	Stock    []*Stock    `json:"stock"`
	SineWave []*SineWave `json:"sine_wave"`
}

type Random struct {
	Name string  `json:"name"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

type SineWave struct {
	Name      string  `json:"name"`
	Amplitude float64 `json:"amplitude"`
}

type Step struct {
	latest float64

	Name  string  `json:"name"`
	Start float64 `json:"min"`
	Step  float64 `json:"max"`
}

type Stock struct {
	latest float64

	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Volatility float64 `json:"volatility"`
}

const SampleConfig = `
  ## Set the metric name to use for reporting
  metric_name = "mock"

  ## Optional string key-value pairs of tags to add to all metrics
  # [inputs.mock.tags]
  # "key" = "value"

  ## One or more mock data fields *must* be defined.
  ##
  ## [[inputs.mock.random_float]]
  ##   name = "rand"
  ##   min = 1.0
  ##   max = 6.0
  ## [[inputs.mock.sine_wave]]
  ##   name = "wave"
  ##   amplitude = 10.0
  ## [[inputs.mock.step]]
  ##   name = "plus_one"
  ##   start = 0.0
  ##   step = 1.0
  ## [[inputs.mock.stock]]
  ##   name = "abc"
  ##   price = 50.00
  ##   volatility = 0.2
`

func (m *Mock) SampleConfig() string {
	return SampleConfig
}

func (m *Mock) Description() string {
	return "Generate metrics for test and demonstration purposes"
}

func (m *Mock) Init() error {
	rand.Seed(time.Now().UnixNano())
	return nil
}

func (m *Mock) Gather(acc telegraf.Accumulator) error {
	fields := make(map[string]interface{})
	m.generateRandomFloat64(fields)
	m.generateStockPrice(fields)
	m.generateSineWave(fields)
	m.generateStep(fields)

	tags := make(map[string]string)
	for key, value := range m.Tags {
		tags[key] = value
	}

	acc.AddFields(m.MetricName, fields, tags)

	m.counter++

	return nil
}

// Generate random value between min and max, inclusivly
func (m *Mock) generateRandomFloat64(fields map[string]interface{}) {
	for _, random := range m.Random {
		fields[random.Name] = random.Min + rand.Float64()*(random.Max-random.Min)
	}
}

// Create sine waves
func (m *Mock) generateSineWave(fields map[string]interface{}) {
	for _, field := range m.SineWave {
		fields[field.Name] = math.Sin((float64(m.counter)*math.Pi)/5.0) * field.Amplitude
	}
}

// Begin at start value and then add step value every tick
func (m *Mock) generateStep(fields map[string]interface{}) {
	for _, step := range m.Step {
		if m.counter == 0 {
			step.latest = step.Start
		} else {
			step.latest += step.Step
		}

		fields[step.Name] = step.latest
	}
}

// Begin at start price and then generate random value
func (m *Mock) generateStockPrice(fields map[string]interface{}) {
	for _, stock := range m.Stock {
		if stock.latest == 0.0 {
			stock.latest = stock.Price
		} else {
			noise := 2 * (rand.Float64() - 0.5)
			stock.latest = stock.latest + (stock.latest * stock.Volatility * noise)

			// avoid going below zero
			if stock.latest < 1.0 {
				stock.latest = 1.0
			}
		}

		fields[stock.Name] = stock.latest
	}
}

func init() {
	inputs.Add("mock", func() telegraf.Input {
		return &Mock{}
	})
}
