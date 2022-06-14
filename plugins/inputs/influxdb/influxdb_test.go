package influxdb_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/influxdb"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
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
	fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write([]byte(influxReturn))
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
	fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, err := w.Write([]byte(influxReturn2))
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

const influxReturn = `
{
"cluster": {"name": "cluster", "tags": {}, "values": {}},
"cmdline": ["influxd"],
"cq": {"name": "cq", "tags": {}, "values": {}},
"database:_internal": {"name": "database", "tags": {"database": "_internal"}, "values": {"numMeasurements": 8, "numSeries": 12}},
"database:udp": {"name": "database", "tags": {"database": "udp"}, "values": {"numMeasurements": 14, "numSeries": 38}},
"hh:/Users/csparr/.influxdb/hh": {"name": "hh", "tags": {"path": "/Users/csparr/.influxdb/hh"}, "values": {}},
"httpd::8086": {"name": "httpd", "tags": {"bind": ":8086"}, "values": {"req": 7, "reqActive": 1, "reqDurationNs": 4488799}},
"measurement:cpu_idle.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "cpu_idle"}, "values": {"numSeries": 1}},
"measurement:cpu_usage.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "cpu_usage"}, "values": {"numSeries": 1}},
"measurement:database._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "database"}, "values": {"numSeries": 2}},
"measurement:database.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "database"}, "values": {"numSeries": 2}},
"measurement:httpd.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "httpd"}, "values": {"numSeries": 1}},
"measurement:measurement.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "measurement"}, "values": {"numSeries": 22}},
"measurement:mem.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "mem"}, "values": {"numSeries": 1}},
"measurement:net.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "net"}, "values": {"numSeries": 1}},
"measurement:runtime._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "runtime"}, "values": {"numSeries": 1}},
"measurement:runtime.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "runtime"}, "values": {"numSeries": 1}},
"measurement:shard._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "shard"}, "values": {"numSeries": 2}},
"measurement:shard.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "shard"}, "values": {"numSeries": 1}},
"measurement:subscriber._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "subscriber"}, "values": {"numSeries": 1}},
"measurement:subscriber.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "subscriber"}, "values": {"numSeries": 1}},
"measurement:swap_used.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "swap_used"}, "values": {"numSeries": 1}},
"measurement:tsm1_cache._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "tsm1_cache"}, "values": {"numSeries": 2}},
"measurement:tsm1_cache.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "tsm1_cache"}, "values": {"numSeries": 2}},
"measurement:tsm1_wal._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "tsm1_wal"}, "values": {"numSeries": 2}},
"measurement:tsm1_wal.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "tsm1_wal"}, "values": {"numSeries": 2}},
"measurement:udp._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "udp"}, "values": {"numSeries": 1}},
"measurement:write._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "write"}, "values": {"numSeries": 1}},
"measurement:write.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "write"}, "values": {"numSeries": 1}},
"memstats": {"Alloc":17034016,"TotalAlloc":201739016,"Sys":38537464,"Lookups":77,"Mallocs":570251,"Frees":381008,"HeapAlloc":17034016,"HeapSys":33849344,"HeapIdle":15802368,"HeapInuse":18046976,"HeapReleased":3473408,"HeapObjects":189243,"StackInuse":753664,"StackSys":753664,"MSpanInuse":97440,"MSpanSys":114688,"MCacheInuse":4800,"MCacheSys":16384,"BuckHashSys":1461583,"GCSys":1112064,"OtherSys":1229737,"NextGC":20843042,"LastGC":1460434886475114239,"PauseTotalNs":5132914,"PauseNs":[195052,117751,139370,156933,263089,165249,713747,103904,122015,294408,213753,170864,175845,114221,121563,122409,113098,162219,229257,126726,250774,254235,117206,293588,144279,124306,127053,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"PauseEnd":[1460433856394860455,1460433856398162739,1460433856405888337,1460433856411784017,1460433856417924684,1460433856428385687,1460433856443782908,1460433856456522851,1460433857392743223,1460433866484394564,1460433866494076235,1460433896472438632,1460433957839825106,1460433976473440328,1460434016473413006,1460434096471892794,1460434126470792929,1460434246480428250,1460434366554468369,1460434396471249528,1460434456471205885,1460434476479487292,1460434536471435965,1460434616469784776,1460434736482078216,1460434856544251733,1460434886475114239,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"NumGC":27,"GCCPUFraction":4.287178819113636e-05,"EnableGC":true,"DebugGC":false,"BySize":[{"Size":0,"Mallocs":0,"Frees":0},{"Size":8,"Mallocs":1031,"Frees":955},{"Size":16,"Mallocs":308485,"Frees":142064},{"Size":32,"Mallocs":64937,"Frees":54321},{"Size":48,"Mallocs":33012,"Frees":29754},{"Size":64,"Mallocs":20299,"Frees":18173},{"Size":80,"Mallocs":8186,"Frees":7597},{"Size":96,"Mallocs":9806,"Frees":8982},{"Size":112,"Mallocs":5671,"Frees":4850},{"Size":128,"Mallocs":2972,"Frees":2684},{"Size":144,"Mallocs":4106,"Frees":3719},{"Size":160,"Mallocs":1324,"Frees":911},{"Size":176,"Mallocs":2574,"Frees":2391},{"Size":192,"Mallocs":4053,"Frees":3863},{"Size":208,"Mallocs":442,"Frees":307},{"Size":224,"Mallocs":336,"Frees":172},{"Size":240,"Mallocs":143,"Frees":125},{"Size":256,"Mallocs":542,"Frees":497},{"Size":288,"Mallocs":15971,"Frees":14761},{"Size":320,"Mallocs":245,"Frees":30},{"Size":352,"Mallocs":1299,"Frees":1065},{"Size":384,"Mallocs":138,"Frees":2},{"Size":416,"Mallocs":54,"Frees":47},{"Size":448,"Mallocs":75,"Frees":29},{"Size":480,"Mallocs":6,"Frees":4},{"Size":512,"Mallocs":452,"Frees":422},{"Size":576,"Mallocs":486,"Frees":395},{"Size":640,"Mallocs":81,"Frees":67},{"Size":704,"Mallocs":421,"Frees":397},{"Size":768,"Mallocs":469,"Frees":468},{"Size":896,"Mallocs":1049,"Frees":1010},{"Size":1024,"Mallocs":1078,"Frees":960},{"Size":1152,"Mallocs":750,"Frees":498},{"Size":1280,"Mallocs":84,"Frees":72},{"Size":1408,"Mallocs":218,"Frees":187},{"Size":1536,"Mallocs":73,"Frees":48},{"Size":1664,"Mallocs":43,"Frees":30},{"Size":2048,"Mallocs":153,"Frees":57},{"Size":2304,"Mallocs":41,"Frees":30},{"Size":2560,"Mallocs":18,"Frees":15},{"Size":2816,"Mallocs":164,"Frees":157},{"Size":3072,"Mallocs":0,"Frees":0},{"Size":3328,"Mallocs":13,"Frees":6},{"Size":4096,"Mallocs":101,"Frees":82},{"Size":4608,"Mallocs":32,"Frees":26},{"Size":5376,"Mallocs":165,"Frees":151},{"Size":6144,"Mallocs":15,"Frees":9},{"Size":6400,"Mallocs":1,"Frees":1},{"Size":6656,"Mallocs":1,"Frees":0},{"Size":6912,"Mallocs":0,"Frees":0},{"Size":8192,"Mallocs":13,"Frees":13},{"Size":8448,"Mallocs":0,"Frees":0},{"Size":8704,"Mallocs":1,"Frees":1},{"Size":9472,"Mallocs":6,"Frees":4},{"Size":10496,"Mallocs":0,"Frees":0},{"Size":12288,"Mallocs":41,"Frees":35},{"Size":13568,"Mallocs":0,"Frees":0},{"Size":14080,"Mallocs":0,"Frees":0},{"Size":16384,"Mallocs":4,"Frees":4},{"Size":16640,"Mallocs":0,"Frees":0},{"Size":17664,"Mallocs":0,"Frees":0}]},
"queryExecutor": {"name": "queryExecutor", "tags": {}, "values": {}},
"shard:/Users/csparr/.influxdb/data/_internal/monitor/2:2": {"name": "shard", "tags": {"database": "_internal", "engine": "tsm1", "id": "2", "path": "/Users/csparr/.influxdb/data/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {}},
"shard:/Users/csparr/.influxdb/data/udp/default/1:1": {"name": "shard", "tags": {"database": "udp", "engine": "tsm1", "id": "1", "path": "/Users/csparr/.influxdb/data/udp/default/1", "retentionPolicy": "default"}, "values": {"fieldsCreate": 61, "seriesCreate": 33, "writePointsOk": 3613, "writeReq": 110}},
"subscriber": {"name": "subscriber", "tags": {}, "values": {"pointsWritten": 3613}},
"tsm1_cache:/Users/csparr/.influxdb/data/_internal/monitor/2": {"name": "tsm1_cache", "tags": {"database": "_internal", "path": "/Users/csparr/.influxdb/data/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {"WALCompactionTimeMs": 0, "cacheAgeMs": 1103932, "cachedBytes": 0, "diskBytes": 0, "memBytes": 40480, "snapshotCount": 0}},
"tsm1_cache:/Users/csparr/.influxdb/data/udp/default/1": {"name": "tsm1_cache", "tags": {"database": "udp", "path": "/Users/csparr/.influxdb/data/udp/default/1", "retentionPolicy": "default"}, "values": {"WALCompactionTimeMs": 0, "cacheAgeMs": 1103029, "cachedBytes": 0, "diskBytes": 0, "memBytes": 2359472, "snapshotCount": 0}},
"tsm1_filestore:/Users/csparr/.influxdb/data/_internal/monitor/2": {"name": "tsm1_filestore", "tags": {"database": "_internal", "path": "/Users/csparr/.influxdb/data/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {}},
"tsm1_filestore:/Users/csparr/.influxdb/data/udp/default/1": {"name": "tsm1_filestore", "tags": {"database": "udp", "path": "/Users/csparr/.influxdb/data/udp/default/1", "retentionPolicy": "default"}, "values": {}},
"tsm1_wal:/Users/csparr/.influxdb/wal/_internal/monitor/2": {"name": "tsm1_wal", "tags": {"database": "_internal", "path": "/Users/csparr/.influxdb/wal/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {"currentSegmentDiskBytes": 0, "oldSegmentsDiskBytes": 69532}},
"tsm1_wal:/Users/csparr/.influxdb/wal/udp/default/1": {"name": "tsm1_wal", "tags": {"database": "udp", "path": "/Users/csparr/.influxdb/wal/udp/default/1", "retentionPolicy": "default"}, "values": {"currentSegmentDiskBytes": 193728, "oldSegmentsDiskBytes": 1008330}},
"write": {"name": "write", "tags": {}, "values": {"pointReq": 3613, "pointReqLocal": 3613, "req": 110, "subWriteOk": 110, "writeOk": 110}}
}`

