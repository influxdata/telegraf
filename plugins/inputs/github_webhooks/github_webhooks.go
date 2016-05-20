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
	defer r.Body.Close()
	eventType := r.Header["X-Github-Event"][0]
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e, err := NewEvent(data, eventType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	gh.Lock()
	gh.events = append(gh.events, e)
	gh.Unlock()
	w.WriteHeader(http.StatusOK)
}

func generateEvent(data []byte, event Event) (Event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func newCommitComment(data []byte) (Event, error) {
	return generateEvent(data, &CommitCommentEvent{})
}

func newCreate(data []byte) (Event, error) {
	return generateEvent(data, &CreateEvent{})
}

func newDelete(data []byte) (Event, error) {
	return generateEvent(data, &DeleteEvent{})
}

func newDeployment(data []byte) (Event, error) {
	return generateEvent(data, &DeploymentEvent{})
}

func newDeploymentStatus(data []byte) (Event, error) {
	return generateEvent(data, &DeploymentStatusEvent{})
}

func newFork(data []byte) (Event, error) {
	return generateEvent(data, &ForkEvent{})
}

func newGollum(data []byte) (Event, error) {
	return generateEvent(data, &GollumEvent{})
}

func newIssueComment(data []byte) (Event, error) {
	return generateEvent(data, &IssueCommentEvent{})
}

func newIssues(data []byte) (Event, error) {
	return generateEvent(data, &IssuesEvent{})
}

func newMember(data []byte) (Event, error) {
	return generateEvent(data, &MemberEvent{})
}

func newMembership(data []byte) (Event, error) {
	return generateEvent(data, &MembershipEvent{})
}

func newPageBuild(data []byte) (Event, error) {
	return generateEvent(data, &PageBuildEvent{})
}

func newPublic(data []byte) (Event, error) {
	return generateEvent(data, &PublicEvent{})
}

func newPullRequest(data []byte) (Event, error) {
	return generateEvent(data, &PullRequestEvent{})
}

func newPullRequestReviewComment(data []byte) (Event, error) {
	return generateEvent(data, &PullRequestReviewCommentEvent{})
}

func newPush(data []byte) (Event, error) {
	return generateEvent(data, &PushEvent{})
}

func newRelease(data []byte) (Event, error) {
	return generateEvent(data, &ReleaseEvent{})
}

func newRepository(data []byte) (Event, error) {
	return generateEvent(data, &RepositoryEvent{})
}

func newStatus(data []byte) (Event, error) {
	return generateEvent(data, &StatusEvent{})
}

func newTeamAdd(data []byte) (Event, error) {
	return generateEvent(data, &TeamAddEvent{})
}

func newWatch(data []byte) (Event, error) {
	return generateEvent(data, &WatchEvent{})
}

type newEventError struct {
	s string
}

func (e *newEventError) Error() string {
	return e.s
}

func NewEvent(r []byte, t string) (Event, error) {
	log.Printf("New %v event received", t)
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
	return nil, &newEventError{"Not a recognized event type"}
}
