package jboss

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// GetHosts constan applied to jboss management query types
const (
	GetHosts          = 0
	GetServers        = 1
	GetDBStat         = 2
	GetJVMStat        = 3
	GetDeployments    = 4
	GetDeploymentStat = 5
	GetWebStat        = 6
	GetJMSQueueStat   = 7
	GetJMSTopicStat   = 8
)

// KeyVal key / value struct
type KeyVal struct {
	Key string
	Val interface{}
}

// OrderedMap Define an ordered map
type OrderedMap []KeyVal

// MarshalJSON Implement the json.Marshaler interface
func (omap OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{")
	for i, kv := range omap {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv.Val)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

// HostResponse expected GetHost response type
type HostResponse struct {
	Outcome string   `json:"outcome"`
	Result  []string `json:"result"`
}

// DatasourceResponse expected GetDBStat response type
type DatasourceResponse struct {
	Outcome string          `json:"outcome"`
	Result  DatabaseMetrics `json:"result"`
}

// JMSResponse expected GetJMSTopicStat/GetJMSQueueStat response type
type JMSResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

// DatabaseMetrics database related metrics
type DatabaseMetrics struct {
	DataSource   map[string]DataSourceMetrics `json:"data-source"`
	XaDataSource map[string]DataSourceMetrics `json:"xa-data-source"`
}

//DataSourceMetrics Datasource related metrics
type DataSourceMetrics struct {
	JndiName   string       `json:"jndi-name"`
	Statistics DBStatistics `json:"statistics"`
}

// DBStatistics DB statistics per pool
type DBStatistics struct {
	Pool DBPoolStatistics `json:"pool"`
}

// DBPoolStatistics pool related statistics
type DBPoolStatistics struct {
	ActiveCount    string `json:"ActiveCount"`
	AvailableCount string `json:"AvailableCount"`
	InUseCount     string `json:"InUseCount"`
}

// JVMResponse GetJVMStat expected response type
type JVMResponse struct {
	Outcome string     `json:"outcome"`
	Result  JVMMetrics `json:"result"`
}

// JVMMetrics JVM related metrics type
type JVMMetrics struct {
	Type map[string]interface{} `json:"type"`
}

// WebResponse getWebStatistics expected response type
type WebResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

// DeploymentResponse GetDeployments expected response type
type DeploymentResponse struct {
	Outcome string            `json:"outcome"`
	Result  DeploymentMetrics `json:"result"`
}

// DeploymentMetrics deployment related type
type DeploymentMetrics struct {
	Name          string                 `json:"name"`
	RuntimeName   string                 `json:"runtime-name"`
	Status        string                 `json:"status"`
	Subdeployment map[string]interface{} `json:"subdeployment"`
	Subsystem     map[string]interface{} `json:"subsystem"`
}

// WebMetrics  Web Modules related metrics
type WebMetrics struct {
	ActiveSessions    string                 `json:"active-sessions"`
	ContextRoot       string                 `json:"context-root"`
	ExpiredSessions   string                 `json:"expired-sessions"`
	MaxActiveSessions string                 `json:"max-active-sessions"`
	SessionsCreated   string                 `json:"sessions-created"`
	Servlet           map[string]interface{} `json:"servlet"`
}

// JBoss the main collectod struct
type JBoss struct {
	Servers []string
	Metrics []string

	Username string
	Password string

	ExecAsDomain bool `toml:"exec_as_domain"`

	Authorization string

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

// HTTPClient HTTP client struct
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

// RealHTTPClient the HTTP client handler
type RealHTTPClient struct {
	client *http.Client
}

// MakeRequest do an HTTP request
func (c *RealHTTPClient) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// SetHTTPClient set http client
func (c *RealHTTPClient) SetHTTPClient(client *http.Client) {
	c.client = client
}

// HTTPClient return the HTTP client
func (c *RealHTTPClient) HTTPClient() *http.Client {
	return c.client
}

var sampleConfig = `
  # Config for get statistics from JBoss AS
  servers = [
    "http://[jboss-server-ip]:9090/management",
  ]
	## Execution Mode
	exec_as_domain = false
  ## Username and password
  username = ""
  password = ""
	## authorization mode could be "basic" or "digest"
  authorization = "digest"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
	## Metric selection
	metrics =[
		"jvm",
		"web_con",
		"deployment",
		"database",
		"jms",
	]
`

// SampleConfig returns a sample configuration block
func (*JBoss) SampleConfig() string {
	return sampleConfig
}

// Description just returns a short description of the JBoss plugin
func (*JBoss) Description() string {
	return "Telegraf plugin for gathering metrics from JBoss AS"
}

func (h *JBoss) checkAuth(host string, uri string) error {
	url := h.Servers[0]

	method := "POST"
	req, err := http.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		log.Printf("Recieved status code '%v' auth skipped\n", resp.StatusCode)
		return nil
	}
	digestParts := digestParts(resp)
	digestParts["uri"] = uri
	digestParts["method"] = method
	digestParts["username"] = h.Username
	digestParts["password"] = h.Password
	postData := []byte("{\"address\":[\"\"],\"child-type\":\"host\",\"json.pretty\":1,\"operation\":\"read-children-names\"}")
	req, err = http.NewRequest(method, url, bytes.NewBuffer(postData))
	h.Authorization = getDigestAuthrization2(digestParts)
	req.Header.Set("Authorization", h.Authorization)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	log.Printf("D! JBoss HTTP response code: %d \n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		log.Printf("D! JBoss HTTP response body: %s \n", string(body))
		return nil
	}
	return nil
}

func digestParts(resp *http.Response) map[string]string {
	result := map[string]string{}
	if len(resp.Header["Www-Authenticate"]) > 0 {
		//        wantedHeaders := []string{"nonce", "realm", "qop"}
		wantedHeaders := []string{"nonce", "realm"}
		responseHeaders := strings.Split(resp.Header["Www-Authenticate"][0], ",")
		for _, r := range responseHeaders {
			for _, w := range wantedHeaders {
				if strings.Contains(r, w) {
					result[w] = strings.Split(r, `"`)[1]
				}
			}
		}
	}
	return result
}

func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getCnonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x", b)[:16]
}

func getDigestAuthrization(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	nonceCount := 00000001
	cnonce := getCnonce()
	response := getMD5(fmt.Sprintf("%s:%s:%v:%s:%s:%s", ha1, d["nonce"], nonceCount, cnonce, d["qop"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", cnonce="%s", nc="%v", qop="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], cnonce, nonceCount, d["qop"], response)
	return authorization
}

func getDigestAuthrization2(digestParts map[string]string) string {
	d := digestParts
	ha1 := getMD5(d["username"] + ":" + d["realm"] + ":" + d["password"])
	ha2 := getMD5(d["method"] + ":" + d["uri"])
	response := getMD5(fmt.Sprintf("%s:%s:%s", ha1, d["nonce"], ha2))
	authorization := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		d["username"], d["realm"], d["nonce"], d["uri"], response)
	return authorization
}

// Gather Gathers data for all servers.
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
			//default as standalone server
			hosts := HostResponse{Outcome: "", Result: []string{"standalone"}}
			log.Printf("I! JBoss Plugin Working as Domain: %t\n", h.ExecAsDomain)
			if h.ExecAsDomain {
				bodyContent, err := h.prepareRequest(GetHosts, nil)
				if err != nil {
					errorChannel <- err
				}

				out, err := h.doRequest(server, bodyContent)

				log.Printf("D! JBoss API Req err: %s", err)
				log.Printf("D! JBoss API Req out: %s", out)

				if err != nil {
					log.Printf("E! JBoss Error handling response 1: %s\n", err)
					log.Printf("E! JBoss server:%s bodyContent %s\n", server, bodyContent)
					errorChannel <- err
					return
				}
				// Unmarshal json

				if err = json.Unmarshal(out, &hosts); err != nil {
					errorChannel <- errors.New("Error decoding JSON response")
				}
				log.Printf("D! JBoss HOSTS %s", hosts)
			}

			h.getServersOnHost(acc, server, hosts.Result)

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
			log.Printf("I! Get Servers from host: %s\n", host)

			servers := HostResponse{Outcome: "", Result: []string{"standalone"}}

			if h.ExecAsDomain {
				//get servers
				adr := OrderedMap{
					{"host", host},
				}

				bodyContent, err := h.prepareRequest(GetServers, adr)
				if err != nil {
					errorChannel <- err
				}

				out, err := h.doRequest(serverURL, bodyContent)

				log.Printf("D! JBoss API Req err: %s", err)
				log.Printf("D! JBoss API Req out: %s", out)

				if err != nil {
					log.Printf("E! JBoss Error handling response 2: %s\n", err)
					errorChannel <- err
					return
				}

				if err = json.Unmarshal(out, &servers); err != nil {
					errorChannel <- errors.New("Error decoding JSON response")
				}
			}

			for _, server := range servers.Result {
				log.Printf("I! JBoss Plugin Processing Servers from host:[ %s ] : Server [ %s ]\n", host, server)
				for _, v := range h.Metrics {
					switch v {
					case "jvm":
						h.getJVMStatistics(acc, serverURL, host, server)
					case "web_con":
						h.getWebStatistics(acc, serverURL, host, server, "ajp")
						h.getWebStatistics(acc, serverURL, host, server, "http")
					case "deployment":
						h.getServerDeploymentStatistics(acc, serverURL, host, server)
					case "database":
						h.getDatasourceStatistics(acc, serverURL, host, server)
					case "jms":
						h.getJMSStatistics(acc, serverURL, host, server, GetJMSQueueStat)
						h.getJMSStatistics(acc, serverURL, host, server, GetJMSTopicStat)
					default:
						log.Printf("E! Jboss doesn't exist the metric set %s\n", v)
					}
				}
			}
		}(host)
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

