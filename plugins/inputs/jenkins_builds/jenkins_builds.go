//go:generate ../../../tools/readme_config_includer/generator
package jenkins

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

//go:embed sample.conf
var sampleConfig string

type JenkinsBuilds struct {
	URL             string
	Username        string
	Password        string
	Source          string
	Port            string
	ResponseTimeout config.Duration

	tls.ClientConfig
	client *JenkinsClient

	Log telegraf.Logger

	MaxIdleConnections int      `toml:"max_idle_connections"`
	MaxWorkers         int      `toml:"max_workers"`
	MaxBuildAge        int      `toml:"max_build_age"`
	MaxNumBuilds       int      `toml:"max_num_builds"`
	JobExclude         []string `toml:"job_exclude"`
	JobInclude         []string `toml:"job_include"`
	jobFilter          filter.Filter
	semaphore          chan int
}

func (j *JenkinsBuilds) SampleConfig() string {
	return sampleConfig
}

const (
	measurementJob      = "jenkins_job_v2"
	measurementExecutor = "jenkins_executors_v2"
)

func (j *JenkinsBuilds) Gather(acc telegraf.Accumulator) error {
	if j.client == nil {
		client, err := j.newHTTPClient()
		if err != nil {
			return err
		}
		if err = j.initialize(client); err != nil {
			return err
		}
	}

	j.gatherExecutorInfo(acc)
	j.gatherJobs(acc)

	return nil
}

func (j *JenkinsBuilds) newHTTPClient() (*http.Client, error) {
	tlsCfg, err := j.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("error parse jenkins config %q: %w", j.URL, err)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			MaxIdleConns:    j.MaxIdleConnections,
		},
		Timeout: time.Duration(j.ResponseTimeout),
	}, nil
}

// separate the client as dependency to use httptest Client for mocking
func (j *JenkinsBuilds) initialize(client *http.Client) error {
	var err error

	// init jenkins tags
	u, err := url.Parse(j.URL)
	if err != nil {
		return err
	}
	if u.Port() == "" {
		if u.Scheme == "http" {
			j.Port = "80"
		} else if u.Scheme == "https" {
			j.Port = "443"
		}
	} else {
		j.Port = u.Port()
	}
	j.Source = u.Hostname()

	if j.MaxIdleConnections <= 0 {
		j.MaxIdleConnections = 10
	}

	if j.MaxWorkers <= 0 {
		j.MaxIdleConnections = 5
	}

	// init filters
	j.jobFilter, err = filter.NewIncludeExcludeFilter(j.JobInclude, j.JobExclude)
	if err != nil {
		return fmt.Errorf("error compiling job filters %q: %w", j.URL, err)
	}

	j.semaphore = make(chan int, j.MaxWorkers)
	ctx := context.Background()
	j.client = newJenkinsClient(client, j.URL, j.Username, j.Password, ctx)

	return j.client.init()
}

func (j *JenkinsBuilds) gatherJobs(acc telegraf.Accumulator) {
	j.Log.Infof("Getting all jobs")
	start := time.Now().Unix()
	jobs, err := j.client.getAllJobs()

	if err != nil {
		acc.AddError(errors.New("unable to get all jobs : " + err.Error()))
		return
	}
	j.Log.Infof("Got %d jobs", len(jobs))
	var wg sync.WaitGroup

	for _, job := range jobs {
		if !j.jobFilter.Match(job.Base) {
			continue
		}
		wg.Add(1)
		j.semaphore <- 1
		go func(job JobInfo, acc telegraf.Accumulator) {
			j.processJobs(job, acc)
			<-j.semaphore
			wg.Done()
		}(job, acc)

	}
	wg.Wait()
	end := time.Now().Unix()
	j.Log.Infof("Finished Gathering jobs in %d", end-start)
}

func (j *JenkinsBuilds) processJobs(job JobInfo, acc telegraf.Accumulator) {
	thisJob, err := j.client.getJob(job.Name, job.Parents)
	if err != nil {
		acc.AddError(errors.New("unable to get job : " + err.Error()))
		return
	}

	builds, err := j.client.getAllBuilds(thisJob)

	if err != nil {
		acc.AddError(errors.New("unable to get all build ids : " + err.Error()))
		return
	}

	if builds == nil {
		return
	}

	if len(builds) > j.MaxNumBuilds {
		builds = builds[0:j.MaxNumBuilds]
	}

	cutoff := time.Now().Add(-time.Hour * time.Duration(j.MaxBuildAge))
	for _, build := range builds {
		buildInfo, err := j.client.getBuildInfo(thisJob, build.Number)
		if err != nil {
			continue
		}

		if buildInfo == nil || buildInfo.GetTimestamp().Before(cutoff) || buildInfo.Raw.Building {
			continue
		}

		j.gatherJobBuild(job, buildInfo, acc)
	}
}

func (j *JenkinsBuilds) gatherJobBuild(job JobInfo, buildInfo *gojenkins.Build, acc telegraf.Accumulator) {
	jobParent := strings.Join(job.Parents, "/")
	tags := map[string]string{"name": job.Name, "parents": jobParent, "result": buildInfo.GetResult(), "server": j.client.getServer()}
	fields := make(map[string]interface{})
	fields["duration"] = buildInfo.GetDuration()
	fields["result_code"] = mapResultCode(buildInfo.GetResult())
	fields["number"] = buildInfo.GetBuildNumber()
	fields["estimated_duration"] = buildInfo.Raw.EstimatedDuration

	acc.AddFields(measurementJob, fields, tags, buildInfo.GetTimestamp())
}

func (j *JenkinsBuilds) gatherExecutorInfo(acc telegraf.Accumulator) {
	total, busy, err := j.client.getExecutors()
	if err != nil {
		acc.AddError(errors.New("unable to get executor info : " + err.Error()))
		return
	}
	tags := map[string]string{"server": j.client.getServer()}
	fields := make(map[string]interface{})
	fields["total_executors"] = total
	fields["busy_executors"] = busy

	acc.AddFields(measurementExecutor, fields, tags, time.Now())
}

func mapResultCode(s string) int {
	switch strings.ToLower(s) {
	case "success":
		return 0
	case "failure":
		return 1
	case "not_built":
		return 2
	case "unstable":
		return 3
	case "aborted":
		return 4
	}
	return -1
}

func init() {
	inputs.Add("jenkins_builds", func() telegraf.Input {
		return &JenkinsBuilds{
			MaxBuildAge:        1,
			MaxIdleConnections: 10,
			MaxNumBuilds:       30,
			MaxWorkers:         5,
		}
	})
}
