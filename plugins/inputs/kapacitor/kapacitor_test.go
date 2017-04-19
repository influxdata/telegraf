package kapacitor_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/kapacitor"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestKapacitor(t *testing.T) {
	fakeInfluxServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(kapacitorReturn))
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
		"gcc_pu_fraction":     float64(0.006757149597237818),
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{server.URL},
	}

	var acc testutil.Accumulator
	plugin.Gather(&acc)

	require.False(t, acc.HasField("kapacitor_memstats", "alloc_bytes"))
	require.True(t, acc.HasField("kapacitor", "num_tasks"))
}

func TestErrorHandling(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte("not json"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer badServer.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{badServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	plugin.Gather(&acc)
	acc.WaitError(1)
	require.Equal(t, uint64(0), acc.NMetrics())
}

func TestErrorHandling404(t *testing.T) {
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer badServer.Close()

	plugin := &kapacitor.Kapacitor{
		URLs: []string{badServer.URL},
	}

	var acc testutil.Accumulator
	plugin.Gather(&acc)
	acc.WaitError(1)
	require.Equal(t, uint64(0), acc.NMetrics())
}

const kapacitorReturn = `{
"cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842",
"cmdline": ["./build/kapacitord","-config","/Users/ross/.kapacitor/config"],
"host": "localhost",
"kapacitor": {"4360b884-4f1b-4915-90b9-e40a27cd366a": {"name": "ingress", "tags": {"task_master": "main", "database": "_internal", "retention_policy": "monitor", "measurement": "cq", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"points_received": 1}}, "b5ca6839-5cda-4330-b7aa-fccc50ce1320": {"name": "edges", "tags": {"type": "stream", "task": "task_master:main", "parent": "write_points", "child": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 72, "emitted": 72}}, "e4f4dd29-96fe-41ca-9107-43030f67d47c": {"name": "edges", "tags": {"host": "localhost", "task": "deadman-test", "parent": "log3", "child": "mean4", "type": "batch", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"collected": 0, "emitted": 0}}, "f8bb9712-52ea-44c8-a0d9-8b6ee47b9cb8": {"name": "edges", "tags": {"task": "derivative-test", "parent": "sum5", "child": "log6", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 0, "emitted": 0}}, "340391e7-8fc7-447d-ad07-547c5dd51f2c": {"tags": {"child": "from1", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "sys-stats", "parent": "stream0"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "2fd00546-3b38-4118-97e6-38489e5732a8": {"name": "edges", "tags": {"task": "sys-stats", "parent": "stream", "child": "stream0", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "type": "stream"}, "values": {"collected": 0, "emitted": 0}}, "5467712d-b3b9-45c5-acae-43a686c345e3": {"name": "edges", "tags": {"host": "localhost", "task": "test", "parent": "stream", "child": "stream0", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"collected": 0, "emitted": 0}}, "2ce99d1d-d26f-4daa-8857-36568a9effc1": {"name": "nodes", "tags": {"task": "deadman-test", "node": "window2", "type": "stream", "kind": "window", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "975e59b2-d0ab-4ebf-9054-4de9751e0e87": {"name": "edges", "tags": {"task": "deadman-test", "parent": "derivative5", "child": "window6", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 0, "emitted": 0}}, "14fc028b-bda9-4b29-bf8e-8f682dd42bab": {"name": "nodes", "tags": {"task": "derivative-test", "node": "from1", "type": "stream", "kind": "from", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "69f374fb-1289-457a-9ae6-79263e2dc430": {"tags": {"child": "derivative2", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "parent": "from1"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "0c0a0463-4588-494b-b871-f20dea1b4c71": {"name": "nodes", "tags": {"task": "derivative-test", "node": "log6", "type": "stream", "kind": "log", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "e6eedf33-a07c-4daf-87fb-9429ee38847f": {"name": "nodes", "tags": {"host": "localhost", "kind": "from", "task": "deadman-test", "node": "from1", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"avg_exec_time_ns": "0s"}}, "c8da7d8a-6ea1-4cac-88aa-0804c413eecb": {"tags": {"child": "from1", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "parent": "stream0"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "ad4ca70d-1e56-422e-b36d-af565a8e1df0": {"tags": {"type": "stream", "kind": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "node": "stream0"}, "values": {"avg_exec_time_ns": "0s"}, "name": "nodes"}, "39fd1e7d-24bf-4cae-9c52-f9d876cc92af": {"name": "edges", "tags": {"task": "test", "parent": "from1", "child": "alert2", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"emitted": 0, "collected": 0}}, "1be18262-c33e-4cd6-bd49-7c82b819d4e1": {"values": {"avg_exec_time_ns": "0s"}, "name": "nodes", "tags": {"host": "localhost", "task": "deadman-test", "node": "stream0", "type": "stream", "kind": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}}, "ce957f67-bef1-4755-acf1-1a335cbf0dcb": {"name": "edges", "tags": {"task": "deadman-test", "parent": "from1", "child": "window2", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 0, "emitted": 0}}, "c33ac068-6216-4ab2-b10d-b82f08891a10": {"name": "nodes", "tags": {"kind": "window", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "node": "window6", "type": "stream"}, "values": {"avg_exec_time_ns": "0s"}}, "2ab14f75-6a7d-49a8-a5bd-b829ec7588b5": {"name": "edges", "tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "parent": "derivative2", "child": "log3", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"collected": 0, "emitted": 0}}, "94031aa9-20fa-448a-8952-ba37d69a991a": {"name": "nodes", "tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "node": "eval4", "type": "stream", "kind": "eval", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"avg_exec_time_ns": "0s", "eval_errors": 0}}, "d3b76aba-b2a2-4e72-a23d-f54c5037d6ee": {"values": {"avg_exec_time_ns": "0s"}, "name": "nodes", "tags": {"host": "localhost", "kind": "stream", "task": "sys-stats", "node": "stream0", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}}, "fe2569b7-275d-411f-86f1-4c75127b0a77": {"name": "nodes", "tags": {"node": "window2", "type": "stream", "kind": "window", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "sys-stats"}, "values": {"avg_exec_time_ns": "0s"}}, "067d801e-48de-4fec-98fe-d6691d3db1fe": {"name": "nodes", "tags": {"task": "derivative-test", "node": "sum5", "type": "stream", "kind": "sum", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "5bc09b36-f1d7-460f-b533-91aa79b79ee3": {"name": "nodes", "tags": {"host": "localhost", "task": "test", "node": "alert2", "type": "stream", "kind": "alert", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"warns_triggered": 0, "crits_triggered": 0, "avg_exec_time_ns": "0s", "alerts_triggered": 0, "oks_triggered": 0, "infos_triggered": 0}}, "874c38c5-380f-4c76-b1ed-af2b7e1bd0f7": {"name": "ingress", "tags": {"database": "_internal", "retention_policy": "monitor", "measurement": "queryExecutor", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task_master": "main"}, "values": {"points_received": 1}}, "38e0634f-9f11-4c3e-8621-96200dfb0ec4": {"name": "nodes", "tags": {"node": "log3", "type": "stream", "kind": "log", "task": "deadman-test", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "193ccaeb-7206-454d-9acc-457506848a09": {"tags": {"child": "sum5", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "parent": "eval4"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "68565c38-e5d2-4563-a2e8-991718154623": {"tags": {"child": "http_out3", "type": "batch", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "sys-stats", "parent": "window2"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "dd1dc51c-ae09-43af-b511-e6d23a7b1940": {"name": "ingress", "tags": {"cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task_master": "main", "database": "_internal", "retention_policy": "monitor", "measurement": "runtime"}, "values": {"points_received": 1}}, "1f60fe6b-3dda-4811-a218-4e54cc6ff20f": {"tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "node": "log7", "type": "stream", "kind": "log", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"avg_exec_time_ns": "0s"}, "name": "nodes"}, "969d3962-ed99-4032-8625-465a2cccb2d9": {"name": "edges", "tags": {"host": "localhost", "task": "derivative-test", "parent": "stream", "child": "stream0", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"collected": 0, "emitted": 0}}, "5d237ee6-8362-427c-982b-f3be4a901e4b": {"tags": {"type": "stream", "kind": "derivative", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "node": "derivative5"}, "values": {"avg_exec_time_ns": "0s"}, "name": "nodes"}, "b92a8b07-713f-4d94-8c99-e1311129018a": {"name": "edges", "tags": {"task": "sys-stats", "parent": "http_out3", "child": "influxdb_out4", "type": "batch", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 0, "emitted": 0}}, "bba6ca07-a577-46b5-90eb-e034b43e2d72": {"name": "nodes", "tags": {"task": "test", "node": "stream0", "type": "stream", "kind": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "056bca53-554b-43e4-8402-7d5e8441f888": {"tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "parent": "window6", "child": "log7", "type": "batch", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "bd5a6337-0a7b-44f1-9768-1d3418fd8d66": {"values": {"collected": 0, "emitted": 0}, "name": "edges", "tags": {"child": "eval4", "type": "stream", "task": "derivative-test", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "parent": "log3"}}, "faec1062-7c9a-44bc-aa7a-3b55bf3c0a45": {"values": {"avg_exec_time_ns": "0s", "points_written": 0, "write_errors": 0}, "name": "nodes", "tags": {"task": "sys-stats", "node": "influxdb_out4", "type": "stream", "kind": "influxdb_out", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}}, "5cb5470f-da25-4d61-977a-9b50d9a3d2d2": {"name": "ingress", "tags": {"host": "localhost", "database": "_internal", "retention_policy": "monitor", "measurement": "tsm1_wal", "task_master": "main", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"points_received": 12}}, "c746289f-c2d3-4c15-8cbc-4ceb9ac72ee4": {"name": "ingress", "tags": {"task_master": "main", "database": "_internal", "retention_policy": "monitor", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "measurement": "database"}, "values": {"points_received": 4}}, "b4e6202b-e826-40bf-835b-f3deb413495f": {"name": "ingress", "tags": {"host": "localhost", "task_master": "main", "database": "_internal", "retention_policy": "monitor", "measurement": "write", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"points_received": 1}}, "bdb9c35c-8f0f-4c64-b111-5afedde6fe26": {"name": "ingress", "tags": {"host": "localhost", "database": "_internal", "retention_policy": "monitor", "measurement": "httpd", "task_master": "main", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"points_received": 1}}, "006be0dc-5861-4fbc-b5af-833d8fccb36c": {"tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "parent": "window2", "child": "log3", "type": "batch", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"collected": 0, "emitted": 0}, "name": "edges"}, "3fa5469d-4599-4307-b67b-934ae74f9266": {"name": "nodes", "tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "deadman-test", "node": "mean4", "type": "stream", "kind": "mean", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"avg_exec_time_ns": "0s"}}, "9636063d-79d2-4480-a03f-10e1b7680a40": {"name": "nodes", "tags": {"kind": "http_out", "task": "sys-stats", "node": "http_out3", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "1ba6153d-227d-4771-ab83-a7bf9f64d1b0": {"name": "edges", "tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "type": "stream", "task": "task_master:main", "parent": "stats", "child": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"collected": 0, "emitted": 0}}, "e7936e59-313c-4fd1-8ae9-06aa11c0ffcb": {"name": "ingress", "tags": {"retention_policy": "monitor", "measurement": "shard", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task_master": "main", "database": "_internal"}, "values": {"points_received": 12}}, "94be0dc3-bee4-4e94-8e3b-7fb856abc490": {"tags": {"type": "stream", "kind": "log", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "node": "log3"}, "values": {"avg_exec_time_ns": "0s"}, "name": "nodes"}, "7cb4de43-c440-477c-beaa-7b402ab16e06": {"name": "edges", "tags": {"child": "window2", "type": "stream", "task": "sys-stats", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "parent": "from1"}, "values": {"collected": 0, "emitted": 0}}, "a2666b71-3057-43d0-887b-8d88ca653b2f": {"name": "ingress", "tags": {"host": "localhost", "task_master": "main", "database": "_internal", "retention_policy": "monitor", "measurement": "subscriber", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"points_received": 2}}, "56727a9f-6f8e-4c34-af3c-32dfb365a407": {"name": "edges", "tags": {"type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "derivative-test", "parent": "stream0", "child": "from1"}, "values": {"collected": 0, "emitted": 0}}, "d2840a0c-b140-419b-b10b-bef3f74cc696": {"name": "nodes", "tags": {"task": "derivative-test", "node": "derivative2", "type": "stream", "kind": "derivative", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "1d223093-e785-4d11-aa9f-8160cf2ddb4d": {"name": "ingress", "tags": {"task_master": "main", "database": "_internal", "retention_policy": "monitor", "measurement": "udp", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"points_received": 1}}, "f1be8ef4-9552-4f6a-83a9-ae2ca0b6e6b5": {"name": "edges", "tags": {"task": "deadman-test", "parent": "mean4", "child": "derivative5", "type": "stream", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 0, "emitted": 0}}, "c1c380da-60b4-422e-a37b-89aff6d11fd1": {"name": "nodes", "tags": {"node": "from1", "type": "stream", "kind": "from", "task": "sys-stats", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"avg_exec_time_ns": "0s"}}, "29321d3f-9f22-4c37-9d48-b8d60aafe1ba": {"name": "nodes", "tags": {"kind": "from", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "task": "test", "node": "from1", "type": "stream"}, "values": {"avg_exec_time_ns": "0s"}}, "901f8a35-b363-47f1-bdfc-c85e3561cc19": {"values": {"collected": 0, "emitted": 0}, "name": "edges", "tags": {"type": "stream", "task": "test", "parent": "stream0", "child": "from1", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}}, "b729d0a0-474e-4bdb-823b-fa810bba3843": {"tags": {"server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost", "measurement": "tsm1_cache", "task_master": "main", "database": "_internal", "retention_policy": "monitor", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842"}, "values": {"points_received": 12}, "name": "ingress"}, "d1526a79-a5f0-45f0-836d-53a924457352": {"name": "edges", "tags": {"type": "stream", "task": "deadman-test", "parent": "stream", "child": "stream0", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"collected": 0, "emitted": 0}}, "fa5165cb-76f0-4127-a9cb-99ab3c34d009": {"name": "ingress", "tags": {"retention_policy": "monitor", "measurement": "tsm1_engine", "task_master": "main", "database": "_internal", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff", "host": "localhost"}, "values": {"points_received": 12}}, "649ffe5d-2682-4675-a943-baec4718fc8f": {"name": "ingress", "tags": {"host": "localhost", "database": "_internal", "retention_policy": "monitor", "measurement": "tsm1_filestore", "task_master": "main", "cluster_id": "aaa1cb78-8277-4886-a8ea-4706bca20842", "server_id": "81e38258-5f18-4201-8d47-db603bfceeff"}, "values": {"points_received": 12}}},
"memstats": {"Alloc":6950624,"TotalAlloc":13475176,"Sys":14625016,"Lookups":40,"Mallocs":157726,"Frees":129656,"HeapAlloc":6950624,"HeapSys":9666560,"HeapIdle":499712,"HeapInuse":9166848,"HeapReleased":0,"HeapObjects":28070,"StackInuse":819200,"StackSys":819200,"MSpanInuse":105600,"MSpanSys":114688,"MCacheInuse":9600,"MCacheSys":16384,"BuckHashSys":1446737,"GCSys":575488,"OtherSys":1985959,"NextGC":10996691,"LastGC":1478813691405406556,"PauseTotalNs":767327,"PauseNs":[106224,484338,77201,99564,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"PauseEnd":[1478813691247788813,1478813691256256020,1478813691262214751,1478813691405406556,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"NumGC":4,"GCCPUFraction":0.006757149597237818,"EnableGC":true,"DebugGC":false,"BySize":[{"Size":0,"Mallocs":0,"Frees":0},{"Size":8,"Mallocs":16193,"Frees":15733},{"Size":16,"Mallocs":54267,"Frees":45512},{"Size":32,"Mallocs":35006,"Frees":25861},{"Size":48,"Mallocs":7130,"Frees":5066},{"Size":64,"Mallocs":3421,"Frees":2554},{"Size":80,"Mallocs":3749,"Frees":3225},{"Size":96,"Mallocs":13973,"Frees":13413},{"Size":112,"Mallocs":1269,"Frees":825},{"Size":128,"Mallocs":1851,"Frees":1755},{"Size":144,"Mallocs":181,"Frees":95},{"Size":160,"Mallocs":558,"Frees":184},{"Size":176,"Mallocs":303,"Frees":249},{"Size":192,"Mallocs":2485,"Frees":278},{"Size":208,"Mallocs":255,"Frees":124},{"Size":224,"Mallocs":316,"Frees":131},{"Size":240,"Mallocs":64,"Frees":28},{"Size":256,"Mallocs":128,"Frees":91},{"Size":288,"Mallocs":734,"Frees":194},{"Size":320,"Mallocs":257,"Frees":15},{"Size":352,"Mallocs":531,"Frees":284},{"Size":384,"Mallocs":63,"Frees":7},{"Size":416,"Mallocs":189,"Frees":66},{"Size":448,"Mallocs":60,"Frees":11},{"Size":480,"Mallocs":19,"Frees":6},{"Size":512,"Mallocs":72,"Frees":56},{"Size":576,"Mallocs":157,"Frees":61},{"Size":640,"Mallocs":27,"Frees":14},{"Size":704,"Mallocs":159,"Frees":146},{"Size":768,"Mallocs":182,"Frees":179},{"Size":896,"Mallocs":82,"Frees":43},{"Size":1024,"Mallocs":108,"Frees":39},{"Size":1152,"Mallocs":245,"Frees":27},{"Size":1280,"Mallocs":49,"Frees":32},{"Size":1408,"Mallocs":70,"Frees":37},{"Size":1536,"Mallocs":61,"Frees":31},{"Size":1664,"Mallocs":29,"Frees":17},{"Size":2048,"Mallocs":102,"Frees":74},{"Size":2304,"Mallocs":27,"Frees":15},{"Size":2560,"Mallocs":17,"Frees":14},{"Size":2816,"Mallocs":92,"Frees":90},{"Size":3072,"Mallocs":4,"Frees":4},{"Size":3328,"Mallocs":20,"Frees":12},{"Size":4096,"Mallocs":85,"Frees":63},{"Size":4608,"Mallocs":25,"Frees":10},{"Size":5376,"Mallocs":50,"Frees":7},{"Size":6144,"Mallocs":13,"Frees":2},{"Size":6400,"Mallocs":0,"Frees":0},{"Size":6656,"Mallocs":2,"Frees":0},{"Size":6912,"Mallocs":0,"Frees":0},{"Size":8192,"Mallocs":31,"Frees":1},{"Size":8448,"Mallocs":0,"Frees":0},{"Size":8704,"Mallocs":1,"Frees":1},{"Size":9472,"Mallocs":0,"Frees":0},{"Size":10496,"Mallocs":0,"Frees":0},{"Size":12288,"Mallocs":5,"Frees":1},{"Size":13568,"Mallocs":0,"Frees":0},{"Size":14080,"Mallocs":0,"Frees":0},{"Size":16384,"Mallocs":4,"Frees":0},{"Size":16640,"Mallocs":0,"Frees":0},{"Size":17664,"Mallocs":0,"Frees":0}]},
"num_enabled_tasks": 5,
"num_subscriptions": 6,
"num_tasks": 5,
"product": "kapacitor",
"server_id": "81e38258-5f18-4201-8d47-db603bfceeff",
"version": "1.1.0~rc2"
}
`
