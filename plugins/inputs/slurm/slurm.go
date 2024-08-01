//go:generate ../../../tools/readme_config_includer/generator
package slurm

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
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
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Token    string `toml:"token"`

	IgnoreDiag         bool `toml:"ignoreDiag"`
	IgnoreJobs         bool `toml:"ignoreJobs"`
	IgnoreNodes        bool `toml:"ignoreNodes"`
	IgnorePartitions   bool `toml:"ignorePartitions"`
	IgnoreReservations bool `toml:"ignoreReservations"`

	ResponseTimeout config.Duration `toml:"response_timeout"`
	tls.ClientConfig

	client  *goslurm.APIClient
	baseURL *url.URL
}

func (s *Slurm) createHTTPClient(host string) (*goslurm.APIClient, error) {
	configuration := goslurm.NewConfiguration()
	configuration.Host = host
	configuration.Scheme = "http"
	configuration.UserAgent = "Telegraf Metrics Agent"
	configuration.HTTPClient = &http.Client{
		Timeout: time.Duration(s.ResponseTimeout),
	}

	return goslurm.NewAPIClient(configuration), nil
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

	if s.URL == "" {
		return fmt.Errorf("empty URL provided")
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
		s.client, err = s.createHTTPClient(u.Host)
		if err != nil {
			return err
		}
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

	records["server_thread_count"] = diag.ServerThreadCount
	records["jobs_canceled"] = diag.JobsCanceled
	records["jobs_submitted"] = diag.JobsSubmitted
	records["jobs_started"] = diag.JobsStarted
	records["jobs_completed"] = diag.JobsCompleted
	records["jobs_failed"] = diag.JobsFailed
	records["jobs_pending"] = diag.JobsPending
	records["jobs_running"] = diag.JobsRunning
	records["schedule_cycle_last"] = diag.ScheduleCycleLast
	records["schedule_cycle_mean"] = diag.ScheduleCycleMean
	records["bf_queue_len"] = diag.BfQueueLen
	records["bf_queue_len_mean"] = diag.BfQueueLenMean
	records["bf_active"] = diag.BfActive

	acc.AddFields("slurm_diag", records, tags)
}

func (s *Slurm) GatherJobsMetrics(acc telegraf.Accumulator,
	jobs []goslurm.V0038JobResponseProperties) {
	for _, job := range jobs {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = *job.Name
		tags["job_id"] = fmt.Sprintf("%d", *job.JobId)

		records["state"] = job.JobState
		records["state_reason"] = job.StateReason
		records["partition"] = job.Partition
		records["nodes"] = job.Nodes
		records["node_count"] = job.NodeCount
		records["priority"] = job.Priority
		records["nice"] = *job.Nice
		records["group_id"] = job.GroupId
		records["command"] = job.Command
		records["standard_output"] = strings.ReplaceAll(
			*job.StandardOutput, "\\", "")
		records["standard_error"] = strings.ReplaceAll(
			*job.StandardError, "\\", "")
		records["standard_input"] = strings.ReplaceAll(
			*job.StandardInput, "\\", "")
		records["current_working_directory"] = strings.ReplaceAll(
			*job.CurrentWorkingDirectory, "\\", "")
		records["submit_time"] = job.SubmitTime
		records["start_time"] = job.StartTime
		records["cpus"] = job.Cpus
		records["cpus_per_task"] = job.CpusPerTask
		records["tasks"] = job.Tasks
		records["time_limit"] = job.TimeLimit
		records["tres_req_str"] = job.TresReqStr

		acc.AddFields("slurm_jobs", records, tags)
	}
}

func (s *Slurm) GatherNodesMetrics(acc telegraf.Accumulator,
	nodes []goslurm.V0038Node) {
	for _, node := range nodes {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = *node.Name

		records["state"] = node.State
		records["cores"] = node.Cores
		records["cpus"] = node.Cpus
		records["cpu_load"] = node.CpuLoad
		records["alloc_cpu"] = node.AllocCpus
		records["real_memory"] = node.RealMemory
		records["free_memory"] = node.FreeMemory
		records["alloc_memory"] = node.AllocMemory
		records["tres"] = node.Tres
		records["tres_used"] = node.TresUsed
		records["weight"] = node.Weight
		records["slurmd_version"] = node.SlurmdVersion
		records["partitions"] = node.Partitions
		records["architecture"] = node.Architecture

		acc.AddFields("slurm_nodes", records, tags)
	}
}

func (s *Slurm) GatherPartitionsMetrics(acc telegraf.Accumulator,
	partitions []goslurm.V0038Partition) {
	for _, partition := range partitions {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = *partition.Name

		records["state"] = partition.State
		records["total_cpu"] = partition.TotalCpus
		records["total_nodes"] = partition.TotalNodes
		records["nodes"] = partition.Nodes
		records["tres"] = partition.Tres

		acc.AddFields("slurm_partitions", records, tags)
	}
}

func (s *Slurm) GatherReservationsMetrics(acc telegraf.Accumulator,
	reservations []goslurm.V0038Reservation) {
	for _, reservation := range reservations {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = *reservation.Name

		records["core_count"] = reservation.CoreCount
		records["core_spec_count"] = reservation.CoreSpecCnt
		records["groups"] = reservation.Groups
		records["users"] = reservation.Users
		records["start_time"] = reservation.StartTime
		records["partition"] = reservation.Partition
		records["accounts"] = reservation.Accounts
		records["node_count"] = reservation.NodeCount
		records["node_list"] = reservation.NodeList
		records["core_count"] = reservation.CoreCount

		acc.AddFields("slurm_reservations", records, tags)
	}
}

func (s *Slurm) Gather(acc telegraf.Accumulator) error {
	auth := context.WithValue(
		context.Background(),
		goslurm.ContextAPIKeys,
		map[string]goslurm.APIKey{
			"user":  {Key: s.Username},
			"token": {Key: s.Token},
		},
	)

	if !s.IgnoreDiag {
		diagResp, _, err := s.client.SlurmAPI.SlurmV0038Diag(auth).Execute()
		if err != nil {
			return err
		}
		diag, ok := diagResp.GetStatisticsOk()
		if !ok {
			return fmt.Errorf("error getting diag: %v", err)
		}
		s.GatherDiagMetrics(acc, diag)
	}

	if !s.IgnoreJobs {
		jobsResp, _, err := s.client.SlurmAPI.SlurmV0038GetJobs(auth).Execute()
		if err != nil {
			return err
		}
		jobs, ok := jobsResp.GetJobsOk()
		if !ok {
			return fmt.Errorf("error getting jobs: %v", err)
		}
		s.GatherJobsMetrics(acc, jobs)
	}

	if !s.IgnoreNodes {
		nodesResp, _, err := s.client.SlurmAPI.SlurmV0038GetNodes(auth).Execute()
		if err != nil {
			return err
		}
		nodes, ok := nodesResp.GetNodesOk()
		if !ok {
			return fmt.Errorf("error getting nodes: %v", err)
		}
		s.GatherNodesMetrics(acc, nodes)
	}

	if !s.IgnorePartitions {
		partitionsResp, _, err := s.client.SlurmAPI.SlurmV0038GetPartitions(
			auth).Execute()
		if err != nil {
			return err
		}
		partitions, ok := partitionsResp.GetPartitionsOk()
		if !ok {
			return fmt.Errorf("error getting partitions: %v", err)
		}
		s.GatherPartitionsMetrics(acc, partitions)
	}

	if !s.IgnoreReservations {
		reservationsResp, _, err := s.client.SlurmAPI.SlurmV0038GetReservations(
			auth).Execute()
		if err != nil {
			return err
		}
		reservations, ok := reservationsResp.GetReservationsOk()
		if !ok {
			return fmt.Errorf("error getting reservations: %v", err)
		}
		s.GatherReservationsMetrics(acc, reservations)
	}

	return nil
}

func init() {
	inputs.Add("slurm", func() telegraf.Input { return &Slurm{} })
}
