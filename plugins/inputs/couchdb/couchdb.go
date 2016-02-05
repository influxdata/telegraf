package couchdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net/http"
	"reflect"
	"strings"
	"sync"
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

type Stats struct {
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
	  # Works with CouchDB stats endpoints out of the box
	  # Multiple HOSTs from which to read CouchDB stats:
	  hosts = ["http://localhost:8086/_stats",...]
	`
}

func (this *CouchDB) Gather(accumulator telegraf.Accumulator) error {
	errorChannel := make(chan error, len(this.HOSTs))
	var wg sync.WaitGroup
	for _, u := range this.HOSTs {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if err := this.fetchAndInsertData(accumulator, host); err != nil {
				errorChannel <- fmt.Errorf("[host=%s]: %s", host, err)
			}
		}(u)
	}

	wg.Wait()
	close(errorChannel)

	// If there weren't any errors, we can return nil now.
	if len(errorChannel) == 0 {
		return nil
	}

	// There were errors, so join them all together as one big error.
	errorStrings := make([]string, 0, len(errorChannel))
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	return errors.New(strings.Join(errorStrings, "\n"))

}

func (this *CouchDB) fetchAndInsertData(accumulator telegraf.Accumulator, host string) error {

	response, error := http.Get(host)
	if error != nil {
		return error
	}
	defer response.Body.Close()

	var stats Stats
	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&stats)

	fields := map[string]interface{}{}

	// CouchDB meta stats:
	this.MapCopy(fields, this.generateFields("couchdb_auth_cache_misses", stats.Couchdb.AuthCacheMisses))
	this.MapCopy(fields, this.generateFields("couchdb_database_writes", stats.Couchdb.DatabaseWrites))
	this.MapCopy(fields, this.generateFields("couchdb_open_databases", stats.Couchdb.OpenDatabases))
	this.MapCopy(fields, this.generateFields("couchdb_auth_cache_hits", stats.Couchdb.AuthCacheHits))
	this.MapCopy(fields, this.generateFields("couchdb_request_time", stats.Couchdb.RequestTime))
	this.MapCopy(fields, this.generateFields("couchdb_database_reads", stats.Couchdb.DatabaseReads))
	this.MapCopy(fields, this.generateFields("couchdb_open_os_files", stats.Couchdb.OpenOsFiles))

	// http request methods stats:
	this.MapCopy(fields, this.generateFields("httpd_request_methods_put", stats.HttpdRequestMethods.Put))
	this.MapCopy(fields, this.generateFields("httpd_request_methods_get", stats.HttpdRequestMethods.Get))
	this.MapCopy(fields, this.generateFields("httpd_request_methods_copy", stats.HttpdRequestMethods.Copy))
	this.MapCopy(fields, this.generateFields("httpd_request_methods_delete", stats.HttpdRequestMethods.Delete))
	this.MapCopy(fields, this.generateFields("httpd_request_methods_post", stats.HttpdRequestMethods.Post))
	this.MapCopy(fields, this.generateFields("httpd_request_methods_head", stats.HttpdRequestMethods.Head))

	// status code stats:
	this.MapCopy(fields, this.generateFields("httpd_status_codes_200", stats.HttpdStatusCodes.Status200))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_201", stats.HttpdStatusCodes.Status201))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_202", stats.HttpdStatusCodes.Status202))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_301", stats.HttpdStatusCodes.Status301))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_304", stats.HttpdStatusCodes.Status304))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_400", stats.HttpdStatusCodes.Status400))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_401", stats.HttpdStatusCodes.Status401))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_403", stats.HttpdStatusCodes.Status403))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_404", stats.HttpdStatusCodes.Status404))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_405", stats.HttpdStatusCodes.Status405))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_409", stats.HttpdStatusCodes.Status409))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_412", stats.HttpdStatusCodes.Status412))
	this.MapCopy(fields, this.generateFields("httpd_status_codes_500", stats.HttpdStatusCodes.Status500))

	// httpd stats:
	this.MapCopy(fields, this.generateFields("httpd_clients_requesting_changes", stats.Httpd.ClientsRequestingChanges))
	this.MapCopy(fields, this.generateFields("httpd_temporary_view_reads", stats.Httpd.TemporaryViewReads))
	this.MapCopy(fields, this.generateFields("httpd_requests", stats.Httpd.Requests))
	this.MapCopy(fields, this.generateFields("httpd_bulk_requests", stats.Httpd.BulkRequests))
	this.MapCopy(fields, this.generateFields("httpd_view_reads", stats.Httpd.ViewReads))

	tags := map[string]string{
		"server": host,
	}
	accumulator.AddFields("couchdb", fields, tags)
	return nil
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

func (this *CouchDB) generateFields(prefix string, obj metaData) map[string]interface{} {
	fields := map[string]interface{}{
		prefix + "_current": this.safeCheck(obj.Current),
		prefix + "_sum":     this.safeCheck(obj.Sum),
		prefix + "_mean":    this.safeCheck(obj.Mean),
		prefix + "_stddev":  this.safeCheck(obj.Stddev),
		prefix + "_min":     this.safeCheck(obj.Min),
		prefix + "_max":     this.safeCheck(obj.Max),
	}
	return fields
}

func init() {
	inputs.Add("couchdb", func() telegraf.Input {
		return &CouchDB{}
	})
}
