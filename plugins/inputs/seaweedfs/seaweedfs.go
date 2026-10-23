//go:generate ../../../tools/readme_config_includer/generator
package seaweedfs

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultResponseTimeout = 5 * time.Second

	// ecDataShards is the number of data shards in an erasure-coded volume set.
	// SeaweedFS uses the same constant to translate the number of EC shards a
	// volume server holds into occupied volume slots, so the free-slot count
	// computed here matches the view of the master.
	ecDataShards = 10

	// defaultCollection is the name reported for volumes created without a
	// collection. An empty tag value would be dropped by Telegraf, so the
	// conventional name is used instead.
	defaultCollection = "default"
)

// SeaweedFS reads volume, disk and collection metrics from the SeaweedFS master
// and volume servers.
type SeaweedFS struct {
	Servers         []string        `toml:"servers"`
	VolumeServers   []string        `toml:"volume_servers"`
	DetailedVolumes bool            `toml:"detailed_volumes"`
	ResponseTimeout config.Duration `toml:"response_timeout"`

	tls.ClientConfig

	Log        telegraf.Logger `toml:"-"`
	httpClient *http.Client
}

// clusterStatus is the reply of the "/cluster/status" endpoint of a master.
type clusterStatus struct {
	IsLeader    bool   `json:"IsLeader"`
	Leader      string `json:"Leader"`
	MaxVolumeID int64  `json:"MaxVolumeId"`
}

// dirStatus is the reply of the "/dir/status" endpoint of a master.
type dirStatus struct {
	Topology topologyInfo `json:"Topology"`
	Version  string       `json:"Version"`
}

type topologyInfo struct {
	Max         int64            `json:"Max"`
	Free        int64            `json:"Free"`
	DataCenters []dataCenterInfo `json:"DataCenters"`
	Layouts     []volumeLayout   `json:"Layouts"`
}

type dataCenterInfo struct {
	ID    string     `json:"Id"`
	Racks []rackInfo `json:"Racks"`
}

type rackInfo struct {
	ID        string         `json:"Id"`
	DataNodes []dataNodeInfo `json:"DataNodes"`
}

type dataNodeInfo struct {
	URL       string `json:"Url"`
	PublicURL string `json:"PublicUrl"`
	Volumes   int64  `json:"Volumes"`
	ECShards  int64  `json:"EcShards"`
	Max       int64  `json:"Max"`
	VolumeIDs string `json:"VolumeIds"`
}

type volumeLayout struct {
	Replication string  `json:"replication"`
	TTL         string  `json:"ttl"`
	Writables   []int64 `json:"writables"`
	Collection  string  `json:"collection"`
	DiskType    string  `json:"diskType"`
}

// volumeServerStatus is the reply of the "/status" endpoint of a volume server.
type volumeServerStatus struct {
	Version      string       `json:"Version"`
	DiskStatuses []diskStatus `json:"DiskStatuses"`
	Volumes      []volumeInfo `json:"Volumes"`
}

