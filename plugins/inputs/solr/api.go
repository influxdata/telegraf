package solr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
)

type apiConfig struct {
	endpointAdmin    string
	endpointMBeans   string
	keyCore          string
	keyCache         string
	keyUpdateHandler string
	keyQueryHandler  string
}

func newAPIv1Config() *apiConfig {
	return &apiConfig{
		endpointAdmin:    "/solr/admin/cores?action=STATUS&wt=json",
		endpointMBeans:   "/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERYHANDLER&cat=UPDATEHANDLER&cat=CACHE",
		keyCore:          `"CORE"`,
		keyCache:         `"CACHE"`,
		keyQueryHandler:  `"QUERYHANDLER"`,
		keyUpdateHandler: `"UPDATEHANDLER"`,
	}
}

func newAPIv2Config() *apiConfig {
	return &apiConfig{
		endpointAdmin:    "/solr/admin/cores?action=STATUS&wt=json",
		endpointMBeans:   "/admin/mbeans?stats=true&wt=json&cat=CORE&cat=QUERY&cat=UPDATE&cat=CACHE",
		keyCore:          `"CORE"`,
		keyCache:         `"CACHE"`,
		keyQueryHandler:  `"QUERY"`,
		keyUpdateHandler: `"UPDATE"`,
	}
}

func (cfg *apiConfig) adminEndpoint(server string) string {
	return strings.TrimSuffix(server, "/") + cfg.endpointAdmin
}

func (cfg *apiConfig) mbeansEndpoint(server, core string) string {
	return strings.TrimSuffix(server, "/") + "/solr/" + strings.Trim(core, "/") + cfg.endpointMBeans
}

func (cfg *apiConfig) parseCore(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the core information element
	var coreData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == cfg.keyCore {
			coreData = data.SolrMbeans[i+1]
			break
		}
	}
	if coreData == nil {
		acc.AddError(errors.New("no core metric data to unmarshal"))
		return
	}

	var coreMetrics map[string]Core
	if err := json.Unmarshal(coreData, &coreMetrics); err != nil {
		acc.AddError(fmt.Errorf("unmarshalling core metrics for %q failed: %w", core, err))
		return
	}

	for name, m := range coreMetrics {
		if strings.Contains(name, "@") {
			continue
		}
		fields := map[string]interface{}{
			"deleted_docs": m.Stats.DeletedDocs,
			"max_docs":     m.Stats.MaxDoc,
			"num_docs":     m.Stats.NumDocs,
		}
		tags := map[string]string{
			"core":    core,
			"handler": name,
		}

		acc.AddFields("solr_core", fields, tags, ts)
	}
}

func (cfg *apiConfig) parseCache(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the cache information element
	var cacheData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == cfg.keyCache {
			cacheData = data.SolrMbeans[i+1]
			break
		}
	}
	if cacheData == nil {
		acc.AddError(errors.New("no cache metric data to unmarshal"))
		return
	}

	var cacheMetrics map[string]Cache
	if err := json.Unmarshal(cacheData, &cacheMetrics); err != nil {
		acc.AddError(fmt.Errorf("unmarshalling update handler for %q failed: %w", core, err))
		return
	}

	for name, metrics := range cacheMetrics {
		fields := make(map[string]interface{}, len(metrics.Stats))
		for key, value := range metrics.Stats {
			splitKey := strings.Split(key, ".")
			newKey := splitKey[len(splitKey)-1]
			switch newKey {
			case "cumulative_evictions", "cumulative_hits", "cumulative_inserts", "cumulative_lookups",
				"eviction", "evictions", "hits", "inserts", "lookups", "size":
				fields[newKey] = getInt(value)
			case "hitratio", "cumulative_hitratio":
				fields[newKey] = getFloat(value)
			case "warmupTime":
				fields["warmup_time"] = getInt(value)
			default:
				continue
			}
		}

		tags := map[string]string{
			"core":    core,
			"handler": name,
		}

		acc.AddFields("solr_cache", fields, tags, ts)
	}
}

func (cfg *apiConfig) parseQueryHandler(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the query-handler information element
	var queryData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == cfg.keyQueryHandler {
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

func (cfg *apiConfig) parseUpdateHandler(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the update-handler information element
	var updateData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == cfg.keyUpdateHandler {
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

	for name, metrics := range updateMetrics {
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
			"handler": name,
		}

		acc.AddFields("solr_updatehandler", fields, tags, ts)
	}
}
