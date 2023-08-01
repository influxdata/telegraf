package solr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

const apiV1MbeansEndpoint = "/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERYHANDLER&cat=UPDATEHANDLER&cat=CACHE"

type metricCollectorV1 struct {
	s        *Solr
	username string
	password string
	filter   filter.Filter
}

func newCollectorV1(s *Solr, coreFilters []string) (MetricCollector, error) {
	f, err := filter.Compile(coreFilters)
	if err != nil {
		return nil, err
	}
	collector := &metricCollectorV1{
		s:      s,
		filter: f,
	}
	return collector, nil
}

// Gather all metrics from server
func (c *metricCollectorV1) Collect(acc telegraf.Accumulator, server string) {
	now := time.Now()

	var coreStatus AdminCoresStatus
	endpoint := server + "/solr/admin/cores?action=STATUS&wt=json"
	if err := c.s.query(endpoint, &coreStatus); err != nil {
		acc.AddError(err)
		return
	}

	var wg sync.WaitGroup
	for core, metrics := range coreStatus.Status {
		fields := map[string]interface{}{
			"deleted_docs":  metrics.Index.DeletedDocs,
			"max_docs":      metrics.Index.MaxDoc,
			"num_docs":      metrics.Index.NumDocs,
			"size_in_bytes": metrics.Index.SizeInBytes,
		}
		tags := map[string]string{"core": core}
		acc.AddFields("solr_admin", fields, tags, now)

		if c.filter != nil && !c.filter.Match(core) {
			continue
		}

		wg.Add(1)
		go func(server string, core string) {
			defer wg.Done()

			endpoint := server + "/solr/" + core + apiV1MbeansEndpoint
			var data MBeansData
			if err := c.s.query(endpoint, &data); err != nil {
				acc.AddError(err)
				return
			}

			parseCore(acc, core, &data, now)
			parseQueryHandlerV1(acc, core, &data, now)
			parseUpdateHandlerV1(acc, core, &data, now)
			parseCache(acc, core, &data, now)
		}(server, core)
	}
	wg.Wait()
}

func parseQueryHandlerV1(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the query-handler information element
	var queryData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == `"QUERYHANDLER"` {
			queryData = data.SolrMbeans[i+1]
			break
		}
	}
	if queryData == nil {
		acc.AddError(errors.New("no query handler metric data to unmarshal"))
		return
	}

	var queryMetrics map[string]QueryHandler
	if err := json.Unmarshal(queryData, &queryMetrics); err != nil {
		acc.AddError(fmt.Errorf("unmarshalling query handler for %q failed: %w", core, err))
		return
	}

	for name, metrics := range queryMetrics {
		if metrics.Stats == nil {
			continue
		}

		var values map[string]interface{}
		switch v := metrics.Stats.(type) {
		case []interface{}:
			values = make(map[string]interface{}, len(v)/2)
			for i := 0; i < len(v); i += 2 {
				key, ok := v[i].(string)
				if !ok {
					continue
				}
				values[key] = v[i+1]
			}
		case map[string]interface{}:
			values = v
		default:
			continue
		}

		fields := map[string]interface{}{
			"15min_rate_reqs_per_second": getFloat(values["15minRateReqsPerSecond"]),
			"5min_rate_reqs_per_second":  getFloat(values["5minRateReqsPerSecond"]),
			"75th_pc_request_time":       getFloat(values["75thPcRequestTime"]),
			"95th_pc_request_time":       getFloat(values["95thPcRequestTime"]),
			"99th_pc_request_time":       getFloat(values["99thPcRequestTime"]),
			"999th_pc_request_time":      getFloat(values["999thPcRequestTime"]),
			"avg_requests_per_second":    getFloat(values["avgRequestsPerSecond"]),
			"avg_time_per_request":       getFloat(values["avgTimePerRequest"]),
			"errors":                     getInt(values["errors"]),
			"handler_start":              getInt(values["handlerStart"]),
			"median_request_time":        getFloat(values["medianRequestTime"]),
			"requests":                   getInt(values["requests"]),
			"timeouts":                   getInt(values["timeouts"]),
			"total_time":                 getFloat(values["totalTime"]),
		}

		tags := map[string]string{
			"core":    core,
			"handler": name,
		}
		acc.AddFields("solr_queryhandler", fields, tags, ts)
	}
}

func parseUpdateHandlerV1(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the update-handler information element
	var updateData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == `"UPDATEHANDLER"` {
			updateData = data.SolrMbeans[i+1]
			break
		}
	}
	if updateData == nil {
		acc.AddError(errors.New("no update handler metric data to unmarshal"))
		return
	}

	var updateMetrics map[string]UpdateHandler
	if err := json.Unmarshal(updateData, &updateMetrics); err != nil {
		acc.AddError(fmt.Errorf("unmarshalling update handler for %q failed: %w", core, err))
		return
	}

	metrics, found := updateMetrics["updateHandler"]
	if !found {
		return
	}
	var autoCommitMaxTime int64
	if len(metrics.Stats.AutocommitMaxTime) > 2 {
		s := metrics.Stats.AutocommitMaxTime[:len(metrics.Stats.AutocommitMaxTime)-2]
		var err error
		autoCommitMaxTime, err = strconv.ParseInt(s, 0, 64)
		if err != nil {
			autoCommitMaxTime = 0
		}
	}

	fields := map[string]interface{}{
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

	tags := map[string]string{
		"core":    core,
		"handler": "updateHandler",
	}

	acc.AddFields("solr_updatehandler", fields, tags, ts)
}
