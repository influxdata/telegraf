package artifactory

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func artifactoryWebhookRequest(t *testing.T, domain, event, jsonString string) {
	var acc testutil.Accumulator
	awh := &Webhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/artifactory", strings.NewReader(jsonString))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+domain+":"+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func artifactoryWebhookRequestWithSignature(t *testing.T, event, jsonString, signature string, expectedStatus int) {
	var acc testutil.Accumulator
	awh := &Webhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
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
	awh := &Webhook{Path: "/artifactory", acc: &acc, log: testutil.Logger{}}
	req, err := http.NewRequest("POST", "/artifactory", strings.NewReader(unsupportedEventJSON()))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	awh.eventHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusBadRequest)
	}
}

func TestArtifactDeployedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "artifact", "deployed", artifactDeployedEventJSON())
}

func TestArtifactDeleted(t *testing.T) {
	artifactoryWebhookRequest(t, "artifact", "deleted", artifactDeletedEventJSON())
}

func TestArtifactMovedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "artifact", "moved", artifactMovedEventJSON())
}

func TestArtifactCopiedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "artifact", "copied", artifactCopiedEventJSON())
}

func TestArtifactPropertiesAddedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "artifact_property", "added", artifactPropertiesAddedEventJSON())
}

func TestArtifactPropertiesDeletedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "artifact_property", "deleted", artifactPropertiesDeletedEventJSON())
}

func TestDockerPushedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "docker", "pushed", dockerPushedEventJSON())
}

func TestDockerDeletedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "docker", "deleted", dockerDeletedEventJSON())
}

func TestDockerPromotedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "docker", "promoted", dockerPromotedEventJSON())
}

func TestBuildUploadedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "build", "uploaded", buildUploadedEventJSON())
}

func TestBuildDeletedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "build", "deleted", buildDeletedEventJSON())
}

func TestBuildPromotedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "build", "promoted", buildPromotedEventJSON())
}

func TestReleaseBundleCreatedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "release_bundle", "created", releaseBundleCreatedEventJSON())
}

func TestReleaseBundleSignedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "release_bundle", "signed", releaseBundleSignedEventJSON())
}

func TestReleaseBundleDeletedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "release_bundle", "deleted", releaseBundleDeletedEventJSON())
}

func TestDistributionStartedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "distribution", "distribute_started", distributionStartedEventJSON())
}

func TestDistributionCompletedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "distribution", "distribute_started", distributionCompletedEventJSON())
}

func TestDistributionAbortedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "distribution", "distribute_aborted", distributionAbortedEventJSON())
}

func TestDistributionFailedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "distribution", "distribute_failed", distributionFailedEventJSON())
}

func TestDestinationReceivedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "destination", "received", destinationReceivedEventJSON())
}

func TestDestinationDeletedStartedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "destination", "delete_started", destinationDeleteStartedEventJSON())
}

func TestDestinationDeletedCompletedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "destination", "delete_completed", destinationDeleteCompletedEventJSON())
}

func TestDestinationDeleteFailedEvent(t *testing.T) {
	artifactoryWebhookRequest(t, "destination", "delete_failed", destinationDeleteFailedEventJSON())
}

func TestEventWithSignatureSuccess(t *testing.T) {
	artifactoryWebhookRequestWithSignature(
		t,
		"watch",
		artifactDeployedEventJSON(),
		generateSignature("signature", []byte(artifactDeployedEventJSON())),
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
