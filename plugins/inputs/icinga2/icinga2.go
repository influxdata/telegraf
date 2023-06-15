//go:generate ../../../tools/readme_config_includer/generator
package icinga2

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Icinga2 struct {
	Server          string
	Objects         []string
	Status          []string
	ObjectType      string `toml:"object_type" deprecated:"1.26.0;2.0.0;use 'objects' instead"`
	Username        string
	Password        string
	ResponseTimeout config.Duration
	tls.ClientConfig

	Log telegraf.Logger

	client *http.Client
}

type ResultObject struct {
	Results []struct {
		Attrs struct {
			CheckCommand string  `json:"check_command"`
			DisplayName  string  `json:"display_name"`
			Name         string  `json:"name"`
			State        float64 `json:"state"`
			HostName     string  `json:"host_name"`
		} `json:"attrs"`
		Name  string   `json:"name"`
		Joins struct{} `json:"joins"`
		Meta  struct{} `json:"meta"`
		Type  string   `json:"type"`
	} `json:"results"`
}

type ResultCIB struct {
	Results []struct {
		Status map[string]interface{} `json:"status"`
	} `json:"results"`
}

type ResultPerfdata struct {
	Results []struct {
		Perfdata []struct {
			Label string  `json:"label"`
			Value float64 `json:"value"`
		} `json:"perfdata"`
	} `json:"results"`
}

var levels = []string{"ok", "warning", "critical", "unknown"}

func (*Icinga2) SampleConfig() string {
	return sampleConfig
}

func (i *Icinga2) Init() error {
	statusEndpoints := []string{"ApiListener", "CIB", "IdoMysqlConnection", "IdoPgsqlConnection"}
	if err := choice.CheckSlice(i.Status, statusEndpoints); err != nil {
		return fmt.Errorf("config option 'status': %w", err)
	}

	if i.ResponseTimeout < config.Duration(time.Second) {
		i.ResponseTimeout = config.Duration(time.Second * 5)
	}

	client, err := i.createHTTPClient()
	if err != nil {
		return err
	}
	i.client = client

	// For backward config compatibility
	if i.ObjectType != "" {
		i.Objects = []string{i.ObjectType}
	}

	objectEndpoints := []string{"services", "hosts"}
	if err := choice.CheckSlice(i.Objects, objectEndpoints); err != nil {
		return fmt.Errorf("config option 'objects': %w", err)
	}

	return nil
}

func (i *Icinga2) gatherObjects(acc telegraf.Accumulator, checks ResultObject, objectType string) {
	for _, check := range checks.Results {
		serverURL, err := url.Parse(i.Server)
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
		if objectType == "services" {
			source = check.Attrs.HostName
		}

		tags := map[string]string{
			"display_name":  check.Attrs.DisplayName,
			"check_command": check.Attrs.CheckCommand,
			"source":        source,
			"state":         levels[state],
			"server":        serverURL.Hostname(),
			"scheme":        serverURL.Scheme,
			"port":          serverURL.Port(),
		}

		acc.AddFields(fmt.Sprintf("icinga2_%s", objectType), fields, tags)
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

func (i *Icinga2) icingaRequest(address string) (*http.Response, error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}

	if i.Username != "" {
		req.SetBasicAuth(i.Username, i.Password)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (i *Icinga2) parseObjectResponse(resp *http.Response, result *ResultObject) error {
	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

func (i *Icinga2) parseCIBResponse(resp *http.Response) (map[string]interface{}, error) {
	result := ResultCIB{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if len(result.Results) == 0 {
		return nil, errors.New("no results in Icinga2 API response")
	}

	return result.Results[0].Status, nil
}

func (i *Icinga2) parsePerfdataResponse(resp *http.Response) (map[string]interface{}, error) {
	result := ResultPerfdata{}

	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if len(result.Results) == 0 {
		return nil, errors.New("no results in Icinga2 API response")
	}

	fields := make(map[string]interface{})
	for _, item := range result.Results[0].Perfdata {
		i := strings.Index(item.Label, "-")
		if i > 0 {
			fields[item.Label[i+1:]] = item.Value
		} else {
			fields[item.Label] = item.Value
		}
	}

	return fields, nil
}

func (i *Icinga2) Gather(acc telegraf.Accumulator) error {
	// Collect /v1/objects
	for _, objectType := range i.Objects {
		requestURL := "%s/v1/objects/%s?attrs=name&attrs=display_name&attrs=state&attrs=check_command"

		// Note: attrs=host_name is only valid for 'services' requests, using check.Attrs.HostName for the host
		//       'hosts' requests will need to use attrs=name only, using check.Attrs.Name for the host
		if objectType == "services" {
			requestURL += "&attrs=host_name"
		}

		address := fmt.Sprintf(requestURL, i.Server, objectType)

		resp, err := i.icingaRequest(address)
		if err != nil {
			return err
		}

		result := ResultObject{}
		err = i.parseObjectResponse(resp, &result)
		if err != nil {
			return fmt.Errorf("could not parse object response: %w", err)
		}

		i.gatherObjects(acc, result, objectType)
	}

	// Collect /v1/status
	for _, statusType := range i.Status {
		address := fmt.Sprintf("%s/v1/status/%s", i.Server, statusType)

		resp, err := i.icingaRequest(address)
		if err != nil {
			return err
		}

		tags := map[string]string{
			"component": statusType,
		}
		var fields map[string]interface{}

		switch statusType {
		case "ApiListener":
			fields, err = i.parsePerfdataResponse(resp)
		case "CIB":
			fields, err = i.parseCIBResponse(resp)
		case "IdoMysqlConnection":
			fields, err = i.parsePerfdataResponse(resp)
		case "IdoPgsqlConnection":
			fields, err = i.parsePerfdataResponse(resp)
		}

		if err != nil {
			return fmt.Errorf("could not parse %s response: %w", statusType, err)
		}

		acc.AddFields("icinga2_status", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("icinga2", func() telegraf.Input {
		return &Icinga2{
			Server:          "https://localhost:5665",
			Objects:         []string{"services"},
			ResponseTimeout: config.Duration(time.Second * 5),
		}
	})
}
