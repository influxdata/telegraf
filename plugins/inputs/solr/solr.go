package solr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/errchan"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const mbeansPath = "/admin/mbeans?stats=true&wt=json"
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
type AdminCores struct {
	Status map[string]json.RawMessage `json:"status"`
}

// Metrics is an exported type that
// contains a response from Solr with metrics
type Metrics struct {
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

// Default settings
func (s *Solr) setDefaults() ([][]string, int, error) {
	var max int
	cores := [][]string{}
	if len(s.Cores) == 0 {
		for n, server := range s.Servers {
			adminCores := &AdminCores{}
			if err := s.gatherData(fmt.Sprintf("%s%s", server, adminCoresPath), adminCores); err != nil {
				return nil, 0, err
			}
			serverCores := []string{}
			for coreName := range adminCores.Status {
				serverCores = append(serverCores, coreName)
			}
			cores = append(cores, serverCores)
			if len(cores[n]) > max {
				max = len(cores[n])
			}
		}
	} else {
		cores = append(cores, s.Cores)
		max = len(s.Cores)
	}
	return cores, max, nil
}

// Gather reads the stats from Solr and writes it to the
// Accumulator.
func (s *Solr) Gather(acc telegraf.Accumulator) error {
	if s.client == nil {
		client := s.createHTTPClient()
		s.client = client
	}
	cores, max, err := s.setDefaults()
	if err != nil {
		return err
	}

	errChan := errchan.New(len(s.Servers) * max)
	var wg sync.WaitGroup

	for n, serv := range s.Servers {
		for _, core := range cores[n] {
			wg.Add(1)
			go func(serv string, core string, acc telegraf.Accumulator) {
				defer wg.Done()
				if err := s.gatherCoreStats(fmt.Sprintf("%s/solr/%s"+mbeansPath, serv, core), core, acc); err != nil {
					errChan.C <- err
					return
				}
			}(serv, core, acc)
		}
	}
	wg.Wait()
	return errChan.Error()
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

func gatherCoreMetrics(mbeansJSON json.RawMessage, core, category string, acc telegraf.Accumulator) error {
	var coreMetrics map[string]Core

	measurementTime := time.Now()
	if err := json.Unmarshal(mbeansJSON, &coreMetrics); err != nil {
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
			fmt.Sprintf("solr_%s", strings.ToLower(category)),
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			measurementTime,
		)
	}
	return nil
}

func gatherQueryHandlerMetrics(mbeansJSON json.RawMessage, core, category string, acc telegraf.Accumulator) error {
	var coreMetrics map[string]QueryHandler

	measurementTime := time.Now()
	if err := json.Unmarshal(mbeansJSON, &coreMetrics); err != nil {
		return err
	}
	for name, metrics := range coreMetrics {
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
			fmt.Sprintf("solr_%s", strings.ToLower(category)),
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			measurementTime,
		)
	}
	return nil
}

func gatherUpdateHandlerMetrics(mbeansJSON json.RawMessage, core, category string, acc telegraf.Accumulator) error {
	var coreMetrics map[string]UpdateHandler

	measurementTime := time.Now()
	if err := json.Unmarshal(mbeansJSON, &coreMetrics); err != nil {
		return err
	}
	for name, metrics := range coreMetrics {
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
			fmt.Sprintf("solr_%s", strings.ToLower(category)),
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			measurementTime,
		)
	}
	return nil
}

func gatherCacheMetrics(mbeansJSON json.RawMessage, core, category string, acc telegraf.Accumulator) error {
	var coreMetrics map[string]Cache

	measurementTime := time.Now()
	if err := json.Unmarshal(mbeansJSON, &coreMetrics); err != nil {
		return err
	}
	for name, metrics := range coreMetrics {
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
			fmt.Sprintf("solr_%s", strings.ToLower(category)),
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			measurementTime,
		)
	}
	return nil
}

func (s *Solr) gatherCoreStats(url string, core string, acc telegraf.Accumulator) error {

	for _, category := range []string{"CORE", "QUERYHANDLER", "UPDATEHANDLER", "CACHE"} {
		metrics := &Metrics{}
		if err := s.gatherData(fmt.Sprintf("%s&cat=%s", url, category), metrics); err != nil {
			acc.AddError(fmt.Errorf("JSON fetch error: %s", err))
		}
		switch category {
		case "CORE":
			if err := gatherCoreMetrics(metrics.SolrMbeans[1], core, category, acc); err != nil {
				acc.AddError(fmt.Errorf("core category: %s", err))
			}
		case "QUERYHANDLER":
			if err := gatherQueryHandlerMetrics(metrics.SolrMbeans[1], core, category, acc); err != nil {
				acc.AddError(fmt.Errorf("query handler category: %s", err))
			}
		case "UPDATEHANDLER":
			if err := gatherUpdateHandlerMetrics(metrics.SolrMbeans[1], core, category, acc); err != nil {
				acc.AddError(fmt.Errorf("update handler category: %s", err))
			}
		case "CACHE":
			if err := gatherCacheMetrics(metrics.SolrMbeans[1], core, category, acc); err != nil {
				acc.AddError(fmt.Errorf("cache category: %s", err))
			}
		default:
			err := errors.New("unrecognized category")
			acc.AddError(fmt.Errorf("category: %s", err))
		}
	}
	return nil
}

func (s *Solr) gatherData(url string, v interface{}) error {
	r, err := s.client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("solr: API responded with status-code %d, expected %d",
			r.StatusCode, http.StatusOK)
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
