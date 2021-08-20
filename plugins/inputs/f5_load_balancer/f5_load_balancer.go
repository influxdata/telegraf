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
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tidwall/gjson"
)

type F5LoadBalancer struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Domain string `toml:"domain"`
	Collectors []string `toml:"collectors"`
	Token string
}

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

func (f5 *F5LoadBalancer) Init() error {
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
	start := time.Now()
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // Due to self signed certificates in many orgs, we don't verify the cert
	f5.Authenticate()
	if f5.Token == "" {
		return fmt.Errorf("No Authentication Token. Exiting...")
	}
	
	// TODO: Use goroutines here
	if contains(f5.Collectors,"node") {
		f5.GatherNode(acc)
	}
	if contains(f5.Collectors,"virtual") {
		f5.GatherVirtual(acc)
	}
	if contains(f5.Collectors,"pool") {
		f5.GatherPool(acc)
	}
	if contains(f5.Collectors,"net_interface") {
		f5.GatherNetInterface(acc)
	}

	f5.ResetToken()
	duration := time.Since(start)
	fmt.Println(duration)

	return nil
}

func (f5 *F5LoadBalancer) GatherNode(acc telegraf.Accumulator) {
	urls := f5.GetUrls("/mgmt/tm/ltm/node")
	var wg sync.WaitGroup
	for _,url := range urls {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
			defer wg.Done()
			resp, tags, err := 	f5.GetTags(url)
			if err != nil {return}
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

func (f5 *F5LoadBalancer) GatherPool(acc telegraf.Accumulator) {
	urls := f5.GetUrls("/mgmt/tm/ltm/pool")
	var wg sync.WaitGroup
	for _,url := range urls {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
			defer wg.Done()
			resp, tags, err := 	f5.GetTags(url)
			if err != nil {return}
			base := gjson.Get(resp, "entries.*.nestedStats.entries").Raw
			fields := make(map[string]interface{})
			fields["pool_active_member_count"],_ = strconv.Atoi(gjson.Get(base,"activeMemberCnt.value").Raw)
			fields["pool_current_sessions"],_ = strconv.Atoi(gjson.Get(base, "curSessions.value").Raw)
			fields["pool_serverside_bits_in"],_ = strconv.Atoi(gjson.Get(base, "serverside\\.bitsIn.value").Raw)
			fields["pool_serverside_bits_out"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.bitsOut.value").Raw)
			fields["pool_serverside_current_connections"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.curConns.value").Raw)
			fields["pool_serverside_packets_in"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.pktsIn.value").Raw)
			fields["pool_serverside_packets_out"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.pktsOut.value").Raw)
			fields["pool_serverside_total_connections"],_ = strconv.Atoi(gjson.Get(base,"serverside\\.totConns.value").Raw)
			fields["pool_total_requests"],_ = strconv.Atoi(gjson.Get(base,"totRequests.value").Raw)
			available := 0
			if gjson.Get(base,"status\\.availabilityState.description").Str == "available"{
				available = 1
			}
			fields["pool_available"] = available

			acc.AddFields("f5_load_balancer", fields, tags)
		}(&wg, url, acc)
	}
	wg.Wait()
}

func (f5 *F5LoadBalancer) GatherVirtual(acc telegraf.Accumulator) {
	urls := f5.GetUrls("/mgmt/tm/ltm/virtual")
	var wg sync.WaitGroup
	for _,url := range urls {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
			defer wg.Done()
			resp, tags, err := 	f5.GetTags(url)		
			if err != nil {return}
			base := gjson.Get(resp, "entries.*.nestedStats.entries").Raw
			fields := make(map[string]interface{})
			fields["virtual_clientside_bits_in"],_ = strconv.Atoi(gjson.Get(base,"clientside\\.bitsIn.value").Raw)
			fields["virtual_clientside_bits_out"],_ = strconv.Atoi(gjson.Get(base, "clientside\\.bitsOut.value").Raw)
			fields["virtual_clientside_current_connections"],_ = strconv.Atoi(gjson.Get(base, "clientside\\.curConns.value").Raw)
			fields["virtual_clientside_packets_in"],_ = strconv.Atoi(gjson.Get(base,"clientside\\.pktsIn.value").Raw)
			fields["virtual_clientside_packets_out"],_ = strconv.Atoi(gjson.Get(base,"clientside\\.pktsOut.value").Raw)
			fields["virtual_total_requests"],_ = strconv.Atoi(gjson.Get(base,"totRequests.value").Raw)
			fields["virtual_one_minute_avg_usage"],_ = strconv.Atoi(gjson.Get(base,"oneMinAvgUsageRatio.value").Raw)
			available := 0
			if gjson.Get(base,"status\\.availabilityState.description").Str == "available"{
				available = 1
			}
			fields["virtual_available"] = available

			acc.AddFields("f5_load_balancer", fields, tags)
		}(&wg, url, acc)
	}
	wg.Wait()
}

func (f5 *F5LoadBalancer) GatherNetInterface(acc telegraf.Accumulator) {
	urls := f5.GetUrls("/mgmt/tm/net/interface")
	var wg sync.WaitGroup
	for _,url := range urls {
		wg.Add(1)
		go func(wg *sync.WaitGroup, url string, acc telegraf.Accumulator) {
			defer wg.Done()
			resp := f5.Call(url)
			tags := map[string]string{}
			tags["name"] = gjson.Get(resp, "entries.*.nestedStats.entries.tmName.description").Raw
			base := gjson.Get(resp, "entries.*.nestedStats.entries").Raw
			fields := make(map[string]interface{})
			fields["net_interface_counter_bits_in"],_ = strconv.Atoi(gjson.Get(base,"counters\\.bitsIn.value").Raw)
			fields["net_interface_counter_bits_out"],_ = strconv.Atoi(gjson.Get(base, "counters\\.bitsOut.value").Raw)
			fields["net_interface_counter_packets_in"],_ = strconv.Atoi(gjson.Get(base, "counters\\.pktsIn.value").Raw)
			fields["net_interface_counter_packets_out"],_ = strconv.Atoi(gjson.Get(base,"counters\\.pktsOut.value").Raw)
			status := 0
			if gjson.Get(base,"status.description").Str == "up"{
				status = 1
			}
			fields["net_interface_status"] = status

			acc.AddFields("f5_load_balancer", fields, tags)
		}(&wg, url, acc)
	}
	wg.Wait()
}

func (f5 *F5LoadBalancer) GetTags(endpoint string) (string, map[string]string, error) {
	resp := f5.Call(endpoint)
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
		// TODO: Write an actual error here
		return resp, tags, nil
	}
	return resp, tags, nil
}

func (f5 *F5LoadBalancer) GetUrls(endpoint string) ([]string) {
	resp := f5.Call(endpoint)
	urls := make([]string,0,2)
	for _,item := range gjson.Get(resp,"items").Array() {
		selfLink := gjson.Get(item.Raw,"selfLink").Str
		urls = append(urls, strings.Split(strings.Split(selfLink, "localhost")[1],"?")[0]+"/stats")
	}
	return urls
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
	// TODO: Should probably return an error with the data if the call failed
	// to prevent issues downstream
	return string(data)
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
	f5.Token = gjson.Get(jsonString, "token.token").Str
}

func (f5 *F5LoadBalancer) ResetToken() {
	f5.Token = ""
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func init() {
	inputs.Add("f5_load_balancer", func() telegraf.Input {
		return &F5LoadBalancer{}
	})
}
