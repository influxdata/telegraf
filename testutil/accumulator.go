package testutil

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Point defines a single point measurement
type Point struct {
	Measurement string
	Tags        map[string]string
	Values      map[string]interface{}
	Time        time.Time
}

// Accumulator defines a mocked out accumulator
type Accumulator struct {
	sync.Mutex
	Points []*Point
}

// Add adds a measurement point to the accumulator
func (a *Accumulator) Add(measurement string, value interface{}, tags map[string]string) {
	a.Lock()
	defer a.Unlock()
	if tags == nil {
		tags = map[string]string{}
	}
	a.Points = append(
		a.Points,
		&Point{
			Measurement: measurement,
			Values:      map[string]interface{}{"value": value},
			Tags:        tags,
		},
	)
}

// AddFieldsWithTime adds a measurement point with a specified timestamp.
func (a *Accumulator) AddFieldsWithTime(
	measurement string,
	values map[string]interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	a.Points = append(
		a.Points,
		&Point{
			Measurement: measurement,
			Values:      values,
			Tags:        tags,
			Time:        timestamp,
		},
	)
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

// CheckValue checks that the accumulators point for the given measurement
// is the same as the given value.
func (a *Accumulator) CheckValue(measurement string, val interface{}) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			return p.Values["value"] == val
		}
	}
	fmt.Printf("CheckValue failed, measurement %s, value %s", measurement, val)
	return false
}

// CheckTaggedValue calls ValidateTaggedValue
func (a *Accumulator) CheckTaggedValue(
	measurement string,
	val interface{},
	tags map[string]string,
) bool {
	return a.ValidateTaggedValue(measurement, val, tags) == nil
}

// ValidateTaggedValue validates that the given measurement and value exist
// in the accumulator and with the given tags.
func (a *Accumulator) ValidateTaggedValue(
	measurement string,
	val interface{},
	tags map[string]string,
) error {
	if tags == nil {
		tags = map[string]string{}
	}
	for _, p := range a.Points {
		if !reflect.DeepEqual(tags, p.Tags) {
			continue
		}

		if p.Measurement == measurement {
			if p.Values["value"] != val {
				return fmt.Errorf("%v (%T) != %v (%T)",
					p.Values["value"], p.Values["value"], val, val)
			}
			return nil
		}
	}

	return fmt.Errorf("unknown measurement %s with tags %v", measurement, tags)
}

// ValidateValue calls ValidateTaggedValue
func (a *Accumulator) ValidateValue(measurement string, val interface{}) error {
	return a.ValidateTaggedValue(measurement, val, nil)
}

// HasIntValue returns true if the measurement has an Int value
func (a *Accumulator) HasIntValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Values["value"].(int64)
			return ok
		}
	}

	return false
}

// HasUIntValue returns true if the measurement has a UInt value
func (a *Accumulator) HasUIntValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Values["value"].(uint64)
			return ok
		}
	}

	return false
}

// HasFloatValue returns true if the given measurement has a float value
func (a *Accumulator) HasFloatValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Values["value"].(float64)
			return ok
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
