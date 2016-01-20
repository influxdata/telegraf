package ghwebhooks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	mod "github.com/influxdata/support-tools/ghWebhooks/models"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	log.Info("Starting ghWebhook server...")
	logFile, err := os.Create("server.log")
	if err != nil {
		log.WithFields(log.Fields{
			"time":  time.Now(),
			"error": err,
		}).Warn("Error in creating log file")
	}
	log.SetLevel(log.InfoLevel)
	log.SetOutput(logFile)

	inputs.Add("ghwebhooks", func() inputs.Input { return &GHWebhooks{} })
}

type GHWebhooks struct {
	ServiceAddress  string
	MeasurementName string

	sync.Mutex

	// Channel for all incoming events from github
	in   chan mod.Event
	done chan struct{}
}

func (gh *GHWebhooks) SampleConfig() string {
	return `
  # Address and port to host Webhook listener on
  service_address = ":1618"
	# Measurement name
	measurement_name = "ghWebhooks"
`
}

func (gh *GHWebhooks) Description() string {
	return "Github Webhook Event collector"
}

// Writes the points from <-gh.in to the Accumulator
func (gh *GHWebhooks) Gather(acc inputs.Accumulator) error {
	gh.Lock()
	defer gh.Unlock()
	for {
		select {
		case <-gh.done:
			return nil
		case e := <-gh.in:
			p := e.NewPoint()
			acc.Add(gh.MeasurementName, p.Fields(), p.Tags(), p.Time())
		}
	}
	return nil
}

func (gh *GHWebhooks) Start() error {
	gh.Lock()
	defer gh.Unlock()
	for {
		select {
		case <-gh.done:
			return nil
		default:
			r := mux.NewRouter()
			r.HandleFunc("/webhooks", gh.webhookHandler).Methods("POST")
			http.ListenAndServe(fmt.Sprintf(":%s", gh.ServiceAddress), r)
		}
	}
}

func (gh *GHWebhooks) Stop() {
	gh.Lock()
	defer gh.Unlock()
	log.Println("Stopping the ghWebhooks service")
	close(gh.done)
	close(gh.in)
}

// Handles the /webhooks route
func (gh *GHWebhooks) webhookHandler(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header["X-Github-Event"][0]
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": eventType, "error": err}
		log.WithFields(fields).Fatal("Error reading Github payload")
	}

	// Send event down chan to GHWebhooks
	e := NewEvent(data, eventType)
	gh.in <- e
	fmt.Printf("%v\n", e.NewPoint())
	w.WriteHeader(http.StatusOK)
}

func newCommitComment(data []byte) mod.Event {
	commitCommentStruct := mod.CommitCommentEvent{}
	err := json.Unmarshal(data, &commitCommentStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "CommitCommentEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return commitCommentStruct
}

func newCreate(data []byte) mod.Event {
	createStruct := mod.CreateEvent{}
	err := json.Unmarshal(data, &createStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "CreateEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return createStruct
}

func newDelete(data []byte) mod.Event {
	deleteStruct := mod.DeleteEvent{}
	err := json.Unmarshal(data, &deleteStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "DeleteEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return deleteStruct
}

func newDeployment(data []byte) mod.Event {
	deploymentStruct := mod.DeploymentEvent{}
	err := json.Unmarshal(data, &deploymentStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "DeploymentEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return deploymentStruct
}

func newDeploymentStatus(data []byte) mod.Event {
	deploymentStatusStruct := mod.DeploymentStatusEvent{}
	err := json.Unmarshal(data, &deploymentStatusStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "DeploymentStatusEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return deploymentStatusStruct
}

func newFork(data []byte) mod.Event {
	forkStruct := mod.ForkEvent{}
	err := json.Unmarshal(data, &forkStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "ForkEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return forkStruct
}

func newGollum(data []byte) mod.Event {
	gollumStruct := mod.GollumEvent{}
	err := json.Unmarshal(data, &gollumStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "GollumEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return gollumStruct
}

func newIssueComment(data []byte) mod.Event {
	issueCommentStruct := mod.IssueCommentEvent{}
	err := json.Unmarshal(data, &issueCommentStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "IssueCommentEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return issueCommentStruct
}

func newIssues(data []byte) mod.Event {
	issuesStruct := mod.IssuesEvent{}
	err := json.Unmarshal(data, &issuesStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "IssuesEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return issuesStruct
}

func newMember(data []byte) mod.Event {
	memberStruct := mod.MemberEvent{}
	err := json.Unmarshal(data, &memberStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "MemberEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return memberStruct
}

func newMembership(data []byte) mod.Event {
	membershipStruct := mod.MembershipEvent{}
	err := json.Unmarshal(data, &membershipStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "MembershipEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return membershipStruct
}

func newPageBuild(data []byte) mod.Event {
	pageBuildEvent := mod.PageBuildEvent{}
	err := json.Unmarshal(data, &pageBuildEvent)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "PageBuildEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return pageBuildEvent
}

func newPublic(data []byte) mod.Event {
	publicEvent := mod.PublicEvent{}
	err := json.Unmarshal(data, &publicEvent)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "PublicEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return publicEvent
}

func newPullRequest(data []byte) mod.Event {
	pullRequestStruct := mod.PullRequestEvent{}
	err := json.Unmarshal(data, &pullRequestStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "PullRequestEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return pullRequestStruct
}

func newPullRequestReviewComment(data []byte) mod.Event {
	pullRequestReviewCommentStruct := mod.PullRequestReviewCommentEvent{}
	err := json.Unmarshal(data, &pullRequestReviewCommentStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "PullRequestReviewCommentEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return pullRequestReviewCommentStruct
}

func newPush(data []byte) mod.Event {
	pushStruct := mod.PushEvent{}
	err := json.Unmarshal(data, &pushStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "PushEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return pushStruct
}

func newRelease(data []byte) mod.Event {
	releaseStruct := mod.ReleaseEvent{}
	err := json.Unmarshal(data, &releaseStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "ReleaseEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return releaseStruct
}

func newRepository(data []byte) mod.Event {
	repositoryStruct := mod.RepositoryEvent{}
	err := json.Unmarshal(data, &repositoryStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "RepositoryEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return repositoryStruct
}

func newStatus(data []byte) mod.Event {
	statusStruct := mod.StatusEvent{}
	err := json.Unmarshal(data, &statusStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "StatusEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return statusStruct
}

func newTeamAdd(data []byte) mod.Event {
	teamAddStruct := mod.TeamAddEvent{}
	err := json.Unmarshal(data, &teamAddStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "TeamAddEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return teamAddStruct
}

func newWatch(data []byte) mod.Event {
	watchStruct := mod.WatchEvent{}
	err := json.Unmarshal(data, &watchStruct)
	if err != nil {
		fields := log.Fields{"time": time.Now(), "event": "WatchEvent", "error": err}
		log.WithFields(fields).Fatalf("Error in unmarshaling JSON")
	}
	return watchStruct
}

func NewEvent(r []byte, t string) mod.Event {
	log.WithFields(log.Fields{"event": t, "time": time.Now()}).Info("Event Recieved")
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
	return nil
}
