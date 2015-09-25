package haproxy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
)

func TestHaproxyGeneratesMetricsWithAuthentication(t *testing.T) {
	//We create a fake server to return test data
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Unauthorized")
			return
		}

		if username == "user" && password == "password" {
			fmt.Fprint(w, csvOutputSample)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Unauthorized")
		}
	}))
	defer ts.Close()

	//Now we tested again above server, with our authentication data
	r := &haproxy{
		Servers: []string{strings.Replace(ts.URL, "http://", "http://user:password@", 1)},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"host":  ts.Listener.Addr().String(),
		"proxy": "be_app",
		"sv":    "host0",
	}

	assert.NoError(t, acc.ValidateTaggedValue("stot", uint64(171014), tags))

	checkInt := []struct {
		name  string
		value uint64
	}{
		{"bin", 5557055817},
		{"scur", 288},
		{"qmax", 81},
		{"http_response.1xx", 0},
		{"http_response.2xx", 1314093},
		{"http_response.3xx", 537036},
		{"http_response.4xx", 123452},
		{"http_response.5xx", 11966},
		{"dreq", 1102},
		{"dresp", 80},
		{"wretr", 17},
		{"wredis", 19},
		{"ereq", 95740},
		{"econ", 0},
		{"eresp", 0},
		{"req_rate", 35},
		{"req_rate_max", 140},
		{"req_tot", 1987928},
		{"bin", 5557055817},
		{"bout", 24096715169},
		{"rate", 18},
		{"rate_max", 102},

		{"throttle", 13},
		{"lbtot", 114},
	}

	for _, c := range checkInt {
		assert.Equal(t, true, acc.CheckValue(c.name, c.value))
	}

	//Here, we should get error because we don't pass authentication data
	r = &haproxy{
		Servers: []string{ts.URL},
	}

	err = r.Gather(&acc)
	require.Error(t, err)
}

func TestHaproxyGeneratesMetricsWithoutAuthentication(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, csvOutputSample)
	}))
	defer ts.Close()

	r := &haproxy{
		Servers: []string{ts.URL},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"proxy": "be_app",
		"host":  ts.Listener.Addr().String(),
		"sv":    "host0",
	}

	assert.NoError(t, acc.ValidateTaggedValue("stot", uint64(171014), tags))
	assert.NoError(t, acc.ValidateTaggedValue("scur", uint64(1), tags))
	assert.NoError(t, acc.ValidateTaggedValue("rate", uint64(3), tags))
	assert.Equal(t, true, acc.CheckValue("bin", uint64(5557055817)))
}

//When not passing server config, we default to localhost
//We just want to make sure we did request stat from localhost
func TestHaproxyDefaultGetFromLocalhost(t *testing.T) {
	r := &haproxy{}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "127.0.0.1:1936/;csv")
}

const csvOutputSample = `
# pxname,svname,qcur,qmax,scur,smax,slim,stot,bin,bout,dreq,dresp,ereq,econ,eresp,wretr,wredis,status,weight,act,bck,chkfail,chkdown,lastchg,downtime,qlimit,pid,iid,sid,throttle,lbtot,tracked,type,rate,rate_lim,rate_max,check_status,check_code,check_duration,hrsp_1xx,hrsp_2xx,hrsp_3xx,hrsp_4xx,hrsp_5xx,hrsp_other,hanafail,req_rate,req_rate_max,req_tot,cli_abrt,srv_abrt,comp_in,comp_out,comp_byp,comp_rsp,lastsess,last_chk,last_agt,qtime,ctime,rtime,ttime,
fe_app,FRONTEND,,81,288,713,2000,1094063,5557055817,24096715169,1102,80,95740,,,17,19,OPEN,,,,,,,,,2,16,113,13,114,,0,18,0,102,,,,0,1314093,537036,123452,11966,1360,,35,140,1987928,,,0,0,0,0,,,,,,,,
be_static,host0,0,0,0,3,,3209,1141294,17389596,,0,,0,0,0,0,no check,1,1,0,,,,,,2,17,1,,3209,,2,0,,7,,,,0,218,1497,1494,0,0,0,,,,0,0,,,,,2,,,0,2,23,545,
be_static,BACKEND,0,0,0,3,200,3209,1141294,17389596,0,0,,0,0,0,0,UP,1,1,0,,0,70698,0,,2,17,0,,3209,,1,0,,7,,,,0,218,1497,1494,0,0,,,,,0,0,0,0,0,0,2,,,0,2,23,545,
be_static,host0,0,0,0,1,,28,17313,466003,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,1,,28,,2,0,,1,L4OK,,1,0,17,6,5,0,0,0,,,,0,0,,,,,2103,,,0,1,1,36,
be_static,host4,0,0,0,1,,28,15358,1281073,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,2,,28,,2,0,,1,L4OK,,1,0,20,5,3,0,0,0,,,,0,0,,,,,2076,,,0,1,1,54,
be_static,host5,0,0,0,1,,28,17547,1970404,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,3,,28,,2,0,,1,L4OK,,0,0,20,5,3,0,0,0,,,,0,0,,,,,1495,,,0,1,1,53,
be_static,host6,0,0,0,1,,28,14105,1328679,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,4,,28,,2,0,,1,L4OK,,0,0,18,8,2,0,0,0,,,,0,0,,,,,1418,,,0,0,1,49,
be_static,host7,0,0,0,1,,28,15258,1965185,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,5,,28,,2,0,,1,L4OK,,0,0,17,8,3,0,0,0,,,,0,0,,,,,935,,,0,0,1,28,
be_static,host8,0,0,0,1,,28,12934,1034779,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,6,,28,,2,0,,1,L4OK,,0,0,17,9,2,0,0,0,,,,0,0,,,,,582,,,0,1,1,66,
be_static,host9,0,0,0,1,,28,13434,134063,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,7,,28,,2,0,,1,L4OK,,0,0,17,8,3,0,0,0,,,,0,0,,,,,539,,,0,0,1,80,
be_static,host1,0,0,0,1,,28,7873,1209688,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,8,,28,,2,0,,1,L4OK,,0,0,22,6,0,0,0,0,,,,0,0,,,,,487,,,0,0,1,36,
be_static,host2,0,0,0,1,,28,13830,1085929,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,9,,28,,2,0,,1,L4OK,,0,0,19,6,3,0,0,0,,,,0,0,,,,,338,,,0,1,1,38,
be_static,host3,0,0,0,1,,28,17959,1259760,,0,,0,0,0,0,UP,1,1,0,0,0,70698,0,,2,18,10,,28,,2,0,,1,L4OK,,1,0,20,6,2,0,0,0,,,,0,0,,,,,92,,,0,1,1,17,
be_static,BACKEND,0,0,0,2,200,307,160276,13322728,0,0,,0,0,0,0,UP,11,11,0,,0,70698,0,,2,18,0,,307,,1,0,,4,,,,0,205,73,29,0,0,,,,,0,0,0,0,0,0,92,,,0,1,3,381,
be_app,host0,0,0,1,32,,171014,510913516,2193856571,,0,,0,1,1,0,UP,100,1,0,1,0,70698,0,,2,19,1,,171013,,2,3,,12,L7OK,301,10,0,119534,48051,2345,1056,0,0,,,,73,1,,,,,0,Moved Permanently,,0,2,312,2341,
be_app,host4,0,0,2,29,,171013,499318742,2195595896,12,34,,0,2,0,0,UP,100,1,0,2,0,70698,0,,2,19,2,,171013,,2,3,,12,L7OK,301,12,0,119572,47882,2441,1088,0,0,,,,84,2,,,,,0,Moved Permanently,,0,2,316,2355,
`
