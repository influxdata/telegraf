package format

import (
	"bytes"
	"context"
	"testing"
	"time"

	"collectd.org/api"
)

func TestWrite(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		ValueList *api.ValueList
		Graphite  *Graphite
		Want      string
	}{
		{ // case 0
			ValueList: &api.ValueList{
				Identifier: api.Identifier{
					Host:           "example.com",
					Plugin:         "golang",
					PluginInstance: "example",
					Type:           "gauge",
					TypeInstance:   "answer",
				},
				Time:     time.Unix(1426975989, 1),
				Interval: 10 * time.Second,
				Values:   []api.Value{api.Gauge(42)},
			},
			Graphite: &Graphite{
				Prefix:            "-->",
				Suffix:            "<--",
				EscapeChar:        "_",
				SeparateInstances: false,
				AlwaysAppendDS:    true,
			},
			Want: "-->example_com<--.golang-example.gauge-answer.value 42 1426975989\r\n",
		},
		{ // case 1
			ValueList: &api.ValueList{
				Identifier: api.Identifier{
					Host:           "example.com",
					Plugin:         "golang",
					PluginInstance: "example",
					Type:           "gauge",
					TypeInstance:   "answer",
				},
				Time:     time.Unix(1426975989, 1),
				Interval: 10 * time.Second,
				Values:   []api.Value{api.Derive(1337)},
			},
			Graphite: &Graphite{
				Prefix:            "collectd.",
				Suffix:            "",
				EscapeChar:        "@",
				SeparateInstances: true,
				AlwaysAppendDS:    false,
			},
			Want: "collectd.example@com.golang.example.gauge.answer 1337 1426975989\r\n",
		},
		{ // case 2
			ValueList: &api.ValueList{
				Identifier: api.Identifier{
					Host:   "example.com",
					Plugin: "golang",
					Type:   "gauge",
				},
				Time:     time.Unix(1426975989, 1),
				Interval: 10 * time.Second,
				Values:   []api.Value{api.Gauge(42), api.Derive(1337)},
			},
			Graphite: &Graphite{
				Prefix:            "collectd.",
				Suffix:            "",
				EscapeChar:        "_",
				SeparateInstances: true,
				AlwaysAppendDS:    false,
			},
			Want: "collectd.example_com.golang.gauge.0 42 1426975989\r\n" +
				"collectd.example_com.golang.gauge.1 1337 1426975989\r\n",
		},
	}

	for i, c := range cases {
		buf := &bytes.Buffer{}
		c.Graphite.W = buf

		if err := c.Graphite.Write(ctx, c.ValueList); err != nil {
			t.Errorf("case %d: got %v, want %v", i, err, nil)
		}

		got := buf.String()
		if got != c.Want {
			t.Errorf("got %q, want %q", got, c.Want)
		}
	}
}
