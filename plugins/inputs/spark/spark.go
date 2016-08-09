package spark

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type JolokiaClient interface {
	MakeRequest(req *http.Request) (*http.Response, error)
}

type JolokiaClientImpl struct {
	client *http.Client
}

func (c JolokiaClientImpl) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

type YarnClient interface {
	MakeRequest(req *http.Request) (*http.Response, error)
}

type YarnClientImpl struct {
	client *http.Client
}

func (c YarnClientImpl) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

type Spark struct {
	jClient     JolokiaClient
	SparkServer []string
	YarnServer  string
}

type javaMetric struct {
	host   string
	metric string
	acc    telegraf.Accumulator
}

type sparkMetric struct {
	host   string
	metric string
	acc    telegraf.Accumulator
}

type Yarn struct {
	yClient       YarnClient
	serverAddress string
}

type yarnMetric struct {
	host string
	acc  telegraf.Accumulator
}

type jmxMetric interface {
	addTagsFields(out map[string]interface{})
}

func newJavaMetric(host string, metric string,
	acc telegraf.Accumulator) *javaMetric {
	return &javaMetric{host: host, metric: metric, acc: acc}
}

func newSparkMetric(host string, metric string,
	acc telegraf.Accumulator) *sparkMetric {
	return &sparkMetric{host: host, metric: metric, acc: acc}
}

func newYarnMetric(host string, acc telegraf.Accumulator) *yarnMetric {
	return &yarnMetric{host: host, acc: acc}
}

func addValuesAsFields(values map[string]interface{}, fields map[string]interface{},
	mname string) {
	for k, v := range values {
		if v != nil {
			fields[mname+"_"+k] = v
		}
	}
}

func parseJmxMetricRequest(mbean string) map[string]string {
	tokens := make(map[string]string)
	classAndPairs := strings.Split(mbean, ":")
	if classAndPairs[0] == "metrics" {
		tokens["class"] = "spark_jolokiaMetrics"
	} else if classAndPairs[0] == "java.lang" {
		tokens["class"] = "java"
	} else {
		return tokens
	}

	pair := strings.Split(classAndPairs[1], "=")
	tokens[pair[0]] = pair[1]

	return tokens
}

func addTokensToTags(tokens map[string]string, tags map[string]string) {
	for k, v := range tokens {
		if k == "name" {
			tags["mname"] = v // name seems to a reserved word in influxdb
		} else if k == "class" || k == "type" {
			continue // class and type are used in the metric name
		} else {
			tags[k] = v
		}
	}
}

func addJavaMetric(class string, c javaMetric,
	values map[string]interface{}) {

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	tags["spark_host"] = c.host
	tags["spark_class"] = class

	if class == "spark_Threading" {
		list := []string{"PeakThreadCount", "CurrentThreadCpuTime", "DaemonThreadCount", "TotalStartedThreadCount", "CurrentThreadUserTime", "ThreadCount"}
		for _, value := range list {
			if values[value] != nil {
				fields[value] = values[value]
			}
		}
	} else {
		for k, v := range values {
			if v != nil {
				fields[k] = v
			}
		}
	}
	c.acc.AddFields(class, fields, tags)

}

func (j javaMetric) addTagsFields(out map[string]interface{}) {
	fmt.Println(out["request"])
	request := out["request"].(map[string]interface{})
	var mbean = request["mbean"].(string)
	var mbeansplit = strings.Split(mbean, "=")
	var class = mbeansplit[1]

	if valuesMap, ok := out["value"]; ok {
		if class == "Memory" {
			addJavaMetric("spark_HeapMemoryUsage", j, valuesMap.(map[string]interface{}))
		} else if class == "Threading" {
			addJavaMetric("spark_Threading", j, valuesMap.(map[string]interface{}))
		} else {
			fmt.Printf("Missing key in '%s' output response\n%v\n",
				j.metric, out)
			return
		}

	}
}

func addSparkMetric(mbean string, c sparkMetric,
	values map[string]interface{}) {

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tokens := parseJmxMetricRequest(mbean)
	addTokensToTags(tokens, tags)
	tags["spark_host"] = c.host

	addValuesAsFields(values, fields, tags["mname"])
	c.acc.AddFields(tokens["class"]+tokens["type"], fields, tags)

}

func (c sparkMetric) addTagsFields(out map[string]interface{}) {
	if valuesMap, ok := out["value"]; ok {
		for k, v := range valuesMap.(map[string]interface{}) {
			addSparkMetric(k, c, v.(map[string]interface{}))
		}
	} else {
		fmt.Printf("Missing key 'value' in '%s' output response\n%v\n",
			c.metric, out)
		return
	}

}

func addYarnMetric(c yarnMetric, value map[string]interface{}, metrictype string) {

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	tags["yarn_host"] = c.host
	for key, val := range value {
		fields[key] = val
	}
	c.acc.AddFields(metrictype, fields, tags)
}

