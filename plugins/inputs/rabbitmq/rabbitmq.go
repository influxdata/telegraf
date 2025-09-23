//go:generate ../../../tools/readme_config_includer/generator
package rabbitmq

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultUsername              = "guest"
	defaultPassword              = "guest"
	defaultURL                   = "http://localhost:15672"
	defaultResponseHeaderTimeout = 3
	defaultClientTimeout         = 4
)

// RabbitMQ defines the configuration necessary for gathering metrics,
// see the sample config for further details
type RabbitMQ struct {
	URL      string        `toml:"url"`
	Username config.Secret `toml:"username"`
	Password config.Secret `toml:"password"`
	tls.ClientConfig

	ResponseHeaderTimeout config.Duration `toml:"header_timeout"`
	ClientTimeout         config.Duration `toml:"client_timeout"`

	Nodes     []string `toml:"nodes"`
	Exchanges []string `toml:"exchanges"`

	MetricInclude             []string `toml:"metric_include"`
	MetricExclude             []string `toml:"metric_exclude"`
	QueueInclude              []string `toml:"queue_name_include"`
	QueueExclude              []string `toml:"queue_name_exclude"`
	FederationUpstreamInclude []string `toml:"federation_upstream_include"`
	FederationUpstreamExclude []string `toml:"federation_upstream_exclude"`

	IncludeQueueTypeTag bool `toml:"include_queue_type_tag"`

	Log telegraf.Logger `toml:"-"`

	client            *http.Client
	excludeEveryQueue bool
	metricFilter      filter.Filter
	queueFilter       filter.Filter
	upstreamFilter    filter.Filter
}

type overviewResponse struct {
	MessageStats *messageStats `json:"message_stats"`
	ObjectTotals *objectTotals `json:"object_totals"`
	QueueTotals  *queueTotals  `json:"queue_totals"`
	Listeners    []listeners   `json:"listeners"`
}

type listeners struct {
	Protocol string `json:"protocol"`
}

type details struct {
	Rate float64 `json:"rate"`
}

