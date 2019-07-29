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
		"proxy":  "git",
		"sv":     "www",
		"type":   "server",
	}

	fields := HaproxyGetFieldValues()
	acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)

	//Here, we should get error because we don't pass authentication data
	r = &haproxy{
		Servers: []string{ts.URL},
	}

	r.Gather(&acc)
	require.NotEmpty(t, acc.Errors)
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
		"server": ts.Listener.Addr().String(),
		"proxy":  "git",
		"sv":     "www",
		"type":   "server",
	}

	fields := HaproxyGetFieldValues()
	acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)
}

func TestHaproxyGeneratesMetricsUsingSocket(t *testing.T) {
	var randomNumber int64
	var sockets [5]net.Listener
	_globmask := "/tmp/test-haproxy*.sock"
	_badmask := "/tmp/test-fail-haproxy*.sock"

	for i := 0; i < 5; i++ {
		binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
		sockname := fmt.Sprintf("/tmp/test-haproxy%d.sock", randomNumber)

		sock, err := net.Listen("unix", sockname)
		if err != nil {
			t.Fatal("Cannot initialize socket ")
		}

		sockets[i] = sock
		defer sock.Close()

		s := statServer{}
		go s.serverSocket(sock)
	}

	r := &haproxy{
		Servers: []string{_globmask},
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	fields := HaproxyGetFieldValues()

	for _, sock := range sockets {
		tags := map[string]string{
			"server": sock.Addr().String(),
			"proxy":  "git",
			"sv":     "www",
			"type":   "server",
		}

		acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)
	}

	// This mask should not match any socket
	r.Servers = []string{_badmask}

	r.Gather(&acc)
	require.NotEmpty(t, acc.Errors)
}

//When not passing server config, we default to localhost
//We just want to make sure we did request stat from localhost
func TestHaproxyDefaultGetFromLocalhost(t *testing.T) {
	r := &haproxy{}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "127.0.0.1:1936/haproxy?stats/;csv")
}

func TestHaproxyKeepFieldNames(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, csvOutputSample)
	}))
	defer ts.Close()

	r := &haproxy{
		Servers:        []string{ts.URL},
		KeepFieldNames: true,
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"server": ts.Listener.Addr().String(),
		"pxname": "git",
		"svname": "www",
		"type":   "server",
	}

	fields := HaproxyGetFieldValues()
	fields["act"] = fields["active_servers"]
	delete(fields, "active_servers")
	fields["bck"] = fields["backup_servers"]
	delete(fields, "backup_servers")
	fields["cli_abrt"] = fields["cli_abort"]
	delete(fields, "cli_abort")
	fields["srv_abrt"] = fields["srv_abort"]
	delete(fields, "srv_abort")
	fields["hrsp_1xx"] = fields["http_response.1xx"]
	delete(fields, "http_response.1xx")
	fields["hrsp_2xx"] = fields["http_response.2xx"]
	delete(fields, "http_response.2xx")
	fields["hrsp_3xx"] = fields["http_response.3xx"]
	delete(fields, "http_response.3xx")
	fields["hrsp_4xx"] = fields["http_response.4xx"]
	delete(fields, "http_response.4xx")
	fields["hrsp_5xx"] = fields["http_response.5xx"]
	delete(fields, "http_response.5xx")
	fields["hrsp_other"] = fields["http_response.other"]
	delete(fields, "http_response.other")

	acc.AssertContainsTaggedFields(t, "haproxy", fields, tags)
}

func HaproxyGetFieldValues() map[string]interface{} {
	fields := map[string]interface{}{
		"active_servers":      uint64(1),
		"backup_servers":      uint64(0),
		"bin":                 uint64(5228218),
		"bout":                uint64(303747244),
		"check_code":          uint64(200),
		"check_duration":      uint64(3),
		"check_fall":          uint64(3),
		"check_health":        uint64(4),
		"check_rise":          uint64(2),
		"check_status":        "L7OK",
		"chkdown":             uint64(84),
		"chkfail":             uint64(559),
		"cli_abort":           uint64(690),
		"ctime":               uint64(1),
		"downtime":            uint64(3352),
		"dresp":               uint64(0),
		"econ":                uint64(0),
		"eresp":               uint64(21),
		"http_response.1xx":   uint64(0),
		"http_response.2xx":   uint64(5668),
		"http_response.3xx":   uint64(8710),
		"http_response.4xx":   uint64(140),
		"http_response.5xx":   uint64(0),
		"http_response.other": uint64(0),
		"iid":                 uint64(4),
		"last_chk":            "OK",
		"lastchg":             uint64(1036557),
		"lastsess":            int64(1342),
		"lbtot":               uint64(9481),
		"mode":                "http",
		"pid":                 uint64(1),
		"qcur":                uint64(0),
		"qmax":                uint64(0),
		"qtime":               uint64(1268),
		"rate":                uint64(0),
		"rate_max":            uint64(2),
		"rtime":               uint64(2908),
		"sid":                 uint64(1),
		"scur":                uint64(0),
		"slim":                uint64(2),
		"smax":                uint64(2),
		"srv_abort":           uint64(0),
		"status":              "UP",
		"stot":                uint64(14539),
		"ttime":               uint64(4500),
		"weight":              uint64(1),
		"wredis":              uint64(0),
		"wretr":               uint64(0),
	}
	return fields
}

