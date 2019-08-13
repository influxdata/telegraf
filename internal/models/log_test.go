package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorCount(t *testing.T) {
	iErrors.Set(0)
	aErrors.Set(0)
	oErrors.Set(0)
	pErrors.Set(0)

	log := Logger{Name: "inputs.test"}
	log.Errorf("something went wrong")
	log.Error("something went wrong")

	log.Name = "aggregators.test"
	log.Error("another thing happened")

	log.Name = "outputs.test"
	log.Error("another thing happened")

	log.Name = "processors.test"
	log.Error("another thing happened")

	require.Equal(t, int64(2), iErrors.Get())
	require.Equal(t, int64(1), aErrors.Get())
	require.Equal(t, int64(1), oErrors.Get())
	require.Equal(t, int64(1), pErrors.Get())
}

func TestPluginConfig(t *testing.T) {
	iErrors.Set(0)
	p := PluginConfig{Log: Logger{Name: "inputs.test"}}
	log := p.Logger()

	log.Debugf("something happened")
	log.Debug("something happened")

	log.Warnf("something happened")
	log.Warn("something happened")
	require.Equal(t, int64(0), iErrors.Get())

	log.Infof("something happened")
	log.Info("something happened")
	require.Equal(t, int64(0), iErrors.Get())

	log.Errorf("something happened")
	log.Error("something happened")
	require.Equal(t, int64(2), iErrors.Get())
}
