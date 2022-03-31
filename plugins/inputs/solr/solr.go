package solr

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const mbeansPath = "/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERYHANDLER&cat=UPDATEHANDLER&cat=CACHE"
const adminCoresPath = "/solr/admin/cores?action=STATUS&wt=json"

// Solr is a plugin to read stats from one or many Solr servers
type Solr struct {
	Local       bool
	Servers     []string
	Username    string
	Password    string
	HTTPTimeout config.Duration
	Cores       []string
	client      *http.Client
}

// AdminCoresStatus is an exported type that
// contains a response with information about Solr cores.
type AdminCoresStatus struct {
	Status map[string]struct {
		Index struct {
			SizeInBytes int64 `json:"sizeInBytes"`
			NumDocs     int64 `json:"numDocs"`
			MaxDoc      int64 `json:"maxDoc"`
			DeletedDocs int64 `json:"deletedDocs"`
		} `json:"index"`
	} `json:"status"`
}

// MBeansData is an exported type that
// contains a response from Solr with metrics
type MBeansData struct {
	Headers    ResponseHeader    `json:"responseHeader"`
	SolrMbeans []json.RawMessage `json:"solr-mbeans"`
}

// ResponseHeader is an exported type that
// contains a response metrics: QTime and Status
type ResponseHeader struct {
	QTime  int64 `json:"QTime"`
	Status int64 `json:"status"`
}

// Core is an exported type that
// contains Core metrics
type Core struct {
	Stats struct {
		DeletedDocs int64 `json:"deletedDocs"`
		MaxDoc      int64 `json:"maxDoc"`
		NumDocs     int64 `json:"numDocs"`
	} `json:"stats"`
}

// QueryHandler is an exported type that
// contains query handler metrics
type QueryHandler struct {
	Stats interface{} `json:"stats"`
}

// UpdateHandler is an exported type that
// contains update handler metrics
type UpdateHandler struct {
	Stats struct {
		Adds                     int64  `json:"adds"`
		AutocommitMaxDocs        int64  `json:"autocommit maxDocs"`
		AutocommitMaxTime        string `json:"autocommit maxTime"`
		Autocommits              int64  `json:"autocommits"`
		Commits                  int64  `json:"commits"`
		CumulativeAdds           int64  `json:"cumulative_adds"`
		CumulativeDeletesByID    int64  `json:"cumulative_deletesById"`
		CumulativeDeletesByQuery int64  `json:"cumulative_deletesByQuery"`
		CumulativeErrors         int64  `json:"cumulative_errors"`
		DeletesByID              int64  `json:"deletesById"`
		DeletesByQuery           int64  `json:"deletesByQuery"`
		DocsPending              int64  `json:"docsPending"`
		Errors                   int64  `json:"errors"`
		ExpungeDeletes           int64  `json:"expungeDeletes"`
		Optimizes                int64  `json:"optimizes"`
		Rollbacks                int64  `json:"rollbacks"`
		SoftAutocommits          int64  `json:"soft autocommits"`
	} `json:"stats"`
}

// Hitratio is an helper interface
// so we can later on convert it to float64
type Hitratio interface{}

// Cache is an exported type that
// contains cache metrics
type Cache struct {
	Stats map[string]interface{} `json:"stats"`
}

// NewSolr return a new instance of Solr
func NewSolr() *Solr {
	return &Solr{
		HTTPTimeout: config.Duration(time.Second * 5),
	}
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
	if err := s.gatherData(s.adminURL(server), adminCoresStatus); err != nil {
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
			acc.AddError(s.gatherData(s.mbeansURL(server, core), mBeansData))
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
	}
	return s.Cores
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
func addAdminCoresStatusToAcc(acc telegraf.Accumulator, adminCoreStatus *AdminCoresStatus, measurementTime time.Time) {
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
			measurementTime,
		)
	}
}

// Add core metrics section to accumulator
func addCoreMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, measurementTime time.Time) error {
	var coreMetrics map[string]Core
	if len(mBeansData.SolrMbeans) < 2 {
		return fmt.Errorf("no core metric data to unmarshal")
	}
	if err := json.Unmarshal(mBeansData.SolrMbeans[1], &coreMetrics); err != nil {
		return err
	}
	for name, metrics := range coreMetrics {
		if strings.Contains(name, "@") {
			continue
		}
		coreFields := map[string]interface{}{
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
			measurementTime,
		)
	}
	return nil
}

// Add query metrics section to accumulator
func addQueryHandlerMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, measurementTime time.Time) error {
	var queryMetrics map[string]QueryHandler

	if len(mBeansData.SolrMbeans) < 4 {
		return fmt.Errorf("no query handler metric data to unmarshal")
	}

	if err := json.Unmarshal(mBeansData.SolrMbeans[3], &queryMetrics); err != nil {
		return err
	}

	for name, metrics := range queryMetrics {
		var coreFields map[string]interface{}

		if metrics.Stats == nil {
			continue
		}

		switch v := metrics.Stats.(type) {
		case []interface{}:
			m := convertArrayToMap(v)
			coreFields = convertQueryHandlerMap(m)
		case map[string]interface{}:
			coreFields = convertQueryHandlerMap(v)
		default:
			continue
		}

		acc.AddFields(
			"solr_queryhandler",
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			measurementTime,
		)
	}
	return nil
}

