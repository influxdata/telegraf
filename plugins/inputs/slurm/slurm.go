//go:generate ../../../tools/readme_config_includer/generator
package slurm

import (
	"context"
	stdTls "crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"

	goslurm "github.com/pcolladosoto/goslurm/v0038"
)

//go:embed sample.conf
var sampleConfig string

type Slurm struct {
	URL              string          `toml:"url"`
	Username         string          `toml:"username"`
	Token            string          `toml:"token"`
	IgnoredEndpoints []string        `toml:"ignored_endpoints"`
	ResponseTimeout  config.Duration `toml:"response_timeout"`
	tls.ClientConfig

	client      *goslurm.APIClient
	baseURL     *url.URL
	endpointMap map[string]bool
}

func (*Slurm) SampleConfig() string {
	return sampleConfig
}

func (s *Slurm) Init() error {
	if s.ResponseTimeout == 0 {
		s.ResponseTimeout = config.Duration(time.Second * 5)
	}

	s.endpointMap = map[string]bool{}
	for _, endpoint := range s.IgnoredEndpoints {
		s.endpointMap[strings.ToLower(endpoint)] = true
	}

	if s.URL == "" {
		return errors.New("empty URL provided")
	}

	u, err := url.Parse(s.URL)
	if err != nil {
		return err
	}

	if u.Hostname() == "" {
		return fmt.Errorf("invalid hostname %q", u.Hostname())
	}

	s.baseURL = u

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid scheme %q", u.Scheme)
	}

	var tlsCfg *stdTls.Config
	if u.Scheme == "https" {
		tlsCfg, err = s.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
	}

	configuration := goslurm.NewConfiguration()
	configuration.Host = u.Host
	configuration.Scheme = u.Scheme
	configuration.UserAgent = internal.ProductToken()
	configuration.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(s.ResponseTimeout),
	}

	s.client = goslurm.NewAPIClient(configuration)

	return nil
}

func (s *Slurm) gatherDiagMetrics(acc telegraf.Accumulator,
	diag *goslurm.V0038DiagStatistics) {
	records := make(map[string]interface{})
	tags := make(map[string]string)

	var (
		int32Ptr *int32
		boolPtr  *bool
		ok       bool
	)

	tags["source"] = s.baseURL.Hostname()

	tc, ok := diag.GetServerThreadCountOk()
	if !ok {
		return
	}
	records["server_thread_count"] = *tc

	int32Ptr, ok = diag.GetJobsCanceledOk()
	if !ok {
		return
	}
	records["jobs_canceled"] = *int32Ptr

	int32Ptr, ok = diag.GetJobsSubmittedOk()
	if !ok {
		return
	}
	records["jobs_submitted"] = *int32Ptr

	int32Ptr, ok = diag.GetJobsStartedOk()
	if !ok {
		return
	}
	records["jobs_started"] = *int32Ptr

	int32Ptr, ok = diag.GetJobsCompletedOk()
	if !ok {
		return
	}
	records["jobs_completed"] = *int32Ptr

	int32Ptr, ok = diag.GetJobsFailedOk()
	if !ok {
		return
	}
	records["jobs_failed"] = *int32Ptr

	int32Ptr, ok = diag.GetJobsPendingOk()
	if !ok {
		return
	}
	records["jobs_pending"] = *int32Ptr

	int32Ptr, ok = diag.GetJobsRunningOk()
	if !ok {
		return
	}
	records["jobs_running"] = *int32Ptr

	int32Ptr, ok = diag.GetScheduleCycleLastOk()
	if !ok {
		return
	}
	records["schedule_cycle_last"] = *int32Ptr

	int32Ptr, ok = diag.GetScheduleCycleMeanOk()
	if !ok {
		return
	}
	records["schedule_cycle_mean"] = *int32Ptr

	int32Ptr, ok = diag.GetBfQueueLenOk()
	if !ok {
		return
	}
	records["bf_queue_len"] = *int32Ptr

	int32Ptr, ok = diag.GetBfQueueLenMeanOk()
	if !ok {
		return
	}
	records["bf_queue_len_mean"] = *int32Ptr

	boolPtr, ok = diag.GetBfActiveOk()
	if !ok {
		return
	}
	records["bf_active"] = *boolPtr

	acc.AddFields("slurm_diag", records, tags)
}

