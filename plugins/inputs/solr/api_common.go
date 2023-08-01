package solr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

func (s *Solr) query(endpoint string, v interface{}) error {
	req, reqErr := http.NewRequest(http.MethodGet, endpoint, nil)
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
		return fmt.Errorf("solr: API endpoint %q responded with %q", endpoint, r.Status)
	}

	return json.NewDecoder(r.Body).Decode(v)
}

func parseCore(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the core information element
	var coreData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == `"CORE"` {
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

func parseCache(acc telegraf.Accumulator, core string, data *MBeansData, ts time.Time) {
	// Determine the cache information element
	var cacheData json.RawMessage
	for i := 0; i < len(data.SolrMbeans); i += 2 {
		if string(data.SolrMbeans[i]) == `"CACHE"` {
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
