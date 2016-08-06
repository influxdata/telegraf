package particle

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/influxdata/telegraf"
)

type ParticleWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

var decoder = schema.NewDecoder()

func (pwh *ParticleWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(pwh.Path, pwh.eventHandler).Methods("POST")
	log.Printf("Started '%s' on %s\n", meas, pwh.Path)
	pwh.acc = acc
}

func (pwh *ParticleWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e, err := NewEvent(r, &ParticleEvent{})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p := e.NewMetric()
	pwh.acc.AddFields(meas, p.Fields(), p.Tags(), p.Time())

	w.WriteHeader(http.StatusOK)
}

func NewEvent(r *http.Request, event Event) (Event, error) {
	if err := decoder.Decode(event, r.PostForm); err != nil {
		return nil, err
	}
	return event, nil
}
