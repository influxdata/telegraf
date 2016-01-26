package github_webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommitCommentEvent(t *testing.T) {
	gh := NewGithubWebhooks()
	jsonString := CommitCommentEventJSON()
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
	jsonString := DeleteEventJSON()
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
	jsonString := DeploymentEventJSON()
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
	jsonString := DeploymentStatusEventJSON()
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
	jsonString := ForkEventJSON()
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
	jsonString := GollumEventJSON()
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
	jsonString := IssueCommentEventJSON()
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
	jsonString := IssuesEventJSON()
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
	jsonString := MemberEventJSON()
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
	jsonString := MembershipEventJSON()
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
	jsonString := PageBuildEventJSON()
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
	jsonString := PublicEventJSON()
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
	jsonString := PullRequestReviewCommentEventJSON()
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
	jsonString := PushEventJSON()
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
	jsonString := ReleaseEventJSON()
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
	jsonString := RepositoryEventJSON()
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

	jsonString := StatusEventJSON()
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
	jsonString := TeamAddEventJSON()
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
	jsonString := WatchEventJSON()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(jsonString))
	req.Header.Add("X-Github-Event", "watch")
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST commit_comment returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}
