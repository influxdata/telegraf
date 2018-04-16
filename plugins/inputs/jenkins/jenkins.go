package jenkins

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Jenkins struct {
	URL    string
	User   string
	Passwd string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool
	// HTTP Timeout specified as a string - 3s, 1m, 1h
	ResponseTimeout internal.Duration
	Instance        *gojenkins.Jenkins

	LastbuildFilterInterval internal.Duration `toml:"lastbuild_interval"`
	JobFilterName           []string          `toml:"job_exclude"`
}

type byBuildNumber []gojenkins.JobBuild

func (a byBuildNumber) Len() int           { return len(a) }
func (a byBuildNumber) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byBuildNumber) Less(i, j int) bool { return a[i].Number > a[j].Number }

const sampleConfig = `
  url = "http://my-jenkins-instance:8080"
  user = "admin"
  passwd = "admin"
  ## Set response_timeout
  response_timeout = "5s"

  ## Optional SSL Config
  # ssl_ca = /path/to/cafile
  # ssl_cert = /path/to/certfile
  # ssl_key = /path/to/keyfile
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Job & build filter
  lastbuild_interval = "1h"
  # job_exclude = [ "MyJob", "MyOtherJob" ]
`

func (j *Jenkins) SampleConfig() string {
	return sampleConfig
}

func (j *Jenkins) Description() string {
	return "Read jobs and cluster metrics from Jenkins instances"
}

func (j *Jenkins) Gather(acc telegraf.Accumulator) error {

	tlsCfg, err := internal.GetTLSConfig(j.SSLCert, j.SSLKey, j.SSLCA, j.InsecureSkipVerify)
	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	httpclient := &http.Client{
		Transport: tr,
		Timeout:   j.ResponseTimeout.Duration,
	}

	instance, err := gojenkins.CreateJenkins(httpclient, j.URL, j.User, j.Passwd).Init()
	if err != nil {
		return fmt.Errorf("E! It was not possible to connect to the Jenkins instance\n")
	}

	j.Instance = instance

	nodes, err := instance.GetAllNodes()
	if err != nil {
		return fmt.Errorf("E! Something went wrong retrieving nodes information from Jenkins\n")
	}

	jobs, err := instance.GetAllJobs()
	if err != nil {
		return fmt.Errorf("E! Something went wrong retrieving jobs information from Jenkins\n")
	}

	acc.AddError(j.GetNodesData(nodes, acc))
	acc.AddError(j.GetJobsData(jobs, acc))

	return nil
}

func (j *Jenkins) GetNodesData(nodes []*gojenkins.Node, acc telegraf.Accumulator) error {

	measurement := "jenkins_node"

	tags := map[string]string{}
	fields := make(map[string]interface{})

	// get node data
	for _, node := range nodes {
		info, _ := node.Info()
		tags["node_name"] = node.GetName()

		// skip nodes without information
		if info.MonitorData.Hudson_NodeMonitors_ArchitectureMonitor == nil {
			continue
		}

		tags["arch"] = info.MonitorData.Hudson_NodeMonitors_ArchitectureMonitor.(string)
		isOnline, _ := node.IsOnline()
		fields["online"] = int(0)
		if isOnline {
			fields["online"] = int(1)
		}

		fields["response_time"] = info.MonitorData.Hudson_NodeMonitors_ResponseTimeMonitor.Average

		if info.MonitorData.Hudson_NodeMonitors_DiskSpaceMonitor != nil {
			diskSpace := info.MonitorData.Hudson_NodeMonitors_DiskSpaceMonitor.(map[string]interface{})
			tags["disk_path"] = diskSpace["path"].(string)
			fields["disk_available"] = diskSpace["size"].(float64)
		}

		if info.MonitorData.Hudson_NodeMonitors_TemporarySpaceMonitor != nil {
			tempSpace := info.MonitorData.Hudson_NodeMonitors_TemporarySpaceMonitor.(map[string]interface{})
			tags["temp_path"] = tempSpace["path"].(string)
			fields["temp_available"] = tempSpace["size"].(float64)
		}

		if info.MonitorData.Hudson_NodeMonitors_SwapSpaceMonitor != nil {
			swapSpace := info.MonitorData.Hudson_NodeMonitors_SwapSpaceMonitor.(map[string]interface{})
			fields["swap_available"] = swapSpace["availableSwapSpace"].(float64)
			fields["swap_total"] = swapSpace["totalSwapSpace"].(float64)
			fields["memory_available"] = swapSpace["availablePhysicalMemory"].(float64)
			fields["memory_total"] = swapSpace["totalPhysicalMemory"].(float64)
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
			log.Printf("E! Error retrieving last build from job %s: %s", jobName, err.Error())
			return err
		}

		// ignore if last build is too old
		if (j.LastbuildFilterInterval != internal.Duration{Duration: 0}) {
			buildSecAgo := time.Now().Sub(jobLastBuild.GetTimestamp()).Seconds()

			if buildSecAgo > j.LastbuildFilterInterval.Duration.Seconds() {
				log.Printf("D! Last job too old, last %s build was %v seconds ago", jobName, buildSecAgo)
				continue
			}
		}

		buildIds, err := job.GetAllBuildIds()
		sort.Sort(byBuildNumber(buildIds))

		if err != nil {
			log.Printf("E! Error retrieving all builds from job \"%s\": %s", jobName, err.Error())
			return err
		}

		for _, buildId := range buildIds {
			build, err := job.GetBuild(buildId.Number)
			log.Printf("D! Reading data from job %s build %v", jobName, buildId.Number)
			if err != nil {
				log.Printf("E! Error retrieving build from job \"%s\": %s", jobName, err.Error())
				return err
			}

			// ignore if build is ongoing
			if build.IsRunning() {
				log.Printf("D! Ignore running build on %s, build %v", jobName, buildId.Number)
				continue
			}

			// stop if build is too old
			if (j.LastbuildFilterInterval != internal.Duration{Duration: 0}) {
				buildSecAgo := time.Now().Sub(build.GetTimestamp()).Seconds()

				if time.Now().Sub(build.GetTimestamp()).Seconds() > j.LastbuildFilterInterval.Duration.Seconds() {
					log.Printf("D! Job %s build %v too old (%v seconds ago), skipping to next job", jobName, buildId.Number, buildSecAgo)
					break
				}
			}

			tags := map[string]string{"job_name": jobName, "result": build.GetResult()}
			fields := make(map[string]interface{})
			fields["duration"] = build.GetDuration()
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
