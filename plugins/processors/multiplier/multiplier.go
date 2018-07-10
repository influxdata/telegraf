package multiplier

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Multiplier struct {
	Config      []string
	VerboseMode bool

	isInitialized bool
	array         map[string]map[string]float64
}

var sampleConfig = `
  ## Config can contain multiply factors for each metrics.
  ## Each config line should be the string in influx format.
  Config = [
    "mem used_percent=100,available_percent=100",
    "swap used_percent=100"
  ]

  # VerboseMode allows to print changes for debug purpose
  VerboseMode = false
`

func (multiplier *Multiplier) SampleConfig() string {
	return sampleConfig
}

func (multiplier *Multiplier) Description() string {
	return "Multiply metrics values on some multiply factor"
}

func (multiplier *Multiplier) Apply(metricsArray ...telegraf.Metric) []telegraf.Metric {
	// Intialization should be only one time
	if !multiplier.isInitialized {
		multiplier.Initialize()
		multiplier.isInitialized = true
	}

	// Loop for all metrics
	for i, metrics := range metricsArray {

		// Check that even one metric should be multiplied
		if _, ok := multiplier.array[metrics.Name()]; ok == true {

			newFields := make(map[string]interface{})

			// Loop for specified metric
			for metricName, metricValue := range metrics.Fields() {

				newValue := metricValue

				// Check that current metric should be multiplied
				if factor, ok := multiplier.array[metrics.Name()][metricName]; ok == true {
					newValue = multiplier.Multiply(metricValue, factor)

					if multiplier.VerboseMode && metricValue != newValue {
						fmt.Printf("Multiplier: [%v.%v] %v * %v => %v\n",
							metrics.Name(), metricName, metricValue, factor, newValue)
					}
				}

				newFields[metricName] = newValue
			}

			newMetric, err := metric.New(metrics.Name(),
				metrics.Tags(), newFields, metrics.Time(), metrics.Type())

			if err != nil {
				fmt.Printf("Multiplier: Cannot make a copy: %v\n", err)
			} else {
				metricsArray[i] = newMetric
			}
		}
	}

	return metricsArray
}

func (multiplier *Multiplier) Multiply(value interface{}, factor float64) interface{} {
	switch data := value.(type) {
	case int:
		return int(factor * float64(data))
	case uint:
		return uint(factor * float64(data))
	case int32:
		return int32(factor * float64(data))
	case uint32:
		return uint32(factor * float64(data))
	case int64:
		return int64(factor * float64(data))
	case uint64:
		return uint64(factor * float64(data))
	case float32:
		return float32(factor * float64(data))
	case float64:
		return float64(factor * float64(data))
	default:
		fmt.Printf("Multiplier plugin couldn't multiply %v [float64] with value: %T '%v'\n",
			factor, value, data)
	}
	return value
}

func toFloat(value interface{}) float64 {
	switch data := value.(type) {
	case int:
		return float64(data)
	case int32:
		return float64(data)
	case int64:
		return float64(data)
	case float32:
		return float64(data)
	case float64:
		return data
	default:
		fmt.Printf("Multiplier plugin couldn't create 'float64' from value: %T '%v'\n",
			value, data)
	}
	return 0
}

func (multiplier *Multiplier) Initialize() error {
	fmt.Printf("Multiplier Config: \n  VerboseMode: %v\n  Config: %v\n",
		multiplier.VerboseMode, multiplier.Config)

	multiplier.array = make(map[string]map[string]float64)

	for _, str := range multiplier.Config {
		parser, _ := parsers.NewInfluxParser()
		metrics, err := parser.ParseLine(str)
		if err != nil {
			fmt.Printf("E! %v\n", err)
			continue
		}

		keeper, ok := multiplier.array[metrics.Name()]
		if !ok {
			keeper = make(map[string]float64)
			multiplier.array[metrics.Name()] = keeper
		}

		for metricName, _metricValue := range metrics.Fields() {
			metricValue := toFloat(_metricValue)
			keeper[metricName] = metricValue
			fmt.Printf("  Multiplication: [%v.%v] * %v\n",
				metrics.Name(), metricName, metricValue)
		}
	}

	return nil
}

func newMultiplier() *Multiplier {
	multiplier := &Multiplier{}
	return multiplier
}

func init() {
	processors.Add("multiplier", func() telegraf.Processor {
		return newMultiplier()
	})
}
