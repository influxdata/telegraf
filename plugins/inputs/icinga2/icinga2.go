package icinga2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Icinga2 struct {
	Server          string
	ObjectType      string
	Username        string
	Password        string
	ResponseTimeout config.Duration
	tls.ClientConfig

	Log telegraf.Logger

	client *http.Client
}

type Result struct {
	Results []Object `json:"results"`
}

type Object struct {
	Attrs Attribute  `json:"attrs"`
	Name  string     `json:"name"`
	Joins struct{}   `json:"joins"`
	Meta  struct{}   `json:"meta"`
	Type  ObjectType `json:"type"`
}

type Attribute struct {
	CheckCommand string  `json:"check_command"`
	DisplayName  string  `json:"display_name"`
	Name         string  `json:"name"`
	State        float64 `json:"state"`
	HostName     string  `json:"host_name"`
}

var levels = []string{"ok", "warning", "critical", "unknown"}

type ObjectType string

var sampleConfig = `
  ## Required Icinga2 server address
  # server = "https://localhost:5665"

  ## Required Icinga2 object type ("services" or "hosts")
  # object_type = "services"

  ## Credentials for basic HTTP authentication
  # username = "admin"
  # password = "admin"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = true
  `

func (i *Icinga2) Description() string {
	return "Gather Icinga2 status"
}

func (i *Icinga2) SampleConfig() string {
	return sampleConfig
}

func (i *Icinga2) GatherStatus(acc telegraf.Accumulator, checks []Object) {
	for _, check := range checks {
		url, err := url.Parse(i.Server)
		if err != nil {
			i.Log.Error(err.Error())
			continue
		}

		state := int64(check.Attrs.State)

		fields := map[string]interface{}{
			"name":       check.Attrs.Name,
			"state_code": state,
		}

		// source is dependent on 'services' or 'hosts' check
		source := check.Attrs.Name
		if i.ObjectType == "services" {
			source = check.Attrs.HostName
		}

		tags := map[string]string{
			"display_name":  check.Attrs.DisplayName,
			"check_command": check.Attrs.CheckCommand,
			"source":        source,
			"state":         levels[state],
			"server":        url.Hostname(),
			"scheme":        url.Scheme,
			"port":          url.Port(),
		}

		acc.AddFields(fmt.Sprintf("icinga2_%s", i.ObjectType), fields, tags)
	}
}

func (i *Icinga2) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := i.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(i.ResponseTimeout),
	}

	return client, nil
}

func (i *Icinga2) Gather(acc telegraf.Accumulator) error {
	if i.ResponseTimeout < config.Duration(time.Second) {
		i.ResponseTimeout = config.Duration(time.Second * 5)
	}

	if i.client == nil {
		client, err := i.createHTTPClient()
		if err != nil {
			return err
		}
		i.client = client
	}

	requestURL := "%s/v1/objects/%s?attrs=name&attrs=display_name&attrs=state&attrs=check_command"

	// Note: attrs=host_name is only valid for 'services' requests, using check.Attrs.HostName for the host
	//       'hosts' requests will need to use attrs=name only, using check.Attrs.Name for the host
	if i.ObjectType == "services" {
		requestURL += "&attrs=host_name"
	}

	url := fmt.Sprintf(requestURL, i.Server, i.ObjectType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if i.Username != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	result := Result{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	i.GatherStatus(acc, result.Results)

	return nil
}

func init() {
	inputs.Add("icinga2", func() telegraf.Input {
		return &Icinga2{
			Server:          "https://localhost:5665",
			ObjectType:      "services",
			ResponseTimeout: config.Duration(time.Second * 5),
		}
	})
}
