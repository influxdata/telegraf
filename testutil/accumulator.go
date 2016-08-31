package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Metric defines a single point measurement
type Metric struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

func (p *Metric) String() string {
	return fmt.Sprintf("%s %v", p.Measurement, p.Fields)
}

// Accumulator defines a mocked out accumulator
type Accumulator struct {
	sync.Mutex

	Metrics  []*Metric
	nMetrics uint64
	Discard  bool
	Errors   []error
	debug    bool
}

func (a *Accumulator) NMetrics() uint64 {
	return atomic.LoadUint64(&a.nMetrics)
}

// AddFields adds a measurement point with a specified timestamp.
func (a *Accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	atomic.AddUint64(&a.nMetrics, 1)
	if a.Discard {
		return
	}
	a.Lock()
	defer a.Unlock()
	if tags == nil {
		tags = map[string]string{}
	}

	if len(fields) == 0 {
		return
	}

	var t time.Time
	if len(timestamp) > 0 {
		t = timestamp[0]
	} else {
		t = time.Now()
	}

	if a.debug {
		pretty, _ := json.MarshalIndent(fields, "", "  ")
		prettyTags, _ := json.MarshalIndent(tags, "", "  ")
		msg := fmt.Sprintf("Adding Measurement [%s]\nFields:%s\nTags:%s\n",
			measurement, string(pretty), string(prettyTags))
		fmt.Print(msg)
	}

	p := &Metric{
		Measurement: measurement,
		Fields:      fields,
		Tags:        tags,
		Time:        t,
	}

	a.Metrics = append(a.Metrics, p)
}

// AddError appends the given error to Accumulator.Errors.
func (a *Accumulator) AddError(err error) {
	if err == nil {
		return
	}
	a.Lock()
	a.Errors = append(a.Errors, err)
	a.Unlock()
}

func (a *Accumulator) SetPrecision(precision, interval time.Duration) {
	return
}

func (a *Accumulator) DisablePrecision() {
	return
}

func (a *Accumulator) Debug() bool {
	// stub for implementing Accumulator interface.
	return a.debug
}

func (a *Accumulator) SetDebug(debug bool) {
	// stub for implementing Accumulator interface.
	a.debug = debug
}

// Get gets the specified measurement point from the accumulator
func (a *Accumulator) Get(measurement string) (*Metric, bool) {
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			return p, true
		}
	}

	return nil, false
}

// NFields returns the total number of fields in the accumulator, across all
// measurements
func (a *Accumulator) NFields() int {
	a.Lock()
	defer a.Unlock()
	counter := 0
	for _, pt := range a.Metrics {
		for _, _ = range pt.Fields {
			counter++
		}
	}
	return counter
}

func (a *Accumulator) AssertContainsTaggedFields(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}

		if p.Measurement == measurement {
			assert.Equal(t, fields, p.Fields)
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s with tags %v", measurement, tags)
	assert.Fail(t, msg)
}

func (a *Accumulator) AssertContainsFields(
	t *testing.T,
	measurement string,
	fields map[string]interface{},
) {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			assert.Equal(t, fields, p.Fields)
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s", measurement)
	assert.Fail(t, msg)
}

// HasIntValue returns true if the measurement has an Int value
func (a *Accumulator) HasIntField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(int64)
					return ok
				}
			}
		}
	}

	return false
}

// HasUIntValue returns true if the measurement has a UInt value
func (a *Accumulator) HasUIntField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(uint64)
					return ok
				}
			}
		}
	}

	return false
}

// HasFloatValue returns true if the given measurement has a float value
func (a *Accumulator) HasFloatField(measurement string, field string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			for fieldname, value := range p.Fields {
				if fieldname == field {
					_, ok := value.(float64)
					return ok
				}
			}
		}
	}

	return false
}

// HasMeasurement returns true if the accumulator has a measurement with the
// given name
func (a *Accumulator) HasMeasurement(measurement string) bool {
	a.Lock()
	defer a.Unlock()
	for _, p := range a.Metrics {
		if p.Measurement == measurement {
			return true
		}
	}
	return false
}
