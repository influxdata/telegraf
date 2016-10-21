package warp10

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWriteWarp10(t *testing.T) {
	w := Warp10{
		Prefix:  "unit.test",
		WarpUrl: "http://localhost:8090",
		Token:   "WRITE",
		Debug:   false,
	}

	//err := i.Connect()
	//require.NoError(t, err)
	err := w.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
