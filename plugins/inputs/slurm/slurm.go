//go:generate ../../../tools/readme_config_includer/generator
package slurm

import (
	"context"
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

func (s *Slurm) createHTTPClient(host string) *goslurm.APIClient {
	configuration := goslurm.NewConfiguration()
	configuration.Host = host
	configuration.Scheme = "http"
	configuration.UserAgent = "Telegraf Metrics Agent"
	configuration.HTTPClient = &http.Client{
		Timeout: time.Duration(s.ResponseTimeout),
	}

	return goslurm.NewAPIClient(configuration)
}

func (s *Slurm) createHTTPSClient(host string) (*goslurm.APIClient, error) {
	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	configuration := goslurm.NewConfiguration()
	configuration.Host = host
	configuration.Scheme = "https"
	configuration.UserAgent = "Telegraf Metrics Agent"
	configuration.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(s.ResponseTimeout),
	}

	return goslurm.NewAPIClient(configuration), nil
}

func (*Slurm) SampleConfig() string {
	return sampleConfig
}

func (s *Slurm) Init() error {
	if s.ResponseTimeout < config.Duration(time.Second) {
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

	switch u.Scheme {
	case "http":
		s.client = s.createHTTPClient(u.Host)
	case "https":
		s.client, err = s.createHTTPSClient(u.Host)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid scheme %q", u.Scheme)
	}

	return nil
}

func (s *Slurm) GatherDiagMetrics(acc telegraf.Accumulator,
	diag *goslurm.V0038DiagStatistics) {
	records := make(map[string]interface{})
	tags := make(map[string]string)

	tags["url"] = s.baseURL.Hostname()

	if tmp, ok := diag.GetServerThreadCountOk(); ok {
		records["server_thread_count"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsCanceledOk(); ok {
		records["jobs_canceled"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsSubmittedOk(); ok {
		records["jobs_submitted"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsStartedOk(); ok {
		records["jobs_started"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsCompletedOk(); ok {
		records["jobs_completed"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsFailedOk(); ok {
		records["jobs_failed"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsPendingOk(); ok {
		records["jobs_pending"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetJobsRunningOk(); ok {
		records["jobs_running"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetScheduleCycleLastOk(); ok {
		records["schedule_cycle_last"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetScheduleCycleMeanOk(); ok {
		records["schedule_cycle_mean"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetBfQueueLenOk(); ok {
		records["bf_queue_len"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetBfQueueLenMeanOk(); ok {
		records["bf_queue_len_mean"] = *tmp
	} else {
		return
	}
	if tmp, ok := diag.GetBfActiveOk(); ok {
		records["bf_active"] = *tmp
	} else {
		return
	}

	acc.AddFields("slurm_diag", records, tags)
}

func (s *Slurm) GatherJobsMetrics(acc telegraf.Accumulator,
	jobs []goslurm.V0038JobResponseProperties) {
	for i := range jobs {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["url"] = s.baseURL.Hostname()
		if tmp, ok := jobs[i].GetNameOk(); ok {
			tags["name"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetJobIdOk(); ok {
			tags["job_id"] = strconv.Itoa(int(*tmp))
		} else {
			continue
		}

		if tmp, ok := jobs[i].GetJobStateOk(); ok {
			records["state"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetStateReasonOk(); ok {
			records["state_reason"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetPartitionOk(); ok {
			records["partition"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetNodesOk(); ok {
			records["nodes"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetNodeCountOk(); ok {
			records["node_count"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetPriorityOk(); ok {
			records["priority"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetNiceOk(); ok {
			records["nice"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetGroupIdOk(); ok {
			records["group_id"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetCommandOk(); ok {
			records["command"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetStandardOutputOk(); ok {
			records["standard_output"] = strings.ReplaceAll(*tmp, "\\", "")
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetStandardErrorOk(); ok {
			records["standard_error"] = strings.ReplaceAll(*tmp, "\\", "")
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetStandardInputOk(); ok {
			records["standard_input"] = strings.ReplaceAll(*tmp, "\\", "")
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetCurrentWorkingDirectoryOk(); ok {
			records["current_working_directory"] = strings.ReplaceAll(
				*tmp, "\\", "")
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetSubmitTimeOk(); ok {
			records["submit_time"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetStartTimeOk(); ok {
			records["start_time"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetCpusOk(); ok {
			records["cpus"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetTasksOk(); ok {
			records["tasks"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetTimeLimitOk(); ok {
			records["time_limit"] = *tmp
		} else {
			continue
		}
		if tmp, ok := jobs[i].GetTresReqStrOk(); ok {
			records["tres_req_str"] = *tmp
		} else {
			continue
		}

		acc.AddFields("slurm_jobs", records, tags)
	}
}

func (s *Slurm) GatherNodesMetrics(acc telegraf.Accumulator,
	nodes []goslurm.V0038Node) {
	for _, node := range nodes {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["url"] = s.baseURL.Hostname()
		if tmp, ok := node.GetNameOk(); ok {
			tags["name"] = *tmp
		} else {
			continue
		}

		if tmp, ok := node.GetStateOk(); ok {
			records["state"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetCoresOk(); ok {
			records["cores"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetCpusOk(); ok {
			records["cpus"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetCpuLoadOk(); ok {
			records["cpu_load"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetAllocCpusOk(); ok {
			records["alloc_cpu"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetRealMemoryOk(); ok {
			records["real_memory"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetFreeMemoryOk(); ok {
			records["free_memory"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetAllocMemoryOk(); ok {
			records["alloc_memory"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetTresOk(); ok {
			records["tres"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetTresUsedOk(); ok {
			records["tres_used"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetWeightOk(); ok {
			records["weight"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetSlurmdVersionOk(); ok {
			records["slurmd_version"] = *tmp
		} else {
			continue
		}
		if tmp, ok := node.GetArchitectureOk(); ok {
			records["architecture"] = *tmp
		} else {
			continue
		}

		acc.AddFields("slurm_nodes", records, tags)
	}
}

func (s *Slurm) GatherPartitionsMetrics(acc telegraf.Accumulator,
	partitions []goslurm.V0038Partition) {
	for _, partition := range partitions {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["url"] = s.baseURL.Hostname()
		if tmp, ok := partition.GetNameOk(); ok {
			tags["name"] = *tmp
		} else {
			continue
		}

		if tmp, ok := partition.GetStateOk(); ok {
			records["state"] = *tmp
		} else {
			continue
		}
		if tmp, ok := partition.GetTotalCpusOk(); ok {
			records["total_cpu"] = *tmp
		} else {
			continue
		}
		if tmp, ok := partition.GetTotalNodesOk(); ok {
			records["total_nodes"] = *tmp
		} else {
			continue
		}
		if tmp, ok := partition.GetNodesOk(); ok {
			records["nodes"] = *tmp
		} else {
			continue
		}
		if tmp, ok := partition.GetTresOk(); ok {
			records["tres"] = *tmp
		} else {
			continue
		}

		acc.AddFields("slurm_partitions", records, tags)
	}
}

func (s *Slurm) GatherReservationsMetrics(acc telegraf.Accumulator,
	reservations []goslurm.V0038Reservation) {
	for _, reservation := range reservations {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["url"] = s.baseURL.Hostname()
		if tmp, ok := reservation.GetNameOk(); ok {
			tags["name"] = *tmp
		} else {
			continue
		}

		if tmp, ok := reservation.GetCoreCountOk(); ok {
			records["core_count"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetCoreSpecCntOk(); ok {
			records["core_spec_count"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetGroupsOk(); ok {
			records["groups"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetUsersOk(); ok {
			records["users"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetStartTimeOk(); ok {
			records["start_time"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetPartitionOk(); ok {
			records["partition"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetAccountsOk(); ok {
			records["accounts"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetNodeCountOk(); ok {
			records["node_count"] = *tmp
		} else {
			continue
		}
		if tmp, ok := reservation.GetNodeListOk(); ok {
			records["node_list"] = *tmp
		} else {
			continue
		}

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
			return err
		}
		defer respRaw.Body.Close()
		diag, ok := diagResp.GetStatisticsOk()
		if !ok {
			return fmt.Errorf("error getting diag: %w", err)
		}
		s.GatherDiagMetrics(acc, diag)
	}

	if !s.endpointMap["jobs"] {
		jobsResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetJobs(auth).Execute()
		if err != nil {
			return err
		}
		defer respRaw.Body.Close()
		jobs, ok := jobsResp.GetJobsOk()
		if !ok {
			return fmt.Errorf("error getting jobs: %w", err)
		}
		s.GatherJobsMetrics(acc, jobs)
	}

	if !s.endpointMap["nodes"] {
		nodesResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetNodes(auth).Execute()
		if err != nil {
			return err
		}
		defer respRaw.Body.Close()
		nodes, ok := nodesResp.GetNodesOk()
		if !ok {
			return fmt.Errorf("error getting nodes: %w", err)
		}
		s.GatherNodesMetrics(acc, nodes)
	}

	if !s.endpointMap["partitions"] {
		partitionsResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetPartitions(
			auth).Execute()
		if err != nil {
			return err
		}
		defer respRaw.Body.Close()
		partitions, ok := partitionsResp.GetPartitionsOk()
		if !ok {
			return fmt.Errorf("error getting partitions: %w", err)
		}
		s.GatherPartitionsMetrics(acc, partitions)
	}

	if !s.endpointMap["reservations"] {
		reservationsResp, respRaw, err := s.client.SlurmAPI.SlurmV0038GetReservations(
			auth).Execute()
		if err != nil {
			return err
		}
		defer respRaw.Body.Close()
		reservations, ok := reservationsResp.GetReservationsOk()
		if !ok {
			return fmt.Errorf("error getting reservations: %w", err)
		}
		s.GatherReservationsMetrics(acc, reservations)
	}

	return nil
}

func init() {
	inputs.Add("slurm", func() telegraf.Input { return &Slurm{} })
}
