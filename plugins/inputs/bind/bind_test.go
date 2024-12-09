package bind

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestBindJsonStats(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	url := ts.Listener.Addr().String()
	host, port, err := net.SplitHostPort(url)
	require.NoError(t, err)
	defer ts.Close()

	b := Bind{
		Urls:                 []string{ts.URL + "/json/v1"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		CountersAsInt:        true,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
	}

	var acc testutil.Accumulator
	err = acc.GatherError(b.Gather)

	require.NoError(t, err)

	// Use subtests for counters, since they are similar structure
	type fieldSet struct {
		fieldKey   string
		fieldValue int64
	}

	testCases := []struct {
		counterType string
		values      []fieldSet
	}{
		{
			"opcode",
			[]fieldSet{
				{"NOTIFY", 0},
				{"UPDATE", 0},
				{"IQUERY", 0},
				{"QUERY", 13},
				{"STATUS", 0},
			},
		},
		{
			"rcode",
			[]fieldSet{
				{"NOERROR", 1732},
				{"FORMERR", 0},
				{"SERVFAIL", 6},
				{"NXDOMAIN", 200},
				{"NOTIMP", 0},
				{"REFUSED", 6},
				{"REFUSED", 0},
				{"YXDOMAIN", 0},
				{"YXRRSET", 0},
				{"NXRRSET", 0},
				{"NOTAUTH", 0},
				{"NOTZONE", 0},
				{"RESERVED11", 0},
				{"RESERVED12", 0},
				{"RESERVED13", 0},
				{"RESERVED14", 0},
				{"RESERVED15", 0},
				{"BADVERS", 0},
				{"17", 0},
				{"18", 0},
				{"19", 0},
				{"20", 0},
				{"21", 0},
				{"22", 0},
				{"BADCOOKIE", 0},
			},
		},
		{
			"qtype",
			[]fieldSet{
				{"A", 2},
				{"AAAA", 2},
				{"PTR", 7},
				{"SRV", 2},
			},
		},
		{
			"nsstat",
			[]fieldSet{
				{"QrySuccess", 6},
				{"QryRecursion", 12},
				{"Requestv4", 13},
				{"QryNXDOMAIN", 4},
				{"QryAuthAns", 1},
				{"QryNxrrset", 1},
				{"QryNoauthAns", 10},
				{"QryUDP", 13},
				{"QryDuplicate", 1},
				{"QrySERVFAIL", 1},
				{"Response", 12},
			},
		},
		{
			"sockstat",
			[]fieldSet{
				{"TCP4Open", 118},
				{"UDP6Close", 112},
				{"UDP4Close", 333},
				{"TCP4Close", 119},
				{"TCP6Active", 2},
				{"UDP4Active", 2},
				{"UDP4RecvErr", 1},
				{"UDP4Open", 335},
				{"TCP4Active", 10},
				{"RawActive", 1},
				{"UDP6ConnFail", 112},
				{"TCP4Conn", 114},
				{"UDP6Active", 1},
				{"UDP6Open", 113},
				{"UDP4Conn", 333},
				{"UDP6SendErr", 112},
				{"RawOpen", 1},
				{"TCP4Accept", 6},
				{"TCP6Open", 2},
			},
		},
		{
			"zonestat",
			[]fieldSet{
				{"NotifyOutv4", 8},
				{"NotifyInv4", 5},
				{"SOAOutv4", 5},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.counterType, func(t *testing.T) {
			tags := map[string]string{
				"url":    url,
				"type":   tc.counterType,
				"source": host,
				"port":   port,
			}

			fields := map[string]interface{}{}

			for _, val := range tc.values {
				fields[val.fieldKey] = val.fieldValue
			}

			acc.AssertContainsTaggedFields(t, "bind_counter", fields, tags)
		})
	}

	// Subtest for memory stats
	t.Run("memory", func(t *testing.T) {
		tags := map[string]string{
			"url":    url,
			"source": host,
			"port":   port,
		}

		fields := map[string]interface{}{
			"block_size":   int64(13893632),
			"context_size": int64(3685480),
			"in_use":       int64(3064368),
			"lost":         int64(0),
			"total_use":    int64(18206566),
		}
		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})

	// Subtest for per-context memory stats
	t.Run("memory_context", func(t *testing.T) {
		require.True(t, acc.HasInt64Field("bind_memory_context", "total"))
		require.True(t, acc.HasInt64Field("bind_memory_context", "in_use"))
	})
}

func TestBindXmlStatsV2(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	url := ts.Listener.Addr().String()
	host, port, err := net.SplitHostPort(url)
	require.NoError(t, err)
	defer ts.Close()

	b := Bind{
		Urls:                 []string{ts.URL + "/xml/v2"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		CountersAsInt:        true,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
	}

	var acc testutil.Accumulator
	err = acc.GatherError(b.Gather)

	require.NoError(t, err)

	// Use subtests for counters, since they are similar structure
	type fieldSet struct {
		fieldKey   string
		fieldValue int64
	}

	testCases := []struct {
		counterType string
		values      []fieldSet
	}{
		{
			"opcode",
			[]fieldSet{
				{"UPDATE", 238},
				{"QUERY", 102312374},
			},
		},
		{
			"qtype",
			[]fieldSet{
				{"ANY", 7},
				{"DNSKEY", 452},
				{"SSHFP", 2987},
				{"SOA", 100415},
				{"AAAA", 37786321},
				{"MX", 441155},
				{"IXFR", 157},
				{"CNAME", 531},
				{"NS", 1999},
				{"TXT", 34628},
				{"A", 58951432},
				{"SRV", 741082},
				{"PTR", 4211487},
				{"NAPTR", 39137},
				{"DS", 584},
			},
		},
		{
			"nsstat",
			[]fieldSet{
				{"XfrReqDone", 157},
				{"ReqEdns0", 441758},
				{"ReqTSIG", 0},
				{"UpdateRespFwd", 0},
				{"RespEDNS0", 441748},
				{"QryDropped", 16},
				{"RPZRewrites", 0},
				{"XfrRej", 0},
				{"RecQryRej", 0},
				{"QryNxrrset", 24423133},
				{"QryFORMERR", 0},
				{"ReqTCP", 1548156},
				{"UpdateDone", 0},
				{"QrySERVFAIL", 14422},
				{"QryRecursion", 2104239},
				{"Requestv4", 102312611},
				{"UpdateFwdFail", 0},
				{"QryReferral", 3},
				{"Response", 102301560},
				{"RespTSIG", 0},
				{"QrySuccess", 63811668},
				{"QryFailure", 0},
				{"RespSIG0", 0},
				{"ReqSIG0", 0},
				{"UpdateRej", 238},
				{"QryAuthAns", 72180718},
				{"UpdateFail", 0},
				{"QryDuplicate", 10879},
				{"RateDropped", 0},
				{"QryNoauthAns", 30106182},
				{"QryNXDOMAIN", 14052096},
				{"ReqBadSIG", 0},
				{"UpdateReqFwd", 0},
				{"RateSlipped", 0},
				{"TruncatedResp", 3787},
				{"Requestv6", 1},
				{"UpdateBadPrereq", 0},
				{"AuthQryRej", 0},
				{"ReqBadEDNSVer", 0},
			},
		},
		{
			"sockstat",
			[]fieldSet{
				{"FdwatchBindFail", 0},
				{"UDP6Open", 238269},
				{"UDP6SendErr", 238250},
				{"TCP4ConnFail", 0},
				{"TCP4Conn", 590},
				{"TCP6AcceptFail", 0},
				{"UDP4SendErr", 0},
				{"FDwatchConn", 0},
				{"TCP4RecvErr", 1},
				{"TCP4OpenFail", 0},
				{"UDP4OpenFail", 0},
				{"UDP6OpenFail", 0},
				{"TCP4Close", 1548268},
				{"TCP6BindFail", 0},
				{"TCP4AcceptFail", 0},
				{"UnixConn", 0},
				{"UDP4Open", 3765532},
				{"TCP6Close", 0},
				{"FDwatchRecvErr", 0},
				{"UDP4Conn", 3764828},
				{"UnixConnFail", 0},
				{"TCP6Conn", 0},
				{"TCP6OpenFail", 0},
				{"TCP6SendErr", 0},
				{"TCP6RecvErr", 0},
				{"FDwatchSendErr", 0},
				{"UDP4RecvErr", 1650},
				{"UDP4ConnFail", 0},
				{"UDP6Close", 238267},
				{"FDWatchClose", 0},
				{"TCP4Accept", 1547672},
				{"UnixAccept", 0},
				{"TCP4Open", 602},
				{"UDP4BindFail", 219},
				{"UDP6ConnFail", 238250},
				{"UnixClose", 0},
				{"TCP4BindFail", 0},
				{"UnixOpenFail", 0},
				{"UDP6BindFail", 16},
				{"UnixOpen", 0},
				{"UnixAcceptFail", 0},
				{"UnixRecvErr", 0},
				{"UDP6RecvErr", 0},
				{"TCP6ConnFail", 0},
				{"FDwatchConnFail", 0},
				{"TCP4SendErr", 0},
				{"UDP4Close", 3765528},
				{"UnixSendErr", 0},
				{"TCP6Open", 2},
				{"UDP6Conn", 1},
				{"TCP6Accept", 0},
				{"UnixBindFail", 0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.counterType, func(t *testing.T) {
			tags := map[string]string{
				"url":    url,
				"type":   tc.counterType,
				"source": host,
				"port":   port,
			}

			fields := make(map[string]interface{}, len(tc.values))
			for _, val := range tc.values {
				fields[val.fieldKey] = val.fieldValue
			}

			acc.AssertContainsTaggedFields(t, "bind_counter", fields, tags)
		})
	}

	// Subtest for memory stats
	t.Run("memory", func(t *testing.T) {
		tags := map[string]string{
			"url":    url,
			"source": host,
			"port":   port,
		}

		fields := map[string]interface{}{
			"block_size":   int64(77070336),
			"context_size": int64(6663840),
			"in_use":       int64(20772579),
			"lost":         int64(0),
			"total_use":    int64(81804609),
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})

	// Subtest for per-context memory stats
	t.Run("memory_context", func(t *testing.T) {
		require.True(t, acc.HasInt64Field("bind_memory_context", "total"))
		require.True(t, acc.HasInt64Field("bind_memory_context", "in_use"))
	})
}

func TestBindXmlStatsV3(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	url := ts.Listener.Addr().String()
	host, port, err := net.SplitHostPort(url)
	require.NoError(t, err)
	defer ts.Close()

	b := Bind{
		Urls:                 []string{ts.URL + "/xml/v3"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		CountersAsInt:        true,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
	}

	var acc testutil.Accumulator
	err = acc.GatherError(b.Gather)
	require.NoError(t, err)

	// Use subtests for counters, since they are similar structure
	type fieldSet struct {
		fieldKey   string
		fieldValue int64
	}

	testCases := []struct {
		counterType string
		values      []fieldSet
	}{
		{
			"opcode",
			[]fieldSet{
				{"NOTIFY", 0},
				{"UPDATE", 0},
				{"IQUERY", 0},
				{"QUERY", 74941},
				{"STATUS", 0},
			},
		},
		{
			"qtype",
			[]fieldSet{
				{"ANY", 22},
				{"SOA", 18},
				{"AAAA", 5735},
				{"MX", 618},
				{"NS", 373},
				{"TXT", 970},
				{"A", 63672},
				{"SRV", 139},
				{"PTR", 3393},
				{"RRSIG", 1},
			},
		},
		{
			"nsstat",
			[]fieldSet{
				{"DNS64", 0},
				{"ExpireOpt", 0},
				{"NSIDOpt", 0},
				{"OtherOpt", 59},
				{"XfrReqDone", 0},
				{"ReqEdns0", 9250},
				{"ReqTSIG", 0},
				{"UpdateRespFwd", 0},
				{"RespEDNS0", 9250},
				{"QryDropped", 11},
				{"RPZRewrites", 0},
				{"XfrRej", 0},
				{"RecQryRej", 35},
				{"QryNxrrset", 2452},
				{"QryFORMERR", 0},
				{"ReqTCP", 260},
				{"QryTCP", 258},
				{"QryUDP", 74648},
				{"UpdateDone", 0},
				{"QrySERVFAIL", 122},
				{"QryRecursion", 53750},
				{"RecursClients", 0},
				{"Requestv4", 74942},
				{"UpdateFwdFail", 0},
				{"QryReferral", 0},
				{"Response", 63264},
				{"RespTSIG", 0},
				{"QrySuccess", 49044},
				{"QryFailure", 35},
				{"RespSIG0", 0},
				{"ReqSIG0", 0},
				{"UpdateRej", 0},
				{"QryAuthAns", 2752},
				{"UpdateFail", 0},
				{"QryDuplicate", 11667},
				{"RateDropped", 0},
				{"QryNoauthAns", 60354},
				{"QryNXDOMAIN", 11610},
				{"ReqBadSIG", 0},
				{"UpdateReqFwd", 0},
				{"RateSlipped", 0},
				{"TruncatedResp", 365},
				{"Requestv6", 0},
				{"UpdateBadPrereq", 0},
				{"AuthQryRej", 0},
				{"ReqBadEDNSVer", 0},
				{"SitBadSize", 0},
				{"SitBadTime", 0},
				{"SitMatch", 0},
				{"SitNew", 0},
				{"SitNoMatch", 0},
				{"SitOpt", 0},
				{"TruncatedResp", 365},
			},
		},
		{
			"sockstat",
			[]fieldSet{
				{"FDwatchConnFail", 0},
				{"UnixClose", 0},
				{"TCP6OpenFail", 0},
				{"TCP6Active", 0},
				{"UDP4RecvErr", 14},
				{"TCP6Conn", 0},
				{"FDWatchClose", 0},
				{"TCP4ConnFail", 0},
				{"UnixConn", 0},
				{"UnixSendErr", 0},
				{"UDP6Close", 0},
				{"UnixOpen", 0},
				{"UDP4Conn", 92535},
				{"TCP4Close", 336},
				{"UnixAcceptFail", 0},
				{"UnixAccept", 0},
				{"TCP6AcceptFail", 0},
				{"UDP6Open", 0},
				{"UDP6BindFail", 0},
				{"UDP6RecvErr", 0},
				{"RawOpenFail", 0},
				{"TCP4Accept", 293},
				{"UDP6SendErr", 0},
				{"UDP6Conn", 0},
				{"TCP4SendErr", 0},
				{"UDP4BindFail", 1},
				{"UDP4Active", 4},
				{"TCP4Active", 297},
				{"UnixConnFail", 0},
				{"UnixOpenFail", 0},
				{"UDP6ConnFail", 0},
				{"TCP6Accept", 0},
				{"UnixRecvErr", 0},
				{"RawActive", 1},
				{"UDP6OpenFail", 0},
				{"RawClose", 0},
				{"UnixBindFail", 0},
				{"UnixActive", 0},
				{"FdwatchBindFail", 0},
				{"UDP4SendErr", 0},
				{"RawRecvErr", 0},
				{"TCP6Close", 0},
				{"FDwatchRecvErr", 0},
				{"TCP4BindFail", 0},
				{"TCP4AcceptFail", 0},
				{"TCP4OpenFail", 0},
				{"UDP4Open", 92542},
				{"UDP4ConnFail", 0},
				{"TCP4Conn", 44},
				{"TCP6ConnFail", 0},
				{"FDwatchConn", 0},
				{"UDP6Active", 0},
				{"RawOpen", 1},
				{"TCP6BindFail", 0},
				{"UDP4Close", 92538},
				{"TCP6Open", 0},
				{"TCP6SendErr", 0},
				{"TCP4Open", 48},
				{"FDwatchSendErr", 0},
				{"TCP6RecvErr", 0},
				{"UDP4OpenFail", 0},
				{"TCP4RecvErr", 0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.counterType, func(t *testing.T) {
			tags := map[string]string{
				"url":    url,
				"type":   tc.counterType,
				"source": host,
				"port":   port,
			}

			fields := make(map[string]interface{}, len(tc.values))
			for _, val := range tc.values {
				fields[val.fieldKey] = val.fieldValue
			}
			acc.AssertContainsTaggedFields(t, "bind_counter", fields, tags)
		})
	}

	// Subtest for memory stats
	t.Run("memory", func(t *testing.T) {
		tags := map[string]string{
			"url":    url,
			"source": host,
			"port":   port,
		}

		fields := map[string]interface{}{
			"block_size":   int64(45875200),
			"context_size": int64(10037400),
			"in_use":       int64(6000232),
			"lost":         int64(0),
			"total_use":    int64(777821909),
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})

	// Subtest for per-context memory stats
	t.Run("memory_context", func(t *testing.T) {
		require.True(t, acc.HasInt64Field("bind_memory_context", "total"))
		require.True(t, acc.HasInt64Field("bind_memory_context", "in_use"))
	})
}

func TestBindXmlStatsV3Signed(t *testing.T) {
	// Setup a mock server to deliver the stats
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	url := ts.Listener.Addr().String()
	host, port, err := net.SplitHostPort(url)
	require.NoError(t, err)
	defer ts.Close()

	// Setup the plugin
	plugin := &Bind{
		Urls:                 []string{ts.URL + "/xml/v3"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		CountersAsInt:        true,
		Timeout:              config.Duration(4 * time.Second),
	}
	require.NoError(t, plugin.Init())

	// Create the expectations
	expected := []telegraf.Metric{
		metric.New(
			"bind_memory",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
			},
			map[string]interface{}{
				"block_size":   int64(45875200),
				"context_size": int64(10037400),
				"in_use":       int64(6000232),
				"lost":         int64(0),
				"total_use":    int64(777821909),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x55fb2e042de0",
				"name":   "main",
			},
			map[string]interface{}{
				"in_use": int64(1454904),
				"total":  int64(2706043),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x55fb2e0507e0",
				"name":   "dst",
			},
			map[string]interface{}{
				"in_use": int64(91776),
				"total":  int64(387478),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x55fb2e0938e0",
				"name":   "zonemgr-pool",
			},
			map[string]interface{}{
				"in_use": int64(143776),
				"total":  int64(742986),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00017d0",
				"name":   "threadkey",
			},
			map[string]interface{}{
				"in_use": int64(0),
				"total":  int64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00475f0",
				"name":   "client",
			},
			map[string]interface{}{
				"in_use": int64(8760),
				"total":  int64(267800),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00dfca0",
				"name":   "cache",
			},
			map[string]interface{}{
				"in_use": int64(83650),
				"total":  int64(288938),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00eaa30",
				"name":   "cache_heap",
			},
			map[string]interface{}{
				"in_use": int64(132096),
				"total":  int64(393216),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d01094e0",
				"name":   "res0",
			},
			map[string]interface{}{
				"in_use": int64(0),
				"total":  int64(262144),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d0114270",
				"name":   "res1",
			},
			map[string]interface{}{
				"in_use": int64(0),
				"total":  int64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d011f000",
				"name":   "res2",
			},
			map[string]interface{}{
				"in_use": int64(0),
				"total":  int64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "opcode",
			},
			map[string]interface{}{
				"IQUERY": int64(0),
				"NOTIFY": int64(0),
				"QUERY":  int64(74941),
				"STATUS": int64(0),
				"UPDATE": int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "qtype",
			},
			map[string]interface{}{
				"A":     int64(63672),
				"AAAA":  int64(5735),
				"ANY":   int64(22),
				"MX":    int64(618),
				"NS":    int64(373),
				"PTR":   int64(3393),
				"RRSIG": int64(1),
				"SOA":   int64(18),
				"SRV":   int64(139),
				"TXT":   int64(970),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "nsstat",
			},
			map[string]interface{}{
				"AuthQryRej":      int64(0),
				"DNS64":           int64(0),
				"ExpireOpt":       int64(0),
				"NSIDOpt":         int64(0),
				"OtherOpt":        int64(59),
				"QryAuthAns":      int64(2752),
				"QryDropped":      int64(11),
				"QryDuplicate":    int64(11667),
				"QryFORMERR":      int64(0),
				"QryFailure":      int64(35),
				"QryNXDOMAIN":     int64(11610),
				"QryNoauthAns":    int64(60354),
				"QryNxrrset":      int64(2452),
				"QryRecursion":    int64(53750),
				"QryReferral":     int64(0),
				"QrySERVFAIL":     int64(122),
				"QrySuccess":      int64(49044),
				"QryTCP":          int64(258),
				"QryUDP":          int64(74648),
				"RPZRewrites":     int64(0),
				"RateDropped":     int64(0),
				"RateSlipped":     int64(0),
				"RecQryRej":       int64(35),
				"RecursClients":   int64(0),
				"ReqBadEDNSVer":   int64(0),
				"ReqBadSIG":       int64(0),
				"ReqEdns0":        int64(9250),
				"ReqSIG0":         int64(0),
				"ReqTCP":          int64(260),
				"ReqTSIG":         int64(0),
				"Requestv4":       int64(74942),
				"Requestv6":       int64(0),
				"RespEDNS0":       int64(9250),
				"RespSIG0":        int64(0),
				"RespTSIG":        int64(0),
				"Response":        int64(63264),
				"SitBadSize":      int64(0),
				"SitBadTime":      int64(0),
				"SitMatch":        int64(0),
				"SitNew":          int64(0),
				"SitNoMatch":      int64(0),
				"SitOpt":          int64(0),
				"TruncatedResp":   int64(365),
				"UpdateBadPrereq": int64(0),
				"UpdateDone":      int64(0),
				"UpdateFail":      int64(0),
				"UpdateFwdFail":   int64(0),
				"UpdateRej":       int64(0),
				"UpdateReqFwd":    int64(0),
				"UpdateRespFwd":   int64(0),
				"XfrRej":          int64(0),
				"XfrReqDone":      int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "zonestat",
			},
			map[string]interface{}{
				"AXFRReqv4":   int64(0),
				"AXFRReqv6":   int64(0),
				"IXFRReqv4":   int64(0),
				"IXFRReqv6":   int64(0),
				"NotifyInv4":  int64(0),
				"NotifyInv6":  int64(0),
				"NotifyOutv4": int64(2),
				"NotifyOutv6": int64(0),
				"NotifyRej":   int64(0),
				"SOAOutv4":    int64(0),
				"SOAOutv6":    int64(0),
				"XfrFail":     int64(0),
				"XfrSuccess":  int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "sockstat",
			},
			map[string]interface{}{
				"FDWatchClose":    int64(0),
				"FDwatchConn":     int64(0),
				"FDwatchConnFail": int64(0),
				"FDwatchRecvErr":  int64(0),
				"FDwatchSendErr":  int64(0),
				"FdwatchBindFail": int64(0),
				"RawActive":       int64(1),
				"RawClose":        int64(0),
				"RawOpen":         int64(1),
				"RawOpenFail":     int64(0),
				"RawRecvErr":      int64(0),
				"TCP4Accept":      int64(293),
				"TCP4AcceptFail":  int64(0),
				"TCP4Active":      int64(297),
				"TCP4BindFail":    int64(0),
				"TCP4Close":       int64(336),
				"TCP4ConnFail":    int64(0),
				"TCP4Conn":        int64(44),
				"TCP4Open":        int64(48),
				"TCP4OpenFail":    int64(0),
				"TCP4RecvErr":     int64(0),
				"TCP4SendErr":     int64(0),
				"TCP6Accept":      int64(0),
				"TCP6AcceptFail":  int64(0),
				"TCP6Active":      int64(0),
				"TCP6BindFail":    int64(0),
				"TCP6Close":       int64(0),
				"TCP6Conn":        int64(0),
				"TCP6ConnFail":    int64(0),
				"TCP6Open":        int64(0),
				"TCP6OpenFail":    int64(0),
				"TCP6RecvErr":     int64(0),
				"TCP6SendErr":     int64(0),
				"UDP4Active":      int64(4),
				"UDP4BindFail":    int64(1),
				"UDP4Close":       int64(92538),
				"UDP4Conn":        int64(92535),
				"UDP4ConnFail":    int64(0),
				"UDP4Open":        int64(92542),
				"UDP4OpenFail":    int64(0),
				"UDP4RecvErr":     int64(14),
				"UDP4SendErr":     int64(0),
				"UDP6Active":      int64(0),
				"UDP6BindFail":    int64(0),
				"UDP6Close":       int64(0),
				"UDP6Conn":        int64(0),
				"UDP6ConnFail":    int64(0),
				"UDP6Open":        int64(0),
				"UDP6OpenFail":    int64(0),
				"UDP6RecvErr":     int64(0),
				"UDP6SendErr":     int64(0),
				"UnixAccept":      int64(0),
				"UnixAcceptFail":  int64(0),
				"UnixActive":      int64(0),
				"UnixBindFail":    int64(0),
				"UnixClose":       int64(0),
				"UnixConn":        int64(0),
				"UnixConnFail":    int64(0),
				"UnixOpen":        int64(0),
				"UnixOpenFail":    int64(0),
				"UnixRecvErr":     int64(0),
				"UnixSendErr":     int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "resqtype",
				"view":   "_default",
			},
			map[string]interface{}{
				"A":      int64(61568),
				"AAAA":   int64(3933),
				"DNSKEY": int64(1699),
				"DS":     int64(13749),
				"MX":     int64(286),
				"NS":     int64(9126),
				"PTR":    int64(1249),
				"SRV":    int64(21),
				"TXT":    int64(942),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "resstats",
				"view":   "_default",
			},
			map[string]interface{}{
				"BadEDNSVersion":  int64(0),
				"BucketSize":      int64(31),
				"EDNS0Fail":       int64(0),
				"FORMERR":         int64(0),
				"GlueFetchv4":     int64(1398),
				"GlueFetchv4Fail": int64(3),
				"GlueFetchv6":     int64(0),
				"GlueFetchv6Fail": int64(0),
				"Lame":            int64(12),
				"Mismatch":        int64(0),
				"NXDOMAIN":        int64(8182),
				"NumFetch":        int64(0),
				"OtherError":      int64(0),
				"QryRTT10":        int64(0),
				"QryRTT100":       int64(45760),
				"QryRTT1600":      int64(75),
				"QryRTT1600+":     int64(0),
				"QryRTT500":       int64(45543),
				"QryRTT800":       int64(743),
				"QueryAbort":      int64(0),
				"QueryCurTCP":     int64(0),
				"QueryCurUDP":     int64(0),
				"QuerySockFail":   int64(0),
				"QueryTimeout":    int64(490),
				"Queryv4":         int64(92573),
				"Queryv6":         int64(0),
				"REFUSED":         int64(34),
				"Responsev4":      int64(92135),
				"Responsev6":      int64(0),
				"Retry":           int64(800),
				"SERVFAIL":        int64(318),
				"ServerQuota":     int64(0),
				"SitClientOk":     int64(0),
				"SitClientOut":    int64(0),
				"SitIn":           int64(0),
				"SitOut":          int64(0),
				"Truncated":       int64(42),
				"ValAttempt":      int64(90256),
				"ValFail":         int64(6),
				"ValNegOk":        int64(22850),
				"ValOk":           int64(67322),
				"ZoneQuota":       int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "adbstat",
				"view":   "_default",
			},
			map[string]interface{}{
				"entriescnt": int64(314),
				"namescnt":   int64(316),
				"nentries":   int64(1021),
				"nnames":     int64(1021),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "cachestats",
				"view":   "_default",
			},
			map[string]interface{}{
				"CacheBuckets": int64(519),
				"CacheHits":    int64(1904593),
				"CacheMisses":  int64(96),
				"CacheNodes":   int64(769),
				"DeleteLRU":    int64(0),
				"DeleteTTL":    int64(47518),
				"HeapMemInUse": int64(132096),
				"HeapMemMax":   int64(132096),
				"HeapMemTotal": int64(393216),
				"QueryHits":    int64(336094),
				"QueryMisses":  int64(369336),
				"TreeMemInUse": int64(392128),
				"TreeMemMax":   int64(828966),
				"TreeMemTotal": int64(1464363),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "resstats",
				"view":   "_bind",
			},
			map[string]interface{}{
				"BadEDNSVersion":  int64(0),
				"BucketSize":      int64(31),
				"EDNS0Fail":       int64(0),
				"FORMERR":         int64(0),
				"GlueFetchv4":     int64(0),
				"GlueFetchv4Fail": int64(0),
				"GlueFetchv6":     int64(0),
				"GlueFetchv6Fail": int64(0),
				"Lame":            int64(0),
				"Mismatch":        int64(0),
				"NXDOMAIN":        int64(0),
				"NumFetch":        int64(0),
				"OtherError":      int64(0),
				"QryRTT10":        int64(0),
				"QryRTT100":       int64(0),
				"QryRTT1600":      int64(0),
				"QryRTT1600+":     int64(0),
				"QryRTT500":       int64(0),
				"QryRTT800":       int64(0),
				"QueryAbort":      int64(0),
				"QueryCurTCP":     int64(0),
				"QueryCurUDP":     int64(0),
				"QuerySockFail":   int64(0),
				"QueryTimeout":    int64(0),
				"Queryv4":         int64(0),
				"Queryv6":         int64(0),
				"REFUSED":         int64(0),
				"Responsev4":      int64(0),
				"Responsev6":      int64(0),
				"Retry":           int64(0),
				"SERVFAIL":        int64(0),
				"ServerQuota":     int64(0),
				"SitClientOk":     int64(0),
				"SitClientOut":    int64(0),
				"SitIn":           int64(0),
				"SitOut":          int64(0),
				"Truncated":       int64(0),
				"ValAttempt":      int64(0),
				"ValFail":         int64(0),
				"ValNegOk":        int64(0),
				"ValOk":           int64(0),
				"ZoneQuota":       int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "adbstat",
				"view":   "_bind",
			},
			map[string]interface{}{
				"entriescnt": int64(0),
				"namescnt":   int64(0),
				"nentries":   int64(1021),
				"nnames":     int64(1021),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "cachestats",
				"view":   "_bind",
			},
			map[string]interface{}{
				"CacheBuckets": int64(64),
				"CacheHits":    int64(0),
				"CacheMisses":  int64(0),
				"CacheNodes":   int64(0),
				"DeleteLRU":    int64(0),
				"DeleteTTL":    int64(0),
				"HeapMemInUse": int64(1024),
				"HeapMemMax":   int64(1024),
				"HeapMemTotal": int64(262144),
				"QueryHits":    int64(0),
				"QueryMisses":  int64(0),
				"TreeMemInUse": int64(29608),
				"TreeMemMax":   int64(29608),
				"TreeMemTotal": int64(287392),
			},
			time.Unix(0, 0),
		),
	}

	// Gather and compare
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
func TestBindXmlStatsV3Unsigned(t *testing.T) {
	// Setup a mock server to deliver the stats
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	url := ts.Listener.Addr().String()
	host, port, err := net.SplitHostPort(url)
	require.NoError(t, err)
	defer ts.Close()

	// Setup the plugin
	plugin := &Bind{
		Urls:                 []string{ts.URL + "/xml/v3"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		Timeout:              config.Duration(4 * time.Second),
	}
	require.NoError(t, plugin.Init())

	// Create the expectations
	expected := []telegraf.Metric{
		metric.New(
			"bind_memory",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
			},
			map[string]interface{}{
				"block_size":   uint64(45875200),
				"context_size": uint64(10037400),
				"in_use":       uint64(6000232),
				"lost":         uint64(0),
				"total_use":    uint64(777821909),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x55fb2e042de0",
				"name":   "main",
			},
			map[string]interface{}{
				"in_use": uint64(1454904),
				"total":  uint64(2706043),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x55fb2e0507e0",
				"name":   "dst",
			},
			map[string]interface{}{
				"in_use": uint64(91776),
				"total":  uint64(387478),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x55fb2e0938e0",
				"name":   "zonemgr-pool",
			},
			map[string]interface{}{
				"in_use": uint64(143776),
				"total":  uint64(742986),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00017d0",
				"name":   "threadkey",
			},
			map[string]interface{}{
				"in_use": uint64(0),
				"total":  uint64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00475f0",
				"name":   "client",
			},
			map[string]interface{}{
				"in_use": uint64(8760),
				"total":  uint64(267800),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00dfca0",
				"name":   "cache",
			},
			map[string]interface{}{
				"in_use": uint64(83650),
				"total":  uint64(288938),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d00eaa30",
				"name":   "cache_heap",
			},
			map[string]interface{}{
				"in_use": uint64(132096),
				"total":  uint64(393216),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d01094e0",
				"name":   "res0",
			},
			map[string]interface{}{
				"in_use": uint64(0),
				"total":  uint64(262144),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d0114270",
				"name":   "res1",
			},
			map[string]interface{}{
				"in_use": uint64(0),
				"total":  uint64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_memory_context",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"id":     "0x7f19d011f000",
				"name":   "res2",
			},
			map[string]interface{}{
				"in_use": uint64(0),
				"total":  uint64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "opcode",
			},
			map[string]interface{}{
				"IQUERY": uint64(0),
				"NOTIFY": uint64(0),
				"QUERY":  uint64(74941),
				"STATUS": uint64(0),
				"UPDATE": uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "qtype",
			},
			map[string]interface{}{
				"A":     uint64(63672),
				"AAAA":  uint64(5735),
				"ANY":   uint64(22),
				"MX":    uint64(618),
				"NS":    uint64(373),
				"PTR":   uint64(3393),
				"RRSIG": uint64(1),
				"SOA":   uint64(18),
				"SRV":   uint64(139),
				"TXT":   uint64(970),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "nsstat",
			},
			map[string]interface{}{
				"AuthQryRej":      uint64(0),
				"DNS64":           uint64(0),
				"ExpireOpt":       uint64(0),
				"NSIDOpt":         uint64(0),
				"OtherOpt":        uint64(59),
				"QryAuthAns":      uint64(2752),
				"QryDropped":      uint64(11),
				"QryDuplicate":    uint64(11667),
				"QryFORMERR":      uint64(0),
				"QryFailure":      uint64(35),
				"QryNXDOMAIN":     uint64(11610),
				"QryNoauthAns":    uint64(60354),
				"QryNxrrset":      uint64(2452),
				"QryRecursion":    uint64(53750),
				"QryReferral":     uint64(0),
				"QrySERVFAIL":     uint64(122),
				"QrySuccess":      uint64(49044),
				"QryTCP":          uint64(258),
				"QryUDP":          uint64(74648),
				"RPZRewrites":     uint64(0),
				"RateDropped":     uint64(0),
				"RateSlipped":     uint64(0),
				"RecQryRej":       uint64(35),
				"RecursClients":   uint64(0),
				"ReqBadEDNSVer":   uint64(0),
				"ReqBadSIG":       uint64(0),
				"ReqEdns0":        uint64(9250),
				"ReqSIG0":         uint64(0),
				"ReqTCP":          uint64(260),
				"ReqTSIG":         uint64(0),
				"Requestv4":       uint64(74942),
				"Requestv6":       uint64(0),
				"RespEDNS0":       uint64(9250),
				"RespSIG0":        uint64(0),
				"RespTSIG":        uint64(0),
				"Response":        uint64(63264),
				"SitBadSize":      uint64(0),
				"SitBadTime":      uint64(0),
				"SitMatch":        uint64(0),
				"SitNew":          uint64(0),
				"SitNoMatch":      uint64(0),
				"SitOpt":          uint64(0),
				"TruncatedResp":   uint64(365),
				"UpdateBadPrereq": uint64(0),
				"UpdateDone":      uint64(0),
				"UpdateFail":      uint64(0),
				"UpdateFwdFail":   uint64(0),
				"UpdateRej":       uint64(0),
				"UpdateReqFwd":    uint64(0),
				"UpdateRespFwd":   uint64(0),
				"XfrRej":          uint64(0),
				"XfrReqDone":      uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "zonestat",
			},
			map[string]interface{}{
				"AXFRReqv4":   uint64(0),
				"AXFRReqv6":   uint64(0),
				"IXFRReqv4":   uint64(0),
				"IXFRReqv6":   uint64(0),
				"NotifyInv4":  uint64(0),
				"NotifyInv6":  uint64(0),
				"NotifyOutv4": uint64(2),
				"NotifyOutv6": uint64(0),
				"NotifyRej":   uint64(0),
				"SOAOutv4":    uint64(0),
				"SOAOutv6":    uint64(0),
				"XfrFail":     uint64(0),
				"XfrSuccess":  uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "sockstat",
			},
			map[string]interface{}{
				"FDWatchClose":    uint64(0),
				"FDwatchConn":     uint64(0),
				"FDwatchConnFail": uint64(0),
				"FDwatchRecvErr":  uint64(0),
				"FDwatchSendErr":  uint64(0),
				"FdwatchBindFail": uint64(0),
				"RawActive":       uint64(1),
				"RawClose":        uint64(0),
				"RawOpen":         uint64(1),
				"RawOpenFail":     uint64(0),
				"RawRecvErr":      uint64(0),
				"TCP4Accept":      uint64(293),
				"TCP4AcceptFail":  uint64(0),
				"TCP4Active":      uint64(297),
				"TCP4BindFail":    uint64(0),
				"TCP4Close":       uint64(336),
				"TCP4ConnFail":    uint64(0),
				"TCP4Conn":        uint64(44),
				"TCP4Open":        uint64(48),
				"TCP4OpenFail":    uint64(0),
				"TCP4RecvErr":     uint64(0),
				"TCP4SendErr":     uint64(0),
				"TCP6Accept":      uint64(0),
				"TCP6AcceptFail":  uint64(0),
				"TCP6Active":      uint64(0),
				"TCP6BindFail":    uint64(0),
				"TCP6Close":       uint64(0),
				"TCP6Conn":        uint64(0),
				"TCP6ConnFail":    uint64(0),
				"TCP6Open":        uint64(0),
				"TCP6OpenFail":    uint64(0),
				"TCP6RecvErr":     uint64(0),
				"TCP6SendErr":     uint64(0),
				"UDP4Active":      uint64(4),
				"UDP4BindFail":    uint64(1),
				"UDP4Close":       uint64(92538),
				"UDP4Conn":        uint64(92535),
				"UDP4ConnFail":    uint64(0),
				"UDP4Open":        uint64(92542),
				"UDP4OpenFail":    uint64(0),
				"UDP4RecvErr":     uint64(14),
				"UDP4SendErr":     uint64(0),
				"UDP6Active":      uint64(0),
				"UDP6BindFail":    uint64(0),
				"UDP6Close":       uint64(0),
				"UDP6Conn":        uint64(0),
				"UDP6ConnFail":    uint64(0),
				"UDP6Open":        uint64(0),
				"UDP6OpenFail":    uint64(0),
				"UDP6RecvErr":     uint64(0),
				"UDP6SendErr":     uint64(0),
				"UnixAccept":      uint64(0),
				"UnixAcceptFail":  uint64(0),
				"UnixActive":      uint64(0),
				"UnixBindFail":    uint64(0),
				"UnixClose":       uint64(0),
				"UnixConn":        uint64(0),
				"UnixConnFail":    uint64(0),
				"UnixOpen":        uint64(0),
				"UnixOpenFail":    uint64(0),
				"UnixRecvErr":     uint64(0),
				"UnixSendErr":     uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "resqtype",
				"view":   "_default",
			},
			map[string]interface{}{
				"A":      uint64(61568),
				"AAAA":   uint64(3933),
				"DNSKEY": uint64(1699),
				"DS":     uint64(13749),
				"MX":     uint64(286),
				"NS":     uint64(9126),
				"PTR":    uint64(1249),
				"SRV":    uint64(21),
				"TXT":    uint64(942),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "resstats",
				"view":   "_default",
			},
			map[string]interface{}{
				"BadEDNSVersion":  uint64(0),
				"BucketSize":      uint64(31),
				"EDNS0Fail":       uint64(0),
				"FORMERR":         uint64(0),
				"GlueFetchv4":     uint64(1398),
				"GlueFetchv4Fail": uint64(3),
				"GlueFetchv6":     uint64(0),
				"GlueFetchv6Fail": uint64(0),
				"Lame":            uint64(12),
				"Mismatch":        uint64(0),
				"NXDOMAIN":        uint64(8182),
				"NumFetch":        uint64(0),
				"OtherError":      uint64(0),
				"QryRTT10":        uint64(0),
				"QryRTT100":       uint64(45760),
				"QryRTT1600":      uint64(75),
				"QryRTT1600+":     uint64(0),
				"QryRTT500":       uint64(45543),
				"QryRTT800":       uint64(743),
				"QueryAbort":      uint64(0),
				"QueryCurTCP":     uint64(0),
				"QueryCurUDP":     uint64(0),
				"QuerySockFail":   uint64(0),
				"QueryTimeout":    uint64(490),
				"Queryv4":         uint64(92573),
				"Queryv6":         uint64(0),
				"REFUSED":         uint64(34),
				"Responsev4":      uint64(92135),
				"Responsev6":      uint64(0),
				"Retry":           uint64(800),
				"SERVFAIL":        uint64(318),
				"ServerQuota":     uint64(0),
				"SitClientOk":     uint64(0),
				"SitClientOut":    uint64(0),
				"SitIn":           uint64(0),
				"SitOut":          uint64(0),
				"Truncated":       uint64(42),
				"ValAttempt":      uint64(90256),
				"ValFail":         uint64(6),
				"ValNegOk":        uint64(22850),
				"ValOk":           uint64(67322),
				"ZoneQuota":       uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "adbstat",
				"view":   "_default",
			},
			map[string]interface{}{
				"entriescnt": uint64(314),
				"namescnt":   uint64(316),
				"nentries":   uint64(1021),
				"nnames":     uint64(1021),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "cachestats",
				"view":   "_default",
			},
			map[string]interface{}{
				"CacheBuckets": uint64(519),
				"CacheHits":    uint64(1904593),
				"CacheMisses":  uint64(96),
				"CacheNodes":   uint64(769),
				"DeleteLRU":    uint64(0),
				"DeleteTTL":    uint64(47518),
				"HeapMemInUse": uint64(132096),
				"HeapMemMax":   uint64(132096),
				"HeapMemTotal": uint64(393216),
				"QueryHits":    uint64(336094),
				"QueryMisses":  uint64(369336),
				"TreeMemInUse": uint64(392128),
				"TreeMemMax":   uint64(828966),
				"TreeMemTotal": uint64(1464363),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "resstats",
				"view":   "_bind",
			},
			map[string]interface{}{
				"BadEDNSVersion":  uint64(0),
				"BucketSize":      uint64(31),
				"EDNS0Fail":       uint64(0),
				"FORMERR":         uint64(0),
				"GlueFetchv4":     uint64(0),
				"GlueFetchv4Fail": uint64(0),
				"GlueFetchv6":     uint64(0),
				"GlueFetchv6Fail": uint64(0),
				"Lame":            uint64(0),
				"Mismatch":        uint64(0),
				"NXDOMAIN":        uint64(0),
				"NumFetch":        uint64(0),
				"OtherError":      uint64(0),
				"QryRTT10":        uint64(0),
				"QryRTT100":       uint64(0),
				"QryRTT1600":      uint64(0),
				"QryRTT1600+":     uint64(0),
				"QryRTT500":       uint64(0),
				"QryRTT800":       uint64(0),
				"QueryAbort":      uint64(0),
				"QueryCurTCP":     uint64(0),
				"QueryCurUDP":     uint64(0),
				"QuerySockFail":   uint64(0),
				"QueryTimeout":    uint64(0),
				"Queryv4":         uint64(0),
				"Queryv6":         uint64(0),
				"REFUSED":         uint64(0),
				"Responsev4":      uint64(0),
				"Responsev6":      uint64(0),
				"Retry":           uint64(0),
				"SERVFAIL":        uint64(0),
				"ServerQuota":     uint64(0),
				"SitClientOk":     uint64(0),
				"SitClientOut":    uint64(0),
				"SitIn":           uint64(0),
				"SitOut":          uint64(0),
				"Truncated":       uint64(0),
				"ValAttempt":      uint64(0),
				"ValFail":         uint64(0),
				"ValNegOk":        uint64(0),
				"ValOk":           uint64(0),
				"ZoneQuota":       uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "adbstat",
				"view":   "_bind",
			},
			map[string]interface{}{
				"entriescnt": uint64(0),
				"namescnt":   uint64(0),
				"nentries":   uint64(1021),
				"nnames":     uint64(1021),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"bind_counter",
			map[string]string{
				"url":    url,
				"source": host,
				"port":   port,
				"type":   "cachestats",
				"view":   "_bind",
			},
			map[string]interface{}{
				"CacheBuckets": uint64(64),
				"CacheHits":    uint64(0),
				"CacheMisses":  uint64(0),
				"CacheNodes":   uint64(0),
				"DeleteLRU":    uint64(0),
				"DeleteTTL":    uint64(0),
				"HeapMemInUse": uint64(1024),
				"HeapMemMax":   uint64(1024),
				"HeapMemTotal": uint64(262144),
				"QueryHits":    uint64(0),
				"QueryMisses":  uint64(0),
				"TreeMemInUse": uint64(29608),
				"TreeMemMax":   uint64(29608),
				"TreeMemTotal": uint64(287392),
			},
			time.Unix(0, 0),
		),
	}

	// Gather and compare
	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestBindUnparsableURL(t *testing.T) {
	b := Bind{
		Urls:          []string{"://example.com"},
		CountersAsInt: true,
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)
	require.Contains(t, err.Error(), "unable to parse address")
}
