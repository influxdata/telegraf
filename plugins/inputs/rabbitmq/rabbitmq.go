package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const DefaultUsername = "guest"
const DefaultPassword = "guest"
const DefaultURL = "http://localhost:15672"

type RabbitMQ struct {
	URL      string
	Name     string
	Username string
	Password string
	Nodes    []string
	Queues   []string

	Client *http.Client
}

type OverviewResponse struct {
	MessageStats *MessageStats `json:"message_stats"`
	ObjectTotals *ObjectTotals `json:"object_totals"`
	QueueTotals  *QueueTotals  `json:"queue_totals"`
}

type Details struct {
	Rate float64
}

type MessageStats struct {
	Ack               int64
	AckDetails        Details `json:"ack_details"`
	Deliver           int64
	DeliverDetails    Details `json:"deliver_details"`
	DeliverGet        int64
	DeliverGetDetails Details `json:"deliver_get_details"`
	Publish           int64
	PublishDetails    Details `json:"publish_details"`
	Redeliver         int64
	RedeliverDetails  Details `json:"redeliver_details"`
}

type ObjectTotals struct {
	Channels    int64
	Connections int64
	Consumers   int64
	Exchanges   int64
	Queues      int64
}

type QueueTotals struct {
	Messages                   int64
	MessagesReady              int64 `json:"messages_ready"`
	MessagesUnacknowledged     int64 `json:"messages_unacknowledged"`
	MessageBytes               int64 `json:"message_bytes"`
	MessageBytesReady          int64 `json:"message_bytes_ready"`
	MessageBytesUnacknowledged int64 `json:"message_bytes_unacknowledged"`
	MessageRam                 int64 `json:"message_bytes_ram"`
	MessagePersistent          int64 `json:"message_bytes_persistent"`
}

type Queue struct {
	QueueTotals         // just to not repeat the same code
	MessageStats        `json:"message_stats"`
	Memory              int64
	Consumers           int64
	ConsumerUtilisation float64 `json:"consumer_utilisation"`
	Name                string
	Node                string
	Vhost               string
	Durable             bool
	AutoDelete          bool `json:"auto_delete"`
}

type Node struct {
	Name string

	DiskFree      int64 `json:"disk_free"`
	DiskFreeLimit int64 `json:"disk_free_limit"`
	FdTotal       int64 `json:"fd_total"`
	FdUsed        int64 `json:"fd_used"`
	MemLimit      int64 `json:"mem_limit"`
	MemUsed       int64 `json:"mem_used"`
	ProcTotal     int64 `json:"proc_total"`
	ProcUsed      int64 `json:"proc_used"`
	RunQueue      int64 `json:"run_queue"`
	SocketsTotal  int64 `json:"sockets_total"`
	SocketsUsed   int64 `json:"sockets_used"`
}

type gatherFunc func(r *RabbitMQ, acc telegraf.Accumulator, errChan chan error)

var gatherFunctions = []gatherFunc{gatherOverview, gatherNodes, gatherQueues}

var sampleConfig = `
  # url = "http://localhost:15672"
  # name = "rmq-server-1" # optional tag
  # username = "guest"
  # password = "guest"

  ## A list of nodes to pull metrics about. If not specified, metrics for
  ## all nodes are gathered.
  # nodes = ["rabbit@node1", "rabbit@node2"]
`

func (r *RabbitMQ) SampleConfig() string {
	return sampleConfig
}

func (r *RabbitMQ) Description() string {
	return "Read metrics from one or many RabbitMQ servers via the management API"
}

func (r *RabbitMQ) Gather(acc telegraf.Accumulator) error {
	if r.Client == nil {
		tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
		r.Client = &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(gatherFunctions))
	errChan := errchan.New(len(gatherFunctions))
	for _, f := range gatherFunctions {
		go func(gf gatherFunc) {
			defer wg.Done()
			gf(r, acc, errChan.C)
		}(f)
	}
	wg.Wait()

	return errChan.Error()
}

