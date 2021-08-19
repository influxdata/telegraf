package f5_load_balancer

import (
	"fmt"
	"net/http"
	"encoding/json"
	"bytes"
	"io"
	"crypto/tls"
	"strings"
	"sync"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tidwall/gjson"
)

// Example struct should be named the same as the Plugin
type F5LoadBalancer struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Domain string `toml:"domain"`
	Collectors []string `toml:"collectors"`
	Token string
}

// Usually the default (example) configuration is contained in this constant.
// Please use '## '' to denote comments and '# ' to specify default settings and start each line with two spaces.
const sampleConfig = `
  ## F5 Load Balancer Username
  username = "" # required
  ## F5 Load Balancer Password
  password = "" # required
  ## F5 Load Balancer User Interface Endpoint
  domain = "https://f5.example.com/" # required
  ## Metrics to collect from the F5
  collectors = ["node","virtual","pool","net_interface"]
`

// Description will appear directly above the plugin definition in the config file
func (f5 *F5LoadBalancer) Description() string {
	return `Gathers metrics from an F5 Load Balancer's API`
}

// SampleConfig will populate the sample configuration portion of the plugin's configuration
func (f5 *F5LoadBalancer) SampleConfig() string {
	return sampleConfig
}

// Init can be implemented to do one-time processing stuff like initializing variables
func (f5 *F5LoadBalancer) Init() error {
	// Check your options according to your requirements
	if f5.Username == "" {
		return fmt.Errorf("Username cannot be empty")
	}
	if f5.Password == "" {
		return fmt.Errorf("Password cannot be empty")
	}
	if f5.Domain == "" {
		return fmt.Errorf("Domain cannot be empty")
	}
	if f5.Collectors == nil {
		f5.Collectors = []string{"node","virtual","pool","net_interface"}
	}
	f5.Token = ""

	return nil
}

func (f5 *F5LoadBalancer) Gather(acc telegraf.Accumulator) error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Due to self signed certificates in many orgs, we don't verify the cert
	f5.Authenticate()
	if f5.Token == "" {
		return fmt.Errorf("No Authentication Token. Exiting...")
	}

	f5.GatherNode(acc)
	fields := make(map[string]interface{})
	fields["f5_node_current_sessions_test"] = "10"
	tags := map[string]string{}
	tags["name"] = "testing"
	acc.AddFields("f5_load_balancer_test", fields, tags)

	f5.ResetToken()

	return nil
}

func (f5 *F5LoadBalancer) Authenticate() {
	body := map[string]string{"username":f5.Username,"password":f5.Password,"loginProviderName":"tmos"}
	jsonBody, _ := json.Marshal(body)
	client := &http.Client{}
	req, err := http.NewRequest("POST",f5.Domain+"/mgmt/shared/authn/login",bytes.NewBuffer(jsonBody))
	req.SetBasicAuth(f5.Username,f5.Password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	jsonString := string(data)
	token := gjson.Get(jsonString, "token.token").Str
	f5.Token = token
}

func (f5 *F5LoadBalancer) ResetToken() {
	f5.Token = ""
}

func (f5 *F5LoadBalancer) GatherNode(acc telegraf.Accumulator) {
	resp := f5.Call("/mgmt/tm/ltm/node")
	urls := make([]string,0,2)
	for _,item := range gjson.Get(resp,"items").Array() {
		selfLink := gjson.Get(item.Raw,"selfLink").Str
		urls = append(urls, strings.Split(strings.Split(selfLink, "localhost")[1],"?")[0]+"/stats")
	}

	var wg sync.WaitGroup
	for _,url := range urls {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
			defer wg.Done()
			resp = f5.Call(url)
			selfLinkSplit := strings.Split(gjson.Get(resp,"selfLink").Str,"~")
			tags := map[string]string{}
			if len(selfLinkSplit) >= 2 {
				selfLinkSplit = strings.Split(selfLinkSplit[2],"/")
				if len(selfLinkSplit) > 0 {
					tags["name"] = selfLinkSplit[0]
				}
			}
			if _, ok := tags["name"]; !ok {
				// Bad or malformed response
				return
			}			
			
			base := gjson.Get(resp, "entries.*.nestedStats.entries").Raw
			fields := make(map[string]interface{})
			fields["node_current_sessions"],_ = strconv.Atoi(gjson.Get(base,"curSessions.value").Raw)
			fields["node_serverside_bits_in"],_ = strconv.Atoi(gjson.Get(base, "serverside\\.bitsIn.value").Raw)
			fields["node_serverside_bits_out"],_ = strconv.Atoi(gjson.Get(base, "serverside\\.bitsOut.value").Raw)
			fields["node_serverside_current_connections"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.curConns.value").Raw)
			fields["node_serverside_packets_in"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.pktsIn.value").Raw)
			fields["node_serverside_packets_out"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.pktsOut.value").Raw)
			fields["node_serverside_total_connections"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.totConns.value").Raw)
			fields["node_total_requests"],_ = strconv.Atoi(gjson.Get(base,"totRequests.value").Raw)
			acc.AddFields("f5_load_balancer", fields, tags)
		}(&wg, url, acc)
	}
	wg.Wait()

}

func (f5 *F5LoadBalancer) Call(endpoint string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET",f5.Domain+endpoint,nil)
	req.Header.Set("Content-Type","application/json")
	req.Header.Set("X-F5-Auth-Token",f5.Token)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	jsonString := string(data)
	return jsonString
}

func init() {
	inputs.Add("f5_load_balancer", func() telegraf.Input {
		return &F5LoadBalancer{}
	})
}
