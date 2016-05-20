package rollbar_webhooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("rollbar_webhooks", func() telegraf.Input { return NewRollbarWebhooks() })
}

type RollbarWebhooks struct {
	ServiceAddress string
	// Lock for the struct
	sync.Mutex
	// Events buffer to store events between Gather calls
	events []Event
}

func NewRollbarWebhooks() *RollbarWebhooks {
	return &RollbarWebhooks{}
}

func (rb *RollbarWebhooks) SampleConfig() string {
	return `
  ## Address and port to host Webhook listener on
  service_address = ":1619"
`
}

func (rb *RollbarWebhooks) Description() string {
	return "A Rollbar Webhook Event collector"
}

func (rb *RollbarWebhooks) Gather(acc telegraf.Accumulator) error {
	rb.Lock()
	defer rb.Unlock()
	for _, event := range rb.events {
		acc.AddFields("rollbar_webhooks", event.Fields(), event.Tags(), time.Now())
	}
	rb.events = make([]Event, 0)
	return nil
}

func (rb *RollbarWebhooks) Listen() {
	r := mux.NewRouter()
	r.HandleFunc("/", rb.eventHandler).Methods("POST")
	err := http.ListenAndServe(fmt.Sprintf("%s", rb.ServiceAddress), r)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

func (rb *RollbarWebhooks) Start(_ telegraf.Accumulator) error {
	go rb.Listen()
	log.Printf("Started the rollbar_webhooks service on %s\n", rb.ServiceAddress)
	return nil
}

func (rb *RollbarWebhooks) Stop() {
	log.Println("Stopping the rbWebhooks service")
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

	rb.Lock()
	rb.events = append(rb.events, event)
	rb.Unlock()

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
