package papertrail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/auth"
)

type PapertrailWebhook struct {
	Path string
	acc  telegraf.Accumulator
	log  telegraf.Logger
	auth.BasicAuth
}

func (pt *PapertrailWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(pt.Path, pt.eventHandler).Methods("POST")
	pt.log = log
	pt.log.Infof("Started the papertrail_webhook on %s", pt.Path)
	pt.acc = acc
}

func (pt *PapertrailWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	if !pt.Verify(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data := r.PostFormValue("payload")
	if data == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var payload Payload
	err := json.Unmarshal([]byte(data), &payload)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if payload.Events != nil {
		// Handle event-based payload
		for _, e := range payload.Events {
			// Warning: Duplicate event timestamps will overwrite each other
			tags := map[string]string{
				"host":  e.Hostname,
				"event": payload.SavedSearch.Name,
			}
			fields := map[string]interface{}{
				"count":       uint64(1),
				"id":          e.ID,
				"source_ip":   e.SourceIP,
				"source_name": e.SourceName,
				"source_id":   int64(e.SourceID),
				"program":     e.Program,
				"severity":    e.Severity,
				"facility":    e.Facility,
				"message":     e.Message,
				"url":         fmt.Sprintf("%s?centered_on_id=%d", payload.SavedSearch.SearchURL, e.ID),
				"search_id":   payload.SavedSearch.ID,
			}
			pt.acc.AddFields("papertrail", fields, tags, e.ReceivedAt)
		}
	} else if payload.Counts != nil { //nolint:revive // Not simplifying here to stay in the structure for better understanding the code
		// Handle count-based payload
		for _, c := range payload.Counts {
			for ts, count := range *c.TimeSeries {
				tags := map[string]string{
					"host":  c.SourceName,
					"event": payload.SavedSearch.Name,
				}
				fields := map[string]interface{}{
					"count": count,
				}
				pt.acc.AddFields("papertrail", fields, tags, time.Unix(ts, 0))
			}
		}
	} else {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
