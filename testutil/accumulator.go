package testutil

import "fmt"

type Point struct {
	Name  string
	Value interface{}
	Tags  map[string]string
}

type Accumulator struct {
	Points []*Point
}

func (a *Accumulator) Add(name string, value interface{}, tags map[string]string) {
	a.Points = append(a.Points, &Point{name, value, tags})
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
