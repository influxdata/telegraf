package burrow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ConsumerOffset struct {
	Offset     int64 `json:"offset"`
	Timestamp  int64 `json:"timestamp"`
	Lag        int64 `json:"lag"`
	Artificial bool  `json:"-"`
}

type PartitionStatus struct {
	Topic     string         `json:"topic"`
	Partition int32          `json:"partition"`
	Status    string         `json:"status"`
	Start     ConsumerOffset `json:"start"`
	End       ConsumerOffset `json:"end"`
}

type ConsumerGroupStatus struct {
	Cluster         string             `json:"cluster"`
	Group           string             `json:"group"`
	Status          string             `json:"status"`
	Complete        bool               `json:"complete"`
	Partitions      []*PartitionStatus `json:"partitions"`
	TotalPartitions int                `json:"partition_count"`
	Maxlag          *PartitionStatus   `json:"maxlag"`
	TotalLag        uint64             `json:"totallag"`
}

type BurrowResponseRequestInfo struct {
	URI     string `json:"url"`
	Host    string `json:"host"`
	Cluster string `json:"cluster"`
	Group   string `json:"group"`
	Topic   string `json:"topic"`
}
type HTTPResponseError struct {
	Error   bool                      `json:"error"`
	Message string                    `json:"message"`
	Result  map[string]string         `json:"result"`
	Request BurrowResponseRequestInfo `json:"request"`
}

type BurrowResponseClusterList struct {
	Error    bool                      `json:"error"`
	Message  string                    `json:"message"`
	Clusters []string                  `json:"clusters"`
	Request  BurrowResponseRequestInfo `json:"request"`
}

type BurrowResponseTopicList struct {
	Error   bool                      `json:"error"`
	Message string                    `json:"message"`
	Topics  []string                  `json:"topics"`
	Request BurrowResponseRequestInfo `json:"request"`
}

type BurrowResponseTopicDetail struct {
	Error   bool                      `json:"error"`
	Message string                    `json:"message"`
	Offsets []int64                   `json:"offsets"`
	Request BurrowResponseRequestInfo `json:"request"`
}

type BurrowResponseConsumerList struct {
	Error     bool                      `json:"error"`
	Message   string                    `json:"message"`
	Consumers []string                  `json:"consumers"`
	Request   BurrowResponseRequestInfo `json:"request"`
}

type BurrowResponseConsumerStatus struct {
	Error   bool                      `json:"error"`
	Message string                    `json:"message"`
	Status  ConsumerGroupStatus       `json:"status"`
	Request BurrowResponseRequestInfo `json:"request"`
}

type Burrow struct {
	client *http.Client

	Urls     []string
	Clusters []string
	Topics   []string
	Groups   []string
}

var sampleConfig = `
  ## Burrow HTTP endpoint urls.
  urls = ["http://burrow-service.com:8000"]
  ## Clusters to fetch data. Default to fetch all.
  #clusters = []
  ## Topics to monitor. Default to monitor all from Burrow.
  #topics = []
  ## Groups to monitor. Default to monitor all from Burrow.
  #groups = []
`

func (b *Burrow) SampleConfig() string {
	return sampleConfig
}

func (b *Burrow) Description() string {
	return "Collect Kafka topics and consumers status from Burrow's (https://github.com/linkedin/Burrow) HTTP Endpoint."
}

func (b *Burrow) Gather(acc telegraf.Accumulator) error {
	if b.client == nil {
		b.client = &http.Client{
			Transport: tr,
			Timeout:   time.Duration(3 * time.Second),
		}
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(b.Urls))

	for _, u := range b.Urls {
		addr, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("Unable to parse address '%s': %s", u, err)
		}

		wg.Add(1)
		go func(addr *url.URL) {
			defer wg.Done()
			errChan.C <- b.gatherUrl(addr, acc)
		}(addr)
	}

	wg.Wait()
	return errChan.Error()
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

func (b *Burrow) gatherUrl(baseUrl *url.URL, acc telegraf.Accumulator) error {
	var err error
	clusters := b.Clusters
	if len(clusters) == 0 {
		clusters, err = b.getClusterList(baseUrl)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(clusters) * 2)

	for _, cluster := range clusters {
		wg.Add(1)
		go func(cluster string) {
			defer wg.Done()
			errChan.C <- b.gatherClusterTopics(baseUrl, cluster, acc)
			errChan.C <- b.gatherClusterConsumers(baseUrl, cluster, acc)
		}(cluster)
	}

	wg.Wait()
	return errChan.Error()
}