// InfluxDB 1.0+ with tags: null instead of tags: {}.
const influxReturn2 = `
{
"cluster": {"name": "cluster", "tags": null, "values": {}},
"cmdline": ["influxd"],
"cq": {"name": "cq", "tags": null, "values": {}},
"database:_internal": {"name": "database", "tags": {"database": "_internal"}, "values": {"numMeasurements": 8, "numSeries": 12}},
"database:udp": {"name": "database", "tags": {"database": "udp"}, "values": {"numMeasurements": 14, "numSeries": 38}},
"hh:/Users/csparr/.influxdb/hh": {"name": "hh", "tags": {"path": "/Users/csparr/.influxdb/hh"}, "values": {}},
"httpd::8086": {"name": "httpd", "tags": {"bind": ":8086"}, "values": {"req": 7, "reqActive": 1, "reqDurationNs": 4488799}},
"measurement:cpu_idle.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "cpu_idle"}, "values": {"numSeries": 1}},
"measurement:cpu_usage.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "cpu_usage"}, "values": {"numSeries": 1}},
"measurement:database._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "database"}, "values": {"numSeries": 2}},
"measurement:database.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "database"}, "values": {"numSeries": 2}},
"measurement:httpd.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "httpd"}, "values": {"numSeries": 1}},
"measurement:measurement.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "measurement"}, "values": {"numSeries": 22}},
"measurement:mem.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "mem"}, "values": {"numSeries": 1}},
"measurement:net.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "net"}, "values": {"numSeries": 1}},
"measurement:runtime._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "runtime"}, "values": {"numSeries": 1}},
"measurement:runtime.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "runtime"}, "values": {"numSeries": 1}},
"measurement:shard._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "shard"}, "values": {"numSeries": 2}},
"measurement:shard.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "shard"}, "values": {"numSeries": 1}},
"measurement:subscriber._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "subscriber"}, "values": {"numSeries": 1}},
"measurement:subscriber.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "subscriber"}, "values": {"numSeries": 1}},
"measurement:swap_used.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "swap_used"}, "values": {"numSeries": 1}},
"measurement:tsm1_cache._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "tsm1_cache"}, "values": {"numSeries": 2}},
"measurement:tsm1_cache.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "tsm1_cache"}, "values": {"numSeries": 2}},
"measurement:tsm1_wal._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "tsm1_wal"}, "values": {"numSeries": 2}},
"measurement:tsm1_wal.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "tsm1_wal"}, "values": {"numSeries": 2}},
"measurement:udp._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "udp"}, "values": {"numSeries": 1}},
"measurement:write._internal": {"name": "measurement", "tags": {"database": "_internal", "measurement": "write"}, "values": {"numSeries": 1}},
"measurement:write.udp": {"name": "measurement", "tags": {"database": "udp", "measurement": "write"}, "values": {"numSeries": 1}},
"memstats": {"Alloc":17034016,"TotalAlloc":201739016,"Sys":38537464,"Lookups":77,"Mallocs":570251,"Frees":381008,"HeapAlloc":17034016,"HeapSys":33849344,"HeapIdle":15802368,"HeapInuse":18046976,"HeapReleased":3473408,"HeapObjects":189243,"StackInuse":753664,"StackSys":753664,"MSpanInuse":97440,"MSpanSys":114688,"MCacheInuse":4800,"MCacheSys":16384,"BuckHashSys":1461583,"GCSys":1112064,"OtherSys":1229737,"NextGC":20843042,"LastGC":1460434886475114239,"PauseTotalNs":5132914,"PauseNs":[195052,117751,139370,156933,263089,165249,713747,103904,122015,294408,213753,170864,175845,114221,121563,122409,113098,162219,229257,126726,250774,254235,117206,293588,144279,124306,127053,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"PauseEnd":[1460433856394860455,1460433856398162739,1460433856405888337,1460433856411784017,1460433856417924684,1460433856428385687,1460433856443782908,1460433856456522851,1460433857392743223,1460433866484394564,1460433866494076235,1460433896472438632,1460433957839825106,1460433976473440328,1460434016473413006,1460434096471892794,1460434126470792929,1460434246480428250,1460434366554468369,1460434396471249528,1460434456471205885,1460434476479487292,1460434536471435965,1460434616469784776,1460434736482078216,1460434856544251733,1460434886475114239,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"NumGC":27,"GCCPUFraction":4.287178819113636e-05,"EnableGC":true,"DebugGC":false,"BySize":[{"Size":0,"Mallocs":0,"Frees":0},{"Size":8,"Mallocs":1031,"Frees":955},{"Size":16,"Mallocs":308485,"Frees":142064},{"Size":32,"Mallocs":64937,"Frees":54321},{"Size":48,"Mallocs":33012,"Frees":29754},{"Size":64,"Mallocs":20299,"Frees":18173},{"Size":80,"Mallocs":8186,"Frees":7597},{"Size":96,"Mallocs":9806,"Frees":8982},{"Size":112,"Mallocs":5671,"Frees":4850},{"Size":128,"Mallocs":2972,"Frees":2684},{"Size":144,"Mallocs":4106,"Frees":3719},{"Size":160,"Mallocs":1324,"Frees":911},{"Size":176,"Mallocs":2574,"Frees":2391},{"Size":192,"Mallocs":4053,"Frees":3863},{"Size":208,"Mallocs":442,"Frees":307},{"Size":224,"Mallocs":336,"Frees":172},{"Size":240,"Mallocs":143,"Frees":125},{"Size":256,"Mallocs":542,"Frees":497},{"Size":288,"Mallocs":15971,"Frees":14761},{"Size":320,"Mallocs":245,"Frees":30},{"Size":352,"Mallocs":1299,"Frees":1065},{"Size":384,"Mallocs":138,"Frees":2},{"Size":416,"Mallocs":54,"Frees":47},{"Size":448,"Mallocs":75,"Frees":29},{"Size":480,"Mallocs":6,"Frees":4},{"Size":512,"Mallocs":452,"Frees":422},{"Size":576,"Mallocs":486,"Frees":395},{"Size":640,"Mallocs":81,"Frees":67},{"Size":704,"Mallocs":421,"Frees":397},{"Size":768,"Mallocs":469,"Frees":468},{"Size":896,"Mallocs":1049,"Frees":1010},{"Size":1024,"Mallocs":1078,"Frees":960},{"Size":1152,"Mallocs":750,"Frees":498},{"Size":1280,"Mallocs":84,"Frees":72},{"Size":1408,"Mallocs":218,"Frees":187},{"Size":1536,"Mallocs":73,"Frees":48},{"Size":1664,"Mallocs":43,"Frees":30},{"Size":2048,"Mallocs":153,"Frees":57},{"Size":2304,"Mallocs":41,"Frees":30},{"Size":2560,"Mallocs":18,"Frees":15},{"Size":2816,"Mallocs":164,"Frees":157},{"Size":3072,"Mallocs":0,"Frees":0},{"Size":3328,"Mallocs":13,"Frees":6},{"Size":4096,"Mallocs":101,"Frees":82},{"Size":4608,"Mallocs":32,"Frees":26},{"Size":5376,"Mallocs":165,"Frees":151},{"Size":6144,"Mallocs":15,"Frees":9},{"Size":6400,"Mallocs":1,"Frees":1},{"Size":6656,"Mallocs":1,"Frees":0},{"Size":6912,"Mallocs":0,"Frees":0},{"Size":8192,"Mallocs":13,"Frees":13},{"Size":8448,"Mallocs":0,"Frees":0},{"Size":8704,"Mallocs":1,"Frees":1},{"Size":9472,"Mallocs":6,"Frees":4},{"Size":10496,"Mallocs":0,"Frees":0},{"Size":12288,"Mallocs":41,"Frees":35},{"Size":13568,"Mallocs":0,"Frees":0},{"Size":14080,"Mallocs":0,"Frees":0},{"Size":16384,"Mallocs":4,"Frees":4},{"Size":16640,"Mallocs":0,"Frees":0},{"Size":17664,"Mallocs":0,"Frees":0}]},
"queryExecutor": {"name": "queryExecutor", "tags": null, "values": {}},
"shard:/Users/csparr/.influxdb/data/_internal/monitor/2:2": {"name": "shard", "tags": {"database": "_internal", "engine": "tsm1", "id": "2", "path": "/Users/csparr/.influxdb/data/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {}},
"shard:/Users/csparr/.influxdb/data/udp/default/1:1": {"name": "shard", "tags": {"database": "udp", "engine": "tsm1", "id": "1", "path": "/Users/csparr/.influxdb/data/udp/default/1", "retentionPolicy": "default"}, "values": {"fieldsCreate": 61, "seriesCreate": 33, "writePointsOk": 3613, "writeReq": 110}},
"subscriber": {"name": "subscriber", "tags": null, "values": {"pointsWritten": 3613}},
"tsm1_cache:/Users/csparr/.influxdb/data/_internal/monitor/2": {"name": "tsm1_cache", "tags": {"database": "_internal", "path": "/Users/csparr/.influxdb/data/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {"WALCompactionTimeMs": 0, "cacheAgeMs": 1103932, "cachedBytes": 0, "diskBytes": 0, "memBytes": 40480, "snapshotCount": 0}},
"tsm1_cache:/Users/csparr/.influxdb/data/udp/default/1": {"name": "tsm1_cache", "tags": {"database": "udp", "path": "/Users/csparr/.influxdb/data/udp/default/1", "retentionPolicy": "default"}, "values": {"WALCompactionTimeMs": 0, "cacheAgeMs": 1103029, "cachedBytes": 0, "diskBytes": 0, "memBytes": 2359472, "snapshotCount": 0}},
"tsm1_filestore:/Users/csparr/.influxdb/data/_internal/monitor/2": {"name": "tsm1_filestore", "tags": {"database": "_internal", "path": "/Users/csparr/.influxdb/data/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {}},
"tsm1_filestore:/Users/csparr/.influxdb/data/udp/default/1": {"name": "tsm1_filestore", "tags": {"database": "udp", "path": "/Users/csparr/.influxdb/data/udp/default/1", "retentionPolicy": "default"}, "values": {}},
"tsm1_wal:/Users/csparr/.influxdb/wal/_internal/monitor/2": {"name": "tsm1_wal", "tags": {"database": "_internal", "path": "/Users/csparr/.influxdb/wal/_internal/monitor/2", "retentionPolicy": "monitor"}, "values": {"currentSegmentDiskBytes": 0, "oldSegmentsDiskBytes": 69532}},
"tsm1_wal:/Users/csparr/.influxdb/wal/udp/default/1": {"name": "tsm1_wal", "tags": {"database": "udp", "path": "/Users/csparr/.influxdb/wal/udp/default/1", "retentionPolicy": "default"}, "values": {"currentSegmentDiskBytes": 193728, "oldSegmentsDiskBytes": 1008330}},
"write": {"name": "write", "tags": null, "values": {"pointReq": 3613, "pointReqLocal": 3613, "req": 110, "subWriteOk": 110, "writeOk": 110}}
}`
