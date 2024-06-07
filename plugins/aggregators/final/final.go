//go:generate ../../../tools/readme_config_includer/generator
package final

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

//go:embed sample.conf
var sampleConfig string

type Final struct {
	OutputStrategy         string          `toml:"output_strategy"`
	SeriesTimeout          config.Duration `toml:"series_timeout"`
	KeepOriginalFieldNames bool            `toml:"keep_original_field_names"`

	// The last metric for all series which are active
	metricCache map[uint64]telegraf.Metric
}

func NewFinal() *Final {
	return &Final{
		SeriesTimeout: config.Duration(5 * time.Minute),
	}
}

func (*Final) SampleConfig() string {
	return sampleConfig
}

func (m *Final) Init() error {
	// Check options and set defaults
	switch m.OutputStrategy {
	case "":
		m.OutputStrategy = "timeout"
	case "timeout", "periodic":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("invalid 'output_strategy': %q", m.OutputStrategy)
	}

	// Initialize the cache
	m.metricCache = make(map[uint64]telegraf.Metric)

	return nil
}

func (m *Final) Add(in telegraf.Metric) {
	id := in.HashID()
	m.metricCache[id] = in
}

func (m *Final) Push(acc telegraf.Accumulator) {
	// Preserve timestamp of original metric
	acc.SetPrecision(time.Nanosecond)

	for id, metric := range m.metricCache {
		if m.OutputStrategy == "timeout" && time.Since(metric.Time()) <= time.Duration(m.SeriesTimeout) {
			// We output on timeout but the last metric of the series was
			// younger than that. So skip the output for this period.
			continue
		}
		var fields map[string]any
		if m.KeepOriginalFieldNames {
			fields = metric.Fields()
		} else {
			fields = map[string]any{}
			for _, field := range metric.FieldList() {
				fields[field.Key+"_final"] = field.Value
			}
		}

		acc.AddFields(metric.Name(), fields, metric.Tags(), metric.Time())
		delete(m.metricCache, id)
	}
}

func (m *Final) Reset() {
}

func init() {
	aggregators.Add("final", func() telegraf.Aggregator {
		return NewFinal()
	})
}
