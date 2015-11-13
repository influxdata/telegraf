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

	p := &Point{
		Measurement: measurement,
		Fields:      fields,
		Tags:        tags,
		Time:        t,
	}

	a.Points = append(
		a.Points,
		p,
	)
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
	return true
}

func (a *Accumulator) SetDebug(debug bool) {
	// stub for implementing Accumulator interface.
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

// CheckValue calls CheckFieldsValue passing a single-value map as fields
func (a *Accumulator) CheckValue(measurement string, val interface{}) bool {
	return a.CheckFieldsValue(measurement, map[string]interface{}{"value": val})
}

// CheckValue checks that the accumulators point for the given measurement
// is the same as the given value.
func (a *Accumulator) CheckFieldsValue(measurement string, fields map[string]interface{}) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			if reflect.DeepEqual(fields, p.Fields) {
				return true
			} else {
				fmt.Printf("Measurement %s Failure, expected: %v, got %v\n",
					measurement, fields, p.Fields)
				return false
			}
		}
	}
	fmt.Printf("Measurement %s, fields %s not found\n", measurement, fields)
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

// ValidateTaggedValue calls ValidateTaggedFieldsValue passing a single-value map as fields
func (a *Accumulator) ValidateTaggedValue(
	measurement string,
	val interface{},
	tags map[string]string,
) error {
	return a.ValidateTaggedFieldsValue(measurement, map[string]interface{}{"value": val}, tags)
}

// ValidateValue calls ValidateTaggedValue
func (a *Accumulator) ValidateValue(measurement string, val interface{}) error {
	return a.ValidateTaggedValue(measurement, val, nil)
}

// CheckTaggedFieldsValue calls ValidateTaggedFieldsValue
func (a *Accumulator) CheckTaggedFieldsValue(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
) bool {
	return a.ValidateTaggedFieldsValue(measurement, fields, tags) == nil
}

// ValidateTaggedValue validates that the given measurement and value exist
// in the accumulator and with the given tags.
func (a *Accumulator) ValidateTaggedFieldsValue(
	measurement string,
	fields map[string]interface{},
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
			if !reflect.DeepEqual(fields, p.Fields) {
				return fmt.Errorf("%v != %v ", fields, p.Fields)
			}
			return nil
		}
	}

	return fmt.Errorf("unknown measurement %s with tags %v", measurement, tags)
}

// ValidateFieldsValue calls ValidateTaggedFieldsValue
func (a *Accumulator) ValidateFieldsValue(
	measurement string,
	fields map[string]interface{},
) error {
	return a.ValidateTaggedValue(measurement, fields, nil)
}

func (a *Accumulator) ValidateTaggedFields(
	measurement string,
	fields map[string]interface{},
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
			if !reflect.DeepEqual(fields, p.Fields) {
				return fmt.Errorf("%v (%T) != %v (%T)",
					p.Fields, p.Fields, fields, fields)
			}
			return nil
		}
	}
	return fmt.Errorf("unknown measurement %s with tags %v", measurement, tags)
}

// HasIntValue returns true if the measurement has an Int value
func (a *Accumulator) HasIntValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Fields["value"].(int64)
			return ok
		}
	}

	return false
}

// HasUIntValue returns true if the measurement has a UInt value
func (a *Accumulator) HasUIntValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Fields["value"].(uint64)
			return ok
		}
	}

	return false
}

// HasFloatValue returns true if the given measurement has a float value
func (a *Accumulator) HasFloatValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Fields["value"].(float64)
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
