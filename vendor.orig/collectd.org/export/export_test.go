package export // import "collectd.org/export"

import (
	"expvar"
	"reflect"
	"testing"

	"collectd.org/api"
)

func TestDerive(t *testing.T) {
	d := NewDerive(api.Identifier{
		Host:   "example.com",
		Plugin: "golang",
		Type:   "derive",
	})

	for i := 0; i < 10; i++ {
		d.Add(i)
	}

	want := &api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "derive",
		},
		Values: []api.Value{api.Derive(45)},
	}
	got := d.ValueList()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}

	s := expvar.Get("example.com/golang/derive").String()
	if s != "45" {
		t.Errorf("got %q, want %q", s, "45")
	}
}

func TestGauge(t *testing.T) {
	g := NewGauge(api.Identifier{
		Host:   "example.com",
		Plugin: "golang",
		Type:   "gauge",
	})

	g.Set(42.0)

	want := &api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Values: []api.Value{api.Gauge(42)},
	}
	got := g.ValueList()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}

	s := expvar.Get("example.com/golang/gauge").String()
	if s != "42" {
		t.Errorf("got %q, want %q", s, "42")
	}
}