// Gathers web data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getWebStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	host string,
	serverName string,
	connector string,
) error {
	adr := OrderedMap{}
	if h.ExecAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"subsystem", "web"},
			{"connector", connector},
		}
	} else {
		adr = OrderedMap{
			{"subsystem", "web"},
			{"connector", connector},
		}
	}

	bodyContent, err := h.prepareRequest(GetWebStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}
	server := WebResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}

	fields := make(map[string]interface{})
	for key, value := range server.Result {
		switch key {
		case "bytesReceived", "bytesSent", "requestCount", "errorCount", "maxTime", "processingTime":
			if value != nil {
				switch value.(type) {
				case int:
					fields[key] = value.(float64)
				case float64:
					fields[key] = value.(float64)
				case string:
					f, err := strconv.ParseFloat(value.(string), 64)
					if err != nil {
						log.Printf("E! JBoss Error decoding Float  from string : %s = %s\n", key, value.(string))
					} else {
						fields[key] = f
					}
				}
			}
		}
	}
	tags := map[string]string{
		"jboss_host":   host,
		"jboss_server": serverName,
		"type":         connector,
	}
	acc.AddFields("jboss_web_con", fields, tags)

	return nil
}

// Gathers database data from a particular host
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
	adr := OrderedMap{}
	if h.ExecAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"subsystem", "datasources"},
		}
	} else {
		adr = OrderedMap{
			{"subsystem", "datasources"},
		}
	}

	bodyContent, err := h.prepareRequest(GetDBStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}
	server := DatasourceResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}

	for database, value := range server.Result.DataSource {
		fields := make(map[string]interface{})
		fields["in-use-count"], _ = strconv.ParseInt(value.Statistics.Pool.InUseCount, 10, 64)
		fields["active-count"], _ = strconv.ParseInt(value.Statistics.Pool.ActiveCount, 10, 64)
		fields["available-count"], _ = strconv.ParseInt(value.Statistics.Pool.AvailableCount, 10, 64)
		tags := map[string]string{
			"jboss_host":   host,
			"jboss_server": serverName,
			"name":         database,
		}
		acc.AddFields("jboss_database", fields, tags)
	}
	for database, value := range server.Result.XaDataSource {
		fields := make(map[string]interface{})
		fields["in-use-count"], _ = strconv.ParseInt(value.Statistics.Pool.InUseCount, 10, 64)
		fields["active-count"], _ = strconv.ParseInt(value.Statistics.Pool.ActiveCount, 10, 64)
		fields["available-count"], _ = strconv.ParseInt(value.Statistics.Pool.AvailableCount, 10, 64)
		tags := map[string]string{
			"jboss_host":   host,
			"jboss_server": serverName,
			"name":         database,
		}
		acc.AddFields("jboss_database", fields, tags)
	}

	return nil
}

