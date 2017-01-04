package jboss

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
        GET_HOSTS = 0
		GET_SERVERS = 1
		GET_DB_STAT = 2
)
		
type HostResponse struct {
	Outcome string `json:"outcome"`
	Result []string `json:"result"`
}

type DatasourceResponse struct {
	Outcome string `json:"outcome"`
	Result DatabaseMetrics `json:"result"`
}


type DatabaseMetrics struct {
//	InstalledDrivers interface `json:"installed-drivers"`
	DataSource  map[string]DataSourceMetrics `json:"data-source"`
	XaDataSource  map[string]DataSourceMetrics `json:"xa-data-source"`
}

type DataSourceMetrics struct {
	JndiName string `json:"jndi-name"`
	Statistics  DBStatistics `json:"statistics"`
}

type DBStatistics struct {
	Pool DBPoolStatistics `json:"pool"`
}

type DBPoolStatistics struct {
	ActiveCount string `json:"ActiveCount"`
	AvailableCount string `json:"AvailableCount"`
	InUseCount string `json:"InUseCount"`
	
}


type ResponseMetrics struct {
	outcome string `json:"outcome"`
	Metrics []Metric `json:"result"`
}


type Metric struct {
	FullName string                 `json:"full_name"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Fields   map[string]interface{} `json:"metric"`
}

type JBoss struct {
	Servers  []string
	Metrics  []string
	Username string
	Password string

	ResponseTimeout internal.Duration
	
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

type HTTPClient interface {
	// Returns the result of an http request
	//
	// Parameters:
	// req: HTTP request object
	//
	// Returns:
	// http.Response:  HTTP respons object
	// error        :  Any error that may have occurred
	MakeRequest(req *http.Request) (*http.Response, error)

	SetHTTPClient(client *http.Client)
	HTTPClient() *http.Client
}

type Messagebody struct {
	Metrics []string `json:"metrics"`
}

type RealHTTPClient struct {
	client *http.Client
}

func (c *RealHTTPClient) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *RealHTTPClient) SetHTTPClient(client *http.Client) {
	c.client = client
}

func (c *RealHTTPClient) HTTPClient() *http.Client {
	return c.client
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

func (j *JBoss) doRequest(req *http.Request) ([]byte, error) {
	resp, err := j.client.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			req.RequestURI,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return nil, err
	}

	// Debug response
//	fmt.Printf("%s", body)

	return []byte(body), nil
}

// Gathers data for all servers.
func (h *JBoss) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if h.ResponseTimeout.Duration < time.Second {
		h.ResponseTimeout.Duration = time.Second * 5
	}

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
			Timeout:   h.ResponseTimeout.Duration,
		}
		h.client.SetHTTPClient(client)
	}

	errorChannel := make(chan error, len(h.Servers))

	for _, server := range h.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			req, err := h.prepareRequest(server, GET_HOSTS, nil);
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(req)

			if err != nil {
				fmt.Printf("Error handling response: %s\n", err)
				errorChannel <- err
			} else {
			// Unmarshal json
				hosts := HostResponse{}
				if err = json.Unmarshal(out, &hosts); err != nil {
					errorChannel <- errors.New("Error decoding JSON response")
				}
// 				fmt.Println(hosts)
				oneH := []string{hosts.Result[0],hosts.Result[1]}
				h.getServersOnHost(acc, server, oneH)
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


// Gathers data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getServersOnHost(
	acc telegraf.Accumulator,
	serverURL string,
	hosts []string,
) error {
	var wg sync.WaitGroup

	errorChannel := make(chan error, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			// fmt.Printf("Get Servers from host %s\n", host)
			adr := make(map[string]interface{})
			adr["host"] = host
			//adr := []string{"host=" + host}
			req, err := h.prepareRequest(serverURL, GET_SERVERS, adr);
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(req)

			if err != nil {
				fmt.Printf("Error handling response: %s\n", err)
				errorChannel <- err
			} else {
				servers := HostResponse{}
				if err = json.Unmarshal(out, &servers); err != nil {
					errorChannel <- errors.New("Error decoding JSON response")
				}
//				fmt.Println(servers)
				for _, server := range servers.Result {
					h.getDatasourceStatistics(acc, serverURL, host, server)
				}
			}
		}(host)
	}

	wg.Wait()
	close(errorChannel)
	
	return nil
}

// Gathers data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getDatasourceStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	host string,
	serverName string,
) error {
	//fmt.Printf("getDatasourceStatistics %s %s\n", host, serverName)
	
	adr := make(map[string]interface{})
	adr["host"] = host
	adr["server"] = serverName
	adr["subsystem"] = "datasources"
	 
	req, err := h.prepareRequest(serverURL, GET_DB_STAT, adr);
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(req)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
		server := DatasourceResponse{}
		if err = json.Unmarshal(out, &server); err != nil {
			return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
		}
//		fmt.Println(server)

	
		for database, value := range server.Result.DataSource {
			fields := make(map[string]interface{})
			fields["InUseCount"] = value.Statistics.Pool.InUseCount
			fields["ActiveCount"] = value.Statistics.Pool.ActiveCount
			fields["AvailableCount"] = value.Statistics.Pool.AvailableCount
			tags := map[string]string{
				"host":   host,
				"server": serverName,
				"name":   database,
				"type":   "datasource",
			}
			acc.AddFields("jboss", fields, tags)
		}
		for database, value := range server.Result.XaDataSource {
			fields := make(map[string]interface{})
			fields["InUseCount"] = value.Statistics.Pool.InUseCount
			fields["ActiveCount"] = value.Statistics.Pool.ActiveCount
			fields["AvailableCount"] = value.Statistics.Pool.AvailableCount
			tags := map[string]string{
				"host":   host,
				"server": serverName,
				"name":   database,
				"type":   "datasource",
			}
			acc.AddFields("jboss", fields, tags)
		}
	}
	
	return nil
}

// Gathers data from a particular server
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     service  : the service being queried
//
// Returns:
//     error: Any error that may have occurred
func (h *JBoss) gatherServer(
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

// Flatten JSON hierarchy to produce field name and field value
// Parameters:
//    item: Item map to flatten
//    fields: Map to store generated fields.
//    id: Prefix for top level metric (empty string "")
// Returns:
//    void
func (h *JBoss) flatten(item map[string]interface{}, fields map[string]interface{}, id string) {
	if id != "" {
		id = id + "_"
	}
	for k, i := range item {
		switch i.(type) {
		case int:
			fields[id+k] = i.(float64)
		case float64:
			fields[id+k] = i.(float64)
		case map[string]interface{}:
			h.flatten(i.(map[string]interface{}), fields, id+k)
		default:
		}
	}
}


func (j *JBoss) prepareRequest(domainUrl string, optype int, adress map[string]interface{}) (*http.Request, error) {
	bodyContent := make(map[string]interface{})
	
//	fmt.Printf("url: %s , optype %d, adress %s\n", domainUrl, optype, adress)
	// Create bodyContent
	switch optype {
	case GET_HOSTS:
		bodyContent["operation"] = "read-children-names"
		bodyContent["child-type"] = "host"
		bodyContent["address"] = []string{}
		bodyContent["json.pretty"] = 1
	case GET_SERVERS:
		bodyContent["operation"] = "read-children-names"
		bodyContent["child-type"] = "server"
		bodyContent["recursive-depth"] = 0
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_DB_STAT:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 2
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	}

	serverUrl, err := url.Parse(domainUrl)
	if err != nil {
		return nil, err
	}
	if j.Username != "" || j.Password != "" {
		serverUrl.User = url.UserPassword(j.Username, j.Password)
	}

	
	requestBody, err := json.Marshal(bodyContent)
	// Debug JSON request
	// fmt.Printf("Req: %s\n", requestBody)

	req, err := http.NewRequest("POST", serverUrl.String(), bytes.NewBuffer(requestBody))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-type", "application/json")

	return req, nil
}



// Sends an HTTP request to the server using the GrayLog object's HTTPClient.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *JBoss) sendRequest(serverURL string) (string, float64, error) {
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


func init() {
	inputs.Add("jboss", func() telegraf.Input {
		return &JBoss{
			client: &RealHTTPClient{},
		}
	})
}

