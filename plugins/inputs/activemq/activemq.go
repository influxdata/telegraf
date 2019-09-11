package activemq

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ActiveMQ struct {
	Server          string            `toml:"server"`
	Port            int               `toml:"port"`
	URL             string            `toml:"url"`
	Username        string            `toml:"username"`
	Password        string            `toml:"password"`
	Webadmin        string            `toml:"webadmin"`
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	tls.ClientConfig

	client  *http.Client
	baseURL *url.URL
}

type Topics struct {
	XMLName    xml.Name `xml:"topics"`
	TopicItems []Topic  `xml:"topic"`
}

type Topic struct {
	XMLName xml.Name `xml:"topic"`
	Name    string   `xml:"name,attr"`
	Stats   Stats    `xml:"stats"`
}

type Subscribers struct {
	XMLName         xml.Name     `xml:"subscribers"`
	SubscriberItems []Subscriber `xml:"subscriber"`
}

type Subscriber struct {
	XMLName          xml.Name `xml:"subscriber"`
	ClientId         string   `xml:"clientId,attr"`
	SubscriptionName string   `xml:"subscriptionName,attr"`
	ConnectionId     string   `xml:"connectionId,attr"`
	DestinationName  string   `xml:"destinationName,attr"`
	Selector         string   `xml:"selector,attr"`
	Active           string   `xml:"active,attr"`
	Stats            Stats    `xml:"stats"`
}

type Queues struct {
	XMLName    xml.Name `xml:"queues"`
	QueueItems []Queue  `xml:"queue"`
}

type Queue struct {
	XMLName xml.Name `xml:"queue"`
	Name    string   `xml:"name,attr"`
	Stats   Stats    `xml:"stats"`
}

type Stats struct {
	XMLName             xml.Name `xml:"stats"`
	Size                int      `xml:"size,attr"`
	ConsumerCount       int      `xml:"consumerCount,attr"`
	EnqueueCount        int      `xml:"enqueueCount,attr"`
	DequeueCount        int      `xml:"dequeueCount,attr"`
	PendingQueueSize    int      `xml:"pendingQueueSize,attr"`
	DispatchedQueueSize int      `xml:"dispatchedQueueSize,attr"`
	DispatchedCounter   int      `xml:"dispatchedCounter,attr"`
	EnqueueCounter      int      `xml:"enqueueCounter,attr"`
	DequeueCounter      int      `xml:"dequeueCounter,attr"`
}

var sampleConfig = `
  ## ActiveMQ WebConsole URL
  url = "http://127.0.0.1:8161"

  ## Required ActiveMQ Endpoint
  ##   deprecated in 1.11; use the url option
  # server = "127.0.0.1"
  # port = 8161

  ## Credentials for basic HTTP authentication
  # username = "admin"
  # password = "admin"

  ## Required ActiveMQ webadmin root path
  # webadmin = "admin"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
  `

func (a *ActiveMQ) Description() string {
	return "Gather ActiveMQ metrics"
}

func (a *ActiveMQ) SampleConfig() string {
	return sampleConfig
}

func (a *ActiveMQ) createHttpClient() (*http.Client, error) {
	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: a.ResponseTimeout.Duration,
	}

	return client, nil
}

func (a *ActiveMQ) Init() error {
	if a.ResponseTimeout.Duration < time.Second {
		a.ResponseTimeout.Duration = time.Second * 5
	}

	var err error
	u := &url.URL{Scheme: "http", Host: a.Server + ":" + strconv.Itoa(a.Port)}
	if a.URL != "" {
		u, err = url.Parse(a.URL)
		if err != nil {
			return err
		}
	}

	if !strings.HasPrefix(u.Scheme, "http") {
		return fmt.Errorf("invalid scheme %q", u.Scheme)
	}

	if u.Hostname() == "" {
		return fmt.Errorf("invalid hostname %q", u.Hostname())
	}

	a.baseURL = u

	a.client, err = a.createHttpClient()
	if err != nil {
		return err
	}
	return nil
}

