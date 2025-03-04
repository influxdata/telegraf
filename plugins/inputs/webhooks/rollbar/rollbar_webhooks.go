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

type Webhook struct {
	Path string
	acc  telegraf.Accumulator
	log  telegraf.Logger
	auth.BasicAuth
}

// Register registers the webhook with the provided router
func (rb *Webhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(rb.Path, rb.eventHandler).Methods("POST")
	rb.log = log
	rb.log.Infof("Started the webhooks_rollbar on %s", rb.Path)
	rb.acc = acc
}

func (rb *Webhook) eventHandler(w http.ResponseWriter, r *http.Request) {
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

	event, err := newEvent(dummyEvent, data)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	rb.acc.AddFields("rollbar_webhooks", event.fields(), event.tags(), time.Now())

	w.WriteHeader(http.StatusOK)
}

func generateEvent(event event, data []byte) (event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func newEvent(dummyEvent *dummyEvent, data []byte) (event, error) {
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
