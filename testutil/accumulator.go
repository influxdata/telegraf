package testutil

import (
	"fmt"
	"time"
)

type Point struct {
	Name   string
	Value  interface{}
	Tags   map[string]string
	Values map[string]interface{}
	Time   time.Time
}

type Accumulator struct {
	Points []*Point
}

func (a *Accumulator) Add(name string, value interface{}, tags map[string]string) {
	a.Points = append(
		a.Points,
		&Point{
			Name:  name,
			Value: value,
			Tags:  tags,
		},
	)
}

func (a *Accumulator) AddValuesWithTime(
	name string,
	values map[string]interface{},
	tags map[string]string,
	timestamp time.Time,
) {
	a.Points = append(
		a.Points,
		&Point{
			Name:   name,
			Values: values,
			Tags:   tags,
			Time:   timestamp,
		},
	)
}

func (a *Accumulator) Get(name string) (*Point, bool) {
	for _, p := range a.Points {
		if p.Name == name {
			return p, true
		}
	}

	return nil, false
}

func (a *Accumulator) CheckValue(name string, val interface{}) bool {
	for _, p := range a.Points {
		if p.Name == name {
			return p.Value == val
		}
	}

	return false
}

func (a *Accumulator) CheckTaggedValue(name string, val interface{}, tags map[string]string) bool {
	return a.ValidateTaggedValue(name, val, tags) == nil
}

func (a *Accumulator) ValidateTaggedValue(name string, val interface{}, tags map[string]string) error {
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

		if found && p.Name == name {
			if p.Value != val {
				return fmt.Errorf("%v (%T) != %v (%T)", p.Value, p.Value, val, val)
			}

			return nil
		}
	}

	return fmt.Errorf("unknown value %s with tags %v", name, tags)
}

func (a *Accumulator) ValidateValue(name string, val interface{}) error {
	return a.ValidateTaggedValue(name, val, nil)
}

func (a *Accumulator) HasIntValue(name string) bool {
	for _, p := range a.Points {
		if p.Name == name {
			_, ok := p.Value.(int64)
			return ok
		}
	}

	return false
}

func (a *Accumulator) HasFloatValue(name string) bool {
	for _, p := range a.Points {
		if p.Name == name {
			_, ok := p.Value.(float64)
			return ok
		}
	}

	return false
}
