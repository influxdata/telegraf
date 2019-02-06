// +build windows

package x509_cert

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Make sure X509Cert implements telegraf.Input
var _ telegraf.Input = &X509Cert{}

func TestGatherWinStore(t *testing.T) {
	sc := X509Cert{
		Sources: []string{"LocalMachine/Root"},
	}
	t.Run("try to open LocalMachine/Root and load certs", func(t *testing.T) {
		var acc testutil.Accumulator
		err := sc.Gather(&acc)
		require.NoError(t, err)
		assert.True(t, acc.HasMeasurement("x509_cert"))
	})
}
