package filestack

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/auth"
)

type Webhook struct {
	Path string
	acc  telegraf.Accumulator
	log  telegraf.Logger
	auth.BasicAuth
}

// Register registers the webhook with the provided router
func (fs *Webhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(fs.Path, fs.eventHandler).Methods("POST")

	fs.log = log
	fs.log.Infof("Started the webhooks_filestack on %s", fs.Path)
	fs.acc = acc
}

func (fs *Webhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if !fs.Verify(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := &filestackEvent{}
	err = json.Unmarshal(body, event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fs.acc.AddFields("filestack_webhooks", event.fields(), event.tags(), time.Unix(event.TimeStamp, 0))

	w.WriteHeader(http.StatusOK)
}
