package solr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const mbeansPath = "/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERYHANDLER&cat=UPDATEHANDLER&cat=CACHE"
const adminCoresPath = "/solr/admin/cores?action=STATUS&wt=json"

type node struct {
	Host string `json:"host"`
}

const sampleConfig = `
  ## specify a list of one or more Solr servers
  servers = ["http://localhost:8983"]

  ## specify a list of one or more Solr cores (default - all)
  # cores = ["main"]
`

// Solr is a plugin to read stats from one or many Solr servers
type Solr struct {
	Local       bool
	Servers     []string
	HTTPTimeout internal.Duration
	Cores       []string
	client      *http.Client
}

// AdminCores is an exported type that
// contains a response with information about Solr cores.
type AdminCoresStatus struct {
	Status map[string]struct {
		Index struct {
			SizeInBytes int64 `json:"sizeInBytes"`
			NumDocs     int   `json:"numDocs"`
			MaxDoc      int   `json:"maxDoc"`
			DeletedDocs int   `json:"deletedDocs"`
		} `json:"index"`
	} `json:"status"`
}

// Metrics is an exported type that
// contains a response from Solr with metrics
type MBeansData struct {
	Headers    ResponseHeader    `json:"responseHeader"`
	SolrMbeans []json.RawMessage `json:"solr-mbeans"`
}

// ResponseHeader is an exported type that
// contains a response metrics: QTime and Status
type ResponseHeader struct {
	QTime  int `json:"QTime"`
	Status int `json:"status"`
}

// Core is an exported type that
// contains Core metrics
type Core struct {
	Class string `json:"class"`
	Stats struct {
		DeletedDocs int `json:"deletedDocs"`
		MaxDoc      int `json:"maxDoc"`
		NumDocs     int `json:"numDocs"`
	} `json:"stats"`
}

// QueryHandler is an exported type that
// contains query handler metrics
type QueryHandler struct {
	Class string `json:"class"`
	Stats struct {
		One5minRateReqsPerSecond float64 `json:"15minRateReqsPerSecond"`
		FiveMinRateReqsPerSecond float64 `json:"5minRateReqsPerSecond"`
		Seven5thPcRequestTime    float64 `json:"75thPcRequestTime"`
		Nine5thPcRequestTime     float64 `json:"95thPcRequestTime"`
		Nine99thPcRequestTime    float64 `json:"999thPcRequestTime"`
		Nine9thPcRequestTime     float64 `json:"99thPcRequestTime"`
		AvgRequestsPerSecond     float64 `json:"avgRequestsPerSecond"`
		AvgTimePerRequest        float64 `json:"avgTimePerRequest"`
		Errors                   int     `json:"errors"`
		HandlerStart             int     `json:"handlerStart"`
		MedianRequestTime        float64 `json:"medianRequestTime"`
		Requests                 int     `json:"requests"`
		Timeouts                 int     `json:"timeouts"`
		TotalTime                float64 `json:"totalTime"`
	} `json:"stats"`
}

// UpdateHandler is an exported type that
// contains update handler metrics
type UpdateHandler struct {
	Class string `json:"class"`
	Stats struct {
		Adds                     int    `json:"adds"`
		AutocommitMaxDocs        int    `json:"autocommit maxDocs"`
		AutocommitMaxTime        string `json:"autocommit maxTime"`
		Autocommits              int    `json:"autocommits"`
		Commits                  int    `json:"commits"`
		CumulativeAdds           int    `json:"cumulative_adds"`
		CumulativeDeletesByID    int    `json:"cumulative_deletesById"`
		CumulativeDeletesByQuery int    `json:"cumulative_deletesByQuery"`
		CumulativeErrors         int    `json:"cumulative_errors"`
		DeletesByID              int    `json:"deletesById"`
		DeletesByQuery           int    `json:"deletesByQuery"`
		DocsPending              int    `json:"docsPending"`
		Errors                   int    `json:"errors"`
		ExpungeDeletes           int    `json:"expungeDeletes"`
		Optimizes                int    `json:"optimizes"`
		Rollbacks                int    `json:"rollbacks"`
		SoftAutocommits          int    `json:"soft autocommits"`
	} `json:"stats"`
}

// Cache is an exported type that
// contains cache metrics
type Cache struct {
	Class string `json:"class"`
	Stats struct {
		CumulativeEvictions int     `json:"cumulative_evictions"`
		CumulativeHitratio  float64 `json:"cumulative_hitratio,string"`
		CumulativeHits      int     `json:"cumulative_hits"`
		CumulativeInserts   int     `json:"cumulative_inserts"`
		CumulativeLookups   int     `json:"cumulative_lookups"`
		Evictions           int     `json:"evictions"`
		Hitratio            float64 `json:"hitratio,string"`
		Hits                int     `json:"hits"`
		Inserts             int     `json:"inserts"`
		Lookups             int     `json:"lookups"`
		Size                int     `json:"size"`
		WarmupTime          int     `json:"warmupTime"`
	} `json:"stats"`
}