func (c yarnMetric) addTagsFields(out map[string]interface{}) {

	if valuesMap, ok := out["clusterMetrics"]; ok {
		addYarnMetric(c, valuesMap.(map[string]interface{}), "spark_clusterMetrics")
	} else if valuesMap, ok := out["clusterInfo"]; ok {
		addYarnMetric(c, valuesMap.(map[string]interface{}), "spark_clusterInfo")
	} else if valuesMap, ok := out["apps"]; ok {
		for _, value := range valuesMap.(map[string]interface{}) {
			for _, vv := range value.([]interface{}) {
				addYarnMetric(c, vv.(map[string]interface{}), "spark_apps")
			}
		}
	} else if valuesMap, ok := out["nodes"]; ok {
		for _, value := range valuesMap.(map[string]interface{}) {
			for _, vv := range value.([]interface{}) {
				addYarnMetric(c, vv.(map[string]interface{}), "spark_nodes")
			}
		}
	} else {
		fmt.Printf("Missing the required key in output response\n%v\n", out)
		return
	}

}

func (j *Spark) SampleConfig() string {
	return `
  ## Spark server exposing jolokia read service
  SparkServer = ["127.0.0.1:8778"] #optional
  ## Server running Yarn Resource Manager
  YarnServer = "127.0.0.1:8088" #optional
`
}

func (j *Spark) Description() string {
	return "Read Spark metrics through Jolokia and Yarn"
}

func (j *Spark) getAttr(requestUrl *url.URL) (map[string]interface{}, error) {
	// Create + send request
	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := j.jClient.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestUrl,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Unmarshal json
	var jsonOut map[string]interface{}
	if err = json.Unmarshal([]byte(body), &jsonOut); err != nil {
		return nil, errors.New("Error decoding JSON response")
	}

	return jsonOut, nil
}

func (j *Yarn) getAttr(requestUrl *url.URL) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := j.yClient.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestUrl,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	// read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal json
	var jsonOut map[string]interface{}
	if err = json.Unmarshal([]byte(body), &jsonOut); err != nil {
		return nil, errors.New("Error decoding JSON response")
	}

	return jsonOut, nil
}

func parseServerTokens(server string) map[string]string {
	serverTokens := make(map[string]string)
	log.Printf("Parsing %s", server)
	hostAndUser := strings.Split(server, "@")
	hostPort := ""

	if len(hostAndUser) == 1 {
		hostPort = hostAndUser[0]
	} else {
		log.Printf("Unsupported Server info, skipping")
		return nil
	}
	log.Printf("%s \n", hostPort)
	hostTokens := strings.Split(hostPort, ":")
	serverTokens["host"] = hostTokens[0]
	serverTokens["port"] = hostTokens[1]
	return serverTokens
}

func (c *Spark) GatherJolokia(acc telegraf.Accumulator, wg *sync.WaitGroup) error {
	context := "/jolokia/read"
	servers := c.SparkServer
	metrics := [...]string{"/metrics:*", "/java.lang:type=Memory/HeapMemoryUsage", "/java.lang:type=Threading"}
	if len(servers) == 0 {
		wg.Done()
		return nil
	}
	for _, server := range servers {
		for _, metric := range metrics {
			serverTokens := parseServerTokens(server)
			var m jmxMetric
			if strings.HasPrefix(metric, "/java.lang:") {
				m = newJavaMetric(serverTokens["host"], metric, acc)
			} else if strings.HasPrefix(metric, "/metrics:") {
				m = newSparkMetric(serverTokens["host"], metric, acc)
			} else {
				log.Printf("Unsupported Spark metric [%s], skipping",
					metric)
				continue
			}

			requestUrl, err := url.Parse("http://" + serverTokens["host"] + ":" +
				serverTokens["port"] + context + metric)

			if err != nil {
				return err
			}
			fmt.Println("Request url is   ", requestUrl)

			out, err := c.getAttr(requestUrl)
			if len(out) == 0 {
				continue
			}
			m.addTagsFields(out)
		}
	}

	wg.Done()
	return nil
}

func (c *Yarn) GatherYarn(acc telegraf.Accumulator, wg *sync.WaitGroup) error {
	contexts := [...]string{"/ws/v1/cluster", "/ws/v1/cluster/metrics", "/ws/v1/cluster/apps", "/ws/v1/cluster/nodes"}
	server := c.serverAddress

	if server == "" {
		wg.Done()
		return nil
	}

	fmt.Println("Going to collect data of server ", server)

	serverTokens := parseServerTokens(server)
	for _, context := range contexts {
		var m = newYarnMetric(server, acc)
		requestUrl, err := url.Parse("http://" + serverTokens["host"] + ":" + serverTokens["port"] + context)
		if err != nil {
			return err
		}

		out, err := c.getAttr(requestUrl)
		if len(out) == 0 {
			continue
		}
		m.addTagsFields(out)

	}
	wg.Done()
	return nil
}

func (c *Spark) Gather(acc telegraf.Accumulator) error {

	log.Println("Config is ", c)
	yarn := Yarn{
		yClient:       &YarnClientImpl{client: &http.Client{}},
		serverAddress: c.YarnServer,
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go yarn.GatherYarn(acc, &wg)
	spark := Spark{
		jClient:     &JolokiaClientImpl{client: &http.Client{}},
		SparkServer: c.SparkServer,
	}
	wg.Add(1)
	go spark.GatherJolokia(acc, &wg)
	wg.Wait()
	return nil
}

func init() {
	inputs.Add("spark", func() telegraf.Input {
		return &Spark{jClient: &JolokiaClientImpl{client: &http.Client{}}}
	})
}