// Gathers JMS data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getJMSStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	host string,
	serverName string,
	opType int,
) error {

	adr := OrderedMap{}

	if h.ExecAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"subsystem", "messaging"},
			{"hornetq-server", "default"},
		}
	} else {
		adr = OrderedMap{
			{"subsystem", "messaging"},
			{"hornetq-server", "default"},
		}
	}

	bodyContent, err := h.prepareRequest(opType, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}
	jmsresponse := JMSResponse{}
	if err = json.Unmarshal(out, &jmsresponse); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}

	for jmsQueue, value := range jmsresponse.Result {
		fields := make(map[string]interface{})
		v := value.(map[string]interface{})
		fields["message-count"] = v["message-count"]
		fields["messages-added"] = v["messages-added"]
		if opType == GetJMSQueueStat {
			fields["consumer-count"] = v["consumer-count"]
		} else {
			fields["subscription-count"] = v["subscription-count"]
		}
		fields["scheduled-count"] = v["scheduled-count"]
		tags := map[string]string{
			"jboss_host":   host,
			"jboss_server": serverName,
			"name":         jmsQueue,
		}
		acc.AddFields("jboss_jms", fields, tags)
	}

	return nil
}

// Gathers JVM data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getJVMStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	host string,
	serverName string,
) error {
	adr := OrderedMap{}
	if h.ExecAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"core-service", "platform-mbean"},
		}
	} else {
		adr = OrderedMap{
			{"core-service", "platform-mbean"},
		}
	}

	bodyContent, err := h.prepareRequest(GetJVMStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	server := JVMResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}

	fields := make(map[string]interface{})

	for typeName, value := range server.Result.Type {

		switch typeName {
		case "threading":
			t := value.(map[string]interface{})
			fields["thread-count"] = t["thread-count"]
			fields["peak-thread-count"] = t["peak-thread-count"]
			fields["daemon-thread-count"] = t["daemon-thread-count"]
		case "memory":
			mem := value.(map[string]interface{})
			heap := mem["heap-memory-usage"].(map[string]interface{})
			nonHeap := mem["non-heap-memory-usage"].(map[string]interface{})
			h.flatten(heap, fields, "heap")
			h.flatten(nonHeap, fields, "nonheap")
		case "garbage-collector":
			gc := value.(map[string]interface{})
			gcName := gc["name"].(map[string]interface{})
			for gcType, gcVal := range gcName {
				object := gcVal.(map[string]interface{})
				fields[gcType+"_count"] = object["collection-count"]
				fields[gcType+"_time"] = object["collection-time"]
			}
		}
	}
	tags := map[string]string{
		"jboss_host":   host,
		"jboss_server": serverName,
	}
	acc.AddFields("jboss_jvm", fields, tags)
	return nil
}

