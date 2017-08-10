package bind

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestBindJsonStats(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))
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
