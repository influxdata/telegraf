package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DefaultUsername will set a default value that corrasponds to the default
// value used by Rabbitmq
const DefaultUsername = "guest"

// DefaultPassword will set a default value that corrasponds to the default
// value used by Rabbitmq
const DefaultPassword = "guest"

// DefaultURL will set a default value that corrasponds to the default value
// used by Rabbitmq
const DefaultURL = "http://localhost:15672"

// Default http timeouts
const DefaultResponseHeaderTimeout = 3
const DefaultClientTimeout = 4

// RabbitMQ defines the configuration necessary for gathering metrics,
// see the sample config for further details
type RabbitMQ struct {
	URL      string
	Name     string
	Username string
	Password string
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	ResponseHeaderTimeout internal.Duration `toml:"header_timeout"`
	ClientTimeout         internal.Duration `toml:"client_timeout"`

	// InsecureSkipVerify bool
	Nodes  []string
	Queues []string

	Client *http.Client
}

// OverviewResponse ...
type OverviewResponse struct {
	MessageStats *MessageStats `json:"message_stats"`
	ObjectTotals *ObjectTotals `json:"object_totals"`
	QueueTotals  *QueueTotals  `json:"queue_totals"`
}

// Details ...
type Details struct {
	Rate float64
}

// MessageStats ...
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

// ObjectTotals ...
type ObjectTotals struct {
	Channels    int64
	Connections int64
	Consumers   int64
	Exchanges   int64
	Queues      int64
}

// QueueTotals ...
type QueueTotals struct {
	Messages                   int64
	MessagesReady              int64 `json:"messages_ready"`
	MessagesUnacknowledged     int64 `json:"messages_unacknowledged"`
	MessageBytes               int64 `json:"message_bytes"`
	MessageBytesReady          int64 `json:"message_bytes_ready"`
	MessageBytesUnacknowledged int64 `json:"message_bytes_unacknowledged"`
	MessageRAM                 int64 `json:"message_bytes_ram"`
	MessagePersistent          int64 `json:"message_bytes_persistent"`
}

// Queue ...
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
	AutoDelete          bool   `json:"auto_delete"`
	IdleSince           string `json:"idle_since"`
}

// Node ...
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

// gatherFunc ...
type gatherFunc func(r *RabbitMQ, acc telegraf.Accumulator)

var gatherFunctions = []gatherFunc{gatherOverview, gatherNodes, gatherQueues}

var sampleConfig = `
  ## Management Plugin url. (default: http://localhost:15672)
  # url = "http://localhost:15672"
  ## Tag added to rabbitmq_overview series; deprecated: use tags
  # name = "rmq-server-1"
  ## Credentials
  # username = "guest"
  # password = "guest"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional request timeouts
  ##
  ## ResponseHeaderTimeout, if non-zero, specifies the amount of time to wait
  ## for a server's response headers after fully writing the request.
  # header_timeout = "3s"
  ##
  ## client_timeout specifies a time limit for requests made by this client.
  ## Includes connection time, any redirects, and reading the response body.
  # client_timeout = "4s"

  ## A list of nodes to pull metrics about. If not specified, metrics for
  ## all nodes are gathered.
  # nodes = ["rabbit@node1", "rabbit@node2"]
`

// SampleConfig ...
func (r *RabbitMQ) SampleConfig() string {
	return sampleConfig
}

// Description ...
func (r *RabbitMQ) Description() string {
	return "Reads metrics from RabbitMQ servers via the Management Plugin"
}

// Gather ...
func (r *RabbitMQ) Gather(acc telegraf.Accumulator) error {
	if r.Client == nil {
		tlsCfg, err := internal.GetTLSConfig(
			r.SSLCert, r.SSLKey, r.SSLCA, r.InsecureSkipVerify)
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: r.ResponseHeaderTimeout.Duration,
			TLSClientConfig:       tlsCfg,
		}
		r.Client = &http.Client{
			Transport: tr,
			Timeout:   r.ClientTimeout.Duration,
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(gatherFunctions))
	for _, f := range gatherFunctions {
		go func(gf gatherFunc) {
			defer wg.Done()
			gf(r, acc)
		}(f)
	}
	wg.Wait()

	return nil
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

func gatherOverview(r *RabbitMQ, acc telegraf.Accumulator) {
	overview := &OverviewResponse{}

	err := r.requestJSON("/api/overview", &overview)
	if err != nil {
		acc.AddError(err)
		return
	}

	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil {
		acc.AddError(fmt.Errorf("Wrong answer from rabbitmq. Probably auth issue"))
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
}

func gatherNodes(r *RabbitMQ, acc telegraf.Accumulator) {
	nodes := make([]Node, 0)
	// Gather information about nodes
	err := r.requestJSON("/api/nodes", &nodes)
	if err != nil {
		acc.AddError(err)
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
}

func gatherQueues(r *RabbitMQ, acc telegraf.Accumulator) {
	// Gather information about queues
	queues := make([]Queue, 0)
	err := r.requestJSON("/api/queues", &queues)
	if err != nil {
		acc.AddError(err)
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
				"idle_since":           queue.IdleSince,
				"memory":               queue.Memory,
				// messages information
				"message_bytes":             queue.MessageBytes,
				"message_bytes_ready":       queue.MessageBytesReady,
				"message_bytes_unacked":     queue.MessageBytesUnacknowledged,
				"message_bytes_ram":         queue.MessageRAM,
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
		return &RabbitMQ{
			ResponseHeaderTimeout: internal.Duration{Duration: DefaultResponseHeaderTimeout * time.Second},
			ClientTimeout:         internal.Duration{Duration: DefaultClientTimeout * time.Second},
		}
	})
}
