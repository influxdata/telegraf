package agent

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/models"
)

func NewAccumulator(
	inputConfig *internal_models.InputConfig,
	metrics chan telegraf.Metric,
	interval time.Duration,
) *accumulator {
	acc := accumulator{}
	acc.metrics = metrics
	acc.inputConfig = inputConfig
	// at what interval are these being accumulated
	acc.interval = interval
	return &acc
}

type accumulator struct {
	sync.Mutex

	metrics chan telegraf.Metric

	defaultTags map[string]string

	debug bool

	inputConfig *internal_models.InputConfig

	prefix string

	interval time.Duration
}

func (ac *accumulator) Add(
	measurement string,
	value interface{},
	tags map[string]string,
	t ...time.Time,
) {
	fields := make(map[string]interface{})
	fields["value"] = value
	ac.AddFields(measurement, fields, tags, t...)
}

func (ac *accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	t ...time.Time,
) {
	if len(fields) == 0 || len(measurement) == 0 {
		return
	}

	if !ac.inputConfig.Filter.ShouldTagsPass(tags) {
		return
	}

	// Override measurement name if set
	if len(ac.inputConfig.NameOverride) != 0 {
		measurement = ac.inputConfig.NameOverride
	}
	// Apply measurement prefix and suffix if set
	if len(ac.inputConfig.MeasurementPrefix) != 0 {
		measurement = ac.inputConfig.MeasurementPrefix + measurement
	}
	if len(ac.inputConfig.MeasurementSuffix) != 0 {
		measurement = measurement + ac.inputConfig.MeasurementSuffix
	}

	if tags == nil {
		tags = make(map[string]string)
	}
	// Apply plugin-wide tags if set
	for k, v := range ac.inputConfig.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	// Apply daemon-wide tags if set
	for k, v := range ac.defaultTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	result := make(map[string]interface{})
	for k, v := range fields {
		// Filter out any filtered fields
		if ac.inputConfig != nil {
			if !ac.inputConfig.Filter.ShouldPass(k) {
				continue
			}
		}
		result[k] = v

		// Validate uint64 and float64 fields
		switch val := v.(type) {
		case uint64:
			// InfluxDB does not support writing uint64
			if val < uint64(9223372036854775808) {
				result[k] = int64(val)
			} else {
				result[k] = int64(9223372036854775807)
			}
		case float64:
			// NaNs are invalid values in influxdb, skip measurement
			if math.IsNaN(val) || math.IsInf(val, 0) {
				if ac.debug {
					log.Printf("Measurement [%s] field [%s] has a NaN or Inf "+
						"field, skipping",
						measurement, k)
				}
				continue
			}
		}
	}
	fields = nil
	if len(result) == 0 {
		return
	}

	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}

	if ac.prefix != "" {
		measurement = ac.prefix + measurement
	}

	m, err := telegraf.NewMetric(measurement, tags, result, timestamp)
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return
	}
	m.SetInterval(ac.interval)
	if ac.debug {
		fmt.Println("> " + m.String())
	}
	ac.metrics <- m
}

func (ac *accumulator) Debug() bool {
	return ac.debug
}

func (ac *accumulator) SetDebug(debug bool) {
	ac.debug = debug
}

func (ac *accumulator) setDefaultTags(tags map[string]string) {
	ac.defaultTags = tags
}

func (ac *accumulator) addDefaultTag(key, value string) {
	ac.defaultTags[key] = value
}