func (r *RabbitMQ) requestJSON(u string, target interface{}) error {
	if r.URL == "" {
		r.URL = DefaultURL
	}
	u = fmt.Sprintf("%s%s", r.URL, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	username := r.Username
	if username == "" {
		username = DefaultUsername
	}

	password := r.Password
	if password == "" {
		password = DefaultPassword
	}

	req.SetBasicAuth(username, password)

	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(target)

	return nil
}

func gatherOverview(r *RabbitMQ, acc telegraf.Accumulator, errChan chan error) {
	overview := &OverviewResponse{}

	err := r.requestJSON("/api/overview", &overview)
	if err != nil {
		errChan <- err
		return
	}

	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil {
		errChan <- fmt.Errorf("Wrong answer from rabbitmq. Probably auth issue")
		return
	}

	tags := map[string]string{"url": r.URL}
	if r.Name != "" {
		tags["name"] = r.Name
	}
	fields := map[string]interface{}{
		"messages":           overview.QueueTotals.Messages,
		"messages_ready":     overview.QueueTotals.MessagesReady,
		"messages_unacked":   overview.QueueTotals.MessagesUnacknowledged,
		"channels":           overview.ObjectTotals.Channels,
		"connections":        overview.ObjectTotals.Connections,
		"consumers":          overview.ObjectTotals.Consumers,
		"exchanges":          overview.ObjectTotals.Exchanges,
		"queues":             overview.ObjectTotals.Queues,
		"messages_acked":     overview.MessageStats.Ack,
		"messages_delivered": overview.MessageStats.Deliver,
		"messages_published": overview.MessageStats.Publish,
	}
	acc.AddFields("rabbitmq_overview", fields, tags)

	errChan <- nil
}

func gatherNodes(r *RabbitMQ, acc telegraf.Accumulator, errChan chan error) {
	nodes := make([]Node, 0)
	// Gather information about nodes
	err := r.requestJSON("/api/nodes", &nodes)
	if err != nil {
		errChan <- err
		return
	}
	now := time.Now()

	for _, node := range nodes {
		if !r.shouldGatherNode(node) {
			continue
		}

		tags := map[string]string{"url": r.URL}
		tags["node"] = node.Name

		fields := map[string]interface{}{
			"disk_free":       node.DiskFree,
			"disk_free_limit": node.DiskFreeLimit,
			"fd_total":        node.FdTotal,
			"fd_used":         node.FdUsed,
			"mem_limit":       node.MemLimit,
			"mem_used":        node.MemUsed,
			"proc_total":      node.ProcTotal,
			"proc_used":       node.ProcUsed,
			"run_queue":       node.RunQueue,
			"sockets_total":   node.SocketsTotal,
			"sockets_used":    node.SocketsUsed,
		}
		acc.AddFields("rabbitmq_node", fields, tags, now)
	}

	errChan <- nil
}

func gatherQueues(r *RabbitMQ, acc telegraf.Accumulator, errChan chan error) {
	// Gather information about queues
	queues := make([]Queue, 0)
	err := r.requestJSON("/api/queues", &queues)
	if err != nil {
		errChan <- err
		return
	}

	for _, queue := range queues {
		if !r.shouldGatherQueue(queue) {
			continue
		}
		tags := map[string]string{
			"url":         r.URL,
			"queue":       queue.Name,
			"vhost":       queue.Vhost,
			"node":        queue.Node,
			"durable":     strconv.FormatBool(queue.Durable),
			"auto_delete": strconv.FormatBool(queue.AutoDelete),
		}

		acc.AddFields(
			"rabbitmq_queue",
			map[string]interface{}{
				// common information
				"consumers":            queue.Consumers,
				"consumer_utilisation": queue.ConsumerUtilisation,
				"memory":               queue.Memory,
				// messages information
				"message_bytes":             queue.MessageBytes,
				"message_bytes_ready":       queue.MessageBytesReady,
				"message_bytes_unacked":     queue.MessageBytesUnacknowledged,
				"message_bytes_ram":         queue.MessageRam,
				"message_bytes_persist":     queue.MessagePersistent,
				"messages":                  queue.Messages,
				"messages_ready":            queue.MessagesReady,
				"messages_unack":            queue.MessagesUnacknowledged,
				"messages_ack":              queue.MessageStats.Ack,
				"messages_ack_rate":         queue.MessageStats.AckDetails.Rate,
				"messages_deliver":          queue.MessageStats.Deliver,
				"messages_deliver_rate":     queue.MessageStats.DeliverDetails.Rate,
				"messages_deliver_get":      queue.MessageStats.DeliverGet,
				"messages_deliver_get_rate": queue.MessageStats.DeliverGetDetails.Rate,
				"messages_publish":          queue.MessageStats.Publish,
				"messages_publish_rate":     queue.MessageStats.PublishDetails.Rate,
				"messages_redeliver":        queue.MessageStats.Redeliver,
				"messages_redeliver_rate":   queue.MessageStats.RedeliverDetails.Rate,
			},
			tags,
		)
	}

	errChan <- nil
}

func (r *RabbitMQ) shouldGatherNode(node Node) bool {
	if len(r.Nodes) == 0 {
		return true
	}

	for _, name := range r.Nodes {
		if name == node.Name {
			return true
		}
	}

	return false
}

func (r *RabbitMQ) shouldGatherQueue(queue Queue) bool {
	if len(r.Queues) == 0 {
		return true
	}

	for _, name := range r.Queues {
		if name == queue.Name {
			return true
		}
	}

	return false
}

func init() {
	inputs.Add("rabbitmq", func() telegraf.Input {
		return &RabbitMQ{}
	})
}
