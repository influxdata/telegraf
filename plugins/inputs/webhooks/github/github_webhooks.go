package github

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type GithubWebhook struct {
	Path   string
	Secret string
	acc    telegraf.Accumulator
}

func (gh *GithubWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(gh.Path, gh.eventHandler).Methods("POST")
	log.Printf("I! Started the webhooks_github on %s\n", gh.Path)
	gh.acc = acc
}

func (gh *GithubWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	eventType := r.Header.Get("X-Github-Event")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if gh.Secret != "" && !checkSignature(gh.Secret, data, r.Header.Get("X-Hub-Signature")) {
		log.Printf("E! Fail to check the github webhook signature\n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, err := NewEvent(data, eventType)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if e != nil {
		p := e.NewMetric()
		gh.acc.AddFields("github_webhooks", p.Fields(), p.Tags(), p.Time())
	}

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
	log.Printf("D! New %v event received", name)
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
	case "ping":
		return nil, nil
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

func checkSignature(secret string, data []byte, signature string) bool {
	return hmac.Equal([]byte(signature), []byte(generateSignature(secret, data)))
}

func generateSignature(secret string, data []byte) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(data)
	result := mac.Sum(nil)
	return "sha1=" + hex.EncodeToString(result)
}
