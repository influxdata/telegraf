package couchdb

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Schema:
type metaData struct {
	Description string  `json:"description"`
	Current     float64 `json:"current"`
	Sum         float64 `json:"sum"`
	Mean        float64 `json:"mean"`
	Stddev      float64 `json:"stddev"`
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
}

type ServerStats struct {
	Couchdb struct {
		AuthCacheMisses metaData `json:"auth_cache_misses"`
		DatabaseWrites  metaData `json:"database_writes"`
		OpenDatabases   metaData `json:"open_databases"`
		AuthCacheHits   metaData `json:"auth_cache_hits"`
		RequestTime     metaData `json:"request_time"`
		DatabaseReads   metaData `json:"database_reads"`
		OpenOsFiles     metaData `json:"open_os_files"`
	} `json:"couchdb"`
	HttpdRequestMethods struct {
		Put    metaData `json:"PUT"`
		Get    metaData `json:"GET"`
		Copy   metaData `json:"COPY"`
		Delete metaData `json:"DELETE"`
		Post   metaData `json:"POST"`
		Head   metaData `json:"HEAD"`
	} `json:"httpd_request_methods"`
	HttpdStatusCodes struct {
		Status200 metaData `json:"200"`
		Status201 metaData `json:"201"`
		Status202 metaData `json:"202"`
		Status301 metaData `json:"301"`
		Status304 metaData `json:"304"`
		Status400 metaData `json:"400"`
		Status401 metaData `json:"401"`
		Status403 metaData `json:"403"`
		Status404 metaData `json:"404"`
		Status405 metaData `json:"405"`
		Status409 metaData `json:"409"`
		Status412 metaData `json:"412"`
		Status500 metaData `json:"500"`
	} `json:"httpd_status_codes"`
	Httpd struct {
		ClientsRequestingChanges metaData `json:"clients_requesting_changes"`
		TemporaryViewReads       metaData `json:"temporary_view_reads"`
		Requests                 metaData `json:"requests"`
		BulkRequests             metaData `json:"bulk_requests"`
		ViewReads                metaData `json:"view_reads"`
	} `json:"httpd"`
}

type CouchDB struct {
	HOSTs []string `toml:"hosts"`
}

func (*CouchDB) Description() string {
	return "Read CouchDB Stats from one or more servers"
}

func (*CouchDB) SampleConfig() string {
	return `
  ## Works with CouchDB stats/all_dbs endpoints out of the box
  ## You will get the stats that are associated with the endpoint
  ## Multiple HOSTs from which to read CouchDB stats:
  hosts = ["http://localhost:5984/_stats","http://localhost:5984/_all_dbs"]
`
}

func (c *CouchDB) Gather(accumulator telegraf.Accumulator) error {

	var wg sync.WaitGroup
	for _, u := range c.HOSTs {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if strings.HasSuffix(host, "_stats") {
				if err := c.fetchAndInsertServerData(accumulator, host); err != nil {
					accumulator.AddError(fmt.Errorf("[host=%s]: %s", host, err))
				}
			}
			if strings.HasSuffix(host, "_all_dbs") {
				if err := c.fetchAndInsertDbData(accumulator, host); err != nil {
					accumulator.AddError(fmt.Errorf("[host=%s]: %s", host, err))
				}
			}
		}(u)
	}

	wg.Wait()

	return nil
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

