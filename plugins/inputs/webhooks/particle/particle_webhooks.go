package particle

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ParticleWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (rb *ParticleWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(rb.Path, rb.eventHandler).Methods("POST")
	log.Printf("I! Started the webhooks_particle on %s\n", rb.Path)
	rb.acc = acc
}

func (rb *ParticleWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dummy := &DummyData{}
	if err := json.Unmarshal(data, dummy); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pd := &ParticleData{}
	if err := json.Unmarshal([]byte(dummy.Data), pd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pTime, err := dummy.Time()
	if err != nil {
		log.Printf("Time Conversion Error")
		pTime = time.Now()
	}
	rb.acc.AddFields(dummy.InfluxDB, pd.Fields, pd.Tags, pTime)
	w.WriteHeader(http.StatusOK)
}
