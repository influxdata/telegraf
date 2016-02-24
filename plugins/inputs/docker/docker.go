package system

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	Endpoint       string
	ContainerNames []string

	client DockerClient
}

type DockerClient interface {
	// Docker Client wrapper
	// Useful for test
	Info() (*docker.Env, error)
	ListContainers(opts docker.ListContainersOptions) ([]docker.APIContainers, error)
	Stats(opts docker.StatsOptions) error
}

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
`

func (d *Docker) Description() string {
	return "Read metrics about docker containers"
}

func (d *Docker) SampleConfig() string { return sampleConfig }

func (d *Docker) Gather(acc telegraf.Accumulator) error {
	if d.client == nil {
		var c *docker.Client
		var err error
		if d.Endpoint == "ENV" {
			c, err = docker.NewClientFromEnv()
			if err != nil {
				return err
			}
		} else if d.Endpoint == "" {
			c, err = docker.NewClient("unix:///var/run/docker.sock")
			if err != nil {
				return err
			}
		} else {
			c, err = docker.NewClient(d.Endpoint)
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
	opts := docker.ListContainersOptions{}
	containers, err := d.client.ListContainers(opts)
	if err != nil {
		return err
	}

	// Get container data
	var wg sync.WaitGroup
	wg.Add(len(containers))
	for _, container := range containers {

		go func(c docker.APIContainers) {
			defer wg.Done()
			err := d.gatherContainer(c, acc)
			if err != nil {
				fmt.Println(err.Error())
			}
		}(container)
	}
	wg.Wait()

	return nil
}

func (d *Docker) gatherInfo(acc telegraf.Accumulator) error {
	// Init vars
	var driverStatus [][]string
	dataFields := make(map[string]interface{})
	metadataFields := make(map[string]interface{})
	now := time.Now()
	// Get info from docker daemon
	info, err := d.client.Info()
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"n_cpus":                  info.GetInt64("NCPU"),
		"n_used_file_descriptors": info.GetInt64("NFd"),
		"n_containers":            info.GetInt64("Containers"),
		"n_images":                info.GetInt64("Images"),
		"n_goroutines":            info.GetInt64("NGoroutines"),
		"n_listener_events":       info.GetInt64("NEventsListener"),
	}
	// Add metrics
	acc.AddFields("docker",
		fields,
		nil,
		now)
	acc.AddFields("docker",
		map[string]interface{}{"memory_total": info.GetInt64("MemTotal")},
		map[string]string{"unit": "bytes"},
		now)
	// Get storage metrics
	driverStatusRaw := []byte(info.Get("DriverStatus"))
	json.Unmarshal(driverStatusRaw, &driverStatus)
	for _, rawData := range driverStatus {
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
			field_name := strings.TrimPrefix(name, "data_space_")
			dataFields[field_name] = value
		} else if strings.HasPrefix(name, "metadata_space_") {
			// metadata space
			field_name := strings.TrimPrefix(name, "metadata_space_")
			metadataFields[field_name] = value
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
	container docker.APIContainers,
	acc telegraf.Accumulator,
) error {
	// Parse container name
	cname := "unknown"
	if len(container.Names) > 0 {
		// Not sure what to do with other names, just take the first.
		cname = strings.TrimPrefix(container.Names[0], "/")
	}

	tags := map[string]string{
		"cont_id":    container.ID,
		"cont_name":  cname,
		"cont_image": container.Image,
	}
	if len(d.ContainerNames) > 0 {
		if !sliceContains(cname, d.ContainerNames) {
			return nil
		}
	}

	statChan := make(chan *docker.Stats)
	done := make(chan bool)
	statOpts := docker.StatsOptions{
		Stream:  false,
		ID:      container.ID,
		Stats:   statChan,
		Done:    done,
		Timeout: time.Duration(time.Second * 5),
	}

	go func() {
		err := d.client.Stats(statOpts)
		if err != nil {
			log.Printf("Error getting docker stats: %s\n", err.Error())
		}
	}()

	stat := <-statChan
	close(done)

	if stat == nil {
		return nil
	}

	// Add labels to tags
	for k, v := range container.Labels {
		tags[k] = v
	}

	gatherContainerStats(stat, acc, tags)

	return nil
}

func gatherContainerStats(
	stat *docker.Stats,
	acc telegraf.Accumulator,
	tags map[string]string,
) {
	now := stat.Read

	memfields := map[string]interface{}{
		"max_usage":                 stat.MemoryStats.MaxUsage,
		"usage":                     stat.MemoryStats.Usage,
		"fail_count":                stat.MemoryStats.Failcnt,
		"limit":                     stat.MemoryStats.Limit,
		"total_pgmafault":           stat.MemoryStats.Stats.TotalPgmafault,
		"cache":                     stat.MemoryStats.Stats.Cache,
		"mapped_file":               stat.MemoryStats.Stats.MappedFile,
		"total_inactive_file":       stat.MemoryStats.Stats.TotalInactiveFile,
		"pgpgout":                   stat.MemoryStats.Stats.Pgpgout,
		"rss":                       stat.MemoryStats.Stats.Rss,
		"total_mapped_file":         stat.MemoryStats.Stats.TotalMappedFile,
		"writeback":                 stat.MemoryStats.Stats.Writeback,
		"unevictable":               stat.MemoryStats.Stats.Unevictable,
		"pgpgin":                    stat.MemoryStats.Stats.Pgpgin,
		"total_unevictable":         stat.MemoryStats.Stats.TotalUnevictable,
		"pgmajfault":                stat.MemoryStats.Stats.Pgmajfault,
		"total_rss":                 stat.MemoryStats.Stats.TotalRss,
		"total_rss_huge":            stat.MemoryStats.Stats.TotalRssHuge,
		"total_writeback":           stat.MemoryStats.Stats.TotalWriteback,
		"total_inactive_anon":       stat.MemoryStats.Stats.TotalInactiveAnon,
		"rss_huge":                  stat.MemoryStats.Stats.RssHuge,
		"hierarchical_memory_limit": stat.MemoryStats.Stats.HierarchicalMemoryLimit,
		"total_pgfault":             stat.MemoryStats.Stats.TotalPgfault,
		"total_active_file":         stat.MemoryStats.Stats.TotalActiveFile,
		"active_anon":               stat.MemoryStats.Stats.ActiveAnon,
		"total_active_anon":         stat.MemoryStats.Stats.TotalActiveAnon,
		"total_pgpgout":             stat.MemoryStats.Stats.TotalPgpgout,
		"total_cache":               stat.MemoryStats.Stats.TotalCache,
		"inactive_anon":             stat.MemoryStats.Stats.InactiveAnon,
		"active_file":               stat.MemoryStats.Stats.ActiveFile,
		"pgfault":                   stat.MemoryStats.Stats.Pgfault,
		"inactive_file":             stat.MemoryStats.Stats.InactiveFile,
		"total_pgpgin":              stat.MemoryStats.Stats.TotalPgpgin,
		"usage_percent":             calculateMemPercent(stat),
	}
	acc.AddFields("docker_mem", memfields, tags, now)

	cpufields := map[string]interface{}{
		"usage_total":                  stat.CPUStats.CPUUsage.TotalUsage,
		"usage_in_usermode":            stat.CPUStats.CPUUsage.UsageInUsermode,
		"usage_in_kernelmode":          stat.CPUStats.CPUUsage.UsageInKernelmode,
		"usage_system":                 stat.CPUStats.SystemCPUUsage,
		"throttling_periods":           stat.CPUStats.ThrottlingData.Periods,
		"throttling_throttled_periods": stat.CPUStats.ThrottlingData.ThrottledPeriods,
		"throttling_throttled_time":    stat.CPUStats.ThrottlingData.ThrottledTime,
		"usage_percent":                calculateCPUPercent(stat),
	}
	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	acc.AddFields("docker_cpu", cpufields, cputags, now)

	for i, percpu := range stat.CPUStats.CPUUsage.PercpuUsage {
		percputags := copyTags(tags)
		percputags["cpu"] = fmt.Sprintf("cpu%d", i)
		acc.AddFields("docker_cpu", map[string]interface{}{"usage_total": percpu}, percputags, now)
	}

	for network, netstats := range stat.Networks {
		netfields := map[string]interface{}{
			"rx_dropped": netstats.RxDropped,
			"rx_bytes":   netstats.RxBytes,
			"rx_errors":  netstats.RxErrors,
			"tx_packets": netstats.TxPackets,
			"tx_dropped": netstats.TxDropped,
			"rx_packets": netstats.RxPackets,
			"tx_errors":  netstats.TxErrors,
			"tx_bytes":   netstats.TxBytes,
		}
		// Create a new network tag dictionary for the "network" tag
		nettags := copyTags(tags)
		nettags["network"] = network
		acc.AddFields("docker_net", netfields, nettags, now)
	}

	gatherBlockIOMetrics(stat, acc, tags, now)
}

func calculateMemPercent(stat *docker.Stats) float64 {
	var memPercent = 0.0
	if stat.MemoryStats.Limit > 0 {
		memPercent = float64(stat.MemoryStats.Usage) / float64(stat.MemoryStats.Limit) * 100.0
	}
	return memPercent
}

func calculateCPUPercent(stat *docker.Stats) float64 {
	var cpuPercent = 0.0
	// calculate the change for the cpu and system usage of the container in between readings
	cpuDelta := float64(stat.CPUStats.CPUUsage.TotalUsage) - float64(stat.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stat.CPUStats.SystemCPUUsage) - float64(stat.PreCPUStats.SystemCPUUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stat.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func gatherBlockIOMetrics(
	stat *docker.Stats,
	acc telegraf.Accumulator,
	tags map[string]string,
	now time.Time,
) {
	blkioStats := stat.BlkioStats
	// Make a map of devices to their block io stats
	deviceStatMap := make(map[string]map[string]interface{})

	for _, metric := range blkioStats.IOServiceBytesRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_service_bytes_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IOServicedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_serviced_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IOQueueRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_queue_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IOServiceTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_service_time_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IOWaitTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_wait_time_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IOMergedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_merged_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IOTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_time_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.SectorsRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("sectors_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for device, fields := range deviceStatMap {
		iotags := copyTags(tags)
		iotags["device"] = device
		acc.AddFields("docker_blkio", fields, iotags, now)
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
		return &Docker{}
	})
}
