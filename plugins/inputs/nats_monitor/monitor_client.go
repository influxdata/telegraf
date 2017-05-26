package natsmonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/influxdata/telegraf"
)

type Varz struct {
	CPU           int64 `json:"cpu"`
	Mem           int64 `json:"mem"`
	Subscriptions int64 `json:"subscriptions"`
	Сonnections   int64 `json:"connections"`
	InMsgs        int64 `json:"in_msgs"`
	OutMsgs       int64 `json:"out_msgs"`
	InBytes       int64 `json:"in_bytes"`
	OutBytes      int64 `json:"out_bytes"`
}

type monitorClient struct {
	serverURL  *url.URL
	httpClient *http.Client
	endpoints  map[string]string
}

func (m *monitorClient) gather(acc telegraf.Accumulator) error {

	tags := map[string]string{"url": m.serverURL.String()}

	fields, err := m.varz()
	if err != nil {
		return err
	}
	acc.AddFields("nats_varz", fields, tags)

	return nil
}

func (m *monitorClient) varz() (map[string]interface{}, error) {

	var f Varz
	err := m.get(m.getEndpointUrl("/varz"), &f)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cpu":           f.CPU,
		"mem":           f.Mem,
		"subscriptions": f.Subscriptions,
		"connections":   f.Сonnections,
		"in_msgs":       f.InMsgs,
		"out_msgs":      f.OutMsgs,
		"in_bytes":      f.InBytes,
		"out_bytes":     f.OutBytes,
	}, nil
}

func (m *monitorClient) get(url string, metrics interface{}) error {

	resp, err := m.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", url, resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, metrics)
	if err != nil {
		return err
	}

	return nil
}

func (m *monitorClient) getEndpointUrl(p string) string {

	url, ok := m.endpoints[p]
	if !ok {
		u := *m.serverURL
		u.Path = path.Join(m.serverURL.Path, p)
		url = u.String()
		m.endpoints[p] = url
	}

	return url
}

func NewMonitorClient(srv *url.URL, cln *http.Client) *monitorClient {
	return &monitorClient{
		serverURL:  srv,
		httpClient: cln,
		endpoints:  make(map[string]string),
	}
}
