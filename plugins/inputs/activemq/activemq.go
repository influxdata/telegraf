//go:generate ../../../tools/readme_config_includer/generator
package activemq

import (
	_ "embed"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type ActiveMQ struct {
	Server          string          `toml:"server" deprecated:"1.11.0;use 'url' instead"`
	Port            int             `toml:"port" deprecated:"1.11.0;use 'url' instead"`
	URL             string          `toml:"url"`
	Username        string          `toml:"username"`
	Password        string          `toml:"password"`
	Webadmin        string          `toml:"webadmin"`
	ResponseTimeout config.Duration `toml:"response_timeout"`
	tls.ClientConfig

	client  *http.Client
	baseURL *url.URL
}

type topics struct {
	XMLName    xml.Name `xml:"topics"`
	TopicItems []topic  `xml:"topic"`
}

type topic struct {
	XMLName xml.Name `xml:"topic"`
	Name    string   `xml:"name,attr"`
	Stats   stats    `xml:"stats"`
}

type subscribers struct {
	XMLName         xml.Name     `xml:"subscribers"`
	SubscriberItems []subscriber `xml:"subscriber"`
}

type subscriber struct {
	XMLName          xml.Name `xml:"subscriber"`
	ClientID         string   `xml:"clientId,attr"`
	SubscriptionName string   `xml:"subscriptionName,attr"`
	ConnectionID     string   `xml:"connectionId,attr"`
	DestinationName  string   `xml:"destinationName,attr"`
	Selector         string   `xml:"selector,attr"`
	Active           string   `xml:"active,attr"`
	Stats            stats    `xml:"stats"`
}

type queues struct {
	XMLName    xml.Name `xml:"queues"`
	QueueItems []queue  `xml:"queue"`
}

type queue struct {
	XMLName xml.Name `xml:"queue"`
	Name    string   `xml:"name,attr"`
	Stats   stats    `xml:"stats"`
}

type stats struct {
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

func (*ActiveMQ) SampleConfig() string {
	return sampleConfig
}

func (a *ActiveMQ) Init() error {
	if a.ResponseTimeout < config.Duration(time.Second) {
		a.ResponseTimeout = config.Duration(time.Second * 5)
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

	a.client, err = a.createHTTPClient()
	if err != nil {
		return err
	}
	return nil
}

func (a *ActiveMQ) Gather(acc telegraf.Accumulator) error {
	dataQueues, err := a.getMetrics(a.queuesURL())
	if err != nil {
		return err
	}
	queues := queues{}
	err = xml.Unmarshal(dataQueues, &queues)
	if err != nil {
		return fmt.Errorf("queues XML unmarshal error: %w", err)
	}

	dataTopics, err := a.getMetrics(a.topicsURL())
	if err != nil {
		return err
	}
	topics := topics{}
	err = xml.Unmarshal(dataTopics, &topics)
	if err != nil {
		return fmt.Errorf("topics XML unmarshal error: %w", err)
	}

	dataSubscribers, err := a.getMetrics(a.subscribersURL())
	if err != nil {
		return err
	}
	subscribers := subscribers{}
	err = xml.Unmarshal(dataSubscribers, &subscribers)
	if err != nil {
		return fmt.Errorf("subscribers XML unmarshal error: %w", err)
	}

	a.gatherQueuesMetrics(acc, queues)
	a.gatherTopicsMetrics(acc, topics)
	a.gatherSubscribersMetrics(acc, subscribers)

	return nil
}

func (a *ActiveMQ) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: time.Duration(a.ResponseTimeout),
	}

	return client, nil
}

func (a *ActiveMQ) getMetrics(u string) ([]byte, error) {
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
		return nil, fmt.Errorf("%s returned HTTP status %s", u, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func (a *ActiveMQ) gatherQueuesMetrics(acc telegraf.Accumulator, queues queues) {
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

func (a *ActiveMQ) gatherTopicsMetrics(acc telegraf.Accumulator, topics topics) {
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

func (a *ActiveMQ) gatherSubscribersMetrics(acc telegraf.Accumulator, subscribers subscribers) {
	for _, subscriber := range subscribers.SubscriberItems {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["client_id"] = subscriber.ClientID
		tags["subscription_name"] = subscriber.SubscriptionName
		tags["connection_id"] = subscriber.ConnectionID
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

func (a *ActiveMQ) queuesURL() string {
	ref := url.URL{Path: path.Join("/", a.Webadmin, "/xml/queues.jsp")}
	return a.baseURL.ResolveReference(&ref).String()
}

func (a *ActiveMQ) topicsURL() string {
	ref := url.URL{Path: path.Join("/", a.Webadmin, "/xml/topics.jsp")}
	return a.baseURL.ResolveReference(&ref).String()
}

func (a *ActiveMQ) subscribersURL() string {
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
