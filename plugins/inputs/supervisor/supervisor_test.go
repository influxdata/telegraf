package supervisor

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupervisorConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s := getConfig()

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)
}

func TestSupervisor_ValidateMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s := getConfig()

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	assert.Equal(t, acc.NFields(), 26)

	intMetrics := []string{
		"Start",
		"Stop",
		"Now",
		"State",
		"ExitStatus",
		"Pid",
	}

	var metricsCounted = 0
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("supervisor", metric))
		metricsCounted++
	}

	assert.Equal(t, metricsCounted, 6)
}

func TestSupervisor_MetricsContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s := getConfig()

	var acc testutil.Accumulator
	err := s.Gather(&acc)
	require.NoError(t, err)

	var foundProcess1 = false
	var foundProcess2 = false

	for _, pnt := range acc.Metrics {
		if pnt.Measurement == "supervisor" {
			if pnt.Tags["server"] == s.Host && pnt.Tags["process"] == "process-1" {
				foundProcess1 = true
				validateProcess(t, pnt.Fields, "process-1")
			}
		}
		if pnt.Measurement == "supervisor" {
			if pnt.Tags["server"] == s.Host && pnt.Tags["process"] == "process-2" {
				foundProcess2 = true
				validateProcess(t, pnt.Fields, "process-2")
			}
		}
	}

	assert.True(t, foundProcess1)
	assert.True(t, foundProcess2)
}

func getConfig() (s *Supervisor) {
	host := fmt.Sprintf("http://" + testutil.GetLocalHost() + ":9001/RPC2")
	s = &Supervisor{
		Host: host,
	}
	return
}

func validateProcess(t *testing.T, fields map[string]interface{}, name string) {
	assert.Equal(t, fields["Name"], name)
	assert.Equal(t, fields["Group"], name)
	assert.True(t, strings.Contains(fields["Description"].(string), "pid"))
	assert.True(t, strings.Contains(fields["Description"].(string), "uptime"))
	assert.True(t, fields["Start"].(int64) > 0)
	assert.Equal(t, fields["Stop"], int64(0))
	assert.True(t, fields["Now"].(int64) > fields["Start"].(int64))
	assert.InDelta(t,
		time.Now().Unix(),
		fields["Now"].(int64),
		2, "The difference should not be more than 2s")
	assert.Equal(t, fields["State"], int64(20))
	assert.Equal(t, fields["Statename"], "RUNNING")
	assert.True(t, strings.Contains(fields["StdoutLogfile"].(string), "stdout"))
	assert.True(t, strings.Contains(fields["StderrLogfile"].(string), "stderr"))
	assert.Equal(t, fields["SpawnErr"], "")
	assert.Equal(t, fields["ExitStatus"], int64(0))

	pidstr := strings.TrimLeft(strings.Split(fields["Description"].(string), ",")[0], "pid ")
	pid, _ := strconv.Atoi(pidstr)
	assert.Equal(t, fields["Pid"].(int64), int64(pid))
}
