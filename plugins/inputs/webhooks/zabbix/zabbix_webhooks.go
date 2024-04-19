package zabbix

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/auth"
)

type ZabbixWebhook struct {
	Path           string
	IgnoreText     bool   `toml:"ignore_text"`
	CreateNameFrom string `toml:"create_name_from_tag"`
	acc            telegraf.Accumulator
	log            telegraf.Logger
	auth.BasicAuth
}

func (zb *ZabbixWebhook) Register(router *mux.Router, acc telegraf.Accumulator, log telegraf.Logger) {
	router.HandleFunc(zb.Path, zb.eventHandler).Methods("POST")
	zb.log = log
	zb.log.Infof("Started the webhooks_zabbix on %s", zb.Path)
	zb.acc = acc
}

func (zb *ZabbixWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	createNameFrom := "component"
	if zb.CreateNameFrom != "" {
		createNameFrom = zb.CreateNameFrom
	}

	if !zb.Verify(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		var err error
		line := []byte(scanner.Text())
		if len(line) <= 10 {
			continue
		}
		var item zabbix_item
		err = json.Unmarshal(line, &item)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("Unmarshal Item faild: %s", err.Error())
			continue
		}
		if !(item.Type == 0 || item.Type == 3) && zb.IgnoreText {
			continue
		}
		zb.acc.AddFields(strings.ToLower("zabbix_"+item.NameFromTag(createNameFrom)), item.Fields(), item.Tags(), item.Time())

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "{\"response\": \"success\"}")
}
