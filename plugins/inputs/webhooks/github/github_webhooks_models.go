package github

import (
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const meas = "github_webhooks"

type event interface {
	newMetric() telegraf.Metric
}

type repository struct {
	Repository string `json:"full_name"`
	Private    bool   `json:"private"`
	Stars      int    `json:"stargazers_count"`
	Forks      int    `json:"forks_count"`
	Issues     int    `json:"open_issues_count"`
}

type sender struct {
	User  string `json:"login"`
	Admin bool   `json:"site_admin"`
}

type commitComment struct {
	Commit string `json:"commit_id"`
	Body   string `json:"body"`
}

type deployment struct {
	Commit      string `json:"sha"`
	Task        string `json:"task"`
	Environment string `json:"environment"`
	Description string `json:"description"`
}

type page struct {
	Name   string `json:"page_name"`
	Title  string `json:"title"`
	Action string `json:"action"`
}

type issue struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Comments int    `json:"comments"`
}

type issueComment struct {
	Body string `json:"body"`
}

type team struct {
	Name string `json:"name"`
}

type pullRequest struct {
	Number       int    `json:"number"`
	State        string `json:"state"`
	Title        string `json:"title"`
	Comments     int    `json:"comments"`
	Commits      int    `json:"commits"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ChangedFiles int    `json:"changed_files"`
}

type pullRequestReviewComment struct {
	File    string `json:"path"`
	Comment string `json:"body"`
}

type workflowJob struct {
	RunAttempt  int       `json:"run_attempt"`
	HeadBranch  string    `json:"head_branch"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	Name        string    `json:"name"`
	Conclusion  string    `json:"conclusion"`
}