// Can obtain from official haproxy demo: 'http://demo.haproxy.org/;csv'
const csvOutputSample = `
# pxname,svname,qcur,qmax,scur,smax,slim,stot,bin,bout,dreq,dresp,ereq,econ,eresp,wretr,wredis,status,weight,act,bck,chkfail,chkdown,lastchg,downtime,qlimit,pid,iid,sid,throttle,lbtot,tracked,type,rate,rate_lim,rate_max,check_status,check_code,check_duration,hrsp_1xx,hrsp_2xx,hrsp_3xx,hrsp_4xx,hrsp_5xx,hrsp_other,hanafail,req_rate,req_rate_max,req_tot,cli_abrt,srv_abrt,comp_in,comp_out,comp_byp,comp_rsp,lastsess,last_chk,last_agt,qtime,ctime,rtime,ttime,agent_status,agent_code,agent_duration,check_desc,agent_desc,check_rise,check_fall,check_health,agent_rise,agent_fall,agent_health,addr,cookie,mode,algo,conn_rate,conn_rate_max,conn_tot,intercepted,dcon,dses,
http-in,FRONTEND,,,3,100,100,2639994,813557487,65937668635,505252,0,47567,,,,,OPEN,,,,,,,,,1,2,0,,,,0,1,0,157,,,,0,1514640,606647,136264,496535,14948,,1,155,2754255,,,36370569635,17435137766,0,642264,,,,,,,,,,,,,,,,,,,,,http,,1,157,2649922,339471,0,0,
http-in,IPv4-direct,,,3,41,100,349801,57445827,1503928881,269899,0,287,,,,,OPEN,,,,,,,,,1,2,1,,,,3,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,http,,,,,,0,0,
http-in,IPv4-cached,,,0,33,100,1786155,644395819,57905460294,60511,0,1,,,,,OPEN,,,,,,,,,1,2,2,,,,3,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,http,,,,,,0,0,
http-in,IPv6-direct,,,0,100,100,325619,92414745,6205208728,3399,0,47279,,,,,OPEN,,,,,,,,,1,2,3,,,,3,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,http,,,,,,0,0,
http-in,local,,,0,0,100,0,0,0,0,0,0,,,,,OPEN,,,,,,,,,1,2,4,,,,3,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,http,,,,,,0,0,
http-in,local-https,,,0,5,100,188347,19301096,323070732,171443,0,0,,,,,OPEN,,,,,,,,,1,2,5,,,,3,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,http,,,,,,0,0,
www,www,0,0,0,20,20,1719698,672044109,64806076656,,0,,0,5285,22,0,UP,1,1,0,561,84,1036557,3356,,1,3,1,,1715117,,2,0,,45,L7OK,200,5,671,1144889,481714,87038,4,0,,,,,105016,167,,,,,5,OK,,0,5,16,1167,,,,Layer7 check passed,,2,3,4,,,,,,http,,,,,,,,
www,bck,0,0,0,10,10,1483,537137,7544118,,0,,0,0,0,0,UP,1,0,1,4,0,5218087,0,,1,3,2,,1371,,2,0,,17,L7OK,200,2,0,629,99,755,0,0,,,,,16,0,,,,,1036557,OK,,756,1,13,1184,,,,Layer7 check passed,,2,5,6,,,,,,http,,,,,,,,
www,BACKEND,0,25,0,46,100,1721835,674684790,64813732170,314,0,,130,5285,22,0,UP,1,1,1,,0,5218087,0,,1,3,0,,1716488,,1,0,,45,,,,0,1145518,481813,88664,5719,121,,,,1721835,105172,167,35669268059,17250148556,0,556042,5,,,0,5,16,1167,,,,,,,,,,,,,,http,,,,,,,,
git,www,0,0,0,2,2,14539,5228218,303747244,,0,,0,21,0,0,UP,1,1,0,559,84,1036557,3352,,1,4,1,,9481,,2,0,,2,L7OK,200,3,0,5668,8710,140,0,0,,,,,690,0,,,,,1342,OK,,1268,1,2908,4500,,,,Layer7 check passed,,2,3,4,,,,,,http,,,,,,,,
git,bck,0,0,0,0,2,0,0,0,,0,,0,0,0,0,UP,1,0,1,2,0,5218087,0,,1,4,2,,0,,2,0,,0,L7OK,200,2,0,0,0,0,0,0,,,,,0,0,,,,,-1,OK,,0,0,0,0,,,,Layer7 check passed,,2,3,4,,,,,,http,,,,,,,,
git,BACKEND,0,6,0,8,2,14541,8082393,303747668,0,0,,2,21,0,0,UP,1,1,1,,0,5218087,0,,1,4,0,,9481,,1,0,,7,,,,0,5668,8710,140,23,0,,,,14541,690,0,133458298,38104818,0,4379,1342,,,1268,1,2908,4500,,,,,,,,,,,,,,http,,,,,,,,
demo,BACKEND,0,0,1,5,20,24063,7876647,659864417,48,0,,1,0,0,0,UP,0,0,0,,0,5218087,,,1,17,0,,0,,1,1,,26,,,,0,23983,21,0,1,57,,,,24062,111,0,567843278,146884392,0,1083,0,,,2706,0,0,887,,,,,,,,,,,,,,http,,,,,,,,
`
