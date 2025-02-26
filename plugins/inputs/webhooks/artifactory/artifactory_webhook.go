package artifactory

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // G505: Blocklisted import crypto/sha1: weak cryptographic primitive - sha1 hash is what is desired in this case
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
)

type Webhook struct {
	Path   string
	Secret string
	acc    telegraf.Accumulator
	log    telegraf.Logger
}

// Register registers the webhook with the provided router
func (awh *Webhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(awh.Path, awh.eventHandler).Methods("POST")

	awh.log = log
	awh.log.Infof("Started webhooks_artifactory on %s", awh.Path)
	awh.acc = acc
}

func (awh *Webhook) eventHandler(rw http.ResponseWriter, r *http.Request) {
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
	ne, err := awh.newEvent(data, et, ed)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	if ne != nil {
		nm := ne.newMetric()
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

func (awh *Webhook) newEvent(data []byte, et, ed string) (event, error) {
	awh.log.Debugf("New %v domain %v event received", ed, et)
	switch ed {
	case "artifact":
		if et == "deployed" || et == "deleted" {
			return generateEvent(data, &artifactDeploymentOrDeletedEvent{})
		}
		if et == "moved" || et == "copied" {
			return generateEvent(data, &artifactMovedOrCopiedEvent{})
		}
		return nil, &newEventError{"Not a recognized event type"}
	case "artifact_property":
		return generateEvent(data, &artifactPropertiesEvent{})
	case "docker":
		return generateEvent(data, &dockerEvent{})
	case "build":
		return generateEvent(data, &buildEvent{})
	case "release_bundle":
		return generateEvent(data, &releaseBundleEvent{})
	case "distribution":
		return generateEvent(data, &distributionEvent{})
	case "destination":
		return generateEvent(data, &destinationEvent{})
	}

	return nil, &newEventError{"Not a recognized event type"}
}

func generateEvent(data []byte, event event) (event, error) {
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
