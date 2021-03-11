package jboss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	dac "github.com/stefa975/go-http-digest-auth-client"
)

// Constants applied to jboss management query types
const (
	GetExecStat = iota
	GetHosts
	GetServers
	GetDBStat
	GetJVMStat
	GetDeployments
	GetDeploymentStat
	GetWebStat
	GetJMSQueueStat
	GetJMSTopicStat
	GetTransactionStat
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

	buf.WriteString("[")
	for i, kv := range omap {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.WriteString("{")
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv.Val)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
		buf.WriteString("}")
	}

	buf.WriteString("]")
	return buf.Bytes(), nil
}

// HostResponse expected GetHost response type
type ExecTypeResponse struct {
	Outcome string                 `json:"outcome"`
	Result  map[string]interface{} `json:"result"`
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

// TransactionResponse transaction related metrics
type TransactionResponse struct {
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
	ActiveCount    interface{} `json:"ActiveCount"`
	AvailableCount interface{} `json:"AvailableCount"`
	InUseCount     interface{} `json:"InUseCount"`
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
	Servers []string `toml:"servers"`
	Metrics []string `toml:"metrics"`

	Username string
	Password string

	Authorization string

	Log telegraf.Logger `toml:"-"`

	ResponseTimeout internal.Duration `toml:"response_timeout"`

	tls.ClientConfig

	client *http.Client
}

var sampleConfig = `
  # Config for get statistics from JBoss AS
  servers = [
    "http://localhost:9090/management",
  ]
  
  ## Username and password
  username = ""
  password = ""
  
  ## authorization mode could be "basic" or "digest"
  authorization = "digest"

  ## Optional SSL Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
  ## Metric selection
  metrics =[
	"jvm",
	"web", 		# Handles both EAP <=6.X/AS <=7.X and EAP >=7.X/Widlfly > 8
	"deployment",
	"database",
	"transaction",
	"jms",
  ]
`

func (*JBoss) SampleConfig() string {
	return sampleConfig
}

func (*JBoss) Description() string {
	return "Telegraf plugin for gathering metrics from JBoss AS"
}

func (h *JBoss) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	// Create an HTTP client that is re-used for each
	// collection interval

	if h.client == nil {
		client, err := h.createHttpClient()
		if err != nil {
			return err
		}
		h.client = client
	}

	for _, server := range h.Servers {

		//Check Exec Mode for each servers

		bodyContent, err := h.createRequestBody(GetExecStat, nil)
		if err != nil {
			acc.AddError(err)
		}

		out, err := h.doRequest(server, bodyContent)
		if err != nil {
			h.Log.Errorf("JBoss Error handling ExecMode Test: %s", err)
			acc.AddError(err)
		}
		// Unmarshal json
		exec := ExecTypeResponse{}
		if err = json.Unmarshal(out, &exec); err != nil {
			acc.AddError(fmt.Errorf("Error decoding JSON response (ExecTypeResponse) %s,%s", out, err))
			return nil
		}

		execAsDomain := exec.Result["launch-type"] == "DOMAIN"
		isEAP7 := isEAP7Version(exec.Result["product-name"].(string), exec.Result["product-version"].(string))
		h.Log.Debugf("JBoss Plugin Working as Domain: %t EAP7: %t  for server %s", execAsDomain, isEAP7, server)

		wg.Add(1)
		go func(server string, execAsDomain bool) {
			defer wg.Done()
			//default as standalone server
			hosts := HostResponse{Outcome: "", Result: []string{"standalone"}}

			if execAsDomain {
				bodyContent, err := h.createRequestBody(GetHosts, nil)
				if err != nil {
					acc.AddError(err)
				}

				out, err := h.doRequest(server, bodyContent)

				if err != nil {
					h.Log.Errorf("JBoss Error handling response 1: %s", err)
					h.Log.Errorf("JBoss server:%s bodyContent %s", server, bodyContent)
					acc.AddError(err)
					return
				}
				// Unmarshal json

				if err = json.Unmarshal(out, &hosts); err != nil {
					acc.AddError(fmt.Errorf("Error decoding JSON response (HostResponse) %s :%s", out, err))
				}
				h.Log.Debugf("JBoss HOSTS %s", hosts)
			}

			h.getServersOnHost(acc, server, execAsDomain, isEAP7, hosts.Result)

		}(server, execAsDomain)
	}

	wg.Wait()

	return nil
}

