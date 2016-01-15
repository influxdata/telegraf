package sqlserver

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSqlServerGeneratesMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s := &SqlServer{}
	s.Servers = append(s.Servers, &Server{ConnectionString: "Server=192.168.1.30;User Id=linuxuser;Password=linuxuser;app name=telegraf;log=1;"})
	
	var acc testutil.Accumulator

	err := s.Gather(&acc)
	require.NoError(t, err)
}