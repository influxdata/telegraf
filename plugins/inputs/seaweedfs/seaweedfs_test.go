package seaweedfs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

const (
	// volumeServerAddress is the address the master reports for the volume
	// server in the recorded topology. It is replaced by the test server
	// address so the volume server can be discovered and reached.
	volumeServerAddress = "seaweedfs-volume-0.seaweedfs-volume.s3:8080"

	clusterStatusPath = "/cluster/status"
	dirStatusPath     = "/dir/status"
	volumeStatusPath  = "/status"

	nonLeaderPath = "/nonleader/cluster/status"
)

// testCluster serves the recorded SeaweedFS API replies.
type testCluster struct {
	server *httptest.Server

	clusterStatus string
	dirStatus     string
	volumeStatus  string

	failClusterStatus bool
	failDirStatus     bool
	failVolumeStatus  bool
	invalidJSON       bool
}

func newTestCluster(t *testing.T) *testCluster {
	t.Helper()

	cluster := &testCluster{
		clusterStatus: readFixture(t, "testdata/cluster_status.json"),
		dirStatus:     readFixture(t, "testdata/dir_status.json"),
		volumeStatus:  readFixture(t, "testdata/volume_status.json"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(clusterStatusPath, func(w http.ResponseWriter, _ *http.Request) {
		if cluster.failClusterStatus {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		cluster.writeJSON(w, cluster.clusterStatus)
	})
	mux.HandleFunc(dirStatusPath, func(w http.ResponseWriter, _ *http.Request) {
		if cluster.failDirStatus {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		cluster.writeJSON(w, cluster.dirStatus)
	})
	mux.HandleFunc(volumeStatusPath, func(w http.ResponseWriter, _ *http.Request) {
		if cluster.failVolumeStatus {
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
		cluster.writeJSON(w, cluster.volumeStatus)
	})
	// A master answering the leader query as a non-leader does.
	mux.HandleFunc(nonLeaderPath, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"IsLeader":false,"Leader":"other:9333.19333","MaxVolumeId":14}`)
	})

	cluster.server = httptest.NewServer(mux)
	t.Cleanup(cluster.server.Close)

	// Point the recorded topology at the test server so that the volume server
	// is discovered and reachable.
	cluster.dirStatus = strings.ReplaceAll(cluster.dirStatus, volumeServerAddress, cluster.address())

	return cluster
}

func (c *testCluster) writeJSON(w http.ResponseWriter, fixture string) {
	if c.invalidJSON {
		fmt.Fprint(w, "{not json")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, fixture)
}

func (c *testCluster) address() string {
	return strings.TrimPrefix(c.server.URL, "http://")
}

func readFixture(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(content)
}

// TestInit tests the configuration validation and the defaults.
func TestInit(t *testing.T) {
	t.Run("no servers", func(t *testing.T) {
		plugin := &SeaweedFS{Log: testutil.Logger{}}
		require.ErrorContains(t, plugin.Init(), "no master server configured")
	})

	t.Run("empty server", func(t *testing.T) {
		plugin := &SeaweedFS{Servers: []string{" "}, Log: testutil.Logger{}}
		require.ErrorContains(t, plugin.Init(), "is empty")
	})

	t.Run("defaults", func(t *testing.T) {
		plugin := &SeaweedFS{Servers: []string{"localhost:9333"}, Log: testutil.Logger{}}
		require.NoError(t, plugin.Init())
		require.Equal(t, config.Duration(defaultResponseTimeout), plugin.Timeout)
		require.Equal(t, defaultResponseTimeout, plugin.httpClient.Timeout)
	})

	t.Run("timeout", func(t *testing.T) {
		plugin := &SeaweedFS{
			Servers: []string{"localhost:9333"},
			HTTPClientConfig: httpconfig.HTTPClientConfig{
				Timeout: config.Duration(2 * time.Second),
			},
			Log: testutil.Logger{},
		}
		require.NoError(t, plugin.Init())
		require.Equal(t, 2*time.Second, plugin.httpClient.Timeout)
	})

	t.Run("proxy", func(t *testing.T) {
		// The proxy is applied by the common HTTP client, which is the reason
		// for using it instead of a hand-rolled transport.
		plugin := &SeaweedFS{
			Servers: []string{"localhost:9333"},
			HTTPClientConfig: httpconfig.HTTPClientConfig{
				TransportConfig: httpconfig.TransportConfig{
					HTTPProxy: proxy.HTTPProxy{HTTPProxyURL: "http://proxy.local:3128"},
				},
			},
			Log: testutil.Logger{},
		}
		require.NoError(t, plugin.Init())

		transport, ok := plugin.httpClient.Transport.(*http.Transport)
		require.True(t, ok)
		require.NotNil(t, transport.Proxy)
	})
}

// TestGather tests the metrics of a fully reachable cluster.
func TestGather(t *testing.T) {
	cluster := newTestCluster(t)

	plugin := &SeaweedFS{
		Servers: []string{cluster.server.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	// Cluster-wide topology, read from the leader
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_cluster",
		map[string]any{
			"max_volume_id":       int64(14),
			"ec_shard_count":      int64(0),
			"free_volume_slots":   int64(21),
			"volume_server_count": int64(1),
			"data_center_count":   int64(1),
			"rack_count":          int64(1),
			"volume_layout_count": int64(2),
		},
		map[string]string{"server": cluster.server.URL},
	)

	// Volume server as reported by the master
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_volume_server",
		map[string]any{
			"ec_shard_count":    int64(0),
			"free_volume_slots": int64(21),
			"disk_count":        int64(2),
		},
		map[string]string{
			"server":     cluster.server.URL,
			"datacenter": "DefaultDataCenter",
			"rack":       "DefaultRack",
		},
	)

	// Per-disk volume counts, which the Prometheus endpoint does not expose
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_disk",
		map[string]any{"volume_count": int64(4)},
		map[string]string{
			"server": cluster.server.URL,
			"path":   "/data1",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_disk",
		map[string]any{"volume_count": int64(0)},
		map[string]string{
			"server":    cluster.server.URL,
			"path":      "/data2",
			"disk_type": "ssd",
		},
	)

	// The needle counts per collection, tagged with the cluster they belong to
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{
			"file_count":   int64(44),
			"delete_count": int64(0),
		},
		map[string]string{
			"server":     cluster.server.URL,
			"collection": "default",
		},
	)

	// Per-volume metrics are only emitted on request
	acc.AssertDoesNotContainMeasurement(t, "seaweedfs_volume")
	// cluster, volume server, two disks and one collection
	require.Len(t, acc.Metrics, 5)
}

// TestGatherDetailedVolumes tests the opt-in per-volume metrics.
func TestGatherDetailedVolumes(t *testing.T) {
	cluster := newTestCluster(t)

	plugin := &SeaweedFS{
		Servers:         []string{cluster.server.URL},
		DetailedVolumes: true,
		Log:             testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	acc.AssertContainsTaggedFields(t,
		"seaweedfs_volume",
		map[string]any{
			"size_bytes":    int64(3828300),
			"file_count":    int64(6),
			"delete_count":  int64(0),
			"deleted_bytes": int64(0),
			"read_only":     false,
			"version":       int64(3),
			"disk_id":       int64(0),
		},
		map[string]string{
			"server":     cluster.server.URL,
			"collection": "default",
			"volume_id":  "1",
		},
	)

	// five cluster-level metrics plus one metric per volume
	require.Len(t, acc.Metrics, 9)
}

// TestGatherMultipleMasters tests that the cluster metrics are read from the
// leader even when a non-leader is configured first.
func TestGatherMultipleMasters(t *testing.T) {
	cluster := newTestCluster(t)

	plugin := &SeaweedFS{
		Servers: []string{cluster.server.URL + "/nonleader", cluster.server.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	// The cluster metric is tagged with the leader
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_cluster",
		map[string]any{
			"max_volume_id":       int64(14),
			"ec_shard_count":      int64(0),
			"free_volume_slots":   int64(21),
			"volume_server_count": int64(1),
			"data_center_count":   int64(1),
			"rack_count":          int64(1),
			"volume_layout_count": int64(2),
		},
		map[string]string{"server": cluster.server.URL},
	)
}

// TestGatherExplicitVolumeServers tests that a configured volume server list is
// used instead of the servers discovered via the master.
func TestGatherExplicitVolumeServers(t *testing.T) {
	cluster := newTestCluster(t)

	const volumeServerStatus = `{
  "DiskStatuses": [
    {"dir": "/data9", "all": 100, "used": 40, "free": 60, "percent_free": 60, "percent_used": 40}
  ],
  "Volumes": [
    {"Collection": "lab", "Size": 10, "Id": 1, "DiskId": 0, "FileCount": 2, "Version": 3}
  ]
}`

	mux := http.NewServeMux()
	mux.HandleFunc(volumeStatusPath, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, volumeServerStatus)
	})
	second := httptest.NewServer(mux)
	defer second.Close()

	plugin := &SeaweedFS{
		Servers:       []string{cluster.server.URL},
		VolumeServers: []string{second.URL},
		Log:           testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	// The configured server is not part of the topology, so neither the
	// datacenter and rack tags nor the volume slots are known.
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_volume_server",
		map[string]any{
			"ec_shard_count": int64(0),
			"disk_count":     int64(1),
		},
		map[string]string{"server": second.URL},
	)

	acc.AssertContainsTaggedFields(t,
		"seaweedfs_disk",
		map[string]any{"volume_count": int64(1)},
		map[string]string{
			"server": second.URL,
			"path":   "/data9",
		},
	)

	// The discovered volume server is not queried, only the configured one
	require.Len(t, acc.Metrics, 4)
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{
			"file_count":   int64(2),
			"delete_count": int64(0),
		},
		map[string]string{
			"server":     cluster.server.URL,
			"collection": "lab",
		},
	)
}

// TestGatherReplication tests that the replicas of a volume are counted once in
// the collection statistics.
func TestGatherReplication(t *testing.T) {
	// Two volume servers holding the same two volumes, as replication "001"
	// does: the replicas report identical file and byte counts.
	const volumeInventory = `{
  "DiskStatuses": [{"dir": "/data1", "all": 100, "used": 10, "free": 90, "percent_free": 90, "percent_used": 10}],
  "Volumes": [
    {"Collection": "lab", "Size": 10, "Id": 1, "DiskId": 0, "FileCount": 5, "Version": 3},
    {"Collection": "lab", "Size": 20, "Id": 2, "DiskId": 0, "FileCount": 6, "Version": 3}
  ]
}`

	newVolumeServer := func() *httptest.Server {
		mux := http.NewServeMux()
		mux.HandleFunc(volumeStatusPath, func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, volumeInventory)
		})
		return httptest.NewServer(mux)
	}
	first := newVolumeServer()
	defer first.Close()
	second := newVolumeServer()
	defer second.Close()

	topology := fmt.Sprintf(`{
  "Topology": {
    "Max": 100,
    "Free": 96,
    "DataCenters": [{"Id": "dc1", "Racks": [{"Id": "rack1", "DataNodes": [
      {"Url": %q, "Volumes": 2, "Max": 50},
      {"Url": %q, "Volumes": 2, "Max": 50}
    ]}]}],
    "Layouts": []
  },
  "Version": "4.48"
}`, strings.TrimPrefix(first.URL, "http://"), strings.TrimPrefix(second.URL, "http://"))

	mux := http.NewServeMux()
	mux.HandleFunc(clusterStatusPath, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"IsLeader":true,"Leader":"self","MaxVolumeId":2}`)
	})
	mux.HandleFunc(dirStatusPath, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, topology)
	})
	master := httptest.NewServer(mux)
	defer master.Close()

	plugin := &SeaweedFS{
		Servers: []string{master.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	// 5 + 6 files in total, not 2*(5+6): the same volume ID is one volume
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{
			"file_count":   int64(11),
			"delete_count": int64(0),
		},
		map[string]string{
			"server":     master.URL,
			"collection": "lab",
		},
	)

	// Both replicas are reachable, so both report their slots
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_volume_server",
		map[string]any{
			"ec_shard_count":    int64(0),
			"free_volume_slots": int64(48),
			"disk_count":        int64(1),
		},
		map[string]string{
			"server":     first.URL,
			"datacenter": "dc1",
			"rack":       "rack1",
		},
	)
}

