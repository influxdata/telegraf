package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type metaData struct {
	Current JSONFloat64 `json:"current"`
	Sum     JSONFloat64 `json:"sum"`
	Mean    JSONFloat64 `json:"mean"`
	Stddev  JSONFloat64 `json:"stddev"`
	Min     JSONFloat64 `json:"min"`
	Max     JSONFloat64 `json:"max"`
	Value   JSONFloat64 `json:"value"`
}

type JSONFloat64 float64

// Used to restore Couchdb <2.0 API behavior
func (j *JSONFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = JSONFloat64(0)
		return nil
	}
	var temp float64
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	*j = JSONFloat64(temp)
	return nil
}

type oldValue struct {
	Value metaData `json:"value"`
	metaData
}

type couchdb struct {
	AuthCacheHits       metaData            `json:"auth_cache_hits"`
	AuthCacheMisses     metaData            `json:"auth_cache_misses"`
	DatabaseWrites      metaData            `json:"database_writes"`
	DatabaseReads       metaData            `json:"database_reads"`
	OpenDatabases       metaData            `json:"open_databases"`
	OpenOsFiles         metaData            `json:"open_os_files"`
	RequestTime         oldValue            `json:"request_time"`
	HttpdRequestMethods httpdRequestMethods `json:"httpd_request_methods"`
	HttpdStatusCodes    httpdStatusCodes    `json:"httpd_status_codes"`
}
type httpdRequestMethods struct {
	Put    metaData `json:"PUT"`
	Get    metaData `json:"GET"`
	Copy   metaData `json:"COPY"`
	Delete metaData `json:"DELETE"`
	Post   metaData `json:"POST"`
	Head   metaData `json:"HEAD"`
}
type httpdStatusCodes struct {
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
}
type httpd struct {
	BulkRequests             metaData `json:"bulk_requests"`
	Requests                 metaData `json:"requests"`
	TemporaryViewReads       metaData `json:"temporary_view_reads"`
	ViewReads                metaData `json:"view_reads"`
	ClientsRequestingChanges metaData `json:"clients_requesting_changes"`
}
type Stats struct {
	Couchdb             couchdb             `json:"couchdb"`
	HttpdRequestMethods httpdRequestMethods `json:"httpd_request_methods"`
	HttpdStatusCodes    httpdStatusCodes    `json:"httpd_status_codes"`
	Httpd               httpd               `json:"httpd"`
}

type CouchDB struct {
	HOSTs []string `toml:"hosts"`
}

func (*CouchDB) Description() string {
	return "Read CouchDB Stats from one or more servers"
}

func (*CouchDB) SampleConfig() string {
	return `
  ## Works with CouchDB stats endpoints out of the box
  ## Multiple HOSTs from which to read CouchDB stats:
  hosts = ["http://localhost:8086/_stats"]
`
}

