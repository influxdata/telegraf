package dockerhub

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type DockerhubWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (dhwh *DockerhubWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(dhwh.Path, dhwh.eventHandler).Methods("POST")
	log.Printf("Started '%s' on %s\n", meas, dhwh.Path)
	dhwh.acc = acc
}

func (dhwh *DockerhubWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, err := NewEvent(data, &DockerhubEvent{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dh := e.NewMetric()
	dhwh.acc.AddFields(meas, dh.Fields(), dh.Tags(), dh.Time())
	w.WriteHeader(http.StatusOK)
}

func NewEvent(data []byte, event Event) (Event, error) {
	err := json.Unmarshal(data, event)
	if err != nil {
		return nil, err
	}
	return event, nil
}
