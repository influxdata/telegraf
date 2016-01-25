package github_webhooks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf/plugins/inputs"
	mod "github.com/influxdata/telegraf/plugins/inputs/github_webhooks/models"
)

func init() {
	inputs.Add("github_webhooks", func() inputs.Input { return &GithubWebhooks{} })
}

type GithubWebhooks struct {
	ServiceAddress  string
	MeasurementName string
	// Lock for the struct
	sync.Mutex
	// Events buffer to store events between Gather calls
	events []mod.Event
}

func NewGithubWebhooks() *GithubWebhooks {
	return &GithubWebhooks{}
}

func (gh *GithubWebhooks) SampleConfig() string {
	return `
  # Address and port to host Webhook listener on
  service_address = ":1618"
  # Measurement name
  measurement_name = "github_webhooks"
`
}

func (gh *GithubWebhooks) Description() string {
	return "A Github Webhook Event collector"
}

// Writes the points from <-gh.in to the Accumulator
func (gh *GithubWebhooks) Gather(acc inputs.Accumulator) error {
	gh.Lock()
	defer gh.Unlock()
	for _, event := range gh.events {
		p := event.NewPoint()
		acc.AddFields(gh.MeasurementName, p.Fields(), p.Tags(), p.Time())
	}
	gh.events = make([]mod.Event, 0)
	return nil
}

