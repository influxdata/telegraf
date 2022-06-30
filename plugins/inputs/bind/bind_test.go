package bind

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestBindJsonStats(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	url := ts.Listener.Addr().String()
	host, port, _ := net.SplitHostPort(url)
	defer ts.Close()

	b := Bind{
		Urls:                 []string{ts.URL + "/json/v1"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

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
	host, port, _ := net.SplitHostPort(url)
	defer ts.Close()

	b := Bind{
		Urls:                 []string{ts.URL + "/xml/v2"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

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
	host, port, _ := net.SplitHostPort(url)
	defer ts.Close()

	b := Bind{
		Urls:                 []string{ts.URL + "/xml/v3"},
		GatherMemoryContexts: true,
		GatherViews:          true,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

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

func TestBindUnparseableURL(t *testing.T) {
	b := Bind{
		Urls: []string{"://example.com"},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)
	require.Contains(t, err.Error(), "unable to parse address")
}