type diskStatus struct {
	Dir         string  `json:"dir"`
	All         uint64  `json:"all"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	PercentFree float64 `json:"percent_free"`
	PercentUsed float64 `json:"percent_used"`
	DiskType    string  `json:"disk_type"`
	Error       string  `json:"error"`
}

type volumeInfo struct {
	Collection        string `json:"Collection"`
	DiskType          string `json:"DiskType"`
	Size              uint64 `json:"Size"`
	DeletedByteCount  uint64 `json:"DeletedByteCount"`
	ModifiedAtSecond  int64  `json:"ModifiedAtSecond"`
	ID                uint32 `json:"Id"`
	DiskID            uint32 `json:"DiskId"`
	CompactRevision   uint32 `json:"CompactRevision"`
	FileCount         uint32 `json:"FileCount"`
	DeleteCount       uint32 `json:"DeleteCount"`
	Version           uint32 `json:"Version"`
	ReadOnly          bool   `json:"ReadOnly"`
	ReadOnlyCanDelete bool   `json:"ReadOnlyCanDelete"`
}

// volumeTarget describes a single volume server to query.
type volumeTarget struct {
	address    string
	datacenter string
	rack       string
	maxSlots   int64
	volumes    int64
	ecShards   int64
	// known is true if the master reported this server, which is required to
	// know the volume slots of the server.
	known bool
}

// clusterInfo is the cluster-wide information read from the master topology.
type clusterInfo struct {
	targets   []volumeTarget
	index     map[string]volumeTarget
	writables map[string]int64
}

// volumeServerCounts are the counts reported for a volume server.
type volumeServerCounts struct {
	volumes  int64
	ecShards int64
	disks    int64
}

// collectionStats aggregates all volumes of one collection.
type collectionStats struct {
	volumeCount         int64
	readonlyVolumeCount int64
	sizeBytes           int64
	fileCount           int64
	deleteCount         int64
	deletedBytes        int64
}

// add accumulates the counters of a volume.
func (c *collectionStats) add(volume volumeInfo) {
	c.volumeCount++
	if volume.ReadOnly {
		c.readonlyVolumeCount++
	}
	c.sizeBytes += int64(volume.Size)
	c.fileCount += int64(volume.FileCount)
	c.deleteCount += int64(volume.DeleteCount)
	c.deletedBytes += int64(volume.DeletedByteCount)
}

// merge adds the counters of another collection statistic.
func (c *collectionStats) merge(other *collectionStats) {
	c.volumeCount += other.volumeCount
	c.readonlyVolumeCount += other.readonlyVolumeCount
	c.sizeBytes += other.sizeBytes
	c.fileCount += other.fileCount
	c.deleteCount += other.deleteCount
	c.deletedBytes += other.deletedBytes
}

// SampleConfig returns the default configuration of the plugin.
func (*SeaweedFS) SampleConfig() string {
	return sampleConfig
}

// Init validates the configuration and prepares the HTTP client.
func (s *SeaweedFS) Init() error {
	if len(s.Servers) == 0 {
		return errors.New("no master server configured in 'servers'")
	}
	for i, server := range s.Servers {
		if strings.TrimSpace(server) == "" {
			return fmt.Errorf("master server %d is empty", i)
		}
	}

	if s.ResponseTimeout <= 0 {
		s.ResponseTimeout = config.Duration(defaultResponseTimeout)
	}

	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	s.httpClient = &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   time.Duration(s.ResponseTimeout),
	}

	return nil
}

// Gather collects the metrics of the cluster.
func (s *SeaweedFS) Gather(acc telegraf.Accumulator) error {
	master := s.gatherMasters(acc)
	if master == "" {
		return nil
	}

	var status dirStatus
	if err := s.getJSON(master, "/dir/status", &status); err != nil {
		acc.AddError(fmt.Errorf("reading topology from master %q failed: %w", master, err))
		return nil
	}

	info := gatherCluster(acc, master, &status.Topology)
	targets := s.selectVolumeServers(info)

	// Every volume server is queried in parallel, each gathering the statistics
	// of the collections it holds. Merging happens afterwards, so no state is
	// shared between the goroutines.
	results := make([]map[string]*collectionStats, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(idx int, target volumeTarget) {
			defer wg.Done()
			results[idx] = s.gatherVolumeServer(acc, target)
		}(i, target)
	}
	wg.Wait()

	emitCollections(acc, mergeCollections(results), info.writables)

	return nil
}

// gatherMasters queries the leader state of every configured master and adds a
// "seaweedfs_master" metric for each of them. It returns the address of the
// master used for the cluster-wide queries, which is the first leader found.
func (s *SeaweedFS) gatherMasters(acc telegraf.Accumulator) string {
	statuses := make([]*clusterStatus, len(s.Servers))

	var wg sync.WaitGroup
	for i, server := range s.Servers {
		wg.Add(1)
		go func(idx int, address string) {
			defer wg.Done()

			var status clusterStatus
			if err := s.getJSON(address, "/cluster/status", &status); err != nil {
				acc.AddError(fmt.Errorf("querying master %q failed: %w", address, err))
				return
			}
			statuses[idx] = &status

			acc.AddFields(
				"seaweedfs_master",
				map[string]any{
					"is_leader":     status.IsLeader,
					"max_volume_id": status.MaxVolumeID,
				},
				map[string]string{"server": normalizeURL(address)},
			)
		}(i, server)
	}
	wg.Wait()

	// Prefer a leader for the cluster-wide queries. Non-leaders proxy those
	// requests to the leader anyway, but going there directly avoids a
	// redirect and keeps working when proxying is unavailable.
	for i, status := range statuses {
		if status != nil && status.IsLeader {
			return normalizeURL(s.Servers[i])
		}
	}
	for i, status := range statuses {
		if status != nil {
			return normalizeURL(s.Servers[i])
		}
	}

	return ""
}

// gatherCluster adds the cluster-wide metrics and a metric for every volume
// server known to the master.
func gatherCluster(
	acc telegraf.Accumulator,
	master string,
	topology *topologyInfo,
) clusterInfo {
	writables := make(map[string]int64, len(topology.Layouts))
	for _, layout := range topology.Layouts {
		writables[normalizeCollection(layout.Collection)] += int64(len(layout.Writables))
	}

	var volumeCount, ecShardCount, serverCount int64
	var dataCenterCount, rackCount int64
	targets := make([]volumeTarget, 0, len(topology.DataCenters))
	index := make(map[string]volumeTarget)

	for _, datacenter := range topology.DataCenters {
		dataCenterCount++
		for _, rack := range datacenter.Racks {
			rackCount++
			for _, node := range rack.DataNodes {
				address := normalizeURL(node.URL)
				if _, found := index[address]; found {
					continue
				}

				target := volumeTarget{
					address:    address,
					datacenter: datacenter.ID,
					rack:       rack.ID,
					maxSlots:   node.Max,
					volumes:    node.Volumes,
					ecShards:   node.ECShards,
					known:      true,
				}
				targets = append(targets, target)
				index[address] = target

				serverCount++
				volumeCount += node.Volumes
				ecShardCount += node.ECShards
			}
		}
	}

	acc.AddFields(
		"seaweedfs_cluster",
		map[string]any{
			"volume_count":        volumeCount,
			"ec_shard_count":      ecShardCount,
			"max_volume_slots":    topology.Max,
			"free_volume_slots":   topology.Free,
			"volume_server_count": serverCount,
			"data_center_count":   dataCenterCount,
			"rack_count":          rackCount,
			"volume_layout_count": int64(len(topology.Layouts)),
		},
		map[string]string{"server": master},
	)

	return clusterInfo{targets: targets, index: index, writables: writables}
}

// selectVolumeServers returns the volume servers to query. Without an explicit
// list all servers discovered via the master are used, otherwise the configured
// list wins and is enriched with the details known to the master.
func (s *SeaweedFS) selectVolumeServers(info clusterInfo) []volumeTarget {
	if len(s.VolumeServers) == 0 {
		return info.targets
	}

	targets := make([]volumeTarget, 0, len(s.VolumeServers))
	seen := make(map[string]struct{}, len(s.VolumeServers))
	for _, server := range s.VolumeServers {
		address := normalizeURL(server)
		if _, found := seen[address]; found {
			continue
		}
		seen[address] = struct{}{}

		if target, found := info.index[address]; found {
			targets = append(targets, target)
			continue
		}
		targets = append(targets, volumeTarget{address: address})
	}

	return targets
}

// gatherVolumeServer adds the disk, volume and collection metrics of a single
// volume server and returns the collection statistics of that server.
func (s *SeaweedFS) gatherVolumeServer(
	acc telegraf.Accumulator,
	target volumeTarget,
) map[string]*collectionStats {
	tags := map[string]string{"server": target.address}
	if target.datacenter != "" {
		tags["datacenter"] = target.datacenter
	}
	if target.rack != "" {
		tags["rack"] = target.rack
	}

	var status volumeServerStatus
	if err := s.getJSON(target.address, "/status", &status); err != nil {
		acc.AddError(fmt.Errorf("querying volume server %q failed: %w", target.address, err))
		// Without the master topology nothing is known about the server, so
		// only a reachable server is reported.
		if target.known {
			counts := volumeServerCounts{volumes: target.volumes, ecShards: target.ecShards}
			addVolumeServer(acc, target, tags, counts)
		}
		return nil
	}

	// The master is the authoritative source for the volume count. For a
	// configured volume server the master does not know, fall back to the
	// inventory reported by the server itself.
	volumeCount := target.volumes
	if !target.known {
		volumeCount = int64(len(status.Volumes))
	}
	counts := volumeServerCounts{
		volumes:  volumeCount,
		ecShards: target.ecShards,
		disks:    int64(len(status.DiskStatuses)),
	}
	addVolumeServer(acc, target, tags, counts)

	// The disk ID of a volume is the index of the disk location on the volume
	// server, which is the order the disk statuses are reported in.
	volumesPerDisk := make([]int64, len(status.DiskStatuses))
	for _, volume := range status.Volumes {
		if int(volume.DiskID) < len(volumesPerDisk) {
			volumesPerDisk[volume.DiskID]++
		}
	}

	for i, disk := range status.DiskStatuses {
		diskTags := map[string]string{
			"server": target.address,
			"path":   disk.Dir,
		}
		if disk.DiskType != "" {
			diskTags["disk_type"] = disk.DiskType
		}
		acc.AddFields(
			"seaweedfs_disk",
			map[string]any{
				"total_bytes":  int64(disk.All),
				"used_bytes":   int64(disk.Used),
				"free_bytes":   int64(disk.Free),
				"used_percent": disk.PercentUsed,
				"free_percent": disk.PercentFree,
				"volume_count": volumesPerDisk[i],
			},
			diskTags,
		)
		if disk.Error != "" && s.Log != nil {
			s.Log.Warnf("Volume server %q reports an error for %q: %s", target.address, disk.Dir, disk.Error)
		}
	}

	collections := make(map[string]*collectionStats)
	for _, volume := range status.Volumes {
		collection := normalizeCollection(volume.Collection)
		stats, found := collections[collection]
		if !found {
			stats = &collectionStats{}
			collections[collection] = stats
		}
		stats.add(volume)

		if s.DetailedVolumes {
			acc.AddFields(
				"seaweedfs_volume",
				map[string]any{
					"size_bytes":    int64(volume.Size),
					"file_count":    int64(volume.FileCount),
					"delete_count":  int64(volume.DeleteCount),
					"deleted_bytes": int64(volume.DeletedByteCount),
					"read_only":     volume.ReadOnly,
					"version":       int64(volume.Version),
					"disk_id":       int64(volume.DiskID),
				},
				map[string]string{
					"server":     target.address,
					"collection": collection,
					"volume_id":  strconv.FormatUint(uint64(volume.ID), 10),
				},
			)
		}
	}

	return collections
}

// addVolumeServer adds the metric describing a volume server. The volume slots
// are only known if the server was reported by the master.
func addVolumeServer(
	acc telegraf.Accumulator,
	target volumeTarget,
	tags map[string]string,
	counts volumeServerCounts,
) {
	fields := map[string]any{
		"volume_count":   counts.volumes,
		"ec_shard_count": counts.ecShards,
		"disk_count":     counts.disks,
	}
	if target.known {
		fields["max_volume_slots"] = target.maxSlots
		fields["free_volume_slots"] = freeVolumeSlots(target.maxSlots, counts.volumes, counts.ecShards)
	}

	acc.AddFields("seaweedfs_volume_server", fields, tags)
}

// mergeCollections combines the collection statistics of all volume servers.
func mergeCollections(results []map[string]*collectionStats) map[string]*collectionStats {
	collections := make(map[string]*collectionStats)
	for _, result := range results {
		for collection, stats := range result {
			merged, found := collections[collection]
			if !found {
				collections[collection] = stats
				continue
			}
			merged.merge(stats)
		}
	}

	return collections
}

// emitCollections adds one metric per collection.
func emitCollections(
	acc telegraf.Accumulator,
	collections map[string]*collectionStats,
	writables map[string]int64,
) {
	for collection, stats := range collections {
		acc.AddFields(
			"seaweedfs_collection",
			map[string]any{
				"volume_count":          stats.volumeCount,
				"readonly_volume_count": stats.readonlyVolumeCount,
				"writable_volume_count": writables[collection],
				"size_bytes":            stats.sizeBytes,
				"file_count":            stats.fileCount,
				"delete_count":          stats.deleteCount,
				"deleted_bytes":         stats.deletedBytes,
			},
			map[string]string{"collection": collection},
		)
	}
}

// getJSON queries an API endpoint and decodes the JSON reply into target.
func (s *SeaweedFS) getJSON(address, path string, target any) error {
	requestURL := normalizeURL(address) + path

	request, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(response.Body, 512))
		if err != nil {
			return fmt.Errorf("%s returned HTTP status %d", requestURL, response.StatusCode)
		}
		detail := strings.TrimSpace(string(body))
		return fmt.Errorf("%s returned HTTP status %d: %s", requestURL, response.StatusCode, detail)
	}

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding %s failed: %w", requestURL, err)
	}

	return nil
}

// normalizeURL adds the http scheme if it is missing, as SeaweedFS reports
// plain "host:port" addresses, and removes any trailing slash.
func normalizeURL(address string) string {
	address = strings.TrimSpace(address)
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	return strings.TrimRight(address, "/")
}

// normalizeCollection returns the collection name to report for a volume.
func normalizeCollection(collection string) string {
	if collection == "" {
		return defaultCollection
	}
	return collection
}

// freeVolumeSlots calculates the number of unused volume slots of a volume
// server the same way the master does.
func freeVolumeSlots(maxSlots, volumes, ecShards int64) int64 {
	used := volumes + (ecShards+ecDataShards-1)/ecDataShards
	if maxSlots <= used {
		return 0
	}
	return maxSlots - used
}

func init() {
	inputs.Add("seaweedfs", func() telegraf.Input {
		return &SeaweedFS{}
	})
}
