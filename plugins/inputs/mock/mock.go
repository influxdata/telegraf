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

	MetricName string            `toml:"metric_name"`
	Tags       map[string]string `toml:"tags"`

	Random   []*random   `toml:"random"`
	Step     []*step     `toml:"step"`
	Stock    []*stock    `toml:"stock"`
	SineWave []*sineWave `toml:"sine_wave"`
}

type random struct {
	Name string  `toml:"name"`
	Min  float64 `toml:"min"`
	Max  float64 `toml:"max"`
}

type sineWave struct {
	Name      string  `toml:"name"`
	Amplitude float64 `toml:"amplitude"`
	Period    float64 `toml:"period"`
}

type step struct {
	latest float64

	Name  string  `toml:"name"`
	Start float64 `toml:"min"`
	Step  float64 `toml:"max"`
}

type stock struct {
	latest float64

	Name       string  `toml:"name"`
	Price      float64 `toml:"price"`
	Volatility float64 `toml:"volatility"`
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
		fields[field.Name] = math.Sin((float64(m.counter) * field.Period * math.Pi)) * field.Amplitude
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