type workflowRun struct {
	HeadBranch   string    `json:"head_branch"`
	CreatedAt    time.Time `json:"created_at"`
	RunStartedAt time.Time `json:"run_started_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RunAttempt   int       `json:"run_attempt"`
	Name         string    `json:"name"`
	Conclusion   string    `json:"conclusion"`
}

type release struct {
	TagName string `json:"tag_name"`
}

type deploymentStatus struct {
	State       string `json:"state"`
	Description string `json:"description"`
}

type commitCommentEvent struct {
	Comment    commitComment `json:"comment"`
	Repository repository    `json:"repository"`
	Sender     sender        `json:"sender"`
}

func (s commitCommentEvent) newMetric() telegraf.Metric {
	event := "commit_comment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"commit":  s.Comment.Commit,
		"comment": s.Comment.Body,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type createEvent struct {
	Ref        string     `json:"ref"`
	RefType    string     `json:"ref_type"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s createEvent) newMetric() telegraf.Metric {
	event := "create"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"ref":     s.Ref,
		"refType": s.RefType,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type deleteEvent struct {
	Ref        string     `json:"ref"`
	RefType    string     `json:"ref_type"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s deleteEvent) newMetric() telegraf.Metric {
	event := "delete"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"ref":     s.Ref,
		"refType": s.RefType,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type deploymentEvent struct {
	Deployment deployment `json:"deployment"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s deploymentEvent) newMetric() telegraf.Metric {
	event := "deployment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":       s.Repository.Stars,
		"forks":       s.Repository.Forks,
		"issues":      s.Repository.Issues,
		"commit":      s.Deployment.Commit,
		"task":        s.Deployment.Task,
		"environment": s.Deployment.Environment,
		"description": s.Deployment.Description,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type deploymentStatusEvent struct {
	Deployment       deployment       `json:"deployment"`
	DeploymentStatus deploymentStatus `json:"deployment_status"`
	Repository       repository       `json:"repository"`
	Sender           sender           `json:"sender"`
}

func (s deploymentStatusEvent) newMetric() telegraf.Metric {
	event := "delete"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":          s.Repository.Stars,
		"forks":          s.Repository.Forks,
		"issues":         s.Repository.Issues,
		"commit":         s.Deployment.Commit,
		"task":           s.Deployment.Task,
		"environment":    s.Deployment.Environment,
		"description":    s.Deployment.Description,
		"depState":       s.DeploymentStatus.State,
		"depDescription": s.DeploymentStatus.Description,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type forkEvent struct {
	Forkee     repository `json:"forkee"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s forkEvent) newMetric() telegraf.Metric {
	event := "fork"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
		"fork":   s.Forkee.Repository,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type gollumEvent struct {
	Pages      []page     `json:"pages"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

// REVIEW: Going to be lazy and not deal with the pages.
func (s gollumEvent) newMetric() telegraf.Metric {
	event := "gollum"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type issueCommentEvent struct {
	Issue      issue        `json:"issue"`
	Comment    issueComment `json:"comment"`
	Repository repository   `json:"repository"`
	Sender     sender       `json:"sender"`
}

func (s issueCommentEvent) newMetric() telegraf.Metric {
	event := "issue_comment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
		"issue":      strconv.Itoa(s.Issue.Number),
	}
	f := map[string]interface{}{
		"stars":    s.Repository.Stars,
		"forks":    s.Repository.Forks,
		"issues":   s.Repository.Issues,
		"title":    s.Issue.Title,
		"comments": s.Issue.Comments,
		"body":     s.Comment.Body,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type issuesEvent struct {
	Action     string     `json:"action"`
	Issue      issue      `json:"issue"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s issuesEvent) newMetric() telegraf.Metric {
	event := "issue"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
		"issue":      strconv.Itoa(s.Issue.Number),
		"action":     s.Action,
	}
	f := map[string]interface{}{
		"stars":    s.Repository.Stars,
		"forks":    s.Repository.Forks,
		"issues":   s.Repository.Issues,
		"title":    s.Issue.Title,
		"comments": s.Issue.Comments,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type memberEvent struct {
	Member     sender     `json:"member"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s memberEvent) newMetric() telegraf.Metric {
	event := "member"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":           s.Repository.Stars,
		"forks":           s.Repository.Forks,
		"issues":          s.Repository.Issues,
		"newMember":       s.Member.User,
		"newMemberStatus": s.Member.Admin,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type membershipEvent struct {
	Action string `json:"action"`
	Member sender `json:"member"`
	Sender sender `json:"sender"`
	Team   team   `json:"team"`
}

func (s membershipEvent) newMetric() telegraf.Metric {
	event := "membership"
	t := map[string]string{
		"event":  event,
		"user":   s.Sender.User,
		"admin":  strconv.FormatBool(s.Sender.Admin),
		"action": s.Action,
	}
	f := map[string]interface{}{
		"newMember":       s.Member.User,
		"newMemberStatus": s.Member.Admin,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type pageBuildEvent struct {
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s pageBuildEvent) newMetric() telegraf.Metric {
	event := "page_build"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type publicEvent struct {
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s publicEvent) newMetric() telegraf.Metric {
	event := "public"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type pullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest pullRequest `json:"pull_request"`
	Repository  repository  `json:"repository"`
	Sender      sender      `json:"sender"`
}

func (s pullRequestEvent) newMetric() telegraf.Metric {
	event := "pull_request"
	t := map[string]string{
		"event":      event,
		"action":     s.Action,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
		"prNumber":   strconv.Itoa(s.PullRequest.Number),
	}
	f := map[string]interface{}{
		"stars":        s.Repository.Stars,
		"forks":        s.Repository.Forks,
		"issues":       s.Repository.Issues,
		"state":        s.PullRequest.State,
		"title":        s.PullRequest.Title,
		"comments":     s.PullRequest.Comments,
		"commits":      s.PullRequest.Commits,
		"additions":    s.PullRequest.Additions,
		"deletions":    s.PullRequest.Deletions,
		"changedFiles": s.PullRequest.ChangedFiles,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type pullRequestReviewCommentEvent struct {
	Comment     pullRequestReviewComment `json:"comment"`
	PullRequest pullRequest              `json:"pull_request"`
	Repository  repository               `json:"repository"`
	Sender      sender                   `json:"sender"`
}

func (s pullRequestReviewCommentEvent) newMetric() telegraf.Metric {
	event := "pull_request_review_comment"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
		"prNumber":   strconv.Itoa(s.PullRequest.Number),
	}
	f := map[string]interface{}{
		"stars":        s.Repository.Stars,
		"forks":        s.Repository.Forks,
		"issues":       s.Repository.Issues,
		"state":        s.PullRequest.State,
		"title":        s.PullRequest.Title,
		"comments":     s.PullRequest.Comments,
		"commits":      s.PullRequest.Commits,
		"additions":    s.PullRequest.Additions,
		"deletions":    s.PullRequest.Deletions,
		"changedFiles": s.PullRequest.ChangedFiles,
		"commentFile":  s.Comment.File,
		"comment":      s.Comment.Comment,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type pushEvent struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s pushEvent) newMetric() telegraf.Metric {
	event := "push"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
		"ref":    s.Ref,
		"before": s.Before,
		"after":  s.After,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type releaseEvent struct {
	Release    release    `json:"release"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s releaseEvent) newMetric() telegraf.Metric {
	event := "release"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":   s.Repository.Stars,
		"forks":   s.Repository.Forks,
		"issues":  s.Repository.Issues,
		"tagName": s.Release.TagName,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type repositoryEvent struct {
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s repositoryEvent) newMetric() telegraf.Metric {
	event := "repository"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type statusEvent struct {
	Commit     string     `json:"sha"`
	State      string     `json:"state"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s statusEvent) newMetric() telegraf.Metric {
	event := "status"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
		"commit": s.Commit,
		"state":  s.State,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type teamAddEvent struct {
	Team       team       `json:"team"`
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s teamAddEvent) newMetric() telegraf.Metric {
	event := "team_add"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":    s.Repository.Stars,
		"forks":    s.Repository.Forks,
		"issues":   s.Repository.Issues,
		"teamName": s.Team.Name,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type watchEvent struct {
	Repository repository `json:"repository"`
	Sender     sender     `json:"sender"`
}

func (s watchEvent) newMetric() telegraf.Metric {
	event := "delete"
	t := map[string]string{
		"event":      event,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
	}
	f := map[string]interface{}{
		"stars":  s.Repository.Stars,
		"forks":  s.Repository.Forks,
		"issues": s.Repository.Issues,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type workflowJobEvent struct {
	Action      string      `json:"action"`
	WorkflowJob workflowJob `json:"workflow_job"`
	Repository  repository  `json:"repository"`
	Sender      sender      `json:"sender"`
}

func (s workflowJobEvent) newMetric() telegraf.Metric {
	event := "workflow_job"
	t := map[string]string{
		"event":      event,
		"action":     s.Action,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
		"name":       s.WorkflowJob.Name,
		"conclusion": s.WorkflowJob.Conclusion,
	}

	var runTimeMs int64
	var queueTimeMs int64
	if s.Action == "in_progress" {
		queueTimeMs = s.WorkflowJob.StartedAt.Sub(s.WorkflowJob.CreatedAt).Milliseconds()
	}
	if s.Action == "completed" {
		runTimeMs = s.WorkflowJob.CompletedAt.Sub(s.WorkflowJob.StartedAt).Milliseconds()
	}
	f := map[string]interface{}{
		"run_attempt": s.WorkflowJob.RunAttempt,
		"queue_time":  queueTimeMs,
		"run_time":    runTimeMs,
		"head_branch": s.WorkflowJob.HeadBranch,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}

type workflowRunEvent struct {
	Action      string      `json:"action"`
	WorkflowRun workflowRun `json:"workflow_run"`
	Repository  repository  `json:"repository"`
	Sender      sender      `json:"sender"`
}

func (s workflowRunEvent) newMetric() telegraf.Metric {
	event := "workflow_run"
	t := map[string]string{
		"event":      event,
		"action":     s.Action,
		"repository": s.Repository.Repository,
		"private":    strconv.FormatBool(s.Repository.Private),
		"user":       s.Sender.User,
		"admin":      strconv.FormatBool(s.Sender.Admin),
		"name":       s.WorkflowRun.Name,
		"conclusion": s.WorkflowRun.Conclusion,
	}
	var runTimeMs int64

	if s.Action == "completed" {
		runTimeMs = s.WorkflowRun.UpdatedAt.Sub(s.WorkflowRun.RunStartedAt).Milliseconds()
	}
	f := map[string]interface{}{
		"run_attempt": s.WorkflowRun.RunAttempt,
		"run_time":    runTimeMs,
		"head_branch": s.WorkflowRun.HeadBranch,
	}
	m := metric.New(meas, t, f, time.Now())
	return m
}
