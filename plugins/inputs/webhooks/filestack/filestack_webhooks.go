package filestack

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type FilestackWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (fs *FilestackWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(fs.Path, fs.eventHandler).Methods("POST")

	log.Printf("I! Started the webhooks_filestack on %s\n", fs.Path)
	fs.acc = acc
}

func (fs *FilestackWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := &FilestackEvent{}
	err = json.Unmarshal(body, event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fs.acc.AddFields("filestack_webhooks", event.Fields(), event.Tags(), time.Unix(event.TimeStamp, 0))

	w.WriteHeader(http.StatusOK)
}