type messageStats struct {
	Ack                     int64
	AckDetails              details `json:"ack_details"`
	Deliver                 int64
	DeliverDetails          details `json:"deliver_details"`
	DeliverGet              int64   `json:"deliver_get"`
	DeliverGetDetails       details `json:"deliver_get_details"`
	Publish                 int64
	PublishDetails          details `json:"publish_details"`
	Redeliver               int64
	RedeliverDetails        details `json:"redeliver_details"`
	PublishIn               int64   `json:"publish_in"`
	PublishInDetails        details `json:"publish_in_details"`
	PublishOut              int64   `json:"publish_out"`
	PublishOutDetails       details `json:"publish_out_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails details `json:"return_unroutable_details"`
}

type objectTotals struct {
	Channels    int64
	Connections int64
	Consumers   int64
	Exchanges   int64
	Queues      int64
}

type queueTotals struct {
	Messages                   int64
	MessagesReady              int64 `json:"messages_ready"`
	MessagesUnacknowledged     int64 `json:"messages_unacknowledged"`
	MessageBytes               int64 `json:"message_bytes"`
	MessageBytesReady          int64 `json:"message_bytes_ready"`
	MessageBytesUnacknowledged int64 `json:"message_bytes_unacknowledged"`
	MessageRAM                 int64 `json:"message_bytes_ram"`
	MessagePersistent          int64 `json:"message_bytes_persistent"`
}

type queue struct {
	queueTotals            // just to not repeat the same code
	messageStats           `json:"message_stats"`
	Memory                 int64
	Consumers              int64
	ConsumerUtilisation    float64 `json:"consumer_utilisation"`
	Name                   string
	Node                   string
	Vhost                  string
	Type                   string
	Durable                bool
	AutoDelete             bool     `json:"auto_delete"`
	IdleSince              string   `json:"idle_since"`
	SlaveNodes             []string `json:"slave_nodes"`
	SynchronisedSlaveNodes []string `json:"synchronised_slave_nodes"`
	HeadMessageTimestamp   *int64   `json:"head_message_timestamp"`
}

type node struct {
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
	MnesiaDiskTxCountDetails details `json:"mnesia_disk_tx_count_details"`
	MnesiaRAMTxCount         int64   `json:"mnesia_ram_tx_count"`
	MnesiaRAMTxCountDetails  details `json:"mnesia_ram_tx_count_details"`
	GcNum                    int64   `json:"gc_num"`
	GcNumDetails             details `json:"gc_num_details"`
	GcBytesReclaimed         int64   `json:"gc_bytes_reclaimed"`
	GcBytesReclaimedDetails  details `json:"gc_bytes_reclaimed_details"`
	IoReadAvgTime            float64 `json:"io_read_avg_time"`
	IoReadAvgTimeDetails     details `json:"io_read_avg_time_details"`
	IoReadBytes              int64   `json:"io_read_bytes"`
	IoReadBytesDetails       details `json:"io_read_bytes_details"`
	IoWriteAvgTime           float64 `json:"io_write_avg_time"`
	IoWriteAvgTimeDetails    details `json:"io_write_avg_time_details"`
	IoWriteBytes             int64   `json:"io_write_bytes"`
	IoWriteBytesDetails      details `json:"io_write_bytes_details"`
}

type exchange struct {
	Name         string
	messageStats `json:"message_stats"`
	Type         string
	Internal     bool
	Vhost        string
	Durable      bool
	AutoDelete   bool `json:"auto_delete"`
}

type federationLinkChannelMessageStats struct {
	Confirm                 int64   `json:"confirm"`
	ConfirmDetails          details `json:"confirm_details"`
	Publish                 int64   `json:"publish"`
	PublishDetails          details `json:"publish_details"`
	ReturnUnroutable        int64   `json:"return_unroutable"`
	ReturnUnroutableDetails details `json:"return_unroutable_details"`
}

type federationLinkChannel struct {
	AcksUncommitted        int64                             `json:"acks_uncommitted"`
	ConsumerCount          int64                             `json:"consumer_count"`
	MessagesUnacknowledged int64                             `json:"messages_unacknowledged"`
	MessagesUncommitted    int64                             `json:"messages_uncommitted"`
	MessagesUnconfirmed    int64                             `json:"messages_unconfirmed"`
	MessageStats           federationLinkChannelMessageStats `json:"message_stats"`
}

type federationLink struct {
	Type             string                `json:"type"`
	Queue            string                `json:"queue"`
	UpstreamQueue    string                `json:"upstream_queue"`
	Exchange         string                `json:"exchange"`
	UpstreamExchange string                `json:"upstream_exchange"`
	Vhost            string                `json:"vhost"`
	Upstream         string                `json:"upstream"`
	LocalChannel     federationLinkChannel `json:"local_channel"`
}

type memoryResponse struct {
	Memory *memory `json:"memory"`
}

// memory details
type memory struct {
	ConnectionReaders   int64       `json:"connection_readers"`
	ConnectionWriters   int64       `json:"connection_writers"`
	ConnectionChannels  int64       `json:"connection_channels"`
	ConnectionOther     int64       `json:"connection_other"`
	QueueProcs          int64       `json:"queue_procs"`
	QueueSlaveProcs     int64       `json:"queue_slave_procs"`
	Plugins             int64       `json:"plugins"`
	OtherProc           int64       `json:"other_proc"`
	Metrics             int64       `json:"metrics"`
	MgmtDB              int64       `json:"mgmt_db"`
	Mnesia              int64       `json:"mnesia"`
	OtherEts            int64       `json:"other_ets"`
	Binary              int64       `json:"binary"`
	MsgIndex            int64       `json:"msg_index"`
	Code                int64       `json:"code"`
	Atom                int64       `json:"atom"`
	OtherSystem         int64       `json:"other_system"`
	AllocatedUnused     int64       `json:"allocated_unused"`
	ReservedUnallocated int64       `json:"reserved_unallocated"`
	Total               interface{} `json:"total"`
}

type errorResponse struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

type gatherFunc func(r *RabbitMQ, acc telegraf.Accumulator)

var gatherFunctions = map[string]gatherFunc{
	"exchange":   gatherExchanges,
	"federation": gatherFederationLinks,
	"node":       gatherNodes,
	"overview":   gatherOverview,
	"queue":      gatherQueues,
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func (*RabbitMQ) SampleConfig() string {
	return sampleConfig
}

func (r *RabbitMQ) Init() error {
	var err error

	// Create gather filters
	if err := r.createQueueFilter(); err != nil {
		return err
	}
	if err := r.createUpstreamFilter(); err != nil {
		return err
	}

	// Create a filter for the metrics
	if r.metricFilter, err = filter.NewIncludeExcludeFilter(r.MetricInclude, r.MetricExclude); err != nil {
		return err
	}

	tlsCfg, err := r.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(r.ResponseHeaderTimeout),
		TLSClientConfig:       tlsCfg,
	}
	r.client = &http.Client{
		Transport: tr,
		Timeout:   time.Duration(r.ClientTimeout),
	}

	return nil
}

func (r *RabbitMQ) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for name, f := range gatherFunctions {
		// Query only metrics that are supported
		if !r.metricFilter.Match(name) {
			continue
		}
		wg.Add(1)
		go func(gf gatherFunc) {
			defer wg.Done()
			gf(r, acc)
		}(f)
	}
	wg.Wait()

	return nil
}

func (r *RabbitMQ) requestEndpoint(u string) ([]byte, error) {
	if r.URL == "" {
		r.URL = defaultURL
	}
	endpoint := r.URL + u
	r.Log.Debugf("Requesting %q...", endpoint)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	username := defaultUsername
	if !r.Username.Empty() {
		usernameSecret, err := r.Username.Get()
		if err != nil {
			return nil, err
		}
		defer usernameSecret.Destroy()
		username = usernameSecret.String()
	}

	password := defaultPassword
	if !r.Password.Empty() {
		passwordSecret, err := r.Password.Get()
		if err != nil {
			return nil, err
		}
		defer passwordSecret.Destroy()
		password = passwordSecret.String()
	}

	req.SetBasicAuth(username, password)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r.Log.Debugf("HTTP status code: %v %v", resp.StatusCode, http.StatusText(resp.StatusCode))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("getting %q failed: %v %v", u, resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return io.ReadAll(resp.Body)
}

func (r *RabbitMQ) requestJSON(u string, target interface{}) error {
	buf, err := r.requestEndpoint(u)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, target); err != nil {
		var jsonErr *json.UnmarshalTypeError
		if errors.As(err, &jsonErr) {
			// Try to get the error reason from the response
			var errResponse errorResponse
			if json.Unmarshal(buf, &errResponse) == nil && errResponse.Error != "" {
				// Return the error reason in the response
				return fmt.Errorf("error response trying to get %q: %q (reason: %q)", u, errResponse.Error, errResponse.Reason)
			}
		}

		return fmt.Errorf("decoding answer from %q failed: %w", u, err)
	}

	return nil
}

func gatherOverview(r *RabbitMQ, acc telegraf.Accumulator) {
	overview := &overviewResponse{}

	err := r.requestJSON("/api/overview", &overview)
	if err != nil {
		acc.AddError(err)
		return
	}

	if overview.QueueTotals == nil || overview.ObjectTotals == nil || overview.MessageStats == nil {
		acc.AddError(errors.New("wrong answer from rabbitmq, probably auth issue"))
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
	allNodes := make([]*node, 0)

	err := r.requestJSON("/api/nodes", &allNodes)
	if err != nil {
		acc.AddError(err)
		return
	}

	nodes := allNodes[:0]
	for _, singleNode := range allNodes {
		if r.shouldGatherNode(singleNode) {
			nodes = append(nodes, singleNode)
		}
	}

	var wg sync.WaitGroup
	for _, singleNode := range nodes {
		wg.Add(1)
		go func(singleNode *node) {
			defer wg.Done()

			tags := map[string]string{"url": r.URL}
			tags["node"] = singleNode.Name

			fields := map[string]interface{}{
				"disk_free":                 singleNode.DiskFree,
				"disk_free_limit":           singleNode.DiskFreeLimit,
				"disk_free_alarm":           boolToInt(singleNode.DiskFreeAlarm),
				"fd_total":                  singleNode.FdTotal,
				"fd_used":                   singleNode.FdUsed,
				"mem_limit":                 singleNode.MemLimit,
				"mem_used":                  singleNode.MemUsed,
				"mem_alarm":                 boolToInt(singleNode.MemAlarm),
				"proc_total":                singleNode.ProcTotal,
				"proc_used":                 singleNode.ProcUsed,
				"run_queue":                 singleNode.RunQueue,
				"sockets_total":             singleNode.SocketsTotal,
				"sockets_used":              singleNode.SocketsUsed,
				"uptime":                    singleNode.Uptime,
				"mnesia_disk_tx_count":      singleNode.MnesiaDiskTxCount,
				"mnesia_disk_tx_count_rate": singleNode.MnesiaDiskTxCountDetails.Rate,
				"mnesia_ram_tx_count":       singleNode.MnesiaRAMTxCount,
				"mnesia_ram_tx_count_rate":  singleNode.MnesiaRAMTxCountDetails.Rate,
				"gc_num":                    singleNode.GcNum,
				"gc_num_rate":               singleNode.GcNumDetails.Rate,
				"gc_bytes_reclaimed":        singleNode.GcBytesReclaimed,
				"gc_bytes_reclaimed_rate":   singleNode.GcBytesReclaimedDetails.Rate,
				"io_read_avg_time":          singleNode.IoReadAvgTime,
				"io_read_avg_time_rate":     singleNode.IoReadAvgTimeDetails.Rate,
				"io_read_bytes":             singleNode.IoReadBytes,
				"io_read_bytes_rate":        singleNode.IoReadBytesDetails.Rate,
				"io_write_avg_time":         singleNode.IoWriteAvgTime,
				"io_write_avg_time_rate":    singleNode.IoWriteAvgTimeDetails.Rate,
				"io_write_bytes":            singleNode.IoWriteBytes,
				"io_write_bytes_rate":       singleNode.IoWriteBytesDetails.Rate,
				"running":                   boolToInt(singleNode.Running),
			}

			var memory memoryResponse
			err = r.requestJSON("/api/nodes/"+singleNode.Name+"/memory", &memory)
			if err != nil {
				acc.AddError(err)
				return
			}

			if memory.Memory != nil {
				fields["mem_connection_readers"] = memory.Memory.ConnectionReaders
				fields["mem_connection_writers"] = memory.Memory.ConnectionWriters
				fields["mem_connection_channels"] = memory.Memory.ConnectionChannels
				fields["mem_connection_other"] = memory.Memory.ConnectionOther
				fields["mem_queue_procs"] = memory.Memory.QueueProcs
				fields["mem_queue_slave_procs"] = memory.Memory.QueueSlaveProcs
				fields["mem_plugins"] = memory.Memory.Plugins
				fields["mem_other_proc"] = memory.Memory.OtherProc
				fields["mem_metrics"] = memory.Memory.Metrics
				fields["mem_mgmt_db"] = memory.Memory.MgmtDB
				fields["mem_mnesia"] = memory.Memory.Mnesia
				fields["mem_other_ets"] = memory.Memory.OtherEts
				fields["mem_binary"] = memory.Memory.Binary
				fields["mem_msg_index"] = memory.Memory.MsgIndex
				fields["mem_code"] = memory.Memory.Code
				fields["mem_atom"] = memory.Memory.Atom
				fields["mem_other_system"] = memory.Memory.OtherSystem
				fields["mem_allocated_unused"] = memory.Memory.AllocatedUnused
				fields["mem_reserved_unallocated"] = memory.Memory.ReservedUnallocated
				switch v := memory.Memory.Total.(type) {
				case float64:
					fields["mem_total"] = int64(v)
				case map[string]interface{}:
					var foundEstimator bool
					for _, estimator := range []string{"rss", "allocated", "erlang"} {
						if x, found := v[estimator]; found {
							if total, ok := x.(float64); ok {
								fields["mem_total"] = int64(total)
								foundEstimator = true
								break
							}
							acc.AddError(fmt.Errorf("unknown type %T for %q total memory", x, estimator))
						}
					}
					if !foundEstimator {
						acc.AddError(fmt.Errorf("no known memory estimation in %v", v))
					}
				default:
					acc.AddError(fmt.Errorf("unknown type %T for total memory", memory.Memory.Total))
				}
			}

			acc.AddFields("rabbitmq_node", fields, tags)
		}(singleNode)
	}

	wg.Wait()
}

func gatherQueues(r *RabbitMQ, acc telegraf.Accumulator) {
	if r.excludeEveryQueue {
		return
	}
	// Gather information about queues
	queues := make([]queue, 0)
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

		if r.IncludeQueueTypeTag {
			if queue.Type == "" {
				tags["type"] = "classic"
			} else {
				tags["type"] = queue.Type
			}
		}

		fields := map[string]interface{}{
			// common information
			"consumers":                queue.Consumers,
			"consumer_utilisation":     queue.ConsumerUtilisation,
			"idle_since":               queue.IdleSince,
			"slave_nodes":              len(queue.SlaveNodes),
			"synchronised_slave_nodes": len(queue.SynchronisedSlaveNodes),
			"memory":                   queue.Memory,
			// messages information
			"message_bytes":             queue.MessageBytes,
			"message_bytes_ready":       queue.MessageBytesReady,
			"message_bytes_unacked":     queue.MessageBytesUnacknowledged,
			"message_bytes_ram":         queue.MessageRAM,
			"message_bytes_persist":     queue.MessagePersistent,
			"messages":                  queue.Messages,
			"messages_ready":            queue.MessagesReady,
			"messages_unack":            queue.MessagesUnacknowledged,
			"messages_ack":              queue.messageStats.Ack,
			"messages_ack_rate":         queue.messageStats.AckDetails.Rate,
			"messages_deliver":          queue.messageStats.Deliver,
			"messages_deliver_rate":     queue.messageStats.DeliverDetails.Rate,
			"messages_deliver_get":      queue.messageStats.DeliverGet,
			"messages_deliver_get_rate": queue.messageStats.DeliverGetDetails.Rate,
			"messages_publish":          queue.messageStats.Publish,
			"messages_publish_rate":     queue.messageStats.PublishDetails.Rate,
			"messages_redeliver":        queue.messageStats.Redeliver,
			"messages_redeliver_rate":   queue.messageStats.RedeliverDetails.Rate,
		}

		if queue.HeadMessageTimestamp != nil {
			fields["head_message_timestamp"] = *queue.HeadMessageTimestamp
		}

		acc.AddFields(
			"rabbitmq_queue",
			fields,
			tags,
		)
	}
}

func gatherExchanges(r *RabbitMQ, acc telegraf.Accumulator) {
	// Gather information about exchanges
	exchanges := make([]exchange, 0)
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
				"messages_publish_in":       exchange.messageStats.PublishIn,
				"messages_publish_in_rate":  exchange.messageStats.PublishInDetails.Rate,
				"messages_publish_out":      exchange.messageStats.PublishOut,
				"messages_publish_out_rate": exchange.messageStats.PublishOutDetails.Rate,
			},
			tags,
		)
	}
}

func gatherFederationLinks(r *RabbitMQ, acc telegraf.Accumulator) {
	// Gather information about federation links
	federationLinks := make([]federationLink, 0)
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

func (r *RabbitMQ) shouldGatherNode(node *node) bool {
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

func (r *RabbitMQ) shouldGatherFederationLink(link federationLink) bool {
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
			ResponseHeaderTimeout: config.Duration(defaultResponseHeaderTimeout * time.Second),
			ClientTimeout:         config.Duration(defaultClientTimeout * time.Second),
		}
	})
}
