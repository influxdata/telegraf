package jboss

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

const (
	GET_HOSTS           = 0
	GET_SERVERS         = 1
	GET_DB_STAT         = 2
	GET_JVM_STAT        = 3
	GET_DEPLOYMENTS     = 4
	GET_DEPLOYMENT_STAT = 5
	GET_WEB_STAT        = 6
	GET_JMS_QUEUE_STAT  = 7
	GET_JMS_TOPIC_STAT  = 8
	GET_TRANSACTION_STAT  = 9
)

type KeyVal struct {
	Key string
	Val interface{}
}

// Define an ordered map
type OrderedMap []KeyVal

// Implement the json.Marshaler interface
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

type HostResponse struct {
	Outcome string   `json:"outcome"`
	Result  []string `json:"result"`
}

type DatasourceResponse struct {
	Outcome string          `json:"outcome"`
	Result  DatabaseMetrics `json:"result"`
}

type JMSResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

type TransactionResponse struct {
	Outcome string                 `json:"outcome"`
    Result  map[string]interface{} `json:"result"`
}

type JMSMetrics struct {
	Count string `json:"message-count"`
	Added string `json:"messages-added"`
}

type DatabaseMetrics struct {
	DataSource   map[string]DataSourceMetrics `json:"data-source"`
	XaDataSource map[string]DataSourceMetrics `json:"xa-data-source"`
}

type DataSourceMetrics struct {
	JndiName   string       `json:"jndi-name"`
	Statistics DBStatistics `json:"statistics"`
}

type DBStatistics struct {
	Pool DBPoolStatistics `json:"pool"`
}

type DBPoolStatistics struct {
	ActiveCount    string `json:"ActiveCount"`
	AvailableCount string `json:"AvailableCount"`
	InUseCount     string `json:"InUseCount"`
}

type JVMResponse struct {
	Outcome string     `json:"outcome"`
	Result  JVMMetrics `json:"result"`
}

type WebResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
}

type JVMMetrics struct {
	Type map[string]interface{} `json:"type"`
}

type DeploymentResponse struct {
	Outcome string            `json:"outcome"`
	Result  DeploymentMetrics `json:"result"`
}

type DeploymentMetrics struct {
	Name          string                 `json:"name"`
	RuntimeName   string                 `json:"runtime-name"`
	Status        string                 `json:"status"`
	Subdeployment map[string]interface{} `json:"subdeployment"`
}

type WebMetrics struct {
	ActiveSessions    string                 `json:"active-sessions"`
	ContextRoot       string                 `json:"context-root"`
	ExpiredSessions   string                 `json:"expired-sessions"`
	MaxActiveSessions string                 `json:"max-active-sessions"`
	SessionsCreated   string                 `json:"sessions-created"`
	Servlet           map[string]interface{} `json:"servlet"`
}

