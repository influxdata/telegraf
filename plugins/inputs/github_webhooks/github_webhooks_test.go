package github_webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func GithubWebhookRequest(event string, jsonString string, t *testing.T) {
	gh := NewGithubWebhooks()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", event)
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestCommitCommentEvent(t *testing.T) {
	GithubWebhookRequest("commit_comment", CommitCommentEventJSON(), t)
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
