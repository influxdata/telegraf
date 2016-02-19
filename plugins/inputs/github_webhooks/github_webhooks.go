package github_webhooks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("github_webhooks", func() telegraf.Input { return &GithubWebhooks{} })
}

type GithubWebhooks struct {
	ServiceAddress string
	// Lock for the struct
	sync.Mutex
	// Events buffer to store events between Gather calls
	events []Event
}

func NewGithubWebhooks() *GithubWebhooks {
	return &GithubWebhooks{}
}

func (gh *GithubWebhooks) SampleConfig() string {
	return `
  ## Address and port to host Webhook listener on
  service_address = ":1618"
`
}

func (gh *GithubWebhooks) Description() string {
	return "A Github Webhook Event collector"
}

// Writes the points from <-gh.in to the Accumulator
func (gh *GithubWebhooks) Gather(acc telegraf.Accumulator) error {
	gh.Lock()
	defer gh.Unlock()
	for _, event := range gh.events {
		p := event.NewMetric()
		acc.AddFields("github_webhooks", p.Fields(), p.Tags(), p.Time())
	}
	gh.events = make([]Event, 0)
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

func (gh *GithubWebhooks) Start(_ telegraf.Accumulator) error {
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

func newCommitComment(data []byte) (Event, error) {
	commitCommentStruct := CommitCommentEvent{}
	err := json.Unmarshal(data, &commitCommentStruct)
	if err != nil {
		return nil, err
	}
	return commitCommentStruct, nil
}

func newCreate(data []byte) (Event, error) {
	createStruct := CreateEvent{}
	err := json.Unmarshal(data, &createStruct)
	if err != nil {
		return nil, err
	}
	return createStruct, nil
}

func newDelete(data []byte) (Event, error) {
	deleteStruct := DeleteEvent{}
	err := json.Unmarshal(data, &deleteStruct)
	if err != nil {
		return nil, err
	}
	return deleteStruct, nil
}

func newDeployment(data []byte) (Event, error) {
	deploymentStruct := DeploymentEvent{}
	err := json.Unmarshal(data, &deploymentStruct)
	if err != nil {
		return nil, err
	}
	return deploymentStruct, nil
}

func newDeploymentStatus(data []byte) (Event, error) {
	deploymentStatusStruct := DeploymentStatusEvent{}
	err := json.Unmarshal(data, &deploymentStatusStruct)
	if err != nil {
		return nil, err
	}
	return deploymentStatusStruct, nil
}

func newFork(data []byte) (Event, error) {
	forkStruct := ForkEvent{}
	err := json.Unmarshal(data, &forkStruct)
	if err != nil {
		return nil, err
	}
	return forkStruct, nil
}

func newGollum(data []byte) (Event, error) {
	gollumStruct := GollumEvent{}
	err := json.Unmarshal(data, &gollumStruct)
	if err != nil {
		return nil, err
	}
	return gollumStruct, nil
}

func newIssueComment(data []byte) (Event, error) {
	issueCommentStruct := IssueCommentEvent{}
	err := json.Unmarshal(data, &issueCommentStruct)
	if err != nil {
		return nil, err
	}
	return issueCommentStruct, nil
}

func newIssues(data []byte) (Event, error) {
	issuesStruct := IssuesEvent{}
	err := json.Unmarshal(data, &issuesStruct)
	if err != nil {
		return nil, err
	}
	return issuesStruct, nil
}

func newMember(data []byte) (Event, error) {
	memberStruct := MemberEvent{}
	err := json.Unmarshal(data, &memberStruct)
	if err != nil {
		return nil, err
	}
	return memberStruct, nil
}

func newMembership(data []byte) (Event, error) {
	membershipStruct := MembershipEvent{}
	err := json.Unmarshal(data, &membershipStruct)
	if err != nil {
		return nil, err
	}
	return membershipStruct, nil
}

func newPageBuild(data []byte) (Event, error) {
	pageBuildEvent := PageBuildEvent{}
	err := json.Unmarshal(data, &pageBuildEvent)
	if err != nil {
		return nil, err
	}
	return pageBuildEvent, nil
}

func newPublic(data []byte) (Event, error) {
	publicEvent := PublicEvent{}
	err := json.Unmarshal(data, &publicEvent)
	if err != nil {
		return nil, err
	}
	return publicEvent, nil
}

func newPullRequest(data []byte) (Event, error) {
	pullRequestStruct := PullRequestEvent{}
	err := json.Unmarshal(data, &pullRequestStruct)
	if err != nil {
		return nil, err
	}
	return pullRequestStruct, nil
}

func newPullRequestReviewComment(data []byte) (Event, error) {
	pullRequestReviewCommentStruct := PullRequestReviewCommentEvent{}
	err := json.Unmarshal(data, &pullRequestReviewCommentStruct)
	if err != nil {
		return nil, err
	}
	return pullRequestReviewCommentStruct, nil
}

func newPush(data []byte) (Event, error) {
	pushStruct := PushEvent{}
	err := json.Unmarshal(data, &pushStruct)
	if err != nil {
		return nil, err
	}
	return pushStruct, nil
}

func newRelease(data []byte) (Event, error) {
	releaseStruct := ReleaseEvent{}
	err := json.Unmarshal(data, &releaseStruct)
	if err != nil {
		return nil, err
	}
	return releaseStruct, nil
}

func newRepository(data []byte) (Event, error) {
	repositoryStruct := RepositoryEvent{}
	err := json.Unmarshal(data, &repositoryStruct)
	if err != nil {
		return nil, err
	}
	return repositoryStruct, nil
}

func newStatus(data []byte) (Event, error) {
	statusStruct := StatusEvent{}
	err := json.Unmarshal(data, &statusStruct)
	if err != nil {
		return nil, err
	}
	return statusStruct, nil
}

func newTeamAdd(data []byte) (Event, error) {
	teamAddStruct := TeamAddEvent{}
	err := json.Unmarshal(data, &teamAddStruct)
	if err != nil {
		return nil, err
	}
	return teamAddStruct, nil
}

func newWatch(data []byte) (Event, error) {
	watchStruct := WatchEvent{}
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

func NewEvent(r []byte, t string) (Event, error) {
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