// TestGatherVolumeServerError tests that a failing volume server is reported
// but the cluster metrics are still gathered.
func TestGatherVolumeServerError(t *testing.T) {
	cluster := newTestCluster(t)
	cluster.failVolumeStatus = true

	plugin := &SeaweedFS{
		Servers: []string{cluster.server.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "querying volume server")

	// The volume server metric is taken from the master topology and therefore
	// still available, only the disk details are missing
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_volume_server",
		map[string]any{
			"ec_shard_count":    int64(0),
			"free_volume_slots": int64(21),
			"disk_count":        int64(0),
		},
		map[string]string{
			"server":     cluster.server.URL,
			"datacenter": "DefaultDataCenter",
			"rack":       "DefaultRack",
		},
	)
	acc.AssertDoesNotContainMeasurement(t, "seaweedfs_disk")
	acc.AssertDoesNotContainMeasurement(t, "seaweedfs_collection")
}

// TestGatherMasterError tests that no metrics are emitted if the master cannot
// be reached.
func TestGatherMasterError(t *testing.T) {
	cluster := newTestCluster(t)
	cluster.failClusterStatus = true

	plugin := &SeaweedFS{
		Servers: []string{cluster.server.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "querying master")
	require.Empty(t, acc.Metrics)
}

// TestGatherTopologyError tests that nothing is reported if the topology cannot
// be read.
func TestGatherTopologyError(t *testing.T) {
	cluster := newTestCluster(t)
	cluster.failDirStatus = true

	plugin := &SeaweedFS{
		Servers: []string{cluster.server.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "reading topology")
	require.Empty(t, acc.Metrics)
}

// TestGatherInvalidJSON tests that a malformed reply is reported as an error.
func TestGatherInvalidJSON(t *testing.T) {
	cluster := newTestCluster(t)
	cluster.invalidJSON = true

	plugin := &SeaweedFS{
		Servers: []string{cluster.server.URL},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.Errors[0], "decoding")
	require.Empty(t, acc.Metrics)
}

// TestGatherAddressWithoutScheme tests that plain "host:port" addresses as
// reported by SeaweedFS are handled.
func TestGatherAddressWithoutScheme(t *testing.T) {
	cluster := newTestCluster(t)

	plugin := &SeaweedFS{
		Servers: []string{cluster.address()},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	acc.AssertContainsTaggedFields(t,
		"seaweedfs_cluster",
		map[string]any{
			"max_volume_id":       int64(14),
			"ec_shard_count":      int64(0),
			"free_volume_slots":   int64(21),
			"volume_server_count": int64(1),
			"data_center_count":   int64(1),
			"rack_count":          int64(1),
			"volume_layout_count": int64(2),
		},
		map[string]string{"server": cluster.server.URL},
	)
}

// TestNormalizeURL tests the address normalization.
func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		address  string
		expected string
	}{
		{"localhost:9333", "http://localhost:9333"},
		{"http://localhost:9333", "http://localhost:9333"},
		{"https://localhost:9333/", "https://localhost:9333"},
		{" http://localhost:9333 ", "http://localhost:9333"},
	}
	for _, test := range tests {
		require.Equal(t, test.expected, normalizeURL(test.address), test.address)
	}
}

// TestNormalizeCollection tests the collection name normalization.
func TestNormalizeCollection(t *testing.T) {
	require.Equal(t, "default", normalizeCollection(""))
	require.Equal(t, "lab", normalizeCollection("lab"))
}

// TestFreeVolumeSlots tests the free volume slot calculation.
func TestFreeVolumeSlots(t *testing.T) {
	tests := []struct {
		name     string
		maxSlots int64
		volumes  int64
		ecShards int64
		expected int64
	}{
		{name: "empty", maxSlots: 35, volumes: 0, ecShards: 0, expected: 35},
		{name: "volumes", maxSlots: 35, volumes: 14, ecShards: 0, expected: 21},
		{name: "ec shards round up", maxSlots: 35, volumes: 0, ecShards: 11, expected: 33},
		{name: "overcommitted", maxSlots: 10, volumes: 12, ecShards: 0, expected: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, freeVolumeSlots(test.maxSlots, test.volumes, test.ecShards))
		})
	}
}

