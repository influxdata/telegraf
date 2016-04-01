package cassandra

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	"net/url"
	//"reflect"
	"strings"
)

type Server struct {
	Host     string
	Username string
	Password string
	Port     string
}

type Metric struct {
	Jmx string
}

type JolokiaClient interface {
	MakeRequest(req *http.Request) (*http.Response, error)
}

type JolokiaClientImpl struct {
	client *http.Client
}

func (c JolokiaClientImpl) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

type Cassandra struct {
	jClient JolokiaClient
	Context string
	Servers []Server
	Metrics []Metric
}

type javaMetric struct {
	server Server
	metric Metric
	acc    telegraf.Accumulator
}

type cassandraMetric struct {
	server Server
	metric Metric
	acc    telegraf.Accumulator
}

type jmxMetric interface {
	addTagsFields(out map[string]interface{})
}

func addServerTags(server Server, tags map[string]string) {
	if server.Host != "" && server.Host != "localhost" &&
		server.Host != "127.0.0.1" {
		tags["host"] = server.Host
	}
}

func newJavaMetric(server Server, metric Metric,
	acc telegraf.Accumulator) *javaMetric {
	return &javaMetric{server: server, metric: metric, acc: acc}
}

func newCassandraMetric(server Server, metric Metric,
	acc telegraf.Accumulator) *cassandraMetric {
	return &cassandraMetric{server: server, metric: metric, acc: acc}
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
	if classAndPairs[0] == "org.apache.cassandra.metrics" {
		tokens["class"] = "cassandra"
	} else if classAndPairs[0] == "java.lang" {
		tokens["class"] = "java"
	} else {
		return tokens
	}
	pairs := strings.Split(classAndPairs[1], ",")
	for _, pair := range pairs {
		p := strings.Split(pair, "=")
		tokens[p[0]] = p[1]
	}
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

func (j javaMetric) addTagsFields(out map[string]interface{}) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	a := out["request"].(map[string]interface{})
	attribute := a["attribute"].(string)
	mbean := a["mbean"].(string)

	tokens := parseJmxMetricRequest(mbean)
	addTokensToTags(tokens, tags)
	addServerTags(j.server, tags)

	if _, ok := tags["mname"]; !ok {
		//Queries for a single value will not return a "name" tag in the response.
		tags["mname"] = attribute
	}

	if values, ok := out["value"]; ok {
		switch t := values.(type) {
		case map[string]interface{}:
			addValuesAsFields(values.(map[string]interface{}), fields, attribute)
		case interface{}:
			fields[attribute] = t
		}
		j.acc.AddFields(tokens["class"]+tokens["type"], fields, tags)
	} else {
		fmt.Printf("Missing key 'value' in '%s' output response\n%v\n",
			j.metric.Jmx, out)
	}
}

func addCassandraMetric(mbean string, c cassandraMetric,
	values map[string]interface{}) {

	tags := make(map[string]string)
	fields := make(map[string]interface{})
	tokens := parseJmxMetricRequest(mbean)
	addTokensToTags(tokens, tags)
	addServerTags(c.server, tags)
	addValuesAsFields(values, fields, tags["mname"])
	c.acc.AddFields(tokens["class"]+tokens["type"], fields, tags)

}

func (c cassandraMetric) addTagsFields(out map[string]interface{}) {

	r := out["request"]

	tokens := parseJmxMetricRequest(r.(map[string]interface{})["mbean"].(string))
	// Requests with wildcards for keyspace or table names will return nested
	// maps in the json response
	if tokens["type"] == "Table" && (tokens["keyspace"] == "*" ||
		tokens["scope"] == "*") {
		if valuesMap, ok := out["value"]; ok {
			for k, v := range valuesMap.(map[string]interface{}) {
				addCassandraMetric(k, c, v.(map[string]interface{}))
			}
		} else {
			fmt.Printf("Missing key 'value' in '%s' output response\n%v\n",
				c.metric.Jmx, out)
			return
		}
	} else {
		if values, ok := out["value"]; ok {
			addCassandraMetric(r.(map[string]interface{})["mbean"].(string),
				c, values.(map[string]interface{}))
		} else {
			fmt.Printf("Missing key 'value' in '%s' output response\n%v\n",
				c.metric.Jmx, out)
			return
		}
	}
}

func (j *Cassandra) SampleConfig() string {
	return `
  # This is the context root used to compose the jolokia url
  context = "/jolokia/read"

  # List of cassandra servers exposing jolokia read service
  [[cassandra.servers]]
    # host can be skipped for localhost. host tag will be set to hostname()
    host = "192.168.103.2"
    port = "8180"
    # username = "myuser"
    # password = "mypassword"

  # List of metrics collected on above servers
  # Each metric consists of a jmx path. Pass or drop slice attributes will be
  # supported in the future.
  # This will collect all heap memory usage metrics from the jvm
  [[cassandra..metrics]]
    jmx  = "/java.lang:type=Memory/HeapMemoryUsage"

  # This will collect ReadLatency metrics for all keyspaces and tables.
  # "type=Table" in the query works with Cassandra3.0. Older versions might need
  # to use "type=ColumnFamily"
  [[cassandra..metrics]]
    jmx  = "/org.apache.cassandra.metrics:type=Table,keyspace=*,scope=*,name=ReadL
atency"
`
}

func (j *Cassandra) Description() string {
	return "Read Cassandra metrics through Jolokia"
}

func (j *Cassandra) getAttr(requestUrl *url.URL) (map[string]interface{}, error) {
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

func (c *Cassandra) Gather(acc telegraf.Accumulator) error {
	context := c.Context
	servers := c.Servers
	metrics := c.Metrics

	for _, server := range servers {
		for _, metric := range metrics {
			var m jmxMetric

			if strings.HasPrefix(metric.Jmx, "/java.lang:") {
				m = newJavaMetric(server, metric, acc)
			} else if strings.HasPrefix(metric.Jmx,
				"/org.apache.cassandra.metrics:") {
				m = newCassandraMetric(server, metric, acc)
			}
			jmxPath := metric.Jmx

			// Prepare URL
			requestUrl, err := url.Parse("http://" + server.Host + ":" +
				server.Port + context + jmxPath)
			fmt.Printf("host %s url %s\n", server.Host, requestUrl)
			if err != nil {
				return err
			}
			if server.Username != "" || server.Password != "" {
				requestUrl.User = url.UserPassword(server.Username, server.Password)
			}

			out, err := c.getAttr(requestUrl)
			if out["status"] != 200.0 {
				fmt.Printf("URL returned with status %v\n", out["status"])
				continue
			}
			m.addTagsFields(out)
		}
	}
	return nil
}

func init() {
	inputs.Add("cassandra", func() telegraf.Input {
		return &Cassandra{jClient: &JolokiaClientImpl{client: &http.Client{}}}
	})
}
