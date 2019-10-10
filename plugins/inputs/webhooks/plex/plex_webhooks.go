package plex

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type PlexWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (p *PlexWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(p.Path, p.eventHandler).Methods("POST")
	log.Printf("I! Started the webhooks_plex on %s\n", p.Path)
	p.acc = acc
}

func (p *PlexWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, err := generateEvent(data, &PlexWebhookEvent{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if e != nil {
		newMetric := e.NewMetric()
		p.acc.AddFields("github_webhooks", newMetric.Fields(), newMetric.Tags(), newMetric.Time())
	}
	w.WriteHeader(http.StatusOK)
}

func generateEvent(data []byte, event Event) (Event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}