// TestMergeVolumes tests that replicas are collapsed into one volume.
func TestMergeVolumes(t *testing.T) {
	first := []volumeRecord{
		{collection: "lab", info: volumeInfo{ID: 1, FileCount: 5}},
		{collection: "lab", info: volumeInfo{ID: 2, FileCount: 6}},
	}
	second := []volumeRecord{
		{collection: "lab", info: volumeInfo{ID: 1, FileCount: 5}},
		{collection: "lab", info: volumeInfo{ID: 3, FileCount: 7}},
	}

	merged := mergeVolumes([][]volumeRecord{first, second, nil})
	require.Len(t, merged, 3)

	ids := make(map[uint32]struct{}, len(merged))
	for _, record := range merged {
		ids[record.info.ID] = struct{}{}
	}
	for _, id := range []uint32{1, 2, 3} {
		_, found := ids[id]
		require.True(t, found, "volume %d missing", id)
	}
}

// TestEmitCollections tests the aggregation of the collection counts.
func TestEmitCollections(t *testing.T) {
	records := []volumeRecord{
		{collection: "lab", info: volumeInfo{ID: 1, FileCount: 10, DeleteCount: 2}},
		{collection: "lab", info: volumeInfo{ID: 2, FileCount: 5, DeleteCount: 1}},
		{collection: "default", info: volumeInfo{ID: 3, FileCount: 7}},
	}

	var acc testutil.Accumulator
	emitCollections(&acc, "http://master:9333", records)

	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{"file_count": int64(15), "delete_count": int64(3)},
		map[string]string{"server": "http://master:9333", "collection": "lab"},
	)
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{"file_count": int64(7), "delete_count": int64(0)},
		map[string]string{"server": "http://master:9333", "collection": "default"},
	)
}

// TestSampleConfig tests that the sample configuration matches the plugin
// options. It fails on unknown or removed keys, e.g. a renamed option that was
// not updated in the sample.
func TestSampleConfig(t *testing.T) {
	cfg := config.NewConfig()
	require.NoError(t, cfg.LoadConfigData([]byte(sampleConfig), config.EmptySourcePath))
	require.Len(t, cfg.Inputs, 1)
}

// TestRegistration tests that the plugin is registered.
func TestRegistration(t *testing.T) {
	creator, found := inputs.Inputs["seaweedfs"]
	require.True(t, found)
	require.IsType(t, &SeaweedFS{}, creator())
}

var _ telegraf.Input = &SeaweedFS{}
