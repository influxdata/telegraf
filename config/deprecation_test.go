package config

import (
	"bufio"
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

func TestPluginDeprecation(t *testing.T) {
	info := telegraf.DeprecationInfo{
		Since:     "1.23.0",
		RemovalIn: "2.0.0",
		Notice:    "please check",
	}
	var tests = []struct {
		name     string
		level    telegraf.LogLevel
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
			printPluginDeprecationNotice(tt.level, "test", info)

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
	var tests = []struct {
		name          string
		since         string
		removal       string
		expected      string
		expectedLevel telegraf.LogLevel
	}{
		{
			name:          "Error level",
			since:         "1.23.0",
			removal:       "1.29.0",
			expectedLevel: telegraf.Error,
			expected:      `Option "option" of plugin "test" deprecated since version 1.23.0 and will be removed in 1.29.0: please check`,
		},
		{
			name:          "Warn level",
			since:         "1.23.0",
			removal:       "2.0.0",
			expectedLevel: telegraf.Warn,
			expected:      `Option "option" of plugin "test" deprecated since version 1.23.0 and will be removed in 2.0.0: please check`,
		},
		{
			name:          "None",
			expectedLevel: telegraf.None,
			expected:      ``,
		},
	}

	// Fake telegraf's version
	version, err := semver.NewVersion("1.30.0")
	require.NoError(t, err)
	telegrafVersion = version

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
			info := telegraf.DeprecationInfo{
				Since:     tt.since,
				RemovalIn: tt.removal,
				Notice:    "please check",
			}
			PrintOptionDeprecationNotice("test", "option", info)
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
				expected := deprecationPrefix(tt.expectedLevel) + ": " + tt.expected
				require.Equal(t, expected, actual)
			} else {
				require.Empty(t, actual)
			}
		})
	}
}

func TestPluginOptionValueDeprecation(t *testing.T) {
	var tests = []struct {
		name          string
		since         string
		removal       string
		value         interface{}
		expected      string
		expectedLevel telegraf.LogLevel
	}{
		{
			name:          "Error level",
			since:         "1.25.0",
			removal:       "1.29.0",
			value:         "foobar",
			expected:      `Value "foobar" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 1.29.0: please check`,
			expectedLevel: telegraf.Error,
		},
		{
			name:          "Warn level",
			since:         "1.25.0",
			removal:       "2.0.0",
			value:         "foobar",
			expected:      `Value "foobar" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 2.0.0: please check`,
			expectedLevel: telegraf.Warn,
		},
		{
			name:          "None",
			expected:      ``,
			expectedLevel: telegraf.None,
		},
		{
			name:          "nil value",
			since:         "1.25.0",
			removal:       "1.29.0",
			value:         nil,
			expected:      `Value "<nil>" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 1.29.0: please check`,
			expectedLevel: telegraf.Error,
		},
		{
			name:          "Boolean value",
			since:         "1.25.0",
			removal:       "1.29.0",
			value:         true,
			expected:      `Value "true" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 1.29.0: please check`,
			expectedLevel: telegraf.Error,
		},
		{
			name:          "Integer value",
			since:         "1.25.0",
			removal:       "1.29.0",
			value:         123,
			expected:      `Value "123" for option "option" of plugin "test" deprecated since version 1.25.0 and will be removed in 1.29.0: please check`,
			expectedLevel: telegraf.Error,
		},
	}

	// Fake telegraf's version
	version, err := semver.NewVersion("1.30.0")
	require.NoError(t, err)
	telegrafVersion = version

	// Switch the logger to log to a buffer
	var buf bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(previous)

	timeout := 1 * time.Second

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			info := telegraf.DeprecationInfo{
				Since:     tt.since,
				RemovalIn: tt.removal,
				Notice:    "please check",
			}
			PrintOptionValueDeprecationNotice("test", "option", tt.value, info)

			if tt.expected != "" {
				require.Eventually(t, func() bool {
					return strings.HasSuffix(buf.String(), "\n")
				}, timeout, 100*time.Millisecond)

				// Remove the time for comparison
				parts := strings.SplitN(strings.TrimSpace(buf.String()), " ", 3)
				require.Len(t, parts, 3)
				actual := parts[2]
				expected := deprecationPrefix(tt.expectedLevel) + ": " + tt.expected
				require.Equal(t, expected, actual)
			} else {
				time.Sleep(timeout)
				require.Empty(t, buf.String())
			}
		})
	}
}