func (c *CouchDB) Gather(accumulator telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, u := range c.HOSTs {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if err := c.fetchAndInsertData(accumulator, host); err != nil {
				accumulator.AddError(fmt.Errorf("[host=%s]: %s", host, err))
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

func (c *CouchDB) fetchAndInsertData(accumulator telegraf.Accumulator, host string) error {
	response, error := client.Get(host)
	if error != nil {
		return error
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Failed to get stats from couchdb: HTTP responded %d", response.StatusCode)
	}

	stats := newStats()
	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&stats)

	fields := map[string]interface{}{}

	// for couchdb 2.0 API changes
	stats.Couchdb.RequestTime.Value = metaData{
		Current: stats.Couchdb.RequestTime.Current,
		Sum:     stats.Couchdb.RequestTime.Sum,
		Mean:    stats.Couchdb.RequestTime.Mean,
		Stddev:  stats.Couchdb.RequestTime.Stddev,
		Min:     stats.Couchdb.RequestTime.Min,
		Max:     stats.Couchdb.RequestTime.Max,
	}
	requestTime := stats.Couchdb.RequestTime.Value
	openOSFiles := stats.Couchdb.OpenOsFiles

	httpdRequestMethodsPut := stats.HttpdRequestMethods.Put
	httpdRequestMethodsGet := stats.HttpdRequestMethods.Get
	httpdRequestMethodsCopy := stats.HttpdRequestMethods.Copy
	httpdRequestMethodsDelete := stats.HttpdRequestMethods.Delete
	httpdRequestMethodsPost := stats.HttpdRequestMethods.Post
	httpdRequestMethodsHead := stats.HttpdRequestMethods.Head

	httpdStatusCodesStatus200 := stats.HttpdStatusCodes.Status200
	httpdStatusCodesStatus201 := stats.HttpdStatusCodes.Status201
	httpdStatusCodesStatus202 := stats.HttpdStatusCodes.Status202
	httpdStatusCodesStatus301 := stats.HttpdStatusCodes.Status301
	httpdStatusCodesStatus304 := stats.HttpdStatusCodes.Status304
	httpdStatusCodesStatus400 := stats.HttpdStatusCodes.Status400
	httpdStatusCodesStatus401 := stats.HttpdStatusCodes.Status401
	httpdStatusCodesStatus403 := stats.HttpdStatusCodes.Status403
	httpdStatusCodesStatus404 := stats.HttpdStatusCodes.Status404
	httpdStatusCodesStatus405 := stats.HttpdStatusCodes.Status405
	httpdStatusCodesStatus409 := stats.HttpdStatusCodes.Status409
	httpdStatusCodesStatus412 := stats.HttpdStatusCodes.Status412
	httpdStatusCodesStatus500 := stats.HttpdStatusCodes.Status500
	// check if couchdb2.0 is used
	if stats.Couchdb.HttpdRequestMethods.Get.Value > -1 {
		requestTime = stats.Couchdb.RequestTime.Value

		httpdRequestMethodsPut = stats.Couchdb.HttpdRequestMethods.Put
		httpdRequestMethodsGet = stats.Couchdb.HttpdRequestMethods.Get
		httpdRequestMethodsCopy = stats.Couchdb.HttpdRequestMethods.Copy
		httpdRequestMethodsDelete = stats.Couchdb.HttpdRequestMethods.Delete
		httpdRequestMethodsPost = stats.Couchdb.HttpdRequestMethods.Post
		httpdRequestMethodsHead = stats.Couchdb.HttpdRequestMethods.Head

		httpdStatusCodesStatus200 = stats.Couchdb.HttpdStatusCodes.Status200
		httpdStatusCodesStatus201 = stats.Couchdb.HttpdStatusCodes.Status201
		httpdStatusCodesStatus202 = stats.Couchdb.HttpdStatusCodes.Status202
		httpdStatusCodesStatus301 = stats.Couchdb.HttpdStatusCodes.Status301
		httpdStatusCodesStatus304 = stats.Couchdb.HttpdStatusCodes.Status304
		httpdStatusCodesStatus400 = stats.Couchdb.HttpdStatusCodes.Status400
		httpdStatusCodesStatus401 = stats.Couchdb.HttpdStatusCodes.Status401
		httpdStatusCodesStatus403 = stats.Couchdb.HttpdStatusCodes.Status403
		httpdStatusCodesStatus404 = stats.Couchdb.HttpdStatusCodes.Status404
		httpdStatusCodesStatus405 = stats.Couchdb.HttpdStatusCodes.Status405
		httpdStatusCodesStatus409 = stats.Couchdb.HttpdStatusCodes.Status409
		httpdStatusCodesStatus412 = stats.Couchdb.HttpdStatusCodes.Status412
		httpdStatusCodesStatus500 = stats.Couchdb.HttpdStatusCodes.Status500
	}

	// CouchDB meta stats:
	c.MapCopy(fields, c.generateFields("couchdb_auth_cache_misses", stats.Couchdb.AuthCacheMisses))
	c.MapCopy(fields, c.generateFields("couchdb_database_writes", stats.Couchdb.DatabaseWrites))
	c.MapCopy(fields, c.generateFields("couchdb_open_databases", stats.Couchdb.OpenDatabases))
	c.MapCopy(fields, c.generateFields("couchdb_auth_cache_hits", stats.Couchdb.AuthCacheHits))
	c.MapCopy(fields, c.generateFields("couchdb_request_time", requestTime))
	c.MapCopy(fields, c.generateFields("couchdb_database_reads", stats.Couchdb.DatabaseReads))
	c.MapCopy(fields, c.generateFields("couchdb_open_os_files", openOSFiles))

	// http request methods stats:
	c.MapCopy(fields, c.generateFields("httpd_request_methods_put", httpdRequestMethodsPut))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_get", httpdRequestMethodsGet))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_copy", httpdRequestMethodsCopy))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_delete", httpdRequestMethodsDelete))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_post", httpdRequestMethodsPost))
	c.MapCopy(fields, c.generateFields("httpd_request_methods_head", httpdRequestMethodsHead))

	// status code stats:
	c.MapCopy(fields, c.generateFields("httpd_status_codes_200", httpdStatusCodesStatus200))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_201", httpdStatusCodesStatus201))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_202", httpdStatusCodesStatus202))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_301", httpdStatusCodesStatus301))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_304", httpdStatusCodesStatus304))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_400", httpdStatusCodesStatus400))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_401", httpdStatusCodesStatus401))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_403", httpdStatusCodesStatus403))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_404", httpdStatusCodesStatus404))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_405", httpdStatusCodesStatus405))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_409", httpdStatusCodesStatus409))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_412", httpdStatusCodesStatus412))
	c.MapCopy(fields, c.generateFields("httpd_status_codes_500", httpdStatusCodesStatus500))

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
	switch v := value.(type) {
	case JSONFloat64:
		return float64(v)
	}
	return value
}

