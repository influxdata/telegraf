package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	
)

func githubWebhookRequest(t *testing.T, event, jsonString string) *testutil.Accumulator {
	var acc testutil.Accumulator
	gh := &Webhook{Path: "/github", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/github", strings.NewReader(jsonString))
	require.NoError(t, err)
	req.Header.Add("X-Github-Event", event)
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
	return &acc
}

func githubWebhookRequestWithSignature(t *testing.T, event, jsonString, signature string, expectedStatus int) {
	var acc testutil.Accumulator
	gh := &Webhook{Path: "/github", Secret: "signature", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/github", strings.NewReader(jsonString))
	require.NoError(t, err)
	req.Header.Add("X-Github-Event", event)
	req.Header.Add("X-Hub-Signature", signature)
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != expectedStatus {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, expectedStatus)
	}
}

func TestCommitCommentEvent(t *testing.T) {
	githubWebhookRequest(t, "commit_comment", commitCommentEventJSON())
}

func TestPingEvent(t *testing.T) {
	githubWebhookRequest(t, "ping", "")
}

func TestDeleteEvent(t *testing.T) {
	githubWebhookRequest(t, "delete", deleteEventJSON())
}

func TestDeploymentEvent(t *testing.T) {
	githubWebhookRequest(t, "deployment", deploymentEventJSON())
}

func TestDeploymentStatusEvent(t *testing.T) {
	githubWebhookRequest(t, "deployment_status", deploymentStatusEventJSON())
}

func TestForkEvent(t *testing.T) {
	githubWebhookRequest(t, "fork", forkEventJSON())
}

func TestGollumEvent(t *testing.T) {
	githubWebhookRequest(t, "gollum", gollumEventJSON())
}

func TestIssueCommentEvent(t *testing.T) {
	githubWebhookRequest(t, "issue_comment", issueCommentEventJSON())
}

func TestIssuesEvent(t *testing.T) {
	githubWebhookRequest(t, "issues", issuesEventJSON())
}

func TestMemberEvent(t *testing.T) {
	githubWebhookRequest(t, "member", memberEventJSON())
}

func TestMembershipEvent(t *testing.T) {
	githubWebhookRequest(t, "membership", membershipEventJSON())
}

func TestPageBuildEvent(t *testing.T) {
	githubWebhookRequest(t, "page_build", pageBuildEventJSON())
}

func TestPublicEvent(t *testing.T) {
	githubWebhookRequest(t, "public", publicEventJSON())
}

func TestPullRequestReviewCommentEvent(t *testing.T) {
	githubWebhookRequest(t, "pull_request_review_comment", pullRequestReviewCommentEventJSON())
}

func TestPushEvent(t *testing.T) {
	githubWebhookRequest(t, "push", pushEventJSON())
}

func TestReleaseEvent(t *testing.T) {
	githubWebhookRequest(t, "release", releaseEventJSON())
}

func TestRepositoryEvent(t *testing.T) {
	githubWebhookRequest(t, "repository", repositoryEventJSON())
}

func TestStatusEvent(t *testing.T) {
	githubWebhookRequest(t, "status", statusEventJSON())
}

func TestTeamAddEvent(t *testing.T) {
	githubWebhookRequest(t, "team_add", teamAddEventJSON())
}

func TestWatchEvent(t *testing.T) {
	githubWebhookRequest(t, "watch", watchEventJSON())
}

func TestEventWithSignatureFail(t *testing.T) {
	githubWebhookRequestWithSignature(t, "watch", watchEventJSON(), "signature", http.StatusBadRequest)
}

func TestEventWithSignatureSuccess(t *testing.T) {
	githubWebhookRequestWithSignature(t, "watch", watchEventJSON(), generateSignature("signature", []byte(watchEventJSON())), http.StatusOK)
}

func TestWorkflowJob(t *testing.T) {
	acc := githubWebhookRequest(t, "workflow_job", WorkflowJobJSON())

	expected := []telegraf.Metric{
		metric.New(
			"github_webhooks",
			map[string]string{
				"event":      "workflow_job",
				"action":     "completed",
				"repository": "DeusDeorum1/yhome-controller",
				"private":    "true",
				"user":       "DeusDeorum1",
				"admin":      "false",
				"name":       "Run build",
				"conclusion": "success",
			},
			map[string]any{
				"run_attempt": 1,
				"queue_time":  int64(0),
				"run_time":    int64(27000),
				"head_branch": "feature/test",
				"run_id":      int64(12537003369),
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestWorkflowRun(t *testing.T) {
	acc := githubWebhookRequest(t, "workflow_run", WorkflowRunJSON())
	expected := []telegraf.Metric{
		metric.New(
			"github_webhooks",
			map[string]string{
				"event":      "workflow_run",
				"action":     "completed",
				"repository": "DeusDeorum1/yhome-controller",
				"private":    "true",
				"user":       "DeusDeorum1",
				"admin":      "false",
				"name":       ".NET",
				"conclusion": "success",
			},
			map[string]any{
				"run_attempt": 1,
				"run_time":    int64(86000),
				"head_branch": "feature/test",
				"run_id":      int64(12537003369),
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
		

func TestCheckSignatureSuccess(t *testing.T) {
	if !checkSignature("my_little_secret", []byte("random-signature-body"), "sha1=3dca279e731c97c38e3019a075dee9ebbd0a99f0") {
		t.Errorf("check signature failed")
	}
}

func TestCheckSignatureFailed(t *testing.T) {
	if checkSignature("m_little_secret", []byte("random-signature-body"), "sha1=3dca279e731c97c38e3019a075dee9ebbd0a99f0") {
		t.Errorf("check signature failed")
	}
}
