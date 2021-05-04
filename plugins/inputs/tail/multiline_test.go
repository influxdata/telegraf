package tail

import (
	"bytes"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/stretchr/testify/assert"
)

func TestMultilineConfigOK(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
	}

	_, err := c.NewMultiline()

	assert.NoError(t, err, "Configuration was OK.")
}

func TestMultilineConfigError(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "\xA0",
		MatchWhichLine: Previous,
	}

	_, err := c.NewMultiline()

	assert.Error(t, err, "The pattern was invalid")
}

func TestMultilineConfigTimeoutSpecified(t *testing.T) {
	duration := config.Duration(10 * time.Second)
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
		Timeout:        &duration,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	assert.Equal(t, duration, *m.config.Timeout)
}

func TestMultilineConfigDefaultTimeout(t *testing.T) {
	duration := config.Duration(5 * time.Second)
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	assert.Equal(t, duration, *m.config.Timeout)
}

func TestMultilineIsEnabled(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	isEnabled := m.IsEnabled()

	assert.True(t, isEnabled, "Should have been enabled")
}

func TestMultilineIsDisabled(t *testing.T) {
	c := &MultilineConfig{
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	isEnabled := m.IsEnabled()

	assert.False(t, isEnabled, "Should have been disabled")
}

func TestMultilineFlushEmpty(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "^=>",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.Flush(&buffer)

	assert.Empty(t, text)
}

func TestMultilineFlush(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "^=>",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer
	_, err = buffer.WriteString("foo")
	assert.NoError(t, err)

	text := m.Flush(&buffer)

	assert.Equal(t, "foo", text)
	assert.Zero(t, buffer.Len())
}

func TestMultiLineProcessLinePrevious(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "^=>",
		MatchWhichLine: Previous,
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
		Pattern:        "=>$",
		MatchWhichLine: Next,
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

func TestMultiLineMatchStringWithInvertMatchFalse(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "=>$",
		MatchWhichLine: Next,
		InvertMatch:    false,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	matches1 := m.matchString("t=>")
	matches2 := m.matchString("t")

	assert.True(t, matches1)
	assert.False(t, matches2)
}

func TestMultiLineMatchStringWithInvertTrue(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "=>$",
		MatchWhichLine: Next,
		InvertMatch:    true,
	}
	m, err := c.NewMultiline()
	assert.NoError(t, err, "Configuration was OK.")

	matches1 := m.matchString("t=>")
	matches2 := m.matchString("t")

	assert.False(t, matches1)
	assert.True(t, matches2)
}

func TestMultilineWhat(t *testing.T) {
	var w1 MultilineMatchWhichLine
	assert.NoError(t, w1.UnmarshalTOML([]byte(`"previous"`)))
	assert.Equal(t, Previous, w1)

	var w2 MultilineMatchWhichLine
	assert.NoError(t, w2.UnmarshalTOML([]byte(`previous`)))
	assert.Equal(t, Previous, w2)

	var w3 MultilineMatchWhichLine
	assert.NoError(t, w3.UnmarshalTOML([]byte(`'previous'`)))
	assert.Equal(t, Previous, w3)

	var w4 MultilineMatchWhichLine
	assert.NoError(t, w4.UnmarshalTOML([]byte(`"next"`)))
	assert.Equal(t, Next, w4)

	var w5 MultilineMatchWhichLine
	assert.NoError(t, w5.UnmarshalTOML([]byte(`next`)))
	assert.Equal(t, Next, w5)

	var w6 MultilineMatchWhichLine
	assert.NoError(t, w6.UnmarshalTOML([]byte(`'next'`)))
	assert.Equal(t, Next, w6)

	var w7 MultilineMatchWhichLine
	assert.Error(t, w7.UnmarshalTOML([]byte(`nope`)))
	assert.Equal(t, MultilineMatchWhichLine(-1), w7)
}