func convertArrayToMap(values []interface{}) map[string]interface{} {
	var key string
	result := make(map[string]interface{})
	for i, item := range values {
		if i%2 == 0 {
			key = fmt.Sprintf("%v", item)
		} else {
			result[key] = item
		}
	}

	return result
}

func convertQueryHandlerMap(value map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"15min_rate_reqs_per_second": getFloat(value["15minRateReqsPerSecond"]),
		"5min_rate_reqs_per_second":  getFloat(value["5minRateReqsPerSecond"]),
		"75th_pc_request_time":       getFloat(value["75thPcRequestTime"]),
		"95th_pc_request_time":       getFloat(value["95thPcRequestTime"]),
		"99th_pc_request_time":       getFloat(value["99thPcRequestTime"]),
		"999th_pc_request_time":      getFloat(value["999thPcRequestTime"]),
		"avg_requests_per_second":    getFloat(value["avgRequestsPerSecond"]),
		"avg_time_per_request":       getFloat(value["avgTimePerRequest"]),
		"errors":                     getInt(value["errors"]),
		"handler_start":              getInt(value["handlerStart"]),
		"median_request_time":        getFloat(value["medianRequestTime"]),
		"requests":                   getInt(value["requests"]),
		"timeouts":                   getInt(value["timeouts"]),
		"total_time":                 getFloat(value["totalTime"]),
	}
}

// Add update metrics section to accumulator
func addUpdateHandlerMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, measurementTime time.Time) error {
	var updateMetrics map[string]UpdateHandler

	if len(mBeansData.SolrMbeans) < 6 {
		return fmt.Errorf("no update handler metric data to unmarshal")
	}
	if err := json.Unmarshal(mBeansData.SolrMbeans[5], &updateMetrics); err != nil {
		return err
	}
	for name, metrics := range updateMetrics {
		var autoCommitMaxTime int64
		if len(metrics.Stats.AutocommitMaxTime) > 2 {
			autoCommitMaxTime, _ = strconv.ParseInt(metrics.Stats.AutocommitMaxTime[:len(metrics.Stats.AutocommitMaxTime)-2], 0, 64)
		}
		coreFields := map[string]interface{}{
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
			measurementTime,
		)
	}
	return nil
}

// Get float64 from interface
func getFloat(unk interface{}) float64 {
	switch i := unk.(type) {
	case float64:
		return i
	case string:
		f, err := strconv.ParseFloat(i, 64)
		if err != nil || math.IsNaN(f) {
			return float64(0)
		}
		return f
	default:
		return float64(0)
	}
}

// Get int64 from interface
func getInt(unk interface{}) int64 {
	switch i := unk.(type) {
	case int64:
		return i
	case float64:
		return int64(i)
	case string:
		v, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			return int64(0)
		}
		return v
	default:
		return int64(0)
	}
}

// Add cache metrics section to accumulator
func addCacheMetricsToAcc(acc telegraf.Accumulator, core string, mBeansData *MBeansData, measurementTime time.Time) error {
	if len(mBeansData.SolrMbeans) < 8 {
		return fmt.Errorf("no cache metric data to unmarshal")
	}
	var cacheMetrics map[string]Cache
	if err := json.Unmarshal(mBeansData.SolrMbeans[7], &cacheMetrics); err != nil {
		return err
	}
	for name, metrics := range cacheMetrics {
		coreFields := make(map[string]interface{})
		for key, value := range metrics.Stats {
			splitKey := strings.Split(key, ".")
			newKey := splitKey[len(splitKey)-1]
			switch newKey {
			case "cumulative_evictions",
				"cumulative_hits",
				"cumulative_inserts",
				"cumulative_lookups",
				"eviction",
				"hits",
				"inserts",
				"lookups",
				"size",
				"evictions":
				coreFields[newKey] = getInt(value)
			case "hitratio",
				"cumulative_hitratio":
				coreFields[newKey] = getFloat(value)
			case "warmupTime":
				coreFields["warmup_time"] = getInt(value)
			default:
				continue
			}
		}
		acc.AddFields(
			"solr_cache",
			coreFields,
			map[string]string{
				"core":    core,
				"handler": name},
			measurementTime,
		)
	}
	return nil
}

// Provide admin url
func (s *Solr) adminURL(server string) string {
	return fmt.Sprintf("%s%s", server, adminCoresPath)
}

// Provide mbeans url
func (s *Solr) mbeansURL(server string, core string) string {
	return fmt.Sprintf("%s/solr/%s%s", server, core, mbeansPath)
}

func (s *Solr) createHTTPClient() *http.Client {
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(s.HTTPTimeout),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(s.HTTPTimeout),
	}

	return client
}

func (s *Solr) gatherData(url string, v interface{}) error {
	req, reqErr := http.NewRequest(http.MethodGet, url, nil)
	if reqErr != nil {
		return reqErr
	}

	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}

	req.Header.Set("User-Agent", internal.ProductToken())

	r, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("solr: API responded with status-code %d, expected %d, url %s",
			r.StatusCode, http.StatusOK, url)
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func init() {
	inputs.Add("solr", func() telegraf.Input {
		return NewSolr()
	})
}
