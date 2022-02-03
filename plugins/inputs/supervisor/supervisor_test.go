package supervisor

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestShort_SampleData(t *testing.T) {
	testCases := []struct {
		desc              string
		supervisorData    supervisorInfo
		sampleProcInfo    []processInfo
		expProcessFields  []map[string]interface{}
		expProcessTags    []map[string]string
		expInstanceFields map[string]interface{}
		expInstancesTags  map[string]string
	}{
		{
			desc: "Case 1",
			sampleProcInfo: []processInfo{
				{
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
				},
				{
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
				},
			},
			supervisorData: supervisorInfo{
				StateCode: int8(1),
			},
			expProcessFields: []map[string]interface{}{
				{
					"uptime":   int32(731),
					"state":    int16(20),
					"pid":      int32(112),
					"exitCode": int8(0),
				},
				{
					"uptime":   int32(731),
					"state":    int16(20),
					"pid":      int32(113),
					"exitCode": int8(0),
				},
			},
			expProcessTags: []map[string]string{
				{
					"process": "Process0",
					"group":   "ProcessGroup0",
					"server":  "example.org:9001",
				},
				{
					"process": "Process1",
					"group":   "ProcessGroup1",
					"server":  "example.org:9001",
				},
			},
			expInstanceFields: map[string]interface{}{
				"state": int8(1),
			},
			expInstancesTags: map[string]string{
				"server": "example.org:9001",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s := &Supervisor{
				Server:      "http://example.org:9001/RPC2",
				UseIdentTag: false,
				MetricsInc:  []string{},
				MetricsExc:  []string{},
			}
			status := supervisorInfo{
				StateCode: tC.supervisorData.StateCode,
				StateName: tC.supervisorData.StateName,
			}
			s.Init()
			for k, v := range tC.sampleProcInfo {
				processTags, processFields, err := s.parseProcessData(v, status)
				require.NoError(t, err)
				require.Equal(t, tC.expProcessFields[k], processFields)
				require.Equal(t, tC.expProcessTags[k], processTags)
			}
			instanceTags, instanceFields, err := s.parseInstanceData(status)
			require.NoError(t, err)
			require.Equal(t, tC.expInstancesTags, instanceTags)
			require.Equal(t, tC.expInstanceFields, instanceFields)
		})
	}
}

func TestIntegration_BasicGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	s := &Supervisor{
		Server:      "http://login:pass@" + testutil.GetLocalHost() + ":9001/RPC2",
		UseIdentTag: true,
		MetricsInc:  []string{},
		MetricsExc:  []string{},
	}
	s.Init()

	var acc testutil.Accumulator
	err := acc.GatherError(s.Gather)
	require.NoError(t, err)
	require.Equal(t, acc.HasField("supervisor_processes", "uptime"), true)
	require.Equal(t, acc.HasField("supervisor_processes", "state"), true)
	require.Equal(t, acc.HasField("supervisor_processes", "pid"), true)
	require.Equal(t, acc.HasField("supervisor_processes", "exitCode"), true)
	require.Equal(t, acc.HasField("supervisor_instance", "state"), true)
}