// Check if JBoss is EE 6 or EE 7
// Parameters:
//     productName   : The tname of the product
//     productVersion: version of the produkt
//
// Returns:
//     bool

func isEAP7Version(productName string, productVersion string) bool {
	if strings.Contains(productName, "EAP") {
		return strings.HasPrefix(productVersion, "7.")
	}
	return strings.HasPrefix(productVersion, "10.")
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
	execAsDomain bool,
	isEAP7 bool,
	hosts []string,
) error {
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			h.Log.Infof("Get Servers from host: %s\n", host)

			servers := HostResponse{Outcome: "", Result: []string{"standalone"}}

			if execAsDomain {
				//get servers
				adr := OrderedMap{
					{"host", host},
				}
				bodyContent, err := h.createRequestBody(GetServers, adr)
				if err != nil {
					acc.AddError(err)
				}

				out, err := h.doRequest(serverURL, bodyContent)

				if err != nil {
					h.Log.Errorf("JBoss Error handling response 2: ERR:%s : OUTPUT:%s", err, out)
					acc.AddError(err)
					return
				}

				if err = json.Unmarshal(out, &servers); err != nil {
					h.Log.Errorf("JBoss Error on JSON decoding")
					acc.AddError(err)
				}
			}

			for _, server := range servers.Result {
				h.Log.Infof("JBoss Plugin Processing Servers from host:[ %s ] : Server [ %s ]", host, server)
				for _, v := range h.Metrics {
					switch v {
					case "jvm":
						h.getJVMStatistics(acc, serverURL, execAsDomain, host, server)
					case "web":
						if isEAP7 {
							h.getUndertowStatistics(acc, serverURL, execAsDomain, host, server, "ajp")
							h.getUndertowStatistics(acc, serverURL, execAsDomain, host, server, "http")
							h.getUndertowStatistics(acc, serverURL, execAsDomain, host, server, "https")
						} else {
							h.getWebStatistics(acc, serverURL, execAsDomain, host, server, "ajp")
							h.getWebStatistics(acc, serverURL, execAsDomain, host, server, "http")
						}
					case "deployment":
						h.getServerDeploymentStatistics(acc, serverURL, execAsDomain, host, server)
					case "database":
						h.getDatasourceStatistics(acc, serverURL, execAsDomain, host, server)
					case "jms":
						if isEAP7 {
							h.getJMSStatistics(acc, serverURL, execAsDomain, host, server, "messaging-activemq", GetJMSQueueStat)
							h.getJMSStatistics(acc, serverURL, execAsDomain, host, server, "messaging-activemq", GetJMSTopicStat)
						} else {
							h.getJMSStatistics(acc, serverURL, execAsDomain, host, server, "messaging", GetJMSQueueStat)
							h.getJMSStatistics(acc, serverURL, execAsDomain, host, server, "messaging", GetJMSTopicStat)
						}
					case "transaction":
						h.getTransactionStatistics(acc, serverURL, execAsDomain, host, server)
					default:
						h.Log.Errorf("Jboss doesn't exist the metric set %s", v)
					}
				}
			}
		}(host)
	}

	wg.Wait()

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
	execAsDomain bool,
	host string,
	serverName string,
) error {
	adr := OrderedMap{}

	if execAsDomain {
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

	bodyContent, err := h.createRequestBody(GetTransactionStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	server := TransactionResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}

	fields := make(map[string]interface{})
	for key, value := range server.Result {
		if strings.Contains(key, "number-of") {
			if v, ok := value.(string); ok {
				fields[key], _ = strconv.ParseInt(v, 10, 64)
			}
		}
	}
	tags := map[string]string{
		"jboss_host":   host,
		"jboss_server": serverName,
	}

	acc.AddFields("jboss_transaction", fields, tags)

	return nil
}

// Gathers web data from a particular host
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     execAsDomain: JBoss runs in domain
//     host     : the host being queried
//     server   : the server being queried
//
// Returns:
//     error: Any error that may have occurred

func (h *JBoss) getWebStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	execAsDomain bool,
	host string,
	serverName string,
	connector string,
) error {
	adr := OrderedMap{}
	if execAsDomain {
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

	bodyContent, err := h.createRequestBody(GetWebStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}
	server := WebResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response (WebResponse): %s : %s", out, err)
	}

	fields := make(map[string]interface{})
	for key, value := range server.Result {
		if value == nil {
			continue
		}
		switch key {
		case "bytesReceived", "bytesSent", "requestCount", "errorCount", "maxTime", "processingTime":
			switch v := value.(type) {
			case int:
				fields[key] = float64(v)
			case float64:
				fields[key] = v
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					fields[key] = f
				}
			}
		}
	}
	tags := map[string]string{
		"jboss_host":   host,
		"jboss_server": serverName,
		"type":         connector,
	}
	acc.AddFields("jboss_web", fields, tags)

	return nil
}

