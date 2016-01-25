package github_webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mod "github.com/influxdata/telegraf/plugins/inputs/github_webhooks/models"
)

func TestCommitCommentEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.CommitCommentEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "commit_comment")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestDeleteEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.DeleteEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "delete")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestDeploymentEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.DeploymentEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "deployment")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestDeploymentStatusEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.DeploymentStatusEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "deployment_status")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestForkEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.ForkEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "fork")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestGollumEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.GollumEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "gollum")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestIssueCommentEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.IssueCommentEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "issue_comment")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestIssuesEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.IssuesEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "issues")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestMemberEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.MemberEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "member")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestMembershipEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.MembershipEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "membership")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestPageBuildEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.PageBuildEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "page_build")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestPublicEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.PublicEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "public")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestPullRequestReviewCommentEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.PullRequestReviewCommentEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "pull_request_review_comment")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestPushEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.PushEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "push")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestReleaseEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.ReleaseEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "release")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestRepositoryEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.RepositoryEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "repository")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestStatusEvent(t *testing.T) {
	gh := NewGithubWebhooks()

	jsonString := mod.StatusEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "status")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestTeamAddEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.TeamAddEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "team_add")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestWatchEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := mod.WatchEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "watch")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}
