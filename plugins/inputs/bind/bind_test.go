package bind

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestBindJsonStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "testdata/bindstats-v1.json")
	}))
	defer ts.Close()

	b := Bind{
		Urls: []string{ts.URL + "/json/v1"},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)

	// Use subtests for counters, since they are similar structure
	testCases := []struct {
		counterType string
		counterName string
		want        int
	}{
		{"opcode", "QUERY", 13},
		{"qtype", "PTR", 7},
		{"nsstat", "QrySuccess", 6},
		{"sockstat", "UDP4Conn", 333},
	}

	for _, tc := range testCases {
		t.Run(tc.counterType, func(t *testing.T) {
			tags := map[string]string{
				"url":  ts.Listener.Addr().String(),
				"type": tc.counterType,
				"name": tc.counterName,
			}

			fields := map[string]interface{}{
				"value": tc.want,
			}

			acc.AssertContainsTaggedFields(t, "bind_counter", fields, tags)
		})
	}

	// Subtest for memory stats
	t.Run("memory", func(t *testing.T) {
		tags := map[string]string{
			"url": ts.Listener.Addr().String(),
		}

		fields := map[string]interface{}{
			"BlockSize":   13893632,
			"ContextSize": 3685480,
			"InUse":       3064368,
			"Lost":        0,
			"TotalUse":    18206566,
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})
}

func TestBindXmlStatsV2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		http.ServeFile(w, r, "testdata/bindstats-v2.xml")
	}))
	defer ts.Close()

	b := Bind{
		Urls: []string{ts.URL + "/xml/v2"},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)

	// Use subtests for counters, since they are similar structure
	testCases := []struct {
		counterType string
		counterName string
		want        int
	}{
		{"opcode", "QUERY", 102312374},
		{"qtype", "PTR", 4211487},
		{"nsstat", "QrySuccess", 63811668},
		{"zonestat", "NotifyOutv4", 663},
		{"sockstat", "UDP4Conn", 3764828},
	}

	for _, tc := range testCases {
		t.Run(tc.counterType, func(t *testing.T) {
			tags := map[string]string{
				"url":  ts.Listener.Addr().String(),
				"type": tc.counterType,
				"name": tc.counterName,
			}

			fields := map[string]interface{}{
				"value": tc.want,
			}

			acc.AssertContainsTaggedFields(t, "bind_counter", fields, tags)
		})
	}

	// Subtest for memory stats
	t.Run("memory", func(t *testing.T) {
		tags := map[string]string{
			"url": ts.Listener.Addr().String(),
		}

		fields := map[string]interface{}{
			"BlockSize":   77070336,
			"ContextSize": 6663840,
			"InUse":       20772579,
			"Lost":        0,
			"TotalUse":    81804609,
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
	})
}

func TestBindXmlStatsV3(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		http.ServeFile(w, r, "testdata/bindstats-v3.xml")
	}))
	defer ts.Close()

	b := Bind{
		Urls: []string{ts.URL + "/xml/v3"},
	}

	var acc testutil.Accumulator
	err := acc.GatherError(b.Gather)

	assert.Nil(t, err)

	// Use subtests for counters, since they are similar structure
	testCases := []struct {
		counterType string
		counterName string
		want        int
	}{
		{"opcode", "QUERY", 74941},
		{"qtype", "PTR", 3393},
		{"nsstat", "QrySuccess", 49044},
		{"zonestat", "NotifyOutv4", 2},
		{"sockstat", "UDP4Conn", 92535},
	}

	for _, tc := range testCases {
		t.Run(tc.counterType, func(t *testing.T) {
			tags := map[string]string{
				"url":  ts.Listener.Addr().String(),
				"type": tc.counterType,
				"name": tc.counterName,
			}

			fields := map[string]interface{}{
				"value": tc.want,
			}

			acc.AssertContainsTaggedFields(t, "bind_counter", fields, tags)
		})
	}

	// Subtest for memory stats
	t.Run("memory", func(t *testing.T) {
		tags := map[string]string{
			"url": ts.Listener.Addr().String(),
		}

		fields := map[string]interface{}{
			"BlockSize":   45875200,
			"ContextSize": 10037400,
			"InUse":       6000232,
			"Lost":        0,
			"TotalUse":    777821909,
		}

		acc.AssertContainsTaggedFields(t, "bind_memory", fields, tags)
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
