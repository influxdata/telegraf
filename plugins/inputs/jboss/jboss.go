package jboss

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

type JBoss struct {
	Servers  []string
	Metrics  []string
	Username string
	Password string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

        client HTTPClient
}

var sampleConfig = `
  # Config for get statistics from JBoss AS
  servers = [
    "http://[jboss-server-ip]:9090/management",
  ]
  ## Username and password
  username = ""
  password = ""
  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

// SampleConfig returns a sample configuration block
func (m *JBoss) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the JBoss plugin
func (m *JBoss) Description() string {
	return "Telegraf plugin for gathering metrics from JBoss AS"
}

// Gathers data for all servers.
func (h *JBoss) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if h.client.HTTPClient() == nil {
		tlsCfg, err := internal.GetTLSConfig(
			h.SSLCert, h.SSLKey, h.SSLCA, h.InsecureSkipVerify)
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
		h.client.SetHTTPClient(client)
	}

	errorChannel := make(chan error, len(h.Servers))

	for _, server := range h.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			if err := h.gatherServer(acc, server); err != nil {
				errorChannel <- err
			}
		}(server)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

// Gathers data from a particular server
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     service  : the service being queried
//
// Returns:
//     error: Any error that may have occurred
func (h *GrayLog) gatherServer(
	acc telegraf.Accumulator,
	serverURL string,
) error {
	resp, _, err := h.sendRequest(serverURL)
	if err != nil {
		return err
	}
	requestURL, err := url.Parse(serverURL)
	host, port, _ := net.SplitHostPort(requestURL.Host)
	var dat ResponseMetrics
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp), &dat); err != nil {
		return err
	}
	for _, m_item := range dat.Metrics {
		fields := make(map[string]interface{})
		tags := map[string]string{
			"server": host,
			"port":   port,
			"name":   m_item.Name,
			"type":   m_item.Type,
		}
		h.flatten(m_item.Fields, fields, "")
		acc.AddFields(m_item.FullName, fields, tags)
	}
	return nil
}



// Sends an HTTP request to the server using the GrayLog object's HTTPClient.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *GrayLog) sendRequest(serverURL string) (string, float64, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	method := "GET"
	content := bytes.NewBufferString("")
	headers["Authorization"] = "Basic " + base64.URLEncoding.EncodeToString([]byte(h.Username+":"+h.Password))
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return "", -1, fmt.Errorf("Invalid server URL \"%s\"", serverURL)
	}
	if strings.Contains(requestURL.String(), "multiple") {
		m := &Messagebody{Metrics: h.Metrics}
		http_body, err := json.Marshal(m)
		if err != nil {
			return "", -1, fmt.Errorf("Invalid list of Metrics %s", h.Metrics)
		}
		method = "POST"
		content = bytes.NewBuffer(http_body)
	}
	req, err := http.NewRequest(method, requestURL.String(), content)
	if err != nil {
		return "", -1, err
	}
	// Add header parameters
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	start := time.Now()
	resp, err := h.client.MakeRequest(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	responseTime := time.Since(start).Seconds()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), responseTime, err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestURL.String(),
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(body), responseTime, err
	}
	return string(body), responseTime, err
}


// This should not belong to the object
func (m *JBoss) gatherMainMetrics(a string, defaultPort string, role Role, acc telegraf.Accumulator) error {
	var jsonOut map[string]interface{}

	host, _, err := net.SplitHostPort(a)
	if err != nil {
		host = a
		a = a + defaultPort
	}

	tags := map[string]string{
		"server": host,
		"role":   string(role),
	}

	ts := strconv.Itoa(m.Timeout) + "ms"

	resp, err := client.Get("http://" + a + "/metrics/snapshot?timeout=" + ts)

	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(data), &jsonOut); err != nil {
		return errors.New("Error decoding JSON response")
	}

	m.filterMetrics(role, &jsonOut)

	jf := jsonparser.JSONFlattener{}

	err = jf.FlattenJSON("", jsonOut)

	if err != nil {
		return err
	}

	if role == MASTER {
		if jf.Fields["master/elected"] != 0.0 {
			tags["state"] = "leader"
		} else {
			tags["state"] = "standby"
		}
	}

	acc.AddFields("mesos", jf.Fields, tags)

	return nil
}


var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}


func init() {
	inputs.Add("mesos", func() telegraf.Input {
		return &Mesos{}
	})
}

