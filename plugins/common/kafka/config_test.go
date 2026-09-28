package kafka

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestBackoffFunc(t *testing.T) {
	b := 250 * time.Millisecond
	limit := 1100 * time.Millisecond

	f := makeBackoffFunc(b, limit)
	require.Equal(t, b, f(0, 0))
	require.Equal(t, b*2, f(1, 0))
	require.Equal(t, b*4, f(2, 0))
	require.Equal(t, limit, f(3, 0)) // would be 2000 but that's greater than max

	f = makeBackoffFunc(b, 0)      // max = 0 means no max
	require.Equal(t, b*8, f(3, 0)) // with no max, it's 2000
}

func TestWriteConfigTimeoutsDefault(t *testing.T) {
	cfg := sarama.NewConfig()
	defaults := sarama.NewConfig()

	k := WriteConfig{}
	require.NoError(t, k.SetConfig(cfg, testutil.Logger{}))

	// Leaving the options unset must not change what sarama defaults to
	require.Equal(t, defaults.Net.DialTimeout, cfg.Net.DialTimeout)
	require.Equal(t, defaults.Net.ReadTimeout, cfg.Net.ReadTimeout)
	require.Equal(t, defaults.Net.WriteTimeout, cfg.Net.WriteTimeout)
	require.Equal(t, defaults.Producer.Timeout, cfg.Producer.Timeout)
}

func TestWriteConfigTimeouts(t *testing.T) {
	cfg := sarama.NewConfig()

	k := WriteConfig{
		NetDialTimeout:  config.Duration(time.Second),
		NetReadTimeout:  config.Duration(2 * time.Second),
		NetWriteTimeout: config.Duration(3 * time.Second),
		ProducerTimeout: config.Duration(4 * time.Second),
	}
	require.NoError(t, k.SetConfig(cfg, testutil.Logger{}))

	require.Equal(t, time.Second, cfg.Net.DialTimeout)
	require.Equal(t, 2*time.Second, cfg.Net.ReadTimeout)
	require.Equal(t, 3*time.Second, cfg.Net.WriteTimeout)
	require.Equal(t, 4*time.Second, cfg.Producer.Timeout)
	require.NoError(t, cfg.Validate())
}

func TestWriteConfigTimeoutsNegative(t *testing.T) {
	tests := []struct {
		name     string
		config   WriteConfig
		expected string
	}{
		{
			name:     "dial",
			config:   WriteConfig{NetDialTimeout: config.Duration(-time.Second)},
			expected: "Net.DialTimeout must be > 0",
		},
		{
			name:     "read",
			config:   WriteConfig{NetReadTimeout: config.Duration(-time.Second)},
			expected: "Net.ReadTimeout must be > 0",
		},
		{
			name:     "write",
			config:   WriteConfig{NetWriteTimeout: config.Duration(-time.Second)},
			expected: "Net.WriteTimeout must be > 0",
		},
		{
			name:     "producer",
			config:   WriteConfig{ProducerTimeout: config.Duration(-time.Second)},
			expected: "Producer.Timeout must be > 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Negative values must reach sarama's validation instead of
			// silently falling back to the default
			cfg := sarama.NewConfig()
			require.NoError(t, tt.config.SetConfig(cfg, testutil.Logger{}))
			require.ErrorContains(t, cfg.Validate(), tt.expected)
		})
	}
}
