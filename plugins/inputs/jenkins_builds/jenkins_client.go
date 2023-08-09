package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BuildInfo struct {
	Class                     string   `json:"_class"`
	Building                  bool     `json:"building"`
	Duration                  float64  `json:"duration"`
	EstimatedDuration         float64  `json:"estimatedDuration"`
	Number                    int64    `json:"number"`
	Result                    string   `json:"result"`
	Timestamp                 int64    `json:"timestamp"`
	URL                       string   `json:"url"`
	Actions                   []Action `json:"actions"`
	BuilderAllocationDuration float64
}

type BuilderAllocation struct {
	Duration float64 `json:"durationMillis"`
}

type Action struct {
	Class                   string  `json:"_class"`
	BlockedDurationMillis   float64 `json:"blockedDurationMillis"`
	BlockedTimeMillis       float64 `json:"blockedTimeMillis"`
	BuildableDurationMillis float64 `json:"buildableDurationMillis"`
	BuildableTimeMillis     float64 `json:"buildableTimeMillis"`
	BuildingDurationMillis  float64 `json:"buildingDurationMillis"`
	ExecutingTimeMillis     float64 `json:"executingTimeMillis"`
	ExecutorUtilization     float64 `json:"executorUtilization"`
	SubTaskCount            float64 `json:"subTaskCount"`
	WaitingDurationMillis   float64 `json:"waitingDurationMillis"`
	WaitingTimeMillis       float64 `json:"waitingTimeMillis"`
}

func (b *BuildInfo) GetTimestamp() time.Time {
	msInt := int64(b.Timestamp)
	return time.Unix(0, msInt*int64(time.Millisecond))
}

type Jenkins struct {
	Server string `json:"url"`
}

type ExecutorsInfo struct {
	BusyExecutors  int `json:"busyExecutors"`
	TotalExecutors int `json:"totalExecutors"`
}

type JobBuild struct {
	Number int64
}

type JenkinsClient struct {
	httpClient    *http.Client
	ctx           context.Context
	url           string
	username      string
	password      string
	sessionCookie *http.Cookie
	jenkins       *Jenkins
}

type APIError struct {
	URL         string
	StatusCode  int
	Title       string
	Description string
}

func (e APIError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("[%s] %s: %s", e.URL, e.Title, e.Description)
	}
	return fmt.Sprintf("[%s] %s", e.URL, e.Title)
}

type Job struct {
	Class string `json:"_class"`
	Url   string `json:"url"`
	Color string `json:"color"`
	Jobs  []Job  `json:"jobs"`
}

type JobsResponse struct {
	Class string `json:"_class"`
	Jobs  []Job  `json:"jobs"`
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
	req, err := http.NewRequest("GET", jc.url, nil)
	if err != nil {
		return err
	}
	if jc.username != "" || jc.password != "" {
		req.SetBasicAuth(jc.username, jc.password)
	}
	resp, err := jc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for _, cc := range resp.Cookies() {
		if strings.Contains(cc.Name, "JSESSIONID") {
			jc.sessionCookie = cc
			break
		}
	}

	jenkins := new(Jenkins)
	err = jc.doGet("/api/json", jenkins)

	if err != nil {
		return err
	}

	jc.jenkins = jenkins
	return nil
}

func (jc *JenkinsClient) isFolder(class string) bool {
	if strings.Contains(class, "com.cloudbees.hudson.plugins.folder.Folder") {
		return true
	}
	return false
}

func (jc *JenkinsClient) isMultiBranch(class string) bool {
	if strings.Contains(class, "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject") {
		return true
	}
	return false
}

func createJobQueryString(depth int) string {
	queryString := "tree="
	queryString += strings.Repeat("jobs[url,color,buildable,", depth-1)
	queryString += "jobs[url]"
	end := strings.Repeat("]", depth-1)
	queryString += end

	return queryString
}

func (jc *JenkinsClient) getBuildInfo(base string, id int64) (*BuildInfo, error) {
	build := new(BuildInfo)
	url := base + "/" + fmt.Sprintf("%d", id) + "/api/json"
	err := jc.doGet(url, build)
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (jc *JenkinsClient) getAllBuildIds(base string) ([]JobBuild, error) {
	var buildsResp struct {
		Builds []JobBuild `json:"allBuilds"`
	}
	url := base + "/api/json?tree=allBuilds[number]"
	err := jc.doGet(url, &buildsResp)
	if err != nil {
		return nil, err
	}
	return buildsResp.Builds, nil
}

func (jc *JenkinsClient) getAllJobs(depth int) ([]JobInfo, error) {

	qr := createJobQueryString(depth)
	url := "/api/json?" + qr
	allJobsResponse := new(JobsResponse)
	err := jc.doGet(url, allJobsResponse)
	if err != nil {
		return nil, err
	}

	var allJobs []JobInfo

	allJobs = append(jc.collectJobs(allJobsResponse.Jobs, allJobs))
	return allJobs, nil
}

func (jc *JenkinsClient) collectJobs(jobs []Job, allJobs []JobInfo) []JobInfo {
	for _, job := range jobs {
		if nil != job.Jobs {
			allJobs = jc.collectJobs(job.Jobs, allJobs)
		}

		if jc.isFolder(job.Class) || jc.isMultiBranch(job.Class) {
			continue
		}

		folder, parents, base := getFolderPath(job.Url)
		allJobs = append(allJobs, JobInfo{
			Base:    base,
			Class:   job.Class,
			Name:    folder,
			Url:     job.Url,
			Color:   job.Color,
			Folder:  folder,
			Parents: parents,
		})
	}
	return allJobs
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

func (jc *JenkinsClient) getExecutorsInfo() (total int, busy int, err error) {
	computers := new(ExecutorsInfo)

	err = jc.doGet("/computer/api/json?depth=1", computers)
	if err != nil {
		return -1, -1, err
	}

	return computers.TotalExecutors, computers.BusyExecutors, nil
}

func (jc *JenkinsClient) getNodeAllocationDuration(base string, id int64) float64 {
	nodeAllocation := new(BuilderAllocation)
	url := base + "/" + fmt.Sprintf("%d", id) + "/execution/node/3/wfapi/describe"
	err := jc.doGet(url, nodeAllocation)
	if err != nil {
		return 0
	}
	return nodeAllocation.Duration
}

func (jc *JenkinsClient) getServer() string {
	server, _ := url.Parse(jc.jenkins.Server)
	return server.Hostname()
}

func (jc *JenkinsClient) doGet(url string, responseType interface{}) error {
	req, err := createGetRequest(jc.url+url, jc.username, jc.password, jc.sessionCookie)
	if err != nil {
		return err
	}
	resp, err := jc.httpClient.Do(req.WithContext(jc.ctx))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		jc.sessionCookie = nil
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}
	if resp.StatusCode == http.StatusNoContent {
		return APIError{
			URL:        url,
			StatusCode: resp.StatusCode,
			Title:      resp.Status,
		}
	}

	return json.NewDecoder(resp.Body).Decode(responseType)
}

func createGetRequest(url string, username, password string, sessionCookie *http.Cookie) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}
	req.Header.Add("Accept", "application/json")
	return req, nil
}
