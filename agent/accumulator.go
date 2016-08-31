package agent

import (
	"fmt"
	"log"
	"math"
	"sync/atomic"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/models"
)

func NewAccumulator(
	inputConfig *models.InputConfig,
	metrics chan telegraf.Metric,
) *accumulator {
	acc := accumulator{}
	acc.metrics = metrics
	acc.inputConfig = inputConfig
	acc.precision = time.Nanosecond
	return &acc
}

type accumulator struct {
	metrics chan telegraf.Metric

	defaultTags map[string]string

	debug bool
	// print every point added to the accumulator
	trace bool

	inputConfig *models.InputConfig

	precision time.Duration

	errCount uint64
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

	if !ac.inputConfig.Filter.ShouldNamePass(measurement) {
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
	ac.inputConfig.Filter.FilterTags(tags)

	result := make(map[string]interface{})
	for k, v := range fields {
		// Filter out any filtered fields
		if ac.inputConfig != nil {
			if !ac.inputConfig.Filter.ShouldFieldsPass(k) {
				continue
			}
		}

		// Validate uint64 and float64 fields
		switch val := v.(type) {
		case uint64:
			// InfluxDB does not support writing uint64
			if val < uint64(9223372036854775808) {
				result[k] = int64(val)
			} else {
				result[k] = int64(9223372036854775807)
			}
			continue
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

		result[k] = v
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
	timestamp = timestamp.Round(ac.precision)

	m, err := telegraf.NewMetric(measurement, tags, result, timestamp)
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return
	}
	if ac.trace {
		fmt.Println("> " + m.String())
	}
	ac.metrics <- m
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	atomic.AddUint64(&ac.errCount, 1)
	//TODO suppress/throttle consecutive duplicate errors?
	log.Printf("ERROR in input [%s]: %s", ac.inputConfig.Name, err)
}

func (ac *accumulator) Debug() bool {
	return ac.debug
}

func (ac *accumulator) SetDebug(debug bool) {
	ac.debug = debug
}

func (ac *accumulator) Trace() bool {
	return ac.trace
}

func (ac *accumulator) SetTrace(trace bool) {
	ac.trace = trace
}

// SetPrecision takes two time.Duration objects. If the first is non-zero,
// it sets that as the precision. Otherwise, it takes the second argument
// as the order of time that the metrics should be rounded to, with the
// maximum being 1s.
func (ac *accumulator) SetPrecision(precision, interval time.Duration) {
	if precision > 0 {
		ac.precision = precision
		return
	}
	switch {
	case interval >= time.Second:
		ac.precision = time.Second
	case interval >= time.Millisecond:
		ac.precision = time.Millisecond
	case interval >= time.Microsecond:
		ac.precision = time.Microsecond
	default:
		ac.precision = time.Nanosecond
	}
}

func (ac *accumulator) DisablePrecision() {
	ac.precision = time.Nanosecond
}

func (ac *accumulator) setDefaultTags(tags map[string]string) {
	ac.defaultTags = tags
}

func (ac *accumulator) addDefaultTag(key, value string) {
	if ac.defaultTags == nil {
		ac.defaultTags = make(map[string]string)
	}
	ac.defaultTags[key] = value
}
