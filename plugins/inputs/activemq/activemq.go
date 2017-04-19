package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ActiveMQ struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Webadmin string `json:"webadmin"`
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

const (
	QUEUES_STATS      = "queues"
	TOPICS_STATS      = "topics"
	SUBSCRIBERS_STATS = "subscribers"
)

var sampleConfig = `
	## Required ActiveMQ Endpoint
	# server = "192.168.50.10"
	## Required ActiveMQ port
	# port = 8161
	## Required username used for request HTTP Basic Authentication
	# username = "admin"
	## Required password used for HTTP Basic Authentication
	# password = "admin"
	## Required ActiveMQ webadmin root path
	# webadmin = "admin"
	`

func (a *ActiveMQ) Description() string {
	return "Gather ActiveMQ metrics"
}

func (a *ActiveMQ) SampleConfig() string {
	return sampleConfig
}

func (a *ActiveMQ) GetMetrics(keyword string) ([]byte, error) {
	client := &http.Client{}

	url := fmt.Sprintf("http://%s:%d/%s/xml/%s.jsp", a.Server, a.Port, a.Webadmin, keyword)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(a.Username, a.Password)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (a *ActiveMQ) GatherQueuesMetrics(acc telegraf.Accumulator, queues Queues) {
	for _, queue := range queues.QueueItems {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = queue.Name

		records["size"] = queue.Stats.Size
		records["consumer_count"] = queue.Stats.ConsumerCount
		records["enqueue_count"] = queue.Stats.EnqueueCount
		records["dequeue_count"] = queue.Stats.DequeueCount

		acc.AddFields("queues_metrics", records, tags)
	}
}

func (a *ActiveMQ) GatherTopicsMetrics(acc telegraf.Accumulator, topics Topics) {
	for _, topic := range topics.TopicItems {
		records := make(map[string]interface{})
		tags := make(map[string]string)

		tags["name"] = topic.Name

		records["size"] = topic.Stats.Size
		records["consumer_count"] = topic.Stats.ConsumerCount
		records["enqueue_count"] = topic.Stats.EnqueueCount
		records["dequeue_count"] = topic.Stats.DequeueCount

		acc.AddFields("topics_metrics", records, tags)
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

		records["pending_queue_size"] = subscriber.Stats.PendingQueueSize
		records["dispatched_queue_size"] = subscriber.Stats.DispatchedQueueSize
		records["dispatched_counter"] = subscriber.Stats.DispatchedCounter
		records["enqueue_counter"] = subscriber.Stats.EnqueueCounter
		records["dequeue_counter"] = subscriber.Stats.DequeueCounter

		acc.AddFields("subscribers_metrics", records, tags)
	}
}

func (a *ActiveMQ) Gather(acc telegraf.Accumulator) error {
	dataQueues, err := a.GetMetrics(QUEUES_STATS)

	queues := Queues{}

	err = xml.Unmarshal(dataQueues, &queues)

	if err != nil {
		return err
	}

	dataTopics, err := a.GetMetrics(TOPICS_STATS)

	topics := Topics{}

	err = xml.Unmarshal(dataTopics, &topics)

	if err != nil {
		return err
	}

	dataSubscribers, err := a.GetMetrics(SUBSCRIBERS_STATS)

	subscribers := Subscribers{}

	err = xml.Unmarshal(dataSubscribers, &subscribers)

	if err != nil {
		return err
	}

	a.GatherQueuesMetrics(acc, queues)

	a.GatherTopicsMetrics(acc, topics)

	a.GatherSubscribersMetrics(acc, subscribers)

	return nil
}

func init() {
	inputs.Add("activemq", func() telegraf.Input { return &ActiveMQ{} })
}
