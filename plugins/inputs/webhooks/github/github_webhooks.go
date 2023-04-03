package github

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // G505: Blocklisted import crypto/sha1: weak cryptographic primitive - sha1 hash is what is desired in this case
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/auth"
)

type GithubWebhook struct {
	Path   string
	Secret string
	acc    telegraf.Accumulator
	log    telegraf.Logger
	auth.BasicAuth
}

func (gh *GithubWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(gh.Path, gh.eventHandler).Methods("POST")

	gh.log = log
	gh.log.Infof("Started the webhooks_github on %s", gh.Path)
	gh.acc = acc
}

func (gh *GithubWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if !gh.Verify(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-Github-Event")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if gh.Secret != "" && !checkSignature(gh.Secret, data, r.Header.Get("X-Hub-Signature")) {
		gh.log.Error("Fail to check the github webhook signature")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, err := gh.NewEvent(data, eventType)
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

func (gh *GithubWebhook) NewEvent(data []byte, name string) (Event, error) {
	gh.log.Debugf("New %v event received", name)
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
	if _, err := mac.Write(data); err != nil {
		return err.Error()
	}
	result := mac.Sum(nil)
	return "sha1=" + hex.EncodeToString(result)
}
