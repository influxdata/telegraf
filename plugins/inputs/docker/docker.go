package docker

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Docker object
type Docker struct {
	Endpoint       string
	ContainerNames []string // deprecated in 1.4; use container_name_include

	GatherServices bool `toml:"gather_services"`

	Timeout        internal.Duration
	PerDevice      bool     `toml:"perdevice"`
	Total          bool     `toml:"total"`
	TagEnvironment []string `toml:"tag_env"`
	LabelInclude   []string `toml:"docker_label_include"`
	LabelExclude   []string `toml:"docker_label_exclude"`

	ContainerInclude []string `toml:"container_name_include"`
	ContainerExclude []string `toml:"container_name_exclude"`

	ContainerStateInclude []string `toml:"container_state_include"`
	ContainerStateExclude []string `toml:"container_state_exclude"`

	tlsint.ClientConfig

	newEnvClient func() (Client, error)
	newClient    func(string, *tls.Config) (Client, error)

	client          Client
	httpClient      *http.Client
	engine_host     string
	serverVersion   string
	filtersCreated  bool
	labelFilter     filter.Filter
	containerFilter filter.Filter
	stateFilter     filter.Filter
}

// KB, MB, GB, TB, PB...human friendly
const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	defaultEndpoint = "unix:///var/run/docker.sock"
)

