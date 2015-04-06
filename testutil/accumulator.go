package testutil

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
