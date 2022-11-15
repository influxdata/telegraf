package influxdb_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/influxdb"
	"github.com/influxdata/telegraf/testutil"
)

func TestBasic(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write([]byte(basicJSON))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{fakeServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 3)
	fields := map[string]interface{}{
		// JSON will truncate floats to integer representations.
		// Since there's no distinction in JSON, we can't assume it's an int.
		"i": -1.0,
		"f": 0.5,
		"b": true,
		"s": "string",
	}
	tags := map[string]string{
		"id":  "ex1",
		"url": fakeServer.URL + "/endpoint",
	}
	acc.AssertContainsTaggedFields(t, "influxdb_foo", fields, tags)

	fields = map[string]interface{}{
		"x": "x",
	}
	tags = map[string]string{
		"id":  "ex2",
		"url": fakeServer.URL + "/endpoint",
	}
	acc.AssertContainsTaggedFields(t, "influxdb_bar", fields, tags)

	acc.AssertContainsTaggedFields(t, "influxdb",
		map[string]interface{}{
			"n_shards": 0,
		}, map[string]string{})
}

func TestInfluxDB(t *testing.T) {
	influxReturn, err := os.ReadFile("./testdata/influx_return.json")
	require.NoError(t, err)

	fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write(influxReturn)
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeInfluxServer.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{fakeInfluxServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 34)

	fields := map[string]interface{}{
		"heap_inuse":      int64(18046976),
		"heap_released":   int64(3473408),
		"mspan_inuse":     int64(97440),
		"total_alloc":     int64(201739016),
		"sys":             int64(38537464),
		"mallocs":         int64(570251),
		"frees":           int64(381008),
		"heap_idle":       int64(15802368),
		"pause_total_ns":  int64(5132914),
		"pause_ns":        int64(127053),
		"lookups":         int64(77),
		"heap_sys":        int64(33849344),
		"mcache_sys":      int64(16384),
		"next_gc":         int64(20843042),
		"gc_cpu_fraction": float64(4.287178819113636e-05),
		"other_sys":       int64(1229737),
		"alloc":           int64(17034016),
		"stack_inuse":     int64(753664),
		"stack_sys":       int64(753664),
		"buck_hash_sys":   int64(1461583),
		"gc_sys":          int64(1112064),
		"num_gc":          int64(27),
		"heap_alloc":      int64(17034016),
		"heap_objects":    int64(189243),
		"mspan_sys":       int64(114688),
		"mcache_inuse":    int64(4800),
		"last_gc":         int64(1460434886475114239),
	}

	tags := map[string]string{
		"url": fakeInfluxServer.URL + "/endpoint",
	}
	acc.AssertContainsTaggedFields(t, "influxdb_memstats", fields, tags)

	acc.AssertContainsTaggedFields(t, "influxdb",
		map[string]interface{}{
			"n_shards": 1,
		}, map[string]string{})
}

func TestInfluxDB2(t *testing.T) {
	// InfluxDB 1.0+ with tags: null instead of tags: {}.
	influxReturn2, err := os.ReadFile("./testdata/influx_return2.json")
	require.NoError(t, err)

	fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write(influxReturn2)
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeInfluxServer.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{fakeInfluxServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 34)

	acc.AssertContainsTaggedFields(t, "influxdb",
		map[string]interface{}{
			"n_shards": 1,
		}, map[string]string{})
}

func TestErrorHandling(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write([]byte("not json"))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer badServer.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{badServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(plugin.Gather))
}

func TestErrorHandling404(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write([]byte(basicJSON))
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer badServer.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{badServer.URL},
	}

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(plugin.Gather))
}

func TestErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte(`{"error": "unable to parse authentication credentials"}`))
		require.NoError(t, err)
	}))
	defer ts.Close()

	plugin := &influxdb.InfluxDB{
		URLs: []string{ts.URL},
	}

	var acc testutil.Accumulator
	err := plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []error{
		&influxdb.APIError{
			StatusCode:  http.StatusUnauthorized,
			Reason:      fmt.Sprintf("%d %s", http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)),
			Description: "unable to parse authentication credentials",
		},
	}
	require.Equal(t, expected, acc.Errors)
}

const basicJSON = `
{
  "_1": {
    "name": "foo",
    "tags": {
      "id": "ex1"
    },
    "values": {
      "i": -1,
      "f": 0.5,
      "b": true,
      "s": "string"
    }
  },
  "ignored": {
    "willBeRecorded": false
  },
  "ignoredAndNested": {
    "hash": {
      "is": "nested"
    }
  },
  "array": [
   "makes parsing more difficult than necessary"
  ],
  "string": "makes parsing more difficult than necessary",
  "_2": {
    "name": "bar",
    "tags": {
      "id": "ex2"
    },
    "values": {
      "x": "x"
    }
  },
  "pointWithoutFields_willNotBeIncluded": {
    "name": "asdf",
    "tags": {
      "id": "ex3"
    },
    "values": {}
  }
}
`
