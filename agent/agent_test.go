package agent

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal/config"

	// needing to load the plugins
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	// needing to load the outputs
	_ "github.com/influxdata/telegraf/plugins/outputs/all"

	"github.com/stretchr/testify/assert"
)

func TestAgent_OmitHostname(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true
	_, err := NewAgent(c)
	assert.NoError(t, err)
	assert.NotContains(t, c.Tags, "host")
}

func TestAgent_LoadPlugin(t *testing.T) {
	c := config.NewConfig()
	c.InputFilters = []string{"mysql"}
	err := c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "redis"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))

	c = config.NewConfig()
	c.InputFilters = []string{"mysql", "foo", "redis", "bar"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Inputs))
}

func TestAgent_LoadOutput(t *testing.T) {
	c := config.NewConfig()
	c.OutputFilters = []string{"influxdb"}
	err := c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ := NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"kafka"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 1, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 0, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 2, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "kafka"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(c.Outputs))
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))

	c = config.NewConfig()
	c.OutputFilters = []string{"influxdb", "foo", "kafka", "bar"}
	err = c.LoadConfig("../internal/config/testdata/telegraf-agent.toml")
	assert.NoError(t, err)
	a, _ = NewAgent(c)
	assert.Equal(t, 3, len(a.Config.Outputs))
}

func TestAgent_ZeroJitter(t *testing.T) {
	flushinterval := jitterInterval(time.Duration(10*time.Second),
		time.Duration(0*time.Second))

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
		flushinterval := jitterInterval(time.Duration(0*time.Second),
			time.Duration(5*time.Second))
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
	flushinterval := jitterInterval(time.Duration(0*time.Second),
		time.Duration(0*time.Second))

	actual := flushinterval
	exp := time.Duration(500 * time.Millisecond)

	if actual != exp {
		t.Errorf("Actual %v, expected %v", actual, exp)
	}
}

func TestAgent_JitterMax(t *testing.T) {
	max := time.Duration(32 * time.Second).Nanoseconds()

	for i := 0; i < 1000; i++ {
		flushinterval := jitterInterval(time.Duration(30*time.Second),
			time.Duration(2*time.Second))
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
		flushinterval := jitterInterval(time.Duration(30*time.Second),
			time.Duration(2*time.Second))
		actual := flushinterval.Nanoseconds()
		if actual < min {
			t.Errorf("Didn't expect interval %d to be < %d", actual, min)
			break
		}
	}
}
