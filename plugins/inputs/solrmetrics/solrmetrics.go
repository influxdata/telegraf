package solrmetrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	metricPath     = "/admin/metrics"
	compact        = "true"
	delimeter      = "."
	wt             = "json"
	dftHTTPTimeout = 5

	sampleConfig = `
		## specify a list of one or more Solr servers
		servers = ["http://localhost:8983"]
	
		## Optional HTTP Basic Auth Credentials
		# username = "username"
		# password = "pa$$word"
		#
		## Optional HTTP timeout, sec
		# httptimeout = 5
		#
		# https://lucene.apache.org/solr/guide/metrics-reporting.html#metrics-api
		## Prefixes 
		prefixes = [
			"REPLICATION./replication.replicationEnabled",
			"REPLICATION./replication.isSlave",
			"REPLICATION./replication.isMaster",
			"CACHE.searcher.queryResultCache",
			"INDEX.sizeInBytes",
			"SEARCHER.searcher.numDocs",
			"SEARCHER.searcher.deletedDocs",
			"SEARCHER.searcher.maxDoc"
			]
		#
		## Keys
		keys = [
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.1xx-responses:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.2xx-responses:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.3xx-responses:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.4xx-responses:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.5xx-responses:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.connect-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.options-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.head-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.move-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.delete-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.get-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.post-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.put-requests:count",
			"solr.jetty:org.eclipse.jetty.server.handler.DefaultHandler.other-requests:count",
			"solr.jvm:buffers.direct.Count",
			"solr.jvm:buffers.direct.MemoryUsed",
			"solr.jvm:buffers.direct.TotalCapacity",
			"solr.jvm:buffers.mapped.Count",
			"solr.jvm:buffers.mapped.MemoryUsed",
			"solr.jvm:buffers.mapped.TotalCapacity",
			"solr.jvm:threads.blocked.count",
			"solr.jvm:threads.count",
			"solr.jvm:threads.daemon.count",
			"solr.jvm:threads.deadlock.count",
			"solr.jvm:threads.new.count",
			"solr.jvm:threads.runnable.count",
			"solr.jvm:threads.terminated.count",
			"solr.jvm:threads.timed_waiting.count",
			"solr.jvm:threads.waiting.count",
			"solr.jvm:os.maxFileDescriptorCount",
			"solr.jvm:os.openFileDescriptorCount",
			"solr.jvm:memory.total.init",
			"solr.jvm:memory.total.max",
			"solr.jvm:memory.total.used",
			"solr.jvm:memory.heap.init",
			"solr.jvm:memory.heap.max",
			"solr.jvm:memory.heap.used",
			"solr.jvm:gc.ConcurrentMarkSweep.count",
			"solr.jvm:gc.ConcurrentMarkSweep.time",
			"solr.jvm:gc.ParNew.count",
			"solr.jvm:gc.ParNew.time",
			"solr.node:CONTAINER.fs.totalSpace",
			"solr.node:CONTAINER.fs.usableSpace",
			"solr.node:ADMIN./admin/zookeeper.errors:count",
			"solr.node:ADMIN./admin/zookeeper.timeouts:count",
			"solr.node:CONTAINER.cores.lazy",
			"solr.node:CONTAINER.cores.loaded",
			"solr.node:CONTAINER.cores.unloaded"
		]
	`
)

// ResponseHeader is an exported type that
// contains a response metrics: QTime and Status
type ResponseHeader struct {
	QTime  int64 `json:"QTime"`
	Status int64 `json:"status"`
}

// Serie is type to represent series
type Serie struct {
	mesName   string // Measurament name
	fields    Fields
	tags      Tags
	timestamp time.Time
}

// Tags for series
type Tags map[string]string

// Metrics are type to represent generic metrics returned by Solr
type Metrics map[string]interface{}

// Fields to export
type Fields map[string]interface{}

// MetricResponse is an exported type that
// contains Metrics of different types and represents response of Solr
type MetricResponse struct {
	ResponseHeader ResponseHeader `json:"responseHeader"`
	Metrics        Metrics        `json:"metrics"`
}

