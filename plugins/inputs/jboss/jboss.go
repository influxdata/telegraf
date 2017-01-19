package jboss

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
		GET_JVM_STAT = 3
		GET_DEPLOYMENTS = 4
		GET_DEPLOYMENT_STAT = 5
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

type JVMResponse struct {
	Outcome string `json:"outcome"`
	Result JVMMetrics `json:"result"`
}

type JVMMetrics struct {
	Type map[string]interface{} `json:"type"`
}

type DeploymentResponse struct {
	Outcome string `json:"outcome"`
	Result DeploymentMetrics `json:"result"`
}

type DeploymentMetrics struct {
	Name string `json:"name"`
	RuntimeName string `json:"runtime-name"`
	Status string `json:"status"`
	Subdeployment map[string]interface{} `json:"subdeployment"`
}

type WebMetrics struct {
	ActiveSessions string `json:"active-sessions"`
	ContextRoot string `json:"context-root"`
	ExpiredSessions string `json:"expired-sessions"`
	MaxActiveSessions string `json:"max-active-sessions"`
	SessionsCreated string `json:"sessions-created"`
	Servlet map[string]interface{} `json:"servlet"`
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
  ## Username and password
  username = ""
  password = ""
  authrization = basic|digest
  
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
        fmt.Printf("Recieved status code '%v' auth skipped\n", resp.StatusCode)
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
    fmt.Printf("response code: %d \n", resp.StatusCode)

    if resp.StatusCode != http.StatusOK {
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }
        fmt.Printf("response body: %s \n", string(body))
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
			bodyContent, err := h.prepareRequest(GET_HOSTS, nil);
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(server, bodyContent)

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
				//oneH := []string{hosts.Result[0],hosts.Result[1]}
				h.getServersOnHost(acc, server, hosts.Result)
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

	for _, host := range hosts {
		fields := make(map[string]interface{})
		fields["name"] = host
		tags := map[string]string{
			"type":   "host",
		}
	
		acc.AddFields("jboss_domain", fields, tags)
	}
	
	errorChannel := make(chan error, len(hosts))

	for _, host := range hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			// fmt.Printf("Get Servers from host %s\n", host)
//			adr := make(map[string]interface{})
//			adr["host"] = host
			adr :=  OrderedMap{
				{"host",  host},
			}
//			adr := []string{"host\":" + host }
			//adr := []string{"host=" + host}
			bodyContent, err := h.prepareRequest(GET_SERVERS, adr);
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(serverURL, bodyContent)

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
					fields := make(map[string]interface{})
					fields["name"] = server
					tags := map[string]string{
						"host":   host,
						"type":   "server",
					}
	
					acc.AddFields("jboss_domain", fields, tags)
					
					h.getDatasourceStatistics(acc, serverURL, host, server)
					h.getJVMStatistics(acc, serverURL, host, server)
					h.getServerDeploymentStatistics(acc, serverURL, host, server)
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
	
	//adr := make(map[string]interface{})
	adr := OrderedMap{
		{"host", host},
		{"server", serverName},
		{"subsystem", "datasources"},
	}
	//adr["host"] = host
	//adr["server"] = serverName
	//adr["subsystem"] = "datasources"
	 
	bodyContent, err := h.prepareRequest(GET_DB_STAT, adr);
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

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
			acc.AddFields("jboss_database", fields, tags)
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
			acc.AddFields("jboss_database", fields, tags)
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
	adr := OrderedMap{
        {"host",  host},
        {"server", serverName},
		{"core-service", "platform-mbean"},
    }

	bodyContent, err := h.prepareRequest(GET_JVM_STAT, adr);
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

//	fmt.Println("out: %s\n", out)
	
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} else {
		server := JVMResponse{}
		if err = json.Unmarshal(out, &server); err != nil {
			return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
		}
//		fmt.Println(server)

	
		for typeName, value := range server.Result.Type {
			fields := make(map[string]interface{})
			if typeName == "threading" {
				t := value.(map[string]interface{})
				fields["thread-count"]  = t["thread-count"]
				fields["peak-thread-count"]  = t["peak-thread-count"]
				fields["daemon-thread-count"]  = t["daemon-thread-count"]
			} else if  typeName == "memory" {
				mem := value.(map[string]interface{})
				heap := mem["heap-memory-usage"].(map[string]interface{})
				nonHeap := mem["non-heap-memory-usage"].(map[string]interface{})
				h.flatten(heap, fields, "heap")
				h.flatten(nonHeap, fields, "nonheap")
			} else if typeName == "garbage-collector" {
				gc := value.(map[string]interface{})
				gc_name := gc["name"].(map[string]interface{})
				Scavenge := gc_name["PS_Scavenge"].(map[string]interface{})
				MarkSweep := gc_name["PS_MarkSweep"].(map[string]interface{})
				fields["scavenge-count"] = Scavenge["collection-count"]
				fields["scavenge-time"] = Scavenge["collection-time"]
				fields["marksweep-count"] = MarkSweep["collection-count"]
				fields["marksweep-time"] = MarkSweep["collection-time"]
			}
			tags := map[string]string{
				"host":   host,
				"server": serverName,
				"name":   typeName,
				"type":   "jvm",
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
	adr := OrderedMap{
        {"host",  host},
        {"server", serverName},
    }

	bodyContent, err := h.prepareRequest(GET_DEPLOYMENTS, adr);
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	}

	out, err := h.doRequest(serverURL, bodyContent)

//	fmt.Println("out: %s\n", out)
	
	if err != nil {
		return fmt.Errorf("error on request to %s : %s\n", serverURL, err)
	} 
	
	deployments := HostResponse{}
	if err = json.Unmarshal(out, &deployments); err != nil {
		return fmt.Errorf("Error decoding JSON response: %s : %s", out, err)
	}
//		fmt.Println(server)

	for _, value := range deployments.Result {
		fields := make(map[string]interface{})
		fields["name"] = value
		tags := map[string]string{
			"host":   host,
			"server": serverName,
			"type":   "deployment",
			}
	
			acc.AddFields("jboss_domain", fields, tags)
	}
	
	errorChannel := make(chan error, len(deployments.Result))

	for _, deployment := range deployments.Result {
		wg.Add(1)
		go func(deployment string) {
			defer wg.Done()
			// fmt.Printf("Get Servers from host %s\n", host)
//			adr := make(map[string]interface{})
//			adr["host"] = host
			adr2 := OrderedMap{
				{"host",  host},
				{"server", serverName},
				{"deployment", deployment},
			}
			
//			adr := []string{"host\":" + host }
			//adr := []string{"host=" + host}
			bodyContent, err := h.prepareRequest(GET_DEPLOYMENT_STAT, adr2);
			if err != nil {
				errorChannel <- err
			}

			out, err := h.doRequest(serverURL, bodyContent)

			if err != nil {
				fmt.Printf("Error handling response: %s\n", err)
				errorChannel <- err
			} else {
				deployment := DeploymentResponse{}
				if err = json.Unmarshal(out, &deployment); err != nil {
					errorChannel <- errors.New("Error decoding JSON response")
				}

	//			fmt.Println(deployment)
				
				for typeName, value := range deployment.Result.Subdeployment {
					fields := make(map[string]interface{})
//					fmt.Println(typeName)
					t := value.(map[string]interface{})
					subsystem := t["subsystem"].(map[string]interface{})
//					fmt.Println(t["subsystem"])
					
					if value, ok := subsystem["ejb"]; ok {
						fmt.Println("EJB data")
						fmt.Println(value)
					}
					
					if webValue, ok := subsystem["web"]; ok {
//						fmt.Println("WEB data")
//						fmt.Println(webValue)
						t2 := webValue.(map[string]interface{})

						fields["active-sessions"]  = t2["active-sessions"]
						fields["expired-sessions"]  = t2["expired-sessions"]
						fields["max-active-sessions"]  = t2["max-active-sessions"]
						fields["sessions-created"]  = t2["sessions-created"]
						tags := map[string]string{
							"host":   host,
							"server": serverName,
							"name":   typeName,
							"system": deployment.Result.RuntimeName,
							"type":   "deployment",
						}
						acc.AddFields("jboss_web", fields, tags)
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
//	fmt.Printf("Req: %s\n", requestBody)

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
		return nil, err
	}
	defer resp.Body.Close()
	
	// Process response
	
	if resp.StatusCode == http.StatusUnauthorized {
		fmt.Printf("Do digest\n")
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
		fmt.Printf("Error: %s", err)
		return nil, err
	}

	// Debug response
//	fmt.Printf("body: %s", body)

	return []byte(body), nil
}

func init() {
	inputs.Add("jboss", func() telegraf.Input {
		return &JBoss{
			client: &RealHTTPClient{},
		}
	})
}

