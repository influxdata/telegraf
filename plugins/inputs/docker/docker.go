//go:generate ../../../tools/readme_config_includer/generator
package docker

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	sizeRegex   = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[bB]?$`)
	sizeUnitMap = map[string]int64{
		"k": 1000,
		"m": 1000 * 1000,
		"g": 1000 * 1000 * 1000,
		"t": 1000 * 1000 * 1000 * 1000,
		"p": 1000 * 1000 * 1000 * 1000 * 1000,
	}
	now = time.Now
)

type Docker struct {
	Endpoint              string          `toml:"endpoint"`
	GatherServices        bool            `toml:"gather_services"`
	PerDeviceInclude      []string        `toml:"perdevice_include"`
	TotalInclude          []string        `toml:"total_include"`
	TagEnvironment        []string        `toml:"tag_env"`
	LabelInclude          []string        `toml:"docker_label_include"`
	LabelExclude          []string        `toml:"docker_label_exclude"`
	ContainerInclude      []string        `toml:"container_name_include"`
	ContainerExclude      []string        `toml:"container_name_exclude"`
	ContainerStateInclude []string        `toml:"container_state_include"`
	ContainerStateExclude []string        `toml:"container_state_exclude"`
	StorageObjects        []string        `toml:"storage_objects"`
	IncludeSourceTag      bool            `toml:"source_tag"`
	Timeout               config.Duration `toml:"timeout"`
	PodmanCacheTTL        config.Duration `toml:"podman_cache_ttl"`
	Log                   telegraf.Logger `toml:"-"`
	common_tls.ClientConfig

	client          *client.Client
	engineHost      string
	serverVersion   string
	isPodman        bool
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
	objectTypes     []types.DiskUsageObject

	// Stats cache for Podman CPU calculation
	statsCache      map[string]*cachedContainerStats
	statsCacheMutex sync.Mutex
}

type cachedContainerStats struct {
	stats     *container.StatsResponse
	timestamp time.Time
}

func (*Docker) SampleConfig() string {
	return sampleConfig
}

func (d *Docker) Init() error {
	// Defaults
	if d.Endpoint == "" {
		d.Endpoint = "unix:///var/run/docker.sock"
	}
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}

	// Check settings
	for _, include := range d.PerDeviceInclude {
		switch include {
		case "cpu", "network", "blkio":
		default:
			return fmt.Errorf("invalid 'perdevice_include' setting %q", include)
		}
	}
	for _, include := range d.TotalInclude {
		switch include {
		case "cpu", "network", "blkio":
		default:
			return fmt.Errorf("invalid 'total_include' setting %q", include)
		}
	}

	// Create storage query objects
	d.objectTypes = make([]types.DiskUsageObject, 0, len(d.StorageObjects))
	for _, object := range d.StorageObjects {
		switch object {
		case "container":
			d.objectTypes = append(d.objectTypes, types.ContainerObject)
		case "image":
			d.objectTypes = append(d.objectTypes, types.ImageObject)
		case "volume":
			d.objectTypes = append(d.objectTypes, types.VolumeObject)
		default:
			return fmt.Errorf("invalid storage object type: %q", object)
		}
	}

	// Create filters
	var err error
	d.labelFilter, err = filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return fmt.Errorf("creating label filter failed: %w", err)
	}
	d.containerFilter, err = filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return fmt.Errorf("creating container name filter failed: %w", err)
	}
	d.stateFilter, err = filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return fmt.Errorf("creating container state filter failed: %w", err)
	}

	return nil
}

func (d *Docker) Start(telegraf.Accumulator) error {
	// Create a new client, this does not connect
	switch d.Endpoint {
	case "ENV":
		c, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			return fmt.Errorf("creating client from environment failed: %w", err)
		}
		d.client = c
	default:
		tlsConfig, err := d.ClientConfig.TLSConfig()
		if err != nil {
			return fmt.Errorf("creating TLS configuration failed: %w", err)
		}

		options := []client.Opt{
			client.WithUserAgent("engine-api-cli-1.0"),
			client.WithAPIVersionNegotiation(),
			client.WithHost(d.Endpoint),
		}
		if tlsConfig != nil {
			httpClient := &http.Client{Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			}}
			options = append(options, client.WithHTTPClient(httpClient))
		}
		c, err := client.NewClientWithOpts(options...)
		if err != nil {
			return fmt.Errorf("creating client failed: %w", err)
		}
		d.client = c
	}

	// Use Ping to check connectivity - this is a lightweight check
	ctxPing, cancelPing := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancelPing()
	if _, err := d.client.Ping(ctxPing); err != nil {
		d.Stop()
		return &internal.StartupError{
			Err:   fmt.Errorf("failed to ping daemon: %w", err),
			Retry: client.IsErrConnectionFailed(err),
		}
	}

	// Check API version compatibility
	version, err := semver.NewVersion(d.client.ClientVersion())
	if err != nil {
		d.Stop()
		return fmt.Errorf("failed to parse client version: %w", err)
	}
	if version.LessThan(semver.New(1, 23, 0, "", "")) {
		return fmt.Errorf("unsupported API version (%s), upgrade to docker engine 1.12+", version)
	} else if version.LessThan(semver.New(1, 42, 0, "", "")) && len(d.objectTypes) > 0 {
		return fmt.Errorf("unsupported API version (%s) for disk usage, upgrade to docker engine 23.0+", version)
	}

	// Get info from docker daemon for Podman detection
	ctxInfo, cancelInfo := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancelInfo()

	info, err := d.client.Info(ctxInfo)
	if err != nil {
		d.Stop()
		return &internal.StartupError{
			Err:   fmt.Errorf("failed to get Docker info: %w", err),
			Retry: client.IsErrConnectionFailed(err),
		}
	}
	d.engineHost = info.Name
	d.serverVersion = info.ServerVersion
	d.isPodman = d.detectPodman(&info)

	// Initialize stats cache only for Podman to save memory for Docker users
	if d.isPodman {
		d.statsCache = make(map[string]*cachedContainerStats)
		msg := "Detected Podman engine (version: %s, name: %s), using stats caching for accurate CPU measurements"
		d.Log.Debugf(msg, info.ServerVersion, info.Name)
	}

	return nil
}

func (d *Docker) Stop() {
	// Close client connection if exists
	if d.client != nil {
		d.client.Close()
		d.client = nil
	}
}

func (d *Docker) Gather(acc telegraf.Accumulator) error {
	// General daemon info
	acc.AddError(d.gatherInfo(acc))

	// Swarm services
	if d.GatherServices {
		acc.AddError(d.gatherSwarmInfo(acc))
	}

	// List containers for detailed metrics
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Timeout))
	defer cancel()

	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.New("timeout retrieving container list")
	}
	if err != nil {
		return fmt.Errorf("listing containers failed: %w", err)
	}

	// Per-container metrics
	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, cntnr := range containers {
		go func(c container.Summary) {
			defer wg.Done()

			// Prepare the tags
			tags, err := d.gatherContainerInfo(acc, c)
			if err != nil {
				acc.AddError(err)
				return
			}
			if tags == nil {
				return
			}
			acc.AddError(d.gatherContainerStats(acc, tags, c.ID))
		}(cntnr)
	}
	wg.Wait()

	// Per disk/volume usage data
	if len(d.objectTypes) > 0 {
		acc.AddError(d.gatherDiskUsage(acc, types.DiskUsageOptions{Types: d.objectTypes}))
	}

	// Clean up stale cache entries for Podman
	if d.isPodman {
		d.cleanupStaleCache()
	}

	return nil
}

// detectPodman detects if we're connected to Podman by checking Docker info response.
// Uses a conservative approach prioritizing explicit indicators over heuristics.
func (d *Docker) detectPodman(info *system.Info) bool {
	sv := strings.ToLower(info.ServerVersion)
	name := strings.ToLower(info.Name)
	endpoint := strings.ToLower(d.Endpoint)

	// 1. Explicit Docker indicators (highest confidence)
	if strings.Contains(sv, "docker") || strings.Contains(name, "docker") ||
		strings.Contains(info.InitBinary, "docker") {
		return false
	}

	// 2. Explicit Podman indicators (highest confidence)
	if strings.Contains(sv, "podman") || strings.Contains(name, "podman") ||
		strings.Contains(endpoint, "podman") {
		return true
	}

	// 3. Exclude other known container runtimes
	if strings.Contains(name, "kubernetes") || strings.Contains(name, "containerd") ||
		strings.Contains(endpoint, "containerd") {
		return false
	}

	// 4. Podman heuristics - conservative approach
	// Common Podman patterns: crun runtime, localhost domains, short names, container sockets
	if info.InitBinary == "crun" ||
		strings.Contains(name, "localhost") ||
		strings.Contains(endpoint, "container.sock") ||
		(len(name) <= 4 && name != "") {
		return true
	}

	// 5. Default to Docker for safety
	return false
}

// fixPodmanCPUStats fixes Podman's CPU stats using cached previous stats
func (d *Docker) fixPodmanCPUStats(containerID string, current *container.StatsResponse) {
	now := time.Now()
	ttl := time.Duration(d.PodmanCacheTTL)

	// Single lock for read-check-update operation
	d.statsCacheMutex.Lock()
	defer d.statsCacheMutex.Unlock()

	if cached, exists := d.statsCache[containerID]; exists && cached != nil && cached.stats != nil {
		// Check if cached stats are recent enough
		age := now.Sub(cached.timestamp)
		if age <= ttl {
			// Use cached stats as PreCPUStats for accurate CPU calculation
			current.PreCPUStats = cached.stats.CPUStats
			d.Log.Tracef("Podman stats cache hit for container %s (age: %v)", hostnameFromID(containerID), age)
		} else {
			d.Log.Tracef("Podman stats cache expired for container %s (age: %v)", hostnameFromID(containerID), age)
		}
	} else {
		d.Log.Tracef("Podman stats cache miss for container %s (first collection)", hostnameFromID(containerID))
	}

	// Update cache with current stats (reuse timestamp)
	d.statsCache[containerID] = &cachedContainerStats{
		stats:     current,
		timestamp: now,
	}
}

// cleanupStaleCache removes expired entries from the stats cache
func (d *Docker) cleanupStaleCache() {
	d.statsCacheMutex.Lock()
	defer d.statsCacheMutex.Unlock()

	if len(d.statsCache) == 0 {
		return // Early exit if cache is empty
	}

	cutoff := time.Now().Add(-time.Duration(d.PodmanCacheTTL))
	expiredCount := 0

	for id, cached := range d.statsCache {
		if cached.timestamp.Before(cutoff) {
			delete(d.statsCache, id)
			expiredCount++
		}
	}

	d.Log.Tracef("Cleaned up %d expired entries from Podman stats cache", expiredCount)
}

func init() {
	inputs.Add("docker", func() telegraf.Input {
		return &Docker{
			// Keep the includes here to allow the user to override the setting
			// with an empty list
			PerDeviceInclude: []string{"cpu"},
			TotalInclude:     []string{"cpu", "blkio", "network"},
			Timeout:          config.Duration(5 * time.Second),
			PodmanCacheTTL:   config.Duration(60 * time.Second),
		}
	})
}
