package rollbar

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/auth"
)

type RollbarWebhook struct {
	Path string
	acc  telegraf.Accumulator
	log  telegraf.Logger
	auth.BasicAuth
}

func (rb *RollbarWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(rb.Path, rb.eventHandler).Methods("POST")
	rb.log = log
	rb.log.Infof("Started the webhooks_rollbar on %s", rb.Path)
	rb.acc = acc
}

func (rb *RollbarWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if !rb.Verify(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dummyEvent := &dummyEvent{}
	err = json.Unmarshal(data, dummyEvent)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := NewEvent(dummyEvent, data)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	rb.acc.AddFields("rollbar_webhooks", event.Fields(), event.Tags(), time.Now())

	w.WriteHeader(http.StatusOK)
}

func generateEvent(event event, data []byte) (event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func NewEvent(dummyEvent *dummyEvent, data []byte) (event, error) {
	switch dummyEvent.EventName {
	case "new_item":
		return generateEvent(&newItem{}, data)
	case "occurrence":
		return generateEvent(&occurrence{}, data)
	case "deploy":
		return generateEvent(&deploy{}, data)
	default:
		return nil, errors.New("Not implemented type: " + dummyEvent.EventName)
	}
}
