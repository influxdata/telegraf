package models

import (
	"testing"

	"github.com/influxdata/telegraf/selfstat"
	"github.com/stretchr/testify/require"
)

func TestErrorCounting(t *testing.T) {
	iLog := Logger{Name: "inputs.test", Errs: selfstat.Register(
		"gather",
		"errors",
		map[string]string{"input": "test"},
	)}
	iLog.Error("something went wrong")
	iLog.Errorf("something went wrong")

	aLog := Logger{Name: "aggregators.test", Errs: selfstat.Register(
		"aggregate",
		"errors",
		map[string]string{"aggregator": "test"},
	)}
	aLog.Name = "aggregators.test"
	aLog.Error("another thing happened")

	oLog := Logger{Name: "outputs.test", Errs: selfstat.Register(
		"write",
		"errors",
		map[string]string{"output": "test"},
	)}
	oLog.Error("another thing happened")

	pLog := Logger{Name: "processors.test", Errs: selfstat.Register(
		"process",
		"errors",
		map[string]string{"processor": "test"},
	)}
	pLog.Error("another thing happened")

	require.Equal(t, int64(2), iLog.Errs.Get())
	require.Equal(t, int64(1), aLog.Errs.Get())
	require.Equal(t, int64(1), oLog.Errs.Get())
	require.Equal(t, int64(1), pLog.Errs.Get())
}

func TestLogging(t *testing.T) {
	log := Logger{Name: "inputs.test", Errs: selfstat.Register(
		"gather",
		"errors",
		map[string]string{"input": "test"},
	)}

	log.Errs.Set(0)

	log.Debugf("something happened")
	log.Debug("something happened")

	log.Warnf("something happened")
	log.Warn("something happened")
	require.Equal(t, int64(0), log.Errs.Get())

	log.Infof("something happened")
	log.Info("something happened")
	require.Equal(t, int64(0), log.Errs.Get())

	log.Errorf("something happened")
	log.Error("something happened")
	require.Equal(t, int64(2), log.Errs.Get())
}
