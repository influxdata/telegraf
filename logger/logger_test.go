package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/selfstat"
)

func TestTextLogTargetDefault(t *testing.T) {
	instance = defaultHandler()
	cfg := &Config{
		Quiet: true,
	}
	require.NoError(t, SetupLogging(cfg))
	defer func() { require.NoError(t, CloseLogging()) }()

	logger, ok := instance.impl.(*textLogger)
	require.Truef(t, ok, "logging instance is not a default-logger but %T", instance.impl)
	require.Equal(t, logger.logger.Writer(), os.Stderr)
}

func TestErrorCounting(t *testing.T) {
	reg := selfstat.Register(
		"gather",
		"errors",
		map[string]string{"input": "test"},
	)
	iLog := New("inputs", "test", "")
	iLog.RegisterErrorCallback(func() {
		reg.Incr(1)
	})
	iLog.Error("something went wrong")
	iLog.Errorf("something went wrong")

	require.Equal(t, int64(2), reg.Get())
}