// Gathers Deployment data from a particular host and server
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) processEJBAppStats(acc telegraf.Accumulator, ejb map[string]interface{}, tags map[string]string) error {
	fields := make(map[string]interface{})
	t := ejb["stateless-session-bean"]
	if t != nil {
		statelessList := t.(map[string]interface{})

		for stateless, ejbVal := range statelessList {
			ejbRuntime := ejbVal.(map[string]interface{})
			fields["invocations"] = ejbRuntime["invocations"]
			fields["peak-concurrent-invocations"] = ejbRuntime["peak-concurrent-invocations"]
			fields["pool-available-count"] = ejbRuntime["pool-available-count"]
			fields["pool-create-count"] = ejbRuntime["pool-create-count"]
			fields["pool-current-size"] = ejbRuntime["pool-current-size"]
			fields["pool-max-size"] = ejbRuntime["pool-max-size"]
			fields["pool-remove-count"] = ejbRuntime["pool-remove-count"]
			fields["wait-time"] = ejbRuntime["wait-time"]
			tags["ejb"] = stateless
			acc.AddFields("jboss_ejb", fields, tags)
		}
	}
	return nil
}

func (h *JBoss) processWebAppStats(acc telegraf.Accumulator, web map[string]interface{}, tags map[string]string) error {
	fields := make(map[string]interface{})
	contextRoot := web["context-root"].(string)
	fields["active-sessions"] = web["active-sessions"]
	fields["expired-sessions"] = web["expired-sessions"]
	fields["max-active-sessions"] = web["max-active-sessions"]
	fields["sessions-created"] = web["sessions-created"]
	tags["context-root"] = contextRoot
	acc.AddFields("jboss_web_app", fields, tags)
	return nil
}

