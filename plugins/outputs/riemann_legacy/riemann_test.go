package riemann_legacy

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	t.Skip("Skipping legacy integration test")

	url := testutil.GetLocalHost() + ":5555"

	r := &Riemann{
		URL:       url,
		Transport: "tcp",
	}

	err := r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
