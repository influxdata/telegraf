package bind

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
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
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)

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
				fieldSet{"NOTIFY", 0},
				fieldSet{"UPDATE", 0},
				fieldSet{"IQUERY", 0},
				fieldSet{"QUERY", 13},
				fieldSet{"STATUS", 0},
			},
		},
		{
			"qtype",
			[]fieldSet{
				fieldSet{"A", 2},
				fieldSet{"AAAA", 2},
				fieldSet{"PTR", 7},
				fieldSet{"SRV", 2},
			},
		},
		{
			"nsstat",
			[]fieldSet{
				fieldSet{"QrySuccess", 6},
				fieldSet{"QryRecursion", 12},
				fieldSet{"Requestv4", 13},
				fieldSet{"QryNXDOMAIN", 4},
				fieldSet{"QryAuthAns", 1},
				fieldSet{"QryNxrrset", 1},
				fieldSet{"QryNoauthAns", 10},
				fieldSet{"QryUDP", 13},
				fieldSet{"QryDuplicate", 1},
				fieldSet{"QrySERVFAIL", 1},
				fieldSet{"Response", 12},
			},
		},
		{
			"sockstat",
			[]fieldSet{
				fieldSet{"TCP4Open", 118},
				fieldSet{"UDP6Close", 112},
				fieldSet{"UDP4Close", 333},
				fieldSet{"TCP4Close", 119},
				fieldSet{"TCP6Active", 2},
				fieldSet{"UDP4Active", 2},
				fieldSet{"UDP4RecvErr", 1},
				fieldSet{"UDP4Open", 335},
				fieldSet{"TCP4Active", 10},
				fieldSet{"RawActive", 1},
				fieldSet{"UDP6ConnFail", 112},
				fieldSet{"TCP4Conn", 114},
				fieldSet{"UDP6Active", 1},
				fieldSet{"UDP6Open", 113},
				fieldSet{"UDP4Conn", 333},
				fieldSet{"UDP6SendErr", 112},
				fieldSet{"RawOpen", 1},
				fieldSet{"TCP4Accept", 6},
				fieldSet{"TCP6Open", 2},
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
			"block_size":   13893632,
			"context_size": 3685480,
			"in_use":       3064368,
			"lost":         0,
			"total_use":    18206566,
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})

	// Subtest for per-context memory stats
	t.Run("memory_context", func(t *testing.T) {
		assert.True(t, acc.HasIntField("bind_memory_context", "total"))
		assert.True(t, acc.HasIntField("bind_memory_context", "in_use"))
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
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)

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
				fieldSet{"UPDATE", 238},
				fieldSet{"QUERY", 102312374},
			},
		},
		{
			"qtype",
			[]fieldSet{
				fieldSet{"ANY", 7},
				fieldSet{"DNSKEY", 452},
				fieldSet{"SSHFP", 2987},
				fieldSet{"SOA", 100415},
				fieldSet{"AAAA", 37786321},
				fieldSet{"MX", 441155},
				fieldSet{"IXFR", 157},
				fieldSet{"CNAME", 531},
				fieldSet{"NS", 1999},
				fieldSet{"TXT", 34628},
				fieldSet{"A", 58951432},
				fieldSet{"SRV", 741082},
				fieldSet{"PTR", 4211487},
				fieldSet{"NAPTR", 39137},
				fieldSet{"DS", 584},
			},
		},
		{
			"nsstat",
			[]fieldSet{
				fieldSet{"XfrReqDone", 157},
				fieldSet{"ReqEdns0", 441758},
				fieldSet{"ReqTSIG", 0},
				fieldSet{"UpdateRespFwd", 0},
				fieldSet{"RespEDNS0", 441748},
				fieldSet{"QryDropped", 16},
				fieldSet{"RPZRewrites", 0},
				fieldSet{"XfrRej", 0},
				fieldSet{"RecQryRej", 0},
				fieldSet{"QryNxrrset", 24423133},
				fieldSet{"QryFORMERR", 0},
				fieldSet{"ReqTCP", 1548156},
				fieldSet{"UpdateDone", 0},
				fieldSet{"QrySERVFAIL", 14422},
				fieldSet{"QryRecursion", 2104239},
				fieldSet{"Requestv4", 102312611},
				fieldSet{"UpdateFwdFail", 0},
				fieldSet{"QryReferral", 3},
				fieldSet{"Response", 102301560},
				fieldSet{"RespTSIG", 0},
				fieldSet{"QrySuccess", 63811668},
				fieldSet{"QryFailure", 0},
				fieldSet{"RespSIG0", 0},
				fieldSet{"ReqSIG0", 0},
				fieldSet{"UpdateRej", 238},
				fieldSet{"QryAuthAns", 72180718},
				fieldSet{"UpdateFail", 0},
				fieldSet{"QryDuplicate", 10879},
				fieldSet{"RateDropped", 0},
				fieldSet{"QryNoauthAns", 30106182},
				fieldSet{"QryNXDOMAIN", 14052096},
				fieldSet{"ReqBadSIG", 0},
				fieldSet{"UpdateReqFwd", 0},
				fieldSet{"RateSlipped", 0},
				fieldSet{"TruncatedResp", 3787},
				fieldSet{"Requestv6", 1},
				fieldSet{"UpdateBadPrereq", 0},
				fieldSet{"AuthQryRej", 0},
				fieldSet{"ReqBadEDNSVer", 0},
			},
		},
		{
			"sockstat",
			[]fieldSet{
				fieldSet{"FdwatchBindFail", 0},
				fieldSet{"UDP6Open", 238269},
				fieldSet{"UDP6SendErr", 238250},
				fieldSet{"TCP4ConnFail", 0},
				fieldSet{"TCP4Conn", 590},
				fieldSet{"TCP6AcceptFail", 0},
				fieldSet{"UDP4SendErr", 0},
				fieldSet{"FDwatchConn", 0},
				fieldSet{"TCP4RecvErr", 1},
				fieldSet{"TCP4OpenFail", 0},
				fieldSet{"UDP4OpenFail", 0},
				fieldSet{"UDP6OpenFail", 0},
				fieldSet{"TCP4Close", 1548268},
				fieldSet{"TCP6BindFail", 0},
				fieldSet{"TCP4AcceptFail", 0},
				fieldSet{"UnixConn", 0},
				fieldSet{"UDP4Open", 3765532},
				fieldSet{"TCP6Close", 0},
				fieldSet{"FDwatchRecvErr", 0},
				fieldSet{"UDP4Conn", 3764828},
				fieldSet{"UnixConnFail", 0},
				fieldSet{"TCP6Conn", 0},
				fieldSet{"TCP6OpenFail", 0},
				fieldSet{"TCP6SendErr", 0},
				fieldSet{"TCP6RecvErr", 0},
				fieldSet{"FDwatchSendErr", 0},
				fieldSet{"UDP4RecvErr", 1650},
				fieldSet{"UDP4ConnFail", 0},
				fieldSet{"UDP6Close", 238267},
				fieldSet{"FDWatchClose", 0},
				fieldSet{"TCP4Accept", 1547672},
				fieldSet{"UnixAccept", 0},
				fieldSet{"TCP4Open", 602},
				fieldSet{"UDP4BindFail", 219},
				fieldSet{"UDP6ConnFail", 238250},
				fieldSet{"UnixClose", 0},
				fieldSet{"TCP4BindFail", 0},
				fieldSet{"UnixOpenFail", 0},
				fieldSet{"UDP6BindFail", 16},
				fieldSet{"UnixOpen", 0},
				fieldSet{"UnixAcceptFail", 0},
				fieldSet{"UnixRecvErr", 0},
				fieldSet{"UDP6RecvErr", 0},
				fieldSet{"TCP6ConnFail", 0},
				fieldSet{"FDwatchConnFail", 0},
				fieldSet{"TCP4SendErr", 0},
				fieldSet{"UDP4Close", 3765528},
				fieldSet{"UnixSendErr", 0},
				fieldSet{"TCP6Open", 2},
				fieldSet{"UDP6Conn", 1},
				fieldSet{"TCP6Accept", 0},
				fieldSet{"UnixBindFail", 0},
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
			"block_size":   77070336,
			"context_size": 6663840,
			"in_use":       20772579,
			"lost":         0,
			"total_use":    81804609,
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})

	// Subtest for per-context memory stats
	t.Run("memory_context", func(t *testing.T) {
		assert.True(t, acc.HasIntField("bind_memory_context", "total"))
		assert.True(t, acc.HasIntField("bind_memory_context", "in_use"))
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
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)

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
				fieldSet{"NOTIFY", 0},
				fieldSet{"UPDATE", 0},
				fieldSet{"IQUERY", 0},
				fieldSet{"QUERY", 13},
				fieldSet{"STATUS", 0},
			},
		},
		{
			"qtype",
			[]fieldSet{
				fieldSet{"A", 2},
				fieldSet{"AAAA", 2},
				fieldSet{"PTR", 7},
				fieldSet{"SRV", 2},
			},
		},
		{
			"nsstat",
			[]fieldSet{
				fieldSet{"QrySuccess", 6},
				fieldSet{"QryRecursion", 12},
				fieldSet{"Requestv4", 13},
				fieldSet{"QryNXDOMAIN", 4},
				fieldSet{"QryAuthAns", 1},
				fieldSet{"QryNxrrset", 1},
				fieldSet{"QryNoauthAns", 10},
				fieldSet{"QryUDP", 13},
				fieldSet{"QryDuplicate", 1},
				fieldSet{"QrySERVFAIL", 1},
				fieldSet{"Response", 12},
			},
		},
		{
			"sockstat",
			[]fieldSet{
				fieldSet{"TCP4Open", 118},
				fieldSet{"UDP6Close", 112},
				fieldSet{"UDP4Close", 333},
				fieldSet{"TCP4Close", 119},
				fieldSet{"TCP6Active", 2},
				fieldSet{"UDP4Active", 2},
				fieldSet{"UDP4RecvErr", 1},
				fieldSet{"UDP4Open", 335},
				fieldSet{"TCP4Active", 10},
				fieldSet{"RawActive", 1},
				fieldSet{"UDP6ConnFail", 112},
				fieldSet{"TCP4Conn", 114},
				fieldSet{"UDP6Active", 1},
				fieldSet{"UDP6Open", 113},
				fieldSet{"UDP4Conn", 333},
				fieldSet{"UDP6SendErr", 112},
				fieldSet{"RawOpen", 1},
				fieldSet{"TCP4Accept", 6},
				fieldSet{"TCP6Open", 2},
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
			"url": url,
		}

		fields := map[string]interface{}{
			"block_size":   45875200,
			"context_size": 10037400,
			"in_use":       6000232,
			"lost":         0,
			"total_use":    777821909,
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})

	// Subtest for per-context memory stats
	t.Run("memory_context", func(t *testing.T) {
		assert.True(t, acc.HasIntField("bind_memory_context", "total"))
		assert.True(t, acc.HasIntField("bind_memory_context", "in_use"))
	})
}

func TestBindUnparseableURL(t *testing.T) {
	b := Bind{
		Urls: []string{"://example.com"},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)
	assert.Contains(t, err.Error(), "Unable to parse address")
}
