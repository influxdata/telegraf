package burrow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	defaultBurrowPrefix          = "/v3/kafka"
	defaultConcurrentConnections = 20
	defaultResponseTimeout       = time.Second * 5
	defaultServer                = "http://localhost:8000"
)

const configSample = `
  ## Burrow API endpoints in format "schema://host:port".
  ## Default is "http://localhost:8000".
  servers = ["http://localhost:8000"]

  ## Override Burrow API prefix.
  ## Useful when Burrow is behind reverse-proxy.
  # api_prefix = "/v3/kafka"

  ## Maximum time to receive response.
  # response_timeout = "5s"

  ## Limit per-server concurrent connections.
  ## Useful in case of large number of topics or consumer groups.
  # concurrent_connections = 20

  ## Filter clusters, default is no filtering.
  ## Values can be specified as glob patterns.
  # clusters_include = []
  # clusters_exclude = []

  ## Filter consumer groups, default is no filtering.
  ## Values can be specified as glob patterns.
  # groups_include = []
  # groups_exclude = []

  ## Filter topics, default is no filtering.
  ## Values can be specified as glob patterns.
  # topics_include = []
  # topics_exclude = []

  ## Credentials for basic HTTP authentication.
  # username = ""
  # password = ""

  ## Optional SSL config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  # insecure_skip_verify = false
`

type (
	burrow struct {
		tls.ClientConfig

		Servers               []string
		Username              string
		Password              string
		ResponseTimeout       internal.Duration
		ConcurrentConnections int

		APIPrefix       string `toml:"api_prefix"`
		ClustersExclude []string
		ClustersInclude []string
		GroupsExclude   []string
		GroupsInclude   []string
		TopicsExclude   []string
		TopicsInclude   []string

		client         *http.Client
		filterClusters filter.Filter
		filterGroups   filter.Filter
		filterTopics   filter.Filter
	}

	// response
	apiResponse struct {
		Clusters []string          `json:"clusters"`
		Groups   []string          `json:"consumers"`
		Topics   []string          `json:"topics"`
		Offsets  []int64           `json:"offsets"`
		Status   apiStatusResponse `json:"status"`
	}

	// response: status field
	apiStatusResponse struct {
		Partitions     []apiStatusResponseLag `json:"partitions"`
		Status         string                 `json:"status"`
		PartitionCount int                    `json:"partition_count"`
		Maxlag         *apiStatusResponseLag  `json:"maxlag"`
		TotalLag       int64                  `json:"totallag"`
	}

	// response: lag field
	apiStatusResponseLag struct {
		Topic      string                   `json:"topic"`
		Partition  int32                    `json:"partition"`
		Status     string                   `json:"status"`
		Start      apiStatusResponseLagItem `json:"start"`
		End        apiStatusResponseLagItem `json:"end"`
		CurrentLag int64                    `json:"current_lag"`
		Owner      string                   `json:"owner"`
	}

	// response: lag field item
	apiStatusResponseLagItem struct {
		Offset    int64 `json:"offset"`
		Timestamp int64 `json:"timestamp"`
		Lag       int64 `json:"lag"`
	}
)

func init() {
	inputs.Add("burrow", func() telegraf.Input {
		return &burrow{}
	})
}

func (b *burrow) SampleConfig() string {
	return configSample
}

func (b *burrow) Description() string {
	return "Collect Kafka topics and consumers status from Burrow HTTP API."
}

func (b *burrow) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(b.Servers) == 0 {
		b.Servers = []string{defaultServer}
	}

	if b.client == nil {
		b.setDefaults()
		if err := b.compileGlobs(); err != nil {
			return err
		}
		c, err := b.createClient()
		if err != nil {
			return err
		}
		b.client = c
	}

	for _, addr := range b.Servers {
		u, err := url.Parse(addr)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to parse address '%s': %s", addr, err))
			continue
		}
		if u.Path == "" {
			u.Path = b.APIPrefix
		}

		wg.Add(1)
		go func(u *url.URL) {
			defer wg.Done()
			acc.AddError(b.gatherServer(u, acc))
		}(u)
	}

	wg.Wait()
	return nil
}

func (b *burrow) setDefaults() {
	if b.APIPrefix == "" {
		b.APIPrefix = defaultBurrowPrefix
	}
	if b.ConcurrentConnections < 1 {
		b.ConcurrentConnections = defaultConcurrentConnections
	}
	if b.ResponseTimeout.Duration < time.Second {
		b.ResponseTimeout = internal.Duration{
			Duration: defaultResponseTimeout,
		}
	}
}

func (b *burrow) compileGlobs() error {
	var err error

	// compile glob patterns
	b.filterClusters, err = filter.NewIncludeExcludeFilter(b.ClustersInclude, b.ClustersExclude)
	if err != nil {
		return err
	}
	b.filterGroups, err = filter.NewIncludeExcludeFilter(b.GroupsInclude, b.GroupsExclude)
	if err != nil {
		return err
	}
	b.filterTopics, err = filter.NewIncludeExcludeFilter(b.TopicsInclude, b.TopicsExclude)
	if err != nil {
		return err
	}
	return nil
}

func (b *burrow) createClient() (*http.Client, error) {
	tlsCfg, err := b.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: b.ResponseTimeout.Duration,
	}

	return client, nil
}

func (b *burrow) getResponse(u *url.URL) (*apiResponse, error) {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if b.Username != "" {
		req.SetBasicAuth(b.Username, b.Password)
	}

	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wrong response: %d", res.StatusCode)
	}

	ares := &apiResponse{}
	dec := json.NewDecoder(res.Body)

	return ares, dec.Decode(ares)
}