func (s *Slurm) gatherJobsMetrics(acc telegraf.Accumulator,
	jobs []goslurm.V0038JobResponseProperties) {
	var (
		int32Ptr *int32
		int64Ptr *int64
		strPtr   *string
		ok       bool
	)

	for i := range jobs {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["source"] = s.baseURL.Hostname()
		strPtr, ok = jobs[i].GetNameOk()
		if !ok {
			continue
		}
		tags["name"] = *strPtr
		int32Ptr, ok = jobs[i].GetJobIdOk()
		if !ok {
			continue
		}
		tags["job_id"] = strconv.Itoa(int(*int32Ptr))

		strPtr, ok = jobs[i].GetJobStateOk()
		if !ok {
			continue
		}
		records["state"] = *strPtr

		strPtr, ok = jobs[i].GetStateReasonOk()
		if !ok {
			continue
		}
		records["state_reason"] = *strPtr

		strPtr, ok = jobs[i].GetPartitionOk()
		if !ok {
			continue
		}
		records["partition"] = *strPtr

		strPtr, ok = jobs[i].GetNodesOk()
		if !ok {
			continue
		}
		records["nodes"] = *strPtr

		int32Ptr, ok = jobs[i].GetNodeCountOk()
		if !ok {
			continue
		}
		records["node_count"] = *int32Ptr

		int64Ptr, ok = jobs[i].GetPriorityOk()
		if !ok {
			continue
		}
		records["priority"] = *int64Ptr

		int32Ptr, ok = jobs[i].GetNiceOk()
		if !ok {
			continue
		}
		records["nice"] = *int32Ptr

		int32Ptr, ok = jobs[i].GetGroupIdOk()
		if !ok {
			continue
		}
		records["group_id"] = *int32Ptr

		strPtr, ok = jobs[i].GetCommandOk()
		if !ok {
			continue
		}
		records["command"] = *strPtr

		strPtr, ok = jobs[i].GetStandardOutputOk()
		if !ok {
			continue
		}
		records["standard_output"] = strings.ReplaceAll(*strPtr, "\\", "")

		strPtr, ok = jobs[i].GetStandardErrorOk()
		if !ok {
			continue
		}
		records["standard_error"] = strings.ReplaceAll(*strPtr, "\\", "")

		strPtr, ok = jobs[i].GetStandardInputOk()
		if !ok {
			continue
		}
		records["standard_input"] = strings.ReplaceAll(*strPtr, "\\", "")

		strPtr, ok = jobs[i].GetCurrentWorkingDirectoryOk()
		if !ok {
			continue
		}
		records["current_working_directory"] = strings.ReplaceAll(
			*strPtr, "\\", "")

		int64Ptr, ok = jobs[i].GetSubmitTimeOk()
		if !ok {
			continue
		}
		records["submit_time"] = *int64Ptr

		int64Ptr, ok = jobs[i].GetStartTimeOk()
		if !ok {
			continue
		}
		records["start_time"] = *int64Ptr

		int32Ptr, ok := jobs[i].GetCpusOk()
		if !ok {
			continue
		}
		records["cpus"] = *int32Ptr

		int32Ptr, ok = jobs[i].GetTasksOk()
		if !ok {
			continue
		}
		records["tasks"] = *int32Ptr

		int64Ptr, ok = jobs[i].GetTimeLimitOk()
		if !ok {
			continue
		}
		records["time_limit"] = *int64Ptr

		strPtr, ok = jobs[i].GetTresReqStrOk()
		if !ok {
			continue
		}
		records["tres_req_str"] = *strPtr

		acc.AddFields("slurm_jobs", records, tags)
	}
}

func (s *Slurm) gatherNodesMetrics(acc telegraf.Accumulator,
	nodes []goslurm.V0038Node) {
	var (
		int32Ptr *int32
		int64Ptr *int64
		strPtr   *string
		ok       bool
	)

	for _, node := range nodes {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["source"] = s.baseURL.Hostname()
		strPtr, ok = node.GetNameOk()
		if !ok {
			continue
		}
		tags["name"] = *strPtr

		strPtr, ok = node.GetStateOk()
		if !ok {
			continue
		}
		records["state"] = *strPtr

		int32Ptr, ok = node.GetCoresOk()
		if !ok {
			continue
		}
		records["cores"] = *int32Ptr

		int32Ptr, ok = node.GetCpusOk()
		if !ok {
			continue
		}
		records["cpus"] = *int32Ptr

		int64Ptr, ok = node.GetCpuLoadOk()
		if !ok {
			continue
		}
		records["cpu_load"] = *int64Ptr

		int64Ptr, ok = node.GetAllocCpusOk()
		if !ok {
			continue
		}
		records["alloc_cpu"] = *int64Ptr

		int32Ptr, ok = node.GetRealMemoryOk()
		if !ok {
			continue
		}
		records["real_memory"] = *int32Ptr

		int32Ptr, ok = node.GetFreeMemoryOk()
		if !ok {
			continue
		}
		records["free_memory"] = *int32Ptr

		int64Ptr, ok = node.GetAllocMemoryOk()
		if !ok {
			continue
		}
		records["alloc_memory"] = *int64Ptr

		strPtr, ok = node.GetTresOk()
		if !ok {
			continue
		}
		records["tres"] = *strPtr

		strPtr, ok = node.GetTresUsedOk()
		if !ok {
			continue
		}
		records["tres_used"] = *strPtr

		int32Ptr, ok = node.GetWeightOk()
		if !ok {
			continue
		}
		records["weight"] = *int32Ptr

		strPtr, ok = node.GetSlurmdVersionOk()
		if !ok {
			continue
		}
		records["slurmd_version"] = *strPtr

		strPtr, ok = node.GetArchitectureOk()
		if !ok {
			continue
		}
		records["architecture"] = *strPtr

		acc.AddFields("slurm_nodes", records, tags)
	}
}

