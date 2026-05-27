package slurm

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	goslurm "github.com/pcolladosoto/goslurm/v0038"

	"github.com/influxdata/telegraf"
)

type slurmV0038 struct {
	client   *goslurm.APIClient
	username string
	token    string
}

func newV0038Client(host, scheme, userAgent string, httpClient *http.Client, username, token string) *slurmV0038 {
	cfg := goslurm.NewConfiguration()
	cfg.Host = host
	cfg.Scheme = scheme
	cfg.UserAgent = userAgent
	cfg.HTTPClient = httpClient
	return &slurmV0038{
		client:   goslurm.NewAPIClient(cfg),
		username: username,
		token:    token,
	}
}

func (s *slurmV0038) authCtx() context.Context {
	return context.WithValue(
		context.Background(),
		goslurm.ContextAPIKeys,
		map[string]goslurm.APIKey{
			"user":  {Key: s.username},
			"token": {Key: s.token},
		},
	)
}

func (s *slurmV0038) gatherDiag(acc telegraf.Accumulator, source string) error {
	resp, raw, err := s.client.SlurmAPI.SlurmV0038Diag(s.authCtx()).Execute()
	if err != nil {
		return fmt.Errorf("error getting diag: %w", err)
	}
	raw.Body.Close()

	diag, ok := resp.GetStatisticsOk()
	if !ok {
		return nil
	}

	records := make(map[string]interface{}, 13)
	tags := map[string]string{"source": source}

	if v, ok := diag.GetServerThreadCountOk(); ok {
		records["server_thread_count"] = *v
	}
	if v, ok := diag.GetJobsCanceledOk(); ok {
		records["jobs_canceled"] = *v
	}
	if v, ok := diag.GetJobsSubmittedOk(); ok {
		records["jobs_submitted"] = *v
	}
	if v, ok := diag.GetJobsStartedOk(); ok {
		records["jobs_started"] = *v
	}
	if v, ok := diag.GetJobsCompletedOk(); ok {
		records["jobs_completed"] = *v
	}
	if v, ok := diag.GetJobsFailedOk(); ok {
		records["jobs_failed"] = *v
	}
	if v, ok := diag.GetJobsPendingOk(); ok {
		records["jobs_pending"] = *v
	}
	if v, ok := diag.GetJobsRunningOk(); ok {
		records["jobs_running"] = *v
	}
	if v, ok := diag.GetScheduleCycleLastOk(); ok {
		records["schedule_cycle_last"] = *v
	}
	if v, ok := diag.GetScheduleCycleMeanOk(); ok {
		records["schedule_cycle_mean"] = *v
	}
	if v, ok := diag.GetBfQueueLenOk(); ok {
		records["bf_queue_len"] = *v
	}
	if v, ok := diag.GetBfQueueLenMeanOk(); ok {
		records["bf_queue_len_mean"] = *v
	}
	if v, ok := diag.GetBfActiveOk(); ok {
		records["bf_active"] = *v
	}

	acc.AddFields("slurm_diag", records, tags)
	return nil
}

func (s *slurmV0038) gatherJobs(acc telegraf.Accumulator, source string) error {
	resp, raw, err := s.client.SlurmAPI.SlurmV0038GetJobs(s.authCtx()).Execute()
	if err != nil {
		return fmt.Errorf("error getting jobs: %w", err)
	}
	raw.Body.Close()

	jobs, ok := resp.GetJobsOk()
	if !ok {
		return nil
	}

	for i := range jobs {
		records := make(map[string]interface{}, 21)
		tags := make(map[string]string, 3)

		tags["source"] = source
		if v, ok := jobs[i].GetNameOk(); ok {
			tags["name"] = *v
		}
		if v, ok := jobs[i].GetJobIdOk(); ok {
			tags["job_id"] = strconv.Itoa(int(*v))
		}

		if v, ok := jobs[i].GetJobStateOk(); ok {
			records["state"] = *v
		}
		if v, ok := jobs[i].GetStateReasonOk(); ok {
			records["state_reason"] = *v
		}
		if v, ok := jobs[i].GetPartitionOk(); ok {
			records["partition"] = *v
		}
		if v, ok := jobs[i].GetNodesOk(); ok {
			records["nodes"] = *v
		}
		if v, ok := jobs[i].GetNodeCountOk(); ok {
			records["node_count"] = *v
		}
		if v, ok := jobs[i].GetPriorityOk(); ok {
			records["priority"] = *v
		}
		if v, ok := jobs[i].GetNiceOk(); ok {
			records["nice"] = *v
		}
		if v, ok := jobs[i].GetGroupIdOk(); ok {
			records["group_id"] = *v
		}
		if v, ok := jobs[i].GetCommandOk(); ok {
			records["command"] = *v
		}
		if v, ok := jobs[i].GetStandardOutputOk(); ok {
			records["standard_output"] = strings.ReplaceAll(*v, "\\", "")
		}
		if v, ok := jobs[i].GetStandardErrorOk(); ok {
			records["standard_error"] = strings.ReplaceAll(*v, "\\", "")
		}
		if v, ok := jobs[i].GetStandardInputOk(); ok {
			records["standard_input"] = strings.ReplaceAll(*v, "\\", "")
		}
		if v, ok := jobs[i].GetCurrentWorkingDirectoryOk(); ok {
			records["current_working_directory"] = strings.ReplaceAll(*v, "\\", "")
		}
		if v, ok := jobs[i].GetSubmitTimeOk(); ok {
			records["submit_time"] = *v
		}
		if v, ok := jobs[i].GetStartTimeOk(); ok {
			records["start_time"] = *v
		}
		if v, ok := jobs[i].GetCpusOk(); ok {
			records["cpus"] = *v
		}
		if v, ok := jobs[i].GetTasksOk(); ok {
			records["tasks"] = *v
		}
		if v, ok := jobs[i].GetTimeLimitOk(); ok {
			records["time_limit"] = *v
		}
		if v, ok := jobs[i].GetTresReqStrOk(); ok {
			for k, val := range parseTres(*v) {
				records["tres_"+k] = val
			}
		}

		acc.AddFields("slurm_jobs", records, tags)
	}
	return nil
}

