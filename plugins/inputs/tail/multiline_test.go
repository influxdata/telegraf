package tail

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
)

func TestMultilineConfigOK(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
	}

	_, err := c.NewMultiline()

	require.NoError(t, err, "Configuration was OK.")
}

func TestMultilineConfigError(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "\xA0",
		MatchWhichLine: Previous,
	}

	_, err := c.NewMultiline()

	require.Error(t, err, "The pattern was invalid")
}

func TestMultilineConfigTimeoutSpecified(t *testing.T) {
	duration := config.Duration(10 * time.Second)
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
		Timeout:        &duration,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")

	require.Equal(t, duration, *m.config.Timeout)
}

func TestMultilineConfigDefaultTimeout(t *testing.T) {
	duration := config.Duration(5 * time.Second)
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")

	require.Equal(t, duration, *m.config.Timeout)
}

func TestMultilineIsEnabled(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        ".*",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")

	isEnabled := m.IsEnabled()

	require.True(t, isEnabled, "Should have been enabled")
}

func TestMultilineIsDisabled(t *testing.T) {
	c := &MultilineConfig{
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")

	isEnabled := m.IsEnabled()

	require.False(t, isEnabled, "Should have been disabled")
}

func TestMultilineFlushEmpty(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "^=>",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.Flush(&buffer)

	require.Empty(t, text)
}

func TestMultilineFlush(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "^=>",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer
	_, err = buffer.WriteString("foo")
	require.NoError(t, err)

	text := m.Flush(&buffer)

	require.Equal(t, "foo", text)
	require.Zero(t, buffer.Len())
}

func TestMultiLineProcessLinePrevious(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "^=>",
		MatchWhichLine: Previous,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.ProcessLine("1", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("=>2", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("=>3", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("4", &buffer)
	require.Equal(t, "1=>2=>3", text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("5", &buffer)
	require.Equal(t, "4", text)
	require.Equal(t, "5", buffer.String())
}

func TestMultiLineProcessLineNext(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "=>$",
		MatchWhichLine: Next,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.ProcessLine("1=>", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("2=>", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("3=>", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("4", &buffer)
	require.Equal(t, "1=>2=>3=>4", text)
	require.Zero(t, buffer.Len())

	text = m.ProcessLine("5", &buffer)
	require.Equal(t, "5", text)
	require.Zero(t, buffer.Len())
}

func TestMultiLineMatchStringWithInvertMatchFalse(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "=>$",
		MatchWhichLine: Next,
		InvertMatch:    false,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")

	matches1 := m.matchString("t=>")
	matches2 := m.matchString("t")

	require.True(t, matches1)
	require.False(t, matches2)
}

func TestMultiLineMatchStringWithInvertTrue(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "=>$",
		MatchWhichLine: Next,
		InvertMatch:    true,
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")

	matches1 := m.matchString("t=>")
	matches2 := m.matchString("t")

	require.False(t, matches1)
	require.True(t, matches2)
}

func TestMultilineWhat(t *testing.T) {
	var w1 MultilineMatchWhichLine
	require.NoError(t, w1.UnmarshalTOML([]byte(`"previous"`)))
	require.Equal(t, Previous, w1)

	var w2 MultilineMatchWhichLine
	require.NoError(t, w2.UnmarshalTOML([]byte(`previous`)))
	require.Equal(t, Previous, w2)

	var w3 MultilineMatchWhichLine
	require.NoError(t, w3.UnmarshalTOML([]byte(`'previous'`)))
	require.Equal(t, Previous, w3)

	var w4 MultilineMatchWhichLine
	require.NoError(t, w4.UnmarshalTOML([]byte(`"next"`)))
	require.Equal(t, Next, w4)

	var w5 MultilineMatchWhichLine
	require.NoError(t, w5.UnmarshalTOML([]byte(`next`)))
	require.Equal(t, Next, w5)

	var w6 MultilineMatchWhichLine
	require.NoError(t, w6.UnmarshalTOML([]byte(`'next'`)))
	require.Equal(t, Next, w6)

	var w7 MultilineMatchWhichLine
	require.Error(t, w7.UnmarshalTOML([]byte(`nope`)))
	require.Equal(t, MultilineMatchWhichLine(-1), w7)
}

func TestMultiLineQuoted(t *testing.T) {
	tests := []struct {
		name      string
		quotation string
		quote     string
		filename  string
	}{
		{
			name:      "single-quotes",
			quotation: "single-quotes",
			quote:     `'`,
			filename:  "multiline_quoted_single.csv",
		},
		{
			name:      "double-quotes",
			quotation: "double-quotes",
			quote:     `"`,
			filename:  "multiline_quoted_double.csv",
		},
		{
			name:      "backticks",
			quotation: "backticks",
			quote:     "`",
			filename:  "multiline_quoted_backticks.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := []string{
				`1660819827410,1,some text without quotes,A`,
				fmt.Sprintf("1660819827411,1,%ssome text all quoted%s,A", tt.quote, tt.quote),
				fmt.Sprintf("1660819827412,1,%ssome text all quoted\nbut wrapped%s,A", tt.quote, tt.quote),
				fmt.Sprintf("1660819827420,2,some text with %squotes%s,B", tt.quote, tt.quote),
				"1660819827430,3,some text with 'multiple \"quotes\" in `one` line',C",
				fmt.Sprintf("1660819827440,4,some multiline text with %squotes\nspanning \\%smultiple\\%s\nlines%s but do not %send\ndirectly%s,D", tt.quote, tt.quote, tt.quote, tt.quote, tt.quote, tt.quote),
				fmt.Sprintf("1660819827450,5,all of %sthis%s should %sbasically%s work...,E", tt.quote, tt.quote, tt.quote, tt.quote),
			}

			c := &MultilineConfig{
				MatchWhichLine: Next,
				Quotation:      tt.quotation,
			}
			m, err := c.NewMultiline()
			require.NoError(t, err)

			f, err := os.Open(filepath.Join("testdata", tt.filename))
			require.NoError(t, err)

			scanner := bufio.NewScanner(f)

			var buffer bytes.Buffer
			var result []string
			for scanner.Scan() {
				line := scanner.Text()

				text := m.ProcessLine(line, &buffer)
				if text == "" {
					continue
				}
				result = append(result, text)
			}

			require.EqualValues(t, expected, result)
		})
	}
}

func TestMultiLineQuotedAndPattern(t *testing.T) {
	c := &MultilineConfig{
		Pattern:        "=>$",
		MatchWhichLine: Next,
		Quotation:      "double-quotes",
	}
	m, err := c.NewMultiline()
	require.NoError(t, err, "Configuration was OK.")
	var buffer bytes.Buffer

	text := m.ProcessLine("1=>", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("2=>", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine(`"a quoted`, &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine(`multiline string"=>`, &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("3=>", &buffer)
	require.Empty(t, text)
	require.NotZero(t, buffer.Len())

	text = m.ProcessLine("4", &buffer)
	require.Equal(t, "1=>2=>\"a quoted\nmultiline string\"=>3=>4", text)
	require.Zero(t, buffer.Len())

	text = m.ProcessLine("5", &buffer)
	require.Equal(t, "5", text)
	require.Zero(t, buffer.Len())
}
