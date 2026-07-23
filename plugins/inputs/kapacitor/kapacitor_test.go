package kapacitor_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/kapacitor"
	"github.com/influxdata/telegraf/testutil"
)

func TestKapacitor(t *testing.T) {
	kapacitorReturn, err := os.ReadFile("./testdata/kapacitor_return.json")
	require.NoError(t, err)

	fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			if _, err := w.Write(kapacitorReturn); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeInfluxServer.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{fakeInfluxServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Metrics, 63)

	fields := map[string]interface{}{
		"alloc_bytes":         int64(6950624),
		"buck_hash_sys_bytes": int64(1446737),
		"frees":               int64(129656),
		"gc_cpu_fraction":     float64(0.006757149597237818),
		"gc_sys_bytes":        int64(575488),
		"heap_alloc_bytes":    int64(6950624),
		"heap_idle_bytes":     int64(499712),
		"heap_in_use_bytes":   int64(9166848),
		"heap_objects":        int64(28070),
		"heap_released_bytes": int64(0),
		"heap_sys_bytes":      int64(9666560),
		"last_gc_ns":          int64(1478813691405406556),
		"lookups":             int64(40),
		"mallocs":             int64(157726),
		"mcache_in_use_bytes": int64(9600),
		"mcache_sys_bytes":    int64(16384),
		"mspan_in_use_bytes":  int64(105600),
		"mspan_sys_bytes":     int64(114688),
		"next_gc_ns":          int64(10996691),
		"num_gc":              int64(4),
		"other_sys_bytes":     int64(1985959),
		"pause_total_ns":      int64(767327),
		"stack_in_use_bytes":  int64(819200),
		"stack_sys_bytes":     int64(819200),
		"sys_bytes":           int64(14625016),
		"total_alloc_bytes":   int64(13475176),
	}

	tags := map[string]string{
		"kap_version": "1.1.0~rc2",
		"url":         fakeInfluxServer.URL + "/endpoint",
	}
	acc.AssertContainsTaggedFields(t, "kapacitor_memstats", fields, tags)

	acc.AssertContainsTaggedFields(t, "kapacitor",
		map[string]interface{}{
			"num_enabled_tasks": 5,
			"num_subscriptions": 6,
			"num_tasks":         5,
		}, tags)
}

func TestMissingStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
			return
		}
	}))
	defer server.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{server.URL},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.False(t, acc.HasField("kapacitor_memstats", "alloc_bytes"))
	require.True(t, acc.HasField("kapacitor", "num_tasks"))
}

func TestErrorHandling(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			if _, err := w.Write([]byte("not json")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer badServer.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{badServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	acc.WaitError(1)
	require.Equal(t, uint64(0), acc.NMetrics())
}

func TestErrorHandling404(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer badServer.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{badServer.URL},
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	acc.WaitError(1)
	require.Equal(t, uint64(0), acc.NMetrics())
}

func TestNestedMetricsTagURL(t *testing.T) {
	kapacitorReturn, err := os.ReadFile("./testdata/kapacitor_return.json")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(kapacitorReturn); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	}))
	defer srv.Close()

	// Default: nested metrics must NOT get a url tag (no cardinality change).
	pluginOff := &kapacitor.Kapacitor{URLs: []string{srv.URL}}
	var accOff testutil.Accumulator
	require.NoError(t, pluginOff.Gather(&accOff))
	for _, m := range accOff.Metrics {
		if m.Measurement == "kapacitor_ingress" {
			_, has := m.Tags["url"]
			require.False(t, has, "default must not tag nested metrics with url")
			require.Equal(t, "1.1.0~rc2", m.Fields["kap_version"])
		}
	}

	// Opt-in: nested metrics get url + kap_version field.
	pluginOn := &kapacitor.Kapacitor{URLs: []string{srv.URL}, TagURL: true}
	var accOn testutil.Accumulator
	require.NoError(t, pluginOn.Gather(&accOn))
	found := false
	for _, m := range accOn.Metrics {
		if m.Measurement == "kapacitor_ingress" {
			found = true
			require.Equal(t, srv.URL, m.Tags["url"])
			require.Equal(t, "1.1.0~rc2", m.Fields["kap_version"])
			// high-cardinality host/cluster_id/server_id still stripped
			_, hasHost := m.Tags["host"]
			require.False(t, hasHost)
		}
	}
	require.True(t, found, "expected kapacitor_ingress metric")
}
