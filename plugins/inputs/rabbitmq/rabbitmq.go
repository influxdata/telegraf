package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
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
	URL      string `toml:"url"`
	Name     string `toml:"name"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	tls.ClientConfig

	ResponseHeaderTimeout internal.Duration `toml:"header_timeout"`
	ClientTimeout         internal.Duration `toml:"client_timeout"`

	Nodes     []string `toml:"nodes"`
	Queues    []string `toml:"queues"`
	Exchanges []string `toml:"exchanges"`

	QueueInclude              []string `toml:"queue_name_include"`
	QueueExclude              []string `toml:"queue_name_exclude"`
	FederationUpstreamInclude []string `toml:"federation_upstream_include"`
	FederationUpstreamExclude []string `toml:"federation_upstream_exclude"`

	Client *http.Client `toml:"-"`

	filterCreated     bool
	excludeEveryQueue bool
	queueFilter       filter.Filter
	upstreamFilter    filter.Filter
}

// OverviewResponse ...
type OverviewResponse struct {
	MessageStats *MessageStats `json:"message_stats"`
	ObjectTotals *ObjectTotals `json:"object_totals"`
	QueueTotals  *QueueTotals  `json:"queue_totals"`
	Listeners    []Listeners   `json:"listeners"`
}

// Listeners ...
type Listeners struct {
	Protocol string `json:"protocol"`
}

// Details ...
type Details struct {
	Rate float64 `json:"rate"`
}

// MessageStats ...
type MessageStats struct {
	Ack                     int64
	AckDetails              Details `json:"ack_details"`
	Deliver                 int64
	DeliverDetails          Details `json:"deliver_details"`
	DeliverGet              int64   `json:"deliver_get"`
	DeliverGetDetails       Details `json:"deliver_get_details"`
	Publish                 int64
	PublishDetails          Details `json:"publish_details"`
	Redeliver               int64
	RedeliverDetails        Details `json:"redeliver_details"`
	PublishIn               int64   `json:"publish_in"`
	PublishInDetails        Details `json:"publish_in_details"`
	PublishOut              int64   `json:"publish_out"`
	PublishOutDetails       Details `json:"publish_out_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails Details `json:"return_unroutable_details"`
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

	DiskFree                 int64   `json:"disk_free"`
	DiskFreeLimit            int64   `json:"disk_free_limit"`
	DiskFreeAlarm            bool    `json:"disk_free_alarm"`
	FdTotal                  int64   `json:"fd_total"`
	FdUsed                   int64   `json:"fd_used"`
	MemLimit                 int64   `json:"mem_limit"`
	MemUsed                  int64   `json:"mem_used"`
	MemAlarm                 bool    `json:"mem_alarm"`
	ProcTotal                int64   `json:"proc_total"`
	ProcUsed                 int64   `json:"proc_used"`
	RunQueue                 int64   `json:"run_queue"`
	SocketsTotal             int64   `json:"sockets_total"`
	SocketsUsed              int64   `json:"sockets_used"`
	Running                  bool    `json:"running"`
	Uptime                   int64   `json:"uptime"`
	MnesiaDiskTxCount        int64   `json:"mnesia_disk_tx_count"`
	MnesiaDiskTxCountDetails Details `json:"mnesia_disk_tx_count_details"`
	MnesiaRamTxCount         int64   `json:"mnesia_ram_tx_count"`
	MnesiaRamTxCountDetails  Details `json:"mnesia_ram_tx_count_details"`
	GcNum                    int64   `json:"gc_num"`
	GcNumDetails             Details `json:"gc_num_details"`
	GcBytesReclaimed         int64   `json:"gc_bytes_reclaimed"`
	GcBytesReclaimedDetails  Details `json:"gc_bytes_reclaimed_details"`
	IoReadAvgTime            int64   `json:"io_read_avg_time"`
	IoReadAvgTimeDetails     Details `json:"io_read_avg_time_details"`
	IoReadBytes              int64   `json:"io_read_bytes"`
	IoReadBytesDetails       Details `json:"io_read_bytes_details"`
	IoWriteAvgTime           int64   `json:"io_write_avg_time"`
	IoWriteAvgTimeDetails    Details `json:"io_write_avg_time_details"`
	IoWriteBytes             int64   `json:"io_write_bytes"`
	IoWriteBytesDetails      Details `json:"io_write_bytes_details"`
}

