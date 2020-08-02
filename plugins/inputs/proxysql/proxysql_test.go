package proxysql

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxysqlDefaultsToLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Proxysql{
		Servers: []string{fmt.Sprintf("radmin:rpass@tcp(%s:6032)/", testutil.GetLocalHost())},
	}

	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("proxysql"))
}