// NewSolr return a new instance of Solr
func NewSolr() *Solr {
	return &Solr{
		HTTPTimeout: internal.Duration{Duration: time.Second * 5},
	}
}

// SampleConfig returns sample configuration for this plugin.
func (s *Solr) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (s *Solr) Description() string {
	return "Read stats from one or more Solr servers or cores"
}

// Gather reads the stats from Solr and writes it to the
// Accumulator.
func (s *Solr) Gather(acc telegraf.Accumulator) error {
	if s.client == nil {
		client := s.createHTTPClient()
		s.client = client
	}

	var wg sync.WaitGroup
	wg.Add(len(s.Servers))

	for _, serv := range s.Servers {
		go func(serv string, acc telegraf.Accumulator) {
			defer wg.Done()
			acc.AddError(s.gatherServerMetrics(serv, acc))
		}(serv, acc)
	}
	wg.Wait()
	return nil
}

// Gather all metrics from server
func (s *Solr) gatherServerMetrics(server string, acc telegraf.Accumulator) error {
	measurementTime := time.Now()
	adminCoresStatus := &AdminCoresStatus{}
	if err := s.gatherData(s.adminUrl(server), adminCoresStatus); err != nil {
		return err
	}
	addAdminCoresStatusToAcc(acc, adminCoresStatus, measurementTime)
	cores := s.filterCores(getCoresFromStatus(adminCoresStatus))
	var wg sync.WaitGroup
	wg.Add(len(cores))
	for _, core := range cores {
		go func(server string, core string, acc telegraf.Accumulator) {
			defer wg.Done()
			mBeansData := &MBeansData{}
			acc.AddError(s.gatherData(s.mbeansUrl(server, core), mBeansData))
			acc.AddError(addCoreMetricsToAcc(acc, core, mBeansData, measurementTime))
			acc.AddError(addQueryHandlerMetricsToAcc(acc, core, mBeansData, measurementTime))
			acc.AddError(addUpdateHandlerMetricsToAcc(acc, core, mBeansData, measurementTime))
			acc.AddError(addCacheMetricsToAcc(acc, core, mBeansData, measurementTime))
		}(server, core, acc)
	}
	wg.Wait()
	return nil
}

// Use cores from configuration if exists, else use cores from server
func (s *Solr) filterCores(serverCores []string) []string {
	if len(s.Cores) == 0 {
		return serverCores
	} else {
		return s.Cores
	}
}

// Return list of cores from solr server
func getCoresFromStatus(adminCoresStatus *AdminCoresStatus) []string {
	serverCores := []string{}
	for coreName := range adminCoresStatus.Status {
		serverCores = append(serverCores, coreName)
	}
	return serverCores
}

// Add core metrics from admin to accumulator
// This is the only point where size_in_bytes is available (as far as I checked)
func addAdminCoresStatusToAcc(acc telegraf.Accumulator, adminCoreStatus *AdminCoresStatus, time time.Time) {
	for core, metrics := range adminCoreStatus.Status {
		coreFields := map[string]interface{}{
			"deleted_docs":  metrics.Index.DeletedDocs,
			"max_docs":      metrics.Index.MaxDoc,
			"num_docs":      metrics.Index.NumDocs,
			"size_in_bytes": metrics.Index.SizeInBytes,
		}
		acc.AddFields(
			"solr_admin",
			coreFields,
			map[string]string{"core": core},
			time,
		)
	}
}

// Add core metrics section to accumulator
func addCoreMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, time time.Time) error {
	var coreMetrics map[string]Core
	if err := json.Unmarshal(mBeansData.SolrMbeans[1], &coreMetrics); err != nil {
		return err
	}
	for name, metrics := range coreMetrics {
		if strings.Contains(name, "@") {
			continue
		}
		coreFields := map[string]interface{}{
			"class_name":   metrics.Class,
			"deleted_docs": metrics.Stats.DeletedDocs,
			"max_docs":     metrics.Stats.MaxDoc,
			"num_docs":     metrics.Stats.NumDocs,
		}
		acc.AddFields(
			"solr_core",
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			time,
		)
	}
	return nil
}

