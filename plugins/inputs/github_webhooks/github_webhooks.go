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

type newEventError struct {
	s string
}

func (e *newEventError) Error() string {
	return e.s
}

func NewEvent(data []byte, name string) (Event, error) {
	log.Printf("New %v event received", name)
	switch name {
	case "commit_comment":
		return generateEvent(data, &CommitCommentEvent{})
	case "create":
		return generateEvent(data, &CreateEvent{})
	case "delete":
		return generateEvent(data, &DeleteEvent{})
	case "deployment":
		return generateEvent(data, &DeploymentEvent{})
	case "deployment_status":
		return generateEvent(data, &DeploymentStatusEvent{})
	case "fork":
		return generateEvent(data, &ForkEvent{})
	case "gollum":
		return generateEvent(data, &GollumEvent{})
	case "issue_comment":
		return generateEvent(data, &IssueCommentEvent{})
	case "issues":
		return generateEvent(data, &IssuesEvent{})
	case "member":
		return generateEvent(data, &MemberEvent{})
	case "membership":
		return generateEvent(data, &MembershipEvent{})
	case "page_build":
		return generateEvent(data, &PageBuildEvent{})
	case "public":
		return generateEvent(data, &PublicEvent{})
	case "pull_request":
		return generateEvent(data, &PullRequestEvent{})
	case "pull_request_review_comment":
		return generateEvent(data, &PullRequestReviewCommentEvent{})
	case "push":
		return generateEvent(data, &PushEvent{})
	case "release":
		return generateEvent(data, &ReleaseEvent{})
	case "repository":
		return generateEvent(data, &RepositoryEvent{})
	case "status":
		return generateEvent(data, &StatusEvent{})
	case "team_add":
		return generateEvent(data, &TeamAddEvent{})
	case "watch":
		return generateEvent(data, &WatchEvent{})
	}
	return nil, &newEventError{"Not a recognized event type"}
}
