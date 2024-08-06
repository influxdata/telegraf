package slurm

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	goslurm "github.com/pcolladosoto/goslurm/v0038"
	"github.com/stretchr/testify/require"
)

func TestURLs(t *testing.T) {
	for _, url := range []string{"http://example.com:6820", "https://example.com:6820"} {
		plugin := Slurm{
			URL: url,
		}
		require.NoError(t, plugin.Init())
	}

	for _, url := range []string{"httpp://example.com:6820", "httpss://example.com:6820"} {
		plugin := Slurm{
			URL: url,
		}
		require.Error(t, plugin.Init())
	}
}

func TestPanicHandling(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slurm/v0.0.38/diag":
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{
				"meta": {},
				"errors": [],
				"statistics": {
					"rpcs_by_message_type": [],
					"rpcs_by_user": [],
					"jobs_running": 100
				}
			}`))
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNotFound)
			t.Fatalf("unexpected path: " + r.URL.Path)
		}
	})

	plugin := &Slurm{
		URL:              "http://" + ts.Listener.Addr().String(),
		IgnoredEndpoints: []string{"jobs", "nodes", "partitions", "reservations"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NotPanics(t, func() { plugin.Gather(&acc) })
}

func TestGatherDiagMetrics(t *testing.T) {
	var (
		bfActive                = false
		bfQueueLen        int32 = 1
		bfQueueLenMean    int32 = 1
		jobsCanceled      int32 = 3
		jobsCompleted     int32 = 396
		jobsFailed        int32 = 2
		jobsPending       int32 = 10
		jobsRunning       int32 = 100
		jobsStarted       int32 = 396
		jobsSubmitted     int32 = 396
		scheduleCycleLast int32 = 301
		scheduleCycleMean int32 = 137
		serverThreadCount int32 = 3
	)
	diag := goslurm.V0038DiagStatistics{
		BfActive:          &bfActive,
		BfQueueLen:        &bfQueueLen,
		BfQueueLenMean:    &bfQueueLenMean,
		JobsCanceled:      &jobsCanceled,
		JobsCompleted:     &jobsCompleted,
		JobsFailed:        &jobsFailed,
		JobsPending:       &jobsPending,
		JobsRunning:       &jobsRunning,
		JobsStarted:       &jobsStarted,
		JobsSubmitted:     &jobsSubmitted,
		ScheduleCycleLast: &scheduleCycleLast,
		ScheduleCycleMean: &scheduleCycleMean,
		ServerThreadCount: &serverThreadCount,
	}

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["url"] = "127.0.0.1"

	records["bf_active"] = bfActive
	records["bf_queue_len"] = bfQueueLen
	records["bf_queue_len_mean"] = bfQueueLenMean
	records["jobs_canceled"] = jobsCanceled
	records["jobs_submitted"] = jobsSubmitted
	records["jobs_started"] = jobsStarted
	records["jobs_completed"] = jobsCompleted
	records["jobs_failed"] = jobsFailed
	records["jobs_pending"] = jobsPending
	records["jobs_running"] = jobsRunning
	records["schedule_cycle_last"] = scheduleCycleLast
	records["schedule_cycle_mean"] = scheduleCycleMean
	records["server_thread_count"] = serverThreadCount

	plugin := &Slurm{
		URL: "http://127.0.0.1:6820",
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	plugin.GatherDiagMetrics(&acc, &diag)
	acc.AssertContainsTaggedFields(t, "slurm_diag", records, tags)
}

func TestGatherJobsMetrics(t *testing.T) {
	var (
		jobName                       = "gridjob"
		jobID                   int32 = 17489
		jobState                      = "RUNNING"
		stateReason                   = "None"
		partition                     = "atlas"
		nodes                         = "naboo222,naboo223"
		nodeCount               int32 = 2
		priority                int64 = 4294884242
		nice                    int32 = 50
		groupID                 int32 = 2005
		command                       = "/tmp/SLURM_job_script.jDwqdW"
		standardOutput                = "/home/sessiondir/IqBMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmLiIKDmRqth1m.comment"
		standardError                 = "/home/sessiondir/IqBMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmLiIKDmRqth1m.comment"
		standardInput                 = "/dev/null"
		currentWorkingDirectory       = "/home/sessiondir/IqBMDmQY2t5nKG01gq4B3BRpm7wtQmABFKDmbnHPDmLiIKDmRqth1m"
		submitTime              int64 = 1722598613
		startTime               int64 = 1722598614
		cpus                    int32 = 1
		tasks                   int32 = 1
		timeLimit               int64 = 3600
		tresReqStr                    = "cpu=1,mem=2000M,node=1,billing=1"
	)
	jobs := []goslurm.V0038JobResponseProperties{
		{
			Name:                    &jobName,
			JobId:                   &jobID,
			JobState:                &jobState,
			StateReason:             &stateReason,
			Partition:               &partition,
			Nodes:                   &nodes,
			NodeCount:               &nodeCount,
			Priority:                &priority,
			Nice:                    &nice,
			GroupId:                 &groupID,
			Command:                 &command,
			StandardOutput:          &standardOutput,
			StandardError:           &standardError,
			StandardInput:           &standardInput,
			CurrentWorkingDirectory: &currentWorkingDirectory,
			SubmitTime:              &submitTime,
			StartTime:               &startTime,
			Cpus:                    &cpus,
			Tasks:                   &tasks,
			TimeLimit:               &timeLimit,
			TresReqStr:              &tresReqStr,
		},
	}

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["url"] = "127.0.0.1"
	tags["name"] = jobName
	tags["job_id"] = strconv.Itoa(int(jobID))

	records["state"] = jobState
	records["state_reason"] = stateReason
	records["partition"] = partition
	records["nodes"] = nodes
	records["node_count"] = nodeCount
	records["priority"] = priority
	records["nice"] = nice
	records["group_id"] = groupID
	records["command"] = command
	records["standard_output"] = standardOutput
	records["standard_error"] = standardError
	records["standard_input"] = standardInput
	records["current_working_directory"] = currentWorkingDirectory
	records["submit_time"] = submitTime
	records["start_time"] = startTime
	records["cpus"] = cpus
	records["tasks"] = tasks
	records["time_limit"] = timeLimit
	records["tres_req_str"] = tresReqStr

	plugin := &Slurm{
		URL: "http://127.0.0.1:6820",
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	plugin.GatherJobsMetrics(&acc, jobs)
	acc.AssertContainsTaggedFields(t, "slurm_jobs", records, tags)
}

func TestGatherNodesMetrics(t *testing.T) {
	var (
		name                = "naboo145"
		state               = "idle"
		cores         int32 = 32
		cpus          int32 = 64
		cpuLoad       int64 = 910
		allocCpus     int64 = 16
		realMemory    int32 = 104223
		freeMemory    int32 = 105203
		allocMemory   int64 = 10
		tres                = "cpu=64,mem=127901M,billing=64"
		tresUsed            = "cpu=8,mem=16000M"
		weight        int32 = 1
		slurmdVersion       = "22.05.9"
		architecture        = "x86_64"
	)
	nodes := []goslurm.V0038Node{
		{
			Name:          &name,
			State:         &state,
			Cores:         &cores,
			Cpus:          &cpus,
			CpuLoad:       &cpuLoad,
			AllocCpus:     &allocCpus,
			RealMemory:    &realMemory,
			FreeMemory:    &freeMemory,
			AllocMemory:   &allocMemory,
			Tres:          &tres,
			TresUsed:      &tresUsed,
			Weight:        &weight,
			SlurmdVersion: &slurmdVersion,
			Architecture:  &architecture,
		},
	}

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["url"] = "127.0.0.1"
	tags["name"] = name

	records["state"] = state
	records["cores"] = cores
	records["cpus"] = cpus
	records["cpu_load"] = cpuLoad
	records["alloc_cpu"] = allocCpus
	records["real_memory"] = realMemory
	records["free_memory"] = freeMemory
	records["alloc_memory"] = allocMemory
	records["tres"] = tres
	records["tres_used"] = tresUsed
	records["weight"] = weight
	records["slurmd_version"] = slurmdVersion
	records["architecture"] = architecture

	plugin := &Slurm{
		URL: "http://127.0.0.1:6820",
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	plugin.GatherNodesMetrics(&acc, nodes)
	acc.AssertContainsTaggedFields(t, "slurm_nodes", records, tags)
}

func TestGatherPartitionsMetrics(t *testing.T) {
	var (
		name             = "atlas"
		state            = "UP"
		totalCpus  int32 = 288
		totalNodes int32 = 6
		nodes            = "naboo145,naboo146,naboo147,naboo216,naboo219,naboo222"
		tres             = "cpu=288,mem=14157M,node=6,billing=288"
	)
	partitions := []goslurm.V0038Partition{
		{
			Name:       &name,
			State:      &state,
			TotalCpus:  &totalCpus,
			TotalNodes: &totalNodes,
			Nodes:      &nodes,
			Tres:       &tres,
		},
	}

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["url"] = "127.0.0.1"
	tags["name"] = name

	records["state"] = state
	records["total_cpu"] = totalCpus
	records["total_nodes"] = totalNodes
	records["nodes"] = nodes
	records["tres"] = tres

	plugin := &Slurm{
		URL: "http://127.0.0.1:6820",
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	plugin.GatherPartitionsMetrics(&acc, partitions)
	acc.AssertContainsTaggedFields(t, "slurm_partitions", records, tags)
}

func TestGatherReservationsMetrics(t *testing.T) {
	var (
		name                = "foo"
		coreCount     int32 = 10
		coreSpecCount int32 = 15
		groups              = "users"
		users               = "me"
		startTime     int32 = 1722598614
		partition           = "atlas"
		accounts            = "physicists"
		nodeCount     int32 = 5
		nodeList            = "naboo123,naboo321"
	)
	reservations := []goslurm.V0038Reservation{
		{
			Name:        &name,
			CoreCount:   &coreCount,
			CoreSpecCnt: &coreSpecCount,
			Groups:      &groups,
			Users:       &users,
			StartTime:   &startTime,
			Partition:   &partition,
			Accounts:    &accounts,
			NodeCount:   &nodeCount,
			NodeList:    &nodeList,
		},
	}

	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["url"] = "127.0.0.1"
	tags["name"] = name

	records["core_count"] = coreCount
	records["core_spec_count"] = coreSpecCount
	records["groups"] = groups
	records["users"] = users
	records["start_time"] = startTime
	records["partition"] = partition
	records["accounts"] = accounts
	records["node_count"] = nodeCount
	records["node_list"] = nodeList

	plugin := &Slurm{
		URL: "http://127.0.0.1:6820",
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	plugin.GatherReservationsMetrics(&acc, reservations)
	acc.AssertContainsTaggedFields(t, "slurm_reservations", records, tags)
}
