package system

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Docker object
type Docker struct {
	Endpoint       string
	ContainerNames []string
	Timeout        internal.Duration

	client DockerClient
}

// DockerClient interface, useful for testing
type DockerClient interface {
	Info(ctx context.Context) (types.Info, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (io.ReadCloser, error)
}

// KB, MB, GB, TB, PB...human friendly
const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB
)

var (
	sizeRegex = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[bB]?$`)
)

var sampleConfig = `
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"
  ## Only collect metrics for these containers, collect all if empty
  container_names = []
  ## Timeout for docker list, info, and stats commands
  timeout = "5s"
`

// Description returns input description
func (d *Docker) Description() string {
	return "Read metrics about docker containers"
}

// SampleConfig prints sampleConfig
func (d *Docker) SampleConfig() string { return sampleConfig }

// Gather starts stats collection
func (d *Docker) Gather(acc telegraf.Accumulator) error {
	if d.client == nil {
		var c *client.Client
		var err error
		defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
		if d.Endpoint == "ENV" {
			c, err = client.NewEnvClient()
			if err != nil {
				return err
			}
		} else if d.Endpoint == "" {
			c, err = client.NewClient("unix:///var/run/docker.sock", "", nil, defaultHeaders)
			if err != nil {
				return err
			}
		} else {
			c, err = client.NewClient(d.Endpoint, "", nil, defaultHeaders)
			if err != nil {
				return err
			}
		}
		d.client = c
	}

	// Get daemon info
	err := d.gatherInfo(acc)
	if err != nil {
		fmt.Println(err.Error())
	}

	// List containers
	opts := types.ContainerListOptions{}
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()
	containers, err := d.client.ContainerList(ctx, opts)
	if err != nil {
		return err
	}

	// Get container data
	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {
		go func(c types.Container) {
			defer wg.Done()
			err := d.gatherContainer(c, acc)
			if err != nil {
				log.Printf("Error gathering container %s stats: %s\n",
					c.Names, err.Error())
			}
		}(container)
	}
	wg.Wait()

	return nil
}

func (d *Docker) gatherInfo(acc telegraf.Accumulator) error {
	// Init vars
	dataFields := make(map[string]interface{})
	metadataFields := make(map[string]interface{})
	now := time.Now()
	// Get info from docker daemon
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()
	info, err := d.client.Info(ctx)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"n_cpus":                  info.NCPU,
		"n_used_file_descriptors": info.NFd,
		"n_containers":            info.Containers,
		"n_images":                info.Images,
		"n_goroutines":            info.NGoroutines,
		"n_listener_events":       info.NEventsListener,
	}
	// Add metrics
	acc.AddFields("docker",
		fields,
		nil,
		now)
	acc.AddFields("docker",
		map[string]interface{}{"memory_total": info.MemTotal},
		map[string]string{"unit": "bytes"},
		now)
	// Get storage metrics
	for _, rawData := range info.DriverStatus {
		// Try to convert string to int (bytes)
		value, err := parseSize(rawData[1])
		if err != nil {
			continue
		}
		name := strings.ToLower(strings.Replace(rawData[0], " ", "_", -1))
		if name == "pool_blocksize" {
			// pool blocksize
			acc.AddFields("docker",
				map[string]interface{}{"pool_blocksize": value},
				map[string]string{"unit": "bytes"},
				now)
		} else if strings.HasPrefix(name, "data_space_") {
			// data space
			fieldName := strings.TrimPrefix(name, "data_space_")
			dataFields[fieldName] = value
		} else if strings.HasPrefix(name, "metadata_space_") {
			// metadata space
			fieldName := strings.TrimPrefix(name, "metadata_space_")
			metadataFields[fieldName] = value
		}
	}
	if len(dataFields) > 0 {
		acc.AddFields("docker_data",
			dataFields,
			map[string]string{"unit": "bytes"},
			now)
	}
	if len(metadataFields) > 0 {
		acc.AddFields("docker_metadata",
			metadataFields,
			map[string]string{"unit": "bytes"},
			now)
	}
	return nil
}

func (d *Docker) gatherContainer(
	container types.Container,
	acc telegraf.Accumulator,
) error {
	var v *types.StatsJSON
	// Parse container name
	cname := "unknown"
	if len(container.Names) > 0 {
		// Not sure what to do with other names, just take the first.
		cname = strings.TrimPrefix(container.Names[0], "/")
	}

	tags := map[string]string{
		"container_name":  cname,
		"container_image": container.Image,
	}
	if len(d.ContainerNames) > 0 {
		if !sliceContains(cname, d.ContainerNames) {
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()
	r, err := d.client.ContainerStats(ctx, container.ID, false)
	if err != nil {
		return fmt.Errorf("Error getting docker stats: %s", err.Error())
	}
	defer r.Close()
	dec := json.NewDecoder(r)
	if err = dec.Decode(&v); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("Error decoding: %s", err.Error())
	}

	// Add labels to tags
	for k, label := range container.Labels {
		tags[k] = label
	}

	gatherContainerStats(v, acc, tags, container.ID)

	return nil
}

func gatherContainerStats(
	stat *types.StatsJSON,
	acc telegraf.Accumulator,
	tags map[string]string,
	id string,
) {
	now := stat.Read

	memfields := map[string]interface{}{
		"max_usage":                 stat.MemoryStats.MaxUsage,
		"usage":                     stat.MemoryStats.Usage,
		"fail_count":                stat.MemoryStats.Failcnt,
		"limit":                     stat.MemoryStats.Limit,
		"total_pgmafault":           stat.MemoryStats.Stats["total_pgmajfault"],
		"cache":                     stat.MemoryStats.Stats["cache"],
		"mapped_file":               stat.MemoryStats.Stats["mapped_file"],
		"total_inactive_file":       stat.MemoryStats.Stats["total_inactive_file"],
		"pgpgout":                   stat.MemoryStats.Stats["pagpgout"],
		"rss":                       stat.MemoryStats.Stats["rss"],
		"total_mapped_file":         stat.MemoryStats.Stats["total_mapped_file"],
		"writeback":                 stat.MemoryStats.Stats["writeback"],
		"unevictable":               stat.MemoryStats.Stats["unevictable"],
		"pgpgin":                    stat.MemoryStats.Stats["pgpgin"],
		"total_unevictable":         stat.MemoryStats.Stats["total_unevictable"],
		"pgmajfault":                stat.MemoryStats.Stats["pgmajfault"],
		"total_rss":                 stat.MemoryStats.Stats["total_rss"],
		"total_rss_huge":            stat.MemoryStats.Stats["total_rss_huge"],
		"total_writeback":           stat.MemoryStats.Stats["total_write_back"],
		"total_inactive_anon":       stat.MemoryStats.Stats["total_inactive_anon"],
		"rss_huge":                  stat.MemoryStats.Stats["rss_huge"],
		"hierarchical_memory_limit": stat.MemoryStats.Stats["hierarchical_memory_limit"],
		"total_pgfault":             stat.MemoryStats.Stats["total_pgfault"],
		"total_active_file":         stat.MemoryStats.Stats["total_active_file"],
		"active_anon":               stat.MemoryStats.Stats["active_anon"],
		"total_active_anon":         stat.MemoryStats.Stats["total_active_anon"],
		"total_pgpgout":             stat.MemoryStats.Stats["total_pgpgout"],
		"total_cache":               stat.MemoryStats.Stats["total_cache"],
		"inactive_anon":             stat.MemoryStats.Stats["inactive_anon"],
		"active_file":               stat.MemoryStats.Stats["active_file"],
		"pgfault":                   stat.MemoryStats.Stats["pgfault"],
		"inactive_file":             stat.MemoryStats.Stats["inactive_file"],
		"total_pgpgin":              stat.MemoryStats.Stats["total_pgpgin"],
		"usage_percent":             calculateMemPercent(stat),
		"container_id":              id,
	}
	acc.AddFields("docker_container_mem", memfields, tags, now)

	cpufields := map[string]interface{}{
		"usage_total":                  stat.CPUStats.CPUUsage.TotalUsage,
		"usage_in_usermode":            stat.CPUStats.CPUUsage.UsageInUsermode,
		"usage_in_kernelmode":          stat.CPUStats.CPUUsage.UsageInKernelmode,
		"usage_system":                 stat.CPUStats.SystemUsage,
		"throttling_periods":           stat.CPUStats.ThrottlingData.Periods,
		"throttling_throttled_periods": stat.CPUStats.ThrottlingData.ThrottledPeriods,
		"throttling_throttled_time":    stat.CPUStats.ThrottlingData.ThrottledTime,
		"usage_percent":                calculateCPUPercent(stat),
		"container_id":                 id,
	}
	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	acc.AddFields("docker_container_cpu", cpufields, cputags, now)

	for i, percpu := range stat.CPUStats.CPUUsage.PercpuUsage {
		percputags := copyTags(tags)
		percputags["cpu"] = fmt.Sprintf("cpu%d", i)
		fields := map[string]interface{}{
			"usage_total":  percpu,
			"container_id": id,
		}
		acc.AddFields("docker_container_cpu", fields, percputags, now)
	}

	for network, netstats := range stat.Networks {
		netfields := map[string]interface{}{
			"rx_dropped":   netstats.RxDropped,
			"rx_bytes":     netstats.RxBytes,
			"rx_errors":    netstats.RxErrors,
			"tx_packets":   netstats.TxPackets,
			"tx_dropped":   netstats.TxDropped,
			"rx_packets":   netstats.RxPackets,
			"tx_errors":    netstats.TxErrors,
			"tx_bytes":     netstats.TxBytes,
			"container_id": id,
		}
		// Create a new network tag dictionary for the "network" tag
		nettags := copyTags(tags)
		nettags["network"] = network
		acc.AddFields("docker_container_net", netfields, nettags, now)
	}

	gatherBlockIOMetrics(stat, acc, tags, now, id)
}

func calculateMemPercent(stat *types.StatsJSON) float64 {
	var memPercent = 0.0
	if stat.MemoryStats.Limit > 0 {
		memPercent = float64(stat.MemoryStats.Usage) / float64(stat.MemoryStats.Limit) * 100.0
	}
	return memPercent
}

func calculateCPUPercent(stat *types.StatsJSON) float64 {
	var cpuPercent = 0.0
	// calculate the change for the cpu and system usage of the container in between readings
	cpuDelta := float64(stat.CPUStats.CPUUsage.TotalUsage) - float64(stat.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stat.CPUStats.SystemUsage) - float64(stat.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stat.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func gatherBlockIOMetrics(
	stat *types.StatsJSON,
	acc telegraf.Accumulator,
	tags map[string]string,
	now time.Time,
	id string,
) {
	blkioStats := stat.BlkioStats
	// Make a map of devices to their block io stats
	deviceStatMap := make(map[string]map[string]interface{})

	for _, metric := range blkioStats.IoServiceBytesRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_service_bytes_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoServicedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_serviced_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoQueuedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_queue_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoServiceTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_service_time_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoWaitTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_wait_time_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoMergedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_merged_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		deviceStatMap[device]["io_time_recursive"] = metric.Value
	}

	for _, metric := range blkioStats.SectorsRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		deviceStatMap[device]["sectors_recursive"] = metric.Value
	}

	for device, fields := range deviceStatMap {
		iotags := copyTags(tags)
		iotags["device"] = device
		fields["container_id"] = id
		acc.AddFields("docker_container_blkio", fields, iotags, now)
	}
}

func copyTags(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func sliceContains(in string, sl []string) bool {
	for _, str := range sl {
		if str == in {
			return true
		}
	}
	return false
}

// Parses the human-readable size string into the amount it represents.
func parseSize(sizeStr string) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 4 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}

	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1, err
	}

	uMap := map[string]int64{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
	unitPrefix := strings.ToLower(matches[3])
	if mul, ok := uMap[unitPrefix]; ok {
		size *= float64(mul)
	}

	return int64(size), nil
}

func init() {
	inputs.Add("docker", func() telegraf.Input {
		return &Docker{
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