func (b *burrow) gatherServer(src *url.URL, acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	r, err := b.getResponse(src)
	if err != nil {
		return err
	}

	guard := make(chan struct{}, b.ConcurrentConnections)
	for _, cluster := range r.Clusters {
		if !b.filterClusters.Match(cluster) {
			continue
		}

		wg.Add(1)
		go func(cluster string) {
			defer wg.Done()

			// fetch topic list
			// endpoint: <api_prefix>/(cluster)/topic
			ut := appendPathToURL(src, cluster, "topic")
			b.gatherTopics(guard, ut, cluster, acc)
		}(cluster)

		wg.Add(1)
		go func(cluster string) {
			defer wg.Done()

			// fetch consumer group list
			// endpoint: <api_prefix>/(cluster)/consumer
			uc := appendPathToURL(src, cluster, "consumer")
			b.gatherGroups(guard, uc, cluster, acc)
		}(cluster)
	}

	wg.Wait()
	return nil
}

func (b *burrow) gatherTopics(guard chan struct{}, src *url.URL, cluster string, acc telegraf.Accumulator) {
	var wg sync.WaitGroup

	r, err := b.getResponse(src)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, topic := range r.Topics {
		if !b.filterTopics.Match(topic) {
			continue
		}

		guard <- struct{}{}
		wg.Add(1)

		go func(topic string) {
			defer func() {
				<-guard
				wg.Done()
			}()

			// fetch topic offsets
			// endpoint: <api_prefix>/<cluster>/topic/<topic>
			tu := appendPathToURL(src, topic)
			tr, err := b.getResponse(tu)
			if err != nil {
				acc.AddError(err)
				return
			}

			b.genTopicMetrics(tr, cluster, topic, acc)
		}(topic)
	}

	wg.Wait()
}

func (b *burrow) genTopicMetrics(r *apiResponse, cluster, topic string, acc telegraf.Accumulator) {
	for i, offset := range r.Offsets {
		tags := map[string]string{
			"cluster":   cluster,
			"topic":     topic,
			"partition": strconv.Itoa(i),
		}

		acc.AddFields(
			"burrow_topic",
			map[string]interface{}{
				"offset": offset,
			},
			tags,
		)
	}
}

func (b *burrow) gatherGroups(guard chan struct{}, src *url.URL, cluster string, acc telegraf.Accumulator) {
	var wg sync.WaitGroup

	r, err := b.getResponse(src)
	if err != nil {
		acc.AddError(err)
		return
	}

	for _, group := range r.Groups {
		if !b.filterGroups.Match(group) {
			continue
		}

		guard <- struct{}{}
		wg.Add(1)

		go func(group string) {
			defer func() {
				<-guard
				wg.Done()
			}()

			// fetch consumer group status
			// endpoint: <api_prefix>/<cluster>/consumer/<group>/lag
			gl := appendPathToURL(src, group, "lag")
			gr, err := b.getResponse(gl)
			if err != nil {
				acc.AddError(err)
				return
			}

			b.genGroupStatusMetrics(gr, cluster, group, acc)
			b.genGroupLagMetrics(gr, cluster, group, acc)
		}(group)
	}

	wg.Wait()
}

func (b *burrow) genGroupStatusMetrics(r *apiResponse, cluster, group string, acc telegraf.Accumulator) {
	partitionCount := r.Status.PartitionCount
	if partitionCount == 0 {
		partitionCount = len(r.Status.Partitions)
	}

	// get max timestamp and total offset from partitions list
	offset := int64(0)
	timestamp := int64(0)
	for _, partition := range r.Status.Partitions {
		offset += partition.End.Offset
		if partition.End.Timestamp > timestamp {
			timestamp = partition.End.Timestamp
		}
	}

	lag := int64(0)
	if r.Status.Maxlag != nil {
		lag = r.Status.Maxlag.CurrentLag
	}

	acc.AddFields(
		"burrow_group",
		map[string]interface{}{
			"status":          r.Status.Status,
			"status_code":     mapStatusToCode(r.Status.Status),
			"partition_count": partitionCount,
			"total_lag":       r.Status.TotalLag,
			"lag":             lag,
			"offset":          offset,
			"timestamp":       timestamp,
		},
		map[string]string{
			"cluster": cluster,
			"group":   group,
		},
	)
}

func (b *burrow) genGroupLagMetrics(r *apiResponse, cluster, group string, acc telegraf.Accumulator) {
	for _, partition := range r.Status.Partitions {
		acc.AddFields(
			"burrow_partition",
			map[string]interface{}{
				"status":      partition.Status,
				"status_code": mapStatusToCode(partition.Status),
				"lag":         partition.CurrentLag,
				"offset":      partition.End.Offset,
				"timestamp":   partition.End.Timestamp,
			},
			map[string]string{
				"cluster":   cluster,
				"group":     group,
				"topic":     partition.Topic,
				"partition": strconv.FormatInt(int64(partition.Partition), 10),
				"owner":     partition.Owner,
			},
		)
	}
}

func appendPathToURL(src *url.URL, parts ...string) *url.URL {
	dst := new(url.URL)
	*dst = *src

	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}

	ext := strings.Join(parts, "/")
	dst.Path = fmt.Sprintf("%s/%s", src.Path, ext)
	return dst
}

func mapStatusToCode(src string) int {
	switch src {
	case "OK":
		return 1
	case "NOT_FOUND":
		return 2
	case "WARN":
		return 3
	case "ERR":
		return 4
	case "STOP":
		return 5
	case "STALL":
		return 6
	default:
		return 0
	}
}
