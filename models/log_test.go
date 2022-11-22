package models

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/stretchr/testify/require"
)

func TestErrorCounting(t *testing.T) {
	reg := selfstat.Register(
		"gather",
		"errors",
		map[string]string{"input": "test"},
	)
	iLog := Logger{Name: "inputs.test"}
	iLog.OnErr(func() {
		reg.Incr(1)
	})
	iLog.Error("something went wrong")
	iLog.Errorf("something went wrong")

	require.Equal(t, int64(2), reg.Get())
}

func TestPluginDeprecation(t *testing.T) {
	info := telegraf.DeprecationInfo{
		Since:     "1.23.0",
		RemovalIn: "2.0.0",
		Notice:    "please check",
	}
	var tests = []struct {
		name     string
		level    telegraf.Escalation
		expected string
	}{
		{
			name:     "Error level",
			level:    telegraf.Error,
			expected: `Plugin "test" deprecated since version 1.23.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "Warn level",
			level:    telegraf.Warn,
			expected: `Plugin "test" deprecated since version 1.23.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "None",
			level:    telegraf.None,
			expected: ``,
		},
	}

	// Switch the logger to log to a buffer
	var buf bytes.Buffer
	scanner := bufio.NewScanner(&buf)

	previous := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(previous)

	msg := make(chan string, 1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			PrintPluginDeprecationNotice(tt.level, "test", info)

			// Wait for a newline to arrive and timeout for cases where
			// we don't see a message.
			go func() {
				scanner.Scan()
				msg <- scanner.Text()
			}()

			// Reduce the timeout if we do not expect a message
			timeout := 1 * time.Second
			if tt.expected == "" {
				timeout = 100 * time.Microsecond
			}

			var actual string
			select {
			case actual = <-msg:
			case <-time.After(timeout):
			}

			if tt.expected != "" {
				// Remove the time for comparison
				parts := strings.SplitN(actual, " ", 3)
				require.Len(t, parts, 3)
				actual = parts[2]
				expected := deprecationPrefix(tt.level) + ": " + tt.expected
				require.Equal(t, expected, actual)
			} else {
				require.Empty(t, actual)
			}
		})
	}
}

func TestPluginOptionDeprecation(t *testing.T) {
	info := telegraf.DeprecationInfo{
		Since:     "1.23.0",
		RemovalIn: "2.0.0",
		Notice:    "please check",
	}
	var tests = []struct {
		name     string
		level    telegraf.Escalation
		expected string
	}{
		{
			name:     "Error level",
			level:    telegraf.Error,
			expected: `Option "option" of plugin "test" deprecated since version 1.23.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "Warn level",
			level:    telegraf.Warn,
			expected: `Option "option" of plugin "test" deprecated since version 1.23.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "None",
			level:    telegraf.None,
			expected: ``,
		},
	}

	// Switch the logger to log to a buffer
	var buf bytes.Buffer
	scanner := bufio.NewScanner(&buf)

	previous := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(previous)

	msg := make(chan string, 1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			PrintOptionDeprecationNotice(tt.level, "test", "option", info)
			// Wait for a newline to arrive and timeout for cases where
			// we don't see a message.
			go func() {
				scanner.Scan()
				msg <- scanner.Text()
			}()

			// Reduce the timeout if we do not expect a message
			timeout := 1 * time.Second
			if tt.expected == "" {
				timeout = 100 * time.Microsecond
			}

			var actual string
			select {
			case actual = <-msg:
			case <-time.After(timeout):
			}

			if tt.expected != "" {
				// Remove the time for comparison
				parts := strings.SplitN(actual, " ", 3)
				require.Len(t, parts, 3)
				actual = parts[2]
				expected := deprecationPrefix(tt.level) + ": " + tt.expected
				require.Equal(t, expected, actual)
			} else {
				require.Empty(t, actual)
			}
		})
	}
}

func TestPluginOptionValueDeprecation(t *testing.T) {
	info := telegraf.DeprecationInfo{
		Since:     "1.25.0",
		RemovalIn: "2.0.0",
		Notice:    "please check",
	}
	var tests = []struct {
		name     string
		level    telegraf.Escalation
		value    interface{}
		expected string
	}{
		{
			name:     "Error level",
			level:    telegraf.Error,
			value:    "foobar",
			expected: `Value "foobar" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "Warn level",
			level:    telegraf.Warn,
			value:    "foobar",
			expected: `Value "foobar" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "None",
			level:    telegraf.None,
			expected: ``,
		},
		{
			name:     "nil value",
			level:    telegraf.Error,
			value:    nil,
			expected: `Value "<nil>" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "Boolean value",
			level:    telegraf.Error,
			value:    true,
			expected: `Value "true" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:     "Integer value",
			level:    telegraf.Error,
			value:    123,
			expected: `Value "123" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 2.0.0: please check`,
		},
	}

	// Switch the logger to log to a buffer
	var buf bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(previous)

	timeout := 1 * time.Second

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			PrintOptionValueDeprecationNotice(tt.level, "test", "option", tt.value, info)

			if tt.expected != "" {
				require.Eventually(t, func() bool {
					return strings.HasSuffix(buf.String(), "\n")
				}, timeout, 100*time.Millisecond)

				// Remove the time for comparison
				parts := strings.SplitN(strings.TrimSpace(buf.String()), " ", 3)
				require.Len(t, parts, 3)
				actual := parts[2]
				expected := deprecationPrefix(tt.level) + ": " + tt.expected
				require.Equal(t, expected, actual)
			} else {
				time.Sleep(timeout)
				require.Empty(t, buf.String())
			}
		})
	}
}
