package jenkins

import (
	"context"
	"github.com/bndr/gojenkins"
	"net/http"
	"net/url"
	"strings"
)

type JenkinsClient struct {
	goJenkinsClient *gojenkins.Jenkins
	httpClient      *http.Client
	ctx             context.Context
	url             string
	username        string
	password        string
}

func newJenkinsClient(client *http.Client, url string, username string, password string,
	ctx context.Context) *JenkinsClient {
	return &JenkinsClient{
		httpClient: client,
		ctx:        ctx,
		url:        url,
		username:   username,
		password:   password,
	}
}

func (jc *JenkinsClient) init() error {
	jenkinsClient, err := gojenkins.CreateJenkins(jc.httpClient, jc.url, jc.username, jc.password).Init(jc.ctx)
	if err != nil {
		return err
	}
	jc.goJenkinsClient = jenkinsClient
	return nil
}

func (jc *JenkinsClient) getJobs() []*gojenkins.Job {
	jobs, err := jc.goJenkinsClient.GetAllJobs(jc.ctx)
	if err != nil {
		panic(err)
	}
	return jobs
}

func (jc *JenkinsClient) isFolder(job gojenkins.InnerJob) bool {
	if strings.Contains(job.Class, "com.cloudbees.hudson.plugins.folder.Folder") {
		return true
	}
	return false
}

func (jc *JenkinsClient) isMultiBranch(job gojenkins.InnerJob) bool {
	if strings.Contains(job.Class, "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject") {
		return true
	}
	return false
}

func (jc *JenkinsClient) getAllJobs() ([]JobInfo, error) {
	var allJobs []JobInfo
	jobs, err := jc.goJenkinsClient.GetAllJobNames(jc.ctx)
	if err != nil {
		return nil, err
	}
	for _, w := range jobs {
		allJobs = jc._findJobs(allJobs, w)
	}
	return allJobs, nil
}

func getFolderPath(jobUrl string) (string, []string, string) {
	_jobUrl, _ := url.Parse(jobUrl)
	jobPath := _jobUrl.Path
	replaced := strings.ReplaceAll(jobPath, "/job/", "/")
	replaced = strings.TrimSuffix(replaced, "/")
	replaced = strings.TrimPrefix(replaced, "/")
	folders := strings.Split(replaced, "/")
	return folders[len(folders)-1], folders[0 : len(folders)-1], jobPath
}

func (jc *JenkinsClient) _findJobs(allJobs []JobInfo, job gojenkins.InnerJob) []JobInfo {
	if jc.isFolder(job) || jc.isMultiBranch(job) {
		folder, parents, _ := getFolderPath(job.Url)
		jobs, err := jc.goJenkinsClient.GetFolder(jc.ctx, folder, parents...)
		if err != nil {
			return allJobs
		}
		for _, w := range jobs.Raw.Jobs {
			allJobs = jc._findJobs(allJobs, w)
		}
	} else {
		folder, parents, base := getFolderPath(job.Url)
		allJobs = append(allJobs, JobInfo{
			Base:    base,
			Class:   job.Class,
			Name:    job.Name,
			Url:     job.Url,
			Color:   job.Color,
			Folder:  folder,
			Parents: parents,
		})
	}
	return allJobs
}

type JobInfo struct {
	Base    string
	Class   string
	Name    string
	Url     string
	Color   string
	Folder  string
	Parents []string
}

func (jc *JenkinsClient) getBuildInfo(job *gojenkins.Job, buildNumber int64) (*gojenkins.Build, error) {
	build, err := job.GetBuild(jc.ctx, buildNumber)
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (jc *JenkinsClient) getAllBuilds(job *gojenkins.Job) ([]gojenkins.JobBuild, error) {
	build, err := job.GetAllBuildIds(jc.ctx)
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (jc *JenkinsClient) getJob(name string, parents []string) (*gojenkins.Job, error) {
	job, err := jc.goJenkinsClient.GetJob(jc.ctx, name, parents...)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (jc *JenkinsClient) getExecutors() (total int, busy int, err error) {
	computers := new(gojenkins.Computers)
	qr := map[string]string{
		"depth": "1",
	}
	_, err = jc.goJenkinsClient.Requester.GetJSON(jc.ctx, "/computer", computers, qr)
	if err != nil {
		return -1, -1, err
	}

	return computers.TotalExecutors, computers.BusyExecutors, nil
}

func (jc *JenkinsClient) getServer() string {
	server, _ := url.Parse(jc.goJenkinsClient.Server)
	return server.Hostname()
}
