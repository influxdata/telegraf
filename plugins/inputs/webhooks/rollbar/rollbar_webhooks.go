package rollbar

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/webhooks_models"
)

func init() {
	webhooks_models.Add("rollbar", func(path string) webhooks_models.Webhook { return NewRollbarWebhooks(path) })
}

// FIXME: rename
type RollbarWebhooks struct {
	Path string
	acc  telegraf.Accumulator
}

func NewRollbarWebhooks(path string) *RollbarWebhooks {
	return &RollbarWebhooks{Path: path}
}

func (rb *RollbarWebhooks) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(rb.Path, rb.eventHandler).Methods("POST")
	log.Printf("Started the webhooks_rollbar on %s\n", rb.Path)
	rb.acc = acc
}

func (rb *RollbarWebhooks) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dummyEvent := &DummyEvent{}
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

func generateEvent(event Event, data []byte) (Event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func NewEvent(dummyEvent *DummyEvent, data []byte) (Event, error) {
	switch dummyEvent.EventName {
	case "new_item":
		return generateEvent(&NewItem{}, data)
	case "deploy":
		return generateEvent(&Deploy{}, data)
	default:
		return nil, errors.New("Not implemented type: " + dummyEvent.EventName)
	}
}