func (gh *GithubWebhooks) Listen() {
	r := mux.NewRouter()
	r.HandleFunc("/", gh.eventHandler).Methods("POST")
	err := http.ListenAndServe(fmt.Sprintf("%s", gh.ServiceAddress), r)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

func (gh *GithubWebhooks) Start() error {
	go gh.Listen()
	log.Printf("Started the github_webhooks service on %s\n", gh.ServiceAddress)
	return nil
}

func (gh *GithubWebhooks) Stop() {
	log.Println("Stopping the ghWebhooks service")
}

// Handles the / route
func (gh *GithubWebhooks) eventHandler(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header["X-Github-Event"][0]
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	e, err := NewEvent(data, eventType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	gh.Lock()
	gh.events = append(gh.events, e)
	gh.Unlock()
	w.WriteHeader(http.StatusOK)
}

func newCommitComment(data []byte) (mod.Event, error) {
	commitCommentStruct := mod.CommitCommentEvent{}
	err := json.Unmarshal(data, &commitCommentStruct)
	if err != nil {
		return nil, err
	}
	return commitCommentStruct, nil
}

func newCreate(data []byte) (mod.Event, error) {
	createStruct := mod.CreateEvent{}
	err := json.Unmarshal(data, &createStruct)
	if err != nil {
		return nil, err
	}
	return createStruct, nil
}

func newDelete(data []byte) (mod.Event, error) {
	deleteStruct := mod.DeleteEvent{}
	err := json.Unmarshal(data, &deleteStruct)
	if err != nil {
		return nil, err
	}
	return deleteStruct, nil
}

func newDeployment(data []byte) (mod.Event, error) {
	deploymentStruct := mod.DeploymentEvent{}
	err := json.Unmarshal(data, &deploymentStruct)
	if err != nil {
		return nil, err
	}
	return deploymentStruct, nil
}

func newDeploymentStatus(data []byte) (mod.Event, error) {
	deploymentStatusStruct := mod.DeploymentStatusEvent{}
	err := json.Unmarshal(data, &deploymentStatusStruct)
	if err != nil {
		return nil, err
	}
	return deploymentStatusStruct, nil
}

func newFork(data []byte) (mod.Event, error) {
	forkStruct := mod.ForkEvent{}
	err := json.Unmarshal(data, &forkStruct)
	if err != nil {
		return nil, err
	}
	return forkStruct, nil
}

func newGollum(data []byte) (mod.Event, error) {
	gollumStruct := mod.GollumEvent{}
	err := json.Unmarshal(data, &gollumStruct)
	if err != nil {
		return nil, err
	}
	return gollumStruct, nil
}

func newIssueComment(data []byte) (mod.Event, error) {
	issueCommentStruct := mod.IssueCommentEvent{}
	err := json.Unmarshal(data, &issueCommentStruct)
	if err != nil {
		return nil, err
	}
	return issueCommentStruct, nil
}

func newIssues(data []byte) (mod.Event, error) {
	issuesStruct := mod.IssuesEvent{}
	err := json.Unmarshal(data, &issuesStruct)
	if err != nil {
		return nil, err
	}
	return issuesStruct, nil
}

func newMember(data []byte) (mod.Event, error) {
	memberStruct := mod.MemberEvent{}
	err := json.Unmarshal(data, &memberStruct)
	if err != nil {
		return nil, err
	}
	return memberStruct, nil
}

func newMembership(data []byte) (mod.Event, error) {
	membershipStruct := mod.MembershipEvent{}
	err := json.Unmarshal(data, &membershipStruct)
	if err != nil {
		return nil, err
	}
	return membershipStruct, nil
}

func newPageBuild(data []byte) (mod.Event, error) {
	pageBuildEvent := mod.PageBuildEvent{}
	err := json.Unmarshal(data, &pageBuildEvent)
	if err != nil {
		return nil, err
	}
	return pageBuildEvent, nil
}

func newPublic(data []byte) (mod.Event, error) {
	publicEvent := mod.PublicEvent{}
	err := json.Unmarshal(data, &publicEvent)
	if err != nil {
		return nil, err
	}
	return publicEvent, nil
}

func newPullRequest(data []byte) (mod.Event, error) {
	pullRequestStruct := mod.PullRequestEvent{}
	err := json.Unmarshal(data, &pullRequestStruct)
	if err != nil {
		return nil, err
	}
	return pullRequestStruct, nil
}

func newPullRequestReviewComment(data []byte) (mod.Event, error) {
	pullRequestReviewCommentStruct := mod.PullRequestReviewCommentEvent{}
	err := json.Unmarshal(data, &pullRequestReviewCommentStruct)
	if err != nil {
		return nil, err
	}
	return pullRequestReviewCommentStruct, nil
}

func newPush(data []byte) (mod.Event, error) {
	pushStruct := mod.PushEvent{}
	err := json.Unmarshal(data, &pushStruct)
	if err != nil {
		return nil, err
	}
	return pushStruct, nil
}

func newRelease(data []byte) (mod.Event, error) {
	releaseStruct := mod.ReleaseEvent{}
	err := json.Unmarshal(data, &releaseStruct)
	if err != nil {
		return nil, err
	}
	return releaseStruct, nil
}

func newRepository(data []byte) (mod.Event, error) {
	repositoryStruct := mod.RepositoryEvent{}
	err := json.Unmarshal(data, &repositoryStruct)
	if err != nil {
		return nil, err
	}
	return repositoryStruct, nil
}

func newStatus(data []byte) (mod.Event, error) {
	statusStruct := mod.StatusEvent{}
	err := json.Unmarshal(data, &statusStruct)
	if err != nil {
		return nil, err
	}
	return statusStruct, nil
}

func newTeamAdd(data []byte) (mod.Event, error) {
	teamAddStruct := mod.TeamAddEvent{}
	err := json.Unmarshal(data, &teamAddStruct)
	if err != nil {
		return nil, err
	}
	return teamAddStruct, nil
}

func newWatch(data []byte) (mod.Event, error) {
	watchStruct := mod.WatchEvent{}
	err := json.Unmarshal(data, &watchStruct)
	if err != nil {
		return nil, err
	}
	return watchStruct, nil
}

type newEventError struct {
	s string
}

func (e *newEventError) Error() string {
	return e.s
}

func NewEvent(r []byte, t string) (mod.Event, error) {
	log.Printf("New %v event recieved", t)
	switch t {
	case "commit_comment":
		return newCommitComment(r)
	case "create":
		return newCreate(r)
	case "delete":
		return newDelete(r)
	case "deployment":
		return newDeployment(r)
	case "deployment_status":
		return newDeploymentStatus(r)
	case "fork":
		return newFork(r)
	case "gollum":
		return newGollum(r)
	case "issue_comment":
		return newIssueComment(r)
	case "issues":
		return newIssues(r)
	case "member":
		return newMember(r)
	case "membership":
		return newMembership(r)
	case "page_build":
		return newPageBuild(r)
	case "public":
		return newPublic(r)
	case "pull_request":
		return newPullRequest(r)
	case "pull_request_review_comment":
		return newPullRequestReviewComment(r)
	case "push":
		return newPush(r)
	case "release":
		return newRelease(r)
	case "repository":
		return newRepository(r)
	case "status":
		return newStatus(r)
	case "team_add":
		return newTeamAdd(r)
	case "watch":
		return newWatch(r)
	}
	return nil, &newEventError{"Not a recgonized event type"}
}