func (s *slurmV0038) gatherNodes(acc telegraf.Accumulator, source string) error {
	resp, raw, err := s.client.SlurmAPI.SlurmV0038GetNodes(s.authCtx()).Execute()
	if err != nil {
		return fmt.Errorf("error getting nodes: %w", err)
	}
	raw.Body.Close()

	nodes, ok := resp.GetNodesOk()
	if !ok {
		return nil
	}

	for _, node := range nodes {
		records := make(map[string]interface{}, 13)
		tags := make(map[string]string, 2)

		tags["source"] = source
		if v, ok := node.GetNameOk(); ok {
			tags["name"] = *v
		}

		if v, ok := node.GetStateOk(); ok {
			records["state"] = *v
		}
		if v, ok := node.GetCoresOk(); ok {
			records["cores"] = *v
		}
		if v, ok := node.GetCpusOk(); ok {
			records["cpus"] = *v
		}
		if v, ok := node.GetCpuLoadOk(); ok {
			records["cpu_load"] = *v
		}
		if v, ok := node.GetAllocCpusOk(); ok {
			records["alloc_cpu"] = *v
		}
		if v, ok := node.GetRealMemoryOk(); ok {
			records["real_memory"] = *v
		}
		if v, ok := node.GetFreeMemoryOk(); ok {
			records["free_memory"] = *v
		}
		if v, ok := node.GetAllocMemoryOk(); ok {
			records["alloc_memory"] = *v
		}
		if v, ok := node.GetTresOk(); ok {
			for k, val := range parseTres(*v) {
				records["tres_"+k] = val
			}
		}
		if v, ok := node.GetTresUsedOk(); ok {
			for k, val := range parseTres(*v) {
				records["tres_used_"+k] = val
			}
		}
		if v, ok := node.GetWeightOk(); ok {
			records["weight"] = *v
		}
		if v, ok := node.GetSlurmdVersionOk(); ok {
			records["slurmd_version"] = *v
		}
		if v, ok := node.GetArchitectureOk(); ok {
			records["architecture"] = *v
		}

		acc.AddFields("slurm_nodes", records, tags)
	}
	return nil
}

func (s *slurmV0038) gatherPartitions(acc telegraf.Accumulator, source string) error {
	resp, raw, err := s.client.SlurmAPI.SlurmV0038GetPartitions(s.authCtx()).Execute()
	if err != nil {
		return fmt.Errorf("error getting partitions: %w", err)
	}
	raw.Body.Close()

	partitions, ok := resp.GetPartitionsOk()
	if !ok {
		return nil
	}

	for _, partition := range partitions {
		records := make(map[string]interface{}, 5)
		tags := make(map[string]string, 2)

		tags["source"] = source
		if v, ok := partition.GetNameOk(); ok {
			tags["name"] = *v
		}

		if v, ok := partition.GetStateOk(); ok {
			records["state"] = *v
		}
		if v, ok := partition.GetTotalCpusOk(); ok {
			records["total_cpu"] = *v
		}
		if v, ok := partition.GetTotalNodesOk(); ok {
			records["total_nodes"] = *v
		}
		if v, ok := partition.GetNodesOk(); ok {
			records["nodes"] = *v
		}
		if v, ok := partition.GetTresOk(); ok {
			for k, val := range parseTres(*v) {
				records["tres_"+k] = val
			}
		}

		acc.AddFields("slurm_partitions", records, tags)
	}
	return nil
}

func (s *slurmV0038) gatherReservations(acc telegraf.Accumulator, source string) error {
	resp, raw, err := s.client.SlurmAPI.SlurmV0038GetReservations(s.authCtx()).Execute()
	if err != nil {
		return fmt.Errorf("error getting reservations: %w", err)
	}
	raw.Body.Close()

	reservations, ok := resp.GetReservationsOk()
	if !ok {
		return nil
	}

	for _, reservation := range reservations {
		records := make(map[string]interface{}, 9)
		tags := make(map[string]string, 2)

		tags["source"] = source
		if v, ok := reservation.GetNameOk(); ok {
			tags["name"] = *v
		}

		if v, ok := reservation.GetCoreCountOk(); ok {
			records["core_count"] = *v
		}
		if v, ok := reservation.GetCoreSpecCntOk(); ok {
			records["core_spec_count"] = *v
		}
		if v, ok := reservation.GetGroupsOk(); ok {
			records["groups"] = *v
		}
		if v, ok := reservation.GetUsersOk(); ok {
			records["users"] = *v
		}
		if v, ok := reservation.GetStartTimeOk(); ok {
			records["start_time"] = *v
		}
		if v, ok := reservation.GetPartitionOk(); ok {
			records["partition"] = *v
		}
		if v, ok := reservation.GetAccountsOk(); ok {
			records["accounts"] = *v
		}
		if v, ok := reservation.GetNodeCountOk(); ok {
			records["node_count"] = *v
		}
		if v, ok := reservation.GetNodeListOk(); ok {
			records["node_list"] = *v
		}

		acc.AddFields("slurm_reservations", records, tags)
	}
	return nil
}
