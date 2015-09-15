package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"

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

type MessageStats struct {
	Ack     int64
	Deliver int64
	Publish int64
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

var sampleConfig = `
	# Specify servers via an array of tables
	[[rabbitmq.servers]]
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

	if len(r.Servers) == 0 {
		r.gatherServer(localhost, acc)
		return nil
	}

	for _, serv := range r.Servers {
		err := r.gatherServer(serv, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) gatherServer(serv *Server, acc plugins.Accumulator) error {
	overview := &OverviewResponse{}

	err := r.requestJSON(serv, "/api/overview", &overview)
	if err != nil {
		return err
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

	if overview.MessageStats != nil {
		acc.Add("messages_acked", overview.MessageStats.Ack, tags)
		acc.Add("messages_delivered", overview.MessageStats.Deliver, tags)
		acc.Add("messages_published", overview.MessageStats.Publish, tags)
	}

	nodes := make([]Node, 0)

	err = r.requestJSON(serv, "/api/nodes", &nodes)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if !shouldGatherNode(node, serv) {
			continue
		}

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

	return nil
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

func init() {
	plugins.Add("rabbitmq", func() plugins.Plugin {
		return &RabbitMQ{}
	})
}