func (s *Slurm) gatherPartitionsMetrics(acc telegraf.Accumulator,
	partitions []goslurm.V0038Partition) {
	var (
		int32Ptr *int32
		strPtr   *string
		ok       bool
	)

	for _, partition := range partitions {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["source"] = s.baseURL.Hostname()
		strPtr, ok = partition.GetNameOk()
		if !ok {
			continue
		}
		tags["name"] = *strPtr

		strPtr, ok = partition.GetStateOk()
		if !ok {
			continue
		}
		records["state"] = *strPtr

		int32Ptr, ok = partition.GetTotalCpusOk()
		if !ok {
			continue
		}
		records["total_cpu"] = *int32Ptr

		int32Ptr, ok = partition.GetTotalNodesOk()
		if !ok {
			continue
		}
		records["total_nodes"] = *int32Ptr

		strPtr, ok = partition.GetNodesOk()
		if !ok {
			continue
		}
		records["nodes"] = *strPtr

		strPtr, ok = partition.GetTresOk()
		if !ok {
			continue
		}
		records["tres"] = *strPtr

		acc.AddFields("slurm_partitions", records, tags)
	}
}

func (s *Slurm) gatherReservationsMetrics(acc telegraf.Accumulator,
	reservations []goslurm.V0038Reservation) {
	var (
		int32Ptr *int32
		strPtr   *string
		ok       bool
	)
	for _, reservation := range reservations {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["source"] = s.baseURL.Hostname()
		strPtr, ok = reservation.GetNameOk()
		if !ok {
			continue
		}
		tags["name"] = *strPtr

		int32Ptr, ok = reservation.GetCoreCountOk()
		if !ok {
			continue
		}
		records["core_count"] = *int32Ptr

		int32Ptr, ok = reservation.GetCoreSpecCntOk()
		if !ok {
			continue
		}
		records["core_spec_count"] = *int32Ptr

		strPtr, ok = reservation.GetGroupsOk()
		if !ok {
			continue
		}
		records["groups"] = *strPtr

		strPtr, ok = reservation.GetUsersOk()
		if !ok {
			continue
		}
		records["users"] = *strPtr

		int32Ptr, ok = reservation.GetStartTimeOk()
		if !ok {
			continue
		}
		records["start_time"] = *int32Ptr

		strPtr, ok = reservation.GetPartitionOk()
		if !ok {
			continue
		}
		records["partition"] = *strPtr

		strPtr, ok = reservation.GetAccountsOk()
		if !ok {
			continue
		}
		records["accounts"] = *strPtr

		int32Ptr, ok = reservation.GetNodeCountOk()
		if !ok {
			continue
		}
		records["node_count"] = *int32Ptr

		strPtr, ok = reservation.GetNodeListOk()
		if !ok {
			continue
		}
		records["node_list"] = *strPtr

		acc.AddFields("slurm_reservations", records, tags)
	}
}

func (s *Slurm) Gather(acc telegraf.Accumulator) (err error) {
	auth := context.WithValue(
		context.Background(),
		goslurm.ContextAPIKeys,
		map[string]goslurm.APIKey{
			"user":  {Key: s.Username},
			"token": {Key: s.Token},
		},
	)

	if !s.endpointMap["diag"] {
		diagResp, respRaw, err := s.client.SlurmAPI.SlurmV0038Diag(auth).Execute()
		if err != nil {
			return fmt.Errorf("error getting diag: %w", err)
		}
		defer respRaw.Body.Close()
		if diag, ok := diagResp.GetStatisticsOk(); ok {
			s.gatherDiagMetrics(acc, diag)
		}
	}

	if !s.endpointMap["jobs"] {
		jobsResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetJobs(auth).Execute()
		if err != nil {
			return fmt.Errorf("error getting jobs: %w", err)
		}
		defer respRaw.Body.Close()
		if jobs, ok := jobsResp.GetJobsOk(); ok {
			s.gatherJobsMetrics(acc, jobs)
		}
	}

	if !s.endpointMap["nodes"] {
		nodesResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetNodes(auth).Execute()
		if err != nil {
			return fmt.Errorf("error getting nodes: %w", err)
		}
		defer respRaw.Body.Close()
		if nodes, ok := nodesResp.GetNodesOk(); ok {
			s.gatherNodesMetrics(acc, nodes)
		}
	}

	if !s.endpointMap["partitions"] {
		partitionsResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetPartitions(
			auth).Execute()
		if err != nil {
			return fmt.Errorf("error getting partitions: %w", err)
		}
		defer respRaw.Body.Close()
		if partitions, ok := partitionsResp.GetPartitionsOk(); ok {
			s.gatherPartitionsMetrics(acc, partitions)
		}
	}

	if !s.endpointMap["reservations"] {
		reservationsResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetReservations(
			auth).Execute()
		if err != nil {
			return fmt.Errorf("error getting reservations: %w", err)
		}
		defer respRaw.Body.Close()
		if reservations, ok := reservationsResp.GetReservationsOk(); ok {
			s.gatherReservationsMetrics(acc, reservations)
		}
	}

	return nil
}

func init() {
	inputs.Add("slurm", func() telegraf.Input { return &Slurm{} })
}
