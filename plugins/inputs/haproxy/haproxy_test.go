package haproxy

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type statServer struct{}

func (s statServer) serverSocket(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		go func(c net.Conn) {
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)

			data := buf[:n]
			if string(data) == "show stat\n" {
				c.Write([]byte(csvOutputSample))
				c.Close()
			}
		}(conn)
	}
}

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
		"server": ts.Listener.Addr().String(),
		"proxy":  "be_app",
		"sv":     "host0",
	}

	fields := map[string]interface{}{
		"active_servers":    uint64(1),
		"backup_servers":    uint64(0),
		"bin":               uint64(510913516),
		"bout":              uint64(2193856571),
		"check_duration":    uint64(10),
		"cli_abort":         uint64(73),
		"ctime":             uint64(2),
		"downtime":          uint64(0),
		"dresp":             uint64(0),
		"econ":              uint64(0),
		"eresp":             uint64(1),
		"http_response.1xx": uint64(0),
		"http_response.2xx": uint64(119534),
		"http_response.3xx": uint64(48051),
		"http_response.4xx": uint64(2345),
		"http_response.5xx": uint64(1056),
		"lbtot":             uint64(171013),
		"qcur":              uint64(0),
		"qmax":              uint64(0),
		"qtime":             uint64(0),
		"rate":              uint64(3),
		"rate_max":          uint64(12),
		"rtime":             uint64(312),
		"scur":              uint64(1),
		"smax":              uint64(32),
		"srv_abort":         uint64(1),
		"stot":              uint64(171014),
		"ttime":             uint64(2341),
		"wredis":            uint64(0),
		"wretr":             uint64(1),
	}
	acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)

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
		"proxy":  "be_app",
		"server": ts.Listener.Addr().String(),
		"sv":     "host0",
	}

	fields := map[string]interface{}{
		"active_servers":    uint64(1),
		"backup_servers":    uint64(0),
		"bin":               uint64(510913516),
		"bout":              uint64(2193856571),
		"check_duration":    uint64(10),
		"cli_abort":         uint64(73),
		"ctime":             uint64(2),
		"downtime":          uint64(0),
		"dresp":             uint64(0),
		"econ":              uint64(0),
		"eresp":             uint64(1),
		"http_response.1xx": uint64(0),
		"http_response.2xx": uint64(119534),
		"http_response.3xx": uint64(48051),
		"http_response.4xx": uint64(2345),
		"http_response.5xx": uint64(1056),
		"lbtot":             uint64(171013),
		"qcur":              uint64(0),
		"qmax":              uint64(0),
		"qtime":             uint64(0),
		"rate":              uint64(3),
		"rate_max":          uint64(12),
		"rtime":             uint64(312),
		"scur":              uint64(1),
		"smax":              uint64(32),
		"srv_abort":         uint64(1),
		"stot":              uint64(171014),
		"ttime":             uint64(2341),
		"wredis":            uint64(0),
		"wretr":             uint64(1),
	}
	acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)
}

func TestHaproxyGeneratesMetricsUsingSocket(t *testing.T) {
	var randomNumber int64
	binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
	sock, err := net.Listen("unix", fmt.Sprintf("/tmp/test-haproxy%d.sock", randomNumber))
	if err != nil {
		t.Fatal("Cannot initialize socket ")
	}

	defer sock.Close()

	s := statServer{}
	go s.serverSocket(sock)

	r := &haproxy{
		Servers: []string{sock.Addr().String()},
	}

	var acc testutil.Accumulator

	err = r.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"proxy":  "be_app",
		"server": sock.Addr().String(),
		"sv":     "host0",
	}

	fields := map[string]interface{}{
		"active_servers":    uint64(1),
		"backup_servers":    uint64(0),
		"bin":               uint64(510913516),
		"bout":              uint64(2193856571),
		"check_duration":    uint64(10),
		"cli_abort":         uint64(73),
		"ctime":             uint64(2),
		"downtime":          uint64(0),
		"dresp":             uint64(0),
		"econ":              uint64(0),
		"eresp":             uint64(1),
		"http_response.1xx": uint64(0),
		"http_response.2xx": uint64(119534),
		"http_response.3xx": uint64(48051),
		"http_response.4xx": uint64(2345),
		"http_response.5xx": uint64(1056),
		"lbtot":             uint64(171013),
		"qcur":              uint64(0),
		"qmax":              uint64(0),
		"qtime":             uint64(0),
		"rate":              uint64(3),
		"rate_max":          uint64(12),
		"rtime":             uint64(312),
		"scur":              uint64(1),
		"smax":              uint64(32),
		"srv_abort":         uint64(1),
		"stot":              uint64(171014),
		"ttime":             uint64(2341),
		"wredis":            uint64(0),
		"wretr":             uint64(1),
	}
	acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)
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