func (c *CouchDB) fetchAndInsertServerData(accumulator telegraf.Accumulator, host string) error {

	response, error := client.Get(host)
	if error != nil {
		return error
	}
	defer response.Body.Close()

	var stats ServerStats
	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&stats)

	fields := map[string]interface{}{}

	// CouchDB meta stats:
	c.MapCopy(fields, c.generateFields("couchdb_auth_cache_misses", stats.Couchdb.AuthCacheMisses))
	c.MapCopy(fields, c.generateFields("couchdb_database_writes", stats.Couchdb.DatabaseWrites))
	c.MapCopy(fields, c.generateFields("couchdb_open_databases", stats.Couchdb.OpenDatabases))
	c.MapCopy(fields, c.generateFields("couchdb_auth_cache_hits", stats.Couchdb.AuthCacheHits))
	c.MapCopy(fields, c.generateFields("couchdb_request_time", stats.Couchdb.RequestTime))
	c.MapCopy(fields, c.generateFields("couchdb_database_reads", stats.Couchdb.DatabaseReads))
	c.MapCopy(fields, c.generateFields("couchdb_open_os_files", stats.Couchdb.OpenOsFiles))

	// http request methods stats:
	c.MapCopy(fields, c.generateFields("httpd_request_methods_put", stats.HttpdRequestMethods.Put))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_get", stats.HttpdRequestMethods.Get))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_copy", stats.HttpdRequestMethods.Copy))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_delete", stats.HttpdRequestMethods.Delete))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_post", stats.HttpdRequestMethods.Post))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_head", stats.HttpdRequestMethods.Head))

	// status code stats:
	c.MapCopy(fields, c.generateFields("httpd_status_codes_200", stats.HttpdStatusCodes.Status200))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_201", stats.HttpdStatusCodes.Status201))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_202", stats.HttpdStatusCodes.Status202))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_301", stats.HttpdStatusCodes.Status301))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_304", stats.HttpdStatusCodes.Status304))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_400", stats.HttpdStatusCodes.Status400))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_401", stats.HttpdStatusCodes.Status401))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_403", stats.HttpdStatusCodes.Status403))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_404", stats.HttpdStatusCodes.Status404))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_405", stats.HttpdStatusCodes.Status405))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_409", stats.HttpdStatusCodes.Status409))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_412", stats.HttpdStatusCodes.Status412))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_500", stats.HttpdStatusCodes.Status500))

	// httpd stats:
	c.MapCopy(fields, c.generateFields("httpd_clients_requesting_changes", stats.Httpd.ClientsRequestingChanges))
	c.MapCopy(fields, c.generateFields("httpd_temporary_view_reads", stats.Httpd.TemporaryViewReads))
	c.MapCopy(fields, c.generateFields("httpd_requests", stats.Httpd.Requests))
	c.MapCopy(fields, c.generateFields("httpd_bulk_requests", stats.Httpd.BulkRequests))
	c.MapCopy(fields, c.generateFields("httpd_view_reads", stats.Httpd.ViewReads))

	tags := map[string]string{
		"server": host,
	}
	accumulator.AddFields("couchdb", fields, tags)
	return nil
}

func (c *CouchDB) fetchAndInsertDbData(accumulator telegraf.Accumulator, host string) error {

	response, error := client.Get(host)
	if error != nil {
		return error
	}
	defer response.Body.Close()

	var dbs []string

	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&dbs)

	for _, db := range dbs {

		fields := map[string]interface{}{}

		response, error := client.Get(strings.Replace(host, "_all_dbs", db, -1))
		if error != nil {
			return error
		}
		defer response.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		var dat map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dat); err != nil {
			return error
		}
		// see http://docs.couchdb.org/en/2.1.0/api/database/common.html for details. Deprecated fields are not mapped
		fields["doc_count"] = dat["doc_count"]
		fields["doc_del_count"] = dat["doc_del_count"]
		fields["compact_running"] = translateBoolToCounter(dat["compact_running"] == "true")
		sizes := dat["sizes"].(map[string]interface{})

		fields["file_size"] = sizes["file"]
		fields["external_size"] = sizes["external"]
		fields["active_size"] = sizes["active"]
		fields["compact_running"] = translateBoolToCounter(dat["compact_running"] == "true")

		tags := map[string]string{
			"server": host,
			"db":     db,
		}

		accumulator.AddFields("couchdb", fields, tags)
	}

	return nil
}

func translateBoolToCounter(v bool) int {
	if v {
		return 1
	} else {
		return 0
	}
}

func (*CouchDB) MapCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)
	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}

func (*CouchDB) safeCheck(value interface{}) interface{} {
	if value == nil {
		return 0.0
	}
	return value
}

func (c *CouchDB) generateFields(prefix string, obj metaData) map[string]interface{} {
	fields := map[string]interface{}{
		prefix + "_current": c.safeCheck(obj.Current),
		prefix + "_sum":     c.safeCheck(obj.Sum),
		prefix + "_mean":    c.safeCheck(obj.Mean),
		prefix + "_stddev":  c.safeCheck(obj.Stddev),
		prefix + "_min":     c.safeCheck(obj.Min),
		prefix + "_max":     c.safeCheck(obj.Max),
	}
	return fields
}

func init() {
	inputs.Add("couchdb", func() telegraf.Input {
		return &CouchDB{}
	})
}
