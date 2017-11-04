package particle

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type event struct {
	Name        string `json:"event"`
	Data        data   `json:"data"`
	TTL         int    `json:"ttl"`
	PublishedAt string `json:"published_at"`
	Database    string `json:"influx_db"`
}

type data struct {
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"values"`
}

func newEvent() *event {
	return &event{
		Data: data{
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
		},
	}
}

func (e *event) Time() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", e.PublishedAt)
}

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
	e := newEvent()
	if err := json.NewDecoder(r.Body).Decode(e); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pTime, err := e.Time()
	if err != nil {
		pTime = time.Now()
		log.Printf("error parsing particle event time: %s. Using telegraf host time instead: %s", e.PublishedAt, pTime)
	}

	rb.acc.AddFields(e.Name, e.Data.Fields, e.Data.Tags, pTime)
	w.WriteHeader(http.StatusOK)
}
