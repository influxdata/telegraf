package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type (
	metaData struct {
		Current *float64 `json:"current"`
		Sum     *float64 `json:"sum"`
		Mean    *float64 `json:"mean"`
		Stddev  *float64 `json:"stddev"`
		Min     *float64 `json:"min"`
		Max     *float64 `json:"max"`
		Value   *float64 `json:"value"`
	}

	oldValue struct {
		Value metaData `json:"value"`
		metaData
	}

	couchdb struct {
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

	httpdRequestMethods struct {
		Put    metaData `json:"PUT"`
		Get    metaData `json:"GET"`
		Copy   metaData `json:"COPY"`
		Delete metaData `json:"DELETE"`
		Post   metaData `json:"POST"`
		Head   metaData `json:"HEAD"`
	}

	httpdStatusCodes struct {
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

	httpd struct {
		BulkRequests             metaData `json:"bulk_requests"`
		Requests                 metaData `json:"requests"`
		TemporaryViewReads       metaData `json:"temporary_view_reads"`
		ViewReads                metaData `json:"view_reads"`
		ClientsRequestingChanges metaData `json:"clients_requesting_changes"`
	}

	Stats struct {
		Couchdb             couchdb             `json:"couchdb"`
		HttpdRequestMethods httpdRequestMethods `json:"httpd_request_methods"`
		HttpdStatusCodes    httpdStatusCodes    `json:"httpd_status_codes"`
		Httpd               httpd               `json:"httpd"`
	}

	CouchDB struct {
		Hosts         []string `toml:"hosts"`
		BasicUsername string   `toml:"basic_username"`
		BasicPassword string   `toml:"basic_password"`

		client *http.Client
	}
)

func (*CouchDB) Description() string {
	return "Read CouchDB Stats from one or more servers"
}

func (*CouchDB) SampleConfig() string {
	return `
  ## Works with CouchDB stats endpoints out of the box
  ## Multiple Hosts from which to read CouchDB stats:
  hosts = ["http://localhost:8086/_stats"]

  ## Use HTTP Basic Authentication.
  # basic_username = "telegraf"
  # basic_password = "p@ssw0rd"
`
}

func (c *CouchDB) Gather(accumulator telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, u := range c.Hosts {
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

func (c *CouchDB) fetchAndInsertData(accumulator telegraf.Accumulator, host string) error {
	if c.client == nil {
		c.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(3 * time.Second),
			},
			Timeout: time.Duration(4 * time.Second),
		}
	}

	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return err
	}

	if c.BasicUsername != "" || c.BasicPassword != "" {
		req.SetBasicAuth(c.BasicUsername, c.BasicPassword)
	}

	response, error := c.client.Do(req)
	if error != nil {
		return error
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Failed to get stats from couchdb: HTTP responded %d", response.StatusCode)
	}

	stats := Stats{}
	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&stats)

	fields := map[string]interface{}{}

	// for couchdb 2.0 API changes
	requestTime := metaData{
		Current: stats.Couchdb.RequestTime.Current,
		Sum:     stats.Couchdb.RequestTime.Sum,
		Mean:    stats.Couchdb.RequestTime.Mean,
		Stddev:  stats.Couchdb.RequestTime.Stddev,
		Min:     stats.Couchdb.RequestTime.Min,
		Max:     stats.Couchdb.RequestTime.Max,
	}

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
	if stats.Couchdb.HttpdRequestMethods.Get.Value != nil {
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
	c.generateFields(fields, "couchdb_auth_cache_misses", stats.Couchdb.AuthCacheMisses)
	c.generateFields(fields, "couchdb_database_writes", stats.Couchdb.DatabaseWrites)
	c.generateFields(fields, "couchdb_open_databases", stats.Couchdb.OpenDatabases)
	c.generateFields(fields, "couchdb_auth_cache_hits", stats.Couchdb.AuthCacheHits)
	c.generateFields(fields, "couchdb_request_time", requestTime)
	c.generateFields(fields, "couchdb_database_reads", stats.Couchdb.DatabaseReads)
	c.generateFields(fields, "couchdb_open_os_files", stats.Couchdb.OpenOsFiles)

	// http request methods stats:
	c.generateFields(fields, "httpd_request_methods_put", httpdRequestMethodsPut)
	c.generateFields(fields, "httpd_request_methods_get", httpdRequestMethodsGet)
	c.generateFields(fields, "httpd_request_methods_copy", httpdRequestMethodsCopy)
	c.generateFields(fields, "httpd_request_methods_delete", httpdRequestMethodsDelete)
	c.generateFields(fields, "httpd_request_methods_post", httpdRequestMethodsPost)
	c.generateFields(fields, "httpd_request_methods_head", httpdRequestMethodsHead)

	// status code stats:
	c.generateFields(fields, "httpd_status_codes_200", httpdStatusCodesStatus200)
	c.generateFields(fields, "httpd_status_codes_201", httpdStatusCodesStatus201)
	c.generateFields(fields, "httpd_status_codes_202", httpdStatusCodesStatus202)
	c.generateFields(fields, "httpd_status_codes_301", httpdStatusCodesStatus301)
	c.generateFields(fields, "httpd_status_codes_304", httpdStatusCodesStatus304)
	c.generateFields(fields, "httpd_status_codes_400", httpdStatusCodesStatus400)
	c.generateFields(fields, "httpd_status_codes_401", httpdStatusCodesStatus401)
	c.generateFields(fields, "httpd_status_codes_403", httpdStatusCodesStatus403)
	c.generateFields(fields, "httpd_status_codes_404", httpdStatusCodesStatus404)
	c.generateFields(fields, "httpd_status_codes_405", httpdStatusCodesStatus405)
	c.generateFields(fields, "httpd_status_codes_409", httpdStatusCodesStatus409)
	c.generateFields(fields, "httpd_status_codes_412", httpdStatusCodesStatus412)
	c.generateFields(fields, "httpd_status_codes_500", httpdStatusCodesStatus500)

	// httpd stats:
	c.generateFields(fields, "httpd_clients_requesting_changes", stats.Httpd.ClientsRequestingChanges)
	c.generateFields(fields, "httpd_temporary_view_reads", stats.Httpd.TemporaryViewReads)
	c.generateFields(fields, "httpd_requests", stats.Httpd.Requests)
	c.generateFields(fields, "httpd_bulk_requests", stats.Httpd.BulkRequests)
	c.generateFields(fields, "httpd_view_reads", stats.Httpd.ViewReads)

	tags := map[string]string{
		"server": host,
	}
	accumulator.AddFields("couchdb", fields, tags)
	return nil
}

func (c *CouchDB) generateFields(fields map[string]interface{}, prefix string, obj metaData) {
	if obj.Value != nil {
		fields[prefix+"_value"] = *obj.Value
	}
	if obj.Current != nil {
		fields[prefix+"_current"] = *obj.Current
	}
	if obj.Sum != nil {
		fields[prefix+"_sum"] = *obj.Sum
	}
	if obj.Mean != nil {
		fields[prefix+"_mean"] = *obj.Mean
	}
	if obj.Stddev != nil {
		fields[prefix+"_stddev"] = *obj.Stddev
	}
	if obj.Min != nil {
		fields[prefix+"_min"] = *obj.Min
	}
	if obj.Max != nil {
		fields[prefix+"_max"] = *obj.Max
	}
}

func init() {
	inputs.Add("couchdb", func() telegraf.Input {
		return &CouchDB{
			client: &http.Client{
				Transport: &http.Transport{
					ResponseHeaderTimeout: time.Duration(3 * time.Second),
				},
				Timeout: time.Duration(4 * time.Second),
			},
		}
	})
}