func (a *ActiveMQ) GetMetrics(u string) ([]byte, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if a.Username != "" || a.Password != "" {
		req.SetBasicAuth(a.Username, a.Password)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned status %q", u, resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func (a *ActiveMQ) GatherQueuesMetrics(acc telegraf.Accumulator, queues Queues) {
	for _, queue := range queues.QueueItems {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = strings.TrimSpace(queue.Name)
		tags["source"] = a.baseURL.Hostname()
		tags["port"] = a.baseURL.Port()

		records["size"] = queue.Stats.Size
		records["consumer_count"] = queue.Stats.ConsumerCount
		records["enqueue_count"] = queue.Stats.EnqueueCount
		records["dequeue_count"] = queue.Stats.DequeueCount

		acc.AddFields("activemq_queues", records, tags)
	}
}

func (a *ActiveMQ) GatherTopicsMetrics(acc telegraf.Accumulator, topics Topics) {
	for _, topic := range topics.TopicItems {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = topic.Name
		tags["source"] = a.baseURL.Hostname()
		tags["port"] = a.baseURL.Port()

		records["size"] = topic.Stats.Size
		records["consumer_count"] = topic.Stats.ConsumerCount
		records["enqueue_count"] = topic.Stats.EnqueueCount
		records["dequeue_count"] = topic.Stats.DequeueCount

		acc.AddFields("activemq_topics", records, tags)
	}
}

func (a *ActiveMQ) GatherSubscribersMetrics(acc telegraf.Accumulator, subscribers Subscribers) {
	for _, subscriber := range subscribers.SubscriberItems {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["client_id"] = subscriber.ClientId
		tags["subscription_name"] = subscriber.SubscriptionName
		tags["connection_id"] = subscriber.ConnectionId
		tags["destination_name"] = subscriber.DestinationName
		tags["selector"] = subscriber.Selector
		tags["active"] = subscriber.Active
		tags["source"] = a.baseURL.Hostname()
		tags["port"] = a.baseURL.Port()

		records["pending_queue_size"] = subscriber.Stats.PendingQueueSize
		records["dispatched_queue_size"] = subscriber.Stats.DispatchedQueueSize
		records["dispatched_counter"] = subscriber.Stats.DispatchedCounter
		records["enqueue_counter"] = subscriber.Stats.EnqueueCounter
		records["dequeue_counter"] = subscriber.Stats.DequeueCounter

		acc.AddFields("activemq_subscribers", records, tags)
	}
}

func (a *ActiveMQ) Gather(acc telegraf.Accumulator) error {
	dataQueues, err := a.GetMetrics(a.QueuesURL())
	if err != nil {
		return err
	}
	queues := Queues{}
	err = xml.Unmarshal(dataQueues, &queues)
	if err != nil {
		return fmt.Errorf("queues XML unmarshal error: %v", err)
	}

	dataTopics, err := a.GetMetrics(a.TopicsURL())
	if err != nil {
		return err
	}
	topics := Topics{}
	err = xml.Unmarshal(dataTopics, &topics)
	if err != nil {
		return fmt.Errorf("topics XML unmarshal error: %v", err)
	}

	dataSubscribers, err := a.GetMetrics(a.SubscribersURL())
	if err != nil {
		return err
	}
	subscribers := Subscribers{}
	err = xml.Unmarshal(dataSubscribers, &subscribers)
	if err != nil {
		return fmt.Errorf("subscribers XML unmarshal error: %v", err)
	}

	a.GatherQueuesMetrics(acc, queues)
	a.GatherTopicsMetrics(acc, topics)
	a.GatherSubscribersMetrics(acc, subscribers)

	return nil
}

func (a *ActiveMQ) QueuesURL() string {
	ref := url.URL{Path: path.Join("/", a.Webadmin, "/xml/queues.jsp")}
	return a.baseURL.ResolveReference(&ref).String()
}

func (a *ActiveMQ) TopicsURL() string {
	ref := url.URL{Path: path.Join("/", a.Webadmin, "/xml/topics.jsp")}
	return a.baseURL.ResolveReference(&ref).String()
}

func (a *ActiveMQ) SubscribersURL() string {
	ref := url.URL{Path: path.Join("/", a.Webadmin, "/xml/subscribers.jsp")}
	return a.baseURL.ResolveReference(&ref).String()
}

func init() {
	inputs.Add("activemq", func() telegraf.Input {
		return &ActiveMQ{
			Server:   "localhost",
			Port:     8161,
			Webadmin: "admin",
		}
	})
}
