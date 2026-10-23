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
		require.Equal(t, defaultResponseTimeout, time.Duration(plugin.ResponseTimeout))
		require.Equal(t, defaultResponseTimeout, plugin.httpClient.Timeout)
	})

	t.Run("response timeout", func(t *testing.T) {
		plugin := &SeaweedFS{
			Servers:         []string{"localhost:9333"},
			ResponseTimeout: config.Duration(2 * time.Second),
			Log:             testutil.Logger{},
		}
		require.NoError(t, plugin.Init())
		require.Equal(t, 2*time.Second, plugin.httpClient.Timeout)
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

	// Leader state of the master
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_master",
		map[string]any{
			"is_leader":     true,
			"max_volume_id": int64(14),
		},
		map[string]string{"server": cluster.server.URL},
	)

	// Cluster-wide topology, read from the leader
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_cluster",
		map[string]any{
			"volume_count":        int64(14),
			"ec_shard_count":      int64(0),
			"max_volume_slots":    int64(35),
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
			"volume_count":      int64(14),
			"ec_shard_count":    int64(0),
			"max_volume_slots":  int64(35),
			"free_volume_slots": int64(21),
			"disk_count":        int64(2),
		},
		map[string]string{
			"server":     cluster.server.URL,
			"datacenter": "DefaultDataCenter",
			"rack":       "DefaultRack",
		},
	)

	// Disk usage of the volume server
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_disk",
		map[string]any{
			"total_bytes":  int64(490164953088),
			"used_bytes":   int64(81732874240),
			"free_bytes":   int64(408432078848),
			"used_percent": 16.674564,
			"free_percent": 83.32543,
			"volume_count": int64(4),
		},
		map[string]string{
			"server": cluster.server.URL,
			"path":   "/data1",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_disk",
		map[string]any{
			"total_bytes":  int64(490164953088),
			"used_bytes":   int64(81732874240),
			"free_bytes":   int64(408432078848),
			"used_percent": 57.5,
			"free_percent": 42.5,
			"volume_count": int64(0),
		},
		map[string]string{
			"server":    cluster.server.URL,
			"path":      "/data2",
			"disk_type": "ssd",
		},
	)

	// Collections are aggregated, the number of writable volumes comes from the
	// volume layouts of the master
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{
			"volume_count":          int64(4),
			"readonly_volume_count": int64(0),
			"writable_volume_count": int64(7),
			"size_bytes":            int64(26456818),
			"file_count":            int64(44),
			"delete_count":          int64(0),
			"deleted_bytes":         int64(0),
		},
		map[string]string{"collection": "default"},
	)

	// Per-volume metrics are only emitted on request
	acc.AssertDoesNotContainMeasurement(t, "seaweedfs_volume")
	// master, cluster, volume server, two disks and one collection
	require.Len(t, acc.Metrics, 6)
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

	// six cluster-wide metrics plus one metric per volume
	require.Len(t, acc.Metrics, 10)
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

	// One metric per master with its own leader state
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_master",
		map[string]any{
			"is_leader":     false,
			"max_volume_id": int64(14),
		},
		map[string]string{"server": cluster.server.URL + "/nonleader"},
	)
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_master",
		map[string]any{
			"is_leader":     true,
			"max_volume_id": int64(14),
		},
		map[string]string{"server": cluster.server.URL},
	)

	// The cluster metrics are tagged with the leader
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_cluster",
		map[string]any{
			"volume_count":        int64(14),
			"ec_shard_count":      int64(0),
			"max_volume_slots":    int64(35),
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
	// datacenter and rack tags nor the volume slots are known. The volume count
	// is taken from the inventory of the server itself.
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_volume_server",
		map[string]any{
			"volume_count":   int64(1),
			"ec_shard_count": int64(0),
			"disk_count":     int64(1),
		},
		map[string]string{"server": second.URL},
	)

	acc.AssertContainsTaggedFields(t,
		"seaweedfs_disk",
		map[string]any{
			"total_bytes":  int64(100),
			"used_bytes":   int64(40),
			"free_bytes":   int64(60),
			"used_percent": 40.0,
			"free_percent": 60.0,
			"volume_count": int64(1),
		},
		map[string]string{
			"server": second.URL,
			"path":   "/data9",
		},
	)

	// The discovered volume server is not queried, only the configured one
	require.Len(t, acc.Metrics, 5)
	acc.AssertContainsTaggedFields(t,
		"seaweedfs_collection",
		map[string]any{
			"volume_count":          int64(1),
			"readonly_volume_count": int64(0),
			"writable_volume_count": int64(7),
			"size_bytes":            int64(10),
			"file_count":            int64(2),
			"delete_count":          int64(0),
			"deleted_bytes":         int64(0),
		},
		map[string]string{"collection": "lab"},
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
			"volume_count":      int64(14),
			"ec_shard_count":    int64(0),
			"max_volume_slots":  int64(35),
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

// TestGatherTopologyError tests that the leader state is reported even if the
// topology cannot be read.
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

	// Only the master metric was added
	require.Len(t, acc.Metrics, 1)
	require.Equal(t, "seaweedfs_master", acc.Metrics[0].Measurement)
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
		"seaweedfs_master",
		map[string]any{
			"is_leader":     true,
			"max_volume_id": int64(14),
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

// TestMergeCollections tests the merge of the per-server collection statistics.
func TestMergeCollections(t *testing.T) {
	first := map[string]*collectionStats{
		"lab": {volumeCount: 2, readonlyVolumeCount: 1, sizeBytes: 100, fileCount: 10},
	}
	second := map[string]*collectionStats{
		"lab":     {volumeCount: 3, sizeBytes: 50, fileCount: 5, deleteCount: 1},
		"default": {volumeCount: 1, sizeBytes: 7},
	}

	merged := mergeCollections([]map[string]*collectionStats{first, second, nil})
	require.Len(t, merged, 2)
	require.Equal(t, &collectionStats{
		volumeCount:         5,
		readonlyVolumeCount: 1,
		sizeBytes:           150,
		fileCount:           15,
		deleteCount:         1,
	}, merged["lab"])
	require.Equal(t, &collectionStats{volumeCount: 1, sizeBytes: 7}, merged["default"])
}

// TestCollectionStatsAdd tests the accumulation of a single volume.
func TestCollectionStatsAdd(t *testing.T) {
	stats := &collectionStats{}
	stats.add(volumeInfo{Size: 10, FileCount: 2, DeleteCount: 1, DeletedByteCount: 3})
	stats.add(volumeInfo{Size: 5, FileCount: 1, ReadOnly: true})

	require.Equal(t, int64(2), stats.volumeCount)
	require.Equal(t, int64(1), stats.readonlyVolumeCount)
	require.Equal(t, int64(15), stats.sizeBytes)
	require.Equal(t, int64(3), stats.fileCount)
	require.Equal(t, int64(1), stats.deleteCount)
	require.Equal(t, int64(3), stats.deletedBytes)
}

// TestRegistration tests that the plugin is registered.
func TestRegistration(t *testing.T) {
	creator, found := inputs.Inputs["seaweedfs"]
	require.True(t, found)
	require.IsType(t, &SeaweedFS{}, creator())
}

var _ telegraf.Input = &SeaweedFS{}