func (h *JBoss) getUndertowStatistics(
	acc telegraf.Accumulator,
	serverURL string,
	execAsDomain bool,
	host string,
	serverName string,
	listener string,
) error {
	adr := OrderedMap{}
	listenerName := "default"
	if listener == "ajp" || listener == "https" {
		listenerName = listener
	}
	listener = listener + "-listener"
	if execAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"subsystem", "undertow"},
			{"server", "default-server"},
			{listener, listenerName},
		}
	} else {
		adr = OrderedMap{
			{"subsystem", "undertow"},
			{"server", "default-server"},
			{listener, listenerName},
		}
	}

	bodyContent, err := h.createRequestBody(GetWebStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	if err != nil {
		return fmt.Errorf("error on request to %s OUT: %s  ERR: %s\n", serverURL, out, err)
	}
	server := WebResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response: OUT: %s  ERR: %s", out, err)
	}

	fields := make(map[string]interface{})
	for key, value := range server.Result {
		h.Log.Debugf("LISTERNER %s : %s", key, value)
		if value == nil {
			continue
		}
		switch key {
		case "bytes-received", "bytes-sent", "request-count", "error-count", "max-processing-time", "processing-time":
			switch v := value.(type) {
			case int:
				fields[key] = float64(v)
			case float64:
				fields[key] = v
			case string:
				if f, err := strconv.ParseFloat(v, 64); err == nil {
					fields[key] = f
				}
			}
		}
	}
	tags := map[string]string{
		"jboss_host":   host,
		"jboss_server": serverName,
		"type":         listener,
	}
	acc.AddFields("jboss_web", fields, tags)

	return nil
}

