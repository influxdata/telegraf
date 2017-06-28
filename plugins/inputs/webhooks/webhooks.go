package webhooks

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/influxdata/telegraf/plugins/inputs/webhooks/filestack"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/github"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/mandrill"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/papertrail"
	"github.com/influxdata/telegraf/plugins/inputs/webhooks/rollbar"
)

type Webhook interface {
	Register(router *mux.Router, acc telegraf.Accumulator)
}

func init() {
	inputs.Add("webhooks", func() telegraf.Input { return NewWebhooks() })
}

type Webhooks struct {
	ServiceAddress string

	Github     *github.GithubWebhook
	Filestack  *filestack.FilestackWebhook
	Mandrill   *mandrill.MandrillWebhook
	Rollbar    *rollbar.RollbarWebhook
	Papertrail *papertrail.PapertrailWebhook
}

func NewWebhooks() *Webhooks {
	return &Webhooks{}
}

func (wb *Webhooks) SampleConfig() string {
	return `
  ## Address and port to host Webhook listener on
  service_address = ":1619"

  [inputs.webhooks.filestack]
    path = "/filestack"

  [inputs.webhooks.github]
    path = "/github"
    # secret = ""

  [inputs.webhooks.mandrill]
    path = "/mandrill"

  [inputs.webhooks.rollbar]
    path = "/rollbar"

  [inputs.webhooks.papertrail]
    path = "/papertrail"
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

	for _, webhook := range wb.AvailableWebhooks() {
		webhook.Register(r, acc)
	}

	err := http.ListenAndServe(fmt.Sprintf("%s", wb.ServiceAddress), r)
	if err != nil {
		acc.AddError(fmt.Errorf("E! Error starting server: %v", err))
	}
}

// Looks for fields which implement Webhook interface
func (wb *Webhooks) AvailableWebhooks() []Webhook {
	webhooks := make([]Webhook, 0)
	s := reflect.ValueOf(wb).Elem()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if !f.CanInterface() {
			continue
		}

		if wbPlugin, ok := f.Interface().(Webhook); ok {
			if !reflect.ValueOf(wbPlugin).IsNil() {
				webhooks = append(webhooks, wbPlugin)
			}
		}
	}

	return webhooks
}

func (wb *Webhooks) Start(acc telegraf.Accumulator) error {
	go wb.Listen(acc)
	log.Printf("I! Started the webhooks service on %s\n", wb.ServiceAddress)
	return nil
}

func (rb *Webhooks) Stop() {
	log.Println("I! Stopping the Webhooks service")
}
