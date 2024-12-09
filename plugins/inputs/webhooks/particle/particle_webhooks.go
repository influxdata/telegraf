package particle

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/auth"
)

type event struct {
	Name        string `json:"event"`
	Data        data   `json:"data"`
	TTL         int    `json:"ttl"`
	PublishedAt string `json:"published_at"`
	Measurement string `json:"measurement"`
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
	log  telegraf.Logger
	auth.BasicAuth
}

func (rb *ParticleWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(rb.Path, rb.eventHandler).Methods("POST")
	rb.log = log
	rb.log.Infof("Started the webhooks_particle on %s", rb.Path)
	rb.acc = acc
}

func (rb *ParticleWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if !rb.Verify(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	e := newEvent()
	if err := json.NewDecoder(r.Body).Decode(e); err != nil {
		rb.acc.AddError(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pTime, err := e.Time()
	if err != nil {
		pTime = time.Now()
	}

	// Use 'measurement' event field as the measurement, or default to the event name.
	measurementName := e.Measurement
	if measurementName == "" {
		measurementName = e.Name
	}

	rb.acc.AddFields(measurementName, e.Data.Fields, e.Data.Tags, pTime)
	w.WriteHeader(http.StatusOK)
}
