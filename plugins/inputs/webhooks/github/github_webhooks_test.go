package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func GithubWebhookRequest(event string, jsonString string, t *testing.T) {
	var acc testutil.Accumulator
	gh := &GithubWebhook{Path: "/github", acc: &acc}
	req, _ := http.NewRequest("POST", "/github", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", event)
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func GithubWebhookRequestWithSignature(event string, jsonString string, t *testing.T, signature string, expectedStatus int) {
	var acc testutil.Accumulator
	gh := &GithubWebhook{Path: "/github", Secret: "signature", acc: &acc}
	req, _ := http.NewRequest("POST", "/github", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", event)
	req.Header.Add("X-Hub-Signature", signature)
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != expectedStatus {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, expectedStatus)
	}
}

func TestCommitCommentEvent(t *testing.T) {
	GithubWebhookRequest("commit_comment", CommitCommentEventJSON(), t)
}

func TestPingEvent(t *testing.T) {
	GithubWebhookRequest("ping", "", t)
}

func TestDeleteEvent(t *testing.T) {
	GithubWebhookRequest("delete", DeleteEventJSON(), t)
}

func TestDeploymentEvent(t *testing.T) {
	GithubWebhookRequest("deployment", DeploymentEventJSON(), t)
}

func TestDeploymentStatusEvent(t *testing.T) {
	GithubWebhookRequest("deployment_status", DeploymentStatusEventJSON(), t)
}

func TestForkEvent(t *testing.T) {
	GithubWebhookRequest("fork", ForkEventJSON(), t)
}

func TestGollumEvent(t *testing.T) {
	GithubWebhookRequest("gollum", GollumEventJSON(), t)
}

func TestIssueCommentEvent(t *testing.T) {
	GithubWebhookRequest("issue_comment", IssueCommentEventJSON(), t)
}

func TestIssuesEvent(t *testing.T) {
	GithubWebhookRequest("issues", IssuesEventJSON(), t)
}

func TestMemberEvent(t *testing.T) {
	GithubWebhookRequest("member", MemberEventJSON(), t)
}

func TestMembershipEvent(t *testing.T) {
	GithubWebhookRequest("membership", MembershipEventJSON(), t)
}

func TestPageBuildEvent(t *testing.T) {
	GithubWebhookRequest("page_build", PageBuildEventJSON(), t)
}

func TestPublicEvent(t *testing.T) {
	GithubWebhookRequest("public", PublicEventJSON(), t)
}

func TestPullRequestReviewCommentEvent(t *testing.T) {
	GithubWebhookRequest("pull_request_review_comment", PullRequestReviewCommentEventJSON(), t)
}

func TestPushEvent(t *testing.T) {
	GithubWebhookRequest("push", PushEventJSON(), t)
}

func TestReleaseEvent(t *testing.T) {
	GithubWebhookRequest("release", ReleaseEventJSON(), t)
}

func TestRepositoryEvent(t *testing.T) {
	GithubWebhookRequest("repository", RepositoryEventJSON(), t)
}

func TestStatusEvent(t *testing.T) {
	GithubWebhookRequest("status", StatusEventJSON(), t)
}

func TestTeamAddEvent(t *testing.T) {
	GithubWebhookRequest("team_add", TeamAddEventJSON(), t)
}

func TestWatchEvent(t *testing.T) {
	GithubWebhookRequest("watch", WatchEventJSON(), t)
}

func TestEventWithSignatureFail(t *testing.T) {
	GithubWebhookRequestWithSignature("watch", WatchEventJSON(), t, "signature", http.StatusBadRequest)
}

func TestEventWithSignatureSuccess(t *testing.T) {
	GithubWebhookRequestWithSignature("watch", WatchEventJSON(), t, generateSignature("signature", []byte(WatchEventJSON())), http.StatusOK)
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
