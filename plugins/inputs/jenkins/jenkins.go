package jenkins

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Jenkins struct {
	URL      string
	Username string
	Password string
	tls.ClientConfig
	ResponseTimeout internal.Duration
	Instance        *gojenkins.Jenkins
	client          *http.Client

	MaxBuildAge   internal.Duration `toml:"max_build_age"`
	JobFilterName []string          `toml:"job_exclude"`
}

const sampleConfig = `
  url = "http://my-jenkins-instance:8080"
  username = "admin"
  password = "admin"
  ## Set response_timeout
  response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = /path/to/cafile
  # tls_cert = /path/to/certfile
  # tls_key = /path/to/keyfile
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Job & build filter
  max_build_age = "1h"
  # job_exclude = [ "MyJob", "MyOtherJob" ]
`

func (j *Jenkins) SampleConfig() string {
	return sampleConfig
}

func (j *Jenkins) Description() string {
	return "Read jobs and cluster metrics from Jenkins instances"
}

func (j *Jenkins) Gather(acc telegraf.Accumulator) error {
	if j.client == nil {
		client, err := j.createHttpClient()
		if err != nil {
			return err
		}
		j.client = client
	}

	instance, err := gojenkins.CreateJenkins(j.client, j.URL, j.Username, j.Password).Init()
	if err != nil {
		return fmt.Errorf("error retrieving connecting to jenkins instance[%s]: %v", j.URL, err)
	}

	j.Instance = instance

	nodes, err := instance.GetAllNodes()
	if err != nil {
		return fmt.Errorf("error retrieving nodes[%s]: %v", j.URL, err)
	}

	jobs, err := instance.GetAllJobs()
	if err != nil {
		return fmt.Errorf("error retrieving jobs[%s]: %v", j.URL, err)
	}

	acc.AddError(j.GetNodesData(nodes, acc))
	acc.AddError(j.GetJobsData(jobs, acc))

	return nil
}

func (j *Jenkins) createHttpClient() (*http.Client, error) {
	tlsCfg, err := j.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	if j.ResponseTimeout.Duration < time.Second {
		j.ResponseTimeout.Duration = time.Second * 5
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: j.ResponseTimeout.Duration,
	}

	return client, nil
}

func (j *Jenkins) GetNodesData(nodes []*gojenkins.Node, acc telegraf.Accumulator) error {

	measurement := "jenkins_node"

	tags := map[string]string{}
	fields := make(map[string]interface{})

	// get node data
	for _, node := range nodes {
		info, err := node.Info()
		if err != nil {
			return fmt.Errorf("error retrieving node information: %v", err)
		}
		tags["node_name"] = node.GetName()

		// skip nodes without information
		if info.MonitorData.Hudson_NodeMonitors_ArchitectureMonitor == nil {
			continue
		}

		tags["arch"] = info.MonitorData.Hudson_NodeMonitors_ArchitectureMonitor.(string)

		isOnline, err := node.IsOnline()
		if err != nil {
			return fmt.Errorf("error retrieving isOnline data from node: %v", err)
		}

		tags["online"] = strconv.FormatBool(isOnline)
		fields["response_time"] = info.MonitorData.Hudson_NodeMonitors_ResponseTimeMonitor.Average

		if info.MonitorData.Hudson_NodeMonitors_DiskSpaceMonitor != nil {
			diskSpace := info.MonitorData.Hudson_NodeMonitors_DiskSpaceMonitor.(map[string]interface{})
			if diskPath, ok := diskSpace["path"].(string); ok {
				tags["disk_path"] = diskPath
			}
			if diskAvailable, ok := diskSpace["size"].(float64); ok {
				fields["disk_available"] = diskAvailable
			}
		}

		if info.MonitorData.Hudson_NodeMonitors_TemporarySpaceMonitor != nil {
			tempSpace := info.MonitorData.Hudson_NodeMonitors_TemporarySpaceMonitor.(map[string]interface{})
			if tempPath, ok := tempSpace["path"].(string); ok {
				tags["temp_path"] = tempPath
			}
			if tempAvailable, ok := tempSpace["size"].(float64); ok {
				fields["temp_available"] = tempAvailable
			}
		}

		if info.MonitorData.Hudson_NodeMonitors_SwapSpaceMonitor != nil {
			swapSpace := info.MonitorData.Hudson_NodeMonitors_SwapSpaceMonitor.(map[string]interface{})
			if swapAvailable, ok := swapSpace["availableSwapSpace"].(float64); ok {
				fields["swap_available"] = swapAvailable
			}
			if swapTotal, ok := swapSpace["totalSwapSpace"].(float64); ok {
				fields["swap_total"] = swapTotal
			}
			if memoryAvailable, ok := swapSpace["availablePhysicalMemory"].(float64); ok {
				fields["memory_available"] = memoryAvailable
			}
			if memoryTotal, ok := swapSpace["totalPhysicalMemory"].(float64); ok {
				fields["memory_total"] = memoryTotal
			}
		}

		acc.AddFields(measurement, fields, tags)

	}

	return nil
}

func (j *Jenkins) GetJobsData(jobs []*gojenkins.Job, acc telegraf.Accumulator) error {

	measurement := "jenkins_job"

	for _, job := range jobs {

		jobName := job.GetName()

		for _, filtername := range j.JobFilterName {
			if filtername == jobName {
				continue
			}
		}

		// ignore if job has no builds
		if job.Raw.LastBuild.Number < 1 {
			continue
		}

		jobLastBuild, err := job.GetLastBuild()
		if err != nil {
			return fmt.Errorf("error retrieving last build from job [%s]: %v", jobName, err)
		}

		// ignore if last build is too old
		if (j.MaxBuildAge != internal.Duration{Duration: 0}) {
			buildSecAgo := time.Now().Sub(jobLastBuild.GetTimestamp()).Seconds()

			if buildSecAgo > j.MaxBuildAge.Duration.Seconds() {
				log.Printf("D! Last job too old, last %s build was %v seconds ago", jobName, buildSecAgo)
				continue
			}
		}

		buildIds, err := job.GetAllBuildIds()
		if err != nil {
			return fmt.Errorf("error retrieving all builds from job [%s]: %v", jobName, err)
		}

		sort.Slice(buildIds, func(i, j int) bool {
			return buildIds[i].Number > buildIds[j].Number
		})

		for _, buildId := range buildIds {
			build, err := job.GetBuild(buildId.Number)
			if err != nil {
				return fmt.Errorf("error retrieving build from job [%s]: %v", jobName, err)
			}

			// ignore if build is ongoing
			if build.IsRunning() {
				log.Printf("D! Ignore running build on %s, build %v", jobName, buildId.Number)
				continue
			}

			// stop if build is too old
			if (j.MaxBuildAge != internal.Duration{Duration: 0}) {
				buildSecAgo := time.Now().Sub(build.GetTimestamp()).Seconds()

				if time.Now().Sub(build.GetTimestamp()).Seconds() > j.MaxBuildAge.Duration.Seconds() {
					log.Printf("D! Job %s build %v too old (%v seconds ago), skipping to next job", jobName, buildId.Number, buildSecAgo)
					break
				}
			}

			tags := map[string]string{"job_name": jobName, "result": build.GetResult()}
			fields := make(map[string]interface{})
			fields["duration_ms"] = build.GetDuration()
			fields["result_code"] = mapResultCode(build.GetResult())

			acc.AddFields(measurement, fields, tags, build.GetTimestamp())

		}
	}

	return nil

}

// perform status mapping
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
	inputs.Add("jenkins", func() telegraf.Input {
		return &Jenkins{}
	})
}
