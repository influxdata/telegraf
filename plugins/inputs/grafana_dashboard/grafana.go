package grafana_dashboard

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/grafana-tools/sdk"
	"github.com/influxdata/telegraf"
)

type Grafana struct {
	url    string
	apiKey string
	client *http.Client
	ctx    context.Context
	log    telegraf.Logger
}

func (g *Grafana) httpDoRequest(method, query string, params url.Values, buf io.Reader) ([]byte, int, error) {
	u, _ := url.Parse(g.url)
	u.Path = path.Join(u.Path, query)
	if params != nil {
		u.RawQuery = params.Encode()
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, 0, err
	}
	req = req.WithContext(g.ctx)
	if !strings.Contains(g.apiKey, ":") {
		req.Header.Set("Authorization", "Bearer "+g.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return data, resp.StatusCode, err
}

func (g *Grafana) httpPost(query string, params url.Values, body []byte) ([]byte, int, error) {
	return g.httpDoRequest("POST", query, params, bytes.NewBuffer(body))
}

func (g *Grafana) httpGet(query string, params url.Values) ([]byte, int, error) {
	return g.httpDoRequest("GET", query, params, nil)
}

func (g *Grafana) datasourceJSONValue(ds *sdk.Datasource, key string) string {

	if ds.JSONData == nil {
		return ""
	}

	m, ok := ds.JSONData.(map[string]interface{})
	if ok {
		t, ok := m[key].(string)
		if ok {
			return t
		}
	}
	return ""
}

func (g *Grafana) datasourceProxyIsPost(ds *sdk.Datasource) bool {

	v := g.datasourceJSONValue(ds, "httpMethod")
	return (v == "POST")
}

func (g *Grafana) getData(ds *sdk.Datasource, query string, params url.Values, body []byte) ([]byte, int, error) {

	var (
		raw  []byte
		code int
		err  error
	)

	if g.datasourceProxyIsPost(ds) {
		if raw, code, err = g.httpPost(query, params, body); err != nil {
			return raw, code, err
		}
	} else {
		if raw, code, err = g.httpGet(query, params); err != nil {
			g.log.Error(err)
			return raw, code, err
		}
	}
	return raw, code, err
}

func (g *Grafana) setVariables(vars map[string]string, query string) string {

	s := query
	for k, v := range vars {
		s = strings.ReplaceAll(s, fmt.Sprintf("$%s", k), v)
		s = strings.ReplaceAll(s, fmt.Sprintf("${%s}", k), v)
	}
	return s
}

func NewGrafana(log telegraf.Logger, url, apiKey string, client *http.Client, ctx context.Context) *Grafana {
	return &Grafana{
		log:    log,
		url:    url,
		apiKey: apiKey,
		client: client,
		ctx:    ctx,
	}
}
