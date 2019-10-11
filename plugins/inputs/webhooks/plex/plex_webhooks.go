package plex

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type PlexWebhook struct {
	Path string
	acc  telegraf.Accumulator

	Log telegraf.Logger
}

func (p *PlexWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(p.Path, p.eventHandler).Methods("POST")
	p.Log.Info("I! Started the webhooks_plex on %s\n", p.Path)
	p.acc = acc
}

func (p *PlexWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Form["payload"] == nil || len(r.Form["payload"]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e := &PlexWebhookEvent{}
	err = generateEvent([]byte(r.Form["payload"][0]), e)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if e != nil {
		newMetric := e.NewMetric()
		p.acc.AddFields("plex_webhooks", newMetric.Fields(), newMetric.Tags(), newMetric.Time())
	}
	w.WriteHeader(http.StatusOK)
}

func generateEvent(data []byte, event Event) error {
	err := json.Unmarshal(data, event)
	return err
}
