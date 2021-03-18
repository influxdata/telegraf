package supervisor

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestShort_SampleData(t *testing.T) {
	sampleProcessInfo := make([]ProcessInfo, 2)
	expectedProcessFields := make([]map[string]interface{}, 2)
	expectedProcessTags := make([]map[string]string, 2)

	sampleProcessInfo[0] = ProcessInfo{
		Name:          "Process0",
		Group:         "ProcessGroup0",
		Description:   "pid 112 uptime 0:12:11",
		Start:         1615632853,
		Stop:          0,
		Now:           1615632853 + 731,
		State:         20,
		Statename:     "RUNNING",
		StdoutLogfile: "/var/log/supervisor/process0-stdout.log",
		StderrLogfile: "/var/log/supervisor/process0-stdout.log",
		SpawnErr:      "",
		ExitStatus:    0,
		Pid:           112,
	}

	sampleProcessInfo[1] = ProcessInfo{
		Name:          "Process1",
		Group:         "ProcessGroup1",
		Description:   "pid 113 uptime 0:12:11",
		Start:         1615632853,
		Stop:          0,
		Now:           1615632853 + 731,
		State:         20,
		Statename:     "RUNNING",
		StdoutLogfile: "/var/log/supervisor/process1-stdout.log",
		StderrLogfile: "/var/log/supervisor/process1-stderr.log",
		SpawnErr:      "",
		ExitStatus:    0,
		Pid:           113,
	}

	expectedProcessTags[0] = map[string]string{
		"process": "Process0",
		"group":   "ProcessGroup0",
		"server":  "sampleInstance",
	}

	expectedProcessFields[0] = map[string]interface{}{
		"uptime":   int32(731),
		"state":    int16(20),
		"pid":      int32(112),
		"exitCode": int8(0),
	}

	expectedProcessTags[1] = map[string]string{
		"process": "Process1",
		"group":   "ProcessGroup1",
		"server":  "sampleInstance",
	}

	expectedProcessFields[1] = map[string]interface{}{
		"uptime":   int32(731),
		"state":    int16(20),
		"pid":      int32(113),
		"exitCode": int8(0),
	}

	expectedInstanceTags := map[string]string{
		"server": "sampleInstance",
	}

	expectedInstanceFields := map[string]interface{}{
		"state": int8(1),
	}

	s := &Supervisor{
		PidGather:      true,
		ExitCodeGather: true,
		UseIdentTag:    true,
		Status: SupervisorInfo{
			StateCode: int8(1),
			StateName: "RUNNING",
			Ident:     "sampleInstance",
		},
	}

	for key, process := range sampleProcessInfo {
		tags, fields, err := s.parseProcessData(process)
		assert.NoError(t, err)
		assert.Equal(t, expectedProcessTags[key], tags)
		assert.Equal(t, expectedProcessFields[key], fields)
	}

	instanceTags, instanceFields, err := s.parseInstanceData()
	assert.NoError(t, err)
	assert.Equal(t, expectedInstanceTags, instanceTags)
	assert.Equal(t, expectedInstanceFields, instanceFields)
}

func TestIntegration_BasicGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	s := &Supervisor{
		Server:         "http://" + testutil.GetLocalHost() + ":9001/RPC2",
		PidGather:      true,
		ExitCodeGather: true,
		UseIdentTag:    true,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(s.Gather)
	require.NoError(t, err)
	assert.Equal(t, acc.HasField("supervisor_processes", "uptime"), true)
	assert.Equal(t, acc.HasField("supervisor_processes", "state"), true)
	assert.Equal(t, acc.HasField("supervisor_processes", "pid"), true)
	assert.Equal(t, acc.HasField("supervisor_processes", "exitCode"), true)
	assert.Equal(t, acc.HasField("supervisor_instance", "state"), true)
}