// Solr is a plugin to read stats from one or many Solr servers
type Solr struct {
	Local       bool
	Servers     []string
	Username    string
	Password    string
	Prefixes    []string
	Keys        []string
	Httptimeout time.Duration
	Cores       []string
	client      *http.Client
	reqs        []*http.Request
}

func copyStrMaps(src, dst map[string]string) {
	for k := range src {
		dst[k] = src[k]
	}
}

// parseMetric parses metric and returns fields with name "preffix.field" and values
func parseMetrics(metrics Metrics, acc telegraf.Accumulator) Metrics {
	f := make(Metrics)
	for k, v := range metrics {
		t := reflect.ValueOf(v)
		switch t.Kind() {
		case reflect.Int:
			f[k] = t.Int()
		case reflect.Float32, reflect.Float64:
			f[k] = t.Float()
		case reflect.String:
			f[k] = t.String()
		case reflect.Bool:
			f[k] = t.Bool()
		case reflect.Map:
			subMetsMap, ok := v.(map[string]interface{})
			if !ok {
				acc.AddError(fmt.Errorf("error of converting map"))
				continue
			}
			subMets := parseMetrics(subMetsMap, acc)
			if len(subMets) == 0 {
				continue
			}
			for km, vm := range subMets {
				key := k + delimeter + km
				f[key] = vm
			}
		default:
			acc.AddError(fmt.Errorf("error of parsing, unknow type. %s=%s", k, v))
			continue
		}
	}
	return f
}

// getMetricTags parces metric name to collection, shard, replica, metric
func getCollectionTags(metricName string) ([]string, error) {
	ct := strings.SplitN(metricName, ".", 4)
	if len(ct) < 4 {
		return nil, fmt.Errorf("wrong metric name %s", metricName)
	}
	return ct, nil
}

func metricsToSeries(metrics Metrics, addTags Tags, acc telegraf.Accumulator) (res []Serie) {
	mTime := time.Now()
	series := make(map[string]Serie)

	// Create series for standard groups
	for _, group := range []string{"jvm", "jetty", "node"} {
		series[group] = Serie{mesName: group, fields: make(Fields), tags: make(Tags), timestamp: mTime}
		copyStrMaps(addTags, series[group].tags)
	}

	// Parce metrics by a group
	for k, v := range metrics {
		// standard groups
		for _, group := range []string{"jvm", "jetty", "node"} {
			if strings.Contains(k, "solr"+delimeter+group) {
				newSerie := strings.Replace(k, "solr."+group+":", "", 1)
				series[group].fields[newSerie] = v
			}
		}
		// Core group
		// Each core of any collection should be present as a separate serie with uniq tags
		if strings.Contains(k, "solr.core") {
			newSerie := strings.Replace(k, "solr.core"+delimeter, "", 1)
			// Decompose a name of a metric
			colTags, err := getCollectionTags(newSerie)
			if err != nil || len(colTags) != 4 {
				acc.AddError(err)
				continue
			}
			collection, shard, replica, metricName := colTags[0], colTags[1], colTags[2], colTags[3]
			newSerie = collection + shard + replica
			// Check and create an unique serie if it doesn't exist
			if _, ok := series[newSerie]; !ok {
				series[newSerie] = Serie{mesName: "core", fields: make(Fields), tags: make(Tags), timestamp: mTime}
				// Add requeremnted tags
				copyStrMaps(addTags, series[newSerie].tags)
				series[newSerie].tags["collection"] = collection
				series[newSerie].tags["shard"] = shard
				series[newSerie].tags["replica"] = replica
			}
			series[newSerie].fields[metricName] = v
		}
	}
	res = []Serie{}
	for _, v := range series {
		// Add a serie, if fileds contain metrics
		if len(v.fields) > 0 {
			res = append(res, v)
		}
	}
	return res
}

// SampleConfig returns sample configuration for this plugin.
func (s *Solr) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (s *Solr) Description() string {
	return "Read stats from Solr servers and cores using Metrics API"
}

// NewSolr return a new instance of Solr
func NewSolr() *Solr {
	return &Solr{}
}

