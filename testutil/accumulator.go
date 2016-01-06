package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Point defines a single point measurement
type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

func (p *Point) String() string {
	return fmt.Sprintf("%s %v", p.Measurement, p.Fields)
}

// Accumulator defines a mocked out accumulator
type Accumulator struct {
	sync.Mutex

	Points []*Point
	debug  bool
}

// Add adds a measurement point to the accumulator
func (a *Accumulator) Add(
	measurement string,
	value interface{},
	tags map[string]string,
	t ...time.Time,
) {
	fields := map[string]interface{}{"value": value}
	a.AddFields(measurement, fields, tags, t...)
}

// AddFields adds a measurement point with a specified timestamp.
func (a *Accumulator) AddFields(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	timestamp ...time.Time,
) {
	a.Lock()
	defer a.Unlock()
	if tags == nil {
		tags = map[string]string{}
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

	p := &Point{
		Measurement: measurement,
		Fields:      fields,
		Tags:        tags,
		Time:        t,
	}

	a.Points = append(a.Points, p)
}

func (a *Accumulator) SetDefaultTags(tags map[string]string) {
	// stub for implementing Accumulator interface.
}

func (a *Accumulator) AddDefaultTag(key, value string) {
	// stub for implementing Accumulator interface.
}

func (a *Accumulator) Prefix() string {
	// stub for implementing Accumulator interface.
	return ""
}

func (a *Accumulator) SetPrefix(prefix string) {
	// stub for implementing Accumulator interface.
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
func (a *Accumulator) Get(measurement string) (*Point, bool) {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			return p, true
		}
	}

	return nil, false
}

// NFields returns the total number of fields in the accumulator, across all
// measurements
func (a *Accumulator) NFields() int {
	counter := 0
	for _, pt := range a.Points {
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
	for _, p := range a.Points {
		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}

		if p.Measurement == measurement {
			if !reflect.DeepEqual(fields, p.Fields) {
				pActual, _ := json.MarshalIndent(p.Fields, "", "  ")
				pExp, _ := json.MarshalIndent(fields, "", "  ")
				msg := fmt.Sprintf("Actual:\n%s\n(%T) \nExpected:\n%s\n(%T)",
					string(pActual), p.Fields, string(pExp), fields)
				assert.Fail(t, msg)
			}
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
	for _, p := range a.Points {
		if p.Measurement == measurement {
			if !reflect.DeepEqual(fields, p.Fields) {
				pActual, _ := json.MarshalIndent(p.Fields, "", "  ")
				pExp, _ := json.MarshalIndent(fields, "", "  ")
				msg := fmt.Sprintf("Actual:\n%s\n(%T) \nExpected:\n%s\n(%T)",
					string(pActual), p.Fields, string(pExp), fields)
				assert.Fail(t, msg)
			}
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s", measurement)
	assert.Fail(t, msg)
}

// HasIntValue returns true if the measurement has an Int value
func (a *Accumulator) HasIntField(measurement string, field string) bool {
	for _, p := range a.Points {
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
	for _, p := range a.Points {
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
	for _, p := range a.Points {
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
	for _, p := range a.Points {
		if p.Measurement == measurement {
			return true
		}
	}
	return false
}