type Metric struct {
	FullName string                 `json:"full_name"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Fields   map[string]interface{} `json:"metric"`
}

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
			log.Printf("D! JBoss Plugin Working as Domain: %t\n", h.ExecAsDomain)
			if h.ExecAsDomain {
				bodyContent, err := h.prepareRequest(GET_HOSTS, nil)
				if err != nil {
					errorChannel <- err
				}

				out, err := h.doRequest(server, bodyContent)

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
			log.Printf("D! Get Servers from host: %s\n", host)

			servers := HostResponse{Outcome: "", Result: []string{"standalone"}}

			if h.ExecAsDomain {
				//get servers
				adr := OrderedMap{
					{"host", host},
				}

				bodyContent, err := h.prepareRequest(GET_SERVERS, adr)
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
				log.Printf("D! JBoss Plugin Processing Servers from host:[ %s ] : Server [ %s ]\n", host, server)
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
						h.getJMSStatistics(acc, serverURL, host, server, GET_JMS_QUEUE_STAT)
						h.getJMSStatistics(acc, serverURL, host, server, GET_JMS_TOPIC_STAT)
					case "transaction":	
						h.getTransactionStatistics(acc, serverURL, host, server)
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

	bodyContent, err := h.prepareRequest(GET_WEB_STAT, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
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
	}

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

	bodyContent, err := h.prepareRequest(GET_DB_STAT, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
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
	}

	return nil
}

// Gathers Transaction data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getTransactionStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	host string,
	serverName string,
) error {
	//fmt.Printf("getTransactionStatistics %s %s\n", host, serverName)

	adr := OrderedMap{}
	
	if h.ExecAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"subsystem", "transactions"},
		}
	} else {
		adr = OrderedMap{
			{"subsystem", "transactions"},
		}
	}

	bodyContent, err := h.prepareRequest(GET_TRANSACTION_STAT, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)
	
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
		server := TransactionResponse{}
		if err = json.Unmarshal(out, &server); err != nil {
			return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
		}

		fields := make(map[string]interface{})
		for key, value := range server.Result {
			if strings.Contains(key, "number-of") {
				fields[key], _ = strconv.ParseInt(value.(string), 10, 64)
			}
		}
		tags := map[string]string{
			"jboss_host":   host,
			"jboss_server": serverName,
		}
		
		acc.AddFields("jboss_transaction", fields, tags)
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
	//fmt.Printf("getDatasourceStatistics %s %s\n", host, serverName)

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

	fmt.Println(bodyContent)

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
		jmsresponse := JMSResponse{}
		if err = json.Unmarshal(out, &jmsresponse); err != nil {
			return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
		}
		fmt.Println("server:")
		fmt.Println(jmsresponse)

		for jmsQueue, value := range jmsresponse.Result {
			fields := make(map[string]interface{})
			v := value.(map[string]interface{})
			fields["message-count"] = v["message-count"]
			fields["messages-added"] = v["messages-added"]
			if opType == GET_JMS_QUEUE_STAT {
				fields["consumer-count"] = v["consumer-count"]
			} else {
				fields["subscription-count"] = v["subscription-count"]
			}
			tags := map[string]string{
				"jboss_host":   host,
				"jboss_server": serverName,
				"name":         jmsQueue,
			}
			acc.AddFields("jboss_jms", fields, tags)
		}
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

	bodyContent, err := h.prepareRequest(GET_JVM_STAT, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	log.Printf("D! JBoss API Req err: %s", err)
	log.Printf("D! JBoss API Req out: %s", out)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
		server := JVMResponse{}
		if err = json.Unmarshal(out, &server); err != nil {
			return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
		}

		for typeName, value := range server.Result.Type {

			fields := make(map[string]interface{})

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
			case "memory-pool":
				data := value.(map[string]interface{})
				name := data["name"].(map[string]interface{})
				for poolName, poolArea := range name {
				    log.Printf("D! PoolName: %s %s", poolName, poolArea)
				    poolData := poolArea.(map[string]interface{})
					usage := poolData["usage"].(map[string]interface{})
					h.flatten(usage, fields, poolName)
				}
			case "garbage-collector":
				gc := value.(map[string]interface{})
				gc_name := gc["name"].(map[string]interface{})
				for gc_type, gc_val := range gc_name {
					object := gc_val.(map[string]interface{})
					fields[gc_type+"_count"] = object["collection-count"]
					fields[gc_type+"_time"] = object["collection-time"]
				}
			}

			tags := map[string]string{
				"jboss_host":   host,
				"jboss_server": serverName,
			}
			acc.AddFields("jboss_jvm", fields, tags)
		}

	}

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

	bodyContent, err := h.prepareRequest(GET_DEPLOYMENTS, adr)
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

			bodyContent, err := h.prepareRequest(GET_DEPLOYMENT_STAT, adr2)
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(serverURL, bodyContent)

			log.Printf("D! JBoss API Req err: %s", err)
			log.Printf("D! JBoss API Req out: %s", out)

			if err != nil {
				log.Printf("E! JBoss Error handling response 3: %s\n", err)
				errorChannel <- err
			} else {
				deployment := DeploymentResponse{}
				if err = json.Unmarshal(out, &deployment); err != nil {
					errorChannel <- errors.New("Error decoding JSON response")
				}

				for typeName, value := range deployment.Result.Subdeployment {
					fields := make(map[string]interface{})

					t := value.(map[string]interface{})
					subsystem := t["subsystem"].(map[string]interface{})

					if value, ok := subsystem["ejb3"]; ok {

						ejb := value.(map[string]interface{})
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
								tags := map[string]string{
									"jboss_host":   host,
									"jboss_server": serverName,
									"name":         typeName,
									"ejb":          stateless,
									"system":       deployment.Result.RuntimeName,
								}
								acc.AddFields("jboss_ejb", fields, tags)
							}
						}
					}

					if webValue, ok := subsystem["web"]; ok {

						t2 := webValue.(map[string]interface{})
						contextRoot := t2["context-root"].(string)
						fields["active-sessions"] = t2["active-sessions"]
						fields["expired-sessions"] = t2["expired-sessions"]
						fields["max-active-sessions"] = t2["max-active-sessions"]
						fields["sessions-created"] = t2["sessions-created"]
						tags := map[string]string{
							"jboss_host":   host,
							"jboss_server": serverName,
							"name":         typeName,
							"context-root": contextRoot,
							"system":       deployment.Result.RuntimeName,
						}
						acc.AddFields("jboss_web_app", fields, tags)
					}
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

// Read memory-pool to produce field name and field value
// Parameters:
//    pool: Memmory-pool map to read
//    fields: Map to store generated fields.
//    name: Name of the memory-pool area
// Returns:
//    void
func (h *JBoss) readMemoryPool(pool map[string]interface{}, fields map[string]interface{}, name string) {

	poolArea := pool[name].(map[string]interface{})
	usage := poolArea["usage"].(map[string]interface{})
	
	h.flatten(usage, fields, name)
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
func (j *JBoss) prepareRequest(optype int, adress OrderedMap) (map[string]interface{}, error) {
	bodyContent := make(map[string]interface{})

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
	case GET_JVM_STAT:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive"] = "true"
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_DEPLOYMENTS:
		bodyContent["operation"] = "read-children-names"
		bodyContent["child-type"] = "deployment"
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_DEPLOYMENT_STAT:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 3
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_WEB_STAT:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 0
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_JMS_QUEUE_STAT:
		bodyContent["operation"] = "read-children-resources"
		bodyContent["child-type"] = "jms-queue"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 2
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_JMS_TOPIC_STAT:
		bodyContent["operation"] = "read-children-resources"
		bodyContent["child-type"] = "jms-topic"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 2
		bodyContent["address"] = adress
		bodyContent["json.pretty"] = 1
	case GET_TRANSACTION_STAT:
		bodyContent["operation"] = "read-resource"
		bodyContent["include-runtime"] = "true"
		bodyContent["recursive-depth"] = 0
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

	req := dac.NewRequest(j.Username, j.Password, method, serverUrl.String(), string(requestBody[:]))
    req.Header.Add("Content-Type", "application/json")

	resp, err := req.Execute()

	if err != nil {
		log.Printf("D! HTTP REQ:%#+v", req)
		log.Printf("D! HTTP RESP:%#+v", resp)
		return nil, err
	}

	log.Printf("D! JBoss API Req HTTP REQ:%#+v", req)
	log.Printf("D! JBoss API Req HTTP RESP:%#+v", resp)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			serverUrl,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}
	
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
