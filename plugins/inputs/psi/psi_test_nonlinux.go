//go:build !linux

package psi

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestPSIStats(t *testing.T) {
	var (
		psi *Psi
		err error
		acc testutil.Accumulator
	)

	err = psi.Gather(&acc)
	require.NoError(t, err)
}
