package testutil

import (
	"fmt"
	"time"
)

type Point struct {
	Measurement string
	Value       interface{}
	Tags        map[string]string
	Values      map[string]interface{}
	Time        time.Time
}

type Accumulator struct {
	Points []*Point
}

func (a *Accumulator) Add(measurement string, value interface{}, tags map[string]string) {
	a.Points = append(
		a.Points,
		&Point{
			Measurement: measurement,
			Value:       value,
			Tags:        tags,
		},
	)
}

func (a *Accumulator) AddValuesWithTime(
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

func (a *Accumulator) Get(measurement string) (*Point, bool) {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			return p, true
		}
	}

	return nil, false
}

func (a *Accumulator) CheckValue(measurement string, val interface{}) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			return p.Value == val
		}
	}

	return false
}

func (a *Accumulator) CheckTaggedValue(measurement string, val interface{}, tags map[string]string) bool {
	return a.ValidateTaggedValue(measurement, val, tags) == nil
}

func (a *Accumulator) ValidateTaggedValue(measurement string, val interface{}, tags map[string]string) error {
	for _, p := range a.Points {
		var found bool

		if p.Tags == nil && tags == nil {
			found = true
		} else {
			for k, v := range p.Tags {
				if tags[k] == v {
					found = true
					break
				}
			}
		}

		if found && p.Measurement == measurement {
			if p.Value != val {
				return fmt.Errorf("%v (%T) != %v (%T)", p.Value, p.Value, val, val)
			}

			return nil
		}
	}

	return fmt.Errorf("unknown value %s with tags %v", measurement, tags)
}

func (a *Accumulator) ValidateValue(measurement string, val interface{}) error {
	return a.ValidateTaggedValue(measurement, val, nil)
}

func (a *Accumulator) HasIntValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Value.(int64)
			return ok
		}
	}

	return false
}

func (a *Accumulator) HasFloatValue(measurement string) bool {
	for _, p := range a.Points {
		if p.Measurement == measurement {
			_, ok := p.Value.(float64)
			return ok
		}
	}

	return false
}
