package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/influxdb/telegraf/plugins"
)

const DefaultUsername = "guest"
const DefaultPassword = "guest"
const DefaultURL = "http://localhost:15672"

type Server struct {
	URL      string
	Name     string
	Username string
	Password string
	Nodes    []string
	Queues   []string
}

type RabbitMQ struct {
	Servers []*Server

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
	Messages               int64
	MessagesReady          int64 `json:"messages_ready"`
	MessagesUnacknowledged int64 `json:"messages_unacknowledged"`
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

type gatherFunc func(r *RabbitMQ, serv *Server, acc plugins.Accumulator, errChan chan error)

var gatherFunctions = []gatherFunc{gatherOverview, gatherNodes, gatherQueues}

var sampleConfig = `
  # Specify servers via an array of tables
  [[plugins.rabbitmq.servers]]
  # name = "rmq-server-1" # optional tag
  # url = "http://localhost:15672"
  # username = "guest"
  # password = "guest"

  # A list of nodes to pull metrics about. If not specified, metrics for
  # all nodes are gathered.
  # nodes = ["rabbit@node1", "rabbit@node2"]
`

func (r *RabbitMQ) SampleConfig() string {
	return sampleConfig
}

func (r *RabbitMQ) Description() string {
	return "Read metrics from one or many RabbitMQ servers via the management API"
}

var localhost = &Server{URL: DefaultURL}

func (r *RabbitMQ) Gather(acc plugins.Accumulator) error {
	if r.Client == nil {
		r.Client = &http.Client{}
	}

	var errChan = make(chan error, len(r.Servers))

	// use localhost is no servers are specified in config
	if len(r.Servers) == 0 {
		r.Servers = append(r.Servers, localhost)
	}

	for _, serv := range r.Servers {
		for _, f := range gatherFunctions {
			go f(r, serv, acc, errChan)
		}
	}

	for i := 1; i <= len(r.Servers)*len(gatherFunctions); i++ {
		err := <-errChan
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) requestJSON(serv *Server, u string, target interface{}) error {
	u = fmt.Sprintf("%s%s", serv.URL, u)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}

	username := serv.Username
	if username == "" {
		username = DefaultUsername
	}

	password := serv.Password
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

func gatherOverview(r *RabbitMQ, serv *Server, acc plugins.Accumulator, errChan chan error) {
	overview := &OverviewResponse{}

	err := r.requestJSON(serv, "/api/overview", &overview)
	if err != nil {
		errChan <- err
		return
	}

	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil {
		errChan <- fmt.Errorf("Wrong answer from rabbitmq. Probably auth issue")
		return
	}

	tags := map[string]string{"url": serv.URL}
	if serv.Name != "" {
		tags["name"] = serv.Name
	}

	acc.Add("messages", overview.QueueTotals.Messages, tags)
	acc.Add("messages_ready", overview.QueueTotals.MessagesReady, tags)
	acc.Add("messages_unacked", overview.QueueTotals.MessagesUnacknowledged, tags)

	acc.Add("channels", overview.ObjectTotals.Channels, tags)
	acc.Add("connections", overview.ObjectTotals.Connections, tags)
	acc.Add("consumers", overview.ObjectTotals.Consumers, tags)
	acc.Add("exchanges", overview.ObjectTotals.Exchanges, tags)
	acc.Add("queues", overview.ObjectTotals.Queues, tags)

	acc.Add("messages_acked", overview.MessageStats.Ack, tags)
	acc.Add("messages_delivered", overview.MessageStats.Deliver, tags)
	acc.Add("messages_published", overview.MessageStats.Publish, tags)

	errChan <- nil
}

func gatherNodes(r *RabbitMQ, serv *Server, acc plugins.Accumulator, errChan chan error) {
	nodes := make([]Node, 0)
	// Gather information about nodes
	err := r.requestJSON(serv, "/api/nodes", &nodes)
	if err != nil {
		errChan <- err
		return
	}

	for _, node := range nodes {
		if !shouldGatherNode(node, serv) {
			continue
		}

		tags := map[string]string{"url": serv.URL}
		tags["node"] = node.Name

		acc.Add("disk_free", node.DiskFree, tags)
		acc.Add("disk_free_limit", node.DiskFreeLimit, tags)
		acc.Add("fd_total", node.FdTotal, tags)
		acc.Add("fd_used", node.FdUsed, tags)
		acc.Add("mem_limit", node.MemLimit, tags)
		acc.Add("mem_used", node.MemUsed, tags)
		acc.Add("proc_total", node.ProcTotal, tags)
		acc.Add("proc_used", node.ProcUsed, tags)
		acc.Add("run_queue", node.RunQueue, tags)
		acc.Add("sockets_total", node.SocketsTotal, tags)
		acc.Add("sockets_used", node.SocketsUsed, tags)
	}

	errChan <- nil
}

func gatherQueues(r *RabbitMQ, serv *Server, acc plugins.Accumulator, errChan chan error) {
	// Gather information about queues
	queues := make([]Queue, 0)
	err := r.requestJSON(serv, "/api/queues", &queues)
	if err != nil {
		errChan <- err
		return
	}

	for _, queue := range queues {
		if !shouldGatherQueue(queue, serv) {
			continue
		}
		tags := map[string]string{
			"url":         serv.URL,
			"queue":       queue.Name,
			"vhost":       queue.Vhost,
			"node":        queue.Node,
			"durable":     strconv.FormatBool(queue.Durable),
			"auto_delete": strconv.FormatBool(queue.AutoDelete),
		}

		acc.AddFields(
			"queue",
			map[string]interface{}{
				// common information
				"consumers":            queue.Consumers,
				"consumer_utilisation": queue.ConsumerUtilisation,
				"memory":               queue.Memory,
				// messages information
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

func shouldGatherNode(node Node, serv *Server) bool {
	if len(serv.Nodes) == 0 {
		return true
	}

	for _, name := range serv.Nodes {
		if name == node.Name {
			return true
		}
	}

	return false
}

func shouldGatherQueue(queue Queue, serv *Server) bool {
	if len(serv.Queues) == 0 {
		return true
	}

	for _, name := range serv.Queues {
		if name == queue.Name {
			return true
		}
	}

	return false
}

func init() {
	plugins.Add("rabbitmq", func() plugins.Plugin {
		return &RabbitMQ{}
	})
}
