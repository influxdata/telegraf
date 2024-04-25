package models

import (
	"testing"

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