// Provide metrics urls
func (s *Solr) metricsURLs(acc telegraf.Accumulator) ([]url.URL, error) {
	if len(s.Prefixes) == 0 && len(s.Keys) == 0 {
		return nil, fmt.Errorf("'keys' and 'prefix' are empty")
	}
	URLs := make([]url.URL, 0, len(s.Servers)*2)
	for _, server := range s.Servers {
		URLPrs := make(map[string]url.Values, 0)
		if len(s.Keys) > 0 {
			URLPrs["keys"] = url.Values{}
		}
		for _, key := range s.Keys {
			URLPrs["keys"].Add("key", key)
		}
		if len(s.Prefixes) > 0 {
			URLPrs["prefixes"] = url.Values{}
			URLPrs["prefixes"].Set("prefix", strings.Join(s.Prefixes, ","))
		}
		// URL for a server
		strURL := server + "/solr" + metricPath
		serverURL, err := url.Parse(strURL)
		if err != nil {
			acc.AddError(err)
			continue
		}
		// Add common parameters and set parameters for the server URL
		for _, v := range URLPrs {
			v.Set("compact", compact)
			v.Set("wt", wt)
			serverURL.RawQuery = v.Encode()
			URLs = append(URLs, *serverURL)
		}

	}
	return URLs, nil
}

func (s *Solr) createHTTPClient() *http.Client {
	// Set default value if the http timeout wasn't set in config
	if s.Httptimeout == 0 {
		s.Httptimeout = dftHTTPTimeout
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: s.Httptimeout * time.Second,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   s.Httptimeout * time.Second,
	}
	return client
}

func (s *Solr) gatherData(req *http.Request, m interface{}) error {
	r, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("solr: API responded with status-code %d, expected %d, url %s",
			r.StatusCode, http.StatusOK, req.URL)
	}
	if err = json.NewDecoder(r.Body).Decode(m); err != nil {
		return err
	}
	return nil
}

// CreateReqs creates request for the each Solr node
func (s *Solr) CreateReqs(acc telegraf.Accumulator) ([]*http.Request, error) {

	urls, err := s.metricsURLs(acc)
	if urls == nil {
		return nil, fmt.Errorf("empty list of URLs to request the metrics. %s", err)
	}

	var reqs []*http.Request
	for _, url := range urls {
		req, reqErr := http.NewRequest(http.MethodGet, url.String(), nil)
		if reqErr != nil {
			acc.AddError(err)
			continue
		}
		if s.Username != "" {
			req.SetBasicAuth(s.Username, s.Password)
		}
		req.Header.Set("User-Agent", "Telegraf/"+internal.Version())
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func (s *Solr) gatherServerMetrics(req *http.Request, acc telegraf.Accumulator) error {
	mr := &MetricResponse{}
	err := s.gatherData(req, mr)
	if err != nil {
		return err
	}
	metrics := parseMetrics(mr.Metrics, acc)
	if len(metrics) == 0 {
		return fmt.Errorf("no Solr metrics handled")
	}
	// Default tags
	tags := make(Tags)
	tags["port"] = req.URL.Port()
	series := metricsToSeries(metrics, tags, acc)
	if len(series) == 0 {
		return fmt.Errorf("no metrics")
	}
	for _, serie := range series {
		acc.AddFields(serie.mesName, serie.fields, serie.tags, serie.timestamp)
	}

	return nil
}

// Gather reads the stats from Solr and writes it to the
// Accumulator.
func (s *Solr) Gather(acc telegraf.Accumulator) (err error) {
	if s.client == nil {
		client := s.createHTTPClient()
		s.client = client
	}
	if s.reqs == nil {
		if s.reqs, err = s.CreateReqs(acc); err != nil {
			acc.AddError(err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(s.reqs))

	for _, req := range s.reqs {
		go func(req *http.Request, acc telegraf.Accumulator) {
			defer wg.Done()
			acc.AddError(s.gatherServerMetrics(req, acc))
		}(req, acc)
	}
	wg.Wait()
	return err
}

func init() {
	inputs.Add("solrmetrics", func() telegraf.Input {
		return NewSolr()
	})
}