// Add query metrics section to accumulator
func addQueryHandlerMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, time time.Time) error {
	var queryMetrics map[string]QueryHandler

	if err := json.Unmarshal(mBeansData.SolrMbeans[3], &queryMetrics); err != nil {
		return err
	}
	for name, metrics := range queryMetrics {
		coreFields := map[string]interface{}{
			"class_name":                 metrics.Class,
			"15min_rate_reqs_per_second": metrics.Stats.One5minRateReqsPerSecond,
			"5min_rate_reqs_per_second":  metrics.Stats.FiveMinRateReqsPerSecond,
			"75th_pc_request_time":       metrics.Stats.Seven5thPcRequestTime,
			"95th_pc_request_time":       metrics.Stats.Nine5thPcRequestTime,
			"999th_pc_request_time":      metrics.Stats.Nine99thPcRequestTime,
			"99th_pc_request_time":       metrics.Stats.Nine9thPcRequestTime,
			"avg_requests_per_second":    metrics.Stats.AvgRequestsPerSecond,
			"avg_time_per_request":       metrics.Stats.AvgTimePerRequest,
			"errors":                     metrics.Stats.Errors,
			"handler_start":              metrics.Stats.HandlerStart,
			"median_request_time":        metrics.Stats.MedianRequestTime,
			"requests":                   metrics.Stats.Requests,
			"timeouts":                   metrics.Stats.Timeouts,
			"total_time":                 metrics.Stats.TotalTime,
		}
		acc.AddFields(
			"solr_queryhandler",
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			time,
		)
	}
	return nil
}

// Add update metrics section to accumulator
func addUpdateHandlerMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, time time.Time) error {
	var updateMetrics map[string]UpdateHandler

	if err := json.Unmarshal(mBeansData.SolrMbeans[5], &updateMetrics); err != nil {
		return err
	}
	for name, metrics := range updateMetrics {
		var autoCommitMaxTime int
		if len(metrics.Stats.AutocommitMaxTime) > 2 {
			autoCommitMaxTime, _ = strconv.Atoi(metrics.Stats.AutocommitMaxTime[:len(metrics.Stats.AutocommitMaxTime)-2])
		}
		coreFields := map[string]interface{}{
			"class_name":                  metrics.Class,
			"adds":                        metrics.Stats.Adds,
			"autocommit_max_docs":         metrics.Stats.AutocommitMaxDocs,
			"autocommit_max_time":         autoCommitMaxTime,
			"autocommits":                 metrics.Stats.Autocommits,
			"commits":                     metrics.Stats.Commits,
			"cumulative_adds":             metrics.Stats.CumulativeAdds,
			"cumulative_deletes_by_id":    metrics.Stats.CumulativeDeletesByID,
			"cumulative_deletes_by_query": metrics.Stats.CumulativeDeletesByQuery,
			"cumulative_errors":           metrics.Stats.CumulativeErrors,
			"deletes_by_id":               metrics.Stats.DeletesByID,
			"deletes_by_query":            metrics.Stats.DeletesByQuery,
			"docs_pending":                metrics.Stats.DocsPending,
			"errors":                      metrics.Stats.Errors,
			"expunge_deletes":             metrics.Stats.ExpungeDeletes,
			"optimizes":                   metrics.Stats.Optimizes,
			"rollbacks":                   metrics.Stats.Rollbacks,
			"soft_autocommits":            metrics.Stats.SoftAutocommits,
		}
		acc.AddFields(
			"solr_updatehandler",
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			time,
		)
	}
	return nil
}

// Add cache metrics section to accumulator
func addCacheMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, time time.Time) error {
	var cacheMetrics map[string]Cache

	if err := json.Unmarshal(mBeansData.SolrMbeans[7], &cacheMetrics); err != nil {
		return err
	}
	for name, metrics := range cacheMetrics {
		coreFields := map[string]interface{}{
			"class_name":           metrics.Class,
			"cumulative_evictions": metrics.Stats.CumulativeEvictions,
			"cumulative_hitratio":  metrics.Stats.CumulativeHitratio,
			"cumulative_hits":      metrics.Stats.CumulativeHits,
			"cumulative_inserts":   metrics.Stats.CumulativeInserts,
			"cumulative_lookups":   metrics.Stats.CumulativeLookups,
			"evictions":            metrics.Stats.Evictions,
			"hitratio":             metrics.Stats.Hitratio,
			"hits":                 metrics.Stats.Hits,
			"inserts":              metrics.Stats.Inserts,
			"lookups":              metrics.Stats.Lookups,
			"size":                 metrics.Stats.Size,
			"warmup_time":          metrics.Stats.WarmupTime,
		}
		acc.AddFields(
			"solr_cache",
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			time,
		)
	}
	return nil
}

// Provide admin url
func (s *Solr) adminUrl(server string) string {
	return fmt.Sprintf("%s%s", server, adminCoresPath)
}

// Provide mbeans url
func (s *Solr) mbeansUrl(server string, core string) string {
	return fmt.Sprintf("%s/solr/%s%s", server, core, mbeansPath)
}

func (s *Solr) createHTTPClient() *http.Client {
	tr := &http.Transport{
		ResponseHeaderTimeout: s.HTTPTimeout.Duration,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   s.HTTPTimeout.Duration,
	}

	return client
}

func (s *Solr) gatherData(url string, v interface{}) error {
	r, err := s.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("solr: API responded with status-code %d, expected %d, url %s",
			r.StatusCode, http.StatusOK, url)
	}
	if err = json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

func init() {
	inputs.Add("solr", func() telegraf.Input {
		return NewSolr()
	})
}
