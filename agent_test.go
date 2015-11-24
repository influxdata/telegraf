package telegraf

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/influxdb/telegraf/internal"
	"github.com/influxdb/telegraf/internal/config"

	// needing to load the plugins
	_ "github.com/influxdb/telegraf/plugins/all"
	// needing to load the outputs
	_ "github.com/influxdb/telegraf/outputs/all"
)

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.PluginFilters = []string{"mysql"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ := NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Plugins))

	c = config.NewConfig()
	c.PluginFilters = []string{"foo"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Plugins))

	c = config.NewConfig()
	c.PluginFilters = []string{"mysql", "foo"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Plugins))

	c = config.NewConfig()
	c.PluginFilters = []string{"mysql", "redis"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Plugins))

	c = config.NewConfig()
	c.PluginFilters = []string{"mysql", "foo", "redis", "bar"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Plugins))
}

func TestAgent_LoadOutput(t *testing.T) {
	c := config.NewConfig()
	c.OutputFilters = []string{"influxdb"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ := NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"foo"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "kafka"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo", "kafka", "bar"}
	c.LoadConfig("./internal/config/testdata/telegraf-agent.toml")
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))
}

func TestAgent_ZeroJitter(t *testing.T) {
	a := &Agent{
		FlushInterval: internal.Duration{10 * time.Second},
		FlushJitter:   internal.Duration{0 * time.Second},
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
			FlushInterval: internal.Duration{0 * time.Second},
			FlushJitter:   internal.Duration{5 * time.Second},
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
		FlushInterval: internal.Duration{0 * time.Second},
		FlushJitter:   internal.Duration{0 * time.Second},
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
			FlushInterval: internal.Duration{30 * time.Second},
			FlushJitter:   internal.Duration{2 * time.Second},
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
			FlushInterval: internal.Duration{30 * time.Second},
			FlushJitter:   internal.Duration{2 * time.Second},
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
