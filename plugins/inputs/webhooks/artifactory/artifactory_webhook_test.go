package artifactory

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func ArtifactoryWebhookRequest(domain string, event string, jsonString string, t *testing.T) {
	var acc testutil.Accumulator
	awh := &ArtifactoryWebhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, _ := http.NewRequest("POST", "/artifactory", strings.NewReader(jsonString))
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+domain+":"+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func ArtifactoryWebhookRequestWithSignature(event string, jsonString string, t *testing.T, signature string, expectedStatus int) {
	var acc testutil.Accumulator
	awh := &ArtifactoryWebhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, _ := http.NewRequest("POST", "/artifactory", strings.NewReader(jsonString))
	req.Header.Add("x-jfrog-event-auth", signature)
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != expectedStatus {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, expectedStatus)
	}
}

func TestUnsupportedEvent(t *testing.T){
	var acc testutil.Accumulator
	awh := &ArtifactoryWebhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, _ := http.NewRequest("POST", "/artifactory", strings.NewReader(UnsupportedEventJSON()))
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != http.StatusBadRequest{
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusBadRequest)
	}
}

func TestArtifactDeployedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("artifact", "deployed", ArtifactDeployedEventJSON(), t)
}

func TestArtifactDeleted(t *testing.T) {
	ArtifactoryWebhookRequest("artifact", "deleted", ArtifactDeletedEventJSON(), t)
}

func TestArtifactMovedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("artifact", "moved", ArtifactMovedEventJSON(), t)
}

func TestArtifactCopiedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("artifact", "copied", ArtifactCopiedEventJSON(), t)
}

func TestArtifactPropertiesAddedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("artifact_property", "added", ArtifactPropertiesAddedEventJSON(), t)
}

func TestArtifactPropertiesDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("artifact_property", "deleted", ArtifactPropertiesDeletedEventJSON(), t)
}

func TestDockerPushedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("docker", "pushed", DockerPushedEventJSON(), t)
}

func TestDockerDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("docker", "deleted", DockerDeletedEventJSON(), t)
}

func TestDockerPromotedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("docker", "promoted", DockerPromotedEventJSON(), t)
}

func TestBuildUploadedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("build", "uploaded", BuildUploadedEventJSON(), t)
}

func TestBuildDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("build", "deleted", BuildDeletedEventJSON(), t)
}

func TestBuildPromotedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("build", "promoted", BuildPromotedEventJSON(), t)
}

func TestReleaseBundleCreatedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("release_bundle", "created", ReleaseBundleCreatedEventJSON(), t)
}

func TestReleaseBundleSignedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("release_bundle", "signed", ReleaseBundleSignedEventJSON(), t)
}

func TestReleaseBundleDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("release_bundle", "deleted", ReleaseBundleDeletedEventJSON(), t)
}

func TestDistributionStartedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("distribution", "distribute_started", DistributionStartedEventJSON(), t)
}

func TestDistributionCompletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("distribution", "distribute_started", DistributionCompletedEventJSON(), t)
}

func TestDistributionAbortedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("distribution", "distribute_aborted", DistributionAbortedEventJSON(), t)
}

func TestDistributionFailedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("distribution", "distribute_failed", DistributionFailedEventJSON(), t)
}

func TestDestinationReceivedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("destination", "received", DestinationReceivedEventJSON(), t)
}

func TestDestinationDeletedStartedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("destination", "delete_started", DestinationDeleteStartedEventJSON(), t)
}

func TestDestinationDeletedCompletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("destination", "delete_completed", DestinationDeleteCompletedEventJSON(), t)
}

func TestDestinationDeleteFailedEvent(t *testing.T) {
	ArtifactoryWebhookRequest("destination", "delete_failed", DestinationDeleteFailedEventJSON(), t)
}

func TestEventWithSignatureSuccess(t *testing.T) {
	ArtifactoryWebhookRequestWithSignature("watch", ArtifactDeployedEventJSON(), t, generateSignature("signature", []byte(ArtifactDeployedEventJSON())), http.StatusOK)
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