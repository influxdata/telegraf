package exec // import "collectd.org/exec"

import (
	"context"
	"os"
	"testing"
	"time"

	"collectd.org/api"
)

func TestSanitizeInterval(t *testing.T) {
	var got, want time.Duration

	got = sanitizeInterval(10 * time.Second)
	want = 10 * time.Second
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	// Environment with seconds
	if err := os.Setenv("COLLECTD_INTERVAL", "42"); err != nil {
		t.Fatalf("os.Setenv: %v", err)
	}
	got = sanitizeInterval(0)
	want = 42 * time.Second
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	// Environment with milliseconds
	if err := os.Setenv("COLLECTD_INTERVAL", "31.337"); err != nil {
		t.Fatalf("os.Setenv: %v", err)
	}
	got = sanitizeInterval(0)
	want = 31337 * time.Millisecond
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func Example() {
	e := NewExecutor()

	// simple "value" callback
	answer := func() api.Value {
		return api.Gauge(42)
	}
	e.ValueCallback(answer, &api.ValueList{
		Identifier: api.Identifier{
			Host:         "example.com",
			Plugin:       "golang",
			Type:         "answer",
			TypeInstance: "live_universe_and_everything",
		},
		Interval: time.Second,
	})

	// "complex" void callback
	bicycles := func(ctx context.Context, interval time.Duration) {
		vl := &api.ValueList{
			Identifier: api.Identifier{
				Host:   "example.com",
				Plugin: "golang",
				Type:   "bicycles",
			},
			Interval: interval,
			Time:     time.Now(),
			Values:   make([]api.Value, 1),
		}

		data := []struct {
			TypeInstance string
			Value        api.Gauge
		}{
			{"beijing", api.Gauge(9000000)},
		}
		for _, d := range data {
			vl.Values[0] = d.Value
			vl.Identifier.TypeInstance = d.TypeInstance
			Putval.Write(ctx, vl)
		}
	}
	e.VoidCallback(bicycles, time.Second)

	// blocks forever
	e.Run(context.Background())
}