// Gathers Deployment data from a particular host and server
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getServerDeploymentStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	host string,
	serverName string,
) error {
	var wg sync.WaitGroup
	adr := OrderedMap{}

	if h.ExecAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
		}
	}

	bodyContent, err := h.prepareRequest(GetDeployments, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	deployments := HostResponse{}
	if err = json.Unmarshal(out, &deployments); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}

	errorChannel := make(chan error, len(deployments.Result))

	for _, deployment := range deployments.Result {
		wg.Add(1)
		go func(deployment string) {
			defer wg.Done()
			adr2 := OrderedMap{}
			if h.ExecAsDomain {
				adr2 = OrderedMap{
					{"host", host},
					{"server", serverName},
					{"deployment", deployment},
				}
			} else {
				adr2 = OrderedMap{
					{"deployment", deployment},
				}
			}

			bodyContent, err := h.prepareRequest(GetDeploymentStat, adr2)
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(serverURL, bodyContent)

			log.Printf("D! JBoss Deployment API Req err: %s", err)
			log.Printf("D! JBoss Deployment API Req out: %s", out)

			if err != nil {
				log.Printf("E! JBoss Deployment Error handling response 3: %s\n", err)
				errorChannel <- err
				return
			}
			// everything ok ! continue with decoding data
			deploy := DeploymentResponse{}
			if err = json.Unmarshal(out, &deploy); err != nil {
				errorChannel <- errors.New("Error decoding JSON response")
			}
			// This struct apply on EAR files
			for typeName, value := range deploy.Result.Subdeployment {
				if value == nil {
					log.Printf("D! JBoss Deployment WARNING Subdeployment value is NULL")
					continue
				}

				t := value.(map[string]interface{})
				if t["subsystem"] == nil {
					log.Printf("D! JBoss Deployment WARNING SUBDEPLOYMENT Subsystem is NULL")
					continue
				}
				subsystem := t["subsystem"].(map[string]interface{})

				if ejbValue, ok := subsystem["ejb3"]; ok {
					ejb := ejbValue.(map[string]interface{})
					tags := map[string]string{
						"jboss_host":   host,
						"jboss_server": serverName,
						"name":         typeName,
						"runtime_name": deploy.Result.RuntimeName,
					}
					h.processEJBAppStats(acc, ejb, tags)
				}

				if webValue, ok := subsystem["web"]; ok {
					web := webValue.(map[string]interface{})
					tags := map[string]string{
						"jboss_host":   host,
						"jboss_server": serverName,
						"name":         typeName,
						"runtime_name": deploy.Result.RuntimeName,
					}
					h.processWebAppStats(acc, web, tags)
				}
			}
			// This struct apply on WAR files
			for typeName, value := range deploy.Result.Subsystem {
				if value == nil {
					log.Printf("D! JBoss Deployment SUBSYSTEM  value NULL")
					continue
				}
				if typeName == "web" {
					web := value.(map[string]interface{})
					tags := map[string]string{
						"jboss_host":   host,
						"jboss_server": serverName,
						"name":         deploy.Result.Name,
						"runtime_name": deploy.Result.RuntimeName,
					}
					h.processWebAppStats(acc, web, tags)
				} else {
					log.Printf("W! JBoss Deployment WAR  from type %s", typeName)
				}
			}

		}(deployment)
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

//func (j *JBoss) prepareRequest(optype int, adress map[string]interface{}) (map[string]interface{}, error) {
func (h *JBoss) prepareRequest(optype int, adress OrderedMap) (map[string]interface{}, error) {
	bodyContent := make(map[string]interface{})

	// Create bodyContent
	switch optype {
	case GetHosts:
		bodyContent["operation"] = "read-children-names"
		bodyContent["child-type"] = "host"
		bodyContent["address"] = []string{}
		bodyContent["json.pretty"] = 1
	case GetServers:
		bodyContent["operation"] = "read-children-names"
		bodyContent["child-type"] = "server"
		bodyContent["recursive-depth"] = 0
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetDBStat:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 2
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetJVMStat:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive"] = "true"
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetDeployments:
		bodyContent["operation"] = "read-children-names"
		bodyContent["child-type"] = "deployment"
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetDeploymentStat:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 3
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetWebStat:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 0
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetJMSQueueStat:
		bodyContent["operation"] = "read-children-resources"
		bodyContent["child-type"] = "jms-queue"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 2
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GetJMSTopicStat:
		bodyContent["operation"] = "read-children-resources"
		bodyContent["child-type"] = "jms-topic"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 2
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	}

	return bodyContent, nil
}

func (j *JBoss) doRequest(domainUrl string, bodyContent map[string]interface{}) ([]byte, error) {

	serverUrl, err := url.Parse(domainUrl)
	if err != nil {
		return nil, err
	}
	requestBody, err := json.Marshal(bodyContent)
	method := "POST"

	// Debug JSON request
	log.Printf("D! Req: %s\n", requestBody)

	req, err := http.NewRequest(method, serverUrl.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	if j.Authorization == "basic" {
		if j.Username != "" || j.Password != "" {
			serverUrl.User = url.UserPassword(j.Username, j.Password)
		}
	}

	resp, err := j.client.MakeRequest(req)
	if err != nil {
		log.Printf("D! HTTP REQ:%#+v", req)
		log.Printf("D! HTTP RESP:%#+v", resp)
		return nil, err
	}
	defer resp.Body.Close()

	// Process response

	if resp.StatusCode == http.StatusUnauthorized {
		digestParts := digestParts(resp)
		digestParts["uri"] = serverUrl.RequestURI()
		digestParts["method"] = method
		digestParts["username"] = j.Username
		digestParts["password"] = j.Password

		req, err = http.NewRequest(method, serverUrl.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		j.Authorization = getDigestAuthrization2(digestParts)
		req.Header.Set("Authorization", j.Authorization)
		req.Header.Set("Content-Type", "application/json")

		resp, err = j.client.MakeRequest(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	log.Printf("D! JBoss API Req HTTP REQ:%#+v", req)
	log.Printf("D! JBoss API Req HTTP RESP:%#+v", resp)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			req.RequestURI,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	//req, err := http.NewRequest("POST", serverUrl.String(), bytes.NewBuffer(requestBody))

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("E! JBoss Error: %s", err)
		return nil, err
	}

	// Debug response
	//log.Printf("D! body: %s\n", body)

	return []byte(body), nil
}

func init() {
	inputs.Add("jboss", func() telegraf.Input {
		return &JBoss{
			client: &RealHTTPClient{},
		}
	})
}
