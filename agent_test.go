package telegraf

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/influxdb/telegraf/duration"

	// needing to load the plugins
	_ "github.com/influxdb/telegraf/plugins/all"
	// needing to load the outputs
	_ "github.com/influxdb/telegraf/outputs/all"
)

func TestAgent_LoadPlugin(t *testing.T) {

	// load a dedicated configuration file
	config, _ := LoadConfig("./testdata/telegraf-agent.toml")
	a, _ := NewAgent(config)

	pluginsEnabled, _ := a.LoadPlugins([]string{"mysql"}, config)
	assert.Equal(t, 1, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"foo"}, config)
	assert.Equal(t, 0, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"mysql", "foo"}, config)
	assert.Equal(t, 1, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"mysql", "redis"}, config)
	assert.Equal(t, 2, len(pluginsEnabled))

	pluginsEnabled, _ = a.LoadPlugins([]string{"mysql", "foo", "redis", "bar"}, config)
	assert.Equal(t, 2, len(pluginsEnabled))
}

func TestAgent_LoadOutput(t *testing.T) {
	// load a dedicated configuration file
	config, _ := LoadConfig("./testdata/telegraf-agent.toml")
	a, _ := NewAgent(config)

	outputsEnabled, _ := a.LoadOutputs([]string{"influxdb"}, config)
	assert.Equal(t, 1, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{}, config)
	assert.Equal(t, 2, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"foo"}, config)
	assert.Equal(t, 0, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"influxdb", "foo"}, config)
	assert.Equal(t, 1, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"influxdb", "kafka"}, config)
	assert.Equal(t, 2, len(outputsEnabled))

	outputsEnabled, _ = a.LoadOutputs([]string{"influxdb", "foo", "kafka", "bar"}, config)
	assert.Equal(t, 2, len(outputsEnabled))
}

func TestAgent_ZeroJitter(t *testing.T) {
	a := &Agent{
		FlushInterval: duration.Duration{10 * time.Second},
		FlushJitter:   duration.Duration{0 * time.Second},
	}
	flushinterval := jitterInterval(a.FlushInterval.Duration,
		a.FlushJitter.Duration)

	actual := flushinterval.Nanoseconds()
	exp := time.Duration(10 * time.Second).Nanoseconds()

	if actual != exp {
		t.Errorf("Actual %v, expected %v", actual, exp)
	}
}

func TestAgent_ZeroInterval(t *testing.T) {
	min := time.Duration(500 * time.Millisecond).Nanoseconds()
	max := time.Duration(5 * time.Second).Nanoseconds()

	for i := 0; i < 1000; i++ {
		a := &Agent{
			FlushInterval: duration.Duration{0 * time.Second},
			FlushJitter:   duration.Duration{5 * time.Second},
		}

		flushinterval := jitterInterval(a.FlushInterval.Duration,
			a.FlushJitter.Duration)
		actual := flushinterval.Nanoseconds()

		if actual > max {
			t.Errorf("Didn't expect interval %d to be > %d", actual, max)
			break
		}
		if actual < min {
			t.Errorf("Didn't expect interval %d to be < %d", actual, min)
			break
		}
	}
}

func TestAgent_ZeroBoth(t *testing.T) {
	a := &Agent{
		FlushInterval: duration.Duration{0 * time.Second},
		FlushJitter:   duration.Duration{0 * time.Second},
	}

	flushinterval := jitterInterval(a.FlushInterval.Duration,
		a.FlushJitter.Duration)

	actual := flushinterval
	exp := time.Duration(500 * time.Millisecond)

	if actual != exp {
		t.Errorf("Actual %v, expected %v", actual, exp)
	}
}

func TestAgent_JitterMax(t *testing.T) {
	max := time.Duration(32 * time.Second).Nanoseconds()

	for i := 0; i < 1000; i++ {
		a := &Agent{
			FlushInterval: duration.Duration{30 * time.Second},
			FlushJitter:   duration.Duration{2 * time.Second},
		}
		flushinterval := jitterInterval(a.FlushInterval.Duration,
			a.FlushJitter.Duration)
		actual := flushinterval.Nanoseconds()
		if actual > max {
			t.Errorf("Didn't expect interval %d to be > %d", actual, max)
			break
		}
	}
}

func TestAgent_JitterMin(t *testing.T) {
	min := time.Duration(30 * time.Second).Nanoseconds()

	for i := 0; i < 1000; i++ {
		a := &Agent{
			FlushInterval: duration.Duration{30 * time.Second},
			FlushJitter:   duration.Duration{2 * time.Second},
		}
		flushinterval := jitterInterval(a.FlushInterval.Duration,
			a.FlushJitter.Duration)
		actual := flushinterval.Nanoseconds()
		if actual < min {
			t.Errorf("Didn't expect interval %d to be < %d", actual, min)
			break
		}
	}
}
