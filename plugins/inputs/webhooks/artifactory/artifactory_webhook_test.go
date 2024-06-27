package artifactory

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func ArtifactoryWebhookRequest(t *testing.T, domain string, event string, jsonString string) {
	var acc testutil.Accumulator
	awh := &ArtifactoryWebhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/artifactory", strings.NewReader(jsonString))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+domain+":"+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func ArtifactoryWebhookRequestWithSignature(event string, jsonString string, t *testing.T, signature string, expectedStatus int) {
	var acc testutil.Accumulator
	awh := &ArtifactoryWebhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/artifactory", strings.NewReader(jsonString))
	require.NoError(t, err)
	req.Header.Add("x-jfrog-event-auth", signature)
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != expectedStatus {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, expectedStatus)
	}
}

func TestUnsupportedEvent(t *testing.T) {
	var acc testutil.Accumulator
	awh := &ArtifactoryWebhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/artifactory", strings.NewReader(UnsupportedEventJSON()))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusBadRequest)
	}
}

func TestArtifactDeployedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "artifact", "deployed", ArtifactDeployedEventJSON())
}

func TestArtifactDeleted(t *testing.T) {
	ArtifactoryWebhookRequest(t, "artifact", "deleted", ArtifactDeletedEventJSON())
}

func TestArtifactMovedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "artifact", "moved", ArtifactMovedEventJSON())
}

func TestArtifactCopiedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "artifact", "copied", ArtifactCopiedEventJSON())
}

func TestArtifactPropertiesAddedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "artifact_property", "added", ArtifactPropertiesAddedEventJSON())
}

func TestArtifactPropertiesDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "artifact_property", "deleted", ArtifactPropertiesDeletedEventJSON())
}

func TestDockerPushedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "docker", "pushed", DockerPushedEventJSON())
}

func TestDockerDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "docker", "deleted", DockerDeletedEventJSON())
}

func TestDockerPromotedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "docker", "promoted", DockerPromotedEventJSON())
}

func TestBuildUploadedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "build", "uploaded", BuildUploadedEventJSON())
}

func TestBuildDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "build", "deleted", BuildDeletedEventJSON())
}

func TestBuildPromotedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "build", "promoted", BuildPromotedEventJSON())
}

func TestReleaseBundleCreatedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "release_bundle", "created", ReleaseBundleCreatedEventJSON())
}

func TestReleaseBundleSignedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "release_bundle", "signed", ReleaseBundleSignedEventJSON())
}

func TestReleaseBundleDeletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "release_bundle", "deleted", ReleaseBundleDeletedEventJSON())
}

func TestDistributionStartedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "distribution", "distribute_started", DistributionStartedEventJSON())
}

func TestDistributionCompletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "distribution", "distribute_started", DistributionCompletedEventJSON())
}

func TestDistributionAbortedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "distribution", "distribute_aborted", DistributionAbortedEventJSON())
}

func TestDistributionFailedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "distribution", "distribute_failed", DistributionFailedEventJSON())
}

func TestDestinationReceivedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "destination", "received", DestinationReceivedEventJSON())
}

func TestDestinationDeletedStartedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "destination", "delete_started", DestinationDeleteStartedEventJSON())
}

func TestDestinationDeletedCompletedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "destination", "delete_completed", DestinationDeleteCompletedEventJSON())
}

func TestDestinationDeleteFailedEvent(t *testing.T) {
	ArtifactoryWebhookRequest(t, "destination", "delete_failed", DestinationDeleteFailedEventJSON())
}

func TestEventWithSignatureSuccess(t *testing.T) {
	ArtifactoryWebhookRequestWithSignature(
		"watch",
		ArtifactDeployedEventJSON(),
		t,
		generateSignature("signature", []byte(ArtifactDeployedEventJSON())),
		http.StatusOK,
	)
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
