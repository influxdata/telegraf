package geth

import (
	"bytes"
	// "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonParser "github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/tidwall/gjson"
)

// Geth is a geth plugin
type Geth struct {
	Servers  []string
	Metrics []string

	client *http.Client
	lock sync.Mutex
}

type GethError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type GethResponse struct {
	Result *json.RawMessage `json:"result"`
	Error *GethError `json:"error"`
}

var sampleConfig = `
  ## Geth HTTP RPC endpoint
  servers = [
    "http://localhost:8545"
  ]

  ## Each metric in this list is a gjson query path to specify a specific chunk of JSON to be parsed.
  ## gjson query paths are described here: https://github.com/tidwall/gjson#path-syntax
  metrics = [
    "chain",
    "db",
    "discv5",
    "eth",
    "les",
    "p2p",
    "system",
    "trie",
    "txpool"
  ]

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (g *Geth) SampleConfig() string {
	return sampleConfig
}

func (g *Geth) Description() string {
	return "Read flattened metrics from one or more Geth HTTP endpoints"
}

// Gathers data for all servers.
func (g *Geth) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if g.client == nil {
		tlsClientConfig := &tls.ClientConfig{}
		tlsCfg, err := tlsClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}
		g.client = client
	}
	if g.Metrics == nil || len(g.Metrics) == 0 {
		g.Metrics = []string{
			"chain",
			"db",
			"discv5",
			"eth",
			"les",
			"p2p",
			"system",
			"trie",
			"txpool",
		}
	}

	for _, server := range g.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(g.gatherServer(acc, server))
		}(server)
	}

	wg.Wait()

	return nil
}

// Gathers data from a particular server
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (g *Geth) gatherServer(acc telegraf.Accumulator, serverURL string) error {
	respBytes, err := g.sendRequest(serverURL)
	if err != nil {
		return err
	}
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return err
	}
	host, port, _ := net.SplitHostPort(requestURL.Host)
	tags := map[string]string{
		"server": host,
		"port":   port,
	}
	return g.parseJSONMetrics(acc, respBytes, tags)
}

func (g *Geth) parseJSONMetrics(acc telegraf.Accumulator, raw []byte, tags map[string]string) error {
	resp := &GethResponse{}
	if err := json.Unmarshal(raw, resp); err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("error returned from geth http call, code: %d, message: %s",
			resp.Error.Code, resp.Error.Message)
	}
	if resp.Result == nil {
		return fmt.Errorf("geth http response is missing result")
	}
	results := gjson.GetMany(string(*resp.Result), g.Metrics...)
	fields := make(map[string]interface{})
	for i, result := range results {
		if result.Index >= 0 {
			var jsonOut map[string]interface{}
			buf := []byte(result.Raw)
			if err := json.Unmarshal(buf, &jsonOut); err != nil {
				return err
			}
			flattener := &jsonParser.JSONFlattener{}
			if err := flattener.FlattenJSON("", jsonOut); err != nil {
				return err
			}
			for k, v := range flattener.Fields {
				metric := strings.Replace(g.Metrics[i], ".", "_", -1)
				fullName := fmt.Sprintf("%s_%s", metric, strings.ToLower(k))
				fields[fullName] = v

			}
		}
	}
	acc.AddFields("geth", fields, tags)
	return nil
}

// Sends an HTTP request to the server using the Geth object's HTTPClient.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (g *Geth) sendRequest(serverURL string) ([]byte, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	method := "POST"
	content := bytes.NewBufferString(`{"jsonrpc":"2.0","method":"debug_metrics","params":[true],"id":1}`)
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid server URL \"%s\"", serverURL)
	}
	req, err := http.NewRequest(method, requestURL.String(), content)
	if err != nil {
		return nil, err
	}
	// Add header parameters
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestURL.String(),
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return body, err
	}
	return body, err
}

func init() {
	inputs.Add("geth", func() telegraf.Input {
		return &Geth{
			client: &http.Client{},
		}
	})
}
