package logparser

import (
	"bytes"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/stretchr/testify/assert"
)

func TestMultilineConfigOK(t *testing.T) {
	c := &MultilineConfig{
		Pattern: ".*",
		What:    Previous,
	}

	_, err := c.NewMultiline()

	assert.NoError(t, err, "Configuration was OK.")
}

func TestMultilineConfigError(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "\xA0",
		What:    Previous,
	}

	_, err := c.NewMultiline()

	assert.Error(t, err, "The pattern was invalid")
}

func TestMultilineConfigTimeoutSpecified(t *testing.T) {
	duration, _ := time.ParseDuration("10s")
	c := &MultilineConfig{
		Pattern: ".*",
		What:    Previous,
		Timeout: &internal.Duration{Duration: duration},
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	assert.Equal(t, duration, m.config.Timeout.Duration)
}

func TestMultilineConfigDefaultTimeout(t *testing.T) {
	duration, _ := time.ParseDuration("5s")
	c := &MultilineConfig{
		Pattern: ".*",
		What:    Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	assert.Equal(t, duration, m.config.Timeout.Duration)
}

func TestMultilineIsEnabled(t *testing.T) {
	c := &MultilineConfig{
		Pattern: ".*",
		What:    Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	isEnabled := m.IsEnabled()

	assert.True(t, isEnabled, "Should have been enabled")
}

func TestMultilineIsDisabled(t *testing.T) {
	c := &MultilineConfig{
		What: Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	isEnabled := m.IsEnabled()

	assert.False(t, isEnabled, "Should have been disabled")
}

func TestMultilineFlushEmpty(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "^=>",
		What:    Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.Flush(&buffer)

	assert.Empty(t, text)
}

func TestMultilineFlush(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "^=>",
		What:    Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer
	buffer.WriteString("foo")

	text := m.Flush(&buffer)

	assert.Equal(t, "foo", text)
	assert.Zero(t, buffer.Len())
}

func TestMultiLineProcessLinePrevious(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "^=>",
		What:    Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.ProcessLine("1", &buffer)
	assert.Empty(t, text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("=>2", &buffer)
	assert.Empty(t, text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("=>3", &buffer)
	assert.Empty(t, text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("4", &buffer)
	assert.Equal(t, "1=>2=>3", text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("5", &buffer)
	assert.Equal(t, "4", text)
	assert.Equal(t, "5", buffer.String())
}

func TestMultiLineProcessLineNext(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "=>$",
		What:    Next,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.ProcessLine("1=>", &buffer)
	assert.Empty(t, text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("2=>", &buffer)
	assert.Empty(t, text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("3=>", &buffer)
	assert.Empty(t, text)
	assert.NotZero(t, buffer.Len())

	text = m.ProcessLine("4", &buffer)
	assert.Equal(t, "1=>2=>3=>4", text)
	assert.Zero(t, buffer.Len())

	text = m.ProcessLine("5", &buffer)
	assert.Equal(t, "5", text)
	assert.Zero(t, buffer.Len())
}

func TestMultiLineMatchStringWithNegateFalse(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "=>$",
		What:    Next,
		Negate:  false,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	matches1 := m.matchString("t=>")
	matches2 := m.matchString("t")

	assert.True(t, matches1)
	assert.False(t, matches2)
}

func TestMultiLineMatchStringWithNegateTrue(t *testing.T) {
	c := &MultilineConfig{
		Pattern: "=>$",
		What:    Next,
		Negate:  true,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	matches1 := m.matchString("t=>")
	matches2 := m.matchString("t")

	assert.False(t, matches1)
	assert.True(t, matches2)
}

func TestMultilineWhat(t *testing.T) {
	var w1 MultilineWhat
	w1.UnmarshalTOML([]byte(`"previous"`))
	assert.Equal(t, Previous, w1)

	var w2 MultilineWhat
	w2.UnmarshalTOML([]byte(`previous`))
	assert.Equal(t, Previous, w2)

	var w3 MultilineWhat
	w3.UnmarshalTOML([]byte(`'previous'`))
	assert.Equal(t, Previous, w3)

	var w4 MultilineWhat
	w4.UnmarshalTOML([]byte(`"next"`))
	assert.Equal(t, Next, w4)

	var w5 MultilineWhat
	w5.UnmarshalTOML([]byte(`next`))
	assert.Equal(t, Next, w5)

	var w6 MultilineWhat
	w6.UnmarshalTOML([]byte(`'next'`))
	assert.Equal(t, Next, w6)

	var w7 MultilineWhat
	err := w7.UnmarshalTOML([]byte(`nope`))
	assert.Equal(t, MultilineWhat(-1), w7)
	assert.Error(t, err)
}