func getPoolFields(pool DBPoolStatistics) map[string]interface{} {
	retmap := make(map[string]interface{})
	//Jboss EAP 6/AS 7.X returns "strings", wildfly 12 returns integers
	switch pool.ActiveCount.(type) {
	case string:
		retmap["in-use-count"], _ = strconv.ParseFloat(pool.InUseCount.(string), 64)
		retmap["active-count"], _ = strconv.ParseFloat(pool.ActiveCount.(string), 64)
		retmap["available-count"], _ = strconv.ParseFloat(pool.AvailableCount.(string), 64)
	case float64:
		retmap["in-use-count"] = pool.InUseCount.(float64)
		retmap["active-count"] = pool.ActiveCount.(float64)
		retmap["available-count"] = pool.AvailableCount.(float64)
	default:
		retmap["in-use-count"] = pool.InUseCount
		retmap["active-count"] = pool.ActiveCount
		retmap["available-count"] = pool.AvailableCount
	}
	return retmap
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
	execAsDomain bool,
	host string,
	serverName string,
) error {
	adr := OrderedMap{}
	if execAsDomain {
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

	bodyContent, err := h.createRequestBody(GetDBStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	if err != nil {
		return fmt.Errorf("error on request to %s : OUT: %s ERR: %s\n", serverURL, out, err)
	}
	server := DatasourceResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response (DataSourceResponse): %s : %s", out, err)
	}

	for database, value := range server.Result.DataSource {
		fields := getPoolFields(value.Statistics.Pool)
		tags := map[string]string{
			"jboss_host":   host,
			"jboss_server": serverName,
			"name":         database,
		}
		acc.AddFields("jboss_database", fields, tags)
	}
	for database, value := range server.Result.XaDataSource {
		fields := getPoolFields(value.Statistics.Pool)
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
	execAsDomain bool,
	host string,
	serverName string,
	subsystem string,
	opType int,
) error {

	adr := OrderedMap{}

	serverID := "hornetq-server"

	if subsystem == "messaging-activemq" {
		serverID = "server"
	}

	if execAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
			{"subsystem", subsystem},
			{serverID, "default"},
		}
	} else {
		adr = OrderedMap{
			{"subsystem", subsystem},
			{serverID, "default"},
		}
	}

	bodyContent, err := h.createRequestBody(opType, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	if err != nil {
		return fmt.Errorf("error on request to %s : OUT: %s ERR: %s\n", serverURL, out, err)
	}
	jmsresponse := JMSResponse{}
	if err = json.Unmarshal(out, &jmsresponse); err != nil {
		return fmt.Errorf("Error decoding JSON response (JMSResponse): %s : %s", out, err)
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
	execAsDomain bool,
	host string,
	serverName string,
) error {
	adr := OrderedMap{}
	if execAsDomain {
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

	bodyContent, err := h.createRequestBody(GetJVMStat, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	if err != nil {
		return fmt.Errorf("error on request to %s : OUT: %s ERR: %s\n", serverURL, out, err)
	}

	server := JVMResponse{}
	if err = json.Unmarshal(out, &server); err != nil {
		return fmt.Errorf("Error decoding JSON response (JVMReponse): %s : %s", out, err)
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

func (h *JBoss) processWebAppStats(acc telegraf.Accumulator, web map[string]interface{}, tags map[string]string) {
	fields := make(map[string]interface{})
	contextRoot := web["context-root"].(string)
	fields["active-sessions"] = web["active-sessions"]
	fields["expired-sessions"] = web["expired-sessions"]
	fields["max-active-sessions"] = web["max-active-sessions"]
	fields["sessions-created"] = web["sessions-created"]
	tags["context-root"] = contextRoot
	acc.AddFields("jboss_web_app", fields, tags)
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
	execAsDomain bool,
	host string,
	serverName string,
) error {
	adr := OrderedMap{}

	if execAsDomain {
		adr = OrderedMap{
			{"host", host},
			{"server", serverName},
		}
	}

	bodyContent, err := h.createRequestBody(GetDeployments, adr)
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

	if err != nil {
		return fmt.Errorf("error on request to  %s OUT: %s  ERR : %s\n", serverURL, out, err)
	}

	deployments := HostResponse{}
	if err = json.Unmarshal(out, &deployments); err != nil {
		return fmt.Errorf("Error decoding JSON response (HostResponse): %s : %s", out, err)
	}

	for _, deployment := range deployments.Result {
		adr2 := OrderedMap{}
		if execAsDomain {
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

		bodyContent, err := h.createRequestBody(GetDeploymentStat, adr2)
		if err != nil {
			acc.AddError(err)
		}

		out, err := h.doRequest(serverURL, bodyContent)

		h.Log.Debugf("JBoss Deployment API Req err: %s", err)
		h.Log.Debugf("JBoss Deployment API Req out: %s", out)

		if err != nil {
			h.Log.Errorf("JBoss Deployment Error handling response 3: %s", err)
			acc.AddError(err)
		}
		// everything ok ! continue with decoding data
		deploy := DeploymentResponse{}
		if err = json.Unmarshal(out, &deploy); err != nil {
			acc.AddError(fmt.Errorf("Error decoding JSON response(DeploymentResponse): %s : %s", out, err))
		}
		// This struct apply on EAR files
		for typeName, value := range deploy.Result.Subdeployment {
			if value == nil {
				h.Log.Debugf("JBoss Deployment WARNING Subdeployment value is NULL")
				continue
			}

			t := value.(map[string]interface{})
			if t["subsystem"] == nil {
				h.Log.Debugf("D! JBoss Deployment WARNING SUBDEPLOYMENT Subsystem is NULL")
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
			//undertow is the new web sybsystem since wildfly 8
			if webValue, ok := subsystem["undertow"]; ok {
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
				h.Log.Debugf("JBoss Deployment SUBSYSTEM  value NULL")
				continue
			}
			if typeName == "web" || typeName == "undertow" {
				web := value.(map[string]interface{})
				tags := map[string]string{
					"jboss_host":   host,
					"jboss_server": serverName,
					"name":         deploy.Result.Name,
					"runtime_name": deploy.Result.RuntimeName,
				}
				h.processWebAppStats(acc, web, tags)
			} else {
				h.Log.Warnf("JBoss Deployment WAR  from type %s", typeName)
			}
		}
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

func (h *JBoss) createRequestBody(optype int, address OrderedMap) (map[string]interface{}, error) {
	bodyContent := make(map[string]interface{})

	// Create bodyContent
	switch optype {
	case GetExecStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-resource",
			"attributes-only": "true",
			"include-runtime": "true",
			"address":         address,
			"json.pretty":     1,
		}
	case GetHosts:
		bodyContent = map[string]interface{}{
			"operation":   "read-children-names",
			"child-type":  "host",
			"address":     address,
			"json.pretty": 1,
		}
	case GetServers:
		bodyContent = map[string]interface{}{
			"operation":       "read-children-names",
			"child-type":      "server",
			"recursive-depth": 0,
			"address":         address,
			"json.pretty":     1,
		}
	case GetDBStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-resource",
			"include-runtime": "true",
			"recursive-depth": 2,
			"address":         address,
			"json.pretty":     1,
		}
	case GetJVMStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-resource",
			"include-runtime": "true",
			"recursive":       "true",
			"address":         address,
			"json.pretty":     1,
		}
	case GetDeployments:
		bodyContent = map[string]interface{}{
			"operation":   "read-children-names",
			"child-type":  "deployment",
			"address":     address,
			"json.pretty": 1,
		}
	case GetDeploymentStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-resource",
			"include-runtime": "true",
			"recursive-depth": 3,
			"address":         address,
			"json.pretty":     1,
		}
	case GetWebStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-resource",
			"include-runtime": "true",
			"recursive-depth": 0,
			"address":         address,
			"json.pretty":     1,
		}
	case GetJMSQueueStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-children-resources",
			"child-type":      "jms-queue",
			"include-runtime": "true",
			"recursive-depth": 2,
			"address":         address,
			"json.pretty":     1,
		}
	case GetJMSTopicStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-children-resources",
			"child-type":      "jms-topic",
			"include-runtime": "true",
			"recursive-depth": 2,
			"address":         address,
			"json.pretty":     1,
		}
	case GetTransactionStat:
		bodyContent = map[string]interface{}{
			"operation":       "read-resources",
			"include-runtime": "true",
			"recursive-depth": 0,
			"address":         address,
			"json.pretty":     1,
		}
	}

	return bodyContent, nil
}

func (h *JBoss) doRequest(domainURL string, bodyContent map[string]interface{}) ([]byte, error) {

	serverURL, err := url.Parse(domainURL)
	if err != nil {
		return nil, err
	}
	requestBody, err := json.Marshal(bodyContent)
	if err != nil {
		h.Log.Errorf("JBoss Marshal error: %s", err)
		return nil, err
	}
	method := "POST"

	// Debug JSON request
	h.Log.Debugf("Req: %s", requestBody)

	dr := dac.NewRequest(h.Username, h.Password, method, serverURL.String(), string(requestBody[:]))
	dr.Header.Add("Content-Type", "application/json")

	resp, err := dr.Execute()

	if err != nil {
		h.Log.Errorf("HTTP REQ:%#+v", dr)
		h.Log.Errorf("HTTP RESP:%#+v", resp)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			serverURL,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.Log.Errorf("JBoss Error: %s", err)
		return nil, err
	}

	return []byte(body), nil
}

func (j *JBoss) createHttpClient() (*http.Client, error) {
	if j.ResponseTimeout.Duration < time.Second {
		j.ResponseTimeout.Duration = time.Second * 5
	}

	tlsConfig, err := j.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: j.ResponseTimeout.Duration,
	}

	return client, nil
}

func init() {
	inputs.Add("jboss", func() telegraf.Input {
		return &JBoss{}
	})
}
