package kinesis

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

func TestLifecycle(t *testing.T) {
	plugin := &Kinesis{}
	err := plugin.Init()
	require.NoError(t, err)

	err = plugin.Connect()
	require.NoError(t, err)

	err = plugin.Write([]telegraf.Metric{})
	require.NoError(t, err)

	err = plugin.Close()
	require.NoError(t, err)
}
