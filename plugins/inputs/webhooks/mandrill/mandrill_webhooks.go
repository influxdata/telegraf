package mandrill

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/gorilla/mux"
)

type MandrillWebhook struct {
	Path string
	acc  telegraf.Accumulator
	log  telegraf.Logger
}

func (md *MandrillWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(md.Path, md.returnOK).Methods("HEAD")
	router.HandleFunc(md.Path, md.eventHandler).Methods("POST")

	md.log = log
	md.log.Infof("Started the webhooks_mandrill on %s", md.Path)
	md.acc = acc
}

func (md *MandrillWebhook) returnOK(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (md *MandrillWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data, err := url.ParseQuery(string(body))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var events []MandrillEvent
	err = json.Unmarshal([]byte(data.Get("mandrill_events")), &events)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, event := range events {
		md.acc.AddFields("mandrill_webhooks", event.Fields(), event.Tags(), time.Unix(event.TimeStamp, 0))
	}

	w.WriteHeader(http.StatusOK)
}