type Exchange struct {
	Name         string
	MessageStats `json:"message_stats"`
	Type         string
	Internal     bool
	Vhost        string
	Durable      bool
	AutoDelete   bool `json:"auto_delete"`
}

// FederationLinkChannelMessageStats ...
type FederationLinkChannelMessageStats struct {
	Confirm                 int64   `json:"confirm"`
	ConfirmDetails          Details `json:"confirm_details"`
	Publish                 int64   `json:"publish"`
	PublishDetails          Details `json:"publish_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails Details `json:"return_unroutable_details"`
}

// FederationLinkChannel ...
type FederationLinkChannel struct {
	AcksUncommitted        int64                             `json:"acks_uncommitted"`
	ConsumerCount          int64                             `json:"consumer_count"`
	MessagesUnacknowledged int64                             `json:"messages_unacknowledged"`
	MessagesUncommitted    int64                             `json:"messages_uncommitted"`
	MessagesUnconfirmed    int64                             `json:"messages_unconfirmed"`
	MessageStats           FederationLinkChannelMessageStats `json:"message_stats"`
}

// FederationLink ...
type FederationLink struct {
	Type             string                `json:"type"`
	Queue            string                `json:"queue"`
	UpstreamQueue    string                `json:"upstream_queue"`
	Exchange         string                `json:"exchange"`
	UpstreamExchange string                `json:"upstream_exchange"`
	Vhost            string                `json:"vhost"`
	Upstream         string                `json:"upstream"`
	LocalChannel     FederationLinkChannel `json:"local_channel"`
}

type HealthCheck struct {
	Status string `json:"status"`
}

// MemoryResponse ...
type MemoryResponse struct {
	Memory *Memory `json:"memory"`
}

// Memory details
type Memory struct {
	ConnectionReaders   int64 `json:"connection_readers"`
	ConnectionWriters   int64 `json:"connection_writers"`
	ConnectionChannels  int64 `json:"connection_channels"`
	ConnectionOther     int64 `json:"connection_other"`
	QueueProcs          int64 `json:"queue_procs"`
	QueueSlaveProcs     int64 `json:"queue_slave_procs"`
	Plugins             int64 `json:"plugins"`
	OtherProc           int64 `json:"other_proc"`
	Metrics             int64 `json:"metrics"`
	MgmtDb              int64 `json:"mgmt_db"`
	Mnesia              int64 `json:"mnesia"`
	OtherEts            int64 `json:"other_ets"`
	Binary              int64 `json:"binary"`
	MsgIndex            int64 `json:"msg_index"`
	Code                int64 `json:"code"`
	Atom                int64 `json:"atom"`
	OtherSystem         int64 `json:"other_system"`
	AllocatedUnused     int64 `json:"allocated_unused"`
	ReservedUnallocated int64 `json:"reserved_unallocated"`
	Total               int64 `json:"total"`
}

// gatherFunc ...
type gatherFunc func(r *RabbitMQ, acc telegraf.Accumulator)

var gatherFunctions = []gatherFunc{gatherOverview, gatherNodes, gatherQueues, gatherExchanges, gatherFederationLinks}

var sampleConfig = `
  ## Management Plugin url. (default: http://localhost:15672)
  # url = "http://localhost:15672"
  ## Tag added to rabbitmq_overview series; deprecated: use tags
  # name = "rmq-server-1"
  ## Credentials
  # username = "guest"
  # password = "guest"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
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

  ## A list of nodes to gather as the rabbitmq_node measurement. If not
  ## specified, metrics for all nodes are gathered.
  # nodes = ["rabbit@node1", "rabbit@node2"]

  ## A list of queues to gather as the rabbitmq_queue measurement. If not
  ## specified, metrics for all queues are gathered.
  # queues = ["telegraf"]

  ## A list of exchanges to gather as the rabbitmq_exchange measurement. If not
  ## specified, metrics for all exchanges are gathered.
  # exchanges = ["telegraf"]

  ## Queues to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all queues
  queue_name_include = []
  queue_name_exclude = []

  ## Federation upstreams include and exclude when gathering the rabbitmq_federation measurement.
  ## If neither are specified, metrics for all federation upstreams are gathered.
  ## Federation link metrics will only be gathered for queues and exchanges
  ## whose non-federation metrics will be collected (e.g a queue excluded
  ## by the 'queue_name_exclude' option will also be excluded from federation).
  ## Globs accepted.
  # federation_upstream_include = ["dataCentre-*"]
  # federation_upstream_exclude = []
`

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

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
		tlsCfg, err := r.ClientConfig.TLSConfig()
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

	// Create gather filters if not already created
	if !r.filterCreated {
		err := r.createQueueFilter()
		if err != nil {
			return err
		}
		err = r.createUpstreamFilter()
		if err != nil {
			return err
		}
		r.filterCreated = true
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

	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil || overview.Listeners == nil {
		acc.AddError(fmt.Errorf("Wrong answer from rabbitmq. Probably auth issue"))
		return
	}

	var clusteringListeners, amqpListeners int64 = 0, 0
	for _, listener := range overview.Listeners {
		if listener.Protocol == "clustering" {
			clusteringListeners++
		} else if listener.Protocol == "amqp" {
			amqpListeners++
		}
	}

	tags := map[string]string{"url": r.URL}
	if r.Name != "" {
		tags["name"] = r.Name
	}
	fields := map[string]interface{}{
		"messages":               overview.QueueTotals.Messages,
		"messages_ready":         overview.QueueTotals.MessagesReady,
		"messages_unacked":       overview.QueueTotals.MessagesUnacknowledged,
		"channels":               overview.ObjectTotals.Channels,
		"connections":            overview.ObjectTotals.Connections,
		"consumers":              overview.ObjectTotals.Consumers,
		"exchanges":              overview.ObjectTotals.Exchanges,
		"queues":                 overview.ObjectTotals.Queues,
		"messages_acked":         overview.MessageStats.Ack,
		"messages_delivered":     overview.MessageStats.Deliver,
		"messages_delivered_get": overview.MessageStats.DeliverGet,
		"messages_published":     overview.MessageStats.Publish,
		"clustering_listeners":   clusteringListeners,
		"amqp_listeners":         amqpListeners,
		"return_unroutable":      overview.MessageStats.ReturnUnroutable,
		"return_unroutable_rate": overview.MessageStats.ReturnUnroutableDetails.Rate,
	}
	acc.AddFields("rabbitmq_overview", fields, tags)
}

func gatherNodes(r *RabbitMQ, acc telegraf.Accumulator) {
	allNodes := make([]Node, 0)
	// Gather information about nodes
	err := r.requestJSON("/api/nodes", &allNodes)
	if err != nil {
		acc.AddError(err)
		return
	}

	nodes := make(map[string]Node)
	for _, node := range allNodes {
		if r.shouldGatherNode(node) {
			nodes[node.Name] = node
		}
	}

	numberNodes := len(nodes)
	if numberNodes == 0 {
		return
	}

	type NodeCheck struct {
		NodeName    string
		HealthCheck HealthCheck
		Memory      *Memory
	}

	nodeChecksChannel := make(chan NodeCheck, numberNodes)

	for _, node := range nodes {
		go func(nodeName string, healthChecksChannel chan NodeCheck) {
			var healthCheck HealthCheck
			var memoryresponse MemoryResponse

			err := r.requestJSON("/api/healthchecks/node/"+nodeName, &healthCheck)
			nodeCheck := NodeCheck{
				NodeName:    nodeName,
				HealthCheck: healthCheck,
			}
			if err != nil {
				acc.AddError(err)
				return
			}

			err = r.requestJSON("/api/nodes/"+nodeName+"/memory", &memoryresponse)
			nodeCheck.Memory = memoryresponse.Memory
			if err != nil {
				acc.AddError(err)
				return
			}

			nodeChecksChannel <- nodeCheck
		}(node.Name, nodeChecksChannel)
	}

	now := time.Now()

	for i := 0; i < len(nodes); i++ {
		nodeCheck := <-nodeChecksChannel

		var healthCheckStatus int64 = 0

		if nodeCheck.HealthCheck.Status == "ok" {
			healthCheckStatus = 1
		}

		node := nodes[nodeCheck.NodeName]

		tags := map[string]string{"url": r.URL}
		tags["node"] = node.Name

		fields := map[string]interface{}{
			"disk_free":                 node.DiskFree,
			"disk_free_limit":           node.DiskFreeLimit,
			"disk_free_alarm":           boolToInt(node.DiskFreeAlarm),
			"fd_total":                  node.FdTotal,
			"fd_used":                   node.FdUsed,
			"mem_limit":                 node.MemLimit,
			"mem_used":                  node.MemUsed,
			"mem_alarm":                 boolToInt(node.MemAlarm),
			"proc_total":                node.ProcTotal,
			"proc_used":                 node.ProcUsed,
			"run_queue":                 node.RunQueue,
			"sockets_total":             node.SocketsTotal,
			"sockets_used":              node.SocketsUsed,
			"uptime":                    node.Uptime,
			"mnesia_disk_tx_count":      node.MnesiaDiskTxCount,
			"mnesia_disk_tx_count_rate": node.MnesiaDiskTxCountDetails.Rate,
			"mnesia_ram_tx_count":       node.MnesiaRamTxCount,
			"mnesia_ram_tx_count_rate":  node.MnesiaRamTxCountDetails.Rate,
			"gc_num":                    node.GcNum,
			"gc_num_rate":               node.GcNumDetails.Rate,
			"gc_bytes_reclaimed":        node.GcBytesReclaimed,
			"gc_bytes_reclaimed_rate":   node.GcBytesReclaimedDetails.Rate,
			"io_read_avg_time":          node.IoReadAvgTime,
			"io_read_avg_time_rate":     node.IoReadAvgTimeDetails.Rate,
			"io_read_bytes":             node.IoReadBytes,
			"io_read_bytes_rate":        node.IoReadBytesDetails.Rate,
			"io_write_avg_time":         node.IoWriteAvgTime,
			"io_write_avg_time_rate":    node.IoWriteAvgTimeDetails.Rate,
			"io_write_bytes":            node.IoWriteBytes,
			"io_write_bytes_rate":       node.IoWriteBytesDetails.Rate,
			"running":                   boolToInt(node.Running),
			"health_check_status":       healthCheckStatus,
		}
		if nodeCheck.Memory != nil {
			fields["mem_connection_readers"] = nodeCheck.Memory.ConnectionReaders
			fields["mem_connection_writers"] = nodeCheck.Memory.ConnectionWriters
			fields["mem_connection_channels"] = nodeCheck.Memory.ConnectionChannels
			fields["mem_connection_other"] = nodeCheck.Memory.ConnectionOther
			fields["mem_queue_procs"] = nodeCheck.Memory.QueueProcs
			fields["mem_queue_slave_procs"] = nodeCheck.Memory.QueueSlaveProcs
			fields["mem_plugins"] = nodeCheck.Memory.Plugins
			fields["mem_other_proc"] = nodeCheck.Memory.OtherProc
			fields["mem_metrics"] = nodeCheck.Memory.Metrics
			fields["mem_mgmt_db"] = nodeCheck.Memory.MgmtDb
			fields["mem_mnesia"] = nodeCheck.Memory.Mnesia
			fields["mem_other_ets"] = nodeCheck.Memory.OtherEts
			fields["mem_binary"] = nodeCheck.Memory.Binary
			fields["mem_msg_index"] = nodeCheck.Memory.MsgIndex
			fields["mem_code"] = nodeCheck.Memory.Code
			fields["mem_atom"] = nodeCheck.Memory.Atom
			fields["mem_other_system"] = nodeCheck.Memory.OtherSystem
			fields["mem_allocated_unused"] = nodeCheck.Memory.AllocatedUnused
			fields["mem_reserved_unallocated"] = nodeCheck.Memory.ReservedUnallocated
			fields["mem_total"] = nodeCheck.Memory.Total
		}
		acc.AddFields("rabbitmq_node", fields, tags, now)
	}
}

func gatherQueues(r *RabbitMQ, acc telegraf.Accumulator) {
	if r.excludeEveryQueue {
		return
	}
	// Gather information about queues
	queues := make([]Queue, 0)
	err := r.requestJSON("/api/queues", &queues)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, queue := range queues {
		if !r.queueFilter.Match(queue.Name) {
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

func gatherExchanges(r *RabbitMQ, acc telegraf.Accumulator) {
	// Gather information about exchanges
	exchanges := make([]Exchange, 0)
	err := r.requestJSON("/api/exchanges", &exchanges)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, exchange := range exchanges {
		if !r.shouldGatherExchange(exchange.Name) {
			continue
		}
		tags := map[string]string{
			"url":         r.URL,
			"exchange":    exchange.Name,
			"type":        exchange.Type,
			"vhost":       exchange.Vhost,
			"internal":    strconv.FormatBool(exchange.Internal),
			"durable":     strconv.FormatBool(exchange.Durable),
			"auto_delete": strconv.FormatBool(exchange.AutoDelete),
		}

		acc.AddFields(
			"rabbitmq_exchange",
			map[string]interface{}{
				"messages_publish_in":       exchange.MessageStats.PublishIn,
				"messages_publish_in_rate":  exchange.MessageStats.PublishInDetails.Rate,
				"messages_publish_out":      exchange.MessageStats.PublishOut,
				"messages_publish_out_rate": exchange.MessageStats.PublishOutDetails.Rate,
			},
			tags,
		)
	}
}

func gatherFederationLinks(r *RabbitMQ, acc telegraf.Accumulator) {
	// Gather information about federation links
	federationLinks := make([]FederationLink, 0)
	err := r.requestJSON("/api/federation-links", &federationLinks)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, link := range federationLinks {
		if !r.shouldGatherFederationLink(link) {
			continue
		}

		tags := map[string]string{
			"url":      r.URL,
			"type":     link.Type,
			"vhost":    link.Vhost,
			"upstream": link.Upstream,
		}

		if link.Type == "exchange" {
			tags["exchange"] = link.Exchange
			tags["upstream_exchange"] = link.UpstreamExchange
		} else {
			tags["queue"] = link.Queue
			tags["upstream_queue"] = link.UpstreamQueue
		}

		acc.AddFields(
			"rabbitmq_federation",
			map[string]interface{}{
				"acks_uncommitted":           link.LocalChannel.AcksUncommitted,
				"consumers":                  link.LocalChannel.ConsumerCount,
				"messages_unacknowledged":    link.LocalChannel.MessagesUnacknowledged,
				"messages_uncommitted":       link.LocalChannel.MessagesUncommitted,
				"messages_unconfirmed":       link.LocalChannel.MessagesUnconfirmed,
				"messages_confirm":           link.LocalChannel.MessageStats.Confirm,
				"messages_publish":           link.LocalChannel.MessageStats.Publish,
				"messages_return_unroutable": link.LocalChannel.MessageStats.ReturnUnroutable,
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

func (r *RabbitMQ) createQueueFilter() error {
	// Backwards compatibility for deprecated `queues` parameter.
	if len(r.Queues) > 0 {
		r.QueueInclude = append(r.QueueInclude, r.Queues...)
	}

	queueFilter, err := filter.NewIncludeExcludeFilter(r.QueueInclude, r.QueueExclude)
	if err != nil {
		return err
	}
	r.queueFilter = queueFilter

	for _, q := range r.QueueExclude {
		if q == "*" {
			r.excludeEveryQueue = true
		}
	}

	return nil
}

func (r *RabbitMQ) createUpstreamFilter() error {
	upstreamFilter, err := filter.NewIncludeExcludeFilter(r.FederationUpstreamInclude, r.FederationUpstreamExclude)
	if err != nil {
		return err
	}
	r.upstreamFilter = upstreamFilter

	return nil
}

func (r *RabbitMQ) shouldGatherExchange(exchangeName string) bool {
	if len(r.Exchanges) == 0 {
		return true
	}

	for _, name := range r.Exchanges {
		if name == exchangeName {
			return true
		}
	}

	return false
}

func (r *RabbitMQ) shouldGatherFederationLink(link FederationLink) bool {
	if !r.upstreamFilter.Match(link.Upstream) {
		return false
	}

	switch link.Type {
	case "exchange":
		return r.shouldGatherExchange(link.Exchange)
	case "queue":
		return r.queueFilter.Match(link.Queue)
	default:
		return false
	}
}

func init() {
	inputs.Add("rabbitmq", func() telegraf.Input {
		return &RabbitMQ{
			ResponseHeaderTimeout: internal.Duration{Duration: DefaultResponseHeaderTimeout * time.Second},
			ClientTimeout:         internal.Duration{Duration: DefaultClientTimeout * time.Second},
		}
	})
}
