package artifactory

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type ArtifactoryWebhook struct {
	Path   string
	Secret string
	acc    telegraf.Accumulator
	log    telegraf.Logger
}

func (awh *ArtifactoryWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(awh.Path, awh.eventHandler).Methods("POST")

	awh.log = log
	awh.log.Infof("Started webhooks_artifactory on %s", awh.Path)
	awh.acc = acc
}

func (awh *ArtifactoryWebhook) eventHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if awh.Secret != "" && !checkSignature(awh.Secret, data, r.Header.Get("x-jfrog-event-auth")) {
		awh.log.Error("Failed to check the artifactory webhook auth signature")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyFields := make(map[string]interface{})
	err = json.Unmarshal(data, &bodyFields)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	et := fmt.Sprintf("%v", bodyFields["event_type"])
	ed := fmt.Sprintf("%v", bodyFields["domain"])
	ne, err := awh.NewEvent(data, et, ed)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	if ne != nil {
		nm := ne.NewMetric()
		awh.acc.AddFields("artifactory_webhooks", nm.Fields(), nm.Tags(), nm.Time())
	}

	rw.WriteHeader(http.StatusOK)
}

type newEventError struct {
	s string
}

func (e *newEventError) Error() string {
	return e.s
}

func (awh *ArtifactoryWebhook) NewEvent(data []byte, et string, ed string) (Event, error) {
	awh.log.Debugf("New %v domain %v event received", ed, et)
	switch ed {
	case "artifact":
		if et == "deployed" || et == "deleted" {
			return generateEvent(data, &ArtifactDeploymentOrDeletedEvent{})
		} else if et == "moved" || et == "copied" {
			return generateEvent(data, &ArtifactMovedOrCopiedEvent{})
		} else {
			return nil, &newEventError{"Not a recognized event type"}
		}
	case "artifact_property":
		return generateEvent(data, &ArtifactPropertiesEvent{})
	case "docker":
		return generateEvent(data, &DockerEvent{})
	case "build":
		return generateEvent(data, &BuildEvent{})
	case "release_bundle":
		return generateEvent(data, &ReleaseBundleEvent{})
	case "distribution":
		return generateEvent(data, &DistributionEvent{})
	case "destination":
		return generateEvent(data, &DestinationEvent{})

	}
	return nil, &newEventError{"Not a recognized event type"}

}

func generateEvent(data []byte, event Event) (Event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
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
