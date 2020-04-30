package becs

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var mockServer *httptest.Server

func TestBecs(t *testing.T) {
	//First request will be from sessionLogin().
	mockServer = httptest.NewServer(http.HandlerFunc(handleSessionLogin))
	defer mockServer.Close()

	b := Becs{}
	b.url.Scheme = "http"
	url := mockServer.Listener.Addr().String()
	b.url.Host = url
	b.IncludePools = true
	b.Resources = []string{"10.0.0.0/8"}

	var acc testutil.Accumulator
	b.sessionLogin()
	b.applicationStatusGet(&acc)

	//Test applicationStatusGet().
	appStatustags := map[string]string{
		"application": "test_app",
		"server":      "labbbecs",
	}

	appStatusfields := make(map[string]interface{})
	appStatusfields["uptime"] = uint(100000)
	appStatusfields["cpuusage"] = uint(1)
	appStatusfields["cpuaverage60"] = uint(2)
	acc.AssertContainsTaggedFields(t, "becs_applications", appStatusfields, appStatustags)

	//Test applicationStatusGet() with include_pools.
	appStatustags["memorypool"] = "test_pool"

	poolFields := make(map[string]interface{})
	poolFields["size"] = uint(1024)
	poolFields["out"] = uint(10)
	poolFields["pages"] = uint(20)
	poolFields["emptypages"] = uint(0)
	acc.AssertContainsTaggedFields(t, "becs_applications", poolFields, appStatustags)

	acc.ClearMetrics()
	b.metricGet(&acc)

	//Test metricGet().
	metricTags := map[string]string{
		"server": "127.0.0.1",
		"metric": "em_elements",
		"emtype": "ibos",
	}

	metricFields := make(map[string]interface{})
	metricFields["elements"] = uint64(100)
	acc.AssertContainsTaggedFields(t, "becs_metrics", metricFields, metricTags)

	acc.ClearMetrics()
	b.clientFind(&acc)

	//Test clientFind().
	clientFindTags := map[string]string{
		"server":   "127.0.0.1",
		"resource": "10.0.0.0/8",
	}

	clientFindFields := make(map[string]interface{})
	clientFindFields["clients"] = uint(3000)
	acc.AssertContainsTaggedFields(t, "becs_clients", clientFindFields, clientFindTags)
}

//Response for sessionLogin().
func handleSessionLogin(w http.ResponseWriter, r *http.Request) {
	s := sessionLoginResponse{}
	s.Body.Response.Out.SessionID = "123456789"

	resp, _ := xml.Marshal(s)

	w.Write(resp)

	//Set next response.
	mockServer.Config.Handler = http.HandlerFunc(handleApplicationList)
}

//Response for applicationList().
func handleApplicationList(w http.ResponseWriter, r *http.Request) {
	a := applicationListResponse{}
	a.Body.Response.Out.Names.Items = []string{"test_app"}

	resp, _ := xml.Marshal(a)

	w.Write(resp)

	//Set next response.
	mockServer.Config.Handler = http.HandlerFunc(handleApplicationStatusGet)
}

//Response for applicationStatusGet().
func handleApplicationStatusGet(w http.ResponseWriter, r *http.Request) {
	a := applicationStatusGetResponse{}
	a.Body.Response.Out.Displayname = "test_app"
	a.Body.Response.Out.Hostname = "labbbecs"
	a.Body.Response.Out.UpTime = uint(100000)
	a.Body.Response.Out.CPUUsage = uint(1)
	a.Body.Response.Out.CPUAverage60 = uint(2)
	pool := memoryPool{
		Name:       "test_pool",
		Size:       uint(1024),
		Out:        uint(10),
		Pages:      uint(20),
		EmptyPages: uint(0),
	}
	a.Body.Response.Out.MemoryPools.Items = append(a.Body.Response.Out.MemoryPools.Items, pool)

	resp, _ := xml.Marshal(a)

	w.Write(resp)

	//Set next response.
	mockServer.Config.Handler = http.HandlerFunc(handleMetricGet)
}

//Response for metricGet().
func handleMetricGet(w http.ResponseWriter, r *http.Request) {
	m := metricGetResponse{}
	metricLabel := metricLabel{
		Name:  "emtype",
		Value: "ibos",
	}

	metricValue := metricValue{
		Value: uint64(100),
	}
	metricValue.Labels.Items = append(metricValue.Labels.Items, metricLabel)

	metric := metric{
		Name: "em_elements",
	}
	metric.Values.Items = append(metric.Values.Items, metricValue)

	m.Body.Response.Out.Metrics.Items = append(m.Body.Response.Out.Metrics.Items, metric)

	resp, _ := xml.Marshal(m)

	w.Write(resp)

	//Set next response.
	mockServer.Config.Handler = http.HandlerFunc(handleClientFind)
}

//Response for clientFind().
func handleClientFind(w http.ResponseWriter, r *http.Request) {
	m := clientFindResponse{}
	m.Body.Response.Out.Actual = uint(3000)

	resp, _ := xml.Marshal(m)

	w.Write(resp)
}
