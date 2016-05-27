package webhooks

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	_ "github.com/influxdata/telegraf/plugins/inputs/webhooks/webhooks_all"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/webhooks_models"
)

func init() {
	inputs.Add("webhooks", func() telegraf.Input { return NewWebhooks() })
}

type Webhooks struct {
	ServiceAddress string

	Webhook []WebhookConfig
}

type WebhookConfig struct {
	Name string
	Path string
}

func NewWebhooks() *Webhooks {
	return &Webhooks{}
}

func (wb *Webhooks) SampleConfig() string {
	return `
  ## Address and port to host Webhook listener on
  service_address = ":1619"
`
}

func (wb *Webhooks) Description() string {
	return "A Webhooks Event collector"
}

func (wb *Webhooks) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (wb *Webhooks) Listen(acc telegraf.Accumulator) {
	r := mux.NewRouter()
	for _, webhook := range wb.Webhook {
		if plugin, ok := webhooks_models.Webhooks[webhook.Name]; ok {
			sub := plugin(webhook.Path)
			sub.Register(r, acc)
		} else {
			log.Printf("Webhook %s is unknow\n", webhook.Name)
		}
	}
	err := http.ListenAndServe(fmt.Sprintf("%s", wb.ServiceAddress), r)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

func (wb *Webhooks) Start(acc telegraf.Accumulator) error {
	go wb.Listen(acc)
	log.Printf("Started the webhooks service on %s\n", wb.ServiceAddress)
	return nil
}

func (rb *Webhooks) Stop() {
	log.Println("Stopping the Webhooks service")
}