func (b *Burrow) gatherClusterTopics(baseUrl *url.URL, cluster string, acc telegraf.Accumulator) error {
	var err error
	topics := b.Topics
	if len(topics) == 0 {
		topics, err = b.getTopicList(baseUrl, cluster)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(topics))
	for _, topic := range topics {
		wg.Add(1)
		go func(topic string) {
			defer wg.Done()
			errChan.C <- b.gatherTopicOffsets(baseUrl, cluster, topic, acc)
		}(topic)
	}

	wg.Wait()
	return errChan.Error()
}

func (b *Burrow) gatherClusterConsumers(baseUrl *url.URL, cluster string, acc telegraf.Accumulator) error {
	var err error
	groups := b.Groups
	if len(groups) == 0 {
		groups, err = b.getConsumerList(baseUrl, cluster)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	errChan := errchan.New(len(groups))
	for _, group := range groups {
		wg.Add(1)
		go func(group string) {
			defer wg.Done()
			errChan.C <- b.gatherConsumerStatus(baseUrl, cluster, group, acc)
		}(group)
	}

	wg.Wait()
	return errChan.Error()
}

func (b *Burrow) getClusterList(baseUrl *url.URL) ([]string, error) {
	u := fmt.Sprintf("%s/v2/kafka", baseUrl.String())
	resp, err := b.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	clusterList := &BurrowResponseClusterList{}
	err = json.NewDecoder(resp.Body).Decode(clusterList)
	if err != nil {
		return nil, err
	}

	return clusterList.Clusters, nil
}

func (b *Burrow) getTopicList(baseUrl *url.URL, cluster string) ([]string, error) {
	u := fmt.Sprintf("%s/v2/kafka/%s/topic", baseUrl.String(), cluster)
	resp, err := b.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	topicList := &BurrowResponseTopicList{}
	err = json.NewDecoder(resp.Body).Decode(topicList)
	if err != nil {
		return nil, err
	}

	return topicList.Topics, nil
}

func (b *Burrow) gatherTopicOffsets(baseUrl *url.URL, cluster string, topic string, acc telegraf.Accumulator) error {
	u := fmt.Sprintf("%s/v2/kafka/%s/topic/%s", baseUrl.String(), cluster, topic)
	resp, err := b.client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	topicDetail := &BurrowResponseTopicDetail{}
	err = json.NewDecoder(resp.Body).Decode(topicDetail)
	if err != nil {
		return err
	}

	for i, offset := range topicDetail.Offsets {
		tags := map[string]string{
			"cluster":   cluster,
			"topic":     topic,
			"partition": strconv.Itoa(i),
		}
		fields := map[string]interface{}{
			"offset": offset,
		}
		acc.AddFields("burrow_topic", fields, tags)
	}

	return nil
}

func (b *Burrow) getConsumerList(baseUrl *url.URL, cluster string) ([]string, error) {
	u := fmt.Sprintf("%s/v2/kafka/%s/consumer", baseUrl.String(), cluster)
	resp, err := b.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	consumerList := &BurrowResponseConsumerList{}
	err = json.NewDecoder(resp.Body).Decode(consumerList)
	if err != nil {
		return nil, err
	}

	return consumerList.Consumers, nil
}

func (b *Burrow) gatherConsumerStatus(baseUrl *url.URL, cluster, group string, acc telegraf.Accumulator) error {
	u := fmt.Sprintf("%s/v2/kafka/%s/consumer/%s/lag", baseUrl.String(), cluster, group)
	resp, err := b.client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	consumerStatus := &BurrowResponseConsumerStatus{}
	err = json.NewDecoder(resp.Body).Decode(consumerStatus)
	if err != nil {
		return err
	}

	for _, partition := range consumerStatus.Status.Partitions {
		tags := map[string]string{
			"cluster":   cluster,
			"group":     group,
			"topic":     partition.Topic,
			"partition": strconv.Itoa(int(partition.Partition)),
			"status":    partition.Status,
		}

		startFields := map[string]interface{}{
			"offset": partition.Start.Offset,
			"lag":    partition.Start.Lag,
		}
		acc.AddFields("burrow_consumer", startFields, tags, time.Unix(0, partition.Start.Timestamp*int64(time.Millisecond)))

		endFields := map[string]interface{}{
			"offset": partition.End.Offset,
			"lag":    partition.End.Lag,
		}
		acc.AddFields("burrow_consumer", endFields, tags, time.Unix(0, partition.End.Timestamp*int64(time.Millisecond)))
	}

	return nil
}

func init() {
	inputs.Add("burrow", func() telegraf.Input {
		return &Burrow{}
	})
}
