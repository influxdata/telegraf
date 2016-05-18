package rollbar_webhooks

import (
	"encoding/json"
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
  service_address = ":1618"
`
}

func (rb *RollbarWebhooks) Description() string {
	return "A Rollbar Webhook Event collector"
}

func (rb *RollbarWebhooks) Gather(acc telegraf.Accumulator) error {
  rb.Lock()
	defer rb.Unlock()
	for _, event := range rb.events {
    fields := map[string]interface{}{
      "value": 1,
    }
    tags := map[string]string{
      "event": event.Name,
    }

		acc.AddFields("rollbar_webhooks", fields, tags, time.Now())
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

type Event struct {
  Name string `json:"event_name"`
}

// Handles the / route
func (rb *RollbarWebhooks) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

  var e Event
  err = json.Unmarshal(data, &e)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rb.Lock()
	rb.events = append(rb.events, e)
	rb.Unlock()

	w.WriteHeader(http.StatusOK)
}