var (
	sizeRegex       = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[bB]?$`)
	containerStates = []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}
)

var sampleConfig = `
  ## Docker Endpoint
  ##   To use TCP, set endpoint = "tcp://[ip]:[port]"
  ##   To use environment variables (ie, docker-machine), set endpoint = "ENV"
  endpoint = "unix:///var/run/docker.sock"

  ## Set to true to collect Swarm metrics(desired_replicas, running_replicas)
  gather_services = false

  ## Only collect metrics for these containers, collect all if empty
  container_names = []

  ## Containers to include and exclude. Globs accepted.
  ## Note that an empty array for both will include all containers
  container_name_include = []
  container_name_exclude = []

  ## Container states to include and exclude. Globs accepted.
  ## When empty only containers in the "running" state will be captured.
  # container_state_include = []
  # container_state_exclude = []

  ## Timeout for docker list, info, and stats commands
  timeout = "5s"

  ## Whether to report for each container per-device blkio (8:0, 8:1...) and
  ## network (eth0, eth1, ...) stats or not
  perdevice = true
  ## Whether to report for each container total blkio and network stats or not
  total = false
  ## Which environment variables should we use as a tag
  ##tag_env = ["JAVA_HOME", "HEAP_SIZE"]

  ## docker labels to include and exclude as tags.  Globs accepted.
  ## Note that an empty array for both will include all labels as tags
  docker_label_include = []
  docker_label_exclude = []

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false
`

func (d *Docker) Description() string {
	return "Read metrics about docker containers"
}

func (d *Docker) SampleConfig() string { return sampleConfig }

func (d *Docker) Gather(acc telegraf.Accumulator) error {
	if d.client == nil {
		c, err := d.getNewClient()
		if err != nil {
			return err
		}
		d.client = c
	}

	// Create label filters if not already created
	if !d.filtersCreated {
		err := d.createLabelFilters()
		if err != nil {
			return err
		}
		err = d.createContainerFilters()
		if err != nil {
			return err
		}
		err = d.createContainerStateFilters()
		if err != nil {
			return err
		}
		d.filtersCreated = true
	}

	// Get daemon info
	err := d.gatherInfo(acc)
	if err != nil {
		acc.AddError(err)
	}

	if d.GatherServices {
		err := d.gatherSwarmInfo(acc)
		if err != nil {
			acc.AddError(err)
		}
	}

	filterArgs := filters.NewArgs()
	for _, state := range containerStates {
		if d.stateFilter.Match(state) {
			filterArgs.Add("status", state)
		}
	}

	// All container states were excluded
	if filterArgs.Len() == 0 {
		return nil
	}

	// List containers
	opts := types.ContainerListOptions{
		Filters: filterArgs,
	}
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
				acc.AddError(fmt.Errorf("E! Error gathering container %s stats: %s\n",
					c.Names, err.Error()))
			}
		}(container)
	}
	wg.Wait()

	return nil
}

func (d *Docker) gatherSwarmInfo(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()
	services, err := d.client.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return err
	}

	if len(services) > 0 {
		tasks, err := d.client.TaskList(ctx, types.TaskListOptions{})
		if err != nil {
			return err
		}

		nodes, err := d.client.NodeList(ctx, types.NodeListOptions{})
		if err != nil {
			return err
		}

		running := map[string]int{}
		tasksNoShutdown := map[string]int{}

		activeNodes := make(map[string]struct{})
		for _, n := range nodes {
			if n.Status.State != swarm.NodeStateDown {
				activeNodes[n.ID] = struct{}{}
			}
		}

		for _, task := range tasks {
			if task.DesiredState != swarm.TaskStateShutdown {
				tasksNoShutdown[task.ServiceID]++
			}

			if task.Status.State == swarm.TaskStateRunning {
				running[task.ServiceID]++
			}
		}

		for _, service := range services {
			tags := map[string]string{}
			fields := make(map[string]interface{})
			now := time.Now()
			tags["service_id"] = service.ID
			tags["service_name"] = service.Spec.Name
			if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
				tags["service_mode"] = "replicated"
				fields["tasks_running"] = running[service.ID]
				fields["tasks_desired"] = *service.Spec.Mode.Replicated.Replicas
			} else if service.Spec.Mode.Global != nil {
				tags["service_mode"] = "global"
				fields["tasks_running"] = running[service.ID]
				fields["tasks_desired"] = tasksNoShutdown[service.ID]
			} else {
				log.Printf("E! Unknow Replicas Mode")
			}
			// Add metrics
			acc.AddFields("docker_swarm",
				fields,
				tags,
				now)
		}
	}

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

	d.engine_host = info.Name
	d.serverVersion = info.ServerVersion

	tags := map[string]string{
		"engine_host":    d.engine_host,
		"server_version": d.serverVersion,
	}

	fields := map[string]interface{}{
		"n_cpus":                  info.NCPU,
		"n_used_file_descriptors": info.NFd,
		"n_containers":            info.Containers,
		"n_containers_running":    info.ContainersRunning,
		"n_containers_stopped":    info.ContainersStopped,
		"n_containers_paused":     info.ContainersPaused,
		"n_images":                info.Images,
		"n_goroutines":            info.NGoroutines,
		"n_listener_events":       info.NEventsListener,
	}
	// Add metrics
	acc.AddFields("docker", fields, tags, now)
	acc.AddFields("docker",
		map[string]interface{}{"memory_total": info.MemTotal},
		tags,
		now)
	// Get storage metrics
	tags["unit"] = "bytes"
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
				tags,
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
		acc.AddFields("docker_data", dataFields, tags, now)
	}
	if len(metadataFields) > 0 {
		acc.AddFields("docker_metadata", metadataFields, tags, now)
	}
	return nil
}

func (d *Docker) gatherContainer(
	container types.Container,
	acc telegraf.Accumulator,
) error {
	var v *types.StatsJSON
	// Parse container name
	var cname string
	for _, name := range container.Names {
		trimmedName := strings.TrimPrefix(name, "/")
		match := d.containerFilter.Match(trimmedName)
		if match {
			cname = trimmedName
			break
		}
	}

	if cname == "" {
		return nil
	}

	// the image name sometimes has a version part, or a private repo
	//   ie, rabbitmq:3-management or docker.someco.net:4443/rabbitmq:3-management
	imageName := ""
	imageVersion := "unknown"
	i := strings.LastIndex(container.Image, ":") // index of last ':' character
	if i > -1 {
		imageVersion = container.Image[i+1:]
		imageName = container.Image[:i]
	} else {
		imageName = container.Image
	}

	tags := map[string]string{
		"engine_host":       d.engine_host,
		"server_version":    d.serverVersion,
		"container_name":    cname,
		"container_image":   imageName,
		"container_version": imageVersion,
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()
	r, err := d.client.ContainerStats(ctx, container.ID, false)
	if err != nil {
		return fmt.Errorf("Error getting docker stats: %s", err.Error())
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&v); err != nil {
		if err == io.EOF {
			return nil
		}
		return fmt.Errorf("Error decoding: %s", err.Error())
	}
	daemonOSType := r.OSType

	// use common (printed at `docker ps`) name for container
	if v.Name != "" {
		tags["container_name"] = strings.TrimPrefix(v.Name, "/")
	}

	// Add labels to tags
	for k, label := range container.Labels {
		if d.labelFilter.Match(k) {
			tags[k] = label
		}
	}

	info, err := d.client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return fmt.Errorf("Error inspecting docker container: %s", err.Error())
	}

	// Add whitelisted environment variables to tags
	if len(d.TagEnvironment) > 0 {
		for _, envvar := range info.Config.Env {
			for _, configvar := range d.TagEnvironment {
				dock_env := strings.SplitN(envvar, "=", 2)
				//check for presence of tag in whitelist
				if len(dock_env) == 2 && len(strings.TrimSpace(dock_env[1])) != 0 && configvar == dock_env[0] {
					tags[dock_env[0]] = dock_env[1]
				}
			}
		}
	}

	if info.State != nil {
		tags["container_status"] = info.State.Status
		statefields := map[string]interface{}{
			"oomkilled": info.State.OOMKilled,
			"pid":       info.State.Pid,
			"exitcode":  info.State.ExitCode,
		}
		container_time, err := time.Parse(time.RFC3339, info.State.StartedAt)
		if err == nil && !container_time.IsZero() {
			statefields["started_at"] = container_time.UnixNano()
		}
		container_time, err = time.Parse(time.RFC3339, info.State.FinishedAt)
		if err == nil && !container_time.IsZero() {
			statefields["finished_at"] = container_time.UnixNano()
		}
		acc.AddFields("docker_container_status", statefields, tags, time.Now())

		if info.State.Health != nil {
			healthfields := map[string]interface{}{
				"health_status":  info.State.Health.Status,
				"failing_streak": info.ContainerJSONBase.State.Health.FailingStreak,
			}
			acc.AddFields("docker_container_health", healthfields, tags, time.Now())
		}
	}

	parseContainerStats(v, acc, tags, container.ID, d.PerDevice, d.Total, daemonOSType)

	return nil
}

func parseContainerStats(
	stat *types.StatsJSON,
	acc telegraf.Accumulator,
	tags map[string]string,
	id string,
	perDevice bool,
	total bool,
	daemonOSType string,
) {
	tm := stat.Read

	if tm.Before(time.Unix(0, 0)) {
		tm = time.Now()
	}

	memfields := map[string]interface{}{
		"container_id": id,
	}

	memstats := []string{
		"active_anon",
		"active_file",
		"cache",
		"hierarchical_memory_limit",
		"inactive_anon",
		"inactive_file",
		"mapped_file",
		"pgfault",
		"pgmajfault",
		"pgpgin",
		"pgpgout",
		"rss",
		"rss_huge",
		"total_active_anon",
		"total_active_file",
		"total_cache",
		"total_inactive_anon",
		"total_inactive_file",
		"total_mapped_file",
		"total_pgfault",
		"total_pgmajfault",
		"total_pgpgin",
		"total_pgpgout",
		"total_rss",
		"total_rss_huge",
		"total_unevictable",
		"total_writeback",
		"unevictable",
		"writeback",
	}
	for _, field := range memstats {
		if value, ok := stat.MemoryStats.Stats[field]; ok {
			memfields[field] = value
		}
	}
	if stat.MemoryStats.Failcnt != 0 {
		memfields["fail_count"] = stat.MemoryStats.Failcnt
	}

	if daemonOSType != "windows" {
		memfields["limit"] = stat.MemoryStats.Limit
		memfields["max_usage"] = stat.MemoryStats.MaxUsage

		mem := calculateMemUsageUnixNoCache(stat.MemoryStats)
		memLimit := float64(stat.MemoryStats.Limit)
		memfields["usage"] = uint64(mem)
		memfields["usage_percent"] = calculateMemPercentUnixNoCache(memLimit, mem)
	} else {
		memfields["commit_bytes"] = stat.MemoryStats.Commit
		memfields["commit_peak_bytes"] = stat.MemoryStats.CommitPeak
		memfields["private_working_set"] = stat.MemoryStats.PrivateWorkingSet
	}

	acc.AddFields("docker_container_mem", memfields, tags, tm)

	cpufields := map[string]interface{}{
		"usage_total":                  stat.CPUStats.CPUUsage.TotalUsage,
		"usage_in_usermode":            stat.CPUStats.CPUUsage.UsageInUsermode,
		"usage_in_kernelmode":          stat.CPUStats.CPUUsage.UsageInKernelmode,
		"usage_system":                 stat.CPUStats.SystemUsage,
		"throttling_periods":           stat.CPUStats.ThrottlingData.Periods,
		"throttling_throttled_periods": stat.CPUStats.ThrottlingData.ThrottledPeriods,
		"throttling_throttled_time":    stat.CPUStats.ThrottlingData.ThrottledTime,
		"container_id":                 id,
	}

	if daemonOSType != "windows" {
		previousCPU := stat.PreCPUStats.CPUUsage.TotalUsage
		previousSystem := stat.PreCPUStats.SystemUsage
		cpuPercent := calculateCPUPercentUnix(previousCPU, previousSystem, stat)
		cpufields["usage_percent"] = cpuPercent
	} else {
		cpuPercent := calculateCPUPercentWindows(stat)
		cpufields["usage_percent"] = cpuPercent
	}

	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	acc.AddFields("docker_container_cpu", cpufields, cputags, tm)

	// If we have OnlineCPUs field, then use it to restrict stats gathering to only Online CPUs
	// (https://github.com/moby/moby/commit/115f91d7575d6de6c7781a96a082f144fd17e400)
	var percpuusage []uint64
	if stat.CPUStats.OnlineCPUs > 0 {
		percpuusage = stat.CPUStats.CPUUsage.PercpuUsage[:stat.CPUStats.OnlineCPUs]
	} else {
		percpuusage = stat.CPUStats.CPUUsage.PercpuUsage
	}

	for i, percpu := range percpuusage {
		percputags := copyTags(tags)
		percputags["cpu"] = fmt.Sprintf("cpu%d", i)
		fields := map[string]interface{}{
			"usage_total":  percpu,
			"container_id": id,
		}
		acc.AddFields("docker_container_cpu", fields, percputags, tm)
	}

	totalNetworkStatMap := make(map[string]interface{})
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
		if perDevice {
			nettags := copyTags(tags)
			nettags["network"] = network
			acc.AddFields("docker_container_net", netfields, nettags, tm)
		}
		if total {
			for field, value := range netfields {
				if field == "container_id" {
					continue
				}

				var uintV uint64
				switch v := value.(type) {
				case uint64:
					uintV = v
				case int64:
					uintV = uint64(v)
				default:
					continue
				}

				_, ok := totalNetworkStatMap[field]
				if ok {
					totalNetworkStatMap[field] = totalNetworkStatMap[field].(uint64) + uintV
				} else {
					totalNetworkStatMap[field] = uintV
				}
			}
		}
	}

	// totalNetworkStatMap could be empty if container is running with --net=host.
	if total && len(totalNetworkStatMap) != 0 {
		nettags := copyTags(tags)
		nettags["network"] = "total"
		totalNetworkStatMap["container_id"] = id
		acc.AddFields("docker_container_net", totalNetworkStatMap, nettags, tm)
	}

	gatherBlockIOMetrics(stat, acc, tags, tm, id, perDevice, total)
}

func gatherBlockIOMetrics(
	stat *types.StatsJSON,
	acc telegraf.Accumulator,
	tags map[string]string,
	tm time.Time,
	id string,
	perDevice bool,
	total bool,
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

	totalStatMap := make(map[string]interface{})
	for device, fields := range deviceStatMap {
		fields["container_id"] = id
		if perDevice {
			iotags := copyTags(tags)
			iotags["device"] = device
			acc.AddFields("docker_container_blkio", fields, iotags, tm)
		}
		if total {
			for field, value := range fields {
				if field == "container_id" {
					continue
				}

				var uintV uint64
				switch v := value.(type) {
				case uint64:
					uintV = v
				case int64:
					uintV = uint64(v)
				default:
					continue
				}

				_, ok := totalStatMap[field]
				if ok {
					totalStatMap[field] = totalStatMap[field].(uint64) + uintV
				} else {
					totalStatMap[field] = uintV
				}
			}
		}
	}
	if total {
		totalStatMap["container_id"] = id
		iotags := copyTags(tags)
		iotags["device"] = "total"
		acc.AddFields("docker_container_blkio", totalStatMap, iotags, tm)
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

func (d *Docker) createContainerFilters() error {
	// Backwards compatibility for deprecated `container_names` parameter.
	if len(d.ContainerNames) > 0 {
		d.ContainerInclude = append(d.ContainerInclude, d.ContainerNames...)
	}

	filter, err := filter.NewIncludeExcludeFilter(d.ContainerInclude, d.ContainerExclude)
	if err != nil {
		return err
	}
	d.containerFilter = filter
	return nil
}

func (d *Docker) createLabelFilters() error {
	filter, err := filter.NewIncludeExcludeFilter(d.LabelInclude, d.LabelExclude)
	if err != nil {
		return err
	}
	d.labelFilter = filter
	return nil
}

func (d *Docker) createContainerStateFilters() error {
	if len(d.ContainerStateInclude) == 0 && len(d.ContainerStateExclude) == 0 {
		d.ContainerStateInclude = []string{"running"}
	}
	filter, err := filter.NewIncludeExcludeFilter(d.ContainerStateInclude, d.ContainerStateExclude)
	if err != nil {
		return err
	}
	d.stateFilter = filter
	return nil
}

func (d *Docker) getNewClient() (Client, error) {
	if d.Endpoint == "ENV" {
		return d.newEnvClient()
	}

	tlsConfig, err := d.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	return d.newClient(d.Endpoint, tlsConfig)
}

func init() {
	inputs.Add("docker", func() telegraf.Input {
		return &Docker{
			PerDevice:      true,
			Timeout:        internal.Duration{Duration: time.Second * 5},
			Endpoint:       defaultEndpoint,
			newEnvClient:   NewEnvClient,
			newClient:      NewClient,
			filtersCreated: false,
		}
	})
}