func (c *CouchDB) generateFields(prefix string, obj metaData) map[string]interface{} {
	fields := map[string]interface{}{}
	if obj.Value > -1 {
		fields[prefix+"_value"] = c.safeCheck(obj.Value)
	}
	if obj.Current > -1 {
		fields[prefix+"_current"] = c.safeCheck(obj.Current)
	}
	if obj.Sum > -1 {
		fields[prefix+"_sum"] = c.safeCheck(obj.Sum)
	}
	if obj.Mean > -1 {
		fields[prefix+"_mean"] = c.safeCheck(obj.Mean)
	}
	if obj.Stddev > -1 {
		fields[prefix+"_stddev"] = c.safeCheck(obj.Stddev)
	}
	if obj.Min > -1 {
		fields[prefix+"_min"] = c.safeCheck(obj.Min)
	}
	if obj.Max > -1 {
		fields[prefix+"_max"] = c.safeCheck(obj.Max)
	}
	return fields
}

func newStats() Stats {
	return Stats{
		Couchdb: couchdb{
			AuthCacheHits:   newMeta(),
			AuthCacheMisses: newMeta(),
			DatabaseWrites:  newMeta(),
			DatabaseReads:   newMeta(),
			OpenDatabases:   newMeta(),
			OpenOsFiles:     newMeta(),
			RequestTime:     oldValue{Value: newMeta()},
			HttpdRequestMethods: httpdRequestMethods{
				Put:    newMeta(),
				Get:    newMeta(),
				Copy:   newMeta(),
				Delete: newMeta(),
				Post:   newMeta(),
				Head:   newMeta(),
			},
			HttpdStatusCodes: httpdStatusCodes{
				Status200: newMeta(),
				Status201: newMeta(),
				Status202: newMeta(),
				Status301: newMeta(),
				Status304: newMeta(),
				Status400: newMeta(),
				Status401: newMeta(),
				Status403: newMeta(),
				Status404: newMeta(),
				Status405: newMeta(),
				Status409: newMeta(),
				Status412: newMeta(),
				Status500: newMeta(),
			},
		},
		HttpdRequestMethods: httpdRequestMethods{
			Put:    newMeta(),
			Get:    newMeta(),
			Copy:   newMeta(),
			Delete: newMeta(),
			Post:   newMeta(),
			Head:   newMeta(),
		},
		HttpdStatusCodes: httpdStatusCodes{
			Status200: newMeta(),
			Status201: newMeta(),
			Status202: newMeta(),
			Status301: newMeta(),
			Status304: newMeta(),
			Status400: newMeta(),
			Status401: newMeta(),
			Status403: newMeta(),
			Status404: newMeta(),
			Status405: newMeta(),
			Status409: newMeta(),
			Status412: newMeta(),
			Status500: newMeta(),
		},
		Httpd: httpd{
			BulkRequests:             newMeta(),
			Requests:                 newMeta(),
			TemporaryViewReads:       newMeta(),
			ViewReads:                newMeta(),
			ClientsRequestingChanges: newMeta(),
		},
	}

}

func newMeta() metaData {
	return metaData{
		Current: -1,
		Sum:     -1,
		Mean:    -1,
		Stddev:  -1,
		Min:     -1,
		Max:     -1,
		Value:   -1,
	}
}

func init() {
	inputs.Add("couchdb", func() telegraf.Input {
		return &CouchDB{}
	})
}
