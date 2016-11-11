package papertrail

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type PapertrailWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (pt *PapertrailWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(pt.Path, pt.eventHandler).Methods("POST")
	log.Printf("I! Started the papertrail_webhook on %s\n", pt.Path)
	pt.acc = acc
}

func (pt *PapertrailWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
	}

	defer r.Body.Close()
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request", 400)
		return
	}

	data, err := url.QueryUnescape(string(reqBody))
	if err != nil {
		http.Error(w, "Invalid request", 400)
		return
	}

	var payload Payload
	// JSON payload is x-www-form-urlencoded, remove this string when unmarshaling
	remove := "payload="
	if len(data) > 0 && data[0:len(remove)] == remove {
		err = json.Unmarshal([]byte(data[len(remove):len(data)]), &payload)
		if err != nil {
			http.Error(w, "Unable to parse request body", 400)
			return
		}
	} else {
		http.Error(w, "Invalid request", 400)
		return
	}

	if payload.Events != nil {

		// Handle event-based payload
		for _, e := range payload.Events {
			// FIXME: Duplicate event timestamps will overwrite each other
			tags := map[string]string{
				"host":  e.Hostname,
				"event": payload.SavedSearch.Name,
			}
			fields := map[string]interface{}{
				"count": 1,
			}
			pt.acc.AddFields("papertrail", fields, tags, e.ReceivedAt)
		}

	} else if payload.Counts != nil {

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
				pt.acc.AddFields("papertrail", fields, tags, time.Unix(int64(ts), 0))
			}
		}
	} else {
		http.Error(w, "Invalid request", 400)
		return
	}

	w.WriteHeader(http.StatusOK)
}
