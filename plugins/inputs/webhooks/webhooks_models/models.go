package webhooks_models

import (
	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
)

type Webhook interface {
	Register(router *mux.Router, acc telegraf.Accumulator)
}

var Webhooks map[string]func(string) Webhook = make(map[string]func(string) Webhook)

func Add(name string, fun func(string) Webhook) {
	Webhooks[name] = fun
}
