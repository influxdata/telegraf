package riemann_legacy

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	// if this needs to run use: `docker run stealthly/docker-riemann`
	t.Skip("Skipping legacy integration test")

	url := testutil.GetLocalHost() + ":5555"

	r := &Riemann{
		URL:       url,
		Transport: "tcp",
		Log:       testutil.Logger{},
	}

	err := r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
