package api // import "collectd.org/api"

import (
	"testing"
)

func TestParseIdentifier(t *testing.T) {
	cases := []struct {
		Input string
		Want  Identifier
	}{
		{
			Input: "example.com/golang/gauge",
			Want: Identifier{
				Host:   "example.com",
				Plugin: "golang",
				Type:   "gauge",
			},
		},
		{
			Input: "example.com/golang-foo/gauge-bar",
			Want: Identifier{
				Host:           "example.com",
				Plugin:         "golang",
				PluginInstance: "foo",
				Type:           "gauge",
				TypeInstance:   "bar",
			},
		},
		{
			Input: "example.com/golang-a-b/gauge-b-c",
			Want: Identifier{
				Host:           "example.com",
				Plugin:         "golang",
				PluginInstance: "a-b",
				Type:           "gauge",
				TypeInstance:   "b-c",
			},
		},
	}

	for i, c := range cases {
		if got, err := ParseIdentifier(c.Input); got != c.Want || err != nil {
			t.Errorf("case %d: got (%v, %v), want (%v, %v)", i, got, err, c.Want, nil)
		}
	}

	failures := []string{
		"example.com/golang",
		"example.com/golang/gauge/extra",
	}

	for _, c := range failures {
		if got, err := ParseIdentifier(c); err == nil {
			t.Errorf("got (%v, %v), want (%v, !%v)", got, err, Identifier{}, nil)
		}
	}
}

func TestIdentifierString(t *testing.T) {
	id := Identifier{
		Host:   "example.com",
		Plugin: "golang",
		Type:   "gauge",
	}

	cases := []struct {
		PluginInstance, TypeInstance string
		Want                         string
	}{
		{"", "", "example.com/golang/gauge"},
		{"foo", "", "example.com/golang-foo/gauge"},
		{"", "foo", "example.com/golang/gauge-foo"},
		{"foo", "bar", "example.com/golang-foo/gauge-bar"},
	}

	for _, c := range cases {
		id.PluginInstance = c.PluginInstance
		id.TypeInstance = c.TypeInstance

		got := id.String()
		if got != c.Want {
			t.Errorf("got %q, want %q", got, c.Want)
		}
	}
}
